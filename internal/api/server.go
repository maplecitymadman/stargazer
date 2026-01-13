package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maplecitymadman/stargazer/internal/k8s"
)

// Server wraps the HTTP server with Kubernetes client and WebSocket support
type Server struct {
	router    *gin.Engine
	k8sClient *k8s.Client
	discovery *k8s.Discovery
	wsHub     *Hub // WebSocket hub for broadcasting
	mu        sync.RWMutex // Protects k8sClient and discovery
}

var (
	serverMu sync.RWMutex // Global mutex for server operations
)


// Config holds server configuration
type Config struct {
	Port          int
	K8sClient     *k8s.Client
	EnableCORS    bool
	RateLimitRPS  int // Requests per second
}

// NewServer creates a new API server
func NewServer(cfg Config) *Server {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add custom logger middleware
	router.Use(loggerMiddleware())

	// Create WebSocket hub
	wsHub := NewHub()
	go wsHub.Run()

	server := &Server{
		router:    router,
		k8sClient: cfg.K8sClient,
		discovery: k8s.NewDiscovery(cfg.K8sClient),
		wsHub:     wsHub,
	}

	// Setup middleware
	if cfg.EnableCORS {
		router.Use(corsMiddleware())
	}
	router.Use(securityHeadersMiddleware())

	if cfg.RateLimitRPS > 0 {
		router.Use(rateLimitMiddleware(cfg.RateLimitRPS))
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// UpdateK8sClient updates the Kubernetes client (for kubeconfig changes)
func (s *Server) UpdateK8sClient(client *k8s.Client) {
	serverMu.Lock()
	defer serverMu.Unlock()
	s.k8sClient = client
	if client != nil {
		s.discovery = k8s.NewDiscovery(client)
	} else {
		s.discovery = nil
	}
}

// GetK8sClient returns the current Kubernetes client (thread-safe)
func (s *Server) GetK8sClient() *k8s.Client {
	serverMu.RLock()
	defer serverMu.RUnlock()
	return s.k8sClient
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check (no rate limiting)
	s.router.GET("/api/health", s.handleHealth)

	// API routes
	api := s.router.Group("/api")
	{
		// Cluster endpoints (v2 style)
		api.GET("/cluster/health", s.handleClusterHealth)
		api.GET("/cluster/issues", s.handleClusterIssues)

		// Context management
		api.GET("/contexts", s.handleGetContexts)
		api.GET("/context/current", s.handleGetCurrentContext)
		api.POST("/context/switch", s.handleSwitchContext)

		// Discovery & scanning
		api.GET("/issues", s.handleGetIssues)
		api.POST("/scan", s.handleScan)

		// Namespace
		api.GET("/namespace", s.handleGetNamespace)
		api.GET("/namespaces", s.handleGetNamespaces)

		// Resources (v2 style - query params instead of path params)
		api.GET("/pods", s.handleGetPodsQuery)
		api.GET("/deployments", s.handleGetDeploymentsQuery)
		api.GET("/events", s.handleGetEventsQuery)
		
		// Legacy path-based routes (for compatibility)
		api.GET("/namespaces/:namespace/pods", s.handleGetPods)
		api.GET("/namespaces/:namespace/pods/:pod", s.handleGetPod)
		api.GET("/namespaces/:namespace/pods/:pod/logs", s.handleGetPodLogs)
		api.GET("/namespaces/:namespace/deployments", s.handleGetDeployments)
		api.GET("/namespaces/:namespace/services", s.handleGetServices)
		api.GET("/namespaces/:namespace/events", s.handleGetEvents)
		api.GET("/nodes", s.handleGetNodes)

		// Troubleshooting (Phase 6)
		api.POST("/troubleshoot", s.handleTroubleshoot)

		// Topology (Phase 2 deferred)
		api.GET("/topology", s.handleGetTopology)
		api.GET("/topology/:service_name", s.handleGetServiceConnections)
		api.GET("/networkpolicy/:policy_name", s.handleGetNetworkPolicyYaml)
		api.GET("/resources/:resource_type/:resource_name/yaml", s.handleGetResourceYaml)

		// Config endpoints (v2 style)
		api.GET("/config", s.handleGetConfig)
		api.GET("/config/providers", s.handleGetProvidersConfig)
		api.GET("/config/kubeconfig/status", s.handleGetKubeconfigStatus)
		api.POST("/config/kubeconfig", s.handleSetKubeconfig)
		api.POST("/config/providers/:provider/model", s.handleSetProviderModel)
		api.POST("/config/providers/:provider/enable", s.handleEnableProvider)
		api.POST("/config/providers/:provider/api-key", s.handleSetProviderApiKey)
		api.POST("/config", s.handleSetConfig)
		api.POST("/config/setup-wizard", s.handleSetupWizard)
	}

	// WebSocket endpoint
	s.router.GET("/ws", func(c *gin.Context) {
		s.handleWebSocket(c.Writer, c.Request)
	})

	// Static files (frontend) will be added when we embed
	// For now, serve a simple message
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Stargazer API Server",
			"version": "0.1.0-dev",
			"endpoints": gin.H{
				"health":    "/api/health",
				"scan":      "/api/scan",
				"issues":    "/api/issues",
				"websocket": "/ws",
			},
		})
	})
}

// Start starts the HTTP server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🚀 Starting Stargazer API server on http://localhost%s\n", addr)
	fmt.Printf("📡 WebSocket endpoint: ws://localhost%s/ws\n", addr)
	fmt.Printf("🔍 API docs: http://localhost%s/api/health\n", addr)

	return s.router.Run(addr)
}

// GetRouter returns the underlying Gin router (for use with custom HTTP server)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close WebSocket hub
	s.wsHub.Shutdown()
	return nil
}

// loggerMiddleware provides custom logging
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		// Skip logging for health checks to reduce noise
		if path == "/api/health" {
			return
		}

		fmt.Printf("[API] %s %s %d %v\n", method, path, statusCode, latency)
	}
}
