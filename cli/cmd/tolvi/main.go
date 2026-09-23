package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	clicmd "github.com/tolvi-labs/tolvi/cli/internal/cli"
	"github.com/tolvi-labs/tolvi/cli/internal/config"
	"github.com/tolvi-labs/tolvi/cli/internal/integrations"
	"github.com/tolvi-labs/tolvi/cli/internal/llm"
	"github.com/tolvi-labs/tolvi/cli/internal/registry"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

var rootCmd = &cobra.Command{
	Use:   "tolvi",
	Short: "Tolvi — engineering vault CLI",
	Long: `Tolvi is a CLI for the per-repo engineering knowledge vault.

It reads decisions, sessions, and patterns stored as Markdown with
frontmatter under <repo>/vault/, and answers questions about them via
the Anthropic API.

For the format spec, see https://tolvilabs.com/tolvi/spec/.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var (
	initWorkspaceFlag string
	initPackFlag      string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Provision a new vault/ at the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := clicmd.RunInit(clicmd.InitOpts{
			Cwd:       cwd,
			Workspace: initWorkspaceFlag,
			Pack:      initPackFlag,
			Stdout:    os.Stdout,
		}); err != nil {
			return err
		}
		// Best-effort: the vault is what init delivers, and the registry is a
		// cache a scan can rebuild, so a failure here warns on stderr and
		// never fails the command.
		clicmd.RegisterAfterInit(os.Stderr, registry.ConfigPath(), cwd)
		return nil
	},
}

var (
	syncSlugFlag         string
	syncStatusFlag       string
	syncBodyFlag         string
	syncNoEditFlag       bool
	syncPrintFlag        bool
	syncVaultFlag        string
	syncPrivateFlag      bool
	syncPrivateVaultFlag string
)

var syncCmd = &cobra.Command{
	Use:   "sync <type> <title...>",
	Short: "Create a new vault doc (decision | session | pattern)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docType := args[0]
		title := strings.Join(args[1:], " ")

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		home, _ := os.UserHomeDir()
		vaultPath, err := vault.Discover(vault.DiscoverOpts{
			StartDir:     cwd,
			HomeDir:      home,
			ExplicitPath: firstNonEmpty(syncVaultFlag, os.Getenv("TOLVI_VAULT")),
		})
		if err != nil {
			return err
		}

		docVisibility := ""
		if syncPrivateFlag {
			docVisibility = "private"
		}
		return clicmd.RunSync(clicmd.SyncOpts{
			VaultPath:     vaultPath,
			DocType:       docType,
			Title:         title,
			Slug:          syncSlugFlag,
			Status:        syncStatusFlag,
			BodyFlag:      syncBodyFlag,
			NoEdit:        syncNoEditFlag,
			PrintPath:     syncPrintFlag,
			DocVisibility: docVisibility,
			PrivateVault:  syncPrivateVaultFlag,
			Stdout:        os.Stdout,
		})
	},
}

var (
	askVaultFlag         string
	askModelFlag         string
	askIncludeStatusFlag string
	askExcludeTypeFlag   string
	askJSONFlag          bool
	askNoStreamFlag      bool
)

var askCmd = &cobra.Command{
	Use:   "ask <query...>",
	Short: "Ask the vault a question (CAG: whole vault → Anthropic)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		cfg := config.Load(config.LoadOpts{
			HomeDir: home,
			Env:     os.Getenv,
		})

		vaultPath, err := vault.Discover(vault.DiscoverOpts{
			StartDir:     cwd,
			HomeDir:      home,
			ExplicitPath: firstNonEmpty(askVaultFlag, os.Getenv("TOLVI_VAULT")),
			DefaultVault: cfg.DefaultVault,
		})
		if err != nil {
			return err
		}

		model := cfg.Model
		if askModelFlag != "" {
			model = askModelFlag
		}
		client, err := llm.NewClient(llm.ClientOpts{
			APIKey:  cfg.AnthropicAPIKey,
			Model:   model,
			BaseURL: os.Getenv("ANTHROPIC_BASE_URL"),
		})
		if err != nil {
			return err
		}

		opts := clicmd.AskOpts{
			VaultPath: vaultPath,
			Query:     query,
			LLM:       client,
			Stdout:    os.Stdout,
			Stderr:    os.Stderr,
			JSON:      askJSONFlag,
			NoStream:  askNoStreamFlag,
			Model:     model,
		}
		if askIncludeStatusFlag != "" {
			opts.IncludeStatuses = parseCSV(askIncludeStatusFlag)
		}
		if askExcludeTypeFlag != "" {
			opts.ExcludeTypes = parseCSV(askExcludeTypeFlag)
		}
		return clicmd.RunAsk(opts)
	},
}

var (
	recallVaultFlag           string
	recallFormatFlag          string
	recallSessionCountFlag    int
	recallDecisionCountFlag   int
	recallMaxBytesFlag        int
	recallIncludePatternsFlag bool
)

var recallCmd = &cobra.Command{
	Use:   "recall",
	Short: "Surface recent sessions and active decisions from the vault",
	Long: `recall reads the vault directly (no API call) and prints a structured
summary of recent sessions and active decisions.

Use --format hook-json to emit the Claude Code SessionStart hook JSON blob,
suitable for piping from a hooks/session-recall.sh script.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		cfg := config.Load(config.LoadOpts{
			HomeDir: home,
			Env:     os.Getenv,
		})

		vaultPath, err := vault.Discover(vault.DiscoverOpts{
			StartDir:     cwd,
			HomeDir:      home,
			ExplicitPath: firstNonEmpty(recallVaultFlag, os.Getenv("TOLVI_VAULT")),
			DefaultVault: cfg.DefaultVault,
		})
		if err != nil {
			return err
		}

		// Flag > config > compiled-in default. Zero values mean "use default"
		// (applied inside RunRecall).
		sessionCount := recallSessionCountFlag
		if sessionCount == 0 {
			sessionCount = cfg.Recall.SessionCount
		}
		decisionCount := recallDecisionCountFlag
		if decisionCount == 0 {
			decisionCount = cfg.Recall.DecisionCount
		}
		maxBytes := recallMaxBytesFlag
		if maxBytes == 0 {
			maxBytes = cfg.Recall.MaxBytes
		}

		return clicmd.RunRecall(clicmd.RecallOpts{
			VaultPath:       vaultPath,
			SessionCount:    sessionCount,
			DecisionCount:   decisionCount,
			MaxBytes:        maxBytes,
			IncludePatterns: recallIncludePatternsFlag || cfg.Recall.IncludePatterns,
			Format:          firstNonEmpty(recallFormatFlag),
			Stdout:          os.Stdout,
		})
	},
}

