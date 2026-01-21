package k8s

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetDeployments(t *testing.T) {
	// Create fake clientset
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-deploy",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
			Labels:            map[string]string{"app": "test"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx:latest"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     2,
			AvailableReplicas: 2,
		},
	}

	clientset := fake.NewSimpleClientset(deployment)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	ctx := context.Background()
	deployments, err := client.GetDeployments(ctx, "default")
	if err != nil {
		t.Fatalf("GetDeployments failed: %v", err)
	}

	if len(deployments) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(deployments))
	}

	d := deployments[0]
	if d.Name != "test-deploy" {
		t.Errorf("Expected name test-deploy, got %s", d.Name)
	}
	if d.Replicas != 3 {
		t.Errorf("Expected 3 replicas, got %d", d.Replicas)
	}
	if d.ReadyReplicas != 2 {
		t.Errorf("Expected 2 ready replicas, got %d", d.ReadyReplicas)
	}
	if len(d.Images) != 1 || d.Images[0] != "nginx:latest" {
		t.Errorf("Unexpected images: %v", d.Images)
	}

	// Test cache
	deployments2, _ := client.GetDeployments(ctx, "default")
	if len(deployments2) != 1 {
		t.Error("Cache failed to return deployment")
	}
}

func TestGetServices(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "10.0.0.1",
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
			},
		},
	}

	clientset := fake.NewSimpleClientset(service)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	services, err := client.GetServices(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
	if services[0].ClusterIP != "10.0.0.1" {
		t.Errorf("Expected cluster IP 10.0.0.1, got %s", services[0].ClusterIP)
	}
}

func TestGetNodes(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"node-role.kubernetes.io/worker": "",
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion: "v1.25.0",
			},
		},
	}

	clientset := fake.NewSimpleClientset(node)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	nodes, err := client.GetNodes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Status != "Ready" {
		t.Errorf("Expected status Ready, got %s", nodes[0].Status)
	}
}

func TestGetNamespaces(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}

	clientset := fake.NewSimpleClientset(ns)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	namespaces, err := client.GetNamespaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(namespaces) < 1 {
		t.Error("Expected at least 1 namespace")
	}
}

func int32Ptr(i int32) *int32 { return &i }
