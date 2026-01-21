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
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Severity    string            `json:"severity"`
	Service     string            `json:"service,omitempty"` // Optional: specific service this applies to
	Namespace   string            `json:"namespace,omitempty"`
	Fix         FixRecommendation `json:"fix"`
	Impact      string            `json:"impact"` // What will improve if applied
}

// FixRecommendation provides the fix for a recommendation
type FixRecommendation struct {
	Type        string   `json:"type"`                   // "networkpolicy", "ciliumpolicy", "istiopolicy", "ingress", "egress"
	Template    string   `json:"template"`               // YAML template
	Command     string   `json:"command,omitempty"`      // Optional: kubectl command
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
		"score":                 score,
		"passed":                passedChecks,
		"total":                 totalChecks,
		"check_details":         details,
		"recommendations_count": len(recommendations),
	}
}

// BestPractices defines the list of networking best practices to check
// These are tailored to the actual infrastructure detected in the cluster
var BestPractices = []NetworkingBestPractice{
	{
		ID:          "np-001",
		Name:        "Services Should Have Network Policies",
		Description: "Non-system services should have NetworkPolicies or CiliumNetworkPolicies for defense in depth",
		Category:    "security",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation
			servicesWithoutPolicy := 0

			// Skip namespaces that typically don't need policies
			skipNamespaces := map[string]bool{
				"kube-system":     true,
				"kube-public":     true,
				"kube-node-lease": true,
				"istio-system":    true,
			}

			for serviceKey, service := range topology.Services {
				// Skip system namespaces
				if skipNamespaces[service.Namespace] ||
					strings.HasPrefix(service.Namespace, "kube-") ||
					strings.HasPrefix(service.Namespace, "istio-") {
					continue
				}

				hasK8sPolicy := false
				hasCiliumPolicy := false

				// Check for K8s NetworkPolicy
				for _, np := range topology.NetworkPolicies {
					if np.Namespace == service.Namespace {
						hasK8sPolicy = true
						break
					}
				}

				// Check for Cilium NetworkPolicy (if Cilium is enabled)
				if topology.Infrastructure.CiliumEnabled {
					for _, cnp := range topology.CiliumPolicies {
						if cnp.Namespace == service.Namespace || cnp.Namespace == "" {
							hasCiliumPolicy = true
							break
						}
					}
				}

				if !hasK8sPolicy && !hasCiliumPolicy {
					servicesWithoutPolicy++

					// Analyze actual connections to provide specific recommendation
					connectivity, hasConnections := topology.Connectivity[serviceKey]
					connectionCount := 0
					ingressConnections := 0
					egressConnections := 0

					if hasConnections {
						connectionCount = len(connectivity.Connections)
						for _, conn := range connectivity.Connections {
							if conn.Allowed {
								targetNs, targetName, _ := ParseServiceKey(conn.Target)
								if targetNs == service.Namespace && targetName == service.Name {
									ingressConnections++
								} else {
									egressConnections++
								}
							}
						}
					}

					// Generate appropriate policy based on infrastructure and actual connections
					var fixType string
					var template string

					if topology.Infrastructure.CiliumEnabled {
						fixType = "ciliumpolicy"
						template = generateCiliumNetworkPolicy(service, topology)
					} else {
						fixType = "networkpolicy"
						template = generateK8sNetworkPolicy(service, topology)
					}

					description := fmt.Sprintf("Service %s/%s has no %s", service.Namespace, service.Name, fixType)
					if hasConnections {
						description += fmt.Sprintf(". This service has %d active connections (%d ingress, %d egress) that should be protected by policy.", connectionCount, ingressConnections, egressConnections)
					} else {
						description += ". Consider adding one for defense in depth."
					}
					if topology.Infrastructure.IstioEnabled {
						description += " This works alongside Istio AuthorizationPolicies for defense in depth."
					}

					recommendations = append(recommendations, Recommendation{
						ID:          fmt.Sprintf("np-001-%s", serviceKey),
						Title:       fmt.Sprintf("Service %s/%s lacks network policy", service.Namespace, service.Name),
						Description: description,
						Category:    "security",
						Severity:    "high",
						Service:     serviceKey,
						Namespace:   service.Namespace,
						Fix: FixRecommendation{
							Type:     fixType,
							Template: template,
							Command:  fmt.Sprintf("kubectl apply -f - <<EOF\n%s\nEOF", template),
						},
						Impact: fmt.Sprintf("Protects %d active connections and adds defense in depth", connectionCount),
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

			// Only check if cert-manager is likely available (check for cert-manager namespace or cluster issuer)
			hasCertManager := false
			for ns := range topology.Services {
				if strings.Contains(ns, "cert-manager") {
					hasCertManager = true
					break
				}
			}

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
								Type:     "ingress",
								Template: generateTLSIngress(ing, hasCertManager),
								ManualSteps: []string{
									"1. Ensure cert-manager is installed (check cert-manager namespace)",
									"2. Verify ClusterIssuer exists: kubectl get clusterissuer",
									"3. Add TLS section to ingress spec with cert-manager annotation",
									"4. Wait for certificate to be issued (check cert-manager logs if needed)",
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
		ID:          "istio-mtls-001",
		Name:        "Istio mTLS Should Be STRICT",
		Description: "If Istio is enabled, PeerAuthentication should use STRICT mode for production security",
		Category:    "security",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Infrastructure.IstioEnabled {
				// Check if we have any PeerAuthentication policies
				hasPeerAuth := false
				for _, istioPolicy := range topology.IstioPolicies {
					if istioPolicy.Type == "peerauthentication" {
						hasPeerAuth = true
						break
					}
				}

				// Recommend STRICT mTLS if Istio is enabled
				// Note: We can't easily check the actual mode without parsing YAML,
				// so we recommend checking and updating to STRICT
				description := "Istio is detected. Ensure PeerAuthentication is set to STRICT mode for production security."
				if !hasPeerAuth {
					description = "Istio is detected but no PeerAuthentication found. Create one with STRICT mTLS mode."
				}

				recommendations = append(recommendations, Recommendation{
					ID:          "istio-mtls-001",
					Title:       "Enable STRICT mTLS mode in Istio",
					Description: description,
					Category:    "security",
					Severity:    "high",
					Fix: FixRecommendation{
						Type:     "istiopolicy",
						Template: generateStrictMTLSPolicy(),
						ManualSteps: []string{
							"1. Start with PERMISSIVE mode to verify all services work",
							"2. Monitor for any connection issues",
							"3. Gradually migrate namespaces to STRICT mode",
							"4. Update PeerAuthentication to STRICT once verified",
						},
					},
					Impact: "Enforces mutual TLS between all services, preventing unauthorized access even if network policies are bypassed",
				})
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "istio-authz-001",
		Name:        "Istio AuthorizationPolicies Should Be Restrictive",
		Description: "AuthorizationPolicies should be more restrictive than allow-all for production",
		Category:    "security",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Infrastructure.IstioEnabled {
				// Check if we have allow-all policies
				hasAllowAll := false
				for _, istioPolicy := range topology.IstioPolicies {
					if istioPolicy.Type == "authorizationpolicy" && strings.Contains(strings.ToLower(istioPolicy.Name), "allow-all") {
						hasAllowAll = true
						break
					}
				}

				if hasAllowAll {
					recommendations = append(recommendations, Recommendation{
						ID:          "istio-authz-001",
						Title:       "Replace allow-all AuthorizationPolicy with restrictive policies",
						Description: "An allow-all AuthorizationPolicy was detected. This should be replaced with namespace or service-specific policies.",
						Category:    "security",
						Severity:    "high",
						Fix: FixRecommendation{
							Type:     "istiopolicy",
							Template: generateRestrictiveAuthzPolicy(),
							ManualSteps: []string{
								"1. Identify which services need to communicate",
								"2. Create namespace-specific AuthorizationPolicies",
								"3. Remove the allow-all policy",
								"4. Test connectivity after changes",
							},
						},
						Impact: "Restricts service-to-service communication to only necessary paths, reducing attack surface",
					})
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
					Title:       "Consider routing egress through Istio EgressGateway",
					Description: "Services are accessing external resources directly. Routing through EgressGateway provides better control, monitoring, and policy enforcement.",
					Category:    "security",
					Severity:    "medium",
					Fix: FixRecommendation{
						Type:     "egress",
						Template: generateEgressGatewayConfig(),
						ManualSteps: []string{
							"1. Deploy Istio EgressGateway (if not already deployed)",
							"2. Create ServiceEntry for external services that need access",
							"3. Create VirtualService to route traffic through gateway",
							"4. Update DestinationRule for egress traffic",
							"5. Test external connectivity",
						},
					},
					Impact: "Provides centralized control, monitoring, and policy enforcement for external traffic",
				})
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "cilium-cnp-001",
		Name:        "Use Cilium Network Policies for Advanced Features",
		Description: "If Cilium is the CNI, consider using CiliumNetworkPolicies for advanced L7 features",
		Category:    "security",
		Severity:    "medium",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Infrastructure.CiliumEnabled {
				// Check if they're using Cilium policies
				ciliumPolicyCount := len(topology.CiliumPolicies)
				k8sPolicyCount := len(topology.NetworkPolicies)

				// If they have Cilium but are only using K8s NetworkPolicies, recommend CNPs
				if ciliumPolicyCount == 0 && k8sPolicyCount > 0 {
					recommendations = append(recommendations, Recommendation{
						ID:          "cilium-cnp-001",
						Title:       "Consider using CiliumNetworkPolicies for advanced features",
						Description: "Cilium is detected but only K8s NetworkPolicies are in use. CiliumNetworkPolicies offer L7 filtering, DNS-based rules, and better performance.",
						Category:    "security",
						Severity:    "medium",
						Fix: FixRecommendation{
							Type:     "ciliumpolicy",
							Template: generateCiliumPolicyExample(),
							ManualSteps: []string{
								"1. Review existing K8s NetworkPolicies",
								"2. Convert to CiliumNetworkPolicy for L7 features",
								"3. Test connectivity after migration",
								"4. Consider using CiliumClusterwideNetworkPolicy for cluster-wide rules",
							},
						},
						Impact: "Enables L7-aware policies, DNS-based rules, and better observability with Hubble",
					})
				}
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "kyverno-001",
		Name:        "Use Kyverno for Policy-as-Code",
		Description: "If Kyverno is installed, use it to enforce network policy standards automatically",
		Category:    "security",
		Severity:    "medium",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			if topology.Infrastructure.KyvernoEnabled {
				kyvernoPolicyCount := len(topology.KyvernoPolicies)

				if kyvernoPolicyCount == 0 {
					recommendations = append(recommendations, Recommendation{
						ID:          "kyverno-001",
						Title:       "Use Kyverno to enforce network policy standards",
						Description: "Kyverno is installed but no policies detected. Use Kyverno ClusterPolicies to automatically require NetworkPolicies on new namespaces.",
						Category:    "security",
						Severity:    "medium",
						Fix: FixRecommendation{
							Type:     "kyverno",
							Template: generateKyvernoNetworkPolicyPolicy(),
							ManualSteps: []string{
								"1. Create Kyverno ClusterPolicy to require NetworkPolicies",
								"2. Create policy to validate NetworkPolicy structure",
								"3. Test by creating a namespace without NetworkPolicy (should be blocked)",
							},
						},
						Impact: "Automatically enforces network policy requirements, preventing services from being deployed without proper network isolation",
					})
				}
			}

			return len(recommendations) == 0, recommendations
		},
	},
	{
		ID:          "mesh-001",
		Name:        "Service Mesh Coverage",
		Description: "If Istio is enabled, most services should be in the mesh for observability and security",
		Category:    "observability",
		Severity:    "medium",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			// Only check if Istio is enabled
			if topology.Infrastructure.IstioEnabled && topology.Summary.TotalServices > 0 {
				coveragePercent := (topology.Summary.ServicesWithMesh * 100) / topology.Summary.TotalServices
				if coveragePercent < 80 {
					recommendations = append(recommendations, Recommendation{
						ID:          "mesh-001",
						Title:       "Low Istio service mesh coverage",
						Description: fmt.Sprintf("Only %d%% of services are in the Istio mesh (target: 80%%). Services not in the mesh miss mTLS, observability, and traffic management.", coveragePercent),
						Category:    "observability",
						Severity:    "medium",
						Fix: FixRecommendation{
							Type:     "istio",
							Template: generateIstioInjectionConfig(),
							ManualSteps: []string{
								"1. Enable Istio sidecar injection for namespaces",
								"2. Label namespaces: kubectl label namespace <ns> istio-injection=enabled",
								"3. Or use annotation: sidecar.istio.io/inject: \"true\" in pod spec",
								"4. Restart pods to inject sidecars",
								"5. Verify sidecars are injected: kubectl get pods -n <ns> -o jsonpath='{.items[*].spec.containers[*].name}'",
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
		Description: "Blocked connections should be investigated and resolved with specific policy fixes",
		Category:    "resilience",
		Severity:    "high",
		Check: func(ctx context.Context, topology *TopologyData) (bool, []Recommendation) {
			var recommendations []Recommendation

			// Analyze actual blocked connections and provide specific fixes
			blockedConnections := make(map[string][]ServiceConnection) // service -> blocked connections

			for serviceKey, connectivity := range topology.Connectivity {
				for _, conn := range connectivity.Connections {
					if conn.BlockedByPolicy || !conn.Allowed {
						if blockedConnections[serviceKey] == nil {
							blockedConnections[serviceKey] = []ServiceConnection{}
						}
						blockedConnections[serviceKey] = append(blockedConnections[serviceKey], conn)
					}
				}
			}

			// Generate specific recommendations for each blocked connection
			for serviceKey, blocked := range blockedConnections {
				if len(blocked) == 0 {
					continue
				}

				serviceNs, serviceName, _ := ParseServiceKey(serviceKey)
				service, serviceExists := topology.Services[serviceKey]
				if !serviceExists {
					continue
				}

				// Group blocked connections by target
				blockedByTarget := make(map[string][]ServiceConnection)
				for _, conn := range blocked {
					blockedByTarget[conn.Target] = append(blockedByTarget[conn.Target], conn)
				}

				// Generate specific recommendation for each blocked target
				for target, conns := range blockedByTarget {
					targetNs, targetName, _ := ParseServiceKey(target)

					// Build specific fix based on blocking policies
					blockingPoliciesList := []string{}
					for _, conn := range conns {
						blockingPoliciesList = append(blockingPoliciesList, conn.BlockingPolicies...)
					}

					// Remove duplicates
					uniquePolicies := make(map[string]bool)
					for _, policy := range blockingPoliciesList {
						uniquePolicies[policy] = true
					}

					policyList := []string{}
					for policy := range uniquePolicies {
						policyList = append(policyList, policy)
					}

					ports := []string{}
					for _, conn := range conns {
						if conn.Port != "" {
							ports = append(ports, conn.Port)
						}
					}

					description := fmt.Sprintf("Connection from %s/%s to %s is blocked", serviceNs, serviceName, target)
					if len(policyList) > 0 {
						description += fmt.Sprintf(" by policy(ies): %s", strings.Join(policyList, ", "))
					}
					if len(ports) > 0 {
						description += fmt.Sprintf(" on port(s): %s", strings.Join(ports, ", "))
					}

					// Generate specific policy fix
					var fixTemplate string
					var fixType string

					if topology.Infrastructure.CiliumEnabled {
						fixType = "ciliumpolicy"
						fixTemplate = generateCiliumPolicyForBlockedConnection(service, targetNs, targetName, ports, policyList)
					} else {
						fixType = "networkpolicy"
						fixTemplate = generateK8sPolicyForBlockedConnection(service, targetNs, targetName, ports, policyList)
					}

					recommendations = append(recommendations, Recommendation{
						ID:          fmt.Sprintf("blocked-001-%s-to-%s", serviceKey, target),
						Title:       fmt.Sprintf("Blocked connection: %s/%s → %s", serviceNs, serviceName, target),
						Description: description,
						Category:    "resilience",
						Severity:    "high",
						Service:     serviceKey,
						Namespace:   serviceNs,
						Fix: FixRecommendation{
							Type:     fixType,
							Template: fixTemplate,
							Command:  fmt.Sprintf("kubectl apply -f - <<EOF\n%s\nEOF", fixTemplate),
							ManualSteps: []string{
								fmt.Sprintf("1. Review blocking policy(ies): %s", strings.Join(policyList, ", ")),
								fmt.Sprintf("2. Update policy to allow connection from %s/%s to %s", serviceNs, serviceName, target),
								"3. Verify connection works after policy update",
								"4. Use path tracer to verify end-to-end connectivity",
							},
						},
						Impact: fmt.Sprintf("Restores connectivity between %s/%s and %s, improving application functionality", serviceNs, serviceName, target),
					})
				}
			}

			// Limit to top 10 most critical blocked connections
			if len(recommendations) > 10 {
				recommendations = recommendations[:10]
			}

			return len(recommendations) == 0, recommendations
		},
	},
}

// Helper function to generate K8s NetworkPolicy template based on actual connections
func generateK8sNetworkPolicy(service ServiceInfo, topology *TopologyData) string {
	serviceKey := ServiceKey(service.Namespace, service.Name)
	connectivity, hasConnections := topology.Connectivity[serviceKey]

	// Check if Istio is enabled to adjust recommendations
	istioNote := ""
	if topology.Infrastructure.IstioEnabled {
		istioNote = "# Note: This works alongside Istio AuthorizationPolicies for defense in depth\n"
	}

	// Build ingress rules based on actual connections
	ingressRules := ""
	ingressSources := make(map[string]map[string]bool) // namespace -> port -> true

	// Check if service receives traffic from ingress
	hasIngress := false
	if topology.Ingress.KubernetesIngress != nil {
		for _, ing := range topology.Ingress.KubernetesIngress {
			if ing.Backend == service.Name && ing.Namespace == service.Namespace {
				hasIngress = true
				break
			}
		}
	}

	if hasIngress {
		ingressRules += `  # Allow from ingress controller
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
      podSelector:
        matchLabels:
          app.kubernetes.io/name: ingress-nginx
    ports:
    - protocol: TCP
      port: 80
    - protocol: TCP
      port: 443
`
	}

	// Analyze actual connections to build specific ingress rules
	if hasConnections {
		for _, conn := range connectivity.Connections {
			if conn.Allowed && !conn.BlockedByPolicy {
				// Parse target to get namespace/name
				targetNs, targetName, _ := ParseServiceKey(conn.Target)
				if targetNs == service.Namespace && targetName == service.Name {
					// This is an incoming connection - find the source
					// We need to find which service is connecting TO this one
					// For now, allow same namespace communication
					if ingressSources[service.Namespace] == nil {
						ingressSources[service.Namespace] = make(map[string]bool)
					}
					if conn.Port != "" {
						ingressSources[service.Namespace][conn.Port] = true
					}
				}
			}
		}

		// Add specific ingress rules for actual connections
		if len(ingressSources) > 0 {
			ingressRules += `  # Allow from same namespace (based on actual connections)
  - from:
    - podSelector: {}
`
			// Add port-specific rules if we have port info
			ports := []string{}
			for port := range ingressSources[service.Namespace] {
				ports = append(ports, port)
			}
			if len(ports) > 0 {
				ingressRules += `    ports:
`
				for _, port := range ports {
					ingressRules += fmt.Sprintf(`    - protocol: TCP
      port: %s
`, port)
				}
			}
		}
	} else {
		// Fallback: allow same namespace if no connection data
		ingressRules += `  # Allow inter-pod communication within namespace
  - from:
    - podSelector: {}
`
	}

	// Build egress rules based on actual connections
	egressRules := `  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
`

	// Analyze actual egress connections
	if hasConnections {
		egressTargets := make(map[string]map[string]bool) // target -> port -> true
		for _, conn := range connectivity.Connections {
			if conn.Allowed && !conn.BlockedByPolicy {
				// This service is connecting TO the target
				targetNs, targetName, _ := ParseServiceKey(conn.Target)
				if targetNs != service.Namespace || targetName != service.Name {
					// This is an egress connection
					targetKey := conn.Target
					if egressTargets[targetKey] == nil {
						egressTargets[targetKey] = make(map[string]bool)
					}
					if conn.Port != "" {
						egressTargets[targetKey][conn.Port] = true
					} else {
						egressTargets[targetKey]["*"] = true
					}
				}
			}
		}

		// Generate specific egress rules for actual connections
		if len(egressTargets) > 0 {
			egressRules += `  # Allow egress to connected services (based on actual traffic)
`
			for target, ports := range egressTargets {
				targetNs, targetName, hasNs := ParseServiceKey(target)
				if hasNs && targetNs != service.Namespace {
					egressRules += fmt.Sprintf(`  - to:
    - namespaceSelector:
        matchLabels:
          name: %s
    ports:
`, targetNs)
					for port := range ports {
						if port != "*" {
							egressRules += fmt.Sprintf(`    - protocol: TCP
      port: %s
`, port)
						} else {
							egressRules += `    - protocol: TCP
`
						}
					}
				} else if targetNs == service.Namespace {
					// Same namespace
					egressRules += fmt.Sprintf(`  - to:
    - podSelector:
        matchLabels:
          app: %s
    ports:
`, targetName)
					for port := range ports {
						if port != "*" {
							egressRules += fmt.Sprintf(`    - protocol: TCP
      port: %s
`, port)
						}
					}
				}
			}
		} else {
			// Fallback: allow same namespace egress
			egressRules += `  # Allow egress to same namespace
  - to:
    - namespaceSelector:
        matchLabels:
          name: ` + service.Namespace + `
`
		}
	} else {
		// Fallback: allow same namespace egress
		egressRules += `  # Allow egress to same namespace
  - to:
    - namespaceSelector:
        matchLabels:
          name: ` + service.Namespace + `
`
	}

	return fmt.Sprintf(`%sapiVersion: networking.k8s.io/v1
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
%s
egress:
%s`, istioNote, service.Name, service.Namespace, service.Name, ingressRules, egressRules)
}

// Helper function to generate Cilium NetworkPolicy template based on actual connections
func generateCiliumNetworkPolicy(service ServiceInfo, topology *TopologyData) string {
	serviceKey := ServiceKey(service.Namespace, service.Name)
	connectivity, hasConnections := topology.Connectivity[serviceKey]

	// Build ingress rules based on actual connections
	ingressRules := ""

	// Check if service receives traffic from ingress
	hasIngress := false
	if topology.Ingress.KubernetesIngress != nil {
		for _, ing := range topology.Ingress.KubernetesIngress {
			if ing.Backend == service.Name && ing.Namespace == service.Namespace {
				hasIngress = true
				break
			}
		}
	}

	if hasIngress {
		ingressRules += `  # Allow from ingress controller
  - fromEndpoints:
    - matchLabels:
        app.kubernetes.io/name: ingress-nginx
    toPorts:
    - ports:
      - port: "80"
        protocol: TCP
      - port: "443"
        protocol: TCP
`
	}

	// Analyze actual connections for ingress
	if hasConnections {
		ingressPorts := make(map[string]bool)
		for _, conn := range connectivity.Connections {
			if conn.Allowed && !conn.BlockedByPolicy {
				targetNs, targetName, _ := ParseServiceKey(conn.Target)
				if targetNs == service.Namespace && targetName == service.Name {
					if conn.Port != "" {
						ingressPorts[conn.Port] = true
					}
				}
			}
		}

		if len(ingressPorts) > 0 || len(connectivity.Connections) > 0 {
			ingressRules += `  # Allow from same namespace (based on actual connections)
  - fromEndpoints:
    - matchLabels:
        app: ` + service.Name + `
    toPorts:
`
			if len(ingressPorts) > 0 {
				for port := range ingressPorts {
					ingressRules += fmt.Sprintf(`    - ports:
      - port: "%s"
        protocol: TCP
`, port)
				}
			} else {
				ingressRules += `    - ports:
      - port: "80"
        protocol: TCP
`
			}
		}
	} else {
		ingressRules += `  # Allow inter-pod communication
  - fromEndpoints:
    - matchLabels:
        app: ` + service.Name + `
`
	}

	// Build egress rules based on actual connections
	egressRules := `  # Allow DNS
  - toEndpoints:
    - matchLabels:
        k8s-app: kube-dns
    toPorts:
    - ports:
      - port: "53"
        protocol: UDP
      - port: "53"
        protocol: TCP
`

	// Analyze actual egress connections
	if hasConnections {
		egressTargets := make(map[string]map[string]bool)
		for _, conn := range connectivity.Connections {
			if conn.Allowed && !conn.BlockedByPolicy {
				targetNs, targetName, _ := ParseServiceKey(conn.Target)
				if targetNs != service.Namespace || targetName != service.Name {
					targetKey := conn.Target
					if egressTargets[targetKey] == nil {
						egressTargets[targetKey] = make(map[string]bool)
					}
					if conn.Port != "" {
						egressTargets[targetKey][conn.Port] = true
					}
				}
			}
		}

		if len(egressTargets) > 0 {
			egressRules += `  # Allow egress to connected services (based on actual traffic)
`
			for target, ports := range egressTargets {
				_, targetName, hasNs := ParseServiceKey(target)
				if hasNs {
					egressRules += fmt.Sprintf(`  - toEndpoints:
    - matchLabels:
        app: %s
    toPorts:
`, targetName)
					for port := range ports {
						egressRules += fmt.Sprintf(`    - ports:
      - port: "%s"
        protocol: TCP
`, port)
					}
				}
			}
		}
	}

	return fmt.Sprintf(`apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: %s-cnp
  namespace: %s
spec:
  endpointSelector:
    matchLabels:
      app: %s
ingress:
%s
egress:
%s`, service.Name, service.Namespace, service.Name, ingressRules, egressRules)
}

// Helper function to generate TLS ingress template (with cert-manager)
func generateTLSIngress(ing KubernetesIngressInfo, hasCertManager bool) string {
	annotation := ""
	if hasCertManager {
		annotation = `  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    # Or use: cert-manager.io/issuer: letsencrypt-prod (for namespace-scoped issuer)
`
	}

	hosts := strings.Join(ing.Hosts, "\n    - ")
	if hosts == "" {
		hosts = "example.com" // fallback
	}

	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s
  namespace: %s
%sspec:
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
`, ing.Name, ing.Namespace, annotation, hosts, ing.Name, ing.Hosts[0], ing.Backend)
}

// Helper function to generate STRICT mTLS policy
func generateStrictMTLSPolicy() string {
	return `apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: istio-system
spec:
  mtls:
    mode: STRICT
---
# For specific namespace (start with PERMISSIVE, then migrate to STRICT)
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: <namespace>
spec:
  mtls:
    mode: PERMISSIVE  # Change to STRICT after testing
`
}

// Helper function to generate restrictive AuthorizationPolicy
func generateRestrictiveAuthzPolicy() string {
	return `apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-namespace-communication
  namespace: <namespace>
spec:
  action: ALLOW
  rules:
  # Allow from same namespace
  - from:
    - source:
        namespaces: ["<namespace>"]
  # Allow from ingress gateway
  - from:
    - source:
        principals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
  # Allow to kube-system (for DNS, metrics, etc.)
  - to:
    - operation:
        hosts: ["*.kube-system.svc.cluster.local"]
`
}

// Helper function to generate EgressGateway config
func generateEgressGatewayConfig() string {
	return `# 1. ServiceEntry for external service
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: external-api
spec:
  hosts:
  - api.example.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  resolution: DNS
  location: MESH_EXTERNAL
---
# 2. VirtualService to route through gateway
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: route-external-api
spec:
  hosts:
  - api.example.com
  gateways:
  - istio-egressgateway
  - mesh
  http:
  - match:
    - gateways:
      - mesh
      port: 443
    route:
    - destination:
        host: istio-egressgateway.istio-system.svc.cluster.local
        port:
          number: 443
      weight: 100
  - match:
    - gateways:
      - istio-egressgateway
      port: 443
    route:
    - destination:
        host: api.example.com
        port:
          number: 443
      weight: 100
`
}

// Helper function to generate Cilium policy example
func generateCiliumPolicyExample() string {
	return `apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: example-l7-policy
  namespace: <namespace>
spec:
  endpointSelector:
    matchLabels:
      app: <app-name>
  ingress:
  # L7 HTTP policy example
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
      rules:
        http:
        - method: "GET"
          path: "/api/v1/*"
  egress:
  # DNS-based egress policy
  - toFQDNs:
    - matchName: "api.example.com"
    toPorts:
    - ports:
      - port: "443"
        protocol: TCP
`
}

// Helper function to generate Kyverno policy
func generateKyvernoNetworkPolicyPolicy() string {
	return `apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-network-policy
  annotations:
    policies.kyverno.io/title: Require NetworkPolicy
    policies.kyverno.io/category: Security
    policies.kyverno.io/severity: high
    policies.kyverno.io/subject: NetworkPolicy
    policies.kyverno.io/description: >-
      Require that all namespaces have at least one NetworkPolicy
spec:
  validationFailureAction: audit
  background: true
  rules:
  - name: check-network-policy
    match:
      resources:
        kinds:
        - Namespace
    validate:
      message: "Namespace must have at least one NetworkPolicy"
      deny:
        conditions:
        - key: "{{count(NetworkPolicy)}}"
          operator: LessThan
          value: 1
`
}

// Helper function to generate Istio injection config
func generateIstioInjectionConfig() string {
	return `# Enable Istio sidecar injection for a namespace
kubectl label namespace <namespace> istio-injection=enabled

# Or use pod-level annotation (in deployment spec)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <app-name>
spec:
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "true"
    spec:
      containers:
      - name: <container-name>
        image: <image>
`
}

// Helper function to generate K8s NetworkPolicy for a specific blocked connection
func generateK8sPolicyForBlockedConnection(source ServiceInfo, targetNs, targetName string, ports []string, blockingPolicies []string) string {
	portRules := ""
	if len(ports) > 0 {
		portRules = "    ports:\n"
		for _, port := range ports {
			portRules += fmt.Sprintf(`    - protocol: TCP
      port: %s
`, port)
		}
	} else {
		portRules = "    # Ports: Add specific ports based on service requirements\n"
	}

	policyNote := ""
	if len(blockingPolicies) > 0 {
		policyNote = fmt.Sprintf("# This policy allows connection blocked by: %s\n", strings.Join(blockingPolicies, ", "))
	}

	if targetNs == source.Namespace {
		// Same namespace
		return fmt.Sprintf(`%sapiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-%s-to-%s
  namespace: %s
spec:
  podSelector:
    matchLabels:
      app: %s
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: %s
%s`, policyNote, source.Name, targetName, source.Namespace, targetName, source.Name, portRules)
	} else {
		// Cross-namespace
		return fmt.Sprintf(`%sapiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-%s-to-%s
  namespace: %s
spec:
  podSelector:
    matchLabels:
      app: %s
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: %s
      podSelector:
        matchLabels:
          app: %s
%s`, policyNote, source.Name, targetName, targetNs, targetName, source.Namespace, source.Name, portRules)
	}
}

// Helper function to generate Cilium NetworkPolicy for a specific blocked connection
func generateCiliumPolicyForBlockedConnection(source ServiceInfo, targetNs, targetName string, ports []string, blockingPolicies []string) string {
	portRules := ""
	if len(ports) > 0 {
		portRules = "    toPorts:\n    - ports:\n"
		for _, port := range ports {
			portRules += fmt.Sprintf(`      - port: "%s"
        protocol: TCP
`, port)
		}
	} else {
		portRules = "    # Ports: Add specific ports based on service requirements\n"
	}

	policyNote := ""
	if len(blockingPolicies) > 0 {
		policyNote = fmt.Sprintf("# This policy allows connection blocked by: %s\n", strings.Join(blockingPolicies, ", "))
	}

	if targetNs == source.Namespace {
		// Same namespace
		return fmt.Sprintf(`%sapiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-%s-to-%s
  namespace: %s
spec:
  endpointSelector:
    matchLabels:
      app: %s
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: %s
%s`, policyNote, source.Name, targetName, source.Namespace, targetName, source.Name, portRules)
	} else {
		// Cross-namespace
		return fmt.Sprintf(`%sapiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-%s-to-%s
  namespace: %s
spec:
  endpointSelector:
    matchLabels:
      app: %s
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: %s
      matchLabels:
        kubernetes.io/metadata.name: %s
%s`, policyNote, source.Name, targetName, targetNs, targetName, source.Name, source.Namespace, portRules)
	}
}
