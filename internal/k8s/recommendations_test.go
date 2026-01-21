package k8s

import (
	"context"
	"strings"
	"testing"
)

func TestGetRecommendations(t *testing.T) {
	client := &Client{}

	// Create topology data with a service that has no policy
	topology := &TopologyData{
		Services: map[string]ServiceInfo{
			"default/app": {
				Name:      "app",
				Namespace: "default",
			},
		},
		NetworkPolicies: []NetworkPolicyInfo{},
		Infrastructure: InfrastructureInfo{
			CiliumEnabled: false,
			IstioEnabled:  false,
		},
		Connectivity: map[string]ConnectivityInfo{
			"default/app": {
				Service: "app",
				Connections: []ServiceConnection{
					{Target: "db", Allowed: true},
				},
			},
		},
	}

	recs, err := client.GetRecommendations(context.Background(), topology)
	if err != nil {
		t.Fatal(err)
	}

	foundPolicyRec := false
	for _, rec := range recs {
		if strings.HasPrefix(rec.ID, "np-001") {
			foundPolicyRec = true
		}
	}

	if !foundPolicyRec {
		t.Error("Expected recommendation for missing network policy (np-001)")
	}
}

func TestGetComplianceScore(t *testing.T) {
	client := &Client{}

	// Poor compliance topology
	topology := &TopologyData{
		Services: map[string]ServiceInfo{
			"default/app": {Name: "app", Namespace: "default"},
		},
		NetworkPolicies: []NetworkPolicyInfo{},
	}

	score, details := client.GetComplianceScore(context.Background(), topology)
	if score > 90 {
		t.Errorf("Expected lower compliance score, got %d", score)
	}
	if details == nil {
		t.Fatal("Expected details map")
	}
}
