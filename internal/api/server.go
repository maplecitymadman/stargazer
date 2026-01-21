package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maplecitymadman/stargazer/internal/k8s"
)

// Fix Issue #4: Handle JSON marshaling errors properly
func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		// Log marshaling error
		slog.Error("JSON marshaling failed", "error", err)
		return "{\"error\": \"marshaling failed\"}"
	}
	return string(b)
}

// Server wraps the HTTP server with Kubernetes client and WebSocket support
type Server struct {
	router        *gin.Engine
	k8sClient     *k8s.Client
	discovery     *k8s.Discovery
	wsHub         *Hub               // WebSocket hub for broadcasting
	policyWatcher *k8s.PolicyWatcher // Policy watcher for real-time updates
	mu            sync.RWMutex       // Protects k8sClient and discovery

	// Fix Issue #3: Add cancel function for policy watcher context
	policyWatcherCancel context.CancelFunc

	// Metrics
	requestCount    atomic.Uint64
	errorCount      atomic.Uint64
	requestDuration atomic.Uint64 // Total duration in nanoseconds
	startTime       time.Time
}

// Config holds server configuration
type Config struct {
	Port         int
	K8sClient    *k8s.Client
	EnableCORS   bool
	RateLimitRPS int // Requests per second
}

// NewServer creates a new API server
func NewServer(cfg Config) *Server {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Create server instance first (we need it for middleware)
	server := &Server{
		router:    router,
		k8sClient: cfg.K8sClient,
		discovery: k8s.NewDiscovery(cfg.K8sClient),
		wsHub:     NewHub(),
		startTime: time.Now(),
	}

	// Start WebSocket hub
	go server.wsHub.Run()

	// Start policy watcher if K8s client is available
	if cfg.K8sClient != nil {
		go server.startPolicyWatcher()
	}

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add custom logger middleware (with metrics)
	router.Use(server.loggerMiddleware())

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
	// Fix Issue #1: Use instance-level mutex instead of global serverMu
	s.mu.Lock()
	defer s.mu.Unlock()
	s.k8sClient = client
	if client != nil {
		s.discovery = k8s.NewDiscovery(client)
	} else {
		s.discovery = nil
	}
}

// GetK8sClient returns the current Kubernetes client (thread-safe)
func (s *Server) GetK8sClient() *k8s.Client {
	// Fix Issue #1: Use instance-level mutex instead of global serverMu
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.k8sClient
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check (no rate limiting)
	s.router.GET("/api/health", s.handleHealth)

	// Metrics endpoint for Prometheus (no rate limiting)
	s.router.GET("/api/metrics", s.handleMetrics)

	// API routes
	api := s.router.Group("/api")
	{
		// Cluster endpoints (v2 style)
		api.GET("/cluster/health", s.handleClusterHealth)
		// Return empty issues array for compatibility (frontend expects this)
		api.GET("/cluster/issues", s.handleGetIssuesEmpty)

		// Context management
		api.GET("/contexts", s.handleGetContexts)
		api.GET("/context/current", s.handleGetCurrentContext)
		api.POST("/context/switch", s.handleSwitchContext)

		// Namespace
		api.GET("/namespace", s.handleGetNamespace)
		api.GET("/namespaces", s.handleGetNamespaces)

		// Events
		api.GET("/events", s.handleGetEventsQuery)

		// Return empty arrays for removed endpoints (frontend compatibility)
		api.GET("/pods", s.handleGetPodsEmpty)
		api.GET("/deployments", s.handleGetDeploymentsEmpty)

		// Legacy path-based routes (for compatibility)
		api.GET("/namespaces/:namespace/services", s.handleGetServices)
		api.GET("/services", s.handleGetAllServices) // Get all services for PathTracer
		api.GET("/namespaces/:namespace/events", s.handleGetEvents)
		api.GET("/nodes", s.handleGetNodes)

		// Troubleshooting
		api.GET("/troubleshoot", s.handleTroubleshoot)

		// Search
		api.GET("/search", s.handleSearch)

		// Topology (Phase 2 deferred) - expensive endpoints, stricter rate limiting
		topologyGroup := api.Group("/topology")
		topologyGroup.Use(rateLimitMiddleware(5)) // 5 req/min for expensive topology calls
		{
			topologyGroup.GET("", s.handleGetTopology)
			topologyGroup.GET("/:service_name", s.handleGetServiceConnections)
			topologyGroup.GET("/trace", s.handleTracePath)
		}
		api.GET("/networkpolicy/:policy_name", s.handleGetNetworkPolicyYaml)

		// Recommendations - expensive endpoints, stricter rate limiting
		recommendationsGroup := api.Group("/recommendations")
		recommendationsGroup.Use(rateLimitMiddleware(10)) // 10 req/min
		{
			recommendationsGroup.GET("", s.handleGetRecommendations)
			recommendationsGroup.GET("/score", s.handleGetComplianceScore)
		}

		// Policy building and management
		api.POST("/policies/network/apply", s.handleApplyNetworkPolicy)
		api.POST("/policies/cilium/build", s.handleBuildCiliumPolicy)
		api.POST("/policies/cilium/apply", s.handleApplyCiliumPolicy)
		api.POST("/policies/cilium/export", s.handleExportCiliumPolicy)
		api.DELETE("/policies/cilium/:name", s.handleDeleteCiliumPolicy)
		api.POST("/policies/kyverno/build", s.handleBuildKyvernoPolicy)
		api.POST("/policies/kyverno/apply", s.handleApplyKyvernoPolicy)
		api.POST("/policies/kyverno/export", s.handleExportKyvernoPolicy)
		api.DELETE("/policies/kyverno/:name", s.handleDeleteKyvernoPolicy)

		// Config endpoints (v2 style)
		api.GET("/config", s.handleGetConfig)

		api.GET("/config/kubeconfig/status", s.handleGetKubeconfigStatus)
		api.POST("/config/kubeconfig", s.handleSetKubeconfig)

		api.POST("/config", s.handleSetConfig)
		api.POST("/config/setup-wizard", s.handleSetupWizard)
	}

	// WebSocket endpoint
	s.router.GET("/ws", func(c *gin.Context) {
		s.handleWebSocket(c.Writer, c.Request)
	})

	// Serve static frontend files from /app/frontend/out
	s.router.Static("/_next/static", "/app/frontend/out/_next/static")
	s.router.StaticFile("/favicon.ico", "/app/frontend/out/favicon.ico")

	// Serve index.html for root and all other routes (SPA routing)
	s.router.NoRoute(func(c *gin.Context) {
		// If it's an API route, return 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		// Otherwise serve the frontend
		c.File("/app/frontend/out/index.html")
	})

	// Serve root as index.html
	s.router.GET("/", func(c *gin.Context) {
		c.File("/app/frontend/out/index.html")
	})
}

