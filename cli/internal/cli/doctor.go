package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// Check is one setup assertion and, when it fails, the command that fixes it.
type Check struct {
	Name   string
	OK     bool
	Detail string // what was actually found
	Fix    string // remediation; empty when OK
}

// DoctorOpts carries the ambient state doctor inspects. Every source of
// truth is injectable so the checks are testable without touching the
// developer's real machine.
type DoctorOpts struct {
	StartDir      string
	HomeDir       string
	ExplicitVault string
	Env           func(string) string
	LookPath      func(string) (string, error)
	Stdout        io.Writer

	// Only scopes the run to a single section. "" runs everything;
	// "vault-health" runs only the content scan.
	Only string

	// SkipVaultHealth suppresses the content scan (used by tests that only
	// care about setup checks).
	SkipVaultHealth bool
}

// goBinHint is the remediation for a binary that exists but is unreachable.
// `go install` putting the binary somewhere PATH does not cover is the most
// common way a tolvi install looks successful and is not, so the fix names
// the PATH export and not just the install command.
// claudeInstallHint points at the repo's Claude Code installer. There is no
// top-level `tolvi install` subcommand; the allow rules are merged by the
// shell installer, so the remediation has to name it.
const claudeInstallHint = "bash skills/tolvi/install.sh --with-hooks   # from a tolvi checkout"

const goBinHint = `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
     export PATH="$PATH:$(go env GOPATH)/bin"   # add to ~/.zshenv, not ~/.zshrc`

// RunDoctor inspects the local setup, writes a report, and returns the
// checks it ran. A failing check is reported, never fatal: the caller
// decides the exit code via DoctorFailures.
func RunDoctor(opts DoctorOpts) ([]Check, error) {
	if opts.Env == nil {
		opts.Env = os.Getenv
	}
	var checks []Check
	if opts.Only != "vault-health" {
		checks = []Check{
			checkPath(opts),
			checkVault(opts),
			checkAPIKey(opts),
			checkClaudePermissions(opts),
		}
		if err := writeDoctorReport(opts.Stdout, checks); err != nil {
			return checks, err
		}
	}
	if !opts.SkipVaultHealth {
		if err := runDoctorVaultHealth(opts); err != nil {
			return checks, err
		}
	}
	return checks, nil
}

// runDoctorVaultHealth appends the content scan. It is skipped silently when
// no vault resolves: "no vault here" is already reported by the vault check,
// and repeating it as a health failure would be noise.
func runDoctorVaultHealth(opts DoctorOpts) error {
	path, err := vault.Discover(vault.DiscoverOpts{
		StartDir:     opts.StartDir,
		HomeDir:      opts.HomeDir,
		ExplicitPath: opts.ExplicitVault,
	})
	if err != nil {
		return nil
	}
	rep, err := RunVaultHealth(path)
	if err != nil {
		return nil
	}
	return RenderVaultHealth(opts.Stdout, rep)
}

// DoctorVaultHealth runs only the content scan and returns its report, for
// callers that need the findings rather than the rendered text.
func DoctorVaultHealth(opts DoctorOpts) (HealthReport, error) {
	path, err := vault.Discover(vault.DiscoverOpts{
		StartDir:     opts.StartDir,
		HomeDir:      opts.HomeDir,
		ExplicitPath: opts.ExplicitVault,
	})
	if err != nil {
		return HealthReport{}, err
	}
	return RunVaultHealth(path)
}

// HealthHighSeverity counts findings that are outright defects rather than
// advisories. `tolvi doctor vault-health` exits non-zero on these.
func HealthHighSeverity(rep HealthReport) int {
	n := 0
	for _, f := range rep.Findings {
		if f.Severity == "high" {
			n++
		}
	}
	return n
}

// DoctorFailures counts the checks that did not pass.
func DoctorFailures(checks []Check) int {
	n := 0
	for _, c := range checks {
		if !c.OK {
			n++
		}
	}
	return n
}

