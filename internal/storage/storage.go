package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/maplecitymadman/stargazer/internal/k8s"
)

// Storage handles persistent storage of scan results and history
type Storage struct {
	basePath string
}

// ScanResult represents a stored scan result
type ScanResult struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Namespace string      `json:"namespace"`
	Issues    []k8s.Issue `json:"issues"`
	Summary   Summary     `json:"summary"`
}

// Summary provides a summary of the scan
type Summary struct {
	TotalIssues   int `json:"total_issues"`
	CriticalCount int `json:"critical_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
	PodsScanned   int `json:"pods_scanned"`
	NodesScanned  int `json:"nodes_scanned"`
}

// NewStorage creates a new storage instance
func NewStorage(basePath string) (*Storage, error) {
	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{
		basePath: basePath,
	}, nil
}

// SaveScanResult saves a scan result to storage
func (s *Storage) SaveScanResult(namespace string, issues []k8s.Issue) (*ScanResult, error) {
	// Generate ID from timestamp
	now := time.Now()
	id := fmt.Sprintf("scan-%d", now.Unix())

	// Calculate summary
	summary := calculateSummary(issues)

	result := &ScanResult{
		ID:        id,
		Timestamp: now,
		Namespace: namespace,
		Issues:    issues,
		Summary:   summary,
	}

	// Create filename
	filename := fmt.Sprintf("%s.json", id)
	path := filepath.Join(s.basePath, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scan result: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write scan result: %w", err)
	}

	return result, nil
}

// GetScanResult retrieves a scan result by ID
func (s *Storage) GetScanResult(id string) (*ScanResult, error) {
	path := filepath.Join(s.basePath, fmt.Sprintf("%s.json", id))

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("scan result not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read scan result: %w", err)
	}

	var result ScanResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse scan result: %w", err)
	}

	return &result, nil
}

// ListScanResults lists all scan results, sorted by timestamp (newest first)
func (s *Storage) ListScanResults(limit int) ([]ScanResult, error) {
	files, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var results []ScanResult

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.basePath, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip files that can't be read
		}

		var result ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue // Skip invalid files
		}

		results = append(results, result)
	}

	// Sort by timestamp (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// DeleteScanResult deletes a scan result by ID
func (s *Storage) DeleteScanResult(id string) error {
	path := filepath.Join(s.basePath, fmt.Sprintf("%s.json", id))

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("scan result not found: %s", id)
		}
		return fmt.Errorf("failed to delete scan result: %w", err)
	}

	return nil
}

// CleanupOld removes scan results older than the specified number of days
func (s *Storage) CleanupOld(days int) (int, error) {
	if days <= 0 {
		return 0, fmt.Errorf("days must be positive")
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	files, err := os.ReadDir(s.basePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read storage directory: %w", err)
	}

	deleted := 0

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.basePath, file.Name())

		// Try to parse timestamp from filename first to avoid IO
		ts, err := parseTimestampFromFilename(file.Name())
		if err == nil {
			if ts.Before(cutoff) {
				if err := os.Remove(path); err == nil {
					deleted++
				}
			}
			continue
		}

		// Fallback to reading file if filename parse fails
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var result ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		// Check if older than cutoff
		if result.Timestamp.Before(cutoff) {
			if err := os.Remove(path); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}

// parseTimestampFromFilename extracts the timestamp from a filename like "scan-1234567890.json"
func parseTimestampFromFilename(filename string) (time.Time, error) {
	// Expected format: scan-<unix-timestamp>.json
	name := strings.TrimSuffix(filename, ".json")
	if !strings.HasPrefix(name, "scan-") {
		return time.Time{}, fmt.Errorf("invalid filename format")
	}

	tsStr := strings.TrimPrefix(name, "scan-")
	tsInt, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(tsInt, 0), nil
}

// GetStats returns storage statistics
func (s *Storage) GetStats() (map[string]interface{}, error) {
	files, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	totalScans := 0
	totalIssues := 0
	var oldestScan, newestScan time.Time

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.basePath, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var result ScanResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		totalScans++
		totalIssues += result.Summary.TotalIssues

		if oldestScan.IsZero() || result.Timestamp.Before(oldestScan) {
			oldestScan = result.Timestamp
		}
		if newestScan.IsZero() || result.Timestamp.After(newestScan) {
			newestScan = result.Timestamp
		}
	}

	stats := map[string]interface{}{
		"total_scans":  totalScans,
		"total_issues": totalIssues,
		"storage_path": s.basePath,
	}

	if !oldestScan.IsZero() {
		stats["oldest_scan"] = oldestScan.Format(time.RFC3339)
	}
	if !newestScan.IsZero() {
		stats["newest_scan"] = newestScan.Format(time.RFC3339)
	}

	return stats, nil
}

// calculateSummary generates a summary from issues
func calculateSummary(issues []k8s.Issue) Summary {
	summary := Summary{
		TotalIssues: len(issues),
	}

	pods := make(map[string]bool)
	nodes := make(map[string]bool)

	for _, issue := range issues {
		switch issue.Priority {
		case k8s.PriorityCritical:
			summary.CriticalCount++
		case k8s.PriorityWarning:
			summary.WarningCount++
		default:
			summary.InfoCount++
		}

		// Track unique resources
		if issue.ResourceType == "pod" {
			pods[issue.Namespace+"/"+issue.ResourceName] = true
		} else if issue.ResourceType == "node" {
			nodes[issue.ResourceName] = true
		}
	}

	summary.PodsScanned = len(pods)
	summary.NodesScanned = len(nodes)

	return summary
}
