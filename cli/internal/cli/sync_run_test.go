package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// fakeEditor returns an editor stub that replaces only the body section
// of the existing CLI-written file, preserving the frontmatter prefix.
func fakeEditor(body string) func(string) error {
	return func(path string) error {
		existing, _ := os.ReadFile(path)
		// Find second "---\n" delimiter and append body after it.
		idx := bytes.Index(existing, []byte("---\n"))
		if idx < 0 {
			return os.WriteFile(path, []byte(body), 0o644)
		}
		next := bytes.Index(existing[idx+4:], []byte("---\n"))
		if next < 0 {
			return os.WriteFile(path, []byte(body), 0o644)
		}
		end := idx + 4 + next + 4
		out := append([]byte{}, existing[:end]...)
		out = append(out, '\n')
		out = append(out, []byte(body)...)
		return os.WriteFile(path, out, 0o644)
	}
}

// mkVaultForTest creates a t.TempDir-rooted vault and returns the vault path.
func mkVaultForTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// Isolate from the developer's real ~/.config/tolvi/roots.json: an empty
	// config dir is single-root mode, which is what an unrouted vault means.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	vaultDir := filepath.Join(root, "vault")
	for _, sub := range []string{"decisions", "sessions", "patterns"} {
		_ = os.MkdirAll(filepath.Join(vaultDir, sub), 0o755)
	}
	_ = vault.WriteMeta(vaultDir, vault.Meta{
		Workspace: "test", Repo: "test", EmbeddingModel: "nomic-embed-text",
		SchemaVersion: vault.SupportedSchemaVersion,
	})
	return vaultDir
}

// declareRoots declares an org root for workspace "test" in the config dir the
// current test is isolated to.
func declareRoots(t *testing.T, path string) {
	t.Helper()
	declareRawRoots(t, `{"roots":[{"role":"org","workspace":"test","path":"`+path+`"}]}`)
}

// declareRawRoots writes roots.json verbatim, for malformed-config cases.
func declareRawRoots(t *testing.T, body string) {
	t.Helper()
	dir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "tolvi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, vault.RootsConfigFileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunSync_DecisionHappyPath(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: vaultDir,
		DocType:   "decision",
		Title:     "Why we chose Postgres",
		Date:      time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		RunEditor: fakeEditor("# Postgres\n\nbecause pgvector.\n"),
		Stdout:    &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	wantPath := filepath.Join(vaultDir, "decisions", "2026-05-14-why-we-chose-postgres.md")
	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	if !bytes.Contains(data, []byte("status: active")) {
		t.Errorf("frontmatter missing status: %s", data)
	}
	if !bytes.Contains(data, []byte("because pgvector")) {
		t.Errorf("body not written: %s", data)
	}
}

func TestRunSync_RefusesOverwrite(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	target := filepath.Join(vaultDir, "decisions", "2026-05-14-existing.md")
	_ = os.WriteFile(target, []byte("preexisting\n"), 0o644)

	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: vaultDir,
		DocType:   "decision",
		Title:     "existing",
		Date:      time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		RunEditor: fakeEditor("# new\n\nbody\n"),
		Stdout:    &out,
	})
	if err == nil {
		t.Fatal("expected overwrite refusal")
	}

	data, _ := os.ReadFile(target)
	if string(data) != "preexisting\n" {
		t.Errorf("file was modified despite refusal: %s", data)
	}
}

// mkPublicVaultForTest creates a repo vault whose workspace declares an org
// root at a sibling private vault. Returns (repoVault, orgRoot).
func mkPublicVaultForTest(t *testing.T, name string) (string, string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	pub := filepath.Join(root, "public", "vault")
	priv := filepath.Join(root, "private", "vault")
	for _, base := range []string{pub, priv} {
		for _, sub := range []string{"decisions", "sessions", "patterns"} {
			_ = os.MkdirAll(filepath.Join(base, sub), 0o755)
		}
	}
	_ = vault.WriteMeta(pub, vault.Meta{
		Workspace: name, Repo: name, EmbeddingModel: "nomic-embed-text",
		SchemaVersion: vault.SupportedSchemaVersion,
	})
	declareRawRoots(t, `{"roots":[{"role":"org","workspace":"`+name+`","path":"`+priv+`"}]}`)
	return pub, priv
}

func TestRunSync_RoutedSession_LandsInTheOrgRootUnderTheRepoName(t *testing.T) {
	pub, priv := mkPublicVaultForTest(t, "acme")
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: pub,
		DocType:   "session",
		Date:      time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		BodyFlag:  "## [10:00] Session — did stuff\n",
		Stdout:    &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	// Session must land in the private vault under a workspace-suffixed name.
	routed := filepath.Join(priv, "sessions", "2026-07-25-acme.md")
	if _, err := os.ReadFile(routed); err != nil {
		t.Fatalf("routed session not written to private vault: %v", err)
	}
	// And must NOT appear in the public vault at all.
	if _, err := os.Stat(filepath.Join(pub, "sessions", "2026-07-25.md")); err == nil {
		t.Error("session leaked into the public vault")
	}
	if _, err := os.Stat(filepath.Join(pub, "sessions", "2026-07-25-acme.md")); err == nil {
		t.Error("session leaked into the public vault (suffixed name)")
	}
}