var (
	commitMessageFlag      string
	commitVaultFlag        string
	commitPrivateVaultFlag string
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Stage vault/ and commit — gated on a session note for today (mechanical; no synthesis)",
	Long: `commit is the controlled, deterministic capture path: it requires a
session note for today to already exist, auto-stages vault/ so the vault
always lands in the commit, then runs git commit. It does NO synthesis and
runs no LLM — what you commit is exactly what is there.

Use this for scripted, CI, or precision commits where you want no surprises.
For comprehensive capture, run the /tolvi-commit skill inside a working
session instead: it synthesizes the whole session (decisions, patterns, log)
from the conversation, then commits. Mechanical for known capture; the skill
for synthesizing messy reality.

If no session note exists for today, commit refuses and points you at
'tolvi sync session' (controlled) or the /tolvi-commit skill (synthesized).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot("")
		if err != nil {
			return err
		}
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		vaultPath, err := vault.Discover(vault.DiscoverOpts{
			StartDir:     cwd,
			HomeDir:      home,
			ExplicitPath: firstNonEmpty(commitVaultFlag, os.Getenv("TOLVI_VAULT")),
		})
		if err != nil {
			return err
		}

		err = clicmd.RunCommit(clicmd.CommitOpts{
			RepoRoot:     repoRoot,
			VaultPath:    vaultPath,
			Message:      commitMessageFlag,
			PrivateVault: commitPrivateVaultFlag,
			Stdin:        os.Stdin,
			Stdout:       os.Stdout,
			Stderr:       os.Stderr,
		})
		if errors.Is(err, clicmd.ErrNoSessionNote) {
			os.Exit(clicmd.ExitVaultState)
		}
		return err
	},
}

var (
	precommitForceFlag      bool
	precommitAppendFlag     bool
	precommitRepoFlag       string
	precommitUninstallForce bool
)

var precommitCmd = &cobra.Command{
	Use:   "precommit",
	Short: "Manage the tolvi git pre-commit hook",
}

var precommitInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the tolvi pre-commit hook into the current repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot(precommitRepoFlag)
		if err != nil {
			return err
		}
		mode := clicmd.InstallModeDefault
		switch {
		case precommitForceFlag && precommitAppendFlag:
			return fmt.Errorf("--force and --append are mutually exclusive")
		case precommitForceFlag:
			mode = clicmd.InstallModeForce
		case precommitAppendFlag:
			mode = clicmd.InstallModeAppend
		}
		result, err := clicmd.InstallShim(clicmd.InstallOpts{RepoRoot: repoRoot, Mode: mode})
		if err != nil {
			fmt.Fprintln(os.Stderr, "tolvi: "+err.Error())
			os.Exit(clicmd.ExitVaultState)
		}
		printInstallResult(result, repoRoot)
		return nil
	},
}

var precommitCheckCmd = &cobra.Command{
	Use:    "check",
	Short:  "Run the precommit heuristics on the staged diff (called by the hook)",
	Hidden: true, // rarely run by humans
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot(precommitRepoFlag)
		if err != nil {
			return nil // check NEVER errors externally
		}
		quiet := os.Getenv("TOLVI_PRECOMMIT_QUIET") != ""
		_ = clicmd.RunCheck(clicmd.CheckOpts{RepoRoot: repoRoot, Stderr: os.Stderr, Quiet: quiet})
		return nil
	},
}

var precommitUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the tolvi pre-commit hook from the current repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := resolveRepoRoot(precommitRepoFlag)
		if err != nil {
			return err
		}
		result, err := clicmd.UninstallShim(clicmd.UninstallOpts{RepoRoot: repoRoot, Force: precommitUninstallForce})
		if err != nil {
			fmt.Fprintln(os.Stderr, "tolvi: "+err.Error())
			os.Exit(clicmd.ExitVaultState)
		}
		printUninstallResult(result)
		return nil
	},
}

// resolveRepoRoot returns the git repo root for the precommit cobra
// handlers. If flag is non-empty, returns it directly. Otherwise walks
// up from $PWD looking for a .git directory.
func resolveRepoRoot(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository (no .git found in any ancestor of %s)", cwd)
		}
		dir = parent
	}
}

func printInstallResult(r clicmd.InstallResult, repoRoot string) {
	fmt.Printf("✓ Repo root: %s\n", repoRoot)
	switch r.Action {
	case clicmd.InstallActionWrote:
		fmt.Printf("✓ Wrote %s (4 lines)\n", r.HookPath)
		fmt.Println("✓ Chmod +x")
		fmt.Println()
		fmt.Println("The hook will print a nudge on commits that touch dependency manifests,")
		fmt.Println("infra config, tooling config, or that add >500 lines.")
		fmt.Println()
		fmt.Println("To silence per-shell: export TOLVI_PRECOMMIT_QUIET=1")
		fmt.Println("To remove: tolvi precommit uninstall")
	case clicmd.InstallActionAlreadyInstalled:
		fmt.Printf("✓ Already installed at %s\n", r.HookPath)
	case clicmd.InstallActionReplaced:
		prevLines := len(strings.Split(strings.TrimSpace(string(r.PrevContent)), "\n"))
		fmt.Printf("✓ Replaced existing hook at %s (was %d lines)\n", r.HookPath, prevLines)
	case clicmd.InstallActionAppended:
		prevLines := len(strings.Split(strings.TrimSpace(string(r.PrevContent)), "\n"))
		fmt.Printf("✓ Appended tolvi check to existing %s (was %d lines, now %d)\n",
			r.HookPath, prevLines, prevLines+2)
	}
}

func printUninstallResult(r clicmd.UninstallResult) {
	switch r.Action {
	case clicmd.UninstallActionRemoved:
		fmt.Printf("✓ Removed %s\n", r.HookPath)
	case clicmd.UninstallActionNoOp:
		fmt.Printf("✓ No hook installed at %s\n", r.HookPath)
	}
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	initCmd.Flags().StringVar(&initWorkspaceFlag, "workspace", "", "workspace name (default: derived from git origin or cwd basename)")
	initCmd.Flags().StringVar(&initPackFlag, "pack", "", "role pack whose templates to write into vault/templates/ (see `tolvi packs list`)")

	syncCmd.Flags().StringVar(&syncSlugFlag, "slug", "", "override the auto-derived slug")
	syncCmd.Flags().StringVar(&syncStatusFlag, "status", "", "frontmatter status (default: active)")
	syncCmd.Flags().StringVar(&syncBodyFlag, "body", "", "body content (skips $EDITOR)")
	syncCmd.Flags().BoolVar(&syncNoEditFlag, "no-edit", false, "write skeleton-only file (no $EDITOR)")
	syncCmd.Flags().BoolVar(&syncPrintFlag, "print-path", false, "print only the resulting path on stdout")
	syncCmd.Flags().StringVar(&syncVaultFlag, "vault", "", "path to vault dir (default: walk up)")
	syncCmd.Flags().BoolVar(&syncPrivateFlag, "private", false, "mark this decision/pattern as private (routes to the private vault under public visibility)")
	syncCmd.Flags().StringVar(&syncPrivateVaultFlag, "private-vault", "", "path to the private vault (overrides the org root declared in roots.json)")

	askCmd.Flags().StringVar(&askVaultFlag, "vault", "", "path to vault dir (default: walk up)")
	askCmd.Flags().StringVar(&askModelFlag, "model", "", "override the configured Anthropic model")
	askCmd.Flags().StringVar(&askIncludeStatusFlag, "include-status", "", "comma-separated statuses to include (default: active,in-progress,historical)")
	askCmd.Flags().StringVar(&askExcludeTypeFlag, "exclude-type", "", "comma-separated doc types to omit (e.g., session)")
	askCmd.Flags().BoolVar(&askJSONFlag, "json", false, "emit JSON instead of streaming text")
	askCmd.Flags().BoolVar(&askNoStreamFlag, "no-stream", false, "buffer output instead of streaming")

	recallCmd.Flags().StringVar(&recallVaultFlag, "vault", "", "path to vault dir (default: walk up)")
	recallCmd.Flags().StringVar(&recallFormatFlag, "format", "", "output format: human (default) | hook-json")
	recallCmd.Flags().IntVar(&recallSessionCountFlag, "session-count", 0, "number of recent sessions to surface (default: 3)")
	recallCmd.Flags().IntVar(&recallDecisionCountFlag, "decision-count", 0, "max decisions to surface (default: 10)")
	recallCmd.Flags().IntVar(&recallMaxBytesFlag, "max-bytes", 0, "byte budget for hook-json additionalContext (0 = unlimited)")
	recallCmd.Flags().BoolVar(&recallIncludePatternsFlag, "include-patterns", false, "include patterns in output (default: false)")

	commitCmd.Flags().StringVarP(&commitMessageFlag, "message", "m", "", "commit message (if omitted, git opens $EDITOR)")
	commitCmd.Flags().StringVar(&commitVaultFlag, "vault", "", "path to vault dir (default: walk up)")
	rootsCmd.Flags().StringVar(&rootsVaultFlag, "vault", "", "path to the vault (default: discovered from the working directory)")
	rootsCmd.Flags().BoolVar(&rootsSessionNoteFlag, "session-note", false, "print only the absolute path of today's session note")

	commitCmd.Flags().StringVar(&commitPrivateVaultFlag, "private-vault", "", "path to the private vault (overrides the org root declared in roots.json)")

	precommitInstallCmd.Flags().BoolVar(&precommitForceFlag, "force", false, "overwrite an existing non-tolvi hook")
	precommitInstallCmd.Flags().BoolVar(&precommitAppendFlag, "append", false, "append tolvi check to an existing hook instead of overwriting")
	precommitInstallCmd.Flags().StringVar(&precommitRepoFlag, "repo", "", "path to the repo root (default: walk up from cwd)")
	precommitCheckCmd.Flags().StringVar(&precommitRepoFlag, "repo", "", "path to the repo root (default: walk up from cwd)")
	precommitUninstallCmd.Flags().StringVar(&precommitRepoFlag, "repo", "", "path to the repo root (default: walk up from cwd)")
	precommitUninstallCmd.Flags().BoolVar(&precommitUninstallForce, "force", false, "remove the hook even if not installed by tolvi")
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

var (
	rootsVaultFlag       string
	rootsSessionNoteFlag bool
)

var rootsCmd = &cobra.Command{
	Use:   "roots",
	Short: "Show the chain of vault roots this repo resolves to",
	Long: `roots prints the repo's identity and the chain of vault roots it resolves
to, nearest scope first, along with where today's session note belongs.

The routing rule lives in one place, and this is how a human, a shell hook, or
a bug report reads it without deriving its own answer.

Use --session-note to print only the absolute path of today's session note,
which is what the commit gate calls.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		cfg := config.Load(config.LoadOpts{HomeDir: home, Env: os.Getenv})

		vaultPath, err := vault.Discover(vault.DiscoverOpts{
			StartDir:     cwd,
			HomeDir:      home,
			ExplicitPath: firstNonEmpty(rootsVaultFlag, os.Getenv("TOLVI_VAULT")),
			DefaultVault: cfg.DefaultVault,
		})
		if err != nil {
			return err
		}
		return clicmd.RunRoots(clicmd.RootsOpts{
			VaultPath:   vaultPath,
			SessionNote: rootsSessionNoteFlag,
			Stdout:      os.Stdout,
		})
	},
}

