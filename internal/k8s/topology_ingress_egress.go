package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// getIngressInfo detects and maps ingress gateways and routes
func (c *Client) getIngressInfo(ctx context.Context, namespace string, services map[string]ServiceInfo, infra InfrastructureInfo,
	networkPolicies []NetworkPolicyInfo,
	ciliumPolicies []CiliumNetworkPolicyInfo,
	istioPolicies []IstioPolicyInfo) (IngressInfo, error) {
	ingress := IngressInfo{
		Gateways:          []GatewayInfo{},
		KubernetesIngress: []KubernetesIngressInfo{},
		Routes:            []IngressRoute{},
		Connections:       []IngressConnection{},
	}

	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	// 1. Detect Istio Gateways
	if infra.IstioEnabled && c.dynamicClient != nil {
		istioVersion, err := c.detectIstioAPIVersion(ctx)
		if err == nil {
			gateways, routes, err := c.getIstioIngress(ctx, ns, istioVersion, services)
			if err == nil {
				ingress.Gateways = append(ingress.Gateways, gateways...)
				ingress.Routes = append(ingress.Routes, routes...)
			}
		}
	}

	// 2. Detect Ingress-Nginx (Kubernetes Ingress)
	if c.clientset != nil {
		ingressList, err := c.clientset.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, ing := range ingressList.Items {
				ingressClass := ""
				if ing.Spec.IngressClassName != nil {
					ingressClass = *ing.Spec.IngressClassName
				}

				// Check if it's nginx or istio
				isNginx := strings.Contains(ingressClass, "nginx") || ingressClass == ""
				isIstio := strings.Contains(ingressClass, "istio")

				if isNginx || isIstio {
					hosts := []string{}
					paths := []string{}
					backend := ""
					backendPort := ""

					for _, rule := range ing.Spec.Rules {
						hosts = append(hosts, rule.Host)
						if rule.HTTP != nil {
							for _, path := range rule.HTTP.Paths {
								paths = append(paths, path.Path)
								if path.Backend.Service != nil {
									backend = path.Backend.Service.Name
									if path.Backend.Service.Port.Number > 0 {
										backendPort = fmt.Sprintf("%d", path.Backend.Service.Port.Number)
									}
								}
							}
						}
					}

					ingressType := "nginx"
					if isIstio {
						ingressType = "istio"
					}

					ingress.KubernetesIngress = append(ingress.KubernetesIngress, KubernetesIngressInfo{
						Name:        ing.Name,
						Namespace:   ing.Namespace,
						Hosts:       hosts,
						Paths:       paths,
						Backend:     backend,
						BackendPort: backendPort,
						TLS:         len(ing.Spec.TLS) > 0,
						Class:       ingressClass,
					})

					// Build routes
					for _, rule := range ing.Spec.Rules {
						if rule.HTTP != nil {
							for _, path := range rule.HTTP.Paths {
								if path.Backend.Service != nil {
									route := IngressRoute{
										Gateway:   ing.Name,
										Host:      rule.Host,
										Path:      path.Path,
										Service:   path.Backend.Service.Name,
										Namespace: ing.Namespace,
										Allowed:   true,
										Type:      ingressType,
									}
									ingress.Routes = append(ingress.Routes, route)
								}
							}
						}
					}
				}
			}
		}
	}

	// 3. Build Ingress → Service connections
	ingress.Connections = c.buildIngressConnections(ingress, services, networkPolicies, ciliumPolicies, istioPolicies)
	// #region agent log
	func() {
		logData := map[string]interface{}{
			"totalConnections": len(ingress.Connections),
			"allowedConnections": func() int {
				count := 0
				for _, c := range ingress.Connections {
					if c.Allowed {
						count++
					}
				}
				return count
			}(),
			"blockedConnections": func() int {
				count := 0
				for _, c := range ingress.Connections {
					if !c.Allowed {
						count++
					}
				}
				return count
			}(),
		}
		logLine := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"topology_ingress_egress.go:120","message":"Ingress connections built","data":%s,"timestamp":%d}`, toJSON(logData), time.Now().UnixMilli())
		f, _ := os.OpenFile("/Users/isaac.sanchezhawkins/talos-deploy/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString(logLine + "\n")
		f.Close()
	}()
	// #endregion
	return ingress, nil
}

// getIstioIngress retrieves Istio Gateways and VirtualServices
func (c *Client) getIstioIngress(ctx context.Context, namespace string, apiVersion string, services map[string]ServiceInfo) ([]GatewayInfo, []IngressRoute, error) {
	gateways := []GatewayInfo{}
	routes := []IngressRoute{}

	gatewayGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  apiVersion,
		Resource: "gateways",
	}

	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  apiVersion,
		Resource: "virtualservices",
	}

	// Get Gateways
	var gatewayList *unstructured.UnstructuredList
	var err error
	if namespace == "" {
		gatewayList, err = c.dynamicClient.Resource(gatewayGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		gatewayList, err = c.dynamicClient.Resource(gatewayGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		for _, item := range gatewayList.Items {
			spec, ok := item.Object["spec"].(map[string]interface{})
			if !ok {
				continue
			}

			hosts := []string{}
			ports := []string{}
			selector := map[string]string{}

			if servers, ok := spec["servers"].([]interface{}); ok {
				for _, server := range servers {
					if s, ok := server.(map[string]interface{}); ok {
						if h, ok := s["hosts"].([]interface{}); ok {
							for _, host := range h {
								if hostStr, ok := host.(string); ok {
									hosts = append(hosts, hostStr)
								}
							}
						}
						if p, ok := s["port"].(map[string]interface{}); ok {
							if num, ok := p["number"].(int64); ok {
								if proto, ok := p["protocol"].(string); ok {
									ports = append(ports, fmt.Sprintf("%d/%s", num, proto))
								}
							}
						}
					}
				}
			}

			if sel, ok := spec["selector"].(map[string]interface{}); ok {
				for k, v := range sel {
					if vStr, ok := v.(string); ok {
						selector[k] = vStr
					}
				}
			}

			gateways = append(gateways, GatewayInfo{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Type:      "istio",
				Hosts:     hosts,
				Ports:     ports,
				Selector:  selector,
			})
		}
	}

	// Get VirtualServices (routes)
	var vsList *unstructured.UnstructuredList
	if namespace == "" {
		vsList, err = c.dynamicClient.Resource(virtualServiceGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		vsList, err = c.dynamicClient.Resource(virtualServiceGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		for _, item := range vsList.Items {
			spec, ok := item.Object["spec"].(map[string]interface{})
			if !ok {
				continue
			}

			// Extract hosts
			hosts := []string{}
			if h, ok := spec["hosts"].([]interface{}); ok {
				for _, host := range h {
					if hostStr, ok := host.(string); ok {
						hosts = append(hosts, hostStr)
					}
				}
			}

			// Extract routes (HTTP routes)
			if http, ok := spec["http"].([]interface{}); ok {
				for _, route := range http {
					if r, ok := route.(map[string]interface{}); ok {
						path := "/"
						if match, ok := r["match"].([]interface{}); ok && len(match) > 0 {
							if m, ok := match[0].(map[string]interface{}); ok {
								if p, ok := m["uri"].(map[string]interface{}); ok {
									if prefix, ok := p["prefix"].(string); ok {
										path = prefix
									}
								}
							}
						}

						// Extract destination service
						if dest, ok := r["route"].([]interface{}); ok && len(dest) > 0 {
							if d, ok := dest[0].(map[string]interface{}); ok {
								if svc, ok := d["destination"].(map[string]interface{}); ok {
									serviceName := ""
									serviceNS := item.GetNamespace()
									if host, ok := svc["host"].(string); ok {
										// host format: service.namespace.svc.cluster.local or just service
										parts := strings.Split(host, ".")
										if len(parts) > 0 {
											serviceName = parts[0]
											if len(parts) > 1 && parts[1] != "svc" {
												serviceNS = parts[1]
											}
										}

										if serviceName != "" {
											for _, host := range hosts {
												route := IngressRoute{
													Gateway:   item.GetName(),
													Host:      host,
													Path:      path,
													Service:   serviceName,
													Namespace: serviceNS,
													Allowed:   true,
													Type:      "istio",
												}
												routes = append(routes, route)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return gateways, routes, nil
}

// buildIngressConnections builds Ingress → Service connections with policy evaluation
func (c *Client) buildIngressConnections(ingress IngressInfo, services map[string]ServiceInfo,
	networkPolicies []NetworkPolicyInfo,
	ciliumPolicies []CiliumNetworkPolicyInfo,
	istioPolicies []IstioPolicyInfo) []IngressConnection {

	connections := []IngressConnection{}
	ctx := context.Background()

	// From routes
	for _, route := range ingress.Routes {
		serviceKey := fmt.Sprintf("%s/%s", route.Namespace, route.Service)
		if service, exists := services[serviceKey]; exists {
			// Evaluate policies
			allowed, reason, blockingPolicies := c.evaluateIngressPolicy(
				ctx, route.Gateway, service, networkPolicies, ciliumPolicies, istioPolicies)

			conn := IngressConnection{
				From:     route.Gateway,
				To:       route.Service,
				Allowed:  allowed,
				Reason:   reason,
				Policies: blockingPolicies,
			}
			connections = append(connections, conn)
		}
	}

	// From Kubernetes Ingress
	for _, ing := range ingress.KubernetesIngress {
		if ing.Backend != "" {
			serviceKey := fmt.Sprintf("%s/%s", ing.Namespace, ing.Backend)
			if service, exists := services[serviceKey]; exists {
				// Evaluate policies
				allowed, reason, blockingPolicies := c.evaluateIngressPolicy(
					ctx, ing.Name, service, networkPolicies, ciliumPolicies, istioPolicies)

				conn := IngressConnection{
					From:     ing.Name,
					To:       ing.Backend,
					Allowed:  allowed,
					Reason:   reason,
					Policies: blockingPolicies,
					Port:     ing.BackendPort,
				}
				connections = append(connections, conn)
			}
		}
	}

	return connections
}

// evaluateIngressPolicy checks if ingress → service connection is allowed
func (c *Client) evaluateIngressPolicy(ctx context.Context,
	from string,
	toService ServiceInfo,
	networkPolicies []NetworkPolicyInfo,
	ciliumPolicies []CiliumNetworkPolicyInfo,
	istioPolicies []IstioPolicyInfo) (allowed bool, reason string, blockingPolicies []string) {
	// #region agent log
	func() {
		logData := map[string]interface{}{
			"from":                 from,
			"toService":            toService.Name,
			"toNamespace":          toService.Namespace,
			"networkPoliciesCount": len(networkPolicies),
			"ciliumPoliciesCount":  len(ciliumPolicies),
			"istioPoliciesCount":   len(istioPolicies),
		}
		logLine := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"topology_ingress_egress.go:373","message":"evaluateIngressPolicy entry","data":%s,"timestamp":%d}`, toJSON(logData), time.Now().UnixMilli())
		f, _ := os.OpenFile("/Users/isaac.sanchezhawkins/talos-deploy/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString(logLine + "\n")
		f.Close()
	}()
	// #endregion
	allowed = true
	reason = "No policy blocking ingress"
	blockingPolicies = []string{}

	// Check NetworkPolicies that might block ingress
	for _, policy := range networkPolicies {
		if policy.Namespace == toService.Namespace {
			// Simplified: if policy exists, it might block
			// In reality, would need to evaluate policy rules
			allowed = false
			reason = fmt.Sprintf("NetworkPolicy %s may block ingress", policy.Name)
			blockingPolicies = append(blockingPolicies, policy.Name)
		}
	}

	// Check Cilium policies
	for _, policy := range ciliumPolicies {
		if policy.Namespace == toService.Namespace || policy.Namespace == "" {
			// Cluster-wide or namespace-specific
			allowed = false
			reason = fmt.Sprintf("CiliumNetworkPolicy %s may block ingress", policy.Name)
			blockingPolicies = append(blockingPolicies, policy.Name)
		}
	}

	// Check Istio AuthorizationPolicy
	for _, policy := range istioPolicies {
		if policy.Type == "authorizationpolicy" && policy.Namespace == toService.Namespace {
			// AuthorizationPolicy might restrict ingress
			allowed = false
			reason = fmt.Sprintf("AuthorizationPolicy %s may restrict ingress", policy.Name)
			blockingPolicies = append(blockingPolicies, policy.Name)
		}
	}
	// #region agent log
	func() {
		logData := map[string]interface{}{
			"allowed":               allowed,
			"reason":                reason,
			"blockingPoliciesCount": len(blockingPolicies),
		}
		logLine := fmt.Sprintf(`{"sessionId":"debug-session","runId":"run1","hypothesisId":"D","location":"topology_ingress_egress.go:424","message":"evaluateIngressPolicy result","data":%s,"timestamp":%d}`, toJSON(logData), time.Now().UnixMilli())
		f, _ := os.OpenFile("/Users/isaac.sanchezhawkins/talos-deploy/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString(logLine + "\n")
		f.Close()
	}()
	// #endregion
	return allowed, reason, blockingPolicies
}

