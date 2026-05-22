package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Bucket is one of the four heuristic categories.
type Bucket string

const (
	BucketDependency  Bucket = "dependency"
	BucketInfra       Bucket = "infra"
	BucketTooling     Bucket = "tooling"
	BucketSubstantial Bucket = "substantial"
)

// substantialDiffThreshold is the strict-greater-than line count that
// fires the "substantial diff" bucket. Excludes lockfiles (which match
// the dependency bucket instead) to prevent double-nudging on
// `npm install`-style commits.
const substantialDiffThreshold = 500

// patternKind tags a pattern's matching strategy.
type patternKind int

const (
	patternExact     patternKind = iota // exact basename match (e.g. "Dockerfile")
	patternGlob                         // basename glob (e.g. "*.tf"); uses filepath.Match
	patternDirPrefix                    // dir prefix; pattern ends with "/**"; matches any path under that dir
	patternGlobPath                     // path glob (single-level, e.g. ".github/workflows/*.yml"); matches full path
)

type pattern struct {
	kind patternKind
	raw  string
}

// matchPath returns true if path matches the pattern according to the
// pattern's kind. path is always repo-relative, forward-slash form.
func (p pattern) matchPath(path string) bool {
	base := filepath.Base(path)
	switch p.kind {
	case patternExact:
		return base == p.raw
	case patternGlob:
		ok, err := filepath.Match(p.raw, base)
		return err == nil && ok
	case patternDirPrefix:
		// p.raw is "<dir>/**"; strip "**" to get "<dir>/" prefix.
		prefix := strings.TrimSuffix(p.raw, "**")
		return strings.HasPrefix(path, prefix)
	case patternGlobPath:
		// Multi-segment glob like ".github/workflows/*.yml".
		// Split into dir + glob, check prefix + last-segment match.
		dir, glob := filepath.Split(p.raw)
		if !strings.HasPrefix(path, dir) {
			return false
		}
		rest := path[len(dir):]
		if strings.Contains(rest, "/") {
			return false // single-level only
		}
		ok, err := filepath.Match(glob, rest)
		return err == nil && ok
	}
	return false
}

// lockfileNames is the subset of dependency patterns whose line
// counts should be EXCLUDED from the substantial-diff LOC sum.
var lockfileNames = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"Cargo.lock":        true,
	"go.sum":            true,
	"poetry.lock":       true,
	"Pipfile.lock":      true,
	"Gemfile.lock":      true,
}

// defaultPatterns is the fixed v1 heuristic set. Tuneability via config
// is a v1.x add (see spec §10).
var defaultPatterns = map[Bucket][]pattern{
	BucketDependency: {
		{patternExact, "package.json"},
		{patternExact, "package-lock.json"},
		{patternExact, "yarn.lock"},
		{patternExact, "pnpm-lock.yaml"},
		{patternExact, "Cargo.toml"},
		{patternExact, "Cargo.lock"},
		{patternExact, "go.mod"},
		{patternExact, "go.sum"},
		{patternExact, "requirements.txt"},
		{patternExact, "pyproject.toml"},
		{patternExact, "poetry.lock"},
		{patternExact, "Pipfile"},
		{patternExact, "Pipfile.lock"},
		{patternExact, "Gemfile"},
		{patternExact, "Gemfile.lock"},
	},
	BucketInfra: {
		{patternExact, "Dockerfile"},
		{patternGlob, "docker-compose*.yml"},
		{patternGlob, "docker-compose*.yaml"},
		{patternGlob, "*.tf"},
		{patternGlob, "*.tfvars"},
		{patternDirPrefix, "helm/**"},
		{patternDirPrefix, "k8s/**"},
		{patternDirPrefix, "kubernetes/**"},
		{patternGlobPath, ".github/workflows/*.yml"},
		{patternGlobPath, ".github/workflows/*.yaml"},
		{patternExact, "Procfile"},
		{patternExact, "fly.toml"},
		{patternExact, "vercel.json"},
		{patternExact, "netlify.toml"},
	},
	BucketTooling: {
		{patternExact, "tsconfig.json"},
		{patternGlob, "babel.config.*"},
		{patternGlob, ".eslintrc*"},
		{patternGlob, ".prettierrc*"},
		{patternGlob, "prettier.config.*"},
		{patternGlob, "webpack.config.*"},
		{patternGlob, "vite.config.*"},
		{patternGlob, "rollup.config.*"},
		{patternGlob, "tsup.config.*"},
		{patternExact, ".editorconfig"},
	},
}

