package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Event represents a Kubernetes event
type Event struct {
	Type              string    `json:"type"`
	Reason            string    `json:"reason"`
	Message           string    `json:"message"`
	InvolvedObject    string    `json:"involved_object"`
	InvolvedNamespace string    `json:"involved_namespace"`
	InvolvedKind      string    `json:"involved_kind"`
	FirstTimestamp    time.Time `json:"first_timestamp"`
	LastTimestamp     time.Time `json:"last_timestamp"`
	Count             int32     `json:"count"`
	Source            string    `json:"source,omitempty"`
}

// GetEvents retrieves events from the specified namespace
// If namespace is empty or "all", retrieves events from all namespaces
// By default, filters out Normal events (only shows Warning and Error)
func (c *Client) GetEvents(ctx context.Context, namespace string, includeNormal bool) ([]Event, error) {
	// Determine cache key and namespace
	var cacheKey string
	var ns string

	if namespace == "" || namespace == "all" {
		cacheKey = "events-all"
		ns = ""
	} else {
		cacheKey = fmt.Sprintf("events-%s", namespace)
		ns = namespace
	}

	if includeNormal {
		cacheKey += "-normal"
	}

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if events, ok := cached.([]Event); ok {
			return events, nil
		}
	}

	// Build field selector to filter out Normal events
	fieldSelector := ""
	if !includeNormal {
		fieldSelector = "type!=Normal"
	}

	// Query Kubernetes API
	var eventList *corev1.EventList
	var err error

	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	if ns == "" {
		eventList, err = c.clientset.CoreV1().Events("").List(ctx, listOptions)
	} else {
		eventList, err = c.clientset.CoreV1().Events(ns).List(ctx, listOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Convert to our Event struct
	events := make([]Event, 0, len(eventList.Items))
	for _, e := range eventList.Items {
		event := convertEvent(&e)
		events = append(events, event)
	}

	// Cache the result
	c.cache.set(cacheKey, events)

	return events, nil
}

// GetPodEvents retrieves events for a specific pod
func (c *Client) GetPodEvents(ctx context.Context, namespace, podName string) ([]Event, error) {
	// Fix Issue #10: Resolve namespace BEFORE creating cache key to prevent race condition
	if namespace == "" {
		namespace = c.namespace
	}

	cacheKey := fmt.Sprintf("events-pod-%s-%s", namespace, podName)

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if events, ok := cached.([]Event); ok {
			return events, nil
		}
	}

	// Query events for this specific pod
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod",
		podName, namespace)

	eventList, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list events for pod %s/%s: %w", namespace, podName, err)
	}

	// Convert to our Event struct
	events := make([]Event, 0, len(eventList.Items))
	for _, e := range eventList.Items {
		event := convertEvent(&e)
		events = append(events, event)
	}

	// Cache the result
	c.cache.set(cacheKey, events)

	return events, nil
}

// convertEvent converts a Kubernetes Event to our Event struct
func convertEvent(e *corev1.Event) Event {
	source := ""
	if e.Source.Component != "" {
		source = e.Source.Component
	}

	return Event{
		Type:              e.Type,
		Reason:            e.Reason,
		Message:           e.Message,
		InvolvedObject:    e.InvolvedObject.Name,
		InvolvedNamespace: e.InvolvedObject.Namespace,
		InvolvedKind:      e.InvolvedObject.Kind,
		FirstTimestamp:    e.FirstTimestamp.Time,
		LastTimestamp:     e.LastTimestamp.Time,
		Count:             e.Count,
		Source:            source,
	}
}
