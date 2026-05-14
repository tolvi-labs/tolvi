package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNoVault is returned by Discover when no vault could be located.
var ErrNoVault = errors.New("no vault found — run 'tolvi init' to create one, or set --vault")

// DiscoverOpts controls vault resolution. Fields are checked in this order:
//  1. ExplicitPath (--vault flag or TOLVI_VAULT env)
//  2. Walk up from StartDir, stopping at HomeDir or filesystem root
//  3. DefaultVault (from config file)
//
// Returns the absolute path to the vault directory (the one containing
// .vault-meta.json), not its parent.
type DiscoverOpts struct {
	StartDir     string // typically os.Getwd()
	HomeDir      string // typically os.UserHomeDir()
	ExplicitPath string // --vault flag / TOLVI_VAULT env (may be "")
	DefaultVault string // config.default_vault (may be "")
}

// Discover resolves the vault to use. See DiscoverOpts for precedence.
func Discover(opts DiscoverOpts) (string, error) {
	if opts.ExplicitPath != "" {
		abs, err := filepath.Abs(opts.ExplicitPath)
		if err != nil {
			return "", fmt.Errorf("resolve --vault path: %w", err)
		}
		if _, err := os.Stat(filepath.Join(abs, ".vault-meta.json")); err != nil {
			return "", fmt.Errorf("%s: not a vault (.vault-meta.json missing)", abs)
		}
		return abs, nil
	}

	if walked, ok := walkUp(opts.StartDir, opts.HomeDir); ok {
		return walked, nil
	}

	if opts.DefaultVault != "" {
		abs, err := filepath.Abs(opts.DefaultVault)
		if err != nil {
			return "", fmt.Errorf("resolve default_vault: %w", err)
		}
		if _, err := os.Stat(filepath.Join(abs, ".vault-meta.json")); err != nil {
			return "", fmt.Errorf("default_vault %s: not a vault (.vault-meta.json missing)", abs)
		}
		return abs, nil
	}

	return "", ErrNoVault
}

// walkUp climbs from dir toward / looking for a `vault/.vault-meta.json`
// at each level. Returns the resolved vault directory on success.
//
// Walk stops at:
//   - dir == filepath.Dir(dir) (filesystem root)
//   - dir == home (don't escape the user's home tree, prevents surprises)
//
// Special case: if the start directory IS itself a vault root (contains
// .vault-meta.json directly), return it without going up. This handles
// `cd vault/decisions && tolvi ask ...` where StartDir resolves up.
func walkUp(start, home string) (string, bool) {
	if abs, err := filepath.Abs(start); err == nil {
		start = abs
	}
	if home != "" {
		if abs, err := filepath.Abs(home); err == nil {
			home = abs
		}
	}

	dir := start
	for {
		// Check if this directory is itself a vault root.
		if _, err := os.Stat(filepath.Join(dir, ".vault-meta.json")); err == nil {
			return dir, true
		}
		// Check for a `vault/` subdirectory containing the meta.
		candidate := filepath.Join(dir, "vault")
		if _, err := os.Stat(filepath.Join(candidate, ".vault-meta.json")); err == nil {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false // filesystem root
		}
		if home != "" && parent == home {
			// Check $HOME itself once before stopping.
			if _, err := os.Stat(filepath.Join(home, "vault", ".vault-meta.json")); err == nil {
				return filepath.Join(home, "vault"), true
			}
			return "", false
		}
		dir = parent
	}
}