// getEgressInfo detects and maps egress gateways and external access
func (c *Client) getEgressInfo(ctx context.Context, namespace string, services map[string]ServiceInfo, infra InfrastructureInfo,
	networkPolicies []NetworkPolicyInfo,
	ciliumPolicies []CiliumNetworkPolicyInfo,
	istioPolicies []IstioPolicyInfo) (EgressInfo, error) {
	egress := EgressInfo{
		Gateways:         []GatewayInfo{},
		ExternalServices: []ExternalServiceInfo{},
		Connections:      []EgressConnection{},
		HasEgressGateway: false,
		DirectEgress:     true,
	}

	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	// 1. Detect Istio EgressGateway
	if infra.IstioEnabled && c.dynamicClient != nil {
		istioVersion, err := c.detectIstioAPIVersion(ctx)
		if err == nil {
			egressGateways, serviceEntries, err := c.getIstioEgress(ctx, ns, istioVersion)
			if err == nil {
				egress.Gateways = append(egress.Gateways, egressGateways...)
				egress.ExternalServices = append(egress.ExternalServices, serviceEntries...)
				egress.HasEgressGateway = len(egressGateways) > 0
			}
		}
	}

	// 2. Detect direct egress (services accessing external without gateway)
	// This is detected by checking for services with external endpoints
	// or by analyzing network policies that allow egress

	// 3. Build Service → Egress connections
	egress.Connections = c.buildEgressConnections(egress, services, infra, networkPolicies, ciliumPolicies, istioPolicies)

	return egress, nil
}

