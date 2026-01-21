package k8s

import (
	"context"
	"fmt"
	"strings"
)

// TracePath traces a complete path from source to destination
func (c *Client) TracePath(ctx context.Context,
	source string,
	destination string,
	namespace string,
	topology *TopologyData) (*PathTrace, error) {

	if topology == nil {
		return &PathTrace{
			Source:      source,
			Destination: destination,
			Path:        []PathHop{},
			Allowed:     false,
			Reason:      "Topology data not available",
		}, fmt.Errorf("topology data is nil")
	}

	trace := &PathTrace{
		Source:      source,
		Destination: destination,
		Path:        []PathHop{},
		Allowed:     true,
	}

	// Parse source
	sourceType, sourceService := c.parseEndpoint(source)
	destType, destService := c.parseEndpoint(destination)

	// Case 1: Ingress → ... (Generic ingress)
	if sourceType == "ingress" {
		return c.traceFromIngress(ctx, destType, destService, namespace, topology)
	}

	// Case 3: Service → Service
	if sourceType == "service" && destType == "service" {
		return c.traceServiceToService(ctx, sourceService, destService, namespace, topology)
	}

	return trace, nil
}

// parseEndpoint parses an endpoint string into type and service
func (c *Client) parseEndpoint(endpoint string) (endpointType string, service string) {
	if endpoint == "ingress-gateway" {
		return "ingress", ""
	}
	if endpoint == "egress-gateway" || endpoint == "external" {
		return "egress", ""
	}
	// Format: "namespace/service" or just "service"
	parts := strings.Split(endpoint, "/")
	if len(parts) == 2 {
		return "service", endpoint
	}
	if len(parts) > 0 {
		return "service", endpoint
	}
	return "service", ""
}

// traceFromIngress traces path from ingress to destination
func (c *Client) traceFromIngress(ctx context.Context,
	destType string, destService string, namespace string, topology *TopologyData) (*PathTrace, error) {

	trace := &PathTrace{
		Source:      "ingress-gateway",
		Destination: destService,
		Path:        []PathHop{},
		Allowed:     true,
	}

	// Step 1: Ingress → First Service
	firstHop := c.findFirstServiceFromIngress(topology)
	if firstHop == nil {
		trace.Allowed = false
		trace.Reason = "No route from ingress gateway"
		return trace, nil
	}

	trace.Path = append(trace.Path, *firstHop)
	if !firstHop.Allowed {
		trace.Allowed = false
		trace.BlockedAt = firstHop
		trace.Reason = fmt.Sprintf("Blocked at ingress: %s", firstHop.Reason)
		return trace, nil
	}

	trace.Path = append(trace.Path, *firstHop)
	if !firstHop.Allowed {
		trace.Allowed = false
		trace.BlockedAt = firstHop
		trace.Reason = fmt.Sprintf("Blocked at ingress: %s", firstHop.Reason)
		return trace, nil
	}

	// If the first hop IS the destination, we're done
	if destType == "service" && (firstHop.To == destService || strings.Contains(firstHop.To, destService)) {
		return trace, nil
	}

	// Step 2: Continue from first service
	if destType == "egress" {
		// Trace to egress
		egressTrace, err := c.traceToEgress(ctx, firstHop.To, namespace, topology)
		if err == nil {
			if !egressTrace.Allowed {
				trace.Allowed = false
				trace.BlockedAt = egressTrace.BlockedAt
				trace.Reason = egressTrace.Reason
			}
			trace.Path = append(trace.Path, egressTrace.Path...)
		}
	} else if destType == "service" {
		// Trace to another service
		serviceTrace, err := c.traceServiceToService(ctx, firstHop.To, destService, namespace, topology)
		if err == nil {
			if !serviceTrace.Allowed {
				trace.Allowed = false
				trace.BlockedAt = serviceTrace.BlockedAt
				trace.Reason = serviceTrace.Reason
			}
			trace.Path = append(trace.Path, serviceTrace.Path...)
		}
	}

	return trace, nil
}

