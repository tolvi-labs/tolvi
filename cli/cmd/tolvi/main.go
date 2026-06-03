package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	clicmd "github.com/tolvi-labs/tolvi/cli/internal/cli"
	"github.com/tolvi-labs/tolvi/cli/internal/config"
	"github.com/tolvi-labs/tolvi/cli/internal/llm"
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

var initWorkspaceFlag string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Provision a new vault/ at the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		return clicmd.RunInit(clicmd.InitOpts{
			Cwd:       cwd,
			Workspace: initWorkspaceFlag,
			Stdout:    os.Stdout,
		})
	},
}

var (
	syncSlugFlag   string
	syncStatusFlag string
	syncBodyFlag   string
	syncNoEditFlag bool
	syncPrintFlag  bool
	syncVaultFlag  string
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

		return clicmd.RunSync(clicmd.SyncOpts{
			VaultPath: vaultPath,
			DocType:   docType,
			Title:     title,
			Slug:      syncSlugFlag,
			Status:    syncStatusFlag,
			BodyFlag:  syncBodyFlag,
			NoEdit:    syncNoEditFlag,
			PrintPath: syncPrintFlag,
			Stdout:    os.Stdout,
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

	syncCmd.Flags().StringVar(&syncSlugFlag, "slug", "", "override the auto-derived slug")
	syncCmd.Flags().StringVar(&syncStatusFlag, "status", "", "frontmatter status (default: active)")
	syncCmd.Flags().StringVar(&syncBodyFlag, "body", "", "body content (skips $EDITOR)")
	syncCmd.Flags().BoolVar(&syncNoEditFlag, "no-edit", false, "write skeleton-only file (no $EDITOR)")
	syncCmd.Flags().BoolVar(&syncPrintFlag, "print-path", false, "print only the resulting path on stdout")
	syncCmd.Flags().StringVar(&syncVaultFlag, "vault", "", "path to vault dir (default: walk up)")

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

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(recallCmd)

	precommitCmd.AddCommand(precommitInstallCmd)
	precommitCmd.AddCommand(precommitCheckCmd)
	precommitCmd.AddCommand(precommitUninstallCmd)
	rootCmd.AddCommand(precommitCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "tolvi:", err)
		os.Exit(clicmd.ExitInternal)
	}
}