// getIstioEgress retrieves Istio EgressGateways and ServiceEntry
func (c *Client) getIstioEgress(ctx context.Context, namespace string, apiVersion string) ([]GatewayInfo, []ExternalServiceInfo, error) {
	gateways := []GatewayInfo{}
	externalServices := []ExternalServiceInfo{}

	serviceEntryGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  apiVersion,
		Resource: "serviceentries",
	}

	var seList *unstructured.UnstructuredList
	var err error
	if namespace == "" {
		seList, err = c.dynamicClient.Resource(serviceEntryGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		seList, err = c.dynamicClient.Resource(serviceEntryGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		for _, item := range seList.Items {
			spec, ok := item.Object["spec"].(map[string]interface{})
			if !ok {
				continue
			}

			hosts := []string{}
			ports := []string{}

			if h, ok := spec["hosts"].([]interface{}); ok {
				for _, host := range h {
					if hostStr, ok := host.(string); ok {
						hosts = append(hosts, hostStr)
					}
				}
			}

			if p, ok := spec["ports"].([]interface{}); ok {
				for _, port := range p {
					if portMap, ok := port.(map[string]interface{}); ok {
						if num, ok := portMap["number"].(int64); ok {
							if proto, ok := portMap["protocol"].(string); ok {
								ports = append(ports, fmt.Sprintf("%d/%s", num, proto))
							}
						}
					}
				}
			}

			externalServices = append(externalServices, ExternalServiceInfo{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Hosts:     hosts,
				Ports:     ports,
				Type:      "serviceentry",
			})
		}
	}

	// Check for EgressGateway deployment (not a CRD, but a Deployment)
	// This indicates egress gateway is configured
	if c.clientset != nil {
		deployments, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
			LabelSelector: "app=istio-egressgateway",
		})
		if err == nil && len(deployments.Items) > 0 {
			// Add gateway info
			for _, dep := range deployments.Items {
				gateways = append(gateways, GatewayInfo{
					Name:      dep.Name,
					Namespace: dep.Namespace,
					Type:      "istio-egress",
				})
			}
		}
	}

	return gateways, externalServices, nil
}

