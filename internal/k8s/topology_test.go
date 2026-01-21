package k8s

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestServiceKey(t *testing.T) {
	tests := []struct {
		namespace string
		name      string
		expected  string
	}{
		{"default", "svc1", "default/svc1"},
		{"", "svc2", "svc2"},
	}

	for _, tt := range tests {
		if got := ServiceKey(tt.namespace, tt.name); got != tt.expected {
			t.Errorf("ServiceKey(%q, %q) = %q, want %q", tt.namespace, tt.name, got, tt.expected)
		}
	}
}

func TestParseServiceKey(t *testing.T) {
	tests := []struct {
		key       string
		wantNS    string
		wantName  string
		wantHasNS bool
	}{
		{"default/svc1", "default", "svc1", true},
		{"svc2", "", "svc2", false},
	}

	for _, tt := range tests {
		ns, name, hasNS := ParseServiceKey(tt.key)
		if ns != tt.wantNS || name != tt.wantName || hasNS != tt.wantHasNS {
			t.Errorf("ParseServiceKey(%q) = (%q, %q, %v), want (%q, %q, %v)", tt.key, ns, name, hasNS, tt.wantNS, tt.wantName, tt.wantHasNS)
		}
	}
}

func TestGetTopologySimple(t *testing.T) {
	// Create common mock objects
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "test"},
			Ports: []corev1.ServicePort{
				{Port: 80, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	clientset := fake.NewSimpleClientset(svc, pod)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	topology, err := client.GetTopology(context.Background(), "default")
	if err != nil {
		t.Fatalf("GetTopology failed: %v", err)
	}

	if len(topology.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(topology.Services))
	}

	sInfo, ok := topology.Services["default/test-svc"]
	if !ok {
		t.Fatal("Service not found in topology map")
	}

	if sInfo.PodCount != 1 {
		t.Errorf("Expected 1 pod, got %d", sInfo.PodCount)
	}
}

func TestGetTopologyWithConnectivity(t *testing.T) {
	// Source Service & Pod
	srcSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "src-svc", Namespace: "default", Labels: map[string]string{"app": "src"}},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "src"}, Ports: []corev1.ServicePort{{Port: 80}}},
	}
	srcPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "src-pod", Namespace: "default", Labels: map[string]string{"app": "src"}},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}

	// Target Service & Pod
	targetSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "target-svc", Namespace: "default", Labels: map[string]string{"app": "target"}},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "target"}, Ports: []corev1.ServicePort{{Port: 80}}},
	}
	targetPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "target-pod", Namespace: "default", Labels: map[string]string{"app": "target"}},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}

	// NetworkPolicy allowing src to target
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "allow-src-to-target", Namespace: "default"},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "target"}},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "src"}}},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}

	clientset := fake.NewSimpleClientset(srcSvc, srcPod, targetSvc, targetPod, policy)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	topology, err := client.GetTopology(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}

	// Check connectivity Info
	connInfo, ok := topology.Connectivity["default/src-svc"]
	if !ok {
		t.Fatal("Source service connectivity info missing")
	}

	foundTarget := false
	for _, conn := range connInfo.Connections {
		if conn.Target == "target-svc" {
			foundTarget = true
			if !conn.Allowed {
				t.Error("Expected connection to target-svc to be allowed by policy")
			}
		}
	}
	if !foundTarget {
		t.Error("Connection to target-svc not found in connectivity map")
	}
}

func TestDetectInfrastructure(t *testing.T) {
	// Cilium DaemonSet
	ciliumDS := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "cilium", Namespace: "kube-system"},
	}

	// Istio Namespace
	istioNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "istio-system"},
	}

	// Istiod Deployment
	istiod := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "istiod", Namespace: "istio-system"},
	}

	clientset := fake.NewSimpleClientset(ciliumDS, istioNS, istiod)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	infra, err := client.detectInfrastructure(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if infra.CNI != "cilium" {
		t.Errorf("Expected CNI cilium, got %s", infra.CNI)
	}
	if !infra.IstioEnabled {
		t.Error("Expected IstioEnabled to be true")
	}
}