// traceToEgress traces path from service to egress
func (c *Client) traceToEgress(ctx context.Context,
	serviceName string, namespace string, topology *TopologyData) (*PathTrace, error) {

	trace := &PathTrace{
		Source:      serviceName,
		Destination: "egress-gateway",
		Path:        []PathHop{},
		Allowed:     true,
	}

	if topology == nil {
		trace.Allowed = false
		trace.Reason = "Topology data not available"
		return trace, fmt.Errorf("topology data is nil")
	}

	// Find egress connection for this service
	for _, conn := range topology.Egress.Connections {
		if conn.From == serviceName {
			hop := PathHop{
				From:     serviceName,
				To:       conn.To,
				Type:     "egress",
				Allowed:  conn.Allowed,
				Reason:   conn.Reason,
				Policies: conn.Policies,
			}
			trace.Path = append(trace.Path, hop)

			if !conn.Allowed {
				trace.Allowed = false
				trace.BlockedAt = &hop
				trace.Reason = fmt.Sprintf("Blocked at egress: %s", conn.Reason)
			}
			return trace, nil
		}
	}

	// No egress connection found
	trace.Allowed = false
	trace.Reason = "No egress route configured for service"
	return trace, nil
}

// traceServiceToService traces path between two services
func (c *Client) traceServiceToService(ctx context.Context,
	fromService string, toService string, namespace string, topology *TopologyData) (*PathTrace, error) {

	trace := &PathTrace{
		Source:      fromService,
		Destination: toService,
		Path:        []PathHop{},
		Allowed:     true,
	}

	// Find connection in connectivity map
	fromKey := fromService
	if !strings.Contains(fromService, "/") && namespace != "" {
		fromKey = fmt.Sprintf("%s/%s", namespace, fromService)
	}

	connInfo, exists := topology.Connectivity[fromKey]
	if !exists {
		trace.Allowed = false
		trace.Reason = "Source service not found"
		return trace, nil
	}

	// Find connection to destination
	for _, conn := range connInfo.Connections {
		if conn.Target == toService || strings.Contains(conn.Target, toService) {
			hop := PathHop{
				From:        fromService,
				To:          toService,
				Type:        "service",
				Allowed:     conn.Allowed,
				Reason:      conn.Reason,
				Policies:    conn.BlockingPolicies,
				ServiceMesh: conn.ServiceMeshType,
			}
			trace.Path = append(trace.Path, hop)

			if !conn.Allowed {
				trace.Allowed = false
				trace.BlockedAt = &hop
				trace.Reason = fmt.Sprintf("Blocked: %s", conn.Reason)
			}
			return trace, nil
		}
	}

	// No direct connection - might need intermediate hops
	// For now, mark as not found
	trace.Allowed = false
	trace.Reason = "No connection path found"
	return trace, nil
}

// findFirstServiceFromIngress finds the first service reachable from ingress
func (c *Client) findFirstServiceFromIngress(topology *TopologyData) *PathHop {
	if topology == nil {
		return nil
	}
	if len(topology.Ingress.Connections) == 0 {
		return nil
	}

	// Get first allowed connection
	for _, conn := range topology.Ingress.Connections {
		if conn.Allowed {
			return &PathHop{
				From:     "ingress-gateway",
				To:       conn.To,
				Type:     "ingress",
				Allowed:  true,
				Reason:   conn.Reason,
				Policies: conn.Policies,
			}
		}
	}

	// Return first connection even if blocked
	if len(topology.Ingress.Connections) > 0 {
		conn := topology.Ingress.Connections[0]
		return &PathHop{
			From:     "ingress-gateway",
			To:       conn.To,
			Type:     "ingress",
			Allowed:  false,
			Reason:   conn.Reason,
			Policies: conn.Policies,
		}
	}

	return nil
}