// buildEgressConnections builds Service → Egress connections with policy evaluation
func (c *Client) buildEgressConnections(egress EgressInfo, services map[string]ServiceInfo,
	infra InfrastructureInfo,
	networkPolicies []NetworkPolicyInfo,
	ciliumPolicies []CiliumNetworkPolicyInfo,
	istioPolicies []IstioPolicyInfo) []EgressConnection {

	connections := []EgressConnection{}
	ctx := context.Background()

	// Services → External Services (ServiceEntry)
	for _, extSvc := range egress.ExternalServices {
		// Find services that might access this external service
		for _, service := range services {
			// Evaluate policies
			allowed, reason, blockingPolicies := c.evaluateEgressPolicy(
				ctx, service, strings.Join(extSvc.Hosts, ","),
				networkPolicies, ciliumPolicies, istioPolicies)

			conn := EgressConnection{
				From:     service.Name,
				To:       strings.Join(extSvc.Hosts, ","),
				Allowed:  allowed,
				Reason:   reason,
				Policies: blockingPolicies,
			}
			connections = append(connections, conn)
		}
	}

	// If no egress gateway, mark as direct egress
	if !egress.HasEgressGateway {
		for _, service := range services {
			// Evaluate policies for direct egress
			allowed, reason, blockingPolicies := c.evaluateEgressPolicy(
				ctx, service, "external", networkPolicies, ciliumPolicies, istioPolicies)

			conn := EgressConnection{
				From:     service.Name,
				To:       "external",
				Allowed:  allowed,
				Reason:   reason,
				Policies: blockingPolicies,
			}
			connections = append(connections, conn)
		}
	}

	return connections
}

