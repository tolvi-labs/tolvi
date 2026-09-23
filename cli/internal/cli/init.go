package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/packs"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// InitOpts controls `tolvi init`.
type InitOpts struct {
	Cwd       string
	Workspace string
	Stdout    io.Writer
	// Pack names a role pack whose templates are written into
	// vault/templates/. Empty provisions the vault with no templates, which
	// is what init did before packs were vendored.
	Pack string
}

// RunInit provisions <Cwd>/vault/ with the three subdirs and
// .vault-meta.json. Refuses if .vault-meta.json already exists.
func RunInit(opts InitOpts) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	vaultDir := filepath.Join(opts.Cwd, "vault")
	metaPath := filepath.Join(vaultDir, ".vault-meta.json")
	if _, err := os.Stat(metaPath); err == nil {
		return fmt.Errorf("vault already exists at %s — refusing to overwrite", vaultDir)
	}

	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}
	for _, sub := range []string{"decisions", "sessions", "patterns"} {
		if err := os.MkdirAll(filepath.Join(vaultDir, sub), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", sub, err)
		}
	}

	workspace, repo := deriveIdentity(opts.Cwd)
	if opts.Workspace != "" {
		workspace = opts.Workspace
	}

	meta := vault.Meta{
		Workspace:      workspace,
		Repo:           repo,
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  vault.SupportedSchemaVersion,
	}
	if err := vault.WriteMeta(vaultDir, meta); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}

	fmt.Fprintf(opts.Stdout, "✓ Created vault/ at %s\n", vaultDir)
	fmt.Fprintf(opts.Stdout, "✓ Created vault/decisions/, vault/sessions/, vault/patterns/\n")
	fmt.Fprintf(opts.Stdout, "✓ Wrote vault/.vault-meta.json (workspace: %s)\n", workspace)

	if opts.Pack != "" {
		written, err := packs.Install(opts.Pack, filepath.Join(vaultDir, "templates"))
		if err != nil {
			// The vault is provisioned by this point, so an unknown pack is
			// reported rather than rolled back: deleting a vault someone may
			// already be writing into would be the worse failure.
			fmt.Fprintf(opts.Stdout, "✗ %v\n", err)
			return err
		}
		fmt.Fprintf(opts.Stdout, "✓ Wrote vault/templates/ from the %s pack (%d templates)\n", opts.Pack, len(written))
	}

	fmt.Fprintln(opts.Stdout)
	fmt.Fprintln(opts.Stdout, "Next steps:")
	fmt.Fprintln(opts.Stdout, "  tolvi sync decision \"your first decision\"")
	fmt.Fprintln(opts.Stdout, "  tolvi ask \"...\"")
	return nil
}

// deriveIdentity splits <cwd>/.git/config's origin URL into the workspace
// (the org segment) and the repo (the segment after it). With no usable
// remote, the directory name serves as both: there is no org to name.
func deriveIdentity(cwd string) (workspace, repo string) {
	gitConfig := filepath.Join(cwd, ".git", "config")
	if data, err := os.ReadFile(gitConfig); err == nil {
		if org, name := extractIdentityFromGitConfig(string(data)); name != "" {
			if org == "" {
				org = name
			}
			return org, name
		}
	}
	base := filepath.Base(cwd)
	return base, base
}

var originURLRe = regexp.MustCompile(`url\s*=\s*(\S+)`)

// extractIdentityFromGitConfig pulls (org, repo) out of the origin URL,
// handling both git@host:org/repo and https://host/org/repo.
func extractIdentityFromGitConfig(config string) (org, repo string) {
	idx := strings.Index(config, `[remote "origin"]`)
	if idx < 0 {
		return "", ""
	}
	m := originURLRe.FindStringSubmatch(config[idx:])
	if m == nil {
		return "", ""
	}
	url := strings.TrimSuffix(m[1], ".git")
	i := strings.LastIndexAny(url, "/:")
	if i < 0 {
		return "", url
	}
	repo = url[i+1:]
	// The org is the segment before the repo, after the host separator.
	if j := strings.LastIndexAny(url[:i], "/:"); j >= 0 {
		org = url[j+1 : i]
	}
	return org, repo
}
