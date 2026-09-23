package integrations

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// The embedded tree is what makes a cask-installed binary able to install the
// integration at all: the cask ships one artifact, the binary, so anything the
// command writes has to be inside it.

func TestEmbedded_CarriesEveryFileTheSkillNeeds(t *testing.T) {
	for _, want := range []string{
		"SKILL.md",
		"hooks.json",
		"hooks/tolvi-recall",
		"hooks/tolvi-sync",
		"commands/tolvi-recall.md",
		"commands/tolvi-sync.md",
		"commands/tolvi-commit.md",
	} {
		if _, err := fs.Stat(Files, filepath.Join("files", want)); err != nil {
			t.Errorf("%s is not embedded: %v", want, err)
		}
	}
}

// go:embed skips names starting with _ unless the pattern says all:. The
// preflight source is deliberately underscore-prefixed, so a bare pattern
// would drop it silently.
func TestEmbedded_IncludesTheUnderscorePrefixedPreflight(t *testing.T) {
	if _, err := fs.Stat(Files, "files/commands/_preflight.md"); err != nil {
		t.Errorf("_preflight.md is not embedded, so the pattern lost its all: prefix: %v", err)
	}
}

func TestInstall_WritesTheSkillAndCommands(t *testing.T) {
	home := t.TempDir()
	res, err := Install(Opts{HomeDir: home})
	if err != nil {
		t.Fatal(err)
	}

	skill := filepath.Join(home, ".claude", "skills", "tolvi", "SKILL.md")
	if _, err := os.Stat(skill); err != nil {
		t.Errorf("SKILL.md not installed: %v", err)
	}
	for _, c := range []string{"tolvi-recall.md", "tolvi-sync.md", "tolvi-commit.md"} {
		if _, err := os.Stat(filepath.Join(home, ".claude", "commands", c)); err != nil {
			t.Errorf("%s not installed: %v", c, err)
		}
	}
	if len(res.Written) == 0 {
		t.Error("Install reported nothing written")
	}
}

// The preflight source is the drift check's canonical copy, not a command a
// user invokes, so installing it would put a non-command in the commands menu.
func TestInstall_DoesNotInstallThePreflightSourceAsACommand(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(Opts{HomeDir: home}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "commands", "_preflight.md")); err == nil {
		t.Error("_preflight.md was installed as a slash command")
	}
}

func TestInstall_RefusesToClobberWithoutForce(t *testing.T) {
	home := t.TempDir()
	skill := filepath.Join(home, ".claude", "skills", "tolvi", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o755); err != nil {
		t.Fatal(err)
	}
	mine := "# my customized skill\n"
	if err := os.WriteFile(skill, []byte(mine), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(Opts{HomeDir: home}); err == nil {
		t.Fatal("overwrote an existing install with no --force")
	}
	got, _ := os.ReadFile(skill)
	if string(got) != mine {
		t.Error("the refusal still modified the file")
	}

	if _, err := Install(Opts{HomeDir: home, Force: true}); err != nil {
		t.Fatalf("--force did not install: %v", err)
	}
	got, _ = os.ReadFile(skill)
	if string(got) == mine {
		t.Error("--force did not overwrite")
	}
}

func TestInstallHooks_MergesIntoExistingSettingsWithoutLosingThem(t *testing.T) {
	home := t.TempDir()
	settings := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `{"permissions":{"allow":["Bash(ls:*)"]},"hooks":{"Stop":[{"matcher":"","hooks":[{"type":"command","command":"say done"}]}]}}`
	if err := os.WriteFile(settings, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(Opts{HomeDir: home, WithHooks: true}); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	raw, _ := os.ReadFile(settings)
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("settings.json is no longer valid JSON: %v", err)
	}
	hooks := got["hooks"].(map[string]any)
	if _, ok := hooks["Stop"]; !ok {
		t.Error("merging dropped an unrelated hook")
	}
	if _, ok := hooks["SessionStart"]; !ok {
		t.Error("SessionStart hook not installed")
	}
	allow := got["permissions"].(map[string]any)["allow"].([]any)
	var sawLs, sawRecall bool
	for _, a := range allow {
		switch a.(string) {
		case "Bash(ls:*)":
			sawLs = true
		case "Bash(tolvi recall:*)":
			sawRecall = true
		}
	}
	if !sawLs {
		t.Error("merging dropped an unrelated allow rule")
	}
	if !sawRecall {
		t.Error("the recall allow rule was not added")
	}
}

// Per 2026-08-03: running the installer twice must not duplicate entries.
func TestInstallHooks_IsIdempotent(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(Opts{HomeDir: home, WithHooks: true}); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))

	if _, err := Install(Opts{HomeDir: home, WithHooks: true, Force: true}); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))

	if string(first) != string(second) {
		t.Errorf("a second install changed settings.json:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// Writes stay a conscious per-call approval, as solo's installer decided.
func TestInstallHooks_AllowlistsReadsOnly(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(Opts{HomeDir: home, WithHooks: true}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	for _, forbidden := range []string{"tolvi sync", "tolvi commit", "tolvi init"} {
		if contains(string(raw), forbidden) {
			t.Errorf("allowlist contains a write command: %q", forbidden)
		}
	}
}

func TestInstallHooks_HookCommandsPointAtInstalledFiles(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(Opts{HomeDir: home, WithHooks: true}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if contains(string(raw), "__HOOKS_DIR__") {
		t.Error("the hooks-dir placeholder was written to settings.json unsubstituted")
	}
	hook := filepath.Join(home, ".claude", "hooks", "tolvi-recall")
	info, err := os.Stat(hook)
	if err != nil {
		t.Fatalf("hook script not installed: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Error("hook script is not executable, so the hook cannot run")
	}
	if !contains(string(raw), hook) {
		t.Errorf("settings.json does not point at the installed hook %s:\n%s", hook, raw)
	}
}

func TestInstall_WithoutHooksLeavesSettingsAlone(t *testing.T) {
	home := t.TempDir()
	if _, err := Install(Opts{HomeDir: home}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Error("a plain install wrote settings.json")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
