package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maplecitymadman/stargazer/internal/k8s"
)

func TestNewStorage(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "stargazer-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Check that base directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Base storage directory was not created")
	}

	if storage.basePath != tmpDir {
		t.Errorf("Expected base path %s, got %s", tmpDir, storage.basePath)
	}
}

func TestSaveAndGetScanResult(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stargazer-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create test issues
	issues := []k8s.Issue{
		{
			ID:           "test-1",
			Title:        "Test Issue 1",
			Description:  "Test description",
			Priority:     k8s.PriorityCritical,
			ResourceType: "pod",
			ResourceName: "test-pod",
			Namespace:    "default",
		},
		{
			ID:           "test-2",
			Title:        "Test Issue 2",
			Description:  "Another test",
			Priority:     k8s.PriorityWarning,
			ResourceType: "deployment",
			ResourceName: "test-deployment",
			Namespace:    "default",
		},
	}

	// Save scan result
	result, err := storage.SaveScanResult("default", issues)
	if err != nil {
		t.Fatalf("Failed to save scan result: %v", err)
	}

	if result.ID == "" {
		t.Error("Result ID should not be empty")
	}

	if result.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", result.Namespace)
	}

	if len(result.Issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(result.Issues))
	}

	// Check summary
	if result.Summary.TotalIssues != 2 {
		t.Errorf("Expected total 2, got %d", result.Summary.TotalIssues)
	}
	if result.Summary.CriticalCount != 1 {
		t.Errorf("Expected 1 critical, got %d", result.Summary.CriticalCount)
	}
	if result.Summary.WarningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", result.Summary.WarningCount)
	}

	// Get the saved result
	retrieved, err := storage.GetScanResult(result.ID)
	if err != nil {
		t.Fatalf("Failed to get scan result: %v", err)
	}

	if retrieved.ID != result.ID {
		t.Errorf("Expected ID %s, got %s", result.ID, retrieved.ID)
	}

	if len(retrieved.Issues) != 2 {
		t.Errorf("Expected 2 issues in retrieved result, got %d", len(retrieved.Issues))
	}
}

func TestListScanResults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stargazer-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Save multiple scan results
	issues := []k8s.Issue{
		{
			ID:           "test-1",
			Title:        "Test Issue",
			Priority:     k8s.PriorityWarning,
			ResourceType: "pod",
			ResourceName: "test-pod",
			Namespace:    "default",
		},
	}

	result1, err1 := storage.SaveScanResult("default", issues)
	if err1 != nil {
		t.Fatal(err1)
	}
	time.Sleep(1100 * time.Millisecond) // Ensure different Unix timestamps (1+ second apart)

	result2, err2 := storage.SaveScanResult("kube-system", issues)
	if err2 != nil {
		t.Fatal(err2)
	}
	time.Sleep(1100 * time.Millisecond)

	result3, err3 := storage.SaveScanResult("default", issues)
	if err3 != nil {
		t.Fatal(err3)
	}
	time.Sleep(100 * time.Millisecond) // Give filesystem time to sync

	// List all results
	results, err := storage.ListScanResults(0)
	if err != nil {
		t.Fatalf("Failed to list scan results: %v", err)
	}

	// Check we have at least the 3 we just created
	if len(results) < 3 {
		t.Errorf("Expected at least 3 results, got %d", len(results))
	}

	// Verify our results are in the list
	foundIDs := make(map[string]bool)
	for _, r := range results {
		foundIDs[r.ID] = true
	}

	if !foundIDs[result1.ID] || !foundIDs[result2.ID] || !foundIDs[result3.ID] {
		t.Error("Not all saved results were found in list")
	}

	// Test limit
	limitedResults, err := storage.ListScanResults(2)
	if err != nil {
		t.Fatalf("Failed to list limited scan results: %v", err)
	}

	if len(limitedResults) > 2 {
		t.Errorf("Expected max 2 results with limit, got %d", len(limitedResults))
	}
}

