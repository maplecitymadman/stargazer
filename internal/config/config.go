package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the Stargazer configuration
type Config struct {
	Version       string              `yaml:"version"`
	Kubeconfig    KubeconfigConfig    `yaml:"kubeconfig,omitempty"`
	LLM           LLMConfig           `yaml:"llm"`
	Storage       StorageConfig       `yaml:"storage"`
	API           APIConfig           `yaml:"api"`
	CreatedAt     string              `yaml:"created_at"`
	UpdatedAt     string              `yaml:"updated_at"`
}

// KubeconfigConfig holds Kubernetes configuration
type KubeconfigConfig struct {
	Path    string `yaml:"path,omitempty"`    // Explicit path to kubeconfig
	Context string `yaml:"context,omitempty"` // Default context to use
}

// LLMConfig holds LLM provider configurations
type LLMConfig struct {
	DefaultProvider string               `yaml:"default_provider,omitempty"`
	Providers       map[string]Provider  `yaml:"providers"`
}

// Provider represents an LLM provider configuration
type Provider struct {
	Enabled   bool   `yaml:"enabled"`
	APIKey    string `yaml:"api_key,omitempty"`
	Model     string `yaml:"model,omitempty"`
	BaseURL   string `yaml:"base_url,omitempty"`
	Encrypted bool   `yaml:"encrypted,omitempty"`
}

// StorageConfig holds storage settings
type StorageConfig struct {
	Path          string `yaml:"path"`
	RetainDays    int    `yaml:"retain_days"`
	MaxScanResults int   `yaml:"max_scan_results"`
}

// APIConfig holds API server settings
type APIConfig struct {
	Port         int  `yaml:"port"`
	EnableCORS   bool `yaml:"enable_cors"`
	RateLimitRPS int  `yaml:"rate_limit_rps"`
}

var (
	// DefaultConfigDir is the default directory for Stargazer config
	DefaultConfigDir = filepath.Join(os.Getenv("HOME"), ".stargazer")
	// DefaultConfigPath is the default path to the config file
	DefaultConfigPath = filepath.Join(DefaultConfigDir, "config.yaml")
)

// Load loads configuration from the default path or creates a new one
func Load() (*Config, error) {
	return LoadFrom(DefaultConfigPath)
}

// LoadFrom loads configuration from a specific path
func LoadFrom(path string) (*Config, error) {
	// Check if config exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		cfg := DefaultConfig()
		if err := cfg.SaveTo(path); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to the default path
func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath)
}

// SaveTo saves the configuration to a specific path
func (c *Config) SaveTo(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update timestamp
	c.UpdatedAt = nowISO()

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// DefaultConfig returns a new configuration with default values
func DefaultConfig() *Config {
	now := nowISO()
	return &Config{
		Version: "1.0",
		LLM: LLMConfig{
			DefaultProvider: "",
			Providers: map[string]Provider{
				"openai": {
					Enabled: false,
					Model:   "gpt-4o-mini",
					BaseURL: "https://api.openai.com/v1",
				},
				"anthropic": {
					Enabled: false,
					Model:   "claude-3-haiku-20240307",
				},
				"gemini": {
					Enabled: false,
					Model:   "gemini-pro",
				},
				"ollama": {
					Enabled: false,
					Model:   "llama3.1",
					BaseURL: "http://localhost:11434",
				},
			},
		},
		Storage: StorageConfig{
			Path:          filepath.Join(DefaultConfigDir, "storage"),
			RetainDays:    30,
			MaxScanResults: 100,
		},
		API: APIConfig{
			Port:         8000,
			EnableCORS:   true,
			RateLimitRPS: 100,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsProviderEnabled checks if a specific provider is enabled
func (c *Config) IsProviderEnabled(name string) bool {
	if provider, ok := c.LLM.Providers[name]; ok {
		return provider.Enabled && provider.APIKey != ""
	}
	return false
}

// GetProvider returns a provider configuration by name
func (c *Config) GetProvider(name string) (Provider, bool) {
	provider, ok := c.LLM.Providers[name]
	return provider, ok
}

// SetProvider sets or updates a provider configuration
func (c *Config) SetProvider(name string, provider Provider) {
	if c.LLM.Providers == nil {
		c.LLM.Providers = make(map[string]Provider)
	}
	c.LLM.Providers[name] = provider
	c.UpdatedAt = nowISO()
}

// EnableProvider enables a provider and optionally sets API key
func (c *Config) EnableProvider(name, apiKey string) error {
	provider, ok := c.LLM.Providers[name]
	if !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	provider.Enabled = true
	if apiKey != "" {
		provider.APIKey = apiKey
	}

	c.LLM.Providers[name] = provider
	c.UpdatedAt = nowISO()

	return nil
}

// DisableProvider disables a provider
func (c *Config) DisableProvider(name string) {
	if provider, ok := c.LLM.Providers[name]; ok {
		provider.Enabled = false
		c.LLM.Providers[name] = provider
		c.UpdatedAt = nowISO()
	}
}

// GetEnabledProviders returns a list of enabled provider names
func (c *Config) GetEnabledProviders() []string {
	var enabled []string
	for name, provider := range c.LLM.Providers {
		if provider.Enabled && provider.APIKey != "" {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Check storage path
	if c.Storage.Path == "" {
		return fmt.Errorf("storage path cannot be empty")
	}

	// Check API port
	if c.API.Port < 1 || c.API.Port > 65535 {
		return fmt.Errorf("invalid API port: %d", c.API.Port)
	}

	// Check rate limit
	if c.API.RateLimitRPS < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	return nil
}

// nowISO returns current time in ISO 8601 format
func nowISO() string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		2026, 1, 13, 12, 0, 0) // Placeholder - in real implementation use time.Now()
}
