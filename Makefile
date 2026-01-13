.PHONY: build clean test run install dev fmt lint help

# Binary name
BINARY_NAME=stargazer
VERSION?=0.1.0-dev
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/stargazer/main.go
	@echo "✅ Built: $(BUILD_DIR)/$(BINARY_NAME)"

build-release: ## Build release binaries for all platforms
	@echo "🏗️  Building release binaries..."
	@mkdir -p $(BUILD_DIR)
	# macOS Intel
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/stargazer/main.go
	# macOS ARM (M1/M2)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/stargazer/main.go
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/stargazer/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 cmd/stargazer/main.go
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/stargazer/main.go
	@echo "✅ Release binaries built in $(BUILD_DIR)/"

install: build ## Install the binary to /usr/local/bin
	@echo "📦 Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed: /usr/local/bin/$(BINARY_NAME)"

clean: ## Remove build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf dist/
	@echo "✅ Clean complete"

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

run: ## Run the application (web mode)
	@echo "🚀 Running $(BINARY_NAME)..."
	go run cmd/stargazer/main.go web

dev: ## Run in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

fmt: ## Format code
	@echo "🎨 Formatting code..."
	gofmt -w $(GO_FILES)
	@echo "✅ Format complete"

lint: ## Run linter (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "❌ golangci-lint not installed. Run: brew install golangci-lint" && exit 1)
	@echo "🔍 Running linter..."
	golangci-lint run ./...

deps: ## Download dependencies
	@echo "📥 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies updated"

build-gui: ## Build desktop app with Wails (requires wails CLI)
	@echo "🖥️  Building desktop app..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin; \
	if ! command -v wails > /dev/null 2>&1; then \
		echo "❌ Wails not installed. Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		echo "💡 Also ensure $$(go env GOPATH)/bin is in your PATH"; \
		exit 1; \
	fi
	@echo "📦 Building frontend..."
	@cd frontend && npm install --silent && npm run build
	@echo "🔨 Building desktop app..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin; wails build
	@echo "✅ Desktop app built in build/bin/"

build-all: build build-gui ## Build both CLI and desktop app

dev-gui: ## Run desktop app in dev mode
	@export PATH=$$PATH:$$(go env GOPATH)/bin; \
	if ! command -v wails > /dev/null 2>&1; then \
		echo "❌ Wails not installed. Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest"; \
		echo "💡 Also ensure $$(go env GOPATH)/bin is in your PATH"; \
		exit 1; \
	fi
	@export PATH=$$PATH:$$(go env GOPATH)/bin; wails dev

# Quick aliases
b: build ## Alias for build
r: run ## Alias for run
t: test ## Alias for test
c: clean ## Alias for clean
g: build-gui ## Alias for build-gui
