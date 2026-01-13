package k8s

import (
	"context"
	"fmt"
	"sync"
)

// Discovery scans Kubernetes resources for common issues
type Discovery struct {
	client *Client
}

// NewDiscovery creates a new Discovery engine
func NewDiscovery(client *Client) *Discovery {
	return &Discovery{
		client: client,
	}
}

// ScanAll scans all resources in the specified namespace for issues
// Runs multiple scans in parallel for efficiency
func (d *Discovery) ScanAll(ctx context.Context, namespace string) ([]Issue, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	allIssues := []Issue{}

	// Channel to collect errors
	errChan := make(chan error, 5)

	// Scan pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		issues, err := d.ScanPods(ctx, namespace)
		if err != nil {
			errChan <- fmt.Errorf("pod scan: %w", err)
			return
		}
		mu.Lock()
		allIssues = append(allIssues, issues...)
		mu.Unlock()
	}()

	// Scan deployments
	wg.Add(1)
	go func() {
		defer wg.Done()
		issues, err := d.ScanDeployments(ctx, namespace)
		if err != nil {
			errChan <- fmt.Errorf("deployment scan: %w", err)
			return
		}
		mu.Lock()
		allIssues = append(allIssues, issues...)
		mu.Unlock()
	}()

	// Scan events
	wg.Add(1)
	go func() {
		defer wg.Done()
		issues, err := d.ScanEvents(ctx, namespace)
		if err != nil {
			errChan <- fmt.Errorf("event scan: %w", err)
			return
		}
		mu.Lock()
		allIssues = append(allIssues, issues...)
		mu.Unlock()
	}()

	// Scan nodes (cluster-wide)
	wg.Add(1)
	go func() {
		defer wg.Done()
		issues, err := d.ScanNodes(ctx)
		if err != nil {
			errChan <- fmt.Errorf("node scan: %w", err)
			return
		}
		mu.Lock()
		allIssues = append(allIssues, issues...)
		mu.Unlock()
	}()

	// Wait for all scans to complete
	wg.Wait()
	close(errChan)

	// Collect any errors
	var scanErrors []error
	for err := range errChan {
		scanErrors = append(scanErrors, err)
	}

	// Return issues even if some scans failed
	if len(scanErrors) > 0 {
		return allIssues, fmt.Errorf("some scans failed: %v", scanErrors)
	}

	return allIssues, nil
}

// ScanPods scans pods for common issues
func (d *Discovery) ScanPods(ctx context.Context, namespace string) ([]Issue, error) {
	pods, err := d.client.GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var issues []Issue

	for _, pod := range pods {
		// Check pod status
		if pod.Status != "Running" && pod.Status != "Succeeded" {
			priority := PriorityWarning
			statusHint := ""

			switch pod.Status {
			case "Failed", "CrashLoopBackOff", "Error":
				priority = PriorityCritical
				statusHint = fmt.Sprintf(" Container is crashing. Check logs with: kubectl logs %s -n %s --previous",
					pod.Name, pod.Namespace)

			case "Pending":
				statusHint = fmt.Sprintf(" Pod cannot be scheduled. Check events with: kubectl get events -n %s --field-selector involvedObject.name=%s",
					pod.Namespace, pod.Name)

			case "ImagePullBackOff", "ErrImagePull":
				priority = PriorityCritical
				statusHint = " Cannot pull container image. Check image name and registry access."

			case "OOMKilled":
				priority = PriorityCritical
				statusHint = " Pod killed due to out of memory. Increase memory limits."
			}

			issues = append(issues, Issue{
				ID:           GenerateIssueID(pod.Name, "status"),
				Title:        fmt.Sprintf("Pod %s in %s state", pod.Name, pod.Status),
				Description:  fmt.Sprintf("Pod is in %s state instead of Running.%s", pod.Status, statusHint),
				Priority:     priority,
				ResourceType: "pod",
				ResourceName: pod.Name,
				Namespace:    pod.Namespace,
			})
		}

		// Check for container-specific issues
		for _, cs := range pod.ContainerStates {
			if cs.State == "waiting" && cs.Reason != "" {
				priority := PriorityWarning

				switch cs.Reason {
				case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull":
					priority = PriorityCritical
				}

				issues = append(issues, Issue{
					ID:          GenerateIssueID(pod.Name, fmt.Sprintf("container-%s", cs.Name)),
					Title:       fmt.Sprintf("Container %s in pod %s: %s", cs.Name, pod.Name, cs.Reason),
					Description: fmt.Sprintf("Container is waiting: %s. %s", cs.Reason, cs.Message),
					Priority:    priority,
					ResourceType: "pod",
					ResourceName: pod.Name,
					Namespace:    pod.Namespace,
				})
			}
		}

		// Check restart count
		if pod.Restarts > 5 {
			issues = append(issues, Issue{
				ID:    GenerateIssueID(pod.Name, "restarts"),
				Title: fmt.Sprintf("High restart count for %s", pod.Name),
				Description: fmt.Sprintf("Pod has restarted %d times. Check crash logs: kubectl logs %s -n %s --previous. Review events: kubectl get events -n %s --field-selector involvedObject.name=%s",
					pod.Restarts, pod.Name, pod.Namespace, pod.Namespace, pod.Name),
				Priority:     PriorityWarning,
				ResourceType: "pod",
				ResourceName: pod.Name,
				Namespace:    pod.Namespace,
			})
		}

		// Check readiness
		if !pod.Ready && pod.Status == "Running" {
			issues = append(issues, Issue{
				ID:          GenerateIssueID(pod.Name, "readiness"),
				Title:       fmt.Sprintf("Pod %s not ready", pod.Name),
				Description: fmt.Sprintf("Pod is running but readiness probe is failing. Check logs and probe configuration. Run: kubectl describe pod %s -n %s to see probe details.", pod.Name, pod.Namespace),
				Priority:    PriorityWarning,
				ResourceType: "pod",
				ResourceName: pod.Name,
				Namespace:    pod.Namespace,
			})
		}
	}

	return issues, nil
}

