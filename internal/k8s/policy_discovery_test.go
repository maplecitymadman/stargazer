package k8s

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetTopologyWithPolicies(t *testing.T) {
	scheme := runtime.NewScheme()

	// Create a Cilium policy
	ciliumPolicy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cilium.io/v2",
			"kind":       "CiliumNetworkPolicy",
			"metadata": map[string]interface{}{
				"name":      "test-cilium",
				"namespace": "default",
			},
			"spec": map[string]interface{}{},
		},
	}

	// Create a Kyverno policy
	kyvernoPolicy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "ClusterPolicy",
			"metadata": map[string]interface{}{
				"name": "test-kyverno",
			},
			"spec": map[string]interface{}{},
		},
	}

	// Create ArgoCD Application (to prevent drift detector panic)
	argoApp := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "test-app",
				"namespace": "argocd",
			},
			"status": map[string]interface{}{
				"sync": map[string]interface{}{
					"status": "Synced",
				},
			},
		},
	}

	// Initialize dynamic client WITH the objects
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, ciliumPolicy, kyvernoPolicy, argoApp)

	client := &Client{
		clientset:     fake.NewSimpleClientset(),
		dynamicClient: dynamicClient,
		cache:         &cache{data: make(map[string]cacheEntry)},
	}

	topology, err := client.GetTopology(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}

	if topology == nil {
		t.Fatal("Expected topology, got nil")
	}
}

func TestGetTopologyWithIstio(t *testing.T) {
	scheme := runtime.NewScheme()

	// Create Istio Gateway
	gateway := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1beta1",
			"kind":       "Gateway",
			"metadata": map[string]interface{}{
				"name":      "test-gateway",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{"istio": "ingressgateway"},
			},
		},
	}

	// Create Istio VirtualService
	vs := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1beta1",
			"kind":       "VirtualService",
			"metadata": map[string]interface{}{
				"name":      "test-vs",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"gateways": []interface{}{"test-gateway"},
				"hosts":    []interface{}{"*"},
				"http": []interface{}{
					map[string]interface{}{
						"route": []interface{}{
							map[string]interface{}{
								"destination": map[string]interface{}{
									"host": "test-service",
								},
							},
						},
					},
				},
			},
		},
	}

	// Create ArgoCD Application (to prevent drift detector panic)
	argoApp := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "test-app",
				"namespace": "argocd",
			},
		},
	}

	// Initialize dynamic client WITH the objects
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, gateway, vs, argoApp)

	client := &Client{
		clientset:     fake.NewSimpleClientset(),
		dynamicClient: dynamicClient,
		cache:         &cache{data: make(map[string]cacheEntry)},
	}

	topology, err := client.GetTopology(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}

	if topology == nil {
		t.Fatal("Expected topology, got nil")
	}
}
