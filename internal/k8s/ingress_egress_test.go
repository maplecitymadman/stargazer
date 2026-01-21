package k8s

import (
	"context"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetIngressInfo(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	client := &Client{
		clientset: clientset,
		cache:     &cache{data: make(map[string]cacheEntry)},
	}

	// Create a dummy ingress
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ing",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "test.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	clientset.NetworkingV1().Ingresses("default").Create(context.Background(), ing, metav1.CreateOptions{})

	services := map[string]ServiceInfo{
		"default/test-service": {Name: "test-service", Namespace: "default"},
	}
	infra := InfrastructureInfo{}
	netPol := []NetworkPolicyInfo{}
	cilPol := []CiliumNetworkPolicyInfo{}
	istPol := []IstioPolicyInfo{}

	info, err := client.getIngressInfo(context.Background(), "default", services, infra, netPol, cilPol, istPol)
	if err != nil {
		t.Fatal(err)
	}

	if len(info.KubernetesIngress) == 0 {
		t.Error("Expected at least one Kubernetes ingress")
	}

	if len(info.Routes) == 0 {
		t.Error("Expected at least one route")
	}

	if len(info.Connections) == 0 {
		t.Error("Expected at least one connection built")
	}
}

func TestGetEgressInfo(t *testing.T) {
	client := &Client{}

	services := map[string]ServiceInfo{
		"default/app": {Name: "app", Namespace: "default"},
	}
	infra := InfrastructureInfo{}
	netPol := []NetworkPolicyInfo{}
	cilPol := []CiliumNetworkPolicyInfo{}
	istPol := []IstioPolicyInfo{}

	info, err := client.getEgressInfo(context.Background(), "default", services, infra, netPol, cilPol, istPol)
	if err != nil {
		t.Fatal(err)
	}

	// buildEgressConnections should create a connection for each service if no gateway exists
	if len(info.Connections) == 0 {
		t.Error("Expected at least one egress connection for services")
	}
}

func TestEvaluateIngressPolicy(t *testing.T) {
	client := &Client{}

	service := ServiceInfo{Name: "app", Namespace: "default"}
	netPol := []NetworkPolicyInfo{
		{Name: "block-all", Namespace: "default"},
	}
	cilPol := []CiliumNetworkPolicyInfo{}
	istPol := []IstioPolicyInfo{}

	allowed, _, policies := client.evaluateIngressPolicy(context.Background(), "ingress", service, netPol, cilPol, istPol)

	if allowed {
		t.Error("Expected ingress to be blocked by policy")
	}
	if len(policies) == 0 {
		t.Error("Expected blocking policy to be listed")
	}
}

func TestEvaluateIngressPolicyWithCilium(t *testing.T) {
	client := &Client{}

	service := ServiceInfo{Name: "app", Namespace: "default"}
	netPol := []NetworkPolicyInfo{}
	cilPol := []CiliumNetworkPolicyInfo{
		{Name: "cilium-block", Namespace: "default"},
	}
	istPol := []IstioPolicyInfo{}

	allowed, _, policies := client.evaluateIngressPolicy(context.Background(), "ingress", service, netPol, cilPol, istPol)

	if allowed {
		t.Error("Expected ingress to be blocked by Cilium policy")
	}
	if len(policies) == 0 {
		t.Error("Expected blocking Cilium policy to be listed")
	}
}

func TestEvaluateIngressPolicyWithIstio(t *testing.T) {
	client := &Client{}

	service := ServiceInfo{Name: "app", Namespace: "default"}
	netPol := []NetworkPolicyInfo{}
	cilPol := []CiliumNetworkPolicyInfo{}
	istPol := []IstioPolicyInfo{
		{Name: "istio-block", Namespace: "default", Type: "authorizationpolicy"},
	}

	allowed, _, policies := client.evaluateIngressPolicy(context.Background(), "ingress", service, netPol, cilPol, istPol)

	if allowed {
		t.Error("Expected ingress to be blocked by Istio policy")
	}
	if len(policies) == 0 {
		t.Error("Expected blocking Istio policy to be listed")
	}
}
