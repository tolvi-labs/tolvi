package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// RootsOpts configures `tolvi roots`.
type RootsOpts struct {
	VaultPath string
	// SessionNote prints only the absolute path of a date's session note,
	// for shell callers that need to know where the note lives without
	// reimplementing the routing rule.
	SessionNote bool
	Today       string // YYYY-MM-DD; empty means today
	Stdout      io.Writer
}

// RunRoots shows the chain of vault roots this repo resolves to, nearest scope
// first. It exists so the routing rule has one answer that humans, hooks and
// the doctor can all read, rather than each deriving its own.
func RunRoots(opts RootsOpts) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Today == "" {
		opts.Today = time.Now().Format("2006-01-02")
	}

	meta, err := vault.ReadMeta(opts.VaultPath)
	if err != nil {
		return err
	}
	chain, err := vault.ChainFor(meta, opts.VaultPath)
	if err != nil {
		return err
	}

	if opts.SessionNote {
		target, err := chain.Target("session", "")
		if err != nil {
			return err
		}
		fmt.Fprintln(opts.Stdout, chain.SessionNotePath(target, opts.Today))
		return nil
	}

	fmt.Fprintf(opts.Stdout, "workspace: %s\n", meta.Workspace)
	if meta.Repo != "" {
		fmt.Fprintf(opts.Stdout, "repo:      %s\n", meta.Repo)
	}
	if meta.Product != "" {
		fmt.Fprintf(opts.Stdout, "product:   %s\n", meta.Product)
	}
	fmt.Fprintln(opts.Stdout)
	fmt.Fprintln(opts.Stdout, "chain (nearest first):")
	for _, r := range chain.ReadRoots() {
		fmt.Fprintf(opts.Stdout, "  %-8s %s\n", r.Role, r.Path)
	}

	target, err := chain.Target("session", "")
	if err != nil {
		fmt.Fprintf(opts.Stdout, "\nsessions:  refused — %v\n", err)
		return nil
	}
	fmt.Fprintf(opts.Stdout, "\nsessions:  %s\n", chain.SessionNotePath(target, opts.Today))
	return nil
}
