package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluate_DependencyBucket(t *testing.T) {
	got := Evaluate([]string{"package.json", "src/main.go"}, map[string]int{
		"package.json": 5, "src/main.go": 10,
	})
	if len(got[BucketDependency]) != 1 || got[BucketDependency][0] != "package.json" {
		t.Errorf("dependency bucket = %v, want [package.json]", got[BucketDependency])
	}
	if len(got[BucketInfra]) != 0 || len(got[BucketTooling]) != 0 {
		t.Errorf("only dependency should fire; got infra=%v tooling=%v", got[BucketInfra], got[BucketTooling])
	}
}

func TestEvaluate_InfraBucket(t *testing.T) {
	got := Evaluate([]string{"Dockerfile", "infra/main.tf"}, nil)
	if len(got[BucketInfra]) != 2 {
		t.Errorf("infra bucket should match both paths: %v", got[BucketInfra])
	}
}

func TestEvaluate_ToolingBucket(t *testing.T) {
	got := Evaluate([]string{"tsconfig.json", ".eslintrc.json"}, nil)
	if len(got[BucketTooling]) != 2 {
		t.Errorf("tooling bucket should match both: %v", got[BucketTooling])
	}
}

func TestEvaluate_SubstantialDiff_BelowThreshold(t *testing.T) {
	got := Evaluate([]string{"src/a.go"}, map[string]int{"src/a.go": 500})
	if len(got[BucketSubstantial]) != 0 {
		t.Errorf("exactly 500 added should NOT fire (strictly >, not >=); got %v", got[BucketSubstantial])
	}
}

func TestEvaluate_SubstantialDiff_AboveThreshold(t *testing.T) {
	got := Evaluate([]string{"src/a.go"}, map[string]int{"src/a.go": 501})
	if len(got[BucketSubstantial]) != 1 {
		t.Errorf("501 added should fire; got %v", got[BucketSubstantial])
	}
}

func TestEvaluate_LockfileExcludedFromLOC(t *testing.T) {
	// package-lock.json gets dependency, but its 5000 lines do NOT count toward the LOC sum
	got := Evaluate(
		[]string{"package-lock.json"},
		map[string]int{"package-lock.json": 5000},
	)
	if len(got[BucketDependency]) != 1 {
		t.Errorf("dependency should fire for lockfile: %v", got[BucketDependency])
	}
	if len(got[BucketSubstantial]) != 0 {
		t.Errorf("lockfile lines should NOT count toward substantial-diff: %v", got[BucketSubstantial])
	}
}

func TestEvaluate_MultipleBuckets(t *testing.T) {
	got := Evaluate(
		[]string{"package.json", "Dockerfile", "src/a.go"},
		map[string]int{"package.json": 5, "Dockerfile": 10, "src/a.go": 600},
	)
	if len(got[BucketDependency]) == 0 {
		t.Error("dependency should fire")
	}
	if len(got[BucketInfra]) == 0 {
		t.Error("infra should fire")
	}
	if len(got[BucketSubstantial]) == 0 {
		t.Error("substantial should fire (600 added, lockfile-excluded code paths)")
	}
}

func TestEvaluate_EmptyStaged(t *testing.T) {
	got := Evaluate(nil, nil)
	for bucket := range got {
		if len(got[bucket]) > 0 {
			t.Errorf("empty staged should fire nothing; got %s: %v", bucket, got[bucket])
		}
	}
}

func TestEvaluate_GlobPattern_TerraformFiles(t *testing.T) {
	got := Evaluate([]string{"main.tf", "modules/db.tf"}, nil)
	if len(got[BucketInfra]) != 2 {
		t.Errorf("*.tf glob should match both: %v", got[BucketInfra])
	}
}

func TestEvaluate_DirPrefixPattern_HelmChart(t *testing.T) {
	got := Evaluate([]string{"helm/templates/deployment.yaml"}, nil)
	if len(got[BucketInfra]) != 1 {
		t.Errorf("helm/** dir-prefix should match: %v", got[BucketInfra])
	}
}