// Evaluate inspects the staged paths and per-path added-line counts
// and returns which buckets fired and which paths matched each bucket.
// Buckets that don't fire are present in the returned map with an empty
// slice — callers must tolerate both (consistency with the map's
// initialization shape).
//
// addedLines may be nil; substantial-diff evaluation simply skips when
// no counts are provided.
func Evaluate(stagedPaths []string, addedLines map[string]int) map[Bucket][]string {
	fired := map[Bucket][]string{
		BucketDependency:  nil,
		BucketInfra:       nil,
		BucketTooling:     nil,
		BucketSubstantial: nil,
	}

	for _, p := range stagedPaths {
		for bucket, pats := range defaultPatterns {
			for _, pat := range pats {
				if pat.matchPath(p) {
					fired[bucket] = append(fired[bucket], p)
					break
				}
			}
		}
	}

	// Substantial-diff: sum addedLines EXCLUDING any path whose basename
	// is in lockfileNames.
	if addedLines != nil {
		var totalAdded int
		var substantialPaths []string
		for _, p := range stagedPaths {
			if lockfileNames[filepath.Base(p)] {
				continue
			}
			added, ok := addedLines[p]
			if !ok {
				continue
			}
			totalAdded += added
			substantialPaths = append(substantialPaths, p)
		}
		if totalAdded > substantialDiffThreshold {
			fired[BucketSubstantial] = substantialPaths
		}
	}

	return fired
}

