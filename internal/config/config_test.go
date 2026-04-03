package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".pbd-cli", "config.yaml")
	got, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() error = %v", err)
	}
	if got != expected {
		t.Errorf("DefaultConfigPath() = %s, want %s", got, expected)
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		BaseURL: "https://example.com",
		Cookie:  "test-session",
		UserID:  123,
	}

	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, _ := os.Stat(configPath)
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %s, want %s", loaded.BaseURL, cfg.BaseURL)
	}
}

func TestLoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BaseURL != GetBaseURL() {
		t.Errorf("default BaseURL = %s, want %s", cfg.BaseURL, GetBaseURL())
	}
}

func TestClearSession(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{BaseURL: "https://example.com", Cookie: "test", UserID: 123}
	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := ClearSession(configPath); err != nil {
		t.Fatalf("ClearSession() error = %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Cookie != "" {
		t.Errorf("Session should be cleared, got %q", loaded.Cookie)
	}
	if loaded.UserID != 0 {
		t.Errorf("UserID should be 0, got %d", loaded.UserID)
	}
}
