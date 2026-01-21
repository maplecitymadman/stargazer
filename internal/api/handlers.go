package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maplecitymadman/stargazer/internal/config"
	"github.com/maplecitymadman/stargazer/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	c.JSON(status, gin.H{
		"status":  map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
		"cluster": map[bool]string{true: "connected", false: "disconnected"}[healthy],
		"version": "0.1.0-dev",
		"error":   errorMsg,
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

	// Get cluster name from kubeconfig
	clusterName := client.GetClusterName()
	if clusterName == "" || clusterName == "unknown" {
		// Fallback to context name if cluster name can't be determined
		clusterName = client.GetContext()
	}

	// Return current context info
	c.JSON(http.StatusOK, gin.H{
		"contexts": []gin.H{
			{
				"name":           client.GetContext(),
				"cluster":        clusterName,
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

	// Get cluster name from kubeconfig
	clusterName := client.GetClusterName()
	if clusterName == "" || clusterName == "unknown" {
		// Fallback to context name if cluster name can't be determined
		clusterName = client.GetContext()
	}

	c.JSON(http.StatusOK, gin.H{
		"context": client.GetContext(),
		"info": gin.H{
			"name":           client.GetContext(),
			"cluster":        clusterName,
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

	// Context switching requires client recreation
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

	// Get events
	var warningEvents, errorEvents int
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
	if errorEvents > 0 || warningEvents > 0 {
		overallHealth = "degraded"
	}

	// Return empty pods/deployments for frontend compatibility
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
			"warnings": warningEvents,
			"errors":   errorEvents,
		},
		"overall_health": overallHealth,
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

// Get all services across all namespaces for PathTracer resource selection
func (s *Server) handleGetAllServices(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusOK, gin.H{
			"services": []gin.H{},
			"count":    0,
			"error":    "Kubernetes client not initialized. Please configure kubeconfig.",
		})
		return
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "all"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	services, err := client.GetServices(ctx, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get services: %v", err),
		})
		return
	}

	// Format for PathTracer: include ingress-gateway and egress-gateway options
	formatted := []gin.H{
		{"name": "ingress-gateway", "namespace": "", "type": "ingress", "display": "Ingress Gateway"},
		{"name": "egress-gateway", "namespace": "", "type": "egress", "display": "Egress Gateway"},
		{"name": "external", "namespace": "", "type": "external", "display": "External"},
	}

	for _, svc := range services {
		displayName := svc.Name
		if svc.Namespace != "" {
			displayName = fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
		}
		formatted = append(formatted, gin.H{
			"name":      svc.Name,
			"namespace": svc.Namespace,
			"type":      "service",
			"display":   displayName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"services": formatted,
		"count":    len(formatted),
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
	if err := validateNamespace(namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid namespace parameter: %v", err),
		})
		return
	}
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

// Troubleshoot endpoint for automated resource analysis
func (s *Server) handleTroubleshoot(c *gin.Context) {
	resourceType := c.Query("type")
	name := c.Query("name")
	namespace := c.Query("namespace")

	if name == "" || resourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name and type query parameters are required",
		})
		return
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	result, err := client.Troubleshoot(c.Request.Context(), resourceType, name, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Troubleshooting failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Get service topology with Cilium, Istio, and Kyverno support
// validateNamespace validates a namespace parameter
func validateNamespace(ns string) error {
	if ns == "" || ns == "all" {
		return nil
	}
	// Kubernetes namespace validation: DNS-1123 label format
	if len(ns) > 253 {
		return fmt.Errorf("namespace too long (max 253 characters)")
	}
	// Basic validation - alphanumeric and hyphens only, must start/end with alphanumeric
	if !strings.HasPrefix(ns, "kube-") && !strings.HasPrefix(ns, "istio-") {
		// Allow system namespaces
		for _, char := range ns {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
				return fmt.Errorf("invalid namespace format: %s", ns)
			}
		}
	}
	return nil
}

func (s *Server) handleGetTopology(c *gin.Context) {
	namespace := c.Query("namespace")
	// Validate namespace parameter
	if err := validateNamespace(namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid namespace parameter: %v", err),
		})
		return
	}
	// Support "all" for all namespaces
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	ctx := c.Request.Context()
	topology, err := client.GetTopology(ctx, ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get topology: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, topology)
}

// Get service connections and detailed network analysis
func (s *Server) handleGetServiceConnections(c *gin.Context) {
	serviceName := c.Param("service_name")
	namespace := c.Query("namespace")

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service_name path parameter is required",
		})
		return
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	ctx := c.Request.Context()
	topology, err := client.GetTopology(ctx, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get topology: %v", err),
		})
		return
	}

	key := k8s.ServiceKey(namespace, serviceName)
	connInfo, ok := topology.Connectivity[key]
	if !ok {
		// Try without namespace if not found (for global search)
		connInfo, ok = topology.Connectivity[serviceName]
	}

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Service %s not found in topology", key),
		})
		return
	}

	c.JSON(http.StatusOK, connInfo)
}