func TestRunSync_PublicDecision_StaysPublic(t *testing.T) {
	pub, priv := mkPublicVaultForTest(t, "acme")
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: pub,
		DocType:   "decision",
		Title:     "Chose Postgres",
		Date:      time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		BodyFlag:  "# Postgres\n\nbecause pgvector.\n",
		Stdout:    &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(pub, "decisions", "2026-07-25-chose-postgres.md")); err != nil {
		t.Fatalf("public decision should stay in the public vault: %v", err)
	}
	if _, err := os.Stat(filepath.Join(priv, "decisions", "2026-07-25-chose-postgres.md")); err == nil {
		t.Error("public decision should not be routed to the private vault")
	}
}

func TestRunSync_PublicPrivateDecision_RoutesToPrivate(t *testing.T) {
	pub, priv := mkPublicVaultForTest(t, "acme")
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath:     pub,
		DocType:       "decision",
		Title:         "Secret sauce",
		DocVisibility: "private",
		Date:          time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		BodyFlag:      "# Secret\n\nhush.\n",
		Stdout:        &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(priv, "decisions", "2026-07-25-secret-sauce.md")); err != nil {
		t.Fatalf("private decision should route to the private vault: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pub, "decisions", "2026-07-25-secret-sauce.md")); err == nil {
		t.Error("private decision leaked into the public vault")
	}
}

func TestRunSync_DeclaredRootsWithoutAPrivateRoot_Errors(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	// Config exists but declares nothing for this workspace. That is a
	// misconfiguration, not a contributor, so the session must be refused
	// rather than written into the repo's public vault.
	declareRawRoots(t, `{"roots":[{"role":"org","workspace":"someone-else","path":"/elsewhere"}]}`)

	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: vaultDir,
		DocType:   "session",
		Date:      time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		BodyFlag:  "## [10:00] Session\n",
		Stdout:    &out,
	})
	if err == nil {
		t.Fatal("expected a refusal: declared roots with no private root for this workspace")
	}
	if _, statErr := os.Stat(filepath.Join(vaultDir, "sessions", "2026-07-25.md")); statErr == nil {
		t.Error("refused session must not have been written to the repo vault")
	}
}

func TestRunSync_PrivateVaultFlagOverridesTheChain(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	priv := t.TempDir()
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath:    vaultDir,
		DocType:      "session",
		Date:         time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		BodyFlag:     "## [10:00] Session\n",
		PrivateVault: priv,
		Stdout:       &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(priv, "sessions", "2026-07-25-test.md")); err != nil {
		t.Fatalf("session should route to --private-vault: %v", err)
	}
}

func TestRunSync_BodyFlag(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: vaultDir,
		DocType:   "pattern",
		Title:     "Idempotent migrations",
		Date:      time.Now(),
		BodyFlag:  "# Pattern\n\nbody from flag.\n",
		Stdout:    &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	created := filepath.Join(vaultDir, "patterns", "idempotent-migrations.md")
	data, _ := os.ReadFile(created)
	if !bytes.Contains(data, []byte("body from flag")) {
		t.Errorf("body flag not honored: %s", data)
	}
}

// TestAppendSessionBlock_ConcurrentAppendsKeepEveryBlock pins the fix for a
// real loss: a routed session note is a shared append target across concurrent
// sessions, and a read-modify-write of the whole file drops whatever another
// session added between the read and the write.
func TestAppendSessionBlock_ConcurrentAppendsKeepEveryBlock(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	day := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)

	// Seed the file so every goroutine takes the append path, not the create one.
	if err := RunSync(SyncOpts{
		VaultPath: vaultDir, DocType: "session", Date: day,
		BodyFlag: "## [09:00] Session — seed\n", Stdout: &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("seed RunSync: %v", err)
	}

	const n = 8
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs <- RunSync(SyncOpts{
				VaultPath: vaultDir, DocType: "session", Date: day,
				BodyFlag: fmt.Sprintf("## [1%d:00] Session — block %d\n", i, i),
				Stdout:   &bytes.Buffer{},
			})
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent RunSync: %v", err)
		}
	}

	data, err := os.ReadFile(filepath.Join(vaultDir, "sessions", "2026-09-12.md"))
	if err != nil {
		t.Fatalf("read session note: %v", err)
	}
	for i := 0; i < n; i++ {
		want := fmt.Sprintf("block %d", i)
		if !bytes.Contains(data, []byte(want)) {
			t.Errorf("%q was dropped by a concurrent append:\n%s", want, data)
		}
	}
}
