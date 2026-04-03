package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL = "https://www.palebluedot.ai"
	DevBaseURL     = "https://dev-q7n1uvyi.palebluedot.top"
)

// env is injected by ldflags at build time: "prod" or "dev"
var env = "prod"

// IsDev returns true if running in development environment
func IsDev() bool {
	return env == "dev"
}

// GetBaseURL returns the appropriate base URL based on environment
func GetBaseURL() string {
	if IsDev() {
		return DevBaseURL
	}
	return DefaultBaseURL
}

type Config struct {
	BaseURL string `yaml:"base_url"`
	Cookie  string `yaml:"cookie"`
	UserID  int    `yaml:"user_id"`
	APIKey  string `yaml:"api_key,omitempty"`
}

// DefaultConfigPath returns the default path to the config file.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".pbd-cli", "config.yaml"), nil
}

// Load reads the config file at path, returning defaults if it doesn't exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				BaseURL: GetBaseURL(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = GetBaseURL()
	}

	return &cfg, nil
}

// Save writes the config to path with secure permissions (0600).
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// ClearSession removes session data from the config file.
func ClearSession(path string) error {
	cfg, err := Load(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Cookie = ""
	cfg.UserID = 0

	return Save(path, cfg)
}
