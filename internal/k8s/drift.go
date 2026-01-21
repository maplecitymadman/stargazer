package k8s

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// getDriftData retrieves drift information from GitOps tools (e.g. ArgoCD)
func (c *Client) getDriftData(ctx context.Context, ns string) (DriftData, error) {
	drift := DriftData{}

	if c.dynamicClient == nil {
		return drift, nil
	}

	// Detect ArgoCD Application CRDs
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	// List all applications (usually in argocd namespace)
	// We'll check the 'argocd' namespace first, then fallback to all namespaces
	appList, err := c.dynamicClient.Resource(gvr).Namespace("argocd").List(ctx, metav1.ListOptions{})
	if err != nil {
		// Fallback to all namespaces if argocd doesn't exist or permissions fail
		appList, err = c.dynamicClient.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		drift.ArgoEnabled = true
		for _, app := range appList.Items {
			status, _, _ := unstructuredGetNestedString(app.Object, "status", "sync", "status")
			repoURL, _, _ := unstructuredGetNestedString(app.Object, "spec", "source", "repoURL")
			targetRevision, _, _ := unstructuredGetNestedString(app.Object, "spec", "source", "targetRevision")

			drift.Applications = append(drift.Applications, ArgoAppInfo{
				Name:           app.GetName(),
				Namespace:      app.GetNamespace(),
				Status:         status,
				RepoURL:        repoURL,
				TargetRevision: targetRevision,
			})
		}
	}

	return drift, nil
}

// Helper to get nested string from unstructured map
func unstructuredGetNestedString(obj map[string]interface{}, fields ...string) (string, bool, error) {
	var val interface{} = obj
	for _, field := range fields {
		if m, ok := val.(map[string]interface{}); ok {
			val = m[field]
		} else {
			return "", false, nil
		}
	}

	if s, ok := val.(string); ok {
		return s, true, nil
	}
	return "", false, nil
}

// mapServiceToDrift maps services in the topology to drift status from applications
func (c *Client) mapServiceToDrift(services map[string]ServiceInfo, drift DriftData) {
	for name, svc := range services {
		// Try to find an application matching this service
		// Simpler logic: match application name or namespace
		// More complex: match resources managed by application
		for _, app := range drift.Applications {
			// This is a heuristic match
			if strings.Contains(strings.ToLower(app.Name), strings.ToLower(svc.Name)) ||
				(app.Namespace == svc.Namespace && strings.Contains(strings.ToLower(app.Name), "app")) {
				svc.DriftStatus = app.Status
				services[name] = svc
				break
			}
		}

		if svc.DriftStatus == "" {
			svc.DriftStatus = "Unknown"
			services[name] = svc
		}
	}
}
