package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_HappyPath(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := RunInit(InitOpts{Cwd: dir, Workspace: "my-project", Stdout: &out})
	if err != nil {
		t.Fatalf("RunInit: %v", err)
	}

	for _, sub := range []string{"decisions", "sessions", "patterns"} {
		info, err := os.Stat(filepath.Join(dir, "vault", sub))
		if err != nil {
			t.Errorf("missing %s/: %v", sub, err)
		} else if !info.IsDir() {
			t.Errorf("%s/ is not a directory", sub)
		}
	}

	metaPath := filepath.Join(dir, "vault", ".vault-meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	if !bytes.Contains(data, []byte(`"workspace": "my-project"`)) {
		t.Errorf("meta missing workspace: %s", data)
	}
}

func TestRunInit_RefusesIfAlreadyVault(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := RunInit(InitOpts{Cwd: dir, Workspace: "first", Stdout: &out}); err != nil {
		t.Fatalf("first RunInit: %v", err)
	}
	err := RunInit(InitOpts{Cwd: dir, Workspace: "second", Stdout: &out})
	if err == nil {
		t.Fatal("expected error on re-init")
	}
}

func TestRunInit_WorkspaceDefaultFromGit(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0o755)
	gitConfig := `[remote "origin"]
	url = git@github.com:tolvi-labs/example-repo.git
`
	_ = os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0o644)

	var out bytes.Buffer
	if err := RunInit(InitOpts{Cwd: dir, Stdout: &out}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "vault", ".vault-meta.json"))
	if !bytes.Contains(data, []byte(`"workspace": "example-repo"`)) {
		t.Errorf("workspace default not derived from git: %s", data)
	}
}

func TestRunInit_WorkspaceFallbackToBasename(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "my-folder-name")
	_ = os.MkdirAll(dir, 0o755)

	var out bytes.Buffer
	if err := RunInit(InitOpts{Cwd: dir, Stdout: &out}); err != nil {
		t.Fatalf("RunInit: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "vault", ".vault-meta.json"))
	if !bytes.Contains(data, []byte(`"workspace": "my-folder-name"`)) {
		t.Errorf("workspace fallback to basename failed: %s", data)
	}
}
