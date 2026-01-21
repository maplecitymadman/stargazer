package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test version
	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}

	// Test API defaults
	if cfg.API.Port != 8000 {
		t.Errorf("Expected port 8000, got %d", cfg.API.Port)
	}
	if !cfg.API.EnableCORS {
		t.Error("Expected CORS to be enabled by default")
	}
	if cfg.API.RateLimitRPS != 100 {
		t.Errorf("Expected rate limit 100, got %d", cfg.API.RateLimitRPS)
	}

	// Test storage defaults
	if cfg.Storage.RetainDays != 30 {
		t.Errorf("Expected retain days 30, got %d", cfg.Storage.RetainDays)
	}
	if cfg.Storage.MaxScanResults != 100 {
		t.Errorf("Expected max results 100, got %d", cfg.Storage.MaxScanResults)
	}

}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "stargazer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override default path for testing
	originalPath := DefaultConfigPath
	DefaultConfigPath = filepath.Join(tmpDir, "config.yaml")
	defer func() { DefaultConfigPath = originalPath }()

	// Create and save config
	cfg := DefaultConfig()
	cfg.API.Port = 9000

}

func TestConfigValidation(t *testing.T) {
	cfg := DefaultConfig()

	// Valid config should pass
	if err := cfg.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}

	// Invalid port should fail
	cfg.API.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for port 0")
	}

	cfg.API.Port = 70000
	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for port > 65535")
	}

	// Reset port
	cfg.API.Port = 8000

	// Invalid storage path should fail
	cfg.Storage.Path = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Expected validation error for empty storage path")
	}
}

func TestLoadNonExistentConfig(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "stargazer-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override path to non-existent file in temp directory
	originalPath := DefaultConfigPath
	DefaultConfigPath = filepath.Join(tmpDir, "config.yaml")
	defer func() { DefaultConfigPath = originalPath }()

	// Load() creates default config if not exists
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should create default config, got error: %v", err)
	}

	// Verify default config was created
	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}

	// Verify file was created
	if _, err := os.Stat(DefaultConfigPath); os.IsNotExist(err) {
		t.Error("Config file should have been created")
	}
}
