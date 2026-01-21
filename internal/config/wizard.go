package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Wizard provides an interactive setup experience
type Wizard struct {
	config *Config
	reader *bufio.Reader
}

// NewWizard creates a new setup wizard
func NewWizard() *Wizard {
	return &Wizard{
		config: DefaultConfig(),
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run executes the setup wizard
func (w *Wizard) Run() (*Config, error) {
	fmt.Println("🧙 Stargazer Setup Wizard")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("This wizard will help you configure Stargazer.")
	fmt.Println("Press Enter to accept default values [in brackets].")
	fmt.Println()

	// API Configuration
	if err := w.setupAPI(); err != nil {
		return nil, err
	}

	// Storage Configuration
	if err := w.setupStorage(); err != nil {
		return nil, err
	}

	// Save configuration
	fmt.Println()
	fmt.Println("💾 Saving configuration...")
	if err := w.config.Save(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Configuration saved to: %s\n", DefaultConfigPath)
	fmt.Println()
	fmt.Println("🎉 Setup complete! You can now run:")
	fmt.Println("   stargazer web    # Start web interface")
	fmt.Println("   stargazer scan   # Scan cluster for issues")
	fmt.Println()

	return w.config, nil
}

func (w *Wizard) setupAPI() error {
	fmt.Println("📡 API Server Configuration")
	fmt.Println("---------------------------")

	// Port
	port, err := w.promptInt("API server port", w.config.API.Port)
	if err != nil {
		return err
	}
	w.config.API.Port = port

	// CORS
	cors, err := w.promptBool("Enable CORS (for development)", w.config.API.EnableCORS)
	if err != nil {
		return err
	}
	w.config.API.EnableCORS = cors

	// Rate limiting
	rateLimit, err := w.promptInt("Rate limit (requests per second)", w.config.API.RateLimitRPS)
	if err != nil {
		return err
	}
	w.config.API.RateLimitRPS = rateLimit

	fmt.Println()
	return nil
}

func (w *Wizard) setupStorage() error {
	fmt.Println("💾 Storage Configuration")
	fmt.Println("------------------------")

	// Storage path
	path, err := w.promptString("Storage directory", w.config.Storage.Path)
	if err != nil {
		return err
	}
	w.config.Storage.Path = path

	// Retention
	retainDays, err := w.promptInt("Retain scan results (days)", w.config.Storage.RetainDays)
	if err != nil {
		return err
	}
	w.config.Storage.RetainDays = retainDays

	// Max results
	maxResults, err := w.promptInt("Maximum stored scan results", w.config.Storage.MaxScanResults)
	if err != nil {
		return err
	}
	w.config.Storage.MaxScanResults = maxResults

	fmt.Println()
	return nil
}

// Prompt helpers

func (w *Wizard) promptString(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := w.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	return input, nil
}

func (w *Wizard) promptInt(prompt string, defaultValue int) (int, error) {
	defaultStr := fmt.Sprintf("%d", defaultValue)
	input, err := w.promptString(prompt, defaultStr)
	if err != nil {
		return 0, err
	}

	var value int
	if _, err := fmt.Sscanf(input, "%d", &value); err != nil {
		return defaultValue, nil
	}

	return value, nil
}

func (w *Wizard) promptBool(prompt string, defaultValue bool) (bool, error) {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	input, err := w.promptString(fmt.Sprintf("%s (y/n)", prompt), defaultStr)
	if err != nil {
		return false, err
	}

	input = strings.ToLower(input)
	return input == "y" || input == "yes", nil
}

// Fix Issue #8: Implement proper password masking using golang.org/x/term
func (w *Wizard) promptPassword(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)

	// Use terminal.ReadPassword to hide password input
	// This reads from stdin (file descriptor 0) and disables echo
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Print newline after password input
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(passwordBytes)), nil
}
