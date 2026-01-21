# Stargazer - Kubernetes Troubleshooting Tool

A **CLI tool** for Kubernetes troubleshooting. Connects to any cluster via kubeconfig - no cluster deployment needed!

## 🌟 Features

- **Standalone Binary**: Single binary that works with any Kubernetes cluster via kubeconfig
- **No Cluster Deployment Required**: Runs locally, connects to remote or local clusters
- **Auto-Discovery**: Scans cluster for issues automatically
- **Multi-Cluster Support**: Switch between Kubernetes contexts
- **Namespace Filtering**: View resources by namespace or cluster-wide
- **Theme Support**: Dark, Light, and Auto themes
- **AI-Powered Troubleshooting**: Configure multiple LLM providers
- **CLI Interface**: Simple command-line interface
- **Read-Only Permissions**: Safe for production environments
- **Fast & Lightweight**: Minimal resource usage, fast startup

## 🚀 Quick Start

```bash
# Build CLI
make build

# Use commands
./bin/stargazer health
./bin/stargazer scan
./bin/stargazer logs my-pod
```

### First Run

```bash
# Verify cluster connection
./bin/stargazer health

# Scan for issues
./bin/stargazer scan
```

## 📋 CLI Commands

```bash
# Health check
stargazer health

# Scan cluster for issues
stargazer scan

# Get pod logs
stargazer logs <pod-name> [--namespace <ns>] [--lines <n>]

# List pods
stargazer pods [--namespace <ns>]

# List deployments
stargazer deployments [--namespace <ns>]

# Get events
stargazer events [--namespace <ns>]

# Configuration
stargazer config setup    # Interactive setup wizard
stargazer config show     # Show current configuration
```

## 🏗️ Architecture

```
stargazer/
├── cmd/
│   └── stargazer/        # CLI entry point
├── internal/
│   ├── api/              # HTTP server & WebSocket
│   ├── k8s/              # Kubernetes client
│   ├── config/           # Configuration management
│   └── storage/          # Local persistence
├── frontend/             # React/Next.js UI (optional web interface)
│   ├── app/              # Next.js app directory
│   ├── components/       # React components
│   └── lib/              # API client
├── go.mod                # Go dependencies
└── Makefile             # Build automation
```

## 🛠️ Technology Stack

- **Backend**: Go 1.21+
- **Frontend**: Next.js 14, React 18, Tailwind CSS (optional web UI)
- **K8s**: client-go (official Kubernetes Go client)
- **Storage**: JSON file persistence (~/.stargazer/)

## 📊 Features

### CLI
- Health checks
- Cluster scanning
- Pod log retrieval
- Resource listing
- Configuration management

## 🔒 Security

- **Read-Only Access**: Only reads cluster state, never modifies
- **Local Storage**: All data stored locally in `~/.stargazer/`
- **Kubeconfig**: Uses standard Kubernetes authentication
- **No External Calls**: Fully self-contained (except configured AI providers)

## 📈 Performance

- **Efficient Caching**: 30s TTL for API responses
- **Background Scanning**: Minimal overhead
- **Compact Data Models**: Optimized for speed
- **Async Operations**: Non-blocking I/O

## 🐛 Development

### Prerequisites

- Go 1.21+
- Node.js 16+ (optional, for web UI development)
- kubectl (for Kubernetes access)

### Building

```bash
# Build CLI
make build

# Run in development mode
make dev
```

### Testing

```bash
# Run Go tests
make test

# Run with coverage
make test-coverage
```

## 📝 Configuration

Configuration is stored in `~/.stargazer/config.yaml`:

- **Kubeconfig**: Auto-detected from `~/.kube/config` or `$KUBECONFIG`
- **AI Providers**: Configure in Settings UI or via config file
- **API Settings**: Rate limiting, CORS, etc.

### Environment Variables

- `KUBECONFIG`: Path to kubeconfig file (auto-detected if not set)
- `KUBECTL_CONTEXT`: Kubernetes context to use
- `LOG_LEVEL`: Logging level (default: INFO)
- `CACHE_TTL`: API cache TTL in seconds (default: 30)

## 🚀 Distribution

### Homebrew (macOS)

```bash
# Install from formula
brew install --build-from-source stargazer.rb
```

### Manual Installation

1. Download binary from releases
2. Add to PATH
3. Run `stargazer config setup` for initial configuration

## 📄 License

MIT License - see LICENSE file for details

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

---

**Stargazer** - Making Kubernetes troubleshooting as easy as looking at the stars ✨
