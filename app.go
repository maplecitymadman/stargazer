package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/maplecitymadman/stargazer/internal/api"
	"github.com/maplecitymadman/stargazer/internal/config"
	"github.com/maplecitymadman/stargazer/internal/k8s"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/out
var assets embed.FS

// App struct for Wails desktop app
type App struct {
	ctx       context.Context
	k8sClient *k8s.Client
	server    *api.Server
	httpSrv   *http.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Load config first to get kubeconfig path
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}
	
	// Initialize K8s client with kubeconfig from config or auto-discovery
	kubeconfigPath := cfg.Kubeconfig.Path
	contextName := cfg.Kubeconfig.Context
	
	client, err := k8s.NewClient(kubeconfigPath, contextName, "")
	if err != nil {
		log.Printf("Warning: Could not initialize K8s client: %v", err)
		log.Println("Stargazer will work in limited mode. Configure kubeconfig in Settings to enable full functionality.")
		// Don't exit - allow user to configure kubeconfig in UI
	} else {
		a.k8sClient = client
		log.Printf("Connected to Kubernetes cluster (context: %s)", client.GetContext())
	}
	
	// Create API server (for HTTP endpoints)
	a.server = api.NewServer(api.Config{
		Port:         8000,
		K8sClient:    a.k8sClient,
		EnableCORS:   cfg.API.EnableCORS,
		RateLimitRPS: cfg.API.RateLimitRPS,
	})
	
	// Start HTTP server in background (frontend uses axios to call it)
	// Try port 8000, if busy try 8001-8010
	go func() {
		port := 8000
		var listener net.Listener
		var err error
		
		// Try to find an available port
		for port <= 8010 {
			listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err == nil {
				// Port is available
				break
			}
			log.Printf("Port %d in use, trying %d...", port, port+1)
			port++
		}
		
		if listener == nil {
			log.Printf("ERROR: Could not find available port (tried 8000-8010)")
			return
		}
		
		// Store the actual port we're using
		actualPort := listener.Addr().(*net.TCPAddr).Port
		log.Printf("✅ Starting API server on http://localhost:%d", actualPort)
		
		// Create HTTP server with our listener
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", actualPort),
			Handler: a.server.GetRouter(),
		}
		a.httpSrv = httpServer
		
		// Serve on the listener
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	
	// Give server a moment to start
	time.Sleep(300 * time.Millisecond)
}

// GetHealth returns cluster health status
func (a *App) GetHealth() map[string]interface{} {
	if a.k8sClient == nil {
		return map[string]interface{}{
			"error": "Kubernetes client not initialized. Please configure kubeconfig.",
		}
	}
	
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	
	if err := a.k8sClient.Health(ctx); err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"status":  "connected",
		"context": a.k8sClient.GetContext(),
	}
}

// GetIssues returns discovered issues
func (a *App) GetIssues() ([]interface{}, error) {
	if a.k8sClient == nil {
		return nil, fmt.Errorf("Kubernetes client not initialized")
	}
	
	discovery := k8s.NewDiscovery(a.k8sClient)
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	
	issues, err := discovery.ScanAll(ctx, "")
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice for JSON
	result := make([]interface{}, len(issues))
	for i, issue := range issues {
		result[i] = issue
	}
	
	return result, nil
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Stargazer",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(ctx context.Context) {
			// Shutdown HTTP server gracefully
			if app.httpSrv != nil {
				app.httpSrv.Shutdown(ctx)
			}
			if app.server != nil {
				app.server.Shutdown(ctx)
			}
		},
	})
	
	if err != nil {
		log.Fatal(err)
	}
}