// Start starts the HTTP server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	slog.Info("Starting Stargazer API server",
		"address", addr,
		"websocket", fmt.Sprintf("ws://localhost%s/ws", addr),
		"health_endpoint", fmt.Sprintf("http://localhost%s/api/health", addr))

	return s.router.Run(addr)
}

// GetRouter returns the underlying Gin router (for use with custom HTTP server)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop policy watcher
	s.mu.Lock()
	if s.policyWatcher != nil {
		s.policyWatcher.Stop()
	}
	// Fix Issue #3: Cancel policy watcher context to properly clean up goroutines
	if s.policyWatcherCancel != nil {
		s.policyWatcherCancel()
	}
	s.mu.Unlock()

	// Close WebSocket hub
	s.wsHub.Shutdown()
	return nil
}

// startPolicyWatcher starts watching for policy changes and broadcasts via WebSocket
func (s *Server) startPolicyWatcher() {
	client := s.GetK8sClient()
	if client == nil {
		return
	}

	s.mu.Lock()
	s.policyWatcher = client.NewPolicyWatcher()
	s.mu.Unlock()

	s.policyWatcher.OnPolicyChange(func(eventType, policyType, name, namespace string) {
		// Broadcast policy change event via WebSocket
		message := Message{
			Type: "policy_change",
			Data: map[string]interface{}{
				"event_type":  eventType,
				"policy_type": policyType,
				"name":        name,
				"namespace":   namespace,
				"timestamp":   time.Now().Unix(),
			},
		}
		s.wsHub.Broadcast(message)
	})

	// Fix Issue #3: Use cancellable context instead of context.Background()
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.policyWatcherCancel = cancel
	s.mu.Unlock()

	if err := s.policyWatcher.Start(ctx); err != nil {
		slog.Error("Policy watcher failed to start", "error", err)
		return
	}

	// Keep watcher running (will be stopped when server shuts down via context cancellation)
	// The goroutines started by watcher.Start() will run until ctx is cancelled
	<-ctx.Done()
}

// loggerMiddleware provides custom logging and metrics tracking
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		// Update metrics
		s.requestCount.Add(1)
		s.requestDuration.Add(uint64(latency.Nanoseconds()))
		if statusCode >= 400 {
			s.errorCount.Add(1)
		}

		// Log API requests (skip health checks to reduce noise)
		if path != "/api/health" {
			slog.Info("API request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency)
		}
	}
}
