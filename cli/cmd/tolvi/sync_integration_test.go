package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildToTmp builds the tolvi binary into a t.TempDir and returns its path.
// Shared by all integration tests in this package.
func buildToTmp(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "tolvi-it")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func TestIntegration_InitThenSync(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()

	// 1. tolvi init
	init := exec.Command(bin, "init", "--workspace", "it-test")
	init.Dir = work
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}

	// 2. tolvi sync pattern (uses --body, no editor invocation)
	sync := exec.Command(bin, "sync", "pattern", "Idempotent migrations",
		"--body", "# Pattern\n\nbody from integration test.\n")
	sync.Dir = work
	if out, err := sync.CombinedOutput(); err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	// 3. Verify the file was created and is well-formed.
	created := filepath.Join(work, "vault", "patterns", "idempotent-migrations.md")
	data, err := os.ReadFile(created)
	if err != nil {
		t.Fatalf("read created: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "status: active") {
		t.Errorf("missing status: %s", s)
	}
	if !strings.Contains(s, "body from integration test") {
		t.Errorf("missing body: %s", s)
	}
}

func TestIntegration_SyncWithFakeEditor(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()

	// init
	init := exec.Command(bin, "init", "--workspace", "it-editor")
	init.Dir = work
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}

	// Locate the fake-editor script relative to the test file.
	editor, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fake-editor.sh"))
	if err != nil {
		t.Fatal(err)
	}

	// sync decision via $EDITOR stub
	sync := exec.Command(bin, "sync", "decision", "Postgres choice")
	sync.Dir = work
	sync.Env = append(os.Environ(),
		"EDITOR="+editor,
		"TOLVI_TEST_BODY=# Postgres\n\nthe editor stub wrote this body.\n",
	)
	out, err := sync.CombinedOutput()
	if err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	created := filepath.Join(work, "vault", "decisions")
	entries, err := os.ReadDir(created)
	if err != nil {
		t.Fatalf("read decisions dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(entries))
	}
	data, _ := os.ReadFile(filepath.Join(created, entries[0].Name()))
	if !strings.Contains(string(data), "the editor stub wrote this body") {
		t.Errorf("editor body not captured: %s", data)
	}
}