// ScanDeployments scans deployments for replica mismatches
func (d *Discovery) ScanDeployments(ctx context.Context, namespace string) ([]Issue, error) {
	deployments, err := d.client.GetDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var issues []Issue

	for _, deployment := range deployments {
		// Check replica mismatches
		if deployment.Replicas != deployment.ReadyReplicas {
			issues = append(issues, Issue{
				ID:          GenerateIssueID(deployment.Name, "replicas"),
				Title:       fmt.Sprintf("Replica mismatch in %s", deployment.Name),
				Description: fmt.Sprintf("Expected %d replicas, %d ready", deployment.Replicas, deployment.ReadyReplicas),
				Priority:    PriorityWarning,
				ResourceType: "deployment",
				ResourceName: deployment.Name,
				Namespace:    deployment.Namespace,
			})
		}

		// Check availability
		if deployment.AvailableReplicas < deployment.Replicas {
			priority := PriorityWarning
			if deployment.AvailableReplicas == 0 {
				priority = PriorityCritical
			}

			issues = append(issues, Issue{
				ID:          GenerateIssueID(deployment.Name, "availability"),
				Title:       fmt.Sprintf("Deployment %s unavailable", deployment.Name),
				Description: fmt.Sprintf("Only %d of %d replicas available", deployment.AvailableReplicas, deployment.Replicas),
				Priority:    priority,
				ResourceType: "deployment",
				ResourceName: deployment.Name,
				Namespace:    deployment.Namespace,
			})
		}
	}

	return issues, nil
}

// ScanEvents scans for Warning and Error events
func (d *Discovery) ScanEvents(ctx context.Context, namespace string) ([]Issue, error) {
	events, err := d.client.GetEvents(ctx, namespace, false) // Only non-Normal events
	if err != nil {
		return nil, err
	}

	var issues []Issue

	// Limit to recent events (e.g., last 10)
	maxEvents := 10
	if len(events) > maxEvents {
		events = events[:maxEvents]
	}

	for _, event := range events {
		priority := PriorityInfo
		if event.Type == "Warning" {
			priority = PriorityWarning
		} else if event.Type == "Error" {
			priority = PriorityCritical
		}

		issues = append(issues, Issue{
			ID:           GenerateIssueID(event.InvolvedObject, event.Reason),
			Title:        fmt.Sprintf("%s: %s", event.InvolvedObject, event.Reason),
			Description:  event.Message,
			Priority:     priority,
			ResourceType: event.InvolvedKind,
			ResourceName: event.InvolvedObject,
			Namespace:    event.InvolvedNamespace,
		})
	}

	return issues, nil
}

// ScanNodes scans nodes for common issues
func (d *Discovery) ScanNodes(ctx context.Context) ([]Issue, error) {
	nodes, err := d.client.GetNodes(ctx)
	if err != nil {
		return nil, err
	}

	var issues []Issue

	for _, node := range nodes {
		// Check node status
		if node.Status != "Ready" {
			issues = append(issues, Issue{
				ID:           GenerateIssueID(node.Name, "status"),
				Title:        fmt.Sprintf("Node %s is %s", node.Name, node.Status),
				Description:  fmt.Sprintf("Node is not in Ready state. Check: kubectl describe node %s", node.Name),
				Priority:     PriorityCritical,
				ResourceType: "node",
				ResourceName: node.Name,
				Namespace:    "", // Nodes are cluster-scoped
			})
		}
	}

	return issues, nil
}
