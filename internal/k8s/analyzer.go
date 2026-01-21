package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AnalysisResult represents the result of a troubleshooting analysis
type AnalysisResult struct {
	Resource        string         `json:"resource"`
	Type            string         `json:"type"`
	Namespace       string         `json:"namespace"`
	Status          string         `json:"status"`
	Issues          []Issue        `json:"issues"`
	Recommendations []string       `json:"recommendations"`
	Events          []corev1.Event `json:"events,omitempty"`
}

// Troubleshoot analyzes a resource and provides automated troubleshooting insights
func (c *Client) Troubleshoot(ctx context.Context, resourceType, name, namespace string) (*AnalysisResult, error) {
	if namespace == "" {
		namespace = c.GetNamespace()
	}

	result := &AnalysisResult{
		Resource:        name,
		Type:            resourceType,
		Namespace:       namespace,
		Issues:          []Issue{},
		Recommendations: []string{},
	}

	switch strings.ToLower(resourceType) {
	case "pod":
		return c.troubleshootPod(ctx, name, namespace, result)
	case "service":
		return c.troubleshootService(ctx, name, namespace, result)
	case "deployment":
		return c.troubleshootDeployment(ctx, name, namespace, result)
	default:
		return nil, fmt.Errorf("troubleshooting not supported for resource type: %s", resourceType)
	}
}

func (c *Client) troubleshootPod(ctx context.Context, name, namespace string, result *AnalysisResult) (*AnalysisResult, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	result.Status = string(pod.Status.Phase)

	// Fetch events for this pod
	events, _ := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", name),
	})
	result.Events = events.Items

	// Analyze container status
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil {
			reason := status.State.Waiting.Reason
			message := status.State.Waiting.Message

			issue := Issue{
				ID:           GenerateIssueID(name, "container-waiting-"+reason),
				Title:        fmt.Sprintf("Container %s is waiting: %s", status.Name, reason),
				Description:  message,
				Priority:     PriorityHigh,
				ResourceType: "Pod",
				ResourceName: name,
				Namespace:    namespace,
			}

			if reason == "CrashLoopBackOff" {
				issue.Priority = PriorityCritical
				result.Recommendations = append(result.Recommendations,
					fmt.Sprintf("Check logs for container %s: kubectl logs %s -c %s", status.Name, name, status.Name),
					"Check if the container entrypoint is correct",
					"Verify required environment variables and secrets are present",
				)
			} else if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
				issue.Priority = PriorityCritical
				result.Recommendations = append(result.Recommendations,
					"Verify the image name and tag are correct",
					"Check image pull secrets (docker-registry secrets)",
					"Verify the registry is accessible from the cluster nodes",
				)
			}
			result.Issues = append(result.Issues, issue)
		}

		if status.State.Terminated != nil && status.State.Terminated.ExitCode != 0 {
			reason := status.State.Terminated.Reason
			exitCode := status.State.Terminated.ExitCode

			issue := Issue{
				ID:           GenerateIssueID(name, "container-terminated"),
				Title:        fmt.Sprintf("Container %s terminated with exit code %d", status.Name, exitCode),
				Description:  fmt.Sprintf("Reason: %s. Message: %s", reason, status.State.Terminated.Message),
				Priority:     PriorityHigh,
				ResourceType: "Pod",
				ResourceName: name,
				Namespace:    namespace,
			}

			if reason == "OOMKilled" {
				issue.Priority = PriorityCritical
				result.Recommendations = append(result.Recommendations,
					"Increase memory limits for the container",
					"Investigate memory leaks in the application",
				)
			}
			result.Issues = append(result.Issues, issue)
		}
	}

	// Analyze pod conditions
	for _, condition := range pod.Status.Conditions {
		if condition.Status == corev1.ConditionFalse {
			if condition.Type == corev1.PodScheduled {
				result.Issues = append(result.Issues, Issue{
					ID:           GenerateIssueID(name, "unschedulable"),
					Title:        "Pod is unschedulable",
					Description:  condition.Message,
					Priority:     PriorityCritical,
					ResourceType: "Pod",
					ResourceName: name,
					Namespace:    namespace,
				})
				result.Recommendations = append(result.Recommendations,
					"Check node resource availability (CPU/Memory)",
					"Verify pod anti-affinity or taint/toleration rules",
				)
			}
		}
	}

	return result, nil
}

