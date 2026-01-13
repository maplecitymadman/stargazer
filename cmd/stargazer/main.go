package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/maplecitymadman/stargazer/internal/api"
	"github.com/maplecitymadman/stargazer/internal/config"
	"github.com/maplecitymadman/stargazer/internal/k8s"
	"github.com/maplecitymadman/stargazer/internal/storage"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0-dev"
)

var rootCmd = &cobra.Command{
	Use:   "stargazer",
	Short: "Stargazer - Kubernetes troubleshooting tool",
	Long: `Stargazer is a lightweight Kubernetes troubleshooting tool with AI-powered diagnostics.

It helps you quickly identify and resolve issues in your Kubernetes clusters by:
- Discovering common problems automatically
- Analyzing logs and events with AI assistance
- Providing actionable recommendations
- Visualizing service topology and dependencies`,
	Version: version,
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface",
	Long:  `Start the web interface on http://localhost:8000`,
	Run: func(cmd *cobra.Command, args []string) {
		runWeb(cmd)
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan cluster for issues",
	Long:  `Scan the Kubernetes cluster for common issues and misconfigurations`,
	Run: func(cmd *cobra.Command, args []string) {
		runScan(cmd)
	},
}

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask AI for troubleshooting help",
	Long:  `Ask the AI assistant for help troubleshooting your Kubernetes cluster`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		question := args[0]
		fmt.Printf("💬 Question: %s\n", question)
		fmt.Println("⚠️  AI troubleshooting not yet implemented - coming in Phase 4-5")
		// TODO: Phase 4-5 - Implement AI engine
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check cluster health",
	Long:  `Check the overall health status of the Kubernetes cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		runHealth(cmd)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs [pod-name]",
	Short: "Get logs from a pod",
	Long:  `Retrieve and display logs from a specific pod`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runLogs(cmd, args[0])
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Stargazer configuration",
	Long:  `Manage Stargazer configuration including LLM providers and API keys`,
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run interactive setup wizard",
	Long:  `Run the interactive setup wizard to configure LLM providers and API keys`,
	Run: func(cmd *cobra.Command, args []string) {
		wizard := config.NewWizard()
		cfg, err := wizard.Run()
		if err != nil {
			fmt.Printf("❌ Setup failed: %v\n", err)
			os.Exit(1)
		}
		_ = cfg // Config is already saved by wizard
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current Stargazer configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Failed to load config: %v\n", err)
			fmt.Println("💡 Run 'stargazer config setup' to create a configuration")
			os.Exit(1)
		}

		fmt.Println("⚙️  Current Configuration")
		fmt.Println("========================")
		fmt.Printf("Config Path: %s\n", config.DefaultConfigPath)
		fmt.Printf("Version: %s\n", cfg.Version)
		fmt.Printf("Created: %s\n", cfg.CreatedAt)
		fmt.Printf("Updated: %s\n\n", cfg.UpdatedAt)

		// API Settings
		fmt.Println("API Settings:")
		fmt.Printf("  Port: %d\n", cfg.API.Port)
		fmt.Printf("  CORS: %v\n", cfg.API.EnableCORS)
		fmt.Printf("  Rate Limit: %d req/s\n\n", cfg.API.RateLimitRPS)

		// Storage Settings
		fmt.Println("Storage:")
		fmt.Printf("  Path: %s\n", cfg.Storage.Path)
		fmt.Printf("  Retain Days: %d\n", cfg.Storage.RetainDays)
		fmt.Printf("  Max Results: %d\n\n", cfg.Storage.MaxScanResults)

		// LLM Providers
		fmt.Println("LLM Providers:")
		fmt.Printf("  Default: %s\n", cfg.LLM.DefaultProvider)

		enabledProviders := cfg.GetEnabledProviders()
		if len(enabledProviders) == 0 {
			fmt.Println("  No providers enabled")
		} else {
			fmt.Printf("  Enabled: %v\n", enabledProviders)
			for name, provider := range cfg.LLM.Providers {
				if provider.Enabled {
					fmt.Printf("    - %s: %s", name, provider.Model)
					if provider.APIKey != "" {
						fmt.Printf(" (API key configured)")
					}
					fmt.Println()
				}
			}
		}
	},
}

func init() {
	// Web command flags
	webCmd.Flags().IntP("port", "p", 8000, "Port to run web server on")

	// Scan command flags
	scanCmd.Flags().StringP("namespace", "n", "", "Namespace to scan (empty = all namespaces)")

	// Logs command flags
	logsCmd.Flags().StringP("namespace", "n", "default", "Namespace of the pod")
	logsCmd.Flags().IntP("tail", "t", 100, "Number of lines to tail")
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")

	// Global flags
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringP("context", "c", "", "Kubernetes context to use")

	// Add subcommands
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(configCmd)

	// Config subcommands
	configCmd.AddCommand(configSetupCmd)
	configCmd.AddCommand(configShowCmd)
}

func main() {
	// Always run CLI mode - GUI is built separately with Wails
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runScan performs a cluster scan for issues
func runScan(cmd *cobra.Command) {
	namespace, _ := cmd.Flags().GetString("namespace")
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	contextName, _ := cmd.Flags().GetString("context")

	// Display scan scope
	if namespace == "" || namespace == "all" {
		fmt.Println("🔍 Scanning all namespaces...")
	} else {
		fmt.Printf("🔍 Scanning namespace: %s\n", namespace)
	}

	// Create K8s client
	client, err := k8s.NewClient(kubeconfig, contextName, namespace)
	if err != nil {
		fmt.Printf("❌ Failed to connect to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Health(ctx); err != nil {
		fmt.Printf("❌ Cannot connect to cluster: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Connected to cluster")
	fmt.Println()

	// Create discovery engine
	discovery := k8s.NewDiscovery(client)

	// Run scan
	fmt.Println("🔬 Running discovery scan...")
	scanCtx, scanCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer scanCancel()

	issues, err := discovery.ScanAll(scanCtx, namespace)
	if err != nil {
		fmt.Printf("⚠️  Scan completed with errors: %v\n", err)
	}

	// Save scan results to storage
	// Load config to get storage path, or use default
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		cfg = config.DefaultConfig()
	}

	store, err := storage.NewStorage(cfg.Storage.Path)
	if err != nil {
		fmt.Printf("⚠️  Could not initialize storage: %v\n", err)
	} else {
		scanNamespace := namespace
		if scanNamespace == "" {
			scanNamespace = "all"
		}
		result, err := store.SaveScanResult(scanNamespace, issues)
		if err != nil {
			fmt.Printf("⚠️  Could not save scan results: %v\n", err)
		} else {
			fmt.Printf("💾 Scan results saved: %s\n", result.ID)
		}
	}

	// Display results
	fmt.Println()
	if len(issues) == 0 {
		fmt.Println("✅ No issues found! Cluster looks healthy.")
		return
	}

	fmt.Printf("Found %d issue(s):\n\n", len(issues))

	// Group issues by priority
	critical := []k8s.Issue{}
	warnings := []k8s.Issue{}
	info := []k8s.Issue{}

	for _, issue := range issues {
		switch issue.Priority {
		case k8s.PriorityCritical:
			critical = append(critical, issue)
		case k8s.PriorityWarning:
			warnings = append(warnings, issue)
		default:
			info = append(info, issue)
		}
	}

	// Display critical issues
	if len(critical) > 0 {
		fmt.Printf("🔴 CRITICAL (%d):\n", len(critical))
		for _, issue := range critical {
			fmt.Printf("  • %s\n", issue.Title)
			fmt.Printf("    %s/%s\n", issue.Namespace, issue.ResourceName)
			fmt.Printf("    %s\n", issue.Description)
			fmt.Println()
		}
	}

	// Display warnings
	if len(warnings) > 0 {
		fmt.Printf("⚠️  WARNING (%d):\n", len(warnings))
		for _, issue := range warnings {
			fmt.Printf("  • %s\n", issue.Title)
			fmt.Printf("    %s/%s\n", issue.Namespace, issue.ResourceName)
			fmt.Printf("    %s\n", issue.Description)
			fmt.Println()
		}
	}

	// Display info
	if len(info) > 0 {
		fmt.Printf("ℹ️  INFO (%d):\n", len(info))
		for _, issue := range info {
			fmt.Printf("  • %s\n", issue.Title)
			if issue.Description != "" {
				fmt.Printf("    %s\n", issue.Description)
			}
			fmt.Println()
		}
	}
}

// runWeb starts the web server
func runWeb(cmd *cobra.Command) {
	port, _ := cmd.Flags().GetInt("port")
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	contextName, _ := cmd.Flags().GetString("context")

	// Load configuration (optional - use defaults if not found)
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("ℹ️  No configuration found, using defaults")
		fmt.Println("💡 Run 'stargazer config setup' to customize settings")
		cfg = config.DefaultConfig()
	}

	// Override port from CLI flag if provided
	if cmd.Flags().Changed("port") {
		cfg.API.Port = port
	} else {
		port = cfg.API.Port
	}

	// Create K8s client
	client, err := k8s.NewClient(kubeconfig, contextName, "")
	if err != nil {
		fmt.Printf("❌ Failed to connect to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Health(ctx); err != nil {
		fmt.Printf("❌ Cannot connect to cluster: %v\n", err)
		fmt.Println("Hint: Check your kubeconfig and cluster connectivity")
		os.Exit(1)
	}

	fmt.Println("✅ Connected to Kubernetes cluster")
	fmt.Printf("📍 Context: %s\n", client.GetContext())
	fmt.Println()

	// Create and start API server with config settings
	server := api.NewServer(api.Config{
		Port:         port,
		K8sClient:    client,
		EnableCORS:   cfg.API.EnableCORS,
		RateLimitRPS: cfg.API.RateLimitRPS,
	})

	if err := server.Start(port); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
		os.Exit(1)
	}
}

// runHealth checks cluster health
func runHealth(cmd *cobra.Command) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	contextName, _ := cmd.Flags().GetString("context")

	fmt.Println("🏥 Checking cluster health...")
	fmt.Println()

	// Create K8s client
	client, err := k8s.NewClient(kubeconfig, contextName, "")
	if err != nil {
		fmt.Printf("❌ Failed to connect to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// Test cluster connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Health(ctx); err != nil {
		fmt.Printf("❌ Cluster unhealthy: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Cluster connection: OK")
	fmt.Printf("📍 Context: %s\n", client.GetContext())
	fmt.Printf("🔧 Kubeconfig: %s\n", client.GetKubeconfigPath())
	fmt.Println()

	// Get nodes
	nodes, err := client.GetNodes(ctx)
	if err != nil {
		fmt.Printf("⚠️  Could not get nodes: %v\n", err)
	} else {
		readyNodes := 0
		for _, node := range nodes {
			if node.Status == "Ready" {
				readyNodes++
			}
		}
		fmt.Printf("🖥️  Nodes: %d total, %d ready\n", len(nodes), readyNodes)

		if readyNodes < len(nodes) {
			fmt.Println("   ⚠️  Some nodes are not ready:")
			for _, node := range nodes {
				if node.Status != "Ready" {
					fmt.Printf("   - %s: %s\n", node.Name, node.Status)
				}
			}
		}
	}

	// Get namespaces
	namespaces, err := client.GetNamespaces(ctx)
	if err != nil {
		fmt.Printf("⚠️  Could not get namespaces: %v\n", err)
	} else {
		fmt.Printf("📦 Namespaces: %d\n", len(namespaces))
	}

	// Run quick discovery scan
	fmt.Println()
	fmt.Println("🔬 Running quick health scan...")
	discovery := k8s.NewDiscovery(client)

	scanCtx, scanCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer scanCancel()

	issues, err := discovery.ScanAll(scanCtx, "")
	if err != nil {
		fmt.Printf("⚠️  Scan completed with errors: %v\n", err)
	}

	// Count issues by priority
	critical := 0
	warnings := 0
	info := 0

	for _, issue := range issues {
		switch issue.Priority {
		case k8s.PriorityCritical:
			critical++
		case k8s.PriorityWarning:
			warnings++
		default:
			info++
		}
	}

	fmt.Println()
	if len(issues) == 0 {
		fmt.Println("✅ No issues found! Cluster is healthy.")
	} else {
		fmt.Printf("Found %d issue(s):\n", len(issues))
		if critical > 0 {
			fmt.Printf("  🔴 Critical: %d\n", critical)
		}
		if warnings > 0 {
			fmt.Printf("  ⚠️  Warnings: %d\n", warnings)
		}
		if info > 0 {
			fmt.Printf("  ℹ️  Info: %d\n", info)
		}
		fmt.Println()
		fmt.Println("💡 Run 'stargazer scan' for detailed issue information")
	}
}

// runLogs retrieves and displays pod logs
func runLogs(cmd *cobra.Command, podName string) {
	namespace, _ := cmd.Flags().GetString("namespace")
	tail, _ := cmd.Flags().GetInt("tail")
	follow, _ := cmd.Flags().GetBool("follow")
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	contextName, _ := cmd.Flags().GetString("context")

	fmt.Printf("📜 Getting logs for pod: %s (namespace: %s)\n", podName, namespace)

	// Create K8s client
	client, err := k8s.NewClient(kubeconfig, contextName, namespace)
	if err != nil {
		fmt.Printf("❌ Failed to connect to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Health(ctx); err != nil {
		fmt.Printf("❌ Cannot connect to cluster: %v\n", err)
		os.Exit(1)
	}

	// Get logs
	logsCtx, logsCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer logsCancel()

	logs, err := client.GetPodLogs(logsCtx, namespace, podName, tail, follow)
	if err != nil {
		fmt.Printf("❌ Failed to get logs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("--- Logs ---")
	fmt.Print(logs)
}
