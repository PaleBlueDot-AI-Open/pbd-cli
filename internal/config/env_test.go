package config

import (
	"strings"
	"testing"
)

func TestGetBaseURL(t *testing.T) {
	url := GetBaseURL()
	if url == "" {
		t.Error("GetBaseURL returned empty string")
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("GetBaseURL should return HTTPS URL, got: %s", url)
	}
}

func TestIsDev(t *testing.T) {
	// This test verifies the function works
	// The actual value depends on build-time ldflags injection
	_ = IsDev() // Just ensure it doesn't panic
}
