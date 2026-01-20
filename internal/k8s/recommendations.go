package k8s

import (
	"context"
	"fmt"
	"strings"
)

// NetworkingBestPractice defines a best practice rule
type NetworkingBestPractice struct {
	ID          string
	Name        string
	Description string
	Category    string // "security", "performance", "observability", "resilience"
	Severity    string // "critical", "high", "medium", "low"
	Check       func(ctx context.Context, topology *TopologyData) (bool, []Recommendation)
}

// Recommendation represents a specific recommendation for the cluster
type Recommendation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Service     string `json:"service,omitempty"` // Optional: specific service this applies to
	Namespace   string `json:"namespace,omitempty"`
	Fix         FixRecommendation `json:"fix"`
	Impact      string `json:"impact"` // What will improve if applied
}

// FixRecommendation provides the fix for a recommendation
type FixRecommendation struct {
	Type        string   `json:"type"`        // "networkpolicy", "ciliumpolicy", "istiopolicy", "ingress", "egress"
	Template    string   `json:"template"`    // YAML template
	Command     string   `json:"command,omitempty"`     // Optional: kubectl command
	ManualSteps []string `json:"manual_steps,omitempty"` // Steps if manual intervention needed
}

// GetRecommendations evaluates the cluster against best practices and returns recommendations
func (c *Client) GetRecommendations(ctx context.Context, topology *TopologyData) ([]Recommendation, error) {
	var allRecommendations []Recommendation

	for _, practice := range BestPractices {
		_, recommendations := practice.Check(ctx, topology)
		allRecommendations = append(allRecommendations, recommendations...)
	}

	return allRecommendations, nil
}

// GetComplianceScore calculates a compliance score (0-100) based on best practices
func (c *Client) GetComplianceScore(ctx context.Context, topology *TopologyData) (int, map[string]interface{}) {
	totalChecks := len(BestPractices)
	passedChecks := 0
	details := make(map[string]interface{})

	recommendations, _ := c.GetRecommendations(ctx, topology)

	for _, practice := range BestPractices {
		passed, _ := practice.Check(ctx, topology)
		if passed {
			passedChecks++
		}
		details[practice.ID] = map[string]interface{}{
			"name":   practice.Name,
			"passed": passed,
		}
	}

	score := 0
	if totalChecks > 0 {
		score = (passedChecks * 100) / totalChecks
	}

	return score, map[string]interface{}{
		"score":                score,
		"passed":               passedChecks,
		"total":                totalChecks,
		"check_details":       details,
		"recommendations_count": len(recommendations),
	}
}

