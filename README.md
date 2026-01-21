# Stargazer - Kubernetes Troubleshooting Tool

A lightweight, efficient Kubernetes troubleshooting tool that works like "OpenCode deployed in a cluster".

## 🌟 Features

- **Single Container Deployment**: Simple, resource-efficient deployment (<50m CPU, <64Mi memory)
- **Auto-Discovery**: Scans cluster every 2 seconds for issues
- **Advanced Issue Detection**: Detects pod issues, deployment problems, service mesh (Istio), network policies, and policy violations (Kyverno)
- **Multi-Namespace Support**: Scan all namespaces or filter by specific namespace
- **Agent System**: Specialized troubleshooting agents (@discovery, @logs, @resource, @network, @security)
- **AI-Powered Troubleshooting**: Real AI integration (OpenAI, Anthropic, Ollama) with fallback to pattern-based recommendations
- **Dual Interface**: CLI + TUI with OpenCode-style interaction
- **Kubectl Command Execution**: Execute validated kubectl commands directly (read-only operations)
- **Read-Only Permissions**: Safe for production environments
- **Issue Persistence**: Stores issue history in persistent volume
- **Fast Startup**: Ready in under 5 seconds

## 🚀 Quick Start

### Development Mode

```bash
# Clone and run demo
git clone <repository>
cd stargazer
python standalone.py --demo

# Interactive mode
python standalone.py --interactive
```

### Production Deployment

```bash
# Build and deploy
chmod +x build.sh
./build.sh --deploy

# Or manual deployment
docker build -t stargazer:latest .
kubectl apply -f kustomization.yaml
```

## 🤖 Agent System

### Main Agent: @troubleshooter
- `scan` - Scan cluster for issues
- `health` - Get cluster health summary
- `analyze <resource>` - Analyze specific resource

### Specialized Agents

#### @discovery
- `pods` - List all pods with status
- `deployments` - Show deployment status
- `events` - Display recent events

#### @logs
- `get <pod-name> [lines]` - Get pod logs
- `errors <pod-name>` - Find error patterns in logs

#### @resource
- `top` - Resource usage analysis
- `pressure` - Check resource pressure
- `describe <type> <name>` - Describe resource

#### @network
- `connectivity` - Check network connectivity
- `policies` - Show network policies
- `endpoints` - Display service endpoints

#### @security
- `rbac` - Check RBAC permissions
- `secrets` - Analyze secret configurations
- `images` - Image security scanning

## 🎮 Interface Modes

### CLI Mode

```bash
# Scan cluster
stargazer scan --continuous --interval 2

# Interactive troubleshooting
stargazer ask

# Check health
stargazer health

# Get logs
stargazer logs web-app-123 --lines 100

# Execute agent command
stargazer exec "@discovery pods"
```

### TUI Mode

```bash
# Start TUI interface
stargazer start --mode tui

# Controls:
# Ctrl+R - Refresh
# Ctrl+H - Health summary
# Ctrl+S - Manual scan
# Tab - Navigate between widgets
# Ctrl+C - Quit
```

## 📋 Commands

### System Commands
- `/agents` - List available agents
- `/help` - Show help
- `/sessions` - Session management
- `/export` - Export issues

### Agent Commands
- `@agentname` - Switch to agent
- `@agentname command` - Execute on specific agent
- `!kubectl command` - Execute kubectl directly

### File References
- `@pod/web-app-123` - Reference specific pod
- `@deployment/api-gateway` - Reference deployment

## 🏗️ Architecture

```
stargazer/
├── Dockerfile                    # Multi-stage, minimal
├── requirements.txt              # Essential deps only
├── kustomization.yaml          # Kustomize deployment
├── deployment.yaml             # Namespace auto-detection
├── serviceaccount.yaml          # Read-only permissions
├── role.yaml                  # RBAC configuration
├── rolebinding.yaml           # Service account binding
├── persistentvolumeclaim.yaml    # Issue history storage
├── build.sh                   # Build and deploy script
├── standalone.py              # Development entry point
└── src/
    ├── main.py               # CLI/TUI entry point
    ├── k8s_client.py        # Efficient K8s client with caching
    ├── discovery.py          # Lightweight discovery engine
    ├── mock_ai.py           # Pattern-based mock AI
    ├── tui_app.py          # Textual interface
    ├── agents.py            # Agent system
    ├── storage.py           # JSON persistence
    └── utils.py            # Shared utilities
```

