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

func TestSessionNotePath_Local(t *testing.T) {
	got, err := sessionNotePath("/repo/vault", vault.Meta{Workspace: "w"}, "2026-07-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/repo/vault", "sessions", "2026-07-25.md")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}

func TestSessionNotePath_Public(t *testing.T) {
	m := vault.Meta{Workspace: "acme", Visibility: "public", PrivateVault: "/other/vault"}
	got, err := sessionNotePath("/repo/vault", m, "2026-07-25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/other/vault", "sessions", "2026-07-25-acme.md")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}

func TestSessionNotePath_PublicNoPrivateVault_Errors(t *testing.T) {
	m := vault.Meta{Workspace: "acme", Visibility: "public"}
	if _, err := sessionNotePath("/repo/vault", m, "2026-07-25"); err == nil {
		t.Fatal("expected error when public but private_vault unset")
	}
}

// TestRunCommit_PublicGate_LooksInPrivateVault verifies that under public
// visibility the gate checks the private vault (workspace-suffixed name),
// not the local vault. A note placed only in the public vault must NOT
// satisfy the gate.
func TestRunCommit_PublicGate_LooksInPrivateVault(t *testing.T) {
	root := t.TempDir()
	pub := filepath.Join(root, "public", "vault")
	priv := filepath.Join(root, "private", "vault")
	for _, base := range []string{pub, priv} {
		if err := os.MkdirAll(filepath.Join(base, "sessions"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := vault.WriteMeta(pub, vault.Meta{
		Workspace: "acme", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1,
		Visibility: "public", PrivateVault: priv,
	}); err != nil {
		t.Fatal(err)
	}

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