// checkPath asks whether `tolvi` is reachable by name. Running doctor proves
// a binary exists; it does not prove PATH can find it, and every skill and
// hook that shells out to `tolvi` depends on the latter.
func checkPath(opts DoctorOpts) Check {
	lookPath := opts.LookPath
	if lookPath == nil {
		return Check{Name: "tolvi on PATH", OK: true, Detail: "not checked"}
	}
	found, err := lookPath("tolvi")
	if err != nil || found == "" {
		return Check{
			Name:   "tolvi on PATH",
			Detail: "not reachable as `tolvi`; skills and hooks that shell out will fall back",
			Fix:    goBinHint,
		}
	}
	return Check{Name: "tolvi on PATH", OK: true, Detail: found}
}

func checkVault(opts DoctorOpts) Check {
	path, err := vault.Discover(vault.DiscoverOpts{
		StartDir:     opts.StartDir,
		HomeDir:      opts.HomeDir,
		ExplicitPath: opts.ExplicitVault,
	})
	if err != nil {
		return Check{
			Name:   "vault",
			Detail: "no vault found from here",
			Fix:    "tolvi init        # or run from inside a repo that has one",
		}
	}
	if _, err := vault.ReadMeta(path); err != nil {
		return Check{
			Name:   "vault",
			Detail: fmt.Sprintf("%s: %v", path, err),
			Fix:    "tolvi init --repair",
		}
	}
	return Check{Name: "vault", OK: true, Detail: path}
}

func checkAPIKey(opts DoctorOpts) Check {
	if key := strings.TrimSpace(opts.Env("ANTHROPIC_API_KEY")); key != "" {
		return Check{Name: "ANTHROPIC_API_KEY", OK: true, Detail: "set"}
	}
	return Check{
		Name:   "ANTHROPIC_API_KEY",
		Detail: "unset; `tolvi ask` is unavailable (the vault itself still works)",
		Fix:    `export ANTHROPIC_API_KEY=sk-ant-...   # add to ~/.zshrc or ~/.bashrc`,
	}
}

// checkClaudePermissions looks for the read-only tolvi allow rules. Without
// them every recall raises a permission prompt, which is the friction that
// pushes people onto the fallback path in the first place.
func checkClaudePermissions(opts DoctorOpts) Check {
	const name = "Claude Code permissions"
	path := filepath.Join(opts.HomeDir, ".claude", "settings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Check{
			Name:   name,
			Detail: "no ~/.claude/settings.json",
			Fix:    claudeInstallHint,
		}
	}
	var settings struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return Check{Name: name, Detail: fmt.Sprintf("%s: %v", path, err), Fix: "fix the JSON, then re-run"}
	}
	var missing []string
	for _, want := range []string{"Bash(tolvi recall:*)", "Bash(tolvi ask:*)"} {
		found := false
		for _, have := range settings.Permissions.Allow {
			if have == want {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, want)
		}
	}
	if len(missing) > 0 {
		return Check{
			Name:   name,
			Detail: "missing allow rules: " + strings.Join(missing, ", "),
			Fix:    claudeInstallHint,
		}
	}
	return Check{Name: name, OK: true, Detail: "recall and ask allowlisted"}
}

func writeDoctorReport(w io.Writer, checks []Check) error {
	if w == nil {
		return nil
	}
	var b strings.Builder
	b.WriteString("\ntolvi doctor\n\n")
	width := 0
	for _, c := range checks {
		if len(c.Name) > width {
			width = len(c.Name)
		}
	}
	for _, c := range checks {
		mark := "x"
		if c.OK {
			mark = "ok"
		}
		fmt.Fprintf(&b, "  %-*s  %-3s %s\n", width, c.Name, mark, c.Detail)
		if !c.OK && c.Fix != "" {
			for _, line := range strings.Split(c.Fix, "\n") {
				fmt.Fprintf(&b, "  %-*s       %s\n", width, "", strings.TrimRight(line, " "))
			}
		}
	}
	failed := DoctorFailures(checks)
	switch failed {
	case 0:
		b.WriteString("\n  Everything checks out.\n\n")
	case 1:
		b.WriteString("\n  1 problem, fixable above.\n\n")
	default:
		fmt.Fprintf(&b, "\n  %d problems, all fixable above.\n\n", failed)
	}
	_, err := io.WriteString(w, b.String())
	return err
}