// BestPractices defines the list of networking best practices to check
var BestPractices = []NetworkingBestPractice{
	{
		ID:          "np-001",
		Name:        "All Services Should Have NetworkPolicies",
		Description: "Every service should have at least one NetworkPolicy defining ingress/egress rules",
		Category:    "security",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation
			servicesWithoutPolicy := 0

			for serviceKey, service := range topology.Services {
				// Skip system namespaces
				if strings.HasPrefix(service.Namespace, "kube-") ||
					strings.HasPrefix(service.Namespace, "istio-") ||
					service.Namespace == "default" && service.Name == "kubernetes" {
					continue
				}

				hasPolicy := false
				for _, np := range topology.NetworkPolicies {
					if np.Namespace == service.Namespace {
						hasPolicy = true
						break
					}
				}

				if !hasPolicy {
					servicesWithoutPolicy++
					recommendations = append(recommendations, Recommendation{
						ID:          fmt.Sprintf("np-001-%s", serviceKey),
						Title:       fmt.Sprintf("Service %s/%s lacks NetworkPolicy", service.Namespace, service.Name),
						Description: "This service has no NetworkPolicy, allowing all ingress/egress traffic by default",
						Category:    "security",
						Severity:    "high",
						Service:     serviceKey,
						Namespace:   service.Namespace,
						Fix: FixRecommendation{
							Type:     "networkpolicy",
							Template: generateDefaultNetworkPolicy(service),
							Command:  fmt.Sprintf("kubectl apply -f - <<EOF\n%s\nEOF", generateDefaultNetworkPolicy(service)),
						},
						Impact: "Restricts traffic to only necessary connections, improving security posture",
					})
				}
			}

			// Limit to top 10 most critical
			if len(recommendations) > 10 {
				recommendations = recommendations[:10]
			}

			return servicesWithoutPolicy == 0, recommendations
		},
	},
	{
		ID:          "ingress-001",
		Name:        "Ingress Should Use TLS",
		Description: "All ingress routes should have TLS certificates configured for secure communication",
		Category:    "security",
		Severity:    "critical",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Ingress.KubernetesIngress != nil {
				for _, ing := range topology.Ingress.KubernetesIngress {
					if !ing.TLS {
						recommendations = append(recommendations, Recommendation{
							ID:          fmt.Sprintf("ingress-001-%s", ing.Name),
							Title:       fmt.Sprintf("Ingress %s/%s missing TLS configuration", ing.Namespace, ing.Name),
							Description: "This ingress route does not have TLS configured, exposing traffic in plaintext",
							Category:    "security",
							Severity:    "critical",
							Namespace:   ing.Namespace,
							Fix: FixRecommendation{
								Type: "ingress",
								Template: generateTLSIngress(ing),
								ManualSteps: []string{
									"1. Ensure cert-manager is installed",
									"2. Create or reference a TLS certificate",
									"3. Add TLS section to ingress spec",
								},
							},
							Impact: "Encrypts traffic between clients and services, preventing man-in-the-middle attacks",
						})
					}
				}
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "egress-001",
		Name:        "Egress Should Route Through Gateway",
		Description: "External egress should route through Istio EgressGateway for better control and observability",
		Category:    "security",
		Severity:    "medium",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Infrastructure.IstioEnabled && !topology.Egress.HasEgressGateway {
				recommendations = append(recommendations, Recommendation{
					ID:          "egress-001",
					Title:       "Egress traffic not routed through gateway",
					Description: "Services are accessing external resources directly instead of through Istio EgressGateway",
					Category:    "security",
					Severity:    "medium",
					Fix: FixRecommendation{
						Type: "egress",
						ManualSteps: []string{
							"1. Configure Istio EgressGateway",
							"2. Create ServiceEntry for external services",
							"3. Update services to route through gateway",
						},
					},
					Impact: "Provides centralized control, monitoring, and policy enforcement for external traffic",
				})
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "np-002",
		Name:        "No Default Allow-All Policies",
		Description: "NetworkPolicies should not allow all traffic (empty rules)",
		Category:    "security",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			// This is a simplified check - in reality would need to parse policy YAML
			// For now, we'll check if there are policies that might be too permissive
			if len(topology.NetworkPolicies) > 0 {
				// Check if we have a reasonable policy-to-service ratio
				serviceCount := len(topology.Services)
				policyCount := len(topology.NetworkPolicies)

				if policyCount < serviceCount/2 {
					recommendations = append(recommendations, Recommendation{
						ID:          "np-002",
						Title:       "Insufficient NetworkPolicy coverage",
						Description: fmt.Sprintf("Only %d NetworkPolicies for %d services - many services may be unprotected", policyCount, serviceCount),
						Category:    "security",
						Severity:    "high",
						Fix: FixRecommendation{
							Type: "networkpolicy",
							ManualSteps: []string{
								"1. Review services without NetworkPolicies",
								"2. Create restrictive policies for each service",
								"3. Test connectivity after applying policies",
							},
						},
						Impact: "Reduces attack surface by restricting unnecessary network access",
					})
				}
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "mesh-001",
		Name:        "Service Mesh Coverage",
		Description: "At least 80% of services should be part of the service mesh for observability and security",
		Category:    "observability",
		Severity:    "medium",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Summary.TotalServices > 0 {
				coveragePercent := (topology.Summary.ServicesWithMesh * 100) / topology.Summary.TotalServices
				if coveragePercent < 80 {
					recommendations = append(recommendations, Recommendation{
						ID:          "mesh-001",
						Title:       "Low service mesh coverage",
						Description: fmt.Sprintf("Only %d%% of services are in the service mesh (target: 80%%)", coveragePercent),
						Category:    "observability",
						Severity:    "medium",
						Fix: FixRecommendation{
							Type: "istio",
							ManualSteps: []string{
								"1. Enable Istio sidecar injection for namespaces",
								"2. Label namespaces: kubectl label namespace <ns> istio-injection=enabled",
								"3. Restart pods to inject sidecars",
							},
						},
						Impact: "Improves observability, security (mTLS), and traffic management capabilities",
					})
				}
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "blocked-001",
		Name:        "Resolve Blocked Connections",
		Description: "Blocked connections should be investigated and resolved",
		Category:    "resilience",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Summary.BlockedConnections > 0 {
				blockedPercent := (topology.Summary.BlockedConnections * 100) / topology.Summary.TotalConnections
				if blockedPercent > 10 {
					recommendations = append(recommendations, Recommendation{
						ID:          "blocked-001",
						Title:       "High number of blocked connections",
						Description: fmt.Sprintf("%d blocked connections (%.1f%%) may indicate misconfigured policies", topology.Summary.BlockedConnections, float64(blockedPercent)),
						Category:    "resilience",
						Severity:    "high",
						Fix: FixRecommendation{
							Type: "diagnostic",
							ManualSteps: []string{
								"1. Review blocked connections in topology view",
								"2. Identify which policies are blocking traffic",
								"3. Update policies to allow necessary traffic",
								"4. Use path tracer to debug connection issues",
							},
						},
						Impact: "Restores service connectivity and improves application reliability",
					})
				}
			}

			return len(recommendations) == 0, recommendations
		},
	},
}

// Helper function to generate default NetworkPolicy template
func generateDefaultNetworkPolicy(service ServiceInfo) string {
	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: %s-network-policy
  namespace: %s
spec:
  podSelector:
    matchLabels:
      app: %s
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 80
    - protocol: TCP
      port: 443
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
  - to:
    - namespaceSelector: {}
`, service.Name, service.Namespace, service.Name)
}

// Helper function to generate TLS ingress template
func generateTLSIngress(ing KubernetesIngressInfo) string {
	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s
  namespace: %s
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - %s
    secretName: %s-tls
  rules:
  - host: %s
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: %s
            port:
              number: 80
`, ing.Name, ing.Namespace, strings.Join(ing.Hosts, "\n    - "), ing.Name, ing.Hosts[0], ing.Backend)
}
