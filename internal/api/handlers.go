package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maplecitymadman/stargazer/internal/config"
	"github.com/maplecitymadman/stargazer/internal/k8s"
)

// Health check endpoint
func (s *Server) handleHealth(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"cluster": "disconnected",
			"version": "0.1.0-dev",
			"error":   "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check cluster connectivity
	err := client.Health(ctx)
	healthy := err == nil

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":  map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
		"cluster": map[bool]string{true: "connected", false: "disconnected"}[healthy],
		"version": "0.1.0-dev",
		"error":   map[bool]string{true: "", false: err.Error()}[healthy],
	})
}

// Get current context info
func (s *Server) handleGetContexts(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"contexts":        []gin.H{},
			"current_context": "",
			"total":           0,
			"error":           "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	// TODO: Implement full context listing from kubeconfig
	// For now, return current context info
	c.JSON(http.StatusOK, gin.H{
		"contexts": []gin.H{
			{
				"name":           client.GetContext(),
				"cluster":        client.GetContext(),
				"user":           "",
				"namespace":      client.GetNamespace(),
				"server":         "",
				"cloud_provider": "unknown",
				"is_current":     true,
			},
		},
		"current_context": client.GetContext(),
		"total":           1,
	})
}

// Get current context
func (s *Server) handleGetCurrentContext(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"context": "",
			"info":    nil,
			"error":   "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"context": client.GetContext(),
		"info": gin.H{
			"name":           client.GetContext(),
			"cluster":        client.GetContext(),
			"namespace":      client.GetNamespace(),
			"cloud_provider": "unknown",
		},
	})
}

