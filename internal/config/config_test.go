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

	// Test LLM providers exist
	expectedProviders := []string{"openai", "anthropic", "gemini", "ollama"}
	for _, name := range expectedProviders {
		if _, exists := cfg.LLM.Providers[name]; !exists {
			t.Errorf("Expected provider %s to exist", name)
		}
	}

	// All providers should be disabled by default
	for name, provider := range cfg.LLM.Providers {
		if provider.Enabled {
			t.Errorf("Provider %s should be disabled by default", name)
		}
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
	cfg.LLM.DefaultProvider = "openai"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if loaded.API.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", loaded.API.Port)
	}
	if loaded.LLM.DefaultProvider != "openai" {
		t.Errorf("Expected default provider openai, got %s", loaded.LLM.DefaultProvider)
	}
}

func TestEnableProvider(t *testing.T) {
	cfg := DefaultConfig()

	// Enable OpenAI with API key
	err := cfg.EnableProvider("openai", "test-api-key")
	if err != nil {
		t.Fatalf("Failed to enable provider: %v", err)
	}

	provider := cfg.LLM.Providers["openai"]
	if !provider.Enabled {
		t.Error("Provider should be enabled")
	}
	if provider.APIKey != "test-api-key" {
		t.Errorf("Expected API key test-api-key, got %s", provider.APIKey)
	}

	// Try to enable non-existent provider
	err = cfg.EnableProvider("nonexistent", "key")
	if err == nil {
		t.Error("Expected error when enabling non-existent provider")
	}
}

func TestDisableProvider(t *testing.T) {
	cfg := DefaultConfig()

	// Enable then disable
	cfg.EnableProvider("openai", "test-key")
	cfg.DisableProvider("openai")

	provider := cfg.LLM.Providers["openai"]
	if provider.Enabled {
		t.Error("Provider should be disabled")
	}

	// Disabling non-existent provider should not panic (it's a no-op)
	cfg.DisableProvider("nonexistent")
}

func TestGetEnabledProviders(t *testing.T) {
	cfg := DefaultConfig()

	// Initially no enabled providers
	enabled := cfg.GetEnabledProviders()
	if len(enabled) != 0 {
		t.Errorf("Expected 0 enabled providers, got %d", len(enabled))
	}

	// Enable some providers
	cfg.EnableProvider("openai", "key1")
	cfg.EnableProvider("anthropic", "key2")

	enabled = cfg.GetEnabledProviders()
	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled providers, got %d", len(enabled))
	}

	// Check if both are in the list
	foundOpenAI := false
	foundAnthropic := false
	for _, name := range enabled {
		if name == "openai" {
			foundOpenAI = true
		}
		if name == "anthropic" {
			foundAnthropic = true
		}
	}

	if !foundOpenAI {
		t.Error("openai should be in enabled providers")
	}
	if !foundAnthropic {
		t.Error("anthropic should be in enabled providers")
	}
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
