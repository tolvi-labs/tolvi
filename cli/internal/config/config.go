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

// RecallConfig controls the tolvi recall command. Zero values mean "use
// the compiled-in default" — RunRecall applies defaults at call time so
// the presence/absence of a config block is transparent to callers.
type RecallConfig struct {
	SessionCount    int  `yaml:"session_count"`    // default: 3
	DecisionCount   int  `yaml:"decision_count"`   // default: 10
	MaxBytes        int  `yaml:"max_bytes"`        // default: 8000 (0 = unlimited)
	IncludePatterns bool `yaml:"include_patterns"` // default: false
}

// Config holds the resolved CLI configuration.
type Config struct {
	AnthropicAPIKey string       `yaml:"anthropic_api_key"`
	Model           string       `yaml:"model"`
	DefaultVault    string       `yaml:"default_vault"`
	Recall          RecallConfig `yaml:"recall"`
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
		// RecallConfig: copy wholesale — zero values mean "use compiled-in
		// defaults", so copying a zero-valued struct from an absent section
		// is equivalent to not setting the field at all.
		cfg.Recall = fileCfg.Recall
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
