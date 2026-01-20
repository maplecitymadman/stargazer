package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Pod represents a Kubernetes pod with relevant metadata
type Pod struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Status          string            `json:"status"`
	Node            string            `json:"node"`
	Ready           bool              `json:"ready"`
	Restarts        int32             `json:"restarts"`
	Age             string            `json:"age"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Containers      []string          `json:"containers"`
	InitContainers  []string          `json:"init_containers"`
	HasServiceMesh  bool              `json:"has_service_mesh"`
	ContainerStates []ContainerState  `json:"container_states,omitempty"`
}

// ContainerState represents the state of a container
type ContainerState struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"` // running, waiting, terminated
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
}

// GetPods retrieves pods from the specified namespace
// If namespace is empty or "all", retrieves pods from all namespaces
func (c *Client) GetPods(ctx context.Context, namespace string) ([]Pod, error) {
	// Determine cache key and namespace to query
	var cacheKey string
	var ns string

	if namespace == "" || namespace == "all" {
		cacheKey = "pods-all"
		ns = "" // Empty string means all namespaces
	} else {
		cacheKey = fmt.Sprintf("pods-%s", namespace)
		ns = namespace
	}

	// Check cache first
	if cached, found := c.cache.get(cacheKey); found {
		if pods, ok := cached.([]Pod); ok {
			return pods, nil
		}
	}

	// Query Kubernetes API
	var podList *corev1.PodList
	var err error

	if ns == "" {
		podList, err = c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		podList, err = c.clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Convert to our Pod struct
	pods := make([]Pod, 0, len(podList.Items))
	for _, p := range podList.Items {
		pod := convertPod(&p)
		pods = append(pods, pod)
	}

	// Cache the result
	c.cache.set(cacheKey, pods)

	return pods, nil
}

// GetPod retrieves a single pod by name and namespace
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*Pod, error) {
	// Fix Issue #10: Resolve namespace BEFORE creating cache key to prevent race condition
	if namespace == "" {
		namespace = c.namespace
	}

	cacheKey := fmt.Sprintf("pod-%s-%s", namespace, name)

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if pod, ok := cached.(*Pod); ok {
			return pod, nil
		}
	}

	// Query Kubernetes API
	p, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, name, err)
	}

	pod := convertPod(p)

	// Cache the result
	c.cache.set(cacheKey, &pod)

	return &pod, nil
}

// convertPod converts a Kubernetes Pod object to our Pod struct
func convertPod(p *corev1.Pod) Pod {
	// Extract container names
	containers := make([]string, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		containers = append(containers, c.Name)
	}

	initContainers := make([]string, 0, len(p.Spec.InitContainers))
	for _, c := range p.Spec.InitContainers {
		initContainers = append(initContainers, c.Name)
	}

	// Calculate restarts
	var restarts int32
	for _, cs := range p.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}

	// Extract container states
	containerStates := make([]ContainerState, 0, len(p.Status.ContainerStatuses))
	for _, cs := range p.Status.ContainerStatuses {
		state := ContainerState{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
		}

		// Determine state
		if cs.State.Running != nil {
			state.State = "running"
		} else if cs.State.Waiting != nil {
			state.State = "waiting"
			state.Reason = cs.State.Waiting.Reason
			state.Message = cs.State.Waiting.Message
		} else if cs.State.Terminated != nil {
			state.State = "terminated"
			state.Reason = cs.State.Terminated.Reason
			state.Message = cs.State.Terminated.Message
		}

		containerStates = append(containerStates, state)
	}

	// Detect service mesh (Istio)
	hasServiceMesh := detectServiceMesh(p)

	// Labels and annotations
	labels := p.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	annotations := p.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	return Pod{
		Name:            p.Name,
		Namespace:       p.Namespace,
		Status:          string(p.Status.Phase),
		Node:            p.Spec.NodeName,
		Ready:           isPodReady(p),
		Restarts:        restarts,
		Age:             calculateAge(p.CreationTimestamp.Time),
		Labels:          labels,
		Annotations:     annotations,
		Containers:      containers,
		InitContainers:  initContainers,
		HasServiceMesh:  hasServiceMesh,
		ContainerStates: containerStates,
	}
}

// isPodReady checks if all containers in the pod are ready
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// detectServiceMesh detects if the pod has a service mesh sidecar (e.g., Istio)
func detectServiceMesh(pod *corev1.Pod) bool {
	// Check container names and images
	allContainers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	for _, container := range allContainers {
		nameLower := strings.ToLower(container.Name)
		imageLower := strings.ToLower(container.Image)

		if strings.Contains(nameLower, "istio") ||
			strings.Contains(nameLower, "envoy") ||
			strings.Contains(imageLower, "istio") ||
			strings.Contains(imageLower, "envoy") {
			return true
		}
	}

	// Check annotations
	if pod.Annotations != nil {
		for key := range pod.Annotations {
			keyLower := strings.ToLower(key)
			if strings.Contains(keyLower, "istio") ||
				strings.Contains(keyLower, "sidecar") {
				return true
			}
		}

		// Specific Istio annotations
		if _, ok := pod.Annotations["sidecar.istio.io/status"]; ok {
			return true
		}
		if _, ok := pod.Annotations["sidecar.istio.io/inject"]; ok {
			return true
		}
	}

	return false
}

// calculateAge returns a human-readable age string
func calculateAge(creationTime time.Time) string {
	duration := time.Since(creationTime)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return "<1m"
}
