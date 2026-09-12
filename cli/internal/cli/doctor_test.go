package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// doctorVault builds a minimal valid vault and returns its path.
func doctorVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	v := filepath.Join(dir, "vault")
	for _, sub := range []string{"decisions", "sessions"} {
		if err := os.MkdirAll(filepath.Join(v, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	meta := `{"workspace":"test","embedding_model":"nomic-embed-text","schema_version":1}`
	if err := os.WriteFile(filepath.Join(v, ".vault-meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// healthyOpts returns options where every check should pass.
func healthyOpts(t *testing.T, out *bytes.Buffer) DoctorOpts {
	t.Helper()
	root := doctorVault(t)
	home := t.TempDir()
	writeClaudeSettings(t, home, `{"permissions":{"allow":["Bash(tolvi recall:*)","Bash(tolvi ask:*)"]}}`)
	return DoctorOpts{
		StartDir: root,
		HomeDir:  home,
		Env: func(k string) string {
			if k == "ANTHROPIC_API_KEY" {
				return "sk-ant-test"
			}
			return ""
		},
		LookPath:        func(string) (string, error) { return "/usr/local/bin/tolvi", nil },
		Stdout:          out,
		SkipVaultHealth: true,
	}
}

func writeClaudeSettings(t *testing.T, home, body string) {
	t.Helper()
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findCheck(t *testing.T, checks []Check, name string) Check {
	t.Helper()
	for _, c := range checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no check named %q in %v", name, checks)
	return Check{}
}

func TestRunDoctor_HealthySetupPassesEverything(t *testing.T) {
	var out bytes.Buffer
	checks, err := RunDoctor(healthyOpts(t, &out))
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) == 0 {
		t.Fatal("expected checks, got none")
	}
	for _, c := range checks {
		if !c.OK {
			t.Errorf("check %q failed on a healthy setup: %s", c.Name, c.Detail)
		}
		if c.OK && c.Fix != "" {
			t.Errorf("check %q passed but still carries a fix: %s", c.Name, c.Fix)
		}
	}
}

// The chicken-and-egg case: doctor runs (so a binary exists) but it was not
// reached through PATH, so every skill that shells out to `tolvi` still fails.
func TestRunDoctor_BinaryNotOnPathIsReported(t *testing.T) {
	var out bytes.Buffer
	opts := healthyOpts(t, &out)
	opts.LookPath = func(string) (string, error) { return "", errors.New("not found") }

	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}
	c := findCheck(t, checks, "tolvi on PATH")
	if c.OK {
		t.Fatal("expected the PATH check to fail")
	}
	// The known trap is `go install` succeeding into a directory PATH misses,
	// so naming only the install command is not a fix.
	if !strings.Contains(c.Fix, "PATH") {
		t.Errorf("PATH remediation must mention PATH, got %q", c.Fix)
	}
	if !strings.Contains(c.Fix, "GOPATH") && !strings.Contains(c.Fix, "go env") {
		t.Errorf("PATH remediation should point at the Go bin dir, got %q", c.Fix)
	}
}

func TestRunDoctor_MissingVaultIsReportedWithInitFix(t *testing.T) {
	var out bytes.Buffer
	opts := healthyOpts(t, &out)
	opts.StartDir = t.TempDir() // no vault anywhere above it
	opts.HomeDir = opts.StartDir

	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}
	c := findCheck(t, checks, "vault")
	if c.OK {
		t.Fatal("expected the vault check to fail")
	}
	if !strings.Contains(c.Fix, "tolvi init") {
		t.Errorf("vault remediation should suggest `tolvi init`, got %q", c.Fix)
	}
}

func TestRunDoctor_MissingAPIKeyOnlyFailsTheAskCheck(t *testing.T) {
	var out bytes.Buffer
	opts := healthyOpts(t, &out)
	opts.Env = func(string) string { return "" }

	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}
	c := findCheck(t, checks, "ANTHROPIC_API_KEY")
	if c.OK {
		t.Fatal("expected the API key check to fail")
	}
	if !strings.Contains(c.Fix, "ANTHROPIC_API_KEY") {
		t.Errorf("remediation should name the variable, got %q", c.Fix)
	}
	// The vault is still fine; a missing key must not cascade.
	if v := findCheck(t, checks, "vault"); !v.OK {
		t.Error("a missing API key should not fail the vault check")
	}
}

func TestRunDoctor_MissingClaudePermissionsIsReported(t *testing.T) {
	var out bytes.Buffer
	opts := healthyOpts(t, &out)
	writeClaudeSettings(t, opts.HomeDir, `{"permissions":{"allow":[]}}`)

	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}
	if c := findCheck(t, checks, "Claude Code permissions"); c.OK {
		t.Fatal("expected the permissions check to fail when no tolvi rules are allowed")
	}
}

func TestRunDoctor_ReportNamesEveryFailureAndItsFix(t *testing.T) {
	var out bytes.Buffer
	opts := healthyOpts(t, &out)
	opts.Env = func(string) string { return "" }
	opts.LookPath = func(string) (string, error) { return "", errors.New("nope") }

	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}
	report := out.String()
	for _, c := range checks {
		if !strings.Contains(report, c.Name) {
			t.Errorf("report omits check %q", c.Name)
		}
		if c.OK || c.Fix == "" {
			continue
		}
		// The report re-indents multi-line fixes, so assert per line.
		for _, line := range strings.Split(c.Fix, "\n") {
			if line = strings.TrimSpace(line); line != "" && !strings.Contains(report, line) {
				t.Errorf("report omits a fix line for %q: %s", c.Name, line)
			}
		}
	}
}

// Remediation must never name a command that does not exist. `tolvi install`
// is not a subcommand; the Claude Code allow rules come from the shell
// installer, and pointing at the wrong one strands the user.
func TestRunDoctor_RemediationNeverInventsASubcommand(t *testing.T) {
	var out bytes.Buffer
	opts := healthyOpts(t, &out)
	writeClaudeSettings(t, opts.HomeDir, `{"permissions":{"allow":[]}}`)
	opts.Env = func(string) string { return "" }
	opts.LookPath = func(string) (string, error) { return "", errors.New("nope") }

	real := map[string]bool{
		"ask": true, "commit": true, "doctor": true, "init": true,
		"precommit": true, "recall": true, "sync": true, "version": true,
	}
	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range checks {
		for _, line := range strings.Split(c.Fix, "\n") {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) < 2 || fields[0] != "tolvi" {
				continue
			}
			if !real[fields[1]] {
				t.Errorf("check %q suggests `tolvi %s`, which is not a subcommand", c.Name, fields[1])
			}
		}
	}
}

func TestDoctorFailures_CountsOnlyFailingChecks(t *testing.T) {
	checks := []Check{{Name: "a", OK: true}, {Name: "b"}, {Name: "c"}}
	if got := DoctorFailures(checks); got != 2 {
		t.Errorf("got %d, want 2", got)
	}
}
