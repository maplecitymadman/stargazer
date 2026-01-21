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

func TestScanPods(t *testing.T) {
	// Pod in CrashLoop
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "crashing-pod", Namespace: "default"},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "main",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "backoff"},
					},
					RestartCount: 10,
				},
			},
		},
	}

	clientset := fake.NewSimpleClientset(pod)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	discovery := NewDiscovery(client)
	issues, err := discovery.ScanPods(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}

	foundCrash := false
	for _, issue := range issues {
		if issue.ResourceName == "crashing-pod" && issue.Priority == PriorityCritical {
			foundCrash = true
		}
	}

	if !foundCrash {
		t.Error("Did not find critical issue for crashing pod")
	}
}

func TestScanDeployments(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "unhealthy-deploy", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "unhealthy"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "unhealthy"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "main", Image: "nginx"}},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     1,
			AvailableReplicas: 1,
		},
	}

	clientset := fake.NewSimpleClientset(deploy)
	client := &Client{
		clientset: clientset,
		cache: &cache{
			data:     make(map[string]cacheEntry),
			cacheTTL: 1 * time.Minute,
		},
	}

	discovery := NewDiscovery(client)
	issues, err := discovery.ScanDeployments(context.Background(), "default")
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) == 0 {
		t.Error("Expected issues for replica mismatch")
	}
}