func TestGetNonExistentScanResult(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stargazer-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.GetScanResult("nonexistent-id")
	if err == nil {
		t.Error("Expected error when getting non-existent result")
	}
}

func TestCleanupOld(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stargazer-cleanup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Helper to create a file with specific timestamp
	createFile := func(timestamp time.Time, nameOverride string) string {
		id := fmt.Sprintf("scan-%d", timestamp.Unix())
		if nameOverride != "" {
			id = nameOverride
		}

		result := &ScanResult{
			ID:        id,
			Timestamp: timestamp,
			Namespace: "default",
			Issues:    []k8s.Issue{},
		}

		filename := fmt.Sprintf("%s.json", id)
		path := filepath.Join(tmpDir, filename)
		data, _ := json.Marshal(result)
		os.WriteFile(path, data, 0644)
		return filename
	}

	now := time.Now()

	// File 1: Old enough to be deleted (using filename)
	oldTime := now.AddDate(0, 0, -10) // 10 days old
	createFile(oldTime, "")

	// File 2: Recent, should NOT be deleted
	recentTime := now.AddDate(0, 0, -1) // 1 day old
	recentFile := createFile(recentTime, "")

	// File 3: Manually named file with old timestamp (fallback to parsing content)
	// We name it "custom-old-scan" so filename parser fails
	createFile(oldTime, "custom-old-scan")

	// Set retention directly to 5 days
	days := 5
	deleted, err := storage.CleanupOld(days)
	if err != nil {
		t.Fatalf("CleanupOld failed: %v", err)
	}

	if deleted != 2 {
		t.Errorf("Expected 2 files deleted, got %d", deleted)
	}

	// Verify what remains
	files, _ := os.ReadDir(tmpDir)
	if len(files) != 1 {
		t.Errorf("Expected 1 file remaining, got %d", len(files))
	}

	// Check that the recent file is the one remaining
	if _, err := os.Stat(filepath.Join(tmpDir, recentFile)); os.IsNotExist(err) {
		t.Error("Recent file was deleted")
	}
}

func TestDeleteScanResult(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stargazer-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := storage.SaveScanResult("default", []k8s.Issue{})
	if err != nil {
		t.Fatal(err)
	}

	// Delete
	err = storage.DeleteScanResult(result.ID)
	if err != nil {
		t.Fatalf("Failed to delete scan result: %v", err)
	}

	// Verify it's gone
	_, err = storage.GetScanResult(result.ID)
	if err == nil {
		t.Error("Scan result still exists after deletion")
	}

	// Delete non-existent
	err = storage.DeleteScanResult("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent result")
	}
}

func TestGetStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stargazer-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage, err := NewStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Stats for empty storage
	stats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats["total_scans"].(int) != 0 {
		t.Errorf("Expected 0 scans, got %v", stats["total_scans"])
	}

	// Add a result
	time.Sleep(1 * time.Second)
	_, err = storage.SaveScanResult("default", []k8s.Issue{{ID: "1", Priority: k8s.PriorityCritical}})
	if err != nil {
		t.Fatal(err)
	}

	stats, err = storage.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats["total_scans"].(int) != 1 {
		t.Errorf("Expected 1 scan, got %v", stats["total_scans"])
	}
	if stats["total_issues"].(int) != 1 {
		t.Errorf("Expected 1 issue, got %v", stats["total_issues"])
	}
	if _, ok := stats["oldest_scan"]; !ok {
		t.Error("Expected oldest_scan in stats")
	}
}

func TestStorageErrorCases(t *testing.T) {
	// Test NewStorage with invalid path
	_, err := NewStorage("/proc/invalid") // Path that should fail
	if err == nil {
		t.Error("Expected error for invalid storage path")
	}

	tmpDir, _ := os.MkdirTemp("", "stargazer-storage-test-*")
	defer os.RemoveAll(tmpDir)
	storage, _ := NewStorage(tmpDir)

	// Test CleanupOld with non-positive days
	_, err = storage.CleanupOld(0)
	if err == nil {
		t.Error("Expected error for CleanupOld with 0 days")
	}
}
