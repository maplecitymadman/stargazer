# Stargazer Desktop App

Stargazer can be built as both a **CLI tool** and a **native desktop application** (like Docker Desktop or Slack).

## Architecture

- **CLI Binary**: `cmd/stargazer/main.go` - Standalone CLI tool
- **Desktop App**: Built with Wails using the React frontend
- **Shared Backend**: Both use the same Go backend (`internal/`)

## Building

### CLI Only (Default)

```bash
# Build CLI binary
make build

# Use CLI commands
./bin/stargazer scan
./bin/stargazer health
./bin/stargazer logs my-pod
```

### Desktop App (Wails)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build desktop app
wails build

# Or run in dev mode
wails dev
```

The desktop app will be built in `build/bin/`:
- macOS: `Stargazer.app`
- Windows: `Stargazer.exe`
- Linux: `Stargazer`

## Distribution

### macOS

```bash
# Build .app bundle
wails build -platform darwin/amd64
wails build -platform darwin/arm64

# Create DMG (optional)
# Use create-dmg or similar tool
```

### Windows

```bash
# Build .exe
wails build -platform windows/amd64

# Creates installer via NSIS (configured in wails.json)
```

### Linux

```bash
# Build binary
wails build -platform linux/amd64

# Create AppImage or .deb package (optional)
```

## Installation

### macOS

1. Download `Stargazer.app`
2. Drag to Applications folder
3. Open from Applications (may need to allow in Security settings)

### Windows

1. Download `Stargazer-installer.exe`
2. Run installer
3. Launch from Start Menu

### Linux

1. Download `Stargazer` binary
2. Make executable: `chmod +x Stargazer`
3. Move to PATH: `sudo mv Stargazer /usr/local/bin/`

## Usage

### Desktop App

- Launch the app from Applications/Start Menu
- GUI opens automatically
- Connects to Kubernetes via kubeconfig
- All features available through the UI

### CLI (Still Available)

Even with the desktop app installed, CLI commands still work:

```bash
# CLI commands work independently
stargazer scan
stargazer health
stargazer logs my-pod
```

## Development

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### Backend Development

```bash
# Run CLI in dev mode
go run cmd/stargazer/main.go scan

# Or run Wails dev mode (includes hot reload)
wails dev
```

## Configuration

Both CLI and GUI use the same config:
- Location: `~/.stargazer/config.yaml`
- Kubeconfig: `~/.kube/config` or `KUBECONFIG` env var

## Features

### Desktop App
- ✅ Native window (not a browser)
- ✅ System tray integration (optional)
- ✅ Native menus and dialogs
- ✅ Auto-updates (future)
- ✅ File associations (future)

### CLI
- ✅ All commands available
- ✅ Scriptable
- ✅ CI/CD integration
- ✅ Remote execution

## Build Scripts

See `Makefile` for build targets:

```bash
make build          # CLI only
make build-gui       # Desktop app (requires Wails)
make build-all       # Both CLI and GUI
```
