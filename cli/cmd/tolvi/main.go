package main

import (
	"fmt"
	"os"
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "tolvi:", err)
		os.Exit(clicmd.ExitInternal)
	}
}