## 🛠️ Technology Stack

- **Base**: Python 3.11-slim
- **CLI**: Click framework
- **TUI**: Textual framework
- **K8s**: Official Python client
- **AI**: Real AI support (OpenAI, Anthropic, Ollama) with pattern-based fallback
- **Storage**: JSON file persistence

## 📊 Resource Usage

- **CPU**: 25m request, 50m limit
- **Memory**: 32Mi request, 64Mi limit
- **Storage**: 100Mi PVC for issue history
- **Startup Time**: <5 seconds
- **Scan Interval**: 2 seconds (configurable)

## 🔒 Security

- **Read-Only RBAC**: Safe for production
- **No Privileges**: Runs as non-root user
- **Minimal Scope**: Only reads cluster state
- **No External Calls**: Fully self-contained

## 📈 Performance Features

- **Efficient Caching**: 30s TTL for API responses
- **Background Scanning**: Minimal overhead
- **Compact Data Models**: Optimized for speed
- **Async Operations**: Non-blocking I/O

## 🐛 Development

### Local Testing

```bash
# Install dependencies
pip install -r requirements.txt

# Run demo mode
python standalone.py --demo

# Interactive testing
python standalone.py --interactive

# Test specific modules
python -c "from src.agents import AgentSystem; print('Agents loaded successfully')"
```

### Building

```bash
# Local build
docker build -t stargazer:test .

# Test container
docker run -it --rm stargazer:test python standalone.py --demo
```

## 📝 Configuration

### Environment Variables

- `POD_NAMESPACE` - Auto-detected via Downward API
- `POD_NAME` - Auto-detected via Downward API
- `SCAN_INTERVAL` - Discovery scan interval (default: 2s)
- `CACHE_TTL` - API cache TTL (default: 30s)
- `LOG_LEVEL` - Logging level (default: INFO)

### AI Configuration (Optional)

To enable real AI troubleshooting, set one of:

- **OpenAI**: `OPENAI_ENABLED=true` and `OPENAI_API_KEY=your-key` (default model: gpt-4o-mini)
- **Anthropic**: `ANTHROPIC_ENABLED=true` and `ANTHROPIC_API_KEY=your-key` (default model: claude-3-haiku-20240307)
- **Ollama**: `OLLAMA_ENABLED=true` and optionally `OLLAMA_URL=http://localhost:11434` (default model: llama3.2)

If no AI is configured, Stargazer will use pattern-based recommendations (MockAI).

### Kubernetes Configuration

The deployment uses:
- Downward API for namespace detection
- Read-only RBAC permissions
- Persistent volume for issue storage
- Resource limits for efficiency

## 🚀 Future Enhancements

- **Additional AI Providers**: Support for more LLM providers (Gemini, Cohere, Mistral, etc.)
- **Metrics Integration**: Prometheus metrics
- **Alerting**: Webhook integrations
- **Multi-Cluster**: Support for multiple clusters
- **Plugin System**: Custom agent plugins
- **GitOps**: Configuration as code

## 📄 License

MIT License - see LICENSE file for details

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## 🔄 Comparison with Stargazer v2

Stargazer and Stargazer v2 are the same application with different deployment models:

### Stargazer (This Version)
- **Deployment**: Kubernetes-native, runs in-cluster
- **Interface**: CLI + TUI (Textual)
- **Focus**: Lightweight, resource-efficient, in-cluster troubleshooting
- **AI**: Simplified AI engine (OpenAI, Anthropic, Ollama) with MockAI fallback
- **Use Case**: Deploy in your cluster for continuous monitoring

### Stargazer v2
- **Deployment**: Standalone application with web UI
- **Interface**: Web UI (Next.js) + CLI
- **Focus**: Feature-rich, multi-cluster support, web-based troubleshooting
- **AI**: Full AI engine (10+ providers) with advanced features
- **Use Case**: Run locally or in-cluster with web interface

### Feature Parity

Both versions now share:
- ✅ Advanced issue discovery (service mesh, network policies, policy violations)
- ✅ Multi-namespace support
- ✅ Real AI integration (with fallback)
- ✅ Kubectl command execution with validation
- ✅ Same agent system

**When to use which:**
- Use **Stargazer** (this version) if you want a lightweight, in-cluster tool with CLI/TUI
- Use **Stargazer v2** if you want a web UI, multi-cluster support, or more AI providers

---

**Stargazer** - Making Kubernetes troubleshooting as easy as looking at the stars ✨