// Trace connection path
func (s *Server) handleTracePath(c *gin.Context) {
	source := c.Query("source")
	destination := c.Query("destination")
	namespace := c.Query("namespace")

	if source == "" || destination == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "source and destination parameters required",
		})
		return
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	ctx := c.Request.Context()

	// Get full topology first
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	topology, err := client.GetTopology(ctx, ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get topology: %v", err),
		})
		return
	}

	// Trace path
	trace, err := client.TracePath(ctx, source, destination, ns, topology)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to trace path: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, trace)
}

// Get network policy YAML
func (s *Server) handleGetNetworkPolicyYaml(c *gin.Context) {
	policyName := c.Param("policy_name")
	namespace := c.Query("namespace")
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	if namespace == "" {
		namespace = client.GetNamespace()
	}

	ctx := c.Request.Context()
	policy, err := client.GetClientset().NetworkingV1().NetworkPolicies(namespace).Get(ctx, policyName, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("NetworkPolicy not found: %v", err),
		})
		return
	}

	// Convert to YAML using JSON as intermediate (simplified)
	// In production, use proper YAML marshaling library like gopkg.in/yaml.v3
	jsonData, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to convert to JSON: %v", err),
		})
		return
	}
	// For now, return JSON-formatted data (can be enhanced with YAML library)
	yamlData := string(jsonData)

	c.JSON(http.StatusOK, gin.H{
		"name":      policy.Name,
		"namespace": policy.Namespace,
		"yaml":      yamlData,
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
	c.JSON(http.StatusOK, cfg)
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

	// Return basic config info
	kubeconfigPath := ""
	if cfg.Kubeconfig.Path != "" {
		kubeconfigPath = cfg.Kubeconfig.Path
	} else {
		kubeconfigPath = "~/.kube/config (default)"
	}

	c.JSON(http.StatusOK, gin.H{
		"kubeconfig": gin.H{
			"path": kubeconfigPath,
		},
		"sops_available": false, // SOPS detection not implemented
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

// Search resources
func (s *Server) handleSearch(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusOK, []k8s.SearchResult{})
		return
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	ctx := c.Request.Context()
	results, err := client.SearchResources(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Search failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, results)
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

	// For now, just save the config as-is
	// Future: support more sophisticated updates

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

// handleMetrics returns Prometheus-compatible metrics
func (s *Server) handleMetrics(c *gin.Context) {
	client := s.GetK8sClient()
	connected := client != nil

	// Get metrics values
	requestCount := s.requestCount.Load()
	errorCount := s.errorCount.Load()
	durationTotal := s.requestDuration.Load()
	uptime := time.Since(s.startTime).Seconds()

	// Calculate average request duration
	avgDuration := 0.0
	if requestCount > 0 {
		avgDuration = float64(durationTotal) / float64(requestCount) / 1e9 // Convert nanoseconds to seconds
	}

	// Prometheus format metrics
	metrics := fmt.Sprintf(`# HELP stargazer_requests_total Total number of HTTP requests
# TYPE stargazer_requests_total counter
stargazer_requests_total %d

# HELP stargazer_errors_total Total number of HTTP errors (4xx, 5xx)
# TYPE stargazer_errors_total counter
stargazer_errors_total %d

# HELP stargazer_request_duration_seconds Average request duration in seconds
# TYPE stargazer_request_duration_seconds gauge
stargazer_request_duration_seconds %.6f

# HELP stargazer_uptime_seconds Server uptime in seconds
# TYPE stargazer_uptime_seconds gauge
stargazer_uptime_seconds %.2f

# HELP stargazer_connected Whether stargazer is connected to Kubernetes cluster
# TYPE stargazer_connected gauge
stargazer_connected %d
`,
		requestCount,
		errorCount,
		avgDuration,
		uptime,
		map[bool]int{true: 1, false: 0}[connected],
	)

	c.Data(http.StatusOK, "text/plain; version=0.0.4", []byte(metrics))
}

func (s *Server) handleGetPodsEmpty(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "all"
	}
	c.JSON(http.StatusOK, gin.H{
		"pods":      []interface{}{},
		"count":     0,
		"namespace": namespace,
	})
}

func (s *Server) handleGetDeploymentsEmpty(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = "all"
	}
	c.JSON(http.StatusOK, gin.H{
		"deployments": []interface{}{},
		"count":       0,
		"namespace":   namespace,
	})
}

// Policy building and management handlers

// Build Cilium Network Policy
func (s *Server) handleBuildCiliumPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	var spec k8s.CiliumNetworkPolicySpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	yamlContent, err := builder.BuildCiliumNetworkPolicy(ctx, spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to build policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"yaml":      yamlContent,
		"name":      spec.Name,
		"namespace": spec.Namespace,
	})
}