// FormatNudge turns the evaluated buckets into the user-facing nudge
// text. Returns empty string when no bucket fired (caller can check
// `out != ""` before printing).
func FormatNudge(fired map[Bucket][]string) string {
	type entry struct {
		bucket  Bucket
		phrase  string
		matched []string
	}
	bucketPhrases := map[Bucket]string{
		BucketDependency:  "This commit changes dependencies — worth capturing the *why* as a decision?",
		BucketInfra:       "This commit touches infra — worth a decision capturing the choice?",
		BucketTooling:     "This commit changes tooling config — record the reasoning?",
		BucketSubstantial: "This is a substantial change — capture context as a session log?",
	}
	// Deterministic order: dependency → infra → tooling → substantial.
	order := []Bucket{BucketDependency, BucketInfra, BucketTooling, BucketSubstantial}

	var entries []entry
	for _, b := range order {
		if matched, ok := fired[b]; ok && len(matched) > 0 {
			entries = append(entries, entry{b, bucketPhrases[b], matched})
		}
	}
	if len(entries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("tolvi:\n")
	for _, e := range entries {
		fmt.Fprintf(&sb, "  ▲ %s\n", e.phrase)
		fmt.Fprintf(&sb, "    Matched: %s\n", strings.Join(e.matched, ", "))
	}
	sb.WriteString("\n  Run: tolvi sync decision \"<title>\"\n")
	sb.WriteString("  Silence: TOLVI_PRECOMMIT_QUIET=1 (env, once or in your shell rc)\n")
	sb.WriteString("          or `tolvi precommit uninstall` (forever)\n")
	return sb.String()
}

// shimContent is the canonical hook file we write.
const shimContent = `#!/usr/bin/env sh
# tolvi precommit hook — installed by ` + "`tolvi precommit install`" + `.
# Exits 0 unconditionally so this never blocks commits.
command -v tolvi >/dev/null 2>&1 && tolvi precommit check || true
exit 0
`

// shimHeaderMarker is the substring used to detect "this file is our shim".
// Picked to be specific enough to avoid false positives on user hooks.
const shimHeaderMarker = "# tolvi precommit hook — installed by"

// isTolviShim returns true if the given file content is a tolvi-installed
// pre-commit shim. Detection is via the canonical header comment.
func isTolviShim(content []byte) bool {
	// Look for the header marker within the first 200 bytes only —
	// avoids false-positives on user hooks that mention "tolvi" deep
	// in the body.
	head := content
	if len(head) > 200 {
		head = head[:200]
	}
	return strings.Contains(string(head), shimHeaderMarker)
}

// InstallMode selects install behavior for an existing hook.
type InstallMode int

const (
	InstallModeDefault InstallMode = iota // refuse if non-tolvi hook exists
	InstallModeForce                      // overwrite existing
	InstallModeAppend                     // chain after existing
)

// InstallOpts is the input shape for InstallShim.
type InstallOpts struct {
	RepoRoot string
	Mode     InstallMode
}

// InstallAction is the result-side enum returned by InstallShim.
type InstallAction int

const (
	InstallActionWrote            InstallAction = iota // fresh install
	InstallActionAlreadyInstalled                      // idempotent re-install of an existing tolvi shim
	InstallActionReplaced                              // --force overwrite of an existing non-tolvi hook
	InstallActionAppended                              // --append chained after existing non-tolvi hook
)

// InstallResult carries the install outcome plus diagnostic info for the caller to print.
type InstallResult struct {
	Action      InstallAction
	HookPath    string
	PrevContent []byte // populated when Action == Replaced or Appended
}

// InstallShim writes the tolvi pre-commit shim into <RepoRoot>/.git/hooks/pre-commit
// per the rules in the spec's install table.
func InstallShim(opts InstallOpts) (InstallResult, error) {
	gitHooks := filepath.Join(opts.RepoRoot, ".git", "hooks")
	if _, err := os.Stat(filepath.Join(opts.RepoRoot, ".git")); err != nil {
		return InstallResult{}, fmt.Errorf("not a git repository: %s", opts.RepoRoot)
	}
	if err := os.MkdirAll(gitHooks, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create .git/hooks: %w", err)
	}
	hookPath := filepath.Join(gitHooks, "pre-commit")

	existing, statErr := os.ReadFile(hookPath)
	hookExists := statErr == nil

	if hookExists && isTolviShim(existing) && opts.Mode != InstallModeAppend {
		return InstallResult{Action: InstallActionAlreadyInstalled, HookPath: hookPath}, nil
	}

	switch opts.Mode {
	case InstallModeDefault:
		if hookExists {
			return InstallResult{}, fmt.Errorf(
				"existing hook at %s is not a tolvi shim. Pass --force to replace or --append to chain after it.",
				hookPath,
			)
		}
		if err := os.WriteFile(hookPath, []byte(shimContent), 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("write %s: %w", hookPath, err)
		}
		return InstallResult{Action: InstallActionWrote, HookPath: hookPath}, nil

	case InstallModeForce:
		if err := os.WriteFile(hookPath, []byte(shimContent), 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("write %s: %w", hookPath, err)
		}
		return InstallResult{Action: InstallActionReplaced, HookPath: hookPath, PrevContent: existing}, nil

	case InstallModeAppend:
		// Append the tolvi check line (not the full shim) to the existing hook.
		// If no hook exists, behave like default install.
		if !hookExists {
			if err := os.WriteFile(hookPath, []byte(shimContent), 0o755); err != nil {
				return InstallResult{}, fmt.Errorf("write %s: %w", hookPath, err)
			}
			return InstallResult{Action: InstallActionWrote, HookPath: hookPath}, nil
		}
		appended := string(existing)
		if !strings.HasSuffix(appended, "\n") {
			appended += "\n"
		}
		appended += "\n" + shimHeaderMarker + " `tolvi precommit install --append`.\n"
		appended += "command -v tolvi >/dev/null 2>&1 && tolvi precommit check || true\n"
		if err := os.WriteFile(hookPath, []byte(appended), 0o755); err != nil {
			return InstallResult{}, fmt.Errorf("write %s: %w", hookPath, err)
		}
		return InstallResult{Action: InstallActionAppended, HookPath: hookPath, PrevContent: existing}, nil
	}
	return InstallResult{}, fmt.Errorf("internal: unknown install mode %d", opts.Mode)
}

// UninstallOpts is the input shape for UninstallShim.
type UninstallOpts struct {
	RepoRoot string
	Force    bool
}

// UninstallAction is the result-side enum returned by UninstallShim.
type UninstallAction int

const (
	UninstallActionRemoved UninstallAction = iota
	UninstallActionNoOp                    // hook didn't exist
)

// UninstallResult carries the uninstall outcome.
type UninstallResult struct {
	Action   UninstallAction
	HookPath string
}

// UninstallShim removes <RepoRoot>/.git/hooks/pre-commit if it's a tolvi
// shim (or if Force is true). Idempotent.
func UninstallShim(opts UninstallOpts) (UninstallResult, error) {
	hookPath := filepath.Join(opts.RepoRoot, ".git", "hooks", "pre-commit")
	existing, statErr := os.ReadFile(hookPath)
	if statErr != nil {
		return UninstallResult{Action: UninstallActionNoOp, HookPath: hookPath}, nil
	}
	if !isTolviShim(existing) && !opts.Force {
		return UninstallResult{}, fmt.Errorf(
			"hook at %s was not installed by tolvi. Pass --force to remove anyway, or remove it manually.",
			hookPath,
		)
	}
	if err := os.Remove(hookPath); err != nil {
		return UninstallResult{}, fmt.Errorf("remove %s: %w", hookPath, err)
	}
	return UninstallResult{Action: UninstallActionRemoved, HookPath: hookPath}, nil
}

// CheckOpts is the input shape for RunCheck.
type CheckOpts struct {
	RepoRoot string
	Stderr   io.Writer
	Quiet    bool // honors TOLVI_PRECOMMIT_QUIET
}

// RunCheck is the entry point invoked by the installed hook. Reads
// staged paths and added-line counts via `git diff --cached`, runs the
// heuristic Evaluate + FormatNudge, prints to opts.Stderr if any bucket
// fires.
//
// CRITICAL: always returns nil. Internal errors are swallowed and exit
// code is 0; the hook script's `|| true` is the second line of defense.
// The caller (the cobra subcommand) should NOT propagate this error to
// os.Exit.
func RunCheck(opts CheckOpts) error {
	if opts.Quiet {
		return nil
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	paths, err := gitStagedPaths(opts.RepoRoot)
	if err != nil || len(paths) == 0 {
		return nil
	}
	lines, _ := gitStagedAddedLines(opts.RepoRoot)

	fired := Evaluate(paths, lines)
	nudge := FormatNudge(fired)
	if nudge == "" {
		return nil
	}
	_, _ = fmt.Fprint(opts.Stderr, nudge)
	return nil
}

// gitStagedPaths returns the list of staged paths via
// `git diff --cached --name-only -z`. The -z separator handles paths
// with newlines/spaces.
func gitStagedPaths(repoRoot string) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only", "-z")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimRight(string(out), "\x00")
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\x00"), nil
}

// gitStagedAddedLines returns a map of staged path → added-line count
// via `git diff --cached --numstat`. Binary diffs (shown as "-") are
// reported as 0 added lines.
func gitStagedAddedLines(repoRoot string) (map[string]int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--numstat")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	result := map[string]int{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// numstat format: "<added>\t<removed>\t<path>"
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		added, err := strconv.Atoi(parts[0])
		if err != nil {
			added = 0 // binary diffs come back as "-"
		}
		result[parts[2]] = added
	}
	return result, nil
}
