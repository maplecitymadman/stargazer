package k8s

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildCiliumNetworkPolicy(t *testing.T) {
	pb := NewPolicyBuilder(fake.NewSimpleClientset(), nil)

	spec := CiliumNetworkPolicySpec{
		Name:      "test-policy",
		Namespace: "default",
		EndpointSelector: map[string]interface{}{
			"matchLabels": map[string]string{"app": "test"},
		},
		Ingress: []CiliumIngressRule{
			{
				FromEndpoints: []map[string]interface{}{
					{"matchLabels": map[string]string{"app": "allowed"}},
				},
				ToPorts: []CiliumPortRule{
					{
						Ports: []CiliumPortProtocol{
							{Port: "80", Protocol: "TCP"},
						},
					},
				},
			},
		},
	}

	yamlContent, err := pb.BuildCiliumNetworkPolicy(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}

	if yamlContent == "" {
		t.Error("Expected non-empty YAML content")
	}
}

func TestBuildKyvernoPolicy(t *testing.T) {
	pb := NewPolicyBuilder(fake.NewSimpleClientset(), nil)

	spec := KyvernoPolicySpec{
		Name: "require-labels",
		Type: "ClusterPolicy",
		Rules: []KyvernoRule{
			{
				Name: "check-labels",
				MatchResources: map[string]interface{}{
					"resources": map[string]interface{}{
						"kinds": []string{"Pod"},
					},
				},
				Validate: map[string]interface{}{
					"message": "Labels are required",
					"pattern": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "?*",
							},
						},
					},
				},
			},
		},
	}

	yamlContent, err := pb.BuildKyvernoPolicy(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}

	if yamlContent == "" {
		t.Error("Expected non-empty YAML content")
	}
}

func TestApplyCiliumNetworkPolicy(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	pb := NewPolicyBuilder(fake.NewSimpleClientset(), dynamicClient)

	yamlContent := `
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: test-policy
  namespace: default
spec:
  endpointSelector:
    matchLabels:
      app: test
`

	err := pb.ApplyCiliumNetworkPolicy(context.Background(), yamlContent, "default")
	if err != nil {
		t.Fatal(err)
	}
}

func TestApplyKyvernoPolicy(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	pb := NewPolicyBuilder(fake.NewSimpleClientset(), dynamicClient)

	yamlContent := `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-labels
spec:
  rules:
  - name: check-labels
    match:
      resources:
        kinds:
        - Pod
`

	err := pb.ApplyKyvernoPolicy(context.Background(), yamlContent, "")
	if err != nil {
		t.Fatal(err)
	}

	// Test namespace-scoped policy
	yamlContentNS := `
apiVersion: kyverno.io/v1
kind: Policy
metadata:
  name: ns-policy
  namespace: default
spec:
  rules: []
`
	err = pb.ApplyKyvernoPolicy(context.Background(), yamlContentNS, "default")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeletePolicies(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	pb := NewPolicyBuilder(fake.NewSimpleClientset(), dynamicClient)

	// Create Cilium Policy
	ciliumYAML := `
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: test
  namespace: default
spec:
  endpointSelector: {}
`
	err := pb.ApplyCiliumNetworkPolicy(context.Background(), ciliumYAML, "default")
	if err != nil {
		t.Fatal(err)
	}

	err = pb.DeleteCiliumNetworkPolicy(context.Background(), "test", "default")
	if err != nil {
		t.Fatal(err)
	}

	// Create Kyverno Policy
	kyvernoYAML := `
apiVersion: kyverno.io/v1
kind: Policy
metadata:
  name: test
  namespace: default
spec:
  rules: []
`
	err = pb.ApplyKyvernoPolicy(context.Background(), kyvernoYAML, "default")
	if err != nil {
		t.Fatal(err)
	}

	err = pb.DeleteKyvernoPolicy(context.Background(), "test", "default", false)
	if err != nil {
		t.Fatal(err)
	}

	// Create Kyverno ClusterPolicy
	clusterYAML := `
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: test-cluster
spec:
  rules: []
`
	err = pb.ApplyKyvernoPolicy(context.Background(), clusterYAML, "")
	if err != nil {
		t.Fatal(err)
	}

	err = pb.DeleteKyvernoPolicy(context.Background(), "test-cluster", "", true)
	if err != nil {
		t.Fatal(err)
	}
}