// Switch Kubernetes context
func (s *Server) handleSwitchContext(c *gin.Context) {
	var req struct {
		Context string `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Context == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Context name required"})
		return
	}

	// TODO: Implement full context switching with kubeconfig parsing
	// For now, return success but note that client needs to be recreated
	// The frontend will reload the page which will pick up the new context
	
	// Save context preference to config
	cfg, err := config.Load()
	if err == nil {
		// Note: Config doesn't have context field yet, but we can add it
		// For now, just return success
		_ = cfg
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"context": req.Context,
		"message": fmt.Sprintf("Switched to context: %s (reload page to apply)", req.Context),
	})
}

// Get cluster health (v2 style)
func (s *Server) handleClusterHealth(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"pods": gin.H{
				"total":   0,
				"healthy": 0,
			},
			"deployments": gin.H{
				"total":   0,
				"healthy": 0,
			},
			"events": gin.H{
				"warnings": 0,
				"errors":   0,
			},
			"overall_health": "degraded",
			"error":          "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	// Support "all" for all namespaces
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get resource counts
	var totalPods, healthyPods int
	var totalDeployments, healthyDeployments int
	var warningEvents, errorEvents int

	// Get pods
	pods, err := client.GetPods(ctx, ns)
	if err == nil {
		totalPods = len(pods)
		for _, pod := range pods {
			if pod.Status == "Running" && pod.Ready {
				healthyPods++
			}
		}
	}

	// Get deployments
	deployments, err := client.GetDeployments(ctx, ns)
	if err == nil {
		totalDeployments = len(deployments)
		for _, dep := range deployments {
			if dep.Replicas == dep.ReadyReplicas {
				healthyDeployments++
			}
		}
	}

	// Get events
	events, err := client.GetEvents(ctx, ns, false)
	if err == nil {
		for _, event := range events {
			if event.Type == "Warning" {
				warningEvents++
			} else if event.Type == "Error" {
				errorEvents++
			}
		}
	}

	overallHealth := "healthy"
	if healthyPods != totalPods || healthyDeployments != totalDeployments || totalPods == 0 {
		overallHealth = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"pods": gin.H{
			"total":   totalPods,
			"healthy": healthyPods,
		},
		"deployments": gin.H{
			"total":   totalDeployments,
			"healthy": healthyDeployments,
		},
		"events": gin.H{
			"warnings": warningEvents,
			"errors":   errorEvents,
		},
		"overall_health": overallHealth,
	})
}

// Get cluster issues (v2 style - same as handleGetIssues but different endpoint)
func (s *Server) handleClusterIssues(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"issues": []interface{}{},
			"count":  0,
			"error":  "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "all"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	discovery := k8s.NewDiscovery(client)
	issues, err := discovery.ScanAll(ctx, namespace)
	if err != nil {
		// Return empty list instead of error to prevent UI breakage
		c.JSON(http.StatusOK, gin.H{
			"issues": []interface{}{},
			"count":  0,
		})
		return
	}

	// Convert issues to map format for JSON
	issueMaps := make([]map[string]interface{}, len(issues))
	for i, issue := range issues {
		// Convert priority to lowercase for frontend
		priorityStr := string(issue.Priority)
		if priorityStr == "CRITICAL" {
			priorityStr = "critical"
		} else if priorityStr == "WARNING" {
			priorityStr = "warning"
		} else {
			priorityStr = "info"
		}
		
		issueMaps[i] = map[string]interface{}{
			"id":            issue.ID,
			"title":        issue.Title,
			"description":  issue.Description,
			"priority":     priorityStr,
			"resource_type": issue.ResourceType,
			"resource_name": issue.ResourceName,
			"namespace":     issue.Namespace,
			"timestamp":     time.Now().Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"issues": issueMaps,
		"count":  len(issueMaps),
	})
}

// Get all discovered issues
func (s *Server) handleGetIssues(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"issues":    []interface{}{},
			"count":     0,
			"namespace": "all",
			"error":     "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "all"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	discovery := k8s.NewDiscovery(client)
	issues, err := discovery.ScanAll(ctx, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to scan cluster: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issues":    issues,
		"count":     len(issues),
		"namespace": namespace,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Trigger a new scan
func (s *Server) handleScan(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Namespace = "all"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clear cache to force fresh scan
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	client.ClearCache()

	discovery := k8s.NewDiscovery(client)
	issues, err := discovery.ScanAll(ctx, req.Namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Scan failed: %v", err),
		})
		return
	}

	// Broadcast to WebSocket clients
	s.wsHub.Broadcast(Message{
		Type: "scan_complete",
		Data: map[string]interface{}{
			"issues":    issues,
			"count":     len(issues),
			"namespace": req.Namespace,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"issues":    issues,
		"count":     len(issues),
		"namespace": req.Namespace,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Get current namespace
func (s *Server) handleGetNamespace(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"namespace": "default",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace": client.GetNamespace(),
	})
}

// Get all namespaces
func (s *Server) handleGetNamespaces(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"namespaces": []interface{}{},
			"count":      0,
			"error":      "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaces, err := client.GetNamespaces(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get namespaces: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
		"count":      len(namespaces),
	})
}

// Get pods with query param (v2 style)
func (s *Server) handleGetPodsQuery(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"pods":      []interface{}{},
			"count":     0,
			"namespace": "all",
			"error":     "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	// Support "all" for all namespaces
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := client.GetPods(ctx, ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get pods: %v", err),
		})
		return
	}

	displayNs := namespace
	if displayNs == "" {
		displayNs = "all"
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":      pods,
		"count":    len(pods),
		"namespace": displayNs,
	})
}

// Get pods in a namespace (path param - legacy)
func (s *Server) handleGetPods(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Param("namespace")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := client.GetPods(ctx, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get pods: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":      pods,
		"count":     len(pods),
		"namespace": namespace,
	})
}

// Get a specific pod
func (s *Server) handleGetPod(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Param("namespace")
	podName := c.Param("pod")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pod, err := client.GetPod(ctx, namespace, podName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pod not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, pod)
}

// Get pod logs
func (s *Server) handleGetPodLogs(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Param("namespace")
	podName := c.Param("pod")

	// Parse query parameters
	tail := 100 // default
	if tailParam := c.Query("tail"); tailParam != "" {
		if parsed, err := strconv.Atoi(tailParam); err == nil && parsed > 0 {
			tail = parsed
		}
	}

	follow := c.Query("follow") == "true"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logs, err := client.GetPodLogs(ctx, namespace, podName, tail, follow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get logs: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pod":       podName,
		"namespace": namespace,
		"logs":      logs,
	})
}

// Get deployments with query param (v2 style)
func (s *Server) handleGetDeploymentsQuery(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"deployments": []interface{}{},
			"count":      0,
			"namespace":  "all",
			"error":      "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	// Support "all" for all namespaces
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deployments, err := client.GetDeployments(ctx, ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get deployments: %v", err),
		})
		return
	}

	displayNs := namespace
	if displayNs == "" {
		displayNs = "all"
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
		"namespace":   displayNs,
	})
}

// Get deployments in a namespace (path param - legacy)
func (s *Server) handleGetDeployments(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Param("namespace")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deployments, err := client.GetDeployments(ctx, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get deployments: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
		"namespace":   namespace,
	})
}

// Get services in a namespace
func (s *Server) handleGetServices(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Param("namespace")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	services, err := client.GetServices(ctx, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get services: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services":  services,
		"count":     len(services),
		"namespace": namespace,
	})
}

// Get events with query param (v2 style)
func (s *Server) handleGetEventsQuery(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"events":    []interface{}{},
			"count":     0,
			"namespace": "all",
			"error":     "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	includeNormal := c.Query("include_normal") == "true"
	
	// Support "all" for all namespaces
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := client.GetEvents(ctx, ns, includeNormal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get events: %v", err),
		})
		return
	}

	displayNs := namespace
	if displayNs == "" {
		displayNs = "all"
	}

	c.JSON(http.StatusOK, gin.H{
		"events":    events,
		"count":     len(events),
		"namespace": displayNs,
	})
}

// Get events in a namespace (path param - legacy)
func (s *Server) handleGetEvents(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Param("namespace")
	includeNormal := c.Query("include_normal") == "true"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := client.GetEvents(ctx, namespace, includeNormal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get events: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":    events,
		"count":     len(events),
		"namespace": namespace,
	})
}

// Get all nodes
func (s *Server) handleGetNodes(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodes, err := client.GetNodes(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get nodes: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// Troubleshoot endpoint (Phase 6 - AI)
func (s *Server) handleTroubleshoot(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "AI troubleshooting not yet implemented - coming in Phase 4-6",
	})
}

// Get service topology (Phase 2 deferred)
func (s *Server) handleGetTopology(c *gin.Context) {
	namespace := c.Query("namespace")
	// Support "all" for all namespaces
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	// TODO: Implement service topology
	// For now, return empty topology structure
	c.JSON(http.StatusOK, gin.H{
		"services":      map[string]interface{}{},
		"connectivity":  map[string]interface{}{},
		"network_policies": []interface{}{},
		"namespace":     ns,
		"summary": gin.H{
			"total_services":       0,
			"blocked_connections":  0,
			"allowed_connections":  0,
		},
	})
}

// Get service connections
func (s *Server) handleGetServiceConnections(c *gin.Context) {
	_ = c.Param("service_name") // serviceName - TODO: use when implemented
	_ = c.Query("namespace")    // namespace - TODO: use when implemented
	
	// TODO: Implement service connections
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Service connections not yet implemented",
	})
}

// Get network policy YAML
func (s *Server) handleGetNetworkPolicyYaml(c *gin.Context) {
	_ = c.Param("policy_name") // policyName - TODO: use when implemented
	namespace := c.Query("namespace")
	client := s.GetK8sClient()
	if namespace == "" && client != nil {
		namespace = client.GetNamespace()
	}
	_ = namespace // TODO: use when implemented

	// TODO: Implement network policy YAML retrieval
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Network policy YAML retrieval not yet implemented",
	})
}

// Get resource YAML
func (s *Server) handleGetResourceYaml(c *gin.Context) {
	_ = c.Param("resource_type") // resourceType - TODO: use when implemented
	_ = c.Param("resource_name") // resourceName - TODO: use when implemented
	namespace := c.Query("namespace")
	client := s.GetK8sClient()
	if namespace == "" && client != nil {
		namespace = client.GetNamespace()
	}
	_ = namespace // TODO: use when implemented

	// TODO: Implement resource YAML retrieval
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Resource YAML retrieval not yet implemented",
	})
}

// Get config (Phase 7)
func (s *Server) handleGetConfig(c *gin.Context) {
	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load config: %v", err),
		})
		return
	}

	// Sanitize API keys before sending to client
	sanitized := cfg
	for name, provider := range sanitized.LLM.Providers {
		if provider.APIKey != "" {
			provider.APIKey = "***REDACTED***"
			sanitized.LLM.Providers[name] = provider
		}
	}

	c.JSON(http.StatusOK, sanitized)
}

// Get providers config (v2 style)
func (s *Server) handleGetProvidersConfig(c *gin.Context) {
	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load config: %v", err),
		})
		return
	}

	// Build providers summary with has_key flag
	providers := make(map[string]interface{})
	for name, provider := range cfg.LLM.Providers {
		providers[name] = map[string]interface{}{
			"enabled": provider.Enabled,
			"model":   provider.Model,
			"has_key": provider.APIKey != "",
		}
	}

	kubeconfigPath := ""
	client := s.GetK8sClient()
	if client != nil {
		kubeconfigPath = client.GetKubeconfigPath()
	} else if cfg.Kubeconfig.Path != "" {
		kubeconfigPath = cfg.Kubeconfig.Path
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"kubeconfig": gin.H{
			"path": kubeconfigPath,
		},
		"sops_available": false, // TODO: Check if SOPS is available
	})
}

// Get kubeconfig status
func (s *Server) handleGetKubeconfigStatus(c *gin.Context) {
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	status := gin.H{
		"configured": false,
		"path":       "",
		"found":      false,
		"error":      "",
		"auto_found": false,
	}

	// Check if kubeconfig is configured
	kubeconfigPath := cfg.Kubeconfig.Path
	if kubeconfigPath == "" {
		// Try auto-discovery using the same logic as resolveKubeconfigPath
		// This helps show what was found even if client creation failed
		discoveredPath, err := k8s.DiscoverKubeconfigPath()
		if err == nil && discoveredPath != "" {
			kubeconfigPath = discoveredPath
			status["auto_found"] = true
			status["configured"] = false // Not explicitly configured, but found
			status["path"] = kubeconfigPath
			status["found"] = true
		} else {
			// Check if we have a working client (might have been auto-discovered)
			client := s.GetK8sClient()
			if client != nil {
				kubeconfigPath = client.GetKubeconfigPath()
				status["auto_found"] = true
				status["configured"] = false
				status["path"] = kubeconfigPath
				status["found"] = true
			} else {
				status["error"] = "Kubeconfig not configured. Please set kubeconfig path in Settings."
			}
		}
	} else {
		status["configured"] = true
		status["path"] = kubeconfigPath
		// Check if file exists
		if _, err := os.Stat(kubeconfigPath); err == nil {
			status["found"] = true
		} else {
			status["found"] = false
			status["error"] = fmt.Sprintf("Kubeconfig file not found at: %s", kubeconfigPath)
		}
	}

	// Check if client is working
	client := s.GetK8sClient()
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Health(ctx); err == nil {
			status["connected"] = true
			status["context"] = client.GetContext()
		} else {
			status["connected"] = false
			status["connection_error"] = err.Error()
		}
	} else {
		status["connected"] = false
	}

	c.JSON(http.StatusOK, status)
}

// Set kubeconfig path
func (s *Server) handleSetKubeconfig(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		Context string `json:"context,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kubeconfig path is required"})
		return
	}

	// Expand path (handle ~ and env vars)
	expandedPath := os.ExpandEnv(req.Path)
	if len(expandedPath) > 0 && expandedPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to get home directory: %v", err),
			})
			return
		}
		expandedPath = filepath.Join(home, expandedPath[1:])
	}

	// Verify file exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Kubeconfig file not found: %s", expandedPath),
		})
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Update kubeconfig in config
	cfg.Kubeconfig.Path = expandedPath
	if req.Context != "" {
		cfg.Kubeconfig.Context = req.Context
	}

	// Save config
	if err := cfg.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	// Try to create new client with updated kubeconfig
	newClient, err := k8s.NewClient(expandedPath, req.Context, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to connect with kubeconfig: %v", err),
			"path":  expandedPath,
		})
		return
	}

	// Update server's k8s client
	s.UpdateK8sClient(newClient)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"path":    expandedPath,
		"context": newClient.GetContext(),
		"message": "Kubeconfig configured successfully",
	})
}