// Apply Cilium Network Policy
func (s *Server) handleApplyCiliumPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	var req struct {
		YAML      string `json:"yaml" binding:"required"`
		Namespace string `json:"namespace"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	namespace := req.Namespace
	if namespace == "" {
		namespace = client.GetNamespace()
	}

	err := builder.ApplyCiliumNetworkPolicy(ctx, req.YAML, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to apply policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Policy applied successfully",
		"namespace": namespace,
	})
}

// Apply Network Policy
func (s *Server) handleApplyNetworkPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	var req struct {
		YAML      string `json:"yaml" binding:"required"`
		Namespace string `json:"namespace"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	namespace := req.Namespace
	if namespace == "" {
		namespace = client.GetNamespace()
	}

	err := builder.ApplyNetworkPolicy(ctx, req.YAML, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to apply policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Network Policy applied successfully",
		"namespace": namespace,
	})
}

// Export Cilium Network Policy (just returns the YAML)
func (s *Server) handleExportCiliumPolicy(c *gin.Context) {
	var req struct {
		YAML string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Return YAML as downloadable content
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", "attachment; filename=cilium-network-policy.yaml")
	c.Data(http.StatusOK, "application/x-yaml", []byte(req.YAML))
}

// Delete Cilium Network Policy
func (s *Server) handleDeleteCiliumPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	name := c.Param("name")
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = client.GetNamespace()
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	err := builder.DeleteCiliumNetworkPolicy(ctx, name, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Policy deleted successfully",
	})
}

// Build Kyverno Policy
func (s *Server) handleBuildKyvernoPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	var spec k8s.KyvernoPolicySpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	yamlContent, err := builder.BuildKyvernoPolicy(ctx, spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to build policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"yaml":      yamlContent,
		"name":      spec.Name,
		"namespace": spec.Namespace,
		"type":      spec.Type,
	})
}

