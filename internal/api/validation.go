package api

import (
	"fmt"
	"strings"
)

// ValidateNamespace validates a Kubernetes namespace parameter
// Returns nil if valid, error otherwise
func ValidateNamespace(ns string) error {
	// Empty or "all" are valid (means all namespaces)
	if ns == "" || ns == "all" {
		return nil
	}

	// Check length (DNS-1123 label max length)
	if len(ns) > 253 {
		return fmt.Errorf("namespace too long (max 253 characters)")
	}

	// Validate DNS-1123 format: lowercase alphanumeric or '-'
	// Must start and end with alphanumeric
	if len(ns) > 0 {
		if !isAlphanumeric(rune(ns[0])) {
			return fmt.Errorf("namespace must start with alphanumeric character")
		}
		if !isAlphanumeric(rune(ns[len(ns)-1])) {
			return fmt.Errorf("namespace must end with alphanumeric character")
		}
	}

	for _, char := range ns {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return fmt.Errorf("invalid namespace format: %s (must be lowercase alphanumeric or '-')", ns)
		}
	}

	return nil
}

// ValidateRequired validates that a required parameter is present
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateContext validates a Kubernetes context name
func ValidateContext(context string) error {
	if context == "" {
		return nil // Empty is valid (use current context)
	}

	// Context names can be more flexible than namespaces
	// Just check for reasonable length
	if len(context) > 253 {
		return fmt.Errorf("context name too long (max 253 characters)")
	}

	return nil
}

// isAlphanumeric checks if a rune is alphanumeric (lowercase)
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