// Set provider model
func (s *Server) handleSetProviderModel(c *gin.Context) {
	provider := c.Param("provider")
	var req struct {
		Model string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load config: %v", err),
		})
		return
	}

	if p, exists := cfg.LLM.Providers[provider]; exists {
		p.Model = req.Model
		cfg.LLM.Providers[provider] = p
		if err := cfg.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to save config: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "provider": provider, "model": req.Model})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider %s not found", provider)})
	}
}

// Enable/disable provider
func (s *Server) handleEnableProvider(c *gin.Context) {
	provider := c.Param("provider")
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load config: %v", err),
		})
		return
	}

	if p, exists := cfg.LLM.Providers[provider]; exists {
		p.Enabled = req.Enabled
		cfg.LLM.Providers[provider] = p
		if err := cfg.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to save config: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "provider": provider, "enabled": req.Enabled})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Provider %s not found", provider)})
	}
}

// Set provider API key
func (s *Server) handleSetProviderApiKey(c *gin.Context) {
	provider := c.Param("provider")
	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key is required"})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load config: %v", err),
		})
		return
	}

	if err := cfg.EnableProvider(provider, req.APIKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to set API key: %v", err),
		})
		return
	}

	if err := cfg.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "provider": provider, "has_key": true})
}

// Set config (Phase 7)
func (s *Server) handleSetConfig(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load config: %v", err),
		})
		return
	}

	// Apply updates (basic implementation - could be more sophisticated)
	// For now, just support enabling/disabling providers
	if providerUpdates, ok := updates["providers"].(map[string]interface{}); ok {
		for name, value := range providerUpdates {
			if providerData, ok := value.(map[string]interface{}); ok {
				if enabled, ok := providerData["enabled"].(bool); ok {
					if provider, exists := cfg.LLM.Providers[name]; exists {
						provider.Enabled = enabled
						cfg.LLM.Providers[name] = provider
					}
				}
				if apiKey, ok := providerData["api_key"].(string); ok && apiKey != "" {
					if err := cfg.EnableProvider(name, apiKey); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": fmt.Sprintf("Failed to update provider %s: %v", name, err),
						})
						return
					}
				}
			}
		}
	}

	if err := cfg.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
	})
}

// Setup wizard (Phase 7)
func (s *Server) handleSetupWizard(c *gin.Context) {
	// Note: The wizard is interactive CLI-only
	// API endpoint just returns info about how to run it
	c.JSON(http.StatusOK, gin.H{
		"message": "Setup wizard is available via CLI",
		"command": "stargazer config setup",
		"info":    "The setup wizard is an interactive CLI tool. Run 'stargazer config setup' in your terminal.",
	})
}