// Apply Kyverno Policy
func (s *Server) handleApplyKyvernoPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	var req struct {
		YAML      string `json:"yaml" binding:"required"`
		Namespace string `json:"namespace"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	namespace := req.Namespace
	if namespace == "" {
		namespace = client.GetNamespace()
	}

	err := builder.ApplyKyvernoPolicy(ctx, req.YAML, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to apply policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Policy applied successfully",
		"namespace": namespace,
	})
}

// Get recommendations based on best practices
func (s *Server) handleGetRecommendations(c *gin.Context) {
	namespace := c.Query("namespace")
	if err := validateNamespace(namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid namespace parameter: %v", err),
		})
		return
	}
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	ctx := c.Request.Context()

	// Get topology first
	topology, err := client.GetTopology(ctx, ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get topology: %v", err),
		})
		return
	}

	// Get recommendations
	recommendations, err := client.GetRecommendations(ctx, topology)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get recommendations: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"count":           len(recommendations),
		"namespace":       namespace,
	})
}

// Get compliance score
func (s *Server) handleGetComplianceScore(c *gin.Context) {
	namespace := c.Query("namespace")
	if err := validateNamespace(namespace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid namespace parameter: %v", err),
		})
		return
	}
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kubernetes client not available",
		})
		return
	}

	ctx := c.Request.Context()

	// Get topology first
	topology, err := client.GetTopology(ctx, ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get topology: %v", err),
		})
		return
	}

	// Get compliance score
	score, scoreDetails := client.GetComplianceScore(ctx, topology)

	c.JSON(http.StatusOK, gin.H{
		"score":                 score,
		"passed":                scoreDetails["passed"],
		"total":                 scoreDetails["total"],
		"details":               scoreDetails["check_details"],
		"recommendations_count": scoreDetails["recommendations_count"],
		"namespace":             namespace,
	})
}

// Export Kyverno Policy (just returns the YAML)
func (s *Server) handleExportKyvernoPolicy(c *gin.Context) {
	var req struct {
		YAML string `json:"yaml" binding:"required"`
		Type string `json:"type"` // "Policy" or "ClusterPolicy"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Determine filename
	filename := "kyverno-policy.yaml"
	if req.Type == "ClusterPolicy" {
		filename = "kyverno-cluster-policy.yaml"
	}

	// Return YAML as downloadable content
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/x-yaml", []byte(req.YAML))
}

// Delete Kyverno Policy
func (s *Server) handleDeleteKyvernoPolicy(c *gin.Context) {
	client := s.GetK8sClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Kubernetes client not initialized",
		})
		return
	}

	name := c.Param("name")
	namespace := c.Query("namespace")
	isClusterPolicy := c.Query("cluster_policy") == "true"

	if namespace == "" && !isClusterPolicy {
		namespace = client.GetNamespace()
	}

	builder := client.GetPolicyBuilder()
	ctx := c.Request.Context()

	err := builder.DeleteKyvernoPolicy(ctx, name, namespace, isClusterPolicy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete policy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Policy deleted successfully",
	})
}

// Set provider model
func (s *Server) handleSetProviderModel(c *gin.Context) {
	provider := c.Param("provider")
	var req struct {
		Model string `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// TODO: Store provider model in config when provider config is implemented
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"provider": provider,
		"model":    req.Model,
		"message":  fmt.Sprintf("Provider %s model set to %s", provider, req.Model),
	})
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
	// TODO: Store provider enabled state in config when provider config is implemented
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"provider": provider,
		"enabled":  req.Enabled,
		"message":  fmt.Sprintf("Provider %s %s", provider, map[bool]string{true: "enabled", false: "disabled"}[req.Enabled]),
	})
}

// Set provider API key
func (s *Server) handleSetProviderApiKey(c *gin.Context) {
	provider := c.Param("provider")
	var req struct {
		ApiKey string `json:"api_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// TODO: Store provider API key securely in config when provider config is implemented
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"provider": provider,
		"message":  fmt.Sprintf("API key for provider %s updated", provider),
	})
}