var (
	doctorVaultFlag string
	doctorJSONFlag  bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [vault-health]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Check that the local tolvi setup is sound, and say how to fix what is not",
	Long: `doctor inspects the things a working tolvi install depends on: whether the
binary is reachable as ` + "`tolvi`" + ` on PATH, whether a vault resolves from here,
whether ANTHROPIC_API_KEY is set, and whether the Claude Code allow rules are
in place. Each failing check prints the command that fixes it.

It then scans the vault's contents for the defects that stop a note being
found: empty tags, unrecognized status values, duplicate titles, unfilled
template placeholders, and escaped unicode.

Pass "vault-health" to run only the content scan.

Exit codes differ by what is being asserted. A plain run exits non-zero when a
SETUP check fails, because that is what stops the tools working; content
findings are reported but do not change it. "tolvi doctor vault-health" exits
non-zero when the scan finds any high-severity defect.

--json emits the same result as machine-readable JSON instead of the report,
against the published schemas in spec/schemas/. Exit codes are unchanged, so a
caller can read either the code or the "ok" field.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		home, _ := os.UserHomeDir()
		only := ""
		if len(args) == 1 {
			if args[0] != "vault-health" {
				return fmt.Errorf("unknown section %q (the only section is \"vault-health\")", args[0])
			}
			only = args[0]
		}
		opts := clicmd.DoctorOpts{
			StartDir:      cwd,
			HomeDir:       home,
			ExplicitVault: firstNonEmpty(doctorVaultFlag, os.Getenv("TOLVI_VAULT")),
			Env:           os.Getenv,
			LookPath:      exec.LookPath,
			Stdout:        os.Stdout,
			Only:          only,
		}

		if only == "vault-health" {
			if doctorJSONFlag {
				// The text report goes to the writer doctor was given; JSON
				// owns stdout alone, so nothing interleaves with it.
				opts.Stdout = io.Discard
			}
			rep, err := clicmd.DoctorVaultHealth(opts)
			if err != nil {
				return err
			}
			if doctorJSONFlag {
				if err := clicmd.PrintVaultHealthJSON(os.Stdout, version, rep); err != nil {
					return err
				}
				if clicmd.HealthHighSeverity(rep) > 0 {
					os.Exit(clicmd.ExitVaultState)
				}
				return nil
			}
			if err := clicmd.RenderVaultHealth(os.Stdout, rep); err != nil {
				return err
			}
			if clicmd.HealthHighSeverity(rep) > 0 {
				os.Exit(clicmd.ExitVaultState)
			}
			return nil
		}

		if doctorJSONFlag {
			// --json reports the setup checks; the content scan has its own
			// shape and its own invocation.
			opts.Stdout = io.Discard
			opts.SkipVaultHealth = true
		}
		checks, err := clicmd.RunDoctor(opts)
		if err != nil {
			return err
		}
		if doctorJSONFlag {
			if err := clicmd.PrintDoctorJSON(os.Stdout, version, checks); err != nil {
				return err
			}
		}
		if clicmd.DoctorFailures(checks) > 0 {
			os.Exit(clicmd.ExitConfig)
		}
		return nil
	},
}

var reposJSONFlag bool

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "The repos on this machine that have a vault",
	Long: `repos maintains a machine-local index at ~/.config/tolvi/repos.json, beside
roots.json and never committed. ` + "`tolvi init`" + ` registers a repo; scan finds ones
that were never registered.

The index is a cache, not a source of truth. Each repo's own .vault-meta.json
is the truth, so entries are verified when they are read: a repo that has moved
is reported as stale rather than trusted, and forgetting it is a deliberate
step rather than something a read does silently.`,
}

var reposListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered repos, verified against the disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		return clicmd.RunReposList(clicmd.ReposOpts{
			RegistryPath: registry.ConfigPath(),
			JSON:         reposJSONFlag,
			Stdout:       os.Stdout,
			Version:      version,
		})
	},
}

var reposScanCmd = &cobra.Command{
	Use:   "scan [dir...]",
	Short: "Rebuild the index from what is on disk",
	Long: `scan searches the directories given and registers every repo that has a
readable vault. With no argument it searches the directories holding this
machine's declared roots, which is the set it already knows about; a machine
with no roots.json is asked to name a directory rather than having one guessed
for it.

scan replaces the index, so a repo that is gone stops being listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return clicmd.RunReposScan(clicmd.ReposOpts{
			RegistryPath: registry.ConfigPath(),
			Dirs:         args,
			Stdout:       os.Stdout,
			Version:      version,
		})
	},
}

