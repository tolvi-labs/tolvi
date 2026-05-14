package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFile_NoEnv(t *testing.T) {
	cfg := Load(LoadOpts{
		HomeDir:    t.TempDir(),
		ConfigPath: "",
		Env:        emptyEnv,
	})
	if cfg.AnthropicAPIKey != "" {
		t.Errorf("api key should be empty")
	}
	if cfg.Model != DefaultModel {
		t.Errorf("model = %q, want default %q", cfg.Model, DefaultModel)
	}
}

func TestLoad_FromFile(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, ".config", "tolvi")
	_ = os.MkdirAll(cfgDir, 0o755)
	yaml := `anthropic_api_key: sk-ant-from-file
model: claude-haiku-4-5-20251001
default_vault: /tmp/my-vault
`
	_ = os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644)

	cfg := Load(LoadOpts{HomeDir: home, Env: emptyEnv})
	if cfg.AnthropicAPIKey != "sk-ant-from-file" {
		t.Errorf("api key = %q", cfg.AnthropicAPIKey)
	}
	if cfg.Model != "claude-haiku-4-5-20251001" {
		t.Errorf("model = %q", cfg.Model)
	}
	if cfg.DefaultVault != "/tmp/my-vault" {
		t.Errorf("default_vault = %q", cfg.DefaultVault)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, ".config", "tolvi")
	_ = os.MkdirAll(cfgDir, 0o755)
	yaml := `anthropic_api_key: sk-ant-from-file
model: claude-haiku-4-5-20251001
`
	_ = os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yaml), 0o644)

	env := func(k string) string {
		switch k {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-from-env"
		case "TOLVI_MODEL":
			return "claude-sonnet-4-7"
		}
		return ""
	}
	cfg := Load(LoadOpts{HomeDir: home, Env: env})
	if cfg.AnthropicAPIKey != "sk-ant-from-env" {
		t.Errorf("env should override file: got %q", cfg.AnthropicAPIKey)
	}
	if cfg.Model != "claude-sonnet-4-7" {
		t.Errorf("env should override file: got %q", cfg.Model)
	}
}

func TestLoad_ExplicitConfigPath(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.yaml")
	_ = os.WriteFile(custom, []byte("anthropic_api_key: custom-key\n"), 0o644)

	cfg := Load(LoadOpts{HomeDir: t.TempDir(), ConfigPath: custom, Env: emptyEnv})
	if cfg.AnthropicAPIKey != "custom-key" {
		t.Errorf("api key = %q", cfg.AnthropicAPIKey)
	}
}

func TestLoad_XDGConfigHome(t *testing.T) {
	xdgHome := t.TempDir()
	cfgDir := filepath.Join(xdgHome, "tolvi")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "config.yaml"),
		[]byte("anthropic_api_key: xdg-path\n"), 0o644)

	env := func(k string) string {
		if k == "XDG_CONFIG_HOME" {
			return xdgHome
		}
		return ""
	}
	cfg := Load(LoadOpts{HomeDir: t.TempDir(), Env: env})
	if cfg.AnthropicAPIKey != "xdg-path" {
		t.Errorf("XDG path not used: got %q", cfg.AnthropicAPIKey)
	}
}

func emptyEnv(_ string) string { return "" }
