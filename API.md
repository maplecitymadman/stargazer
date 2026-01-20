# Stargazer API Documentation

Version: 0.1.0-dev

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [REST API Endpoints](#rest-api-endpoints)
  - [Health & Metrics](#health--metrics)
  - [Context Management](#context-management)
  - [Namespace Operations](#namespace-operations)
  - [Cluster Resources](#cluster-resources)
  - [Events](#events)
  - [Topology & Network Tracing](#topology--network-tracing)
  - [Recommendations & Compliance](#recommendations--compliance)
  - [Policy Management](#policy-management)
  - [Configuration](#configuration)
- [WebSocket API](#websocket-api)
- [Error Responses](#error-responses)
- [Data Models](#data-models)

---

## Overview

Stargazer provides a RESTful API for Kubernetes cluster monitoring, network topology visualization, and policy management with support for Cilium, Istio, and Kyverno.

**Base URL**: `http://localhost:8080/api`

**Content Type**: `application/json`

---

## Authentication

Currently, Stargazer operates as a **local-only** service with no authentication required. It connects to your Kubernetes cluster using your local kubeconfig file.

**Security Note**: Do not expose this API to the public internet without implementing proper authentication and authorization.

---

## Rate Limiting

Rate limits are applied per endpoint to protect cluster resources:

- **Standard endpoints**: No explicit rate limiting (reasonable use expected)
- **Topology endpoints**: 5 requests/minute (expensive operations)
- **Recommendations endpoints**: 10 requests/minute (expensive operations)
- **Health & Metrics endpoints**: No rate limiting

Rate limit responses return `429 Too Many Requests` when exceeded.

---

## REST API Endpoints

### Health & Metrics

#### GET /api/health

Check API server health and Kubernetes cluster connectivity.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "status": "healthy",
  "cluster": "connected",
  "version": "0.1.0-dev",
  "error": ""
}
```

**Response (Unhealthy)**: `503 Service Unavailable`
```json
{
  "status": "unhealthy",
  "cluster": "disconnected",
  "version": "0.1.0-dev",
  "error": "Kubernetes client not initialized. Please configure kubeconfig."
}
```

**Example**:
```bash
curl http://localhost:8080/api/health
```

---

#### GET /api/metrics

Get Prometheus-compatible metrics for monitoring.

**Query Parameters**: None

**Response**: `200 OK` (text/plain)
```
# HELP stargazer_requests_total Total number of HTTP requests
# TYPE stargazer_requests_total counter
stargazer_requests_total 1234

# HELP stargazer_errors_total Total number of HTTP errors (4xx, 5xx)
# TYPE stargazer_errors_total counter
stargazer_errors_total 5

# HELP stargazer_request_duration_seconds Average request duration in seconds
# TYPE stargazer_request_duration_seconds gauge
stargazer_request_duration_seconds 0.025000

# HELP stargazer_uptime_seconds Server uptime in seconds
# TYPE stargazer_uptime_seconds gauge
stargazer_uptime_seconds 3600.00

# HELP stargazer_connected Whether stargazer is connected to Kubernetes cluster
# TYPE stargazer_connected gauge
stargazer_connected 1
```

**Example**:
```bash
curl http://localhost:8080/api/metrics
```

---

### Context Management

#### GET /api/contexts

List all available Kubernetes contexts.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "contexts": [
    {
      "name": "minikube",
      "cluster": "minikube",
      "user": "",
      "namespace": "default",
      "server": "",
      "cloud_provider": "unknown",
      "is_current": true
    }
  ],
  "current_context": "minikube",
  "total": 1
}
```

**Example**:
```bash
curl http://localhost:8080/api/contexts
```

---

#### GET /api/context/current

Get the current Kubernetes context.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "context": "minikube",
  "info": {
    "name": "minikube",
    "cluster": "minikube",
    "namespace": "default",
    "cloud_provider": "unknown"
  }
}
```

**Example**:
```bash
curl http://localhost:8080/api/context/current
```

---

#### POST /api/context/switch

Switch to a different Kubernetes context.

**Request Body**:
```json
{
  "context": "production-cluster"
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "context": "production-cluster",
  "message": "Switched to context: production-cluster (reload page to apply)"
}
```

**Error Response**: `400 Bad Request`
```json
{
  "error": "Context name required"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/context/switch \
  -H "Content-Type: application/json" \
  -d '{"context": "production-cluster"}'
```

---

### Namespace Operations

#### GET /api/namespace

Get the current default namespace.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "namespace": "default"
}
```

**Example**:
```bash
curl http://localhost:8080/api/namespace
```

---

#### GET /api/namespaces

List all namespaces in the cluster.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "namespaces": [
    {
      "name": "default",
      "status": "Active",
      "age": "30d"
    },
    {
      "name": "kube-system",
      "status": "Active",
      "age": "30d"
    },
    {
      "name": "production",
      "status": "Active",
      "age": "15d"
    }
  ],
  "count": 3
}
```

**Example**:
```bash
curl http://localhost:8080/api/namespaces
```

---

### Cluster Resources

#### GET /api/cluster/health

Get cluster health summary.

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "pods": {
    "total": 0,
    "healthy": 0
  },
  "deployments": {
    "total": 0,
    "healthy": 0
  },
  "events": {
    "warnings": 3,
    "errors": 0
  },
  "overall_health": "healthy"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/cluster/health?namespace=production"
```

---

#### GET /api/nodes

List all nodes in the cluster.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "nodes": [
    {
      "name": "minikube",
      "status": "Ready",
      "roles": ["control-plane"],
      "age": "30d",
      "version": "v1.28.0",
      "cpu": "4",
      "memory": "8Gi"
    }
  ],
  "count": 1
}
```

**Example**:
```bash
curl http://localhost:8080/api/nodes
```

---

#### GET /api/services

Get all services across namespaces (for PathTracer).

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "services": [
    {
      "name": "ingress-gateway",
      "namespace": "",
      "type": "ingress",
      "display": "Ingress Gateway"
    },
    {
      "name": "egress-gateway",
      "namespace": "",
      "type": "egress",
      "display": "Egress Gateway"
    },
    {
      "name": "frontend",
      "namespace": "production",
      "type": "service",
      "display": "production/frontend"
    },
    {
      "name": "backend",
      "namespace": "production",
      "type": "service",
      "display": "production/backend"
    }
  ],
  "count": 4
}
```

**Example**:
```bash
curl "http://localhost:8080/api/services?namespace=production"
```

---

#### GET /api/namespaces/:namespace/services

Get services in a specific namespace.

**Path Parameters**:
- `namespace`: Namespace name

**Response**: `200 OK`
```json
{
  "services": [
    {
      "name": "frontend",
      "namespace": "production",
      "type": "ClusterIP",
      "cluster_ip": "10.96.0.1",
      "ports": ["80/TCP", "443/TCP"]
    },
    {
      "name": "backend",
      "namespace": "production",
      "type": "ClusterIP",
      "cluster_ip": "10.96.0.2",
      "ports": ["8080/TCP"]
    }
  ],
  "count": 2,
  "namespace": "production"
}
```

**Example**:
```bash
curl http://localhost:8080/api/namespaces/production/services
```

---

#### GET /api/pods

Get pods (returns empty array for frontend compatibility).

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "pods": [],
  "count": 0,
  "namespace": "all"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/pods?namespace=production"
```

---

#### GET /api/deployments

Get deployments (returns empty array for frontend compatibility).

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "deployments": [],
  "count": 0,
  "namespace": "all"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/deployments?namespace=production"
```

---

### Events

#### GET /api/events

Get cluster events.

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.
- `include_normal` (optional): Include normal events. Default: `false` (warnings/errors only).

**Response**: `200 OK`
```json
{
  "events": [
    {
      "type": "Warning",
      "reason": "BackOff",
      "message": "Back-off restarting failed container",
      "object": "pod/frontend-7d9f8c4b5-xyz",
      "namespace": "production",
      "timestamp": "2026-01-20T10:30:00Z",
      "count": 5
    },
    {
      "type": "Warning",
      "reason": "FailedScheduling",
      "message": "0/3 nodes are available: insufficient memory",
      "object": "pod/backend-9c8d7b6a4-abc",
      "namespace": "production",
      "timestamp": "2026-01-20T10:25:00Z",
      "count": 1
    }
  ],
  "count": 2,
  "namespace": "all"
}
```

**Example**:
```bash
# Get warning/error events across all namespaces
curl "http://localhost:8080/api/events?namespace=all"

# Include normal events for specific namespace
curl "http://localhost:8080/api/events?namespace=production&include_normal=true"
```

---

#### GET /api/namespaces/:namespace/events

Get events for a specific namespace (legacy endpoint).

**Path Parameters**:
- `namespace`: Namespace name

**Query Parameters**:
- `include_normal` (optional): Include normal events. Default: `false`.

**Response**: Same as `/api/events`

**Example**:
```bash
curl "http://localhost:8080/api/namespaces/production/events?include_normal=true"
```

---

### Topology & Network Tracing

#### GET /api/topology

Get service topology with network policies and service mesh information.

**Rate Limit**: 5 requests/minute

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "namespace": "production",
  "services": {
    "production/frontend": {
      "name": "frontend",
      "namespace": "production",
      "type": "ClusterIP",
      "cluster_ip": "10.96.0.1",
      "ports": ["80/TCP"],
      "pods": ["frontend-7d9f8c4b5-xyz"],
      "pod_count": 1,
      "healthy_pods": 1,
      "deployment": "frontend",
      "has_service_mesh": true,
      "service_mesh_type": "istio",
      "has_cilium_proxy": false
    }
  },
  "connectivity": {
    "production/frontend": {
      "service": "production/frontend",
      "connections": [
        {
          "target": "production/backend",
          "allowed": true,
          "reason": "NetworkPolicy allows",
          "via_service_mesh": true,
          "service_mesh_type": "istio",
          "blocked_by_policy": false,
          "port": "8080",
          "protocol": "TCP"
        }
      ],
      "can_reach": ["production/backend", "production/database"],
      "blocked_from": [],
      "istio_rules": [
        {
          "type": "virtualservice",
          "name": "frontend-route",
          "description": "HTTP routing rules"
        }
      ]
    }
  },
  "network_policies": [
    {
      "name": "allow-frontend",
      "namespace": "production",
      "type": "kubernetes"
    }
  ],
  "cilium_policies": [
    {
      "name": "frontend-egress",
      "namespace": "production",
      "type": "ciliumnetworkpolicy"
    }
  ],
  "istio_policies": [
    {
      "name": "frontend-route",
      "namespace": "production",
      "type": "virtualservice"
    }
  ],
  "kyverno_policies": [],
  "infrastructure": {
    "cni": "cilium",
    "cilium_enabled": true,
    "istio_enabled": true,
    "kyverno_enabled": false,
    "network_policies": 5,
    "cilium_policies": 3,
    "istio_policies": 2,
    "kyverno_policies": 0
  },
  "summary": {
    "total_services": 3,
    "total_connections": 4,
    "blocked_connections": 0,
    "service_mesh_services": 2
  },
  "hubble_enabled": true
}
```

**Example**:
```bash
curl "http://localhost:8080/api/topology?namespace=production"
```

---

#### GET /api/topology/trace

Trace network path between two services.

**Rate Limit**: 5 requests/minute

**Query Parameters**:
- `source` (required): Source service name
- `destination` (required): Destination service name
- `namespace` (optional): Namespace scope. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "source": "production/frontend",
  "destination": "production/backend",
  "path": [
    {
      "hop": 1,
      "service": "production/frontend",
      "type": "source"
    },
    {
      "hop": 2,
      "service": "istio-ingressgateway",
      "type": "ingress",
      "policies": ["allow-ingress"]
    },
    {
      "hop": 3,
      "service": "production/backend",
      "type": "destination"
    }
  ],
  "allowed": true,
  "blocked_by": [],
  "service_mesh_hops": ["istio-ingressgateway"],
  "network_policies": ["allow-frontend-to-backend"],
  "details": "Path allowed via Istio service mesh with NetworkPolicy allow-frontend-to-backend"
}
```

**Error Response**: `400 Bad Request`
```json
{
  "error": "source and destination parameters required"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/topology/trace?source=production/frontend&destination=production/backend&namespace=production"
```

---

#### GET /api/networkpolicy/:policy_name

Get NetworkPolicy YAML.

**Path Parameters**:
- `policy_name`: NetworkPolicy name

**Query Parameters**:
- `namespace` (optional): Namespace name. Defaults to current namespace.

**Response**: `200 OK`
```json
{
  "name": "allow-frontend",
  "namespace": "production",
  "yaml": "{\n  \"apiVersion\": \"networking.k8s.io/v1\",\n  \"kind\": \"NetworkPolicy\",\n  ..."
}
```

**Error Response**: `404 Not Found`
```json
{
  "error": "NetworkPolicy not found: allow-frontend"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/networkpolicy/allow-frontend?namespace=production"
```

---

### Recommendations & Compliance

#### GET /api/recommendations

Get security and best practice recommendations.

**Rate Limit**: 10 requests/minute

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "recommendations": [
    {
      "id": "rec-001",
      "title": "Add NetworkPolicy for service 'frontend'",
      "description": "Service 'frontend' has no NetworkPolicy restricting ingress/egress traffic",
      "priority": "HIGH",
      "category": "security",
      "resource_type": "service",
      "resource_name": "frontend",
      "namespace": "production",
      "remediation": "Create a NetworkPolicy to restrict traffic to only required services"
    },
    {
      "id": "rec-002",
      "title": "Enable mTLS for service mesh",
      "description": "Istio is installed but mTLS is not enforced",
      "priority": "CRITICAL",
      "category": "security",
      "resource_type": "cluster",
      "namespace": "",
      "remediation": "Enable strict mTLS mode in Istio configuration"
    }
  ],
  "count": 2,
  "namespace": "production"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/recommendations?namespace=production"
```

---

#### GET /api/recommendations/score

Get compliance score based on best practices.

**Rate Limit**: 10 requests/minute

**Query Parameters**:
- `namespace` (optional): Filter by namespace. Use `all` or omit for all namespaces.

**Response**: `200 OK`
```json
{
  "score": 75.5,
  "passed": 15,
  "total": 20,
  "details": [
    {
      "check": "network_policies_exist",
      "passed": true,
      "description": "NetworkPolicies are configured"
    },
    {
      "check": "mtls_enabled",
      "passed": false,
      "description": "mTLS is not enforced"
    },
    {
      "check": "pod_security_standards",
      "passed": true,
      "description": "Pod Security Standards are applied"
    }
  ],
  "recommendations_count": 5,
  "namespace": "production"
}
```

**Example**:
```bash
curl "http://localhost:8080/api/recommendations/score?namespace=production"
```

---

### Policy Management

#### POST /api/policies/cilium/build

Build a Cilium NetworkPolicy from specification.

**Request Body**:
```json
{
  "name": "allow-frontend-egress",
  "namespace": "production",
  "description": "Allow frontend to access backend",
  "endpoint_selector": {
    "matchLabels": {
      "app": "frontend"
    }
  },
  "egress": [
    {
      "to_endpoints": [
        {
          "matchLabels": {
            "app": "backend"
          }
        }
      ],
      "to_ports": [
        {
          "ports": [
            {
              "port": "8080",
              "protocol": "TCP"
            }
          ]
        }
      ]
    }
  ]
}
```

**Response**: `200 OK`
```json
{
  "yaml": "apiVersion: cilium.io/v2\nkind: CiliumNetworkPolicy\nmetadata:\n  name: allow-frontend-egress\n  namespace: production\nspec:\n  endpointSelector:\n    matchLabels:\n      app: frontend\n  egress:\n  - toEndpoints:\n    - matchLabels:\n        app: backend\n    toPorts:\n    - ports:\n      - port: \"8080\"\n        protocol: TCP\n",
  "name": "allow-frontend-egress",
  "namespace": "production"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/policies/cilium/build \
  -H "Content-Type: application/json" \
  -d '{
    "name": "allow-frontend-egress",
    "namespace": "production",
    "endpoint_selector": {
      "matchLabels": {"app": "frontend"}
    },
    "egress": [{
      "to_endpoints": [{"matchLabels": {"app": "backend"}}],
      "to_ports": [{
        "ports": [{"port": "8080", "protocol": "TCP"}]
      }]
    }]
  }'
```

---

#### POST /api/policies/cilium/apply

Apply a Cilium NetworkPolicy to the cluster.

**Request Body**:
```json
{
  "yaml": "apiVersion: cilium.io/v2\nkind: CiliumNetworkPolicy\nmetadata:\n  name: allow-frontend-egress\n  namespace: production\nspec:\n  ...",
  "namespace": "production"
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "message": "Policy applied successfully",
  "namespace": "production"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/policies/cilium/apply \
  -H "Content-Type: application/json" \
  -d '{
    "yaml": "apiVersion: cilium.io/v2\nkind: CiliumNetworkPolicy\n...",
    "namespace": "production"
  }'
```

---

#### POST /api/policies/cilium/export

Export a Cilium NetworkPolicy as downloadable YAML file.

**Request Body**:
```json
{
  "yaml": "apiVersion: cilium.io/v2\nkind: CiliumNetworkPolicy\n..."
}
```

**Response**: `200 OK` (application/x-yaml)
```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-frontend-egress
  namespace: production
spec:
  ...
```

**Headers**:
- `Content-Type: application/x-yaml`
- `Content-Disposition: attachment; filename=cilium-network-policy.yaml`

**Example**:
```bash
curl -X POST http://localhost:8080/api/policies/cilium/export \
  -H "Content-Type: application/json" \
  -d '{"yaml": "apiVersion: cilium.io/v2\n..."}' \
  -o policy.yaml
```

---

#### DELETE /api/policies/cilium/:name

Delete a Cilium NetworkPolicy.

**Path Parameters**:
- `name`: Policy name

**Query Parameters**:
- `namespace` (optional): Namespace name. Defaults to current namespace.

**Response**: `200 OK`
```json
{
  "success": true,
  "message": "Policy deleted successfully"
}
```

**Example**:
```bash
curl -X DELETE "http://localhost:8080/api/policies/cilium/allow-frontend-egress?namespace=production"
```

---

#### POST /api/policies/kyverno/build

Build a Kyverno Policy from specification.

**Request Body**:
```json
{
  "name": "require-labels",
  "namespace": "production",
  "type": "Policy",
  "description": "Require specific labels on pods",
  "rules": [
    {
      "name": "check-labels",
      "match_resources": {
        "kinds": ["Pod"]
      },
      "validate": {
        "message": "Label 'app' is required",
        "pattern": {
          "metadata": {
            "labels": {
              "app": "?*"
            }
          }
        }
      }
    }
  ]
}
```

**Response**: `200 OK`
```json
{
  "yaml": "apiVersion: kyverno.io/v1\nkind: Policy\nmetadata:\n  name: require-labels\n  namespace: production\nspec:\n  ...",
  "name": "require-labels",
  "namespace": "production",
  "type": "Policy"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/policies/kyverno/build \
  -H "Content-Type: application/json" \
  -d '{
    "name": "require-labels",
    "namespace": "production",
    "type": "Policy",
    "rules": [{
      "name": "check-labels",
      "match_resources": {"kinds": ["Pod"]},
      "validate": {
        "message": "Label app is required",
        "pattern": {"metadata": {"labels": {"app": "?*"}}}
      }
    }]
  }'
```

---

#### POST /api/policies/kyverno/apply

Apply a Kyverno Policy to the cluster.

**Request Body**: Same as Cilium apply endpoint

**Response**: Same as Cilium apply endpoint

**Example**:
```bash
curl -X POST http://localhost:8080/api/policies/kyverno/apply \
  -H "Content-Type: application/json" \
  -d '{
    "yaml": "apiVersion: kyverno.io/v1\nkind: Policy\n...",
    "namespace": "production"
  }'
```

---

#### POST /api/policies/kyverno/export

Export a Kyverno Policy as downloadable YAML file.

**Request Body**:
```json
{
  "yaml": "apiVersion: kyverno.io/v1\nkind: Policy\n...",
  "type": "Policy"
}
```

**Response**: `200 OK` (application/x-yaml)

**Headers**:
- `Content-Type: application/x-yaml`
- `Content-Disposition: attachment; filename=kyverno-policy.yaml` (or `kyverno-cluster-policy.yaml` for ClusterPolicy)

**Example**:
```bash
curl -X POST http://localhost:8080/api/policies/kyverno/export \
  -H "Content-Type: application/json" \
  -d '{"yaml": "apiVersion: kyverno.io/v1\n...", "type": "Policy"}' \
  -o policy.yaml
```

---

#### DELETE /api/policies/kyverno/:name

Delete a Kyverno Policy.

**Path Parameters**:
- `name`: Policy name

**Query Parameters**:
- `namespace` (optional): Namespace name. Required for Policy (not ClusterPolicy).
- `cluster_policy` (optional): Set to `true` for ClusterPolicy. Default: `false`.

**Response**: `200 OK`
```json
{
  "success": true,
  "message": "Policy deleted successfully"
}
```

**Example**:
```bash
# Delete namespaced Policy
curl -X DELETE "http://localhost:8080/api/policies/kyverno/require-labels?namespace=production"

# Delete ClusterPolicy
curl -X DELETE "http://localhost:8080/api/policies/kyverno/global-policy?cluster_policy=true"
```

---

### Configuration

#### GET /api/config

Get complete configuration.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "kubeconfig": {
    "path": "/home/user/.kube/config",
    "context": "minikube"
  },
  "llm": {
    "providers": {
      "openai": {
        "enabled": true,
        "model": "gpt-4",
        "api_key": "***REDACTED***"
      },
      "anthropic": {
        "enabled": false,
        "model": "claude-3-opus",
        "api_key": ""
      }
    },
    "default_provider": "openai"
  },
  "server": {
    "port": 8080,
    "enable_cors": true
  }
}
```

**Note**: API keys are redacted in responses for security.

**Example**:
```bash
curl http://localhost:8080/api/config
```

---

#### GET /api/config/providers

Get LLM providers configuration summary.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "providers": {
    "openai": {
      "enabled": true,
      "model": "gpt-4",
      "has_key": true
    },
    "anthropic": {
      "enabled": false,
      "model": "claude-3-opus",
      "has_key": false
    }
  },
  "kubeconfig": {
    "path": "/home/user/.kube/config"
  },
  "sops_available": false
}
```

**Example**:
```bash
curl http://localhost:8080/api/config/providers
```

---

#### GET /api/config/kubeconfig/status

Get kubeconfig status and connectivity.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "configured": true,
  "path": "/home/user/.kube/config",
  "found": true,
  "auto_found": false,
  "connected": true,
  "context": "minikube",
  "error": ""
}
```

**Response (Not Configured)**: `200 OK`
```json
{
  "configured": false,
  "path": "",
  "found": false,
  "auto_found": false,
  "connected": false,
  "error": "Kubeconfig not configured. Please set kubeconfig path in Settings."
}
```

**Example**:
```bash
curl http://localhost:8080/api/config/kubeconfig/status
```

---

#### POST /api/config/kubeconfig

Set kubeconfig path and context.

**Request Body**:
```json
{
  "path": "/home/user/.kube/config",
  "context": "minikube"
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "path": "/home/user/.kube/config",
  "context": "minikube",
  "message": "Kubeconfig configured successfully"
}
```

**Error Response**: `400 Bad Request`
```json
{
  "error": "Kubeconfig file not found: /invalid/path"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/config/kubeconfig \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/home/user/.kube/config",
    "context": "minikube"
  }'
```

---

#### POST /api/config/providers/:provider/model

Set model for an LLM provider.

**Path Parameters**:
- `provider`: Provider name (e.g., `openai`, `anthropic`)

**Request Body**:
```json
{
  "model": "gpt-4-turbo"
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "provider": "openai",
  "model": "gpt-4-turbo"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/config/providers/openai/model \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4-turbo"}'
```

---

#### POST /api/config/providers/:provider/enable

Enable or disable an LLM provider.

**Path Parameters**:
- `provider`: Provider name

**Request Body**:
```json
{
  "enabled": true
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "provider": "openai",
  "enabled": true
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/config/providers/openai/enable \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'
```

---

#### POST /api/config/providers/:provider/api-key

Set API key for an LLM provider.

**Path Parameters**:
- `provider`: Provider name

**Request Body**:
```json
{
  "api_key": "sk-..."
}
```

**Response**: `200 OK`
```json
{
  "success": true,
  "provider": "openai",
  "has_key": true
}
```

**Error Response**: `400 Bad Request`
```json
{
  "error": "API key is required"
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/config/providers/openai/api-key \
  -H "Content-Type: application/json" \
  -d '{"api_key": "sk-..."}'
```

---

#### PUT /api/config

Update configuration (partial update).

**Request Body**:
```json
{
  "providers": {
    "openai": {
      "enabled": true,
      "api_key": "sk-..."
    }
  }
}
```

**Response**: `200 OK`
```json
{
  "message": "Configuration updated successfully"
}
```

**Example**:
```bash
curl -X PUT http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "providers": {
      "openai": {"enabled": true}
    }
  }'
```

---

#### POST /api/config/setup-wizard

Get information about the setup wizard.

**Query Parameters**: None

**Response**: `200 OK`
```json
{
  "message": "Setup wizard is available via CLI",
  "command": "stargazer config setup",
  "info": "The setup wizard is an interactive CLI tool. Run 'stargazer config setup' in your terminal."
}
```

**Example**:
```bash
curl -X POST http://localhost:8080/api/config/setup-wizard
```

---

## WebSocket API

### Connection

**Endpoint**: `ws://localhost:8080/ws`

**Protocol**: WebSocket (RFC 6455)

### Connection Flow

1. **Connect**: Client establishes WebSocket connection
2. **Welcome Message**: Server sends initial message
3. **Bidirectional Communication**: Client and server exchange messages
4. **Heartbeat**: Server sends ping every 54 seconds
5. **Disconnect**: Either party closes connection

### Message Format

All WebSocket messages use JSON format:

```json
{
  "type": "message_type",
  "data": {
    // message-specific data
  }
}
```

### Server → Client Messages

#### Connected Event

Sent immediately after connection is established.

```json
{
  "type": "connected",
  "data": {
    "message": "Connected to Stargazer",
    "timestamp": "2026-01-20T10:30:00Z"
  }
}
```

#### Policy Change Event

Sent when a policy is created, updated, or deleted.

```json
{
  "type": "policy_change",
  "data": {
    "event_type": "ADDED",
    "policy_type": "CiliumNetworkPolicy",
    "name": "allow-frontend-egress",
    "namespace": "production",
    "timestamp": 1705748400
  }
}
```

**Event Types**:
- `ADDED`: Policy created
- `MODIFIED`: Policy updated
- `DELETED`: Policy deleted

**Policy Types**:
- `NetworkPolicy`: Kubernetes NetworkPolicy
- `CiliumNetworkPolicy`: Cilium NetworkPolicy
- `CiliumClusterwideNetworkPolicy`: Cilium cluster-wide policy
- `Policy`: Kyverno Policy
- `ClusterPolicy`: Kyverno ClusterPolicy

### Client → Server Messages

Currently, client messages are logged but not processed. Future versions may support:
- Subscription to specific events
- Real-time query requests
- Custom event filtering

**Example Client Message**:
```json
{
  "type": "subscribe",
  "data": {
    "events": ["policy_change"],
    "namespace": "production"
  }
}
```

### Connection Management

**Keep-Alive**: Server sends ping frames every 54 seconds. Client must respond with pong.

**Timeouts**:
- Read timeout: 60 seconds
- Write timeout: 10 seconds

**Reconnection**: Client should implement exponential backoff for reconnection attempts.

### JavaScript Example

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected to Stargazer');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message.type, message.data);

  if (message.type === 'policy_change') {
    console.log(`Policy ${message.data.event_type}: ${message.data.name}`);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from Stargazer');
  // Implement reconnection logic
};
```

### Go Example

```go
package main

import (
    "encoding/json"
    "log"
    "github.com/gorilla/websocket"
)

type Message struct {
    Type string                 `json:"type"`
    Data map[string]interface{} `json:"data"`
}

func main() {
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    for {
        var msg Message
        err := conn.ReadJSON(&msg)
        if err != nil {
            log.Println("Read error:", err)
            break
        }

        log.Printf("Received: %s - %v\n", msg.Type, msg.Data)
    }
}
```

---

## Error Responses

### Standard Error Format

All error responses follow this format:

```json
{
  "error": "Error description"
}
```

### Common HTTP Status Codes

| Code | Status | Description |
|------|--------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request parameters or body |
| 404 | Not Found | Resource not found |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 501 | Not Implemented | Feature not yet implemented |
| 503 | Service Unavailable | Kubernetes client not initialized or cluster unreachable |

### Error Examples

#### Invalid Namespace

```json
{
  "error": "Invalid namespace parameter: namespace too long (max 253 characters)"
}
```

#### Kubernetes Client Not Available

```json
{
  "error": "Kubernetes client not initialized. Please configure kubeconfig."
}
```

#### Resource Not Found

```json
{
  "error": "NetworkPolicy not found: allow-frontend"
}
```

#### Rate Limit Exceeded

```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

#### Missing Required Parameters

```json
{
  "error": "source and destination parameters required"
}
```

---

## Data Models

### ServiceInfo

```typescript
interface ServiceInfo {
  name: string;
  namespace: string;
  type: string; // "ClusterIP", "NodePort", "LoadBalancer", etc.
  cluster_ip: string;
  ports: string[]; // e.g., ["80/TCP", "443/TCP"]
  pods: string[];
  pod_count: number;
  healthy_pods: number;
  deployment?: string;
  has_service_mesh: boolean;
  service_mesh_type?: string; // "istio", "cilium", ""
  has_cilium_proxy: boolean;
}
```

### ConnectivityInfo

```typescript
interface ConnectivityInfo {
  service: string;
  connections: ServiceConnection[];
  can_reach: string[];
  blocked_from: string[];
  istio_rules?: IstioRuleInfo[];
  cilium_rules?: CiliumRuleInfo[];
}
```

### ServiceConnection

```typescript
interface ServiceConnection {
  target: string;
  allowed: boolean;
  reason: string;
  via_service_mesh: boolean;
  service_mesh_type?: string;
  blocked_by_policy: boolean;
  blocking_policies?: string[];
  port?: string;
  protocol?: string;
}
```

### NetworkPolicyInfo

```typescript
interface NetworkPolicyInfo {
  name: string;
  namespace: string;
  type: string; // "kubernetes" or "cilium"
  yaml?: string;
}
```

### Event

```typescript
interface Event {
  type: string; // "Normal", "Warning", "Error"
  reason: string;
  message: string;
  object: string; // e.g., "pod/frontend-7d9f8c4b5-xyz"
  namespace: string;
  timestamp: string; // ISO 8601 format
  count: number;
}
```

### Recommendation

```typescript
interface Recommendation {
  id: string;
  title: string;
  description: string;
  priority: "CRITICAL" | "HIGH" | "WARNING" | "INFO";
  category: string; // e.g., "security", "performance"
  resource_type: string; // e.g., "service", "pod", "cluster"
  resource_name: string;
  namespace: string;
  remediation: string;
}
```

### CiliumNetworkPolicySpec

```typescript
interface CiliumNetworkPolicySpec {
  name: string;
  namespace: string;
  description?: string;
  endpoint_selector?: {
    matchLabels?: Record<string, string>;
  };
  ingress?: CiliumIngressRule[];
  egress?: CiliumEgressRule[];
}

interface CiliumIngressRule {
  from_endpoints?: Array<{ matchLabels?: Record<string, string> }>;
  from_cidr?: string[];
  from_entities?: string[];
  to_ports?: CiliumPortRule[];
}

interface CiliumEgressRule {
  to_endpoints?: Array<{ matchLabels?: Record<string, string> }>;
  to_cidr?: string[];
  to_entities?: string[];
  to_ports?: CiliumPortRule[];
  to_services?: Array<{ k8sService?: { serviceName: string; namespace: string } }>;
}

interface CiliumPortRule {
  ports?: CiliumPortProtocol[];
  rules?: Record<string, any>;
}

interface CiliumPortProtocol {
  port: string;
  protocol: string; // "TCP", "UDP", "SCTP"
}
```

### KyvernoPolicySpec

```typescript
interface KyvernoPolicySpec {
  name: string;
  namespace?: string; // Empty for ClusterPolicy
  type: "Policy" | "ClusterPolicy";
  description?: string;
  rules: KyvernoRule[];
  validation?: boolean;
  mutation?: boolean;
  generation?: boolean;
}

interface KyvernoRule {
  name: string;
  match_resources: Record<string, any>;
  exclude_resources?: Record<string, any>;
  validate?: Record<string, any>;
  mutate?: Record<string, any>;
  generate?: Record<string, any>;
}
```

---

## Best Practices

### Rate Limiting

- Cache expensive topology calls locally
- Use WebSocket for real-time updates instead of polling
- Implement exponential backoff for failed requests

### Error Handling

- Always check HTTP status codes
- Parse error messages for user-friendly display
- Implement retry logic for transient failures (503, 429)

### WebSocket

- Implement reconnection with exponential backoff
- Handle ping/pong for connection keep-alive
- Buffer messages during disconnection

### Security

- Never expose this API to the public internet without authentication
- Use HTTPS in production with proper TLS certificates
- Rotate API keys regularly
- Use RBAC in Kubernetes to limit service account permissions

### Performance

- Use `namespace` parameter to scope queries when possible
- Leverage WebSocket for real-time updates
- Monitor `/api/metrics` endpoint for performance tracking

---

## Version History

### v0.1.0-dev (Current)

- Initial API release
- REST API endpoints for cluster resources
- WebSocket support for real-time updates
- Cilium and Kyverno policy management
- Network topology and path tracing
- Recommendations and compliance scoring

---

## Support

For issues, feature requests, or questions:

- **GitHub**: https://github.com/maplecitymadman/stargazer
- **Documentation**: See project README.md

---

**Last Updated**: 2026-01-20