var reposForgetCmd = &cobra.Command{
	Use:   "forget <path>",
	Short: "Drop one repo from the index",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return clicmd.RunReposForget(clicmd.ReposOpts{
			RegistryPath: registry.ConfigPath(),
			Dirs:         args,
			Stdout:       os.Stdout,
			Version:      version,
		})
	},
}

var (
	integrationsForceFlag bool
	integrationsHooksFlag bool
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Install this repo's Claude Code integration from the binary",
}

var integrationsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Write the Tolvi skill, slash commands, and optionally the session hooks",
	Long: `install writes the Claude Code integration carried inside this binary:
the Tolvi skill, the three slash commands, and with --with-hooks the session
hooks and the read-only allow rules.

The files ride inside the binary because the Homebrew cask ships exactly one
artifact. A brew install therefore had no route to the integration at all: the
only installer was skills/tolvi/install.sh, which needs a checkout.

Only what this repo owns is installed. Stack skills live in their own product
repos and are installed from there, so what is found on this machine is
reported rather than wired up.

Existing files are never overwritten without --force, and the hook merge is
idempotent: running install twice changes nothing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := integrations.Install(integrations.Opts{
			Force:     integrationsForceFlag,
			WithHooks: integrationsHooksFlag,
		})
		if err != nil {
			return err
		}
		for _, w := range res.Written {
			fmt.Printf("  wrote %s\n", w)
		}
		if res.HooksWired {
			fmt.Printf("  hooks wired in %s\n", res.SettingsPath)
			for _, rule := range res.AllowAdded {
				fmt.Printf("  allowed %s\n", rule)
			}
			if len(res.AllowAdded) == 0 {
				fmt.Println("  allow rules already present")
			}
		} else {
			fmt.Println("\nRun with --with-hooks to wire recall and the pre-commit vault check.")
		}
		reportStackSkills()
		return nil
	},
}

// reportStackSkills names stack skills found in registered repos without
// touching them. They are installed by their own product repos, per
// vault/decisions/2026-09-12-stack-skills-live-in-their-product-repos.md, and
// a binary cannot recreate a symlink into a checkout it does not own.
func reportStackSkills() {
	f, err := registry.Load(registry.ConfigPath())
	if err != nil || len(f.Repos) == 0 {
		return
	}
	var found []string
	for _, e := range f.Repos {
		if e.Repo == "tolvi" {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(e.Path, "skills", "*", "SKILL.md"))
		for _, m := range matches {
			found = append(found, filepath.Base(filepath.Dir(m))+" in "+e.Path)
		}
	}
	if len(found) == 0 {
		return
	}
	fmt.Println("\nStack skills on this machine, installed from their own repos:")
	for _, s := range found {
		fmt.Printf("  %s\n", s)
	}
}

var packsJSONFlag bool

var packsCmd = &cobra.Command{
	Use:   "packs",
	Short: "The role packs this binary can provision a vault with",
	Long: `Role packs are sets of vault templates tuned to a kind of work. They are
