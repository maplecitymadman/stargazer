# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-01-19

### Added

#### Core Infrastructure
- Desktop application built with Wails v2 for native cross-platform experience
- CLI tool built with Cobra for command-line operations
- Kubernetes client integration using official client-go library
- Multi-cluster support with context switching
- Intelligent kubeconfig discovery across multiple standard locations
- Configuration management with YAML-based storage in `~/.stargazer/`
- RESTful API server with Gin framework
- WebSocket support for real-time updates and notifications

#### Kubernetes Features
- Cluster health monitoring and dashboard
- Issue discovery and automatic scanning
- Resource viewing for pods, deployments, services, and events
- Namespace filtering and cluster-wide operations
- Pod log retrieval with configurable line limits
- Real-time event streaming via WebSocket
- Read-only permissions for production safety

#### Networking & Topology
- Service topology visualization with React Flow
- Interactive, draggable topology graph with custom node styling
- Network-focused metrics (Network Policies count and breakdown)
- Service policy coverage tracking
- Ingress detection (Istio Gateways, Kubernetes Ingress)
- Egress detection (Istio EgressGateways, ServiceEntry, direct egress)
- Connection path tracing from ingress to egress
- Policy evaluation for ingress/egress connections
- Real-time policy change notifications via PolicyWatcher
- Service mesh indicators and policy blocking visualization
- Gateway node visualization with connection status (allowed/blocked)
- Circular layout algorithm for service topology
- Zoom, pan, minimap, and interactive controls for topology view
- Node selection with detailed information panel

#### Network Policies & Security
- Network policy viewing and management
- Comprehensive networking recommendations engine
- Best practices compliance scoring (security, performance, observability, resilience)
- Policy-based connection testing
- Istio AuthorizationPolicy detection with correct API version
- Template-based policy generation

#### User Interface
- Redesigned dashboard with compact widget-based layout
- Expandable sidebar navigation with grouped sections
- Traffic Analysis page with topology, path-trace, ingress, and egress tabs
- Network Policies page with view, build, and test tabs
- Troubleshooting page with blocked connections, services, and recommendations
- Clickable dashboard widgets for navigation to detailed views
- Theme support (Dark, Light, Auto modes)
- Settings UI for kubeconfig configuration with status display
- Context selector displaying cluster name instead of context name
- Responsive design with Tailwind CSS
- Toast notifications for user feedback

#### Developer Experience
- Makefile for build automation
- Configuration wizard for interactive setup
- API caching with configurable 30s TTL
- Rate limiting middleware for API protection
- CORS middleware for cross-origin requests
- Security headers middleware
- Comprehensive error handling and logging
- Request metrics and monitoring

### Changed
- Replaced Pods card with Network Policies metrics on dashboard
- Replaced Deployments card with Services metrics and policy coverage percentage
- Changed cluster selector to display cluster name instead of context name
- Improved compliance score API response structure
- Enhanced dashboard for network troubleshooting with intelligent resource selection
- Optimized React Flow styling to match application theme
- Animated service mesh connections for better visualization

### Fixed
- Unused variables in handlers and improved gateway node handles
- Invalid icon names causing rendering issues
- Tailwind CSS syntax errors (text-[#e4e4e7]-dim to text-space-text-dim)
- React useEffect dependency arrays to include all read variables
- Declaration order by moving loadData before useEffect hooks
- Nil pointer dereferences in ingress/egress detection code
- Resource leaks in PolicyWatcher shutdown
- Context cancellation handling in watchers
- Button click issues and improved UI responsiveness
- Istio AuthorizationPolicy API version detection

### Removed
- Unused frontend components (DeploymentsDetail, PodsDetail, IssuesList, IssueTroubleshooter, ResourcesView)
- Desktop app build artifacts and configuration files from repository
- Redundant build-and-push.sh script
- Python artifacts from initial project setup
- Console.log statements and TODO comments
- Dead code and unused imports

### Security
- Read-only cluster access by default
- Local-only data storage with no external calls (except configured AI providers)
- Standard Kubernetes authentication via kubeconfig
- Rate limiting to prevent API abuse
- Security headers middleware for HTTP responses
- Safe production environment operation

## [0.0.1] - 2026-01-12

### Added
- Initial project structure
- Basic Kubernetes client setup
- Foundation for desktop and CLI applications

---

## Version History

### [0.1.0] - 2026-01-19
First production-ready release with comprehensive networking troubleshooting features, interactive topology visualization, and professional desktop application.

### [0.0.1] - 2026-01-12
Initial commit and project setup.

---

## Links
- [Repository](https://github.com/maplecitymadman/stargazer)
- [Documentation](./README.md)
- [Desktop App Guide](./DESKTOP_APP.md)
