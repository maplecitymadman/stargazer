package k8s

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetTopologyWithPolicies(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)

	// Create a Cilium policy
	ciliumGVR := schema.GroupVersionResource{Group: "cilium.io", Version: "v2", Resource: "ciliumnetworkpolicies"}
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
	dynamicClient.Resource(ciliumGVR).Namespace("default").Create(context.Background(), ciliumPolicy, metav1.CreateOptions{})

	// Create a Kyverno policy
	kyvernoGVR := schema.GroupVersionResource{Group: "kyverno.io", Version: "v1", Resource: "clusterpolicies"}
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
	dynamicClient.Resource(kyvernoGVR).Create(context.Background(), kyvernoPolicy, metav1.CreateOptions{})

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
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)

	// Create Istio Gateway
	gatewayGVR := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "gateways"}
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
	dynamicClient.Resource(gatewayGVR).Namespace("default").Create(context.Background(), gateway, metav1.CreateOptions{})

	// Create Istio VirtualService
	vsGVR := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "virtualservices"}
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
	dynamicClient.Resource(vsGVR).Namespace("default").Create(context.Background(), vs, metav1.CreateOptions{})

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
