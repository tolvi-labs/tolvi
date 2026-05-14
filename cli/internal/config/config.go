// Package config resolves Tolvi CLI configuration from XDG config file
// and environment variables. Last writer wins: defaults → file → env →
// (flags applied by caller).
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultModel is the model used when no override is supplied.
const DefaultModel = "claude-sonnet-4-7"

// Config holds the resolved CLI configuration.
type Config struct {
	AnthropicAPIKey string `yaml:"anthropic_api_key"`
	Model           string `yaml:"model"`
	DefaultVault    string `yaml:"default_vault"`
}

// LoadOpts allows tests to inject HomeDir, ConfigPath, and a custom env
// reader. Production callers pass HomeDir = os.UserHomeDir() and
// Env = os.Getenv.
type LoadOpts struct {
	HomeDir    string
	ConfigPath string              // explicit path override (--config flag / TOLVI_CONFIG env)
	Env        func(string) string // env-var reader; required
}

// Load resolves the configuration. Never returns an error: missing files
// and unset env vars degrade gracefully to defaults.
func Load(opts LoadOpts) Config {
	cfg := Config{Model: DefaultModel}

	// 1. Determine config file path.
	cfgPath := opts.ConfigPath
	if cfgPath == "" {
		if v := opts.Env("TOLVI_CONFIG"); v != "" {
			cfgPath = v
		}
	}
	if cfgPath == "" {
		xdg := opts.Env("XDG_CONFIG_HOME")
		if xdg == "" {
			xdg = filepath.Join(opts.HomeDir, ".config")
		}
		cfgPath = filepath.Join(xdg, "tolvi", "config.yaml")
	}

	// 2. Read file (best-effort).
	if data, err := os.ReadFile(cfgPath); err == nil {
		var fileCfg Config
		_ = yaml.Unmarshal(data, &fileCfg)
		if fileCfg.AnthropicAPIKey != "" {
			cfg.AnthropicAPIKey = fileCfg.AnthropicAPIKey
		}
		if fileCfg.Model != "" {
			cfg.Model = fileCfg.Model
		}
		if fileCfg.DefaultVault != "" {
			cfg.DefaultVault = fileCfg.DefaultVault
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// Permission denied / read error. Silently fall through to env;
		// `tolvi ask` fails clearly later if no API key was provided.
		_ = err
	}

	// 3. Env overrides.
	if v := opts.Env("ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := opts.Env("TOLVI_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := opts.Env("TOLVI_VAULT"); v != "" {
		cfg.DefaultVault = v
	}

	return cfg
}
