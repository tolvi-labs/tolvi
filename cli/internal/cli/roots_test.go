package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRoots_ListsTheChainNearestFirst(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	org := t.TempDir()
	declareRoots(t, org)

	var out bytes.Buffer
	if err := RunRoots(RootsOpts{VaultPath: vaultDir, Stdout: &out}); err != nil {
		t.Fatalf("RunRoots: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "repo") || !strings.Contains(got, vaultDir) {
		t.Errorf("repo root missing: %s", got)
	}
	if !strings.Contains(got, "org") || !strings.Contains(got, org) {
		t.Errorf("org root missing: %s", got)
	}
	if strings.Index(got, "repo") > strings.Index(got, "org") {
		t.Errorf("chain should list nearest first: %s", got)
	}
}

func TestRunRoots_SessionNotePrintsWhereTodaysNoteLives(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	org := t.TempDir()
	declareRoots(t, org)

	var out bytes.Buffer
	err := RunRoots(RootsOpts{VaultPath: vaultDir, SessionNote: true, Today: "2026-09-13", Stdout: &out})
	if err != nil {
		t.Fatalf("RunRoots: %v", err)
	}
	want := filepath.Join(org, "sessions", "2026-09-13-test.md")
	if got := strings.TrimSpace(out.String()); got != want {
		t.Errorf("session note path = %q, want %q", got, want)
	}
}

func TestRunRoots_SessionNoteInSingleRootMode(t *testing.T) {
	vaultDir := mkVaultForTest(t) // no roots.json declared

	var out bytes.Buffer
	if err := RunRoots(RootsOpts{VaultPath: vaultDir, SessionNote: true, Today: "2026-09-13", Stdout: &out}); err != nil {
		t.Fatalf("RunRoots: %v", err)
	}
	want := filepath.Join(vaultDir, "sessions", "2026-09-13.md")
	if got := strings.TrimSpace(out.String()); got != want {
		t.Errorf("session note path = %q, want %q", got, want)
	}
}
