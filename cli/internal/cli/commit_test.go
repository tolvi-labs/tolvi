package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

func TestHasSessionNote(t *testing.T) {
	vault := t.TempDir()
	sessions := filepath.Join(vault, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}

	const date = "2026-06-07"

	// No file → false.
	if hasSessionNote(vault, date) {
		t.Error("expected false when session file is absent")
	}

	// File exists but no "## " heading → false (frontmatter-only skeleton).
	skeleton := "---\ntags: [session]\ndate: 2026-06-07\nstatus: active\n---\n"
	if err := os.WriteFile(filepath.Join(sessions, date+".md"), []byte(skeleton), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasSessionNote(vault, date) {
		t.Error("expected false when file has no '## ' session block")
	}

	// File with a session block → true.
	withBlock := skeleton + "\n## [10:00] Session — did the thing\n\n### What happened\n- stuff\n"
	if err := os.WriteFile(filepath.Join(sessions, date+".md"), []byte(withBlock), 0o644); err != nil {
		t.Fatal(err)
	}
	if !hasSessionNote(vault, date) {
		t.Error("expected true when file has a '## ' session block")
	}

	// A different date is unaffected.
	if hasSessionNote(vault, "2026-06-06") {
		t.Error("expected false for a date with no file")
	}
}

// chainForTest builds a chain directly, without touching the filesystem.
func chainForTest(t *testing.T, id vault.Identity, repoVault, rootsJSON string) vault.Chain {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "roots.json")
	if rootsJSON != "" {
		if err := os.WriteFile(path, []byte(rootsJSON), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r, err := vault.LoadRoots(path)
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	return r.Chain(id, repoVault)
}

func TestSessionNotePath_SingleRootMode(t *testing.T) {
	c := chainForTest(t, vault.Identity{Workspace: "w", Repo: "w"}, "/repo/vault", "")
	got, err := sessionNotePath(c, "2026-07-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/repo/vault", "sessions", "2026-07-25.md")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}

func TestSessionNotePath_Routed(t *testing.T) {
	c := chainForTest(t, vault.Identity{Workspace: "acme-org", Repo: "acme"}, "/repo/vault",
		`{"roots":[{"role":"org","workspace":"acme-org","path":"/other/vault"}]}`)
	got, err := sessionNotePath(c, "2026-07-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/other/vault", "sessions", "2026-07-25-acme.md")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}

func TestSessionNotePath_DeclaredRootsWithoutAPrivateRoot_Errors(t *testing.T) {
	c := chainForTest(t, vault.Identity{Workspace: "acme-org", Repo: "acme"}, "/repo/vault",
		`{"roots":[{"role":"org","workspace":"someone-else","path":"/elsewhere"}]}`)
	if _, err := sessionNotePath(c, "2026-07-25"); err == nil {
		t.Fatal("expected an error when no private root is declared for this workspace")
	}
}

// TestRunCommit_RoutedGate_LooksInTheOrgRoot verifies that when a workspace
// declares an org root the gate checks there, under the repo-suffixed name,
// rather than in the repo's own vault. A note placed only in the repo vault
// must NOT satisfy the gate.
func TestRunCommit_RoutedGate_LooksInTheOrgRoot(t *testing.T) {
	root := t.TempDir()
	pub := filepath.Join(root, "public", "vault")
	priv := filepath.Join(root, "private", "vault")
	for _, base := range []string{pub, priv} {
		if err := os.MkdirAll(filepath.Join(base, "sessions"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	if err := vault.WriteMeta(pub, vault.Meta{
		Workspace: "acme", Repo: "acme", EmbeddingModel: "nomic-embed-text",
		SchemaVersion: vault.SupportedSchemaVersion,
	}); err != nil {
		t.Fatal(err)
	}
	declareRawRoots(t, `{"roots":[{"role":"org","workspace":"acme","path":"`+priv+`"}]}`)

	const today = "2026-07-25"
	noteBody := "---\ntags: [session]\ndate: " + today + "\nstatus: active\n---\n\n## [10:00] Session — x\n"

	// A note in the PUBLIC vault (wrong place) must not satisfy the gate.
	if err := os.WriteFile(filepath.Join(pub, "sessions", today+".md"), []byte(noteBody), 0o644); err != nil {
		t.Fatal(err)
	}

	var errBuf bytes.Buffer
	err := RunCommit(CommitOpts{
		RepoRoot:  root,
		VaultPath: pub,
		Message:   "x",
		Today:     today,
		Stdout:    &bytes.Buffer{},
		Stderr:    &errBuf,
	})
	if !errors.Is(err, ErrNoSessionNote) {
		t.Fatalf("expected ErrNoSessionNote (note is in public vault, gate must look in private), got %v", err)
	}
}

// A malformed .vault-routing.local.json must fail loudly rather than quietly
// reverting to local-only routing, which would gate the commit on a session
// note in the public vault.
func TestRunCommit_BrokenRoutingConfig_Errors(t *testing.T) {
	vaultDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(vaultDir, ".vault-routing.local.json"), []byte(`{"private_vault":`), 0o644); err != nil {
		t.Fatalf("write routing config: %v", err)
	}

	err := RunCommit(CommitOpts{
		VaultPath: vaultDir,
		RepoRoot:  t.TempDir(),
		Stdout:    &bytes.Buffer{},
		Stderr:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected an error for a malformed routing config")
	}
	if errors.Is(err, ErrNoSessionNote) {
		t.Errorf("got ErrNoSessionNote, want a config error: %v", err)
	}
}