func TestEvaluate_DirPrefixPattern_GitHubWorkflows(t *testing.T) {
	got := Evaluate(
		[]string{".github/workflows/ci.yml", ".github/workflows/release.yml"},
		nil,
	)
	if len(got[BucketInfra]) != 2 {
		t.Errorf(".github/workflows/*.yml should match both: %v", got[BucketInfra])
	}
}

func TestFormatNudge_SingleBucket(t *testing.T) {
	out := FormatNudge(map[Bucket][]string{
		BucketDependency: {"package.json"},
	})
	if !strings.Contains(out, "changes dependencies") {
		t.Errorf("output missing dependency phrasing: %s", out)
	}
	if !strings.Contains(out, "Matched: package.json") {
		t.Errorf("output missing matched paths: %s", out)
	}
	if !strings.Contains(out, "tolvi sync decision") {
		t.Errorf("output missing tolvi sync hint: %s", out)
	}
	if !strings.Contains(out, "TOLVI_PRECOMMIT_QUIET") {
		t.Errorf("output missing silence hint: %s", out)
	}
}

func TestFormatNudge_MultipleBuckets(t *testing.T) {
	out := FormatNudge(map[Bucket][]string{
		BucketDependency: {"package.json", "package-lock.json"},
		BucketInfra:      {"Dockerfile"},
	})
	for _, want := range []string{"changes dependencies", "touches infra", "package.json, package-lock.json", "Dockerfile"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}

func TestFormatNudge_NoFires_ReturnsEmpty(t *testing.T) {
	out := FormatNudge(map[Bucket][]string{})
	if out != "" {
		t.Errorf("empty fires should produce empty string; got %q", out)
	}
}

func TestFormatNudge_OnlyZeroLengthBuckets_ReturnsEmpty(t *testing.T) {
	out := FormatNudge(map[Bucket][]string{
		BucketDependency: {}, BucketInfra: {}, BucketTooling: {}, BucketSubstantial: {},
	})
	if out != "" {
		t.Errorf("only-zero-length should produce empty; got %q", out)
	}
}

func TestIsTolviShim_Recognizes(t *testing.T) {
	shim := `#!/usr/bin/env sh
# tolvi precommit hook — installed by ` + "`tolvi precommit install`" + `.
# Exits 0 unconditionally so this never blocks commits.
command -v tolvi >/dev/null 2>&1 && tolvi precommit check || true
exit 0
`
	if !isTolviShim([]byte(shim)) {
		t.Error("canonical shim not recognized")
	}
}

func TestIsTolviShim_DoesNotFalsePositive(t *testing.T) {
	userHook := `#!/usr/bin/env bash
# my custom hook that runs go vet
# (mentions tolvi in this comment — should not match)
go vet ./...
`
	if isTolviShim([]byte(userHook)) {
		t.Error("user hook mentioning 'tolvi' in a comment should not be detected as a shim")
	}
}

func TestInstallShim_FreshInstall(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeDefault})
	if err != nil {
		t.Fatalf("InstallShim: %v", err)
	}
	if result.Action != InstallActionWrote {
		t.Errorf("Action = %v, want Wrote", result.Action)
	}
	hookPath := filepath.Join(gitDir, "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if !isTolviShim(data) {
		t.Errorf("written file is not a tolvi shim: %s", data)
	}
	info, _ := os.Stat(hookPath)
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("hook not executable: mode = %o", info.Mode().Perm())
	}
}

func TestInstallShim_IdempotentReInstall(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)

	if _, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeDefault}); err != nil {
		t.Fatalf("first install: %v", err)
	}
	result, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeDefault})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if result.Action != InstallActionAlreadyInstalled {
		t.Errorf("Action = %v, want AlreadyInstalled", result.Action)
	}
}

