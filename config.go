package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	defaultLimit    = 200
	defaultProvider = "openai"
	defaultModel    = "gpt-4o-mini"
)

// ProviderConfig holds provider-specific configuration.
type ProviderConfig struct {
	Type    string `json:"type"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// Config holds the full application configuration.
type Config struct {
	Limit    int            `json:"limit"`
	Provider ProviderConfig `json:"provider"`
}

// configPath returns the path to the config file.
func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "q", "config.json")
}

// loadConfig loads configuration from the config file.
// Returns an error if the file doesn't exist and no env vars can compensate.
func loadConfig() (*Config, error) {
	path := configPath()
	if path == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}

	cfg := &Config{
		Limit: defaultLimit,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if env vars can provide enough config
			if os.Getenv("Q_PROVIDER") != "" && os.Getenv("Q_MODEL") != "" {
				// Env vars will fill in the rest
				return cfg, nil
			}
			return nil, fmt.Errorf("no configuration found. Run \"q init\" to set up your config at %s", path)
		}
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %v", err)
	}

	// Apply default limit if not specified in file
	if cfg.Limit == 0 {
		// Distinguish between "limit: 0" (explicit unlimited) and "limit not set"
		// We need to check if the key was present in the JSON
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err == nil {
			if _, exists := raw["limit"]; !exists {
				cfg.Limit = defaultLimit
			}
		}
	}

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("Q_API_KEY"); v != "" {
		cfg.Provider.APIKey = v
	}
	if v := os.Getenv("Q_MODEL"); v != "" {
		cfg.Provider.Model = v
	}
	if v := os.Getenv("Q_PROVIDER"); v != "" {
		cfg.Provider.Type = v
	}
	if v := os.Getenv("Q_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.Limit = n
		}
	}
}

// resolveDefaults fills in default values for unset config fields.
func resolveDefaults(cfg *Config) {
	if cfg.Provider.Type == "" {
		cfg.Provider.Type = defaultProvider
	}
	if cfg.Provider.Model == "" {
		cfg.Provider.Model = defaultModel
	}
	if cfg.Provider.BaseURL == "" {
		switch cfg.Provider.Type {
		case "ollama":
			cfg.Provider.BaseURL = "http://localhost:11434/v1"
		default: // "openai" and anything else
			cfg.Provider.BaseURL = "https://api.openai.com/v1"
		}
	}
}

// validateConfig checks that required config values are present.
func validateConfig(cfg *Config) error {
	if cfg.Provider.Type == "openai" && cfg.Provider.APIKey == "" {
		return fmt.Errorf("API key required for OpenAI provider. Set it in config or via Q_API_KEY env var")
	}
	return nil
}
