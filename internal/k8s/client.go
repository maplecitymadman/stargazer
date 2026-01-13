package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes client-go with caching and helper methods
type Client struct {
	clientset       *kubernetes.Clientset
	namespace       string
	cache           *cache
	kubeconfigPath  string
	context         string
	restConfig      *rest.Config
}

// cache holds cached K8s API responses with TTL
type cache struct {
	mu       sync.RWMutex
	data     map[string]cacheEntry
	cacheTTL time.Duration
}

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

// DiscoverKubeconfigPath attempts to discover kubeconfig path without creating a client
// This is useful for status checks and UI display
func DiscoverKubeconfigPath() (string, error) {
	return resolveKubeconfigPath("")
}

// NewClient creates a new Kubernetes client
// If kubeconfigPath is empty, it tries: KUBECONFIG env var, then ~/.kube/config
// If contextName is empty, it uses the current context from kubeconfig
func NewClient(kubeconfigPath, contextName, namespace string) (*Client, error) {
	// Build config using client-go's default loading rules
	// This handles all the standard kubeconfig discovery automatically
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	
	// If explicit path provided, use it; otherwise let client-go discover it
	var resolvedPath string
	if kubeconfigPath != "" {
		// Use explicitly provided path
		expanded := expandPath(kubeconfigPath)
		if _, err := os.Stat(expanded); os.IsNotExist(err) {
			return nil, fmt.Errorf("kubeconfig file not found at specified path: %s", expanded)
		}
		resolvedPath = expanded
		loadingRules.ExplicitPath = resolvedPath
	} else {
		// Let client-go use its default discovery (same as kubectl)
		// This will check KUBECONFIG env var, then ~/.kube/config
		// Don't set ExplicitPath - let client-go handle discovery
		resolvedPath = loadingRules.GetDefaultFilename()
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		configOverrides.CurrentContext = contextName
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		// Try to get the actual path that was used
		actualPath := loadingRules.GetDefaultFilename()
		if loadingRules.ExplicitPath != "" {
			actualPath = loadingRules.ExplicitPath
		}
		
		// Provide helpful error message
		home := homedir.HomeDir()
		defaultPath := ""
		if home != "" {
			defaultPath = filepath.Join(home, ".kube", "config")
		}
		
		return nil, fmt.Errorf(`failed to load kubeconfig from %s: %w

To fix:
1. Set kubeconfig path in Settings → Cluster
2. Or ensure kubeconfig exists at: %s
3. Or set KUBECONFIG environment variable`, 
			actualPath, err, defaultPath)
	}
	
	// Get the actual path that was loaded (for logging/storage)
	// After successful load, get the actual path from the loading rules
	if resolvedPath == "" || loadingRules.ExplicitPath == "" {
		// The actual path used will be in GetDefaultFilename() or ExplicitPath
		if loadingRules.ExplicitPath != "" {
			resolvedPath = loadingRules.ExplicitPath
		} else {
			resolvedPath = loadingRules.GetDefaultFilename()
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	// Determine cache TTL from environment (default: 30s)
	cacheTTL := 30 * time.Second
	if ttlStr := os.Getenv("CACHE_TTL"); ttlStr != "" {
		if duration, err := time.ParseDuration(ttlStr + "s"); err == nil {
			cacheTTL = duration
		}
	}

	// Default namespace
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}
	}

	client := &Client{
		clientset:      clientset,
		namespace:      namespace,
		kubeconfigPath: resolvedPath,
		context:        contextName,
		restConfig:     config,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: cacheTTL,
		},
	}

	return client, nil
}

