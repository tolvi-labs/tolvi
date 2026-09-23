// Package integrations installs this repo's Claude Code integration from
// files carried inside the binary.
//
// The cask ships exactly one artifact, the binary, so a brew install used to
// yield a CLI with no route to the skill, hooks or slash commands: the only
// installer was skills/tolvi/install.sh, which needs a checkout. Embedding the
// tree closes that.
//
// The embedded copy is a forced duplicate of skills/tolvi/, which the plugin
// loader requires at the repo root and //go:embed cannot reach from here. It
// is pinned by .github/scripts/integration-copy-parity-check.sh, and
// skills/tolvi/ is canonical. See
// vault/decisions/2026-09-23-duplicate-and-pin-is-a-mitigation-not-a-default.md.
package integrations

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Files carries the integration tree. The `all:` prefix is load-bearing:
// without it //go:embed silently skips commands/_preflight.md, because embed
// ignores names beginning with _ or . unless asked not to.
//
//go:embed all:files
var Files embed.FS

// Opts is what an install needs. HomeDir is injectable so tests never touch
// the developer's real ~/.claude.
type Opts struct {
	HomeDir string
	// Force overwrites an existing install. Off by default, because the
	// files are ones people customize.
	Force bool
	// WithHooks also wires the session hooks and the read-only allow rules
	// into ~/.claude/settings.json.
	WithHooks bool
}

// Result reports what an install did, so the command can print it rather than
// the package printing anything itself.
type Result struct {
	Written      []string
	HooksWired   bool
	AllowAdded   []string
	SettingsPath string
}

// readOnlyAllowRules are the rules an install adds. Reads only: `tolvi sync`
// and `tolvi commit` mutate the vault and the git tree, so they stay a
// conscious per-call approval. This mirrors what skills/tolvi/install.sh
// allowlists, deliberately.
var readOnlyAllowRules = []string{"Bash(tolvi recall:*)", "Bash(tolvi ask:*)"}

// commandFiles are the slash commands an install publishes. _preflight.md is
// not among them: it is the drift check's canonical source, physically
// duplicated into each command file, and installing it would put a
// non-command in the commands menu.
var commandFiles = []string{"tolvi-recall.md", "tolvi-sync.md", "tolvi-commit.md"}

// Install writes the skill, the slash commands, and optionally the hooks.
func Install(opts Opts) (Result, error) {
	var res Result
	if opts.HomeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return res, err
		}
		opts.HomeDir = home
	}
	claude := filepath.Join(opts.HomeDir, ".claude")

	skillPath := filepath.Join(claude, "skills", "tolvi", "SKILL.md")
	if !opts.Force {
		if _, err := os.Stat(skillPath); err == nil {
			return res, fmt.Errorf("%s already exists; pass --force to overwrite it", skillPath)
		}
	}

	if err := copyEmbedded("files/SKILL.md", skillPath); err != nil {
		return res, err
	}
	res.Written = append(res.Written, skillPath)

	for _, name := range commandFiles {
		dst := filepath.Join(claude, "commands", name)
		if err := copyEmbedded(filepath.Join("files/commands", name), dst); err != nil {
			return res, err
		}
		res.Written = append(res.Written, dst)
	}

	if !opts.WithHooks {
		return res, nil
	}

	hooksDir := filepath.Join(claude, "hooks")
	for _, name := range []string{"tolvi-recall", "tolvi-sync"} {
		dst := filepath.Join(hooksDir, name)
		if err := copyEmbedded(filepath.Join("files/hooks", name), dst); err != nil {
			return res, err
		}
		// A hook that is not executable cannot run, and the failure is silent.
		if err := os.Chmod(dst, 0o755); err != nil {
			return res, err
		}
		res.Written = append(res.Written, dst)
	}

	added, settingsPath, err := mergeHooks(claude, hooksDir)
	if err != nil {
		return res, err
	}
	res.HooksWired = true
	res.AllowAdded = added
	res.SettingsPath = settingsPath
	return res, nil
}

// mergeHooks merges the hook entries and the read-only allow rules into
// settings.json, adding what is missing and touching nothing else. Running it
// twice changes nothing, per
// vault/decisions/2026-08-03-installer-hook-merge-must-be-idempotent.
func mergeHooks(claudeDir, hooksDir string) ([]string, string, error) {
	settingsPath := filepath.Join(claudeDir, "settings.json")

	settings := map[string]any{}
	if raw, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(raw, &settings); err != nil {
			return nil, settingsPath, fmt.Errorf("%s is not valid JSON, so it will not be edited: %w", settingsPath, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, settingsPath, err
	}

	tmplRaw, err := Files.ReadFile("files/hooks.json")
	if err != nil {
		return nil, settingsPath, err
	}
	var tmpl struct {
		Hooks map[string][]any `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(string(tmplRaw), "__HOOKS_DIR__", hooksDir)), &tmpl); err != nil {
		return nil, settingsPath, err
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	for event, entries := range tmpl.Hooks {
		existing, _ := hooks[event].([]any)
		for _, entry := range entries {
			if !containsJSON(existing, entry) {
				existing = append(existing, entry)
			}
		}
		hooks[event] = existing
	}
	settings["hooks"] = hooks

	perms, _ := settings["permissions"].(map[string]any)
	if perms == nil {
		perms = map[string]any{}
	}
	allow, _ := perms["allow"].([]any)
	var added []string
	for _, rule := range readOnlyAllowRules {
		found := false
		for _, a := range allow {
			if s, ok := a.(string); ok && s == rule {
				found = true
				break
			}
		}
		if !found {
			allow = append(allow, rule)
			added = append(added, rule)
		}
	}
	perms["allow"] = allow
	settings["permissions"] = perms

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, settingsPath, err
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return nil, settingsPath, err
	}
	return added, settingsPath, os.WriteFile(settingsPath, append(out, '\n'), 0o644)
}

// containsJSON reports whether an equivalent entry is already present. Hook
// entries are nested maps, so equality is compared on their canonical JSON
// rather than with ==, which does not work on maps at all.
func containsJSON(haystack []any, needle any) bool {
	want, err := json.Marshal(needle)
	if err != nil {
		return false
	}
	for _, h := range haystack {
		got, err := json.Marshal(h)
		if err == nil && string(got) == string(want) {
			return true
		}
	}
	return false
}

func copyEmbedded(name, dst string) error {
	data, err := fs.ReadFile(Files, name)
	if err != nil {
		return fmt.Errorf("embedded %s: %w", name, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
