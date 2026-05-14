package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	clicmd "github.com/tolvi-labs/tolvi/cli/internal/cli"
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

func init() {
	initCmd.Flags().StringVar(&initWorkspaceFlag, "workspace", "", "workspace name (default: derived from git origin or cwd basename)")
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "tolvi:", err)
		os.Exit(clicmd.ExitInternal)
	}
}