// evaluateEgressPolicy checks if service → egress connection is allowed
func (c *Client) evaluateEgressPolicy(ctx context.Context,
	fromService ServiceInfo,
	to string,
	networkPolicies []NetworkPolicyInfo,
	ciliumPolicies []CiliumNetworkPolicyInfo,
	istioPolicies []IstioPolicyInfo) (allowed bool, reason string, blockingPolicies []string) {

	allowed = true
	reason = "No policy blocking egress"
	blockingPolicies = []string{}

	// Check NetworkPolicies that might block egress
	for _, policy := range networkPolicies {
		if policy.Namespace == fromService.Namespace {
			// Check if policy has egress rules
			// Simplified check - would need actual policy evaluation
			allowed = false
			reason = fmt.Sprintf("NetworkPolicy %s may block egress", policy.Name)
			blockingPolicies = append(blockingPolicies, policy.Name)
		}
	}

	// Check Cilium policies
	for _, policy := range ciliumPolicies {
		if policy.Namespace == fromService.Namespace || policy.Namespace == "" {
			allowed = false
			reason = fmt.Sprintf("CiliumNetworkPolicy %s may block egress", policy.Name)
			blockingPolicies = append(blockingPolicies, policy.Name)
		}
	}

	// Check Istio AuthorizationPolicy for egress
	for _, policy := range istioPolicies {
		if policy.Type == "authorizationpolicy" && policy.Namespace == fromService.Namespace {
			allowed = false
			reason = fmt.Sprintf("AuthorizationPolicy %s may restrict egress", policy.Name)
			blockingPolicies = append(blockingPolicies, policy.Name)
		}
	}

	return allowed, reason, blockingPolicies
}