func (c *Client) troubleshootService(ctx context.Context, name, namespace string, result *AnalysisResult) (*AnalysisResult, error) {
	svc, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	result.Status = "Active"

	// Check for endpoints
	endpoints, err := c.clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil || len(endpoints.Subsets) == 0 {
		result.Issues = append(result.Issues, Issue{
			ID:           GenerateIssueID(name, "no-endpoints"),
			Title:        "Service has no active endpoints",
			Description:  "No pods are currently matching the service selector or none are ready.",
			Priority:     PriorityCritical,
			ResourceType: "Service",
			ResourceName: name,
			Namespace:    namespace,
		})
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Check if pods with labels %v exist in namespace %s", svc.Spec.Selector, namespace),
			"Verify if the pods are in 'Ready' state",
			"Check if the containerPort matches the service targetPort",
		)
	}

	// Check for Zombie Service (no traffic)
	if c.prometheusURL != "" {
		query := fmt.Sprintf("sum(rate(istio_requests_total{destination_service_name=\"%s\", destination_service_namespace=\"%s\"}[24h]))", name, namespace)
		metrics, err := c.queryPrometheus(ctx, query)
		if err == nil && len(metrics.Data.Result) > 0 {
			// If result exists, traffic is being recorded. Check value.
			val := metrics.Data.Result[0].Value
			if len(val) >= 2 {
				var rps float64
				rpsStr, ok := val[1].(string)
				if ok {
					fmt.Sscanf(rpsStr, "%f", &rps)
				}

				if rps < 0.001 {
					result.Issues = append(result.Issues, Issue{
						ID:           GenerateIssueID(name, "zombie-service"),
						Title:        "Zombie Service (No Traffic)",
						Description:  fmt.Sprintf("This service has received near-zero traffic (%f RPS) over the last 24 hours.", rps),
						Priority:     PriorityLow,
						ResourceType: "Service",
						ResourceName: name,
						Namespace:    namespace,
					})
					result.Recommendations = append(result.Recommendations,
						"Consider downscaling or decommissioning this service to save resources",
						"Verify if this service is still required by other components",
					)
				}
			}
		} else if err == nil && len(metrics.Data.Result) == 0 {
			// No results in Prometheus usually means no traffic recorded at all
			result.Issues = append(result.Issues, Issue{
				ID:           GenerateIssueID(name, "zombie-service"),
				Title:        "Zombie Service (No Traffic)",
				Description:  "No traffic data found for this service in the last 24 hours.",
				Priority:     PriorityLow,
				ResourceType: "Service",
				ResourceName: name,
				Namespace:    namespace,
			})
			result.Recommendations = append(result.Recommendations,
				"Investigate if this service is obsolete",
				"Consider reducing the number of replicas to 1 or 0",
			)
		}
	}

	return result, nil
}

func (c *Client) troubleshootDeployment(ctx context.Context, name, namespace string, result *AnalysisResult) (*AnalysisResult, error) {
	deploy, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	result.Status = fmt.Sprintf("%d/%d replicas ready", deploy.Status.ReadyReplicas, deploy.Status.Replicas)

	if deploy.Status.ReadyReplicas < deploy.Status.Replicas {
		result.Issues = append(result.Issues, Issue{
			ID:           GenerateIssueID(name, "replicas-unavailable"),
			Title:        fmt.Sprintf("Deployment has %d unavailable replicas", deploy.Status.Replicas-deploy.Status.ReadyReplicas),
			Description:  "The desired number of replicas is not yet reached.",
			Priority:     PriorityHigh,
			ResourceType: "Deployment",
			ResourceName: name,
			Namespace:    namespace,
		})
		result.Recommendations = append(result.Recommendations,
			"Check the status of the associated pods",
			"Check for resource quota limits in the namespace",
			"Verify if there are any rolling update issues in progress",
		)
	}

	return result, nil
}