owned by tolvi-solo and vendored into this binary, so ` + "`tolvi init --pack`" + `
works with no sibling checkout.`,
}

var packsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the vendored role packs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return clicmd.RunPacksList(clicmd.PacksOpts{
			JSON:    packsJSONFlag,
			Stdout:  os.Stdout,
			Version: version,
		})
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	reposCmd.AddCommand(reposListCmd, reposScanCmd, reposForgetCmd)
	reposListCmd.Flags().BoolVar(&reposJSONFlag, "json", false, "emit JSON against spec/schemas/repos-list.json")
	rootCmd.AddCommand(reposCmd)
	integrationsInstallCmd.Flags().BoolVar(&integrationsForceFlag, "force", false, "overwrite an existing install")
	integrationsInstallCmd.Flags().BoolVar(&integrationsHooksFlag, "with-hooks", false, "also wire the session hooks and read-only allow rules")
	integrationsCmd.AddCommand(integrationsInstallCmd)
	rootCmd.AddCommand(integrationsCmd)
	packsListCmd.Flags().BoolVar(&packsJSONFlag, "json", false, "emit JSON against spec/schemas/packs-list.json")
	packsCmd.AddCommand(packsListCmd)
	rootCmd.AddCommand(packsCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(recallCmd)
	rootCmd.AddCommand(commitCmd)
	doctorCmd.Flags().StringVar(&doctorVaultFlag, "vault", "", "path to the vault (default: discovered from $PWD)")
	doctorCmd.Flags().BoolVar(&doctorJSONFlag, "json", false, "emit JSON against spec/schemas/ instead of the text report")
	rootCmd.AddCommand(rootsCmd)
	rootCmd.AddCommand(doctorCmd)

	precommitCmd.AddCommand(precommitInstallCmd)
	precommitCmd.AddCommand(precommitCheckCmd)
	precommitCmd.AddCommand(precommitUninstallCmd)
	rootCmd.AddCommand(precommitCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "tolvi:", err)
		os.Exit(clicmd.ExitInternal)
	}
}
