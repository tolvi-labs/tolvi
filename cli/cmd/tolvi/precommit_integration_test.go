package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGit runs a git command in dir and returns combined output. Fails the
// test on non-zero exit.
func runGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Suppress git's complaints about missing user.name / user.email in CI.
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return out
}

func TestIntegration_Precommit_FreshInstall(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	out, err := exec.Command(bin, "precommit", "install", "--repo", work).CombinedOutput()
	if err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}
	hookPath := filepath.Join(work, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("hook not executable: mode = %o", info.Mode().Perm())
	}
	body, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(body), "tolvi precommit check") {
		t.Errorf("hook missing tolvi check: %s", body)
	}
}

func TestIntegration_Precommit_Idempotent(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	if out, err := exec.Command(bin, "precommit", "install", "--repo", work).CombinedOutput(); err != nil {
		t.Fatalf("first install: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "precommit", "install", "--repo", work).CombinedOutput()
	if err != nil {
		t.Fatalf("second install: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Already installed") {
		t.Errorf("second install should be idempotent: %s", out)
	}
}

func TestIntegration_Precommit_RefusesExistingUserHook(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	hookPath := filepath.Join(work, ".git", "hooks", "pre-commit")
	_ = os.WriteFile(hookPath, []byte("#!/bin/sh\necho user-hook\n"), 0o755)

	cmd := exec.Command(bin, "precommit", "install", "--repo", work)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit; output: %s", out)
	}
	if !strings.Contains(string(out), "not a tolvi shim") {
		t.Errorf("error message missing 'not a tolvi shim': %s", out)
	}
}

func TestIntegration_Precommit_ForceReplacesUserHook(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	hookPath := filepath.Join(work, ".git", "hooks", "pre-commit")
	_ = os.WriteFile(hookPath, []byte("#!/bin/sh\necho user-hook\n"), 0o755)

	out, err := exec.Command(bin, "precommit", "install", "--force", "--repo", work).CombinedOutput()
	if err != nil {
		t.Fatalf("--force install: %v\n%s", err, out)
	}
	body, _ := os.ReadFile(hookPath)
	if strings.Contains(string(body), "user-hook") {
		t.Errorf("--force should have replaced user hook; body: %s", body)
	}
	if !strings.Contains(string(body), "tolvi precommit check") {
		t.Errorf("--force result is not a tolvi shim: %s", body)
	}
}

func TestIntegration_Precommit_AppendPreservesUserHook(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	hookPath := filepath.Join(work, ".git", "hooks", "pre-commit")
	_ = os.WriteFile(hookPath, []byte("#!/bin/sh\necho user-hook\n"), 0o755)

	out, err := exec.Command(bin, "precommit", "install", "--append", "--repo", work).CombinedOutput()
	if err != nil {
		t.Fatalf("--append install: %v\n%s", err, out)
	}
	body, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(body), "user-hook") {
		t.Errorf("--append should preserve user hook content; body: %s", body)
	}
	if !strings.Contains(string(body), "tolvi precommit check") {
		t.Errorf("--append should add tolvi check; body: %s", body)
	}
}

func TestIntegration_Precommit_Check_FiresDependencyBucket(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	// Create and stage a package.json — fires dependency bucket.
	_ = os.WriteFile(filepath.Join(work, "package.json"), []byte(`{"name":"x"}`+"\n"), 0o644)
	runGit(t, work, "add", "package.json")

	out, err := exec.Command(bin, "precommit", "check", "--repo", work).CombinedOutput()
	if err != nil {
		t.Fatalf("check: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "changes dependencies") {
		t.Errorf("check output missing dependency nudge: %s", s)
	}
	if !strings.Contains(s, "package.json") {
		t.Errorf("check output missing matched path: %s", s)
	}
}

func TestIntegration_Precommit_Check_QuietEnvSilences(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")
	_ = os.WriteFile(filepath.Join(work, "package.json"), []byte(`{"name":"x"}`+"\n"), 0o644)
	runGit(t, work, "add", "package.json")

	cmd := exec.Command(bin, "precommit", "check", "--repo", work)
	cmd.Env = append(os.Environ(), "TOLVI_PRECOMMIT_QUIET=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check quiet: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("quiet should produce empty output; got: %q", out)
	}
}

// symlinkBinAsTolvi creates a "tolvi" symlink → bin in a fresh temp dir
// and returns the temp dir. Use this when a test needs the binary to be
// invokable as the literal name "tolvi" (e.g., for hook scripts that
// call `command -v tolvi`). The default `buildToTmp` names the binary
// "tolvi-it" which doesn't satisfy that literal lookup.
func symlinkBinAsTolvi(t *testing.T, bin string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Symlink(bin, filepath.Join(dir, "tolvi")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	return dir
}

func TestIntegration_Precommit_InstalledHookInvocation(t *testing.T) {
	// This test invokes the installed pre-commit hook script directly,
	// asserting that (1) it calls `tolvi precommit check` correctly via
	// the PATH-resolved binary, (2) the nudge is printed to stderr, and
	// (3) the script exits 0 (never blocks).
	//
	// We don't drive this via `git commit` because git captures hook
	// stdout/stderr in implementation-specific ways across versions and
	// platforms; testing the shim directly is more reliable and exercises
	// the same code path the user sees.
	bin := buildToTmp(t)
	pathDir := symlinkBinAsTolvi(t, bin) // make `tolvi` resolve to our binary
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	if out, err := exec.Command(bin, "precommit", "install", "--repo", work).CombinedOutput(); err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}

	_ = os.WriteFile(filepath.Join(work, "package.json"), []byte(`{"name":"x"}`+"\n"), 0o644)
	runGit(t, work, "add", "package.json")

	// Invoke the installed hook script directly with tolvi on PATH.
	hookPath := filepath.Join(work, ".git", "hooks", "pre-commit")
	hookCmd := exec.Command(hookPath)
	hookCmd.Dir = work
	hookCmd.Env = append(os.Environ(),
		"PATH="+pathDir+":"+os.Getenv("PATH"),
	)
	out, err := hookCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook should exit 0 (never blocks); got err=%v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "changes dependencies") {
		t.Errorf("hook nudge missing from output:\n%s", s)
	}
	if !strings.Contains(s, "package.json") {
		t.Errorf("hook output should list matched path:\n%s", s)
	}
}

func TestIntegration_Precommit_Uninstall(t *testing.T) {
	bin := buildToTmp(t)
	work := t.TempDir()
	runGit(t, work, "init", "-q")

	if out, err := exec.Command(bin, "precommit", "install", "--repo", work).CombinedOutput(); err != nil {
		t.Fatalf("install: %v\n%s", err, out)
	}
	hookPath := filepath.Join(work, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err != nil {
		t.Fatalf("hook not installed: %v", err)
	}

	if out, err := exec.Command(bin, "precommit", "uninstall", "--repo", work).CombinedOutput(); err != nil {
		t.Fatalf("uninstall: %v\n%s", err, out)
	}
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Errorf("hook still exists after uninstall: %v", err)
	}
}
