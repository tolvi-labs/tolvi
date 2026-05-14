package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
	vaultDir := filepath.Join(root, "vault")
	for _, sub := range []string{"decisions", "sessions", "patterns"} {
		_ = os.MkdirAll(filepath.Join(vaultDir, sub), 0o755)
	}
	_ = vault.WriteMeta(vaultDir, vault.Meta{
		Workspace: "test", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1,
	})
	return vaultDir
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
