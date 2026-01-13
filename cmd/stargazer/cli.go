package main

import (
	"fmt"
	"os"
)

// RunCLI is the entry point for CLI mode
// This is called when the binary is run with CLI commands
func RunCLI() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
