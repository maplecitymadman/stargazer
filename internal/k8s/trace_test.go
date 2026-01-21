package k8s

import (
	"context"
	"testing"
)

func TestTracePath(t *testing.T) {
	client := &Client{}

	topology := &TopologyData{
		Services: map[string]ServiceInfo{
			"default/src": {Name: "src", Namespace: "default"},
			"default/dst": {Name: "dst", Namespace: "default"},
		},
		Connectivity: map[string]ConnectivityInfo{
			"default/src": {
				Connections: []ServiceConnection{
					{Target: "default/dst", Allowed: true},
				},
			},
		},
	}

	// Trace from src to dst
	trace, err := client.TracePath(context.Background(), "default/src", "default/dst", "default", topology)
	if err != nil {
		t.Fatal(err)
	}

	if !trace.Allowed {
		t.Error("Expected trace allowed")
	}

	if len(trace.Path) == 0 {
		t.Error("Expected at least one hop in path")
	}
}

func TestTraceFromIngress(t *testing.T) {
	client := &Client{}

	topology := &TopologyData{
		Ingress: IngressInfo{
			Connections: []IngressConnection{
				{To: "default/app", Allowed: true},
			},
		},
		Connectivity: map[string]ConnectivityInfo{
			"default/app": {
				Connections: []ServiceConnection{
					{Target: "default/db", Allowed: true},
				},
			},
		},
		Services: map[string]ServiceInfo{
			"default/app": {Name: "app", Namespace: "default"},
			"default/db":  {Name: "db", Namespace: "default"},
		},
	}

	// Trace from ingress to db (multi-hop)
	trace, err := client.TracePath(context.Background(), "ingress-gateway", "default/db", "default", topology)
	if err != nil {
		t.Fatal(err)
	}

	if !trace.Allowed {
		t.Errorf("Expected trace allowed, got reason: %s", trace.Reason)
	}
}

func TestTraceToEgress(t *testing.T) {
	client := &Client{}

	topology := &TopologyData{
		Egress: EgressInfo{
			Connections: []EgressConnection{
				{From: "default/app", To: "external", Allowed: true},
			},
		},
		Services: map[string]ServiceInfo{
			"default/app": {Name: "app", Namespace: "default"},
		},
	}

	trace, err := client.TracePath(context.Background(), "default/app", "egress-gateway", "default", topology)
	if err != nil {
		t.Fatal(err)
	}

	if !trace.Allowed {
		t.Error("Expected trace allowed")
	}
}

func TestTraceIngressToService(t *testing.T) {
	client := &Client{}

	topology := &TopologyData{
		Ingress: IngressInfo{
			Connections: []IngressConnection{
				{To: "default/app", Allowed: true},
			},
		},
	}

	// Trace from ingress to app (direct)
	trace, err := client.TracePath(context.Background(), "ingress-gateway", "default/app", "default", topology)
	if err != nil {
		t.Fatal(err)
	}

	if !trace.Allowed {
		t.Error("Expected trace allowed")
	}
}
