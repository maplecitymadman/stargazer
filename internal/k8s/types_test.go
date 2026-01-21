package k8s

import (
	"testing"
	"time"
)

func TestGenerateIssueID(t *testing.T) {
	// Same inputs should produce same ID
	id1 := GenerateIssueID("pod-1", "status")
	id2 := GenerateIssueID("pod-1", "status")

	if id1 != id2 {
		t.Errorf("Same inputs should produce same ID, got %s and %s", id1, id2)
	}

	// Different inputs should produce different IDs
	id3 := GenerateIssueID("pod-2", "status")
	if id1 == id3 {
		t.Error("Different inputs should produce different IDs")
	}

	id4 := GenerateIssueID("pod-1", "restarts")
	if id1 == id4 {
		t.Error("Different issues for same resource should produce different IDs")
	}

	// ID should be deterministic and not empty
	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	// ID should be reasonable length (first 8 bytes of SHA256 = 16 hex chars)
	if len(id1) != 16 {
		t.Errorf("Expected ID length 16, got %d", len(id1))
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityCritical, "CRITICAL"},
		{PriorityWarning, "WARNING"},
		{PriorityInfo, "INFO"},
	}

	for _, tt := range tests {
		if string(tt.priority) != tt.expected {
			t.Errorf("Expected priority string %s, got %s", tt.expected, string(tt.priority))
		}
	}
}

func TestIssueStructure(t *testing.T) {
	issue := Issue{
		ID:           "test-id",
		Title:        "Test Issue",
		Description:  "Test description",
		Priority:     PriorityCritical,
		ResourceType: "pod",
		ResourceName: "test-pod",
		Namespace:    "default",
	}

	if issue.ID != "test-id" {
		t.Errorf("Expected ID test-id, got %s", issue.ID)
	}

	if issue.Priority != PriorityCritical {
		t.Errorf("Expected priority critical, got %s", issue.Priority)
	}

	if issue.ResourceType != "pod" {
		t.Errorf("Expected resource type pod, got %s", issue.ResourceType)
	}
}

func TestPodStructure(t *testing.T) {
	pod := Pod{
		Name:      "test-pod",
		Namespace: "default",
		Status:    "Running",
		Ready:     true,
		Restarts:  5,
		Age:       "2d",
	}

	if pod.Name != "test-pod" {
		t.Errorf("Expected name test-pod, got %s", pod.Name)
	}

	if !pod.Ready {
		t.Error("Pod should be ready")
	}

	if pod.Restarts != 5 {
		t.Errorf("Expected 5 restarts, got %d", pod.Restarts)
	}
}

func TestDeploymentStructure(t *testing.T) {
	deployment := Deployment{
		Name:              "test-deployment",
		Namespace:         "default",
		Replicas:          3,
		ReadyReplicas:     2,
		AvailableReplicas: 2,
		Age:               "5d",
	}

	if deployment.Replicas != 3 {
		t.Errorf("Expected 3 replicas, got %d", deployment.Replicas)
	}

	if deployment.ReadyReplicas != 2 {
		t.Errorf("Expected 2 ready replicas, got %d", deployment.ReadyReplicas)
	}

	// Test replica mismatch detection
	if deployment.Replicas == deployment.ReadyReplicas {
		t.Error("This deployment should have a replica mismatch")
	}
}

func TestNodeStructure(t *testing.T) {
	node := Node{
		Name:   "node-1",
		Status: "Ready",
		Roles:  []string{"master", "control-plane"},
		Age:    "30d",
	}

	if node.Name != "node-1" {
		t.Errorf("Expected name node-1, got %s", node.Name)
	}

	if node.Status != "Ready" {
		t.Errorf("Expected status Ready, got %s", node.Status)
	}

	if len(node.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(node.Roles))
	}
}

func TestServiceStructure(t *testing.T) {
	service := Service{
		Name:        "test-service",
		Namespace:   "default",
		Type:        "ClusterIP",
		ClusterIP:   "10.0.0.1",
		ExternalIP:  "",
		Ports:       []ServicePort{{Protocol: "TCP", Port: 80, TargetPort: "8080"}},
		Age:         "10d",
		Annotations: map[string]string{"key": "value"},
		Labels:      map[string]string{"app": "test"},
	}

	if service.Type != "ClusterIP" {
		t.Errorf("Expected type ClusterIP, got %s", service.Type)
	}

	if len(service.Ports) != 1 {
		t.Errorf("Expected 1 port, got %d", len(service.Ports))
	}

	if service.Labels["app"] != "test" {
		t.Errorf("Expected label app=test, got %s", service.Labels["app"])
	}
}

func TestEventStructure(t *testing.T) {
	now := time.Now()
	event := Event{
		Type:              "Warning",
		Reason:            "BackOff",
		Message:           "Back-off restarting failed container",
		InvolvedObject:    "pod/test-pod",
		InvolvedKind:      "Pod",
		InvolvedNamespace: "default",
		FirstTimestamp:    now,
		LastTimestamp:     now.Add(5 * time.Minute),
		Count:             5,
	}

	if event.Type != "Warning" {
		t.Errorf("Expected type Warning, got %s", event.Type)
	}

	if event.Count != 5 {
		t.Errorf("Expected count 5, got %d", event.Count)
	}

	if event.Reason != "BackOff" {
		t.Errorf("Expected reason BackOff, got %s", event.Reason)
	}
}

func TestContainerStateStructure(t *testing.T) {
	state := ContainerState{
		Name:    "app",
		State:   "running",
		Ready:   true,
		Reason:  "",
		Message: "",
	}

	if state.Name != "app" {
		t.Errorf("Expected name app, got %s", state.Name)
	}

	if !state.Ready {
		t.Error("Container should be ready")
	}

	// Test waiting state
	waitingState := ContainerState{
		Name:    "app",
		State:   "waiting",
		Ready:   false,
		Reason:  "CrashLoopBackOff",
		Message: "Container crashed",
	}

	if waitingState.Ready {
		t.Error("Waiting container should not be ready")
	}

	if waitingState.Reason != "CrashLoopBackOff" {
		t.Errorf("Expected reason CrashLoopBackOff, got %s", waitingState.Reason)
	}
}
