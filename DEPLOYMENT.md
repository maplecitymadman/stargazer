# Stargazer Deployment Guide

This guide covers all deployment options for Stargazer, including desktop app, CLI, Docker, and Kubernetes in-cluster deployments.

## Table of Contents

1. [Deployment Options Overview](#deployment-options-overview)
2. [Desktop App Deployment](#desktop-app-deployment)
3. [CLI Deployment](#cli-deployment)
4. [Docker Deployment](#docker-deployment)
5. [Kubernetes In-Cluster Deployment](#kubernetes-in-cluster-deployment)
6. [Configuration Management](#configuration-management)
7. [Environment Variables](#environment-variables)
8. [Kubeconfig Setup](#kubeconfig-setup)
9. [Multi-Cluster Setup](#multi-cluster-setup)
10. [Troubleshooting](#troubleshooting)
11. [Upgrading](#upgrading)
12. [Uninstallation](#uninstallation)

---

## Deployment Options Overview

Stargazer can be deployed in three primary modes:

### 1. Desktop Application (Recommended for Local Use)
- **Platform**: macOS, Windows, Linux
- **Use Case**: Local troubleshooting, single-user workstation
- **Pros**: Native GUI, no browser required, easy to use
- **Cons**: Requires installation on each workstation

### 2. CLI Tool
- **Platform**: macOS, Windows, Linux
- **Use Case**: Terminal users, automation, scripts
- **Pros**: Lightweight, scriptable, quick commands
- **Cons**: No visual interface

### 3. Containerized (Docker/Kubernetes)
- **Platform**: Any platform with Docker/Kubernetes
- **Use Case**: Multi-user teams, in-cluster monitoring
- **Pros**: Centralized deployment, no local installation
- **Cons**: Requires cluster deployment, network access

---

## Desktop App Deployment

The desktop app provides a native GUI experience using Wails (Go + React).

### Prerequisites

- Go 1.21 or higher
- Node.js 16 or higher
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Platform-specific requirements:
  - **macOS**: Xcode Command Line Tools
  - **Windows**: WebView2 runtime (usually pre-installed on Windows 10+)
  - **Linux**: `webkit2gtk` package

### macOS Deployment

#### Build from Source

```bash
# Clone repository
git clone https://github.com/maplecitymadman/stargazer.git
cd stargazer

# Build desktop app
make build-gui
# or
wails build

# App location: build/bin/Stargazer.app
```

#### Install to Applications

```bash
# Copy to Applications folder
cp -r build/bin/Stargazer.app /Applications/

# Or create a DMG for distribution
# (Requires create-dmg: brew install create-dmg)
create-dmg \
  --volname "Stargazer Installer" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --app-drop-link 425 120 \
  Stargazer.dmg \
  build/bin/Stargazer.app
```

#### Homebrew Installation

```bash
# Using Homebrew formula
brew install --build-from-source stargazer.rb

# Or from a tap (once published)
brew tap your-org/stargazer
brew install stargazer
```

### Windows Deployment

#### Build from Source

```bash
# Clone repository
git clone https://github.com/maplecitymadman/stargazer.git
cd stargazer

# Build desktop app
wails build -platform windows/amd64

# Executable location: build/bin/Stargazer.exe
```

#### Create Installer (Optional)

Use tools like NSIS or Inno Setup to create a Windows installer:

```nsis
; Example NSIS script (stargazer-installer.nsi)
!include "MUI2.nsh"

Name "Stargazer"
OutFile "Stargazer-Setup.exe"
InstallDir "$PROGRAMFILES\Stargazer"

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Install"
  SetOutPath "$INSTDIR"
  File "build\bin\Stargazer.exe"
  CreateShortcut "$DESKTOP\Stargazer.lnk" "$INSTDIR\Stargazer.exe"
SectionEnd
```

### Linux Deployment

#### Build from Source

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev

# Or for Fedora/RHEL
sudo dnf install gtk3-devel webkit2gtk3-devel

# Clone and build
git clone https://github.com/maplecitymadman/stargazer.git
cd stargazer
make build-gui

# Binary location: build/bin/Stargazer
```

#### Create .deb Package (Debian/Ubuntu)

```bash
# Create package structure
mkdir -p stargazer-deb/DEBIAN
mkdir -p stargazer-deb/usr/local/bin
mkdir -p stargazer-deb/usr/share/applications

# Copy binary
cp build/bin/Stargazer stargazer-deb/usr/local/bin/

# Create control file
cat > stargazer-deb/DEBIAN/control << EOF
Package: stargazer
Version: 0.1.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Your Name <your@email.com>
Description: Kubernetes troubleshooting tool
 Stargazer provides AI-powered Kubernetes diagnostics.
EOF

# Create .desktop file
cat > stargazer-deb/usr/share/applications/stargazer.desktop << EOF
[Desktop Entry]
Name=Stargazer
Comment=Kubernetes Troubleshooting Tool
Exec=/usr/local/bin/Stargazer
Icon=stargazer
Type=Application
Categories=Development;
EOF

# Build package
dpkg-deb --build stargazer-deb
```

#### Create AppImage (Universal Linux)

```bash
# Install appimagetool
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
chmod +x appimagetool-x86_64.AppImage

# Create AppDir structure
mkdir -p Stargazer.AppDir/usr/bin
cp build/bin/Stargazer Stargazer.AppDir/usr/bin/

# Create AppRun script
cat > Stargazer.AppDir/AppRun << 'EOF'
#!/bin/bash
SELF=$(readlink -f "$0")
HERE=${SELF%/*}
exec "${HERE}/usr/bin/Stargazer" "$@"
EOF
chmod +x Stargazer.AppDir/AppRun

# Create .desktop file
cat > Stargazer.AppDir/stargazer.desktop << EOF
[Desktop Entry]
Name=Stargazer
Exec=Stargazer
Icon=stargazer
Type=Application
Categories=Development;
EOF

# Build AppImage
./appimagetool-x86_64.AppImage Stargazer.AppDir
```

### First Launch

1. Launch Stargazer from Applications/Start Menu/Desktop
2. App automatically detects kubeconfig from `~/.kube/config`
3. Configure AI providers in Settings (optional)
4. Start troubleshooting

---

## CLI Deployment

### Build from Source

```bash
# Clone repository
git clone https://github.com/maplecitymadman/stargazer.git
cd stargazer

# Build CLI binary
make build

# Binary location: bin/stargazer
```

### Platform-Specific Builds

```bash
# Build for all platforms
make build-release

# This creates:
# - bin/stargazer-darwin-amd64 (macOS Intel)
# - bin/stargazer-darwin-arm64 (macOS M1/M2)
# - bin/stargazer-linux-amd64 (Linux x64)
# - bin/stargazer-linux-arm64 (Linux ARM)
# - bin/stargazer-windows-amd64.exe (Windows x64)
```

### System Installation

#### macOS/Linux

```bash
# Install to system path
sudo make install
# Installs to /usr/local/bin/stargazer

# Or manually
sudo cp bin/stargazer /usr/local/bin/
sudo chmod +x /usr/local/bin/stargazer

# Verify installation
stargazer --version
```

#### Windows

```powershell
# Copy to a directory in PATH
Copy-Item bin\stargazer-windows-amd64.exe C:\Windows\System32\stargazer.exe

# Or add to user PATH
$env:Path += ";C:\path\to\stargazer"

# Verify installation
stargazer --version
```

### Initial Configuration

```bash
# Run setup wizard
stargazer config setup

# Verify configuration
stargazer config show

# Test cluster connection
stargazer health
```

---

## Docker Deployment

Docker deployment is ideal for running Stargazer as a web service accessible to multiple users.

### Using Existing Dockerfile

The project includes a Dockerfile for containerized deployment:

```dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY bin/stargazer-linux-amd64 /usr/local/bin/stargazer
RUN chmod 755 /usr/local/bin/stargazer
COPY frontend/out /app/frontend/out
EXPOSE 8000
WORKDIR /app
CMD ["stargazer"]
```

### Build Docker Image

```bash
# Build CLI binary for Linux
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/stargazer-linux-amd64 cmd/stargazer/main.go

# Build frontend
cd frontend
npm install
npm run build
cd ..

# Build Docker image
docker build -t stargazer:latest .

# Or with version tag
docker build -t stargazer:0.1.0 .
```

### Run Docker Container

#### Basic Usage (Local Kubeconfig)

```bash
# Mount local kubeconfig
docker run -d \
  --name stargazer \
  -p 8000:8000 \
  -v ~/.kube/config:/root/.kube/config:ro \
  -v ~/.stargazer:/root/.stargazer \
  stargazer:latest web --port 8000
```

#### With Environment Variables

```bash
docker run -d \
  --name stargazer \
  -p 8000:8000 \
  -e KUBECONFIG=/root/.kube/config \
  -e KUBECTL_CONTEXT=production \
  -e LOG_LEVEL=INFO \
  -v ~/.kube/config:/root/.kube/config:ro \
  -v ~/.stargazer:/root/.stargazer \
  stargazer:latest web
```

#### With Service Account (In-Cluster Mode)

```bash
# Run without kubeconfig (uses in-cluster config)
docker run -d \
  --name stargazer \
  -p 8000:8000 \
  -e KUBERNETES_SERVICE_HOST=kubernetes.default.svc \
  -e KUBERNETES_SERVICE_PORT=443 \
  -v /var/run/secrets/kubernetes.io/serviceaccount:/var/run/secrets/kubernetes.io/serviceaccount:ro \
  stargazer:latest web
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  stargazer:
    image: stargazer:latest
    container_name: stargazer
    ports:
      - "8000:8000"
    volumes:
      - ~/.kube/config:/root/.kube/config:ro
      - stargazer-data:/root/.stargazer
    environment:
      - KUBECONFIG=/root/.kube/config
      - KUBECTL_CONTEXT=production
      - LOG_LEVEL=INFO
      - CACHE_TTL=30
    restart: unless-stopped
    command: ["web", "--port", "8000"]

volumes:
  stargazer-data:
```

Run with Docker Compose:

```bash
# Start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Push to Container Registry

```bash
# Tag image
docker tag stargazer:latest your-registry.io/stargazer:latest
docker tag stargazer:latest your-registry.io/stargazer:0.1.0

# Push to registry
docker push your-registry.io/stargazer:latest
docker push your-registry.io/stargazer:0.1.0
```

---

## Kubernetes In-Cluster Deployment

Deploy Stargazer as a pod within your Kubernetes cluster for centralized troubleshooting.

### Prerequisites

- Kubernetes cluster with RBAC enabled
- kubectl configured with cluster access
- Container image pushed to accessible registry

### ServiceAccount and RBAC

Create `stargazer-rbac.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: stargazer

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: stargazer
  namespace: stargazer

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: stargazer-reader
rules:
  # Read-only access to cluster resources
  - apiGroups: [""]
    resources:
      - nodes
      - namespaces
      - pods
      - pods/log
      - services
      - endpoints
      - events
      - configmaps
      - persistentvolumeclaims
      - persistentvolumes
    verbs: ["get", "list", "watch"]

  - apiGroups: ["apps"]
    resources:
      - deployments
      - replicasets
      - statefulsets
      - daemonsets
    verbs: ["get", "list", "watch"]

  - apiGroups: ["batch"]
    resources:
      - jobs
      - cronjobs
    verbs: ["get", "list", "watch"]

  - apiGroups: ["networking.k8s.io"]
    resources:
      - ingresses
      - networkpolicies
    verbs: ["get", "list", "watch"]

  - apiGroups: ["policy"]
    resources:
      - poddisruptionbudgets
    verbs: ["get", "list", "watch"]

  - apiGroups: ["metrics.k8s.io"]
    resources:
      - pods
      - nodes
    verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: stargazer-reader-binding
subjects:
  - kind: ServiceAccount
    name: stargazer
    namespace: stargazer
roleRef:
  kind: ClusterRole
  name: stargazer-reader
  apiGroup: rbac.authorization.k8s.io
```

Apply RBAC:

```bash
kubectl apply -f stargazer-rbac.yaml
```

### Deployment Manifest

Create `stargazer-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stargazer
  namespace: stargazer
  labels:
    app: stargazer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: stargazer
  template:
    metadata:
      labels:
        app: stargazer
    spec:
      serviceAccountName: stargazer
      containers:
      - name: stargazer
        image: your-registry.io/stargazer:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
          name: http
          protocol: TCP
        env:
        - name: LOG_LEVEL
          value: "INFO"
        - name: CACHE_TTL
          value: "30"
        # In-cluster config is automatically detected
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 10
        volumeMounts:
        - name: storage
          mountPath: /root/.stargazer
      volumes:
      - name: storage
        persistentVolumeClaim:
          claimName: stargazer-storage

---
apiVersion: v1
kind: Service
metadata:
  name: stargazer
  namespace: stargazer
spec:
  type: ClusterIP
  selector:
    app: stargazer
  ports:
  - port: 8000
    targetPort: 8000
    protocol: TCP
    name: http

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: stargazer-storage
  namespace: stargazer
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

Deploy to Kubernetes:

```bash
kubectl apply -f stargazer-deployment.yaml
```

### Ingress (Optional)

Create `stargazer-ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: stargazer
  namespace: stargazer
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    # Add TLS/authentication annotations as needed
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - stargazer.yourdomain.com
    secretName: stargazer-tls
  rules:
  - host: stargazer.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: stargazer
            port:
              number: 8000
```

Apply Ingress:

```bash
kubectl apply -f stargazer-ingress.yaml
```

### ConfigMap for Configuration

Create `stargazer-config.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: stargazer-config
  namespace: stargazer
data:
  config.yaml: |
    version: "1.0"
    api:
      port: 8000
      enable_cors: true
      rate_limit_rps: 100
    storage:
      path: /root/.stargazer/storage
      retain_days: 30
      max_scan_results: 100
    llm:
      default_provider: ""
      providers:
        openai:
          enabled: false
          model: gpt-4o-mini
        anthropic:
          enabled: false
          model: claude-3-haiku-20240307
```

Mount ConfigMap in Deployment:

```yaml
# Add to spec.template.spec.volumes
- name: config
  configMap:
    name: stargazer-config

# Add to spec.template.spec.containers[0].volumeMounts
- name: config
  mountPath: /root/.stargazer/config.yaml
  subPath: config.yaml
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -n stargazer

# Check logs
kubectl logs -n stargazer -l app=stargazer

# Port-forward for local access
kubectl port-forward -n stargazer svc/stargazer 8000:8000

# Access at http://localhost:8000
```

---

## Configuration Management

### Configuration File Location

Stargazer stores configuration in `~/.stargazer/config.yaml`:

```yaml
version: "1.0"
created_at: "2026-01-13 12:00:00"
updated_at: "2026-01-13 12:00:00"

kubeconfig:
  path: ""  # Auto-detected if empty
  context: ""  # Uses current context if empty

api:
  port: 8000
  enable_cors: true
  rate_limit_rps: 100

storage:
  path: ~/.stargazer/storage
  retain_days: 30
  max_scan_results: 100

llm:
  default_provider: "openai"
  providers:
    openai:
      enabled: true
      api_key: "sk-..."
      model: "gpt-4o-mini"
      base_url: "https://api.openai.com/v1"
    anthropic:
      enabled: false
      api_key: ""
      model: "claude-3-haiku-20240307"
    gemini:
      enabled: false
      api_key: ""
      model: "gemini-pro"
    ollama:
      enabled: false
      model: "llama3.1"
      base_url: "http://localhost:11434"
```

### Configuration Commands

```bash
# Interactive setup wizard
stargazer config setup

# Show current configuration
stargazer config show

# Edit configuration manually
vim ~/.stargazer/config.yaml

# Validate configuration
stargazer config show  # Automatically validates
```

### Storage Management

```bash
# Storage location
~/.stargazer/
├── config.yaml          # Configuration file
├── storage/
│   ├── scans/          # Scan results
│   └── issues/         # Discovered issues

# Clear old scan results
rm -rf ~/.stargazer/storage/scans/*

# Backup configuration
cp ~/.stargazer/config.yaml ~/.stargazer/config.yaml.backup
```

---

## Environment Variables

Stargazer supports the following environment variables:

### Kubernetes Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to kubeconfig file | `~/.kube/config` |
| `KUBECTL_CONTEXT` | Kubernetes context to use | Current context |
| `KUBERNETES_SERVICE_HOST` | K8s API server host (in-cluster) | Auto-detected |
| `KUBERNETES_SERVICE_PORT` | K8s API server port (in-cluster) | Auto-detected |

### Application Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | `INFO` |
| `CACHE_TTL` | API cache TTL in seconds | `30` |
| `PORT` | API server port | `8000` |
| `ENABLE_CORS` | Enable CORS for API | `true` |
| `RATE_LIMIT_RPS` | Rate limit (requests per second) | `100` |

### LLM Provider Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | - |
| `ANTHROPIC_API_KEY` | Anthropic API key | - |
| `GEMINI_API_KEY` | Google Gemini API key | - |
| `OLLAMA_BASE_URL` | Ollama API base URL | `http://localhost:11434` |

### Storage Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `STORAGE_PATH` | Storage directory path | `~/.stargazer/storage` |
| `RETAIN_DAYS` | Days to retain scan results | `30` |
| `MAX_SCAN_RESULTS` | Maximum scan results to keep | `100` |

### Example: Environment Variables in Docker

```bash
docker run -d \
  --name stargazer \
  -p 8000:8000 \
  -e KUBECONFIG=/root/.kube/config \
  -e KUBECTL_CONTEXT=production \
  -e LOG_LEVEL=DEBUG \
  -e CACHE_TTL=60 \
  -e OPENAI_API_KEY=sk-... \
  -v ~/.kube/config:/root/.kube/config:ro \
  stargazer:latest web
```

### Example: Environment Variables in Kubernetes

```yaml
env:
- name: LOG_LEVEL
  value: "INFO"
- name: CACHE_TTL
  value: "30"
- name: OPENAI_API_KEY
  valueFrom:
    secretKeyRef:
      name: stargazer-secrets
      key: openai-api-key
```

---

## Kubeconfig Setup

### Local Kubeconfig (Default)

Stargazer automatically detects kubeconfig from standard locations:

1. `KUBECONFIG` environment variable
2. `~/.kube/config` (default)
3. Explicit path in config file

```bash
# Use default kubeconfig
stargazer health

# Use specific kubeconfig
stargazer --kubeconfig=/path/to/kubeconfig health

# Use specific context
stargazer --context=production health
```

### In-Cluster Configuration

When running as a pod in Kubernetes, Stargazer automatically uses in-cluster service account credentials:

```go
// Automatic detection in code
// If /var/run/secrets/kubernetes.io/serviceaccount exists,
// use in-cluster config
```

No kubeconfig needed when deployed in-cluster with proper ServiceAccount.

### Multiple Kubeconfig Files

Merge multiple kubeconfig files:

```bash
# Set KUBECONFIG with multiple files
export KUBECONFIG=~/.kube/config:~/.kube/cluster2:~/.kube/cluster3

# Or merge into single file
KUBECONFIG=~/.kube/config:~/.kube/cluster2 kubectl config view --flatten > ~/.kube/merged-config
export KUBECONFIG=~/.kube/merged-config
```

### Kubeconfig Security

Protect kubeconfig files:

```bash
# Secure permissions
chmod 600 ~/.kube/config

# Read-only mount in Docker
docker run -v ~/.kube/config:/root/.kube/config:ro stargazer:latest

# Use service accounts instead of admin credentials
kubectl create serviceaccount stargazer-readonly
kubectl create clusterrolebinding stargazer-readonly --clusterrole=view --serviceaccount=default:stargazer-readonly
```

---

## Multi-Cluster Setup

### Using Multiple Contexts

List available contexts:

```bash
kubectl config get-contexts
```

Switch contexts in Stargazer:

#### Desktop App
1. Open Settings
2. Navigate to Kubernetes section
3. Select context from dropdown
4. Click "Switch Context"

#### CLI
```bash
# Use specific context
stargazer --context=production health

# Or set in config file
stargazer config setup
# Select context: production

# Or set environment variable
export KUBECTL_CONTEXT=production
stargazer health
```

### Unified Multi-Cluster Dashboard

Deploy Stargazer in each cluster and use a reverse proxy:

Create `nginx-multi-cluster.conf`:

```nginx
upstream cluster1 {
    server stargazer-cluster1.example.com:8000;
}

upstream cluster2 {
    server stargazer-cluster2.example.com:8000;
}

upstream cluster3 {
    server stargazer-cluster3.example.com:8000;
}

server {
    listen 80;
    server_name stargazer.example.com;

    location /cluster1/ {
        proxy_pass http://cluster1/;
    }

    location /cluster2/ {
        proxy_pass http://cluster2/;
    }

    location /cluster3/ {
        proxy_pass http://cluster3/;
    }
}
```

### Hub-and-Spoke Architecture

Deploy one central Stargazer instance with access to multiple clusters:

```yaml
# centralized-deployment.yaml
apiVersion: v1
kind: Secret
metadata:
  name: multi-cluster-kubeconfig
  namespace: stargazer
type: Opaque
stringData:
  config: |
    # Merged kubeconfig with multiple clusters
    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        server: https://cluster1.example.com
      name: cluster1
    - cluster:
        server: https://cluster2.example.com
      name: cluster2
    contexts:
    - context:
        cluster: cluster1
        user: cluster1-user
      name: cluster1
    - context:
        cluster: cluster2
        user: cluster2-user
      name: cluster2
    current-context: cluster1
    users:
    - name: cluster1-user
      user:
        token: <token1>
    - name: cluster2-user
      user:
        token: <token2>

---
# Mount in deployment
spec:
  template:
    spec:
      containers:
      - name: stargazer
        volumeMounts:
        - name: kubeconfig
          mountPath: /root/.kube
      volumes:
      - name: kubeconfig
        secret:
          secretName: multi-cluster-kubeconfig
```

---

## Troubleshooting

### Common Issues

#### 1. Cannot Connect to Cluster

**Symptoms:**
```
❌ Cannot connect to cluster: connection refused
```

**Solutions:**
```bash
# Verify kubeconfig
kubectl cluster-info

# Check context
kubectl config current-context

# Test cluster access
kubectl get nodes

# Use explicit kubeconfig
stargazer --kubeconfig=/path/to/config health

# Check network connectivity
curl https://your-k8s-api:6443
```

#### 2. Permission Denied Errors

**Symptoms:**
```
❌ Failed to list pods: pods is forbidden
```

**Solutions:**
```bash
# Check RBAC permissions
kubectl auth can-i get pods --as=system:serviceaccount:stargazer:stargazer

# Grant read permissions
kubectl create clusterrolebinding stargazer-reader \
  --clusterrole=view \
  --serviceaccount=stargazer:stargazer

# Or use custom role (see RBAC section)
kubectl apply -f stargazer-rbac.yaml
```

#### 3. Desktop App Won't Start

**Symptoms:**
- App crashes on launch
- Blank window
- "WebView2 not found" (Windows)

**Solutions:**

macOS:
```bash
# Check permissions
xattr -d com.apple.quarantine /Applications/Stargazer.app

# Reinstall
rm -rf /Applications/Stargazer.app
cp -r build/bin/Stargazer.app /Applications/
```

Windows:
```powershell
# Install WebView2 Runtime
winget install Microsoft.EdgeWebView2Runtime

# Or download from:
# https://developer.microsoft.com/microsoft-edge/webview2/
```

Linux:
```bash
# Install webkit2gtk
sudo apt-get install webkit2gtk-4.0  # Ubuntu/Debian
sudo dnf install webkit2gtk3        # Fedora/RHEL
```

#### 4. API Server Port Already in Use

**Symptoms:**
```
❌ Port 8000 in use, trying 8001...
❌ Could not find available port (tried 8000-8010)
```

**Solutions:**
```bash
# Find process using port
lsof -i :8000
netstat -an | grep 8000

# Kill process
kill <PID>

# Or use different port
stargazer web --port 8080

# In Docker
docker run -p 8080:8000 stargazer:latest web --port 8000
```

#### 5. Frontend Not Loading

**Symptoms:**
- Blank page
- 404 errors
- Assets not found

**Solutions:**
```bash
# Rebuild frontend
cd frontend
npm install
npm run build
cd ..

# Rebuild desktop app
make build-gui

# For Docker, ensure frontend is copied
docker build --no-cache -t stargazer:latest .
```

#### 6. High Memory Usage

**Symptoms:**
- App becomes slow
- OOM errors in Kubernetes

**Solutions:**

Desktop/CLI:
```bash
# Reduce cache TTL
export CACHE_TTL=10
stargazer web

# Limit scan results
# Edit ~/.stargazer/config.yaml
storage:
  max_scan_results: 50
```

Kubernetes:
```yaml
# Increase resource limits
resources:
  requests:
    memory: "256Mi"
  limits:
    memory: "1Gi"
```

#### 7. AI Features Not Working

**Symptoms:**
- AI troubleshooting returns errors
- "Provider not configured"

**Solutions:**
```bash
# Configure provider
stargazer config setup
# Select provider and enter API key

# Or set environment variable
export OPENAI_API_KEY=sk-...

# Verify configuration
stargazer config show
```

### Debug Mode

Enable debug logging:

```bash
# CLI
export LOG_LEVEL=DEBUG
stargazer health

# Desktop app - check logs
# macOS: ~/Library/Logs/Stargazer/
# Windows: %APPDATA%\Stargazer\logs\
# Linux: ~/.local/share/Stargazer/logs/

# Docker
docker logs -f stargazer

# Kubernetes
kubectl logs -n stargazer -l app=stargazer -f
```

### Collect Diagnostic Information

```bash
# System info
stargazer --version
kubectl version
kubectl cluster-info

# Configuration
stargazer config show

# Test connectivity
stargazer health

# Network diagnostics
kubectl get nodes
kubectl get pods --all-namespaces
```

---

## Upgrading

### Desktop App

#### macOS
```bash
# Using Homebrew
brew upgrade stargazer

# Manual upgrade
cd stargazer
git pull
make build-gui
cp -r build/bin/Stargazer.app /Applications/
```

#### Windows
```bash
# Download new version
# Replace existing .exe file
# Or run new installer

# Using winget (if available)
winget upgrade stargazer
```

#### Linux
```bash
# From source
cd stargazer
git pull
make build-gui
sudo cp build/bin/Stargazer /usr/local/bin/

# Using package manager
sudo apt-get update && sudo apt-get upgrade stargazer  # Debian/Ubuntu
sudo dnf upgrade stargazer  # Fedora/RHEL
```

### CLI

```bash
# Rebuild from source
cd stargazer
git pull
make build
sudo make install

# Or download new binary
wget https://github.com/maplecitymadman/stargazer/releases/download/v0.2.0/stargazer-linux-amd64
chmod +x stargazer-linux-amd64
sudo mv stargazer-linux-amd64 /usr/local/bin/stargazer
```

### Docker

```bash
# Pull new image
docker pull your-registry.io/stargazer:latest

# Stop old container
docker stop stargazer
docker rm stargazer

# Start new container
docker run -d \
  --name stargazer \
  -p 8000:8000 \
  -v ~/.kube/config:/root/.kube/config:ro \
  -v ~/.stargazer:/root/.stargazer \
  your-registry.io/stargazer:latest web

# Or with docker-compose
docker-compose pull
docker-compose up -d
```

### Kubernetes

```bash
# Update image tag
kubectl set image deployment/stargazer \
  stargazer=your-registry.io/stargazer:v0.2.0 \
  -n stargazer

# Or apply updated manifest
kubectl apply -f stargazer-deployment.yaml

# Watch rollout
kubectl rollout status deployment/stargazer -n stargazer

# Rollback if needed
kubectl rollout undo deployment/stargazer -n stargazer
```

### Configuration Migration

After upgrading, check for configuration changes:

```bash
# Backup current config
cp ~/.stargazer/config.yaml ~/.stargazer/config.yaml.backup

# Check for new options
stargazer config show

# Re-run setup if needed
stargazer config setup
```

---

## Uninstallation

### Desktop App

#### macOS
```bash
# Remove app
rm -rf /Applications/Stargazer.app

# Remove configuration
rm -rf ~/.stargazer

# Using Homebrew
brew uninstall stargazer
```

#### Windows
```powershell
# Remove executable
Remove-Item C:\Windows\System32\stargazer.exe
# Or use Control Panel > Uninstall Programs

# Remove configuration
Remove-Item -Recurse $env:USERPROFILE\.stargazer
```

#### Linux
```bash
# Remove binary
sudo rm /usr/local/bin/Stargazer

# Remove configuration
rm -rf ~/.stargazer

# Using package manager
sudo apt-get remove stargazer  # Debian/Ubuntu
sudo dnf remove stargazer      # Fedora/RHEL
```

### CLI

```bash
# Remove binary
sudo rm /usr/local/bin/stargazer

# Remove configuration
rm -rf ~/.stargazer

# Using Homebrew
brew uninstall stargazer
```

### Docker

```bash
# Stop and remove container
docker stop stargazer
docker rm stargazer

# Remove image
docker rmi stargazer:latest

# Remove volumes (caution: deletes data)
docker volume rm stargazer-data

# Using docker-compose
docker-compose down -v
```

### Kubernetes

```bash
# Delete all resources
kubectl delete namespace stargazer

# Or delete individually
kubectl delete -f stargazer-deployment.yaml
kubectl delete -f stargazer-rbac.yaml
kubectl delete -f stargazer-ingress.yaml

# Verify deletion
kubectl get all -n stargazer
```

### Clean Up Storage

```bash
# Remove all Stargazer data
rm -rf ~/.stargazer

# Or selectively remove
rm -rf ~/.stargazer/storage  # Scan results only
rm ~/.stargazer/config.yaml   # Configuration only
```

---

## Additional Resources

- [README.md](./README.md) - General overview and quick start
- [GitHub Repository](https://github.com/maplecitymadman/stargazer)
- [Issue Tracker](https://github.com/maplecitymadman/stargazer/issues)

---

## Support

For deployment issues or questions:

1. Check [Troubleshooting](#troubleshooting) section
2. Search [GitHub Issues](https://github.com/maplecitymadman/stargazer/issues)
3. Open a new issue with:
   - Deployment method (desktop/CLI/Docker/K8s)
   - Platform (macOS/Windows/Linux)
   - Error messages
   - `stargazer --version` output
   - `kubectl version` output

---

**Last Updated:** 2026-01-20
**Version:** 0.1.0
