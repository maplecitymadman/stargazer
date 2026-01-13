package k8s

import (
	"crypto/sha256"
	"fmt"
)

// Priority represents the severity of an issue
type Priority string

const (
	PriorityCritical Priority = "CRITICAL"
	PriorityHigh     Priority = "HIGH"
	PriorityWarning  Priority = "WARNING"
	PriorityInfo     Priority = "INFO"
)

// Issue represents a discovered problem in the cluster
type Issue struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Priority     Priority `json:"priority"`
	ResourceType string   `json:"resource_type"`
	ResourceName string   `json:"resource_name"`
	Namespace    string   `json:"namespace"`
	Timestamp    string   `json:"timestamp,omitempty"`
}

// GenerateIssueID creates a deterministic ID for an issue
func GenerateIssueID(resourceName, issueType string) string {
	data := fmt.Sprintf("%s-%s", resourceName, issueType)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:8]) // First 8 bytes as hex
}