func TestInstallShim_RefusesExistingNonTolviHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)
	_ = os.WriteFile(filepath.Join(gitDir, "pre-commit"), []byte("#!/bin/sh\nrun-some-check\n"), 0o755)

	_, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeDefault})
	if err == nil {
		t.Fatal("expected refusal for existing non-tolvi hook")
	}
}

func TestInstallShim_Force(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)
	_ = os.WriteFile(filepath.Join(gitDir, "pre-commit"), []byte("#!/bin/sh\nrun-some-check\n"), 0o755)

	result, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeForce})
	if err != nil {
		t.Fatalf("InstallShim force: %v", err)
	}
	if result.Action != InstallActionReplaced {
		t.Errorf("Action = %v, want Replaced", result.Action)
	}
	data, _ := os.ReadFile(filepath.Join(gitDir, "pre-commit"))
	if !isTolviShim(data) {
		t.Errorf("after --force, file is not a tolvi shim: %s", data)
	}
}

func TestInstallShim_Append(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)
	userHook := "#!/bin/sh\nrun-some-check\n"
	_ = os.WriteFile(filepath.Join(gitDir, "pre-commit"), []byte(userHook), 0o755)

	result, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeAppend})
	if err != nil {
		t.Fatalf("InstallShim append: %v", err)
	}
	if result.Action != InstallActionAppended {
		t.Errorf("Action = %v, want Appended", result.Action)
	}
	data, _ := os.ReadFile(filepath.Join(gitDir, "pre-commit"))
	s := string(data)
	if !strings.Contains(s, "run-some-check") {
		t.Error("user hook content lost after --append")
	}
	if !strings.Contains(s, "tolvi precommit check") {
		t.Error("tolvi line not appended")
	}
}

func TestInstallShim_NotAGitRepo(t *testing.T) {
	dir := t.TempDir() // no .git
	_, err := InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeDefault})
	if err == nil {
		t.Fatal("expected error for missing .git")
	}
}

func TestUninstallShim_RemovesTolviShim(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)
	_, _ = InstallShim(InstallOpts{RepoRoot: dir, Mode: InstallModeDefault})

	result, err := UninstallShim(UninstallOpts{RepoRoot: dir})
	if err != nil {
		t.Fatalf("UninstallShim: %v", err)
	}
	if result.Action != UninstallActionRemoved {
		t.Errorf("Action = %v, want Removed", result.Action)
	}
	if _, err := os.Stat(filepath.Join(gitDir, "pre-commit")); !os.IsNotExist(err) {
		t.Error("hook still exists after uninstall")
	}
}

func TestUninstallShim_NoHookExists(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)

	result, err := UninstallShim(UninstallOpts{RepoRoot: dir})
	if err != nil {
		t.Fatalf("UninstallShim no-hook: %v", err)
	}
	if result.Action != UninstallActionNoOp {
		t.Errorf("Action = %v, want NoOp", result.Action)
	}
}

func TestUninstallShim_RefusesUserHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)
	_ = os.WriteFile(filepath.Join(gitDir, "pre-commit"), []byte("#!/bin/sh\nrun-check\n"), 0o755)

	_, err := UninstallShim(UninstallOpts{RepoRoot: dir})
	if err == nil {
		t.Fatal("expected refusal for non-tolvi hook")
	}
}

func TestUninstallShim_ForceRemovesUserHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	_ = os.MkdirAll(gitDir, 0o755)
	_ = os.WriteFile(filepath.Join(gitDir, "pre-commit"), []byte("#!/bin/sh\nrun-check\n"), 0o755)

	result, err := UninstallShim(UninstallOpts{RepoRoot: dir, Force: true})
	if err != nil {
		t.Fatalf("UninstallShim force: %v", err)
	}
	if result.Action != UninstallActionRemoved {
		t.Errorf("Action = %v, want Removed", result.Action)
	}
}