// resolveKubeconfigPath finds the kubeconfig file from various sources
// It checks in order: explicit path, config file, KUBECONFIG env (with multiple paths), default locations
func resolveKubeconfigPath(explicitPath string) (string, error) {
	var triedPaths []string

	// 1. Use explicitly provided path
	if explicitPath != "" {
		expanded := expandPath(explicitPath)
		triedPaths = append(triedPaths, expanded)
		if _, err := os.Stat(expanded); err == nil {
			return expanded, nil
		}
	}

	// 2. Check config file (if available)
	if configPath := getConfigKubeconfigPath(); configPath != "" {
		expanded := expandPath(configPath)
		triedPaths = append(triedPaths, expanded)
		if _, err := os.Stat(expanded); err == nil {
			return expanded, nil
		}
	}

	// 3. Check KUBECONFIG environment variable (can contain multiple paths separated by : or ;)
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		// Split by : on Unix or ; on Windows
		separator := ":"
		if runtime.GOOS == "windows" {
			separator = ";"
		}
		
		paths := strings.Split(envPath, separator)
		for _, path := range paths {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			expanded := expandPath(path)
			triedPaths = append(triedPaths, expanded)
			if _, err := os.Stat(expanded); err == nil {
				return expanded, nil
			}
		}
	}

	// 4. Try default locations in order of preference
	if home := homedir.HomeDir(); home != "" {
		defaultPaths := []string{
			filepath.Join(home, ".kube", "config"),           // Standard location
			filepath.Join(home, ".kube", "kubeconfig"),       // Alternative name
			filepath.Join(home, "kubeconfig"),                // Home directory
			filepath.Join(home, ".kube", "config.yaml"),      // YAML extension
			filepath.Join(home, ".kube", "config.yml"),       // YML extension
		}
		
		for _, defaultPath := range defaultPaths {
			triedPaths = append(triedPaths, defaultPath)
			if _, err := os.Stat(defaultPath); err == nil {
				return defaultPath, nil
			}
		}
	}

	// 5. Check current directory and common locations
	currentDir, _ := os.Getwd()
	commonPaths := []string{
		filepath.Join(currentDir, "kubeconfig"),
		filepath.Join(currentDir, ".kube", "config"),
		filepath.Join(currentDir, "config"),
		"/etc/kubernetes/admin.conf",                          // Common in kubeadm setups
		"/etc/kubernetes/kubeconfig",                          // Alternative system location
	}
	
	for _, commonPath := range commonPaths {
		triedPaths = append(triedPaths, commonPath)
		if _, err := os.Stat(commonPath); err == nil {
			return commonPath, nil
		}
	}

	// 6. Try using client-go's default loading rules (as fallback)
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if loadingRules.GetLoadingPrecedence() != nil {
		for _, path := range loadingRules.GetLoadingPrecedence() {
			if path != "" {
				expanded := expandPath(path)
				triedPaths = append(triedPaths, expanded)
				if _, err := os.Stat(expanded); err == nil {
					return expanded, nil
				}
			}
		}
	}

	// 7. Nothing found - return helpful error with all tried paths
	triedList := strings.Join(triedPaths, "\n  - ")
	return "", fmt.Errorf(`could not find Kubernetes kubeconfig file

Tried the following locations:
  - %s

To fix this:
  1. Configure kubeconfig path in Settings → Cluster → Kubeconfig Path
  2. Set KUBECONFIG environment variable: export KUBECONFIG=/path/to/kubeconfig
  3. Place kubeconfig at ~/.kube/config
  4. Use kubectl to verify your kubeconfig: kubectl config view`,
		triedList)
}

// expandPath expands environment variables and ~ in a path
func expandPath(path string) string {
	// Trim whitespace
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}
	
	// Expand environment variables
	expanded := os.ExpandEnv(path)
	
	// Handle ~ for home directory
	if len(expanded) > 0 && expanded[0] == '~' {
		home := homedir.HomeDir()
		if home != "" {
			if len(expanded) == 1 {
				return home
			}
		// Handle ~/path and ~user/path
		if expanded[1] == '/' || expanded[1] == filepath.Separator {
			expanded = filepath.Join(home, expanded[2:])
		}
		// Note: ~user format is left as-is (would require user.Lookup)
		}
	}
	
	// Clean the path (resolve . and ..)
	expanded = filepath.Clean(expanded)
	
	return expanded
}

// getConfigKubeconfigPath loads kubeconfig path from config file
// Uses a simple approach to avoid circular dependency with config package
func getConfigKubeconfigPath() string {
	configPath := filepath.Join(homedir.HomeDir(), ".stargazer", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return ""
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	// Simple YAML parsing - look for kubeconfig.path
	content := string(data)
	lines := strings.Split(content, "\n")
	
	inKubeconfig := false
	indentLevel := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		// Check if we're entering kubeconfig section
		if strings.HasPrefix(trimmed, "kubeconfig:") {
			inKubeconfig = true
			indentLevel = len(line) - len(strings.TrimLeft(line, " \t"))
			continue
		}
		
		// Check if we've left the kubeconfig section
		if inKubeconfig {
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			if currentIndent <= indentLevel && trimmed != "" {
				inKubeconfig = false
				continue
			}
		}
		
		// Look for path field within kubeconfig section
		if inKubeconfig && strings.HasPrefix(trimmed, "path:") {
			parts := strings.SplitN(trimmed, "path:", 2)
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				path = strings.Trim(path, `"`)
				path = strings.Trim(path, "'")
				if path != "" && path != "null" {
					return path
				}
			}
		}
	}

	return ""
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetNamespace returns the configured namespace
func (c *Client) GetNamespace() string {
	return c.namespace
}

// SetNamespace updates the default namespace
func (c *Client) SetNamespace(namespace string) {
	c.namespace = namespace
}

// GetContext returns the current Kubernetes context
func (c *Client) GetContext() string {
	return c.context
}

// GetKubeconfigPath returns the path to the kubeconfig file
func (c *Client) GetKubeconfigPath() string {
	return c.kubeconfigPath
}

// Cache methods

func (c *cache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid
	if time.Since(entry.timestamp) > c.cacheTTL {
		return nil, false
	}

	return entry.data, true
}

func (c *cache) set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

func (c *cache) invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (c *cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheEntry)
}

// ClearCache clears all cached data
func (c *Client) ClearCache() {
	c.cache.clear()
}

// Health checks if the client can connect to the cluster
func (c *Client) Health(ctx context.Context) error {
	_, err := c.clientset.Discovery().ServerVersion()
	return err
}
