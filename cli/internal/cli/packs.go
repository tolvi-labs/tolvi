package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/packs"
)

// PacksOpts is what `tolvi packs list` needs.
type PacksOpts struct {
	JSON    bool
	Stdout  io.Writer
	Version string
}

// RunPacksList prints the packs this binary carries.
func RunPacksList(opts PacksOpts) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	all, err := packs.List()
	if err != nil {
		return err
	}
	if opts.JSON {
		return printPacksListJSON(opts.Stdout, opts.Version, all)
	}
	for _, p := range all {
		fmt.Fprintf(opts.Stdout, "%-14s %d templates  %s\n", p.Name, len(p.Templates), strings.Join(p.Verticals, ", "))
	}
	fmt.Fprintf(opts.Stdout, "\n`tolvi init --pack <name>` provisions a vault with one of these.\n")
	return nil
}
