package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// InitOpts controls `tolvi init`.
type InitOpts struct {
	Cwd       string
	Workspace string
	Stdout    io.Writer
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

	workspace := opts.Workspace
	if workspace == "" {
		workspace = deriveWorkspace(opts.Cwd)
	}

	meta := vault.Meta{
		Workspace:      workspace,
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  vault.SupportedSchemaVersion,
	}
	if err := vault.WriteMeta(vaultDir, meta); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}

	fmt.Fprintf(opts.Stdout, "✓ Created vault/ at %s\n", vaultDir)
	fmt.Fprintf(opts.Stdout, "✓ Created vault/decisions/, vault/sessions/, vault/patterns/\n")
	fmt.Fprintf(opts.Stdout, "✓ Wrote vault/.vault-meta.json (workspace: %s)\n", workspace)
	fmt.Fprintln(opts.Stdout)
	fmt.Fprintln(opts.Stdout, "Next steps:")
	fmt.Fprintln(opts.Stdout, "  tolvi sync decision \"your first decision\"")
	fmt.Fprintln(opts.Stdout, "  tolvi ask \"...\"")
	return nil
}

// deriveWorkspace tries to extract a workspace name from <cwd>/.git/config's
// origin URL (the segment after the last "/" minus ".git"), falling back
// to filepath.Base(cwd).
func deriveWorkspace(cwd string) string {
	gitConfig := filepath.Join(cwd, ".git", "config")
	if data, err := os.ReadFile(gitConfig); err == nil {
		if name := extractRepoNameFromGitConfig(string(data)); name != "" {
			return name
		}
	}
	return filepath.Base(cwd)
}

var originURLRe = regexp.MustCompile(`url\s*=\s*(\S+)`)

func extractRepoNameFromGitConfig(config string) string {
	// Find the [remote "origin"] section.
	idx := strings.Index(config, `[remote "origin"]`)
	if idx < 0 {
		return ""
	}
	tail := config[idx:]
	m := originURLRe.FindStringSubmatch(tail)
	if m == nil {
		return ""
	}
	url := m[1]
	url = strings.TrimSuffix(url, ".git")
	// Handle both git@host:org/repo and https://host/org/repo.
	if i := strings.LastIndexAny(url, "/:"); i >= 0 {
		return url[i+1:]
	}
	return url
}
