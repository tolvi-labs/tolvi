package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// exitCode extracts the process exit code from an *exec.ExitError, or -1.
func exitCode(err error) int {
	var ee *exec.ExitError
	if as, ok := err.(*exec.ExitError); ok {
		ee = as
		return ee.ExitCode()
	}
	return -1
}

func TestIntegration_Commit_GateThenCommit(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	// Provision a vault inside the repo.
	initCmd := exec.Command(bin, "init", "--workspace", "commit-it")
	initCmd.Dir = work
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}

	gitEnv := append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test",
	)
	today := time.Now().Format("2006-01-02")

	// Stage some work.
	if err := os.WriteFile(filepath.Join(work, "foo.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", "foo.txt")

	// --- Negative: no session note → refuse with ExitVaultState (3). ---
	no := exec.Command(bin, "commit", "-m", "should fail")
	no.Dir, no.Env = work, gitEnv
	out, err := no.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure without a session note, got success:\n%s", out)
	}
	if ec := exitCode(err); ec != 3 {
		t.Errorf("expected exit 3 (ExitVaultState), got %d\n%s", ec, out)
	}
	if !strings.Contains(string(out), "no session note") {
		t.Errorf("missing guidance in output:\n%s", out)
	}

	// --- Positive: write a session note for today, then commit succeeds. ---
	sess := filepath.Join(work, "vault", "sessions", today+".md")
	body := "---\ntags: [session]\ndate: " + today + "\nstatus: active\n---\n\n" +
		"## [10:00] Session — test\n\n### What happened\n- did it\n"
	if err := os.WriteFile(sess, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	ok := exec.Command(bin, "commit", "-m", "[T-1] add foo + session")
	ok.Dir, ok.Env = work, gitEnv
	if out, err := ok.CombinedOutput(); err != nil {
		t.Fatalf("commit: %v\n%s", err, out)
	}

	if log := runGit(t, work, "log", "-1", "--pretty=%s"); !strings.Contains(string(log), "add foo + session") {
		t.Errorf("commit subject wrong: %s", log)
	}
	files := string(runGit(t, work, "show", "--name-only", "--pretty=format:"))
	if !strings.Contains(files, "vault/sessions/"+today+".md") {
		t.Errorf("vault note auto-staged but not in commit:\n%s", files)
	}
	if !strings.Contains(files, "foo.txt") {
		t.Errorf("pre-staged work not in commit:\n%s", files)
	}
}
