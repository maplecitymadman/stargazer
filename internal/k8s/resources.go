package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Kubernetes deployment
type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	Labels            map[string]string `json:"labels"`
	Selector          map[string]string `json:"selector"`
	Images            []string          `json:"images"`
	Age               string            `json:"age"`
}

// Service represents a Kubernetes service
type Service struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"cluster_ip"`
	ExternalIP  string            `json:"external_ip,omitempty"`
	Ports       []ServicePort     `json:"ports"`
	Selector    map[string]string `json:"selector"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Age         string            `json:"age"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	NodePort   int32  `json:"node_port,omitempty"`
}

// Node represents a Kubernetes node
type Node struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"` // Ready, NotReady
	Roles             []string          `json:"roles"`
	Version           string            `json:"version"`
	OS                string            `json:"os"`
	ContainerRuntime  string            `json:"container_runtime"`
	CPUCapacity       string            `json:"cpu_capacity"`
	MemoryCapacity    string            `json:"memory_capacity"`
	PodCapacity       string            `json:"pod_capacity"`
	CPUAllocatable    string            `json:"cpu_allocatable"`
	MemoryAllocatable string            `json:"memory_allocatable"`
	PodAllocatable    string            `json:"pod_allocatable"`
	Labels            map[string]string `json:"labels"`
	Age               string            `json:"age"`
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels"`
	Age    string            `json:"age"`
}

// GetDeployments retrieves deployments from the specified namespace
func (c *Client) GetDeployments(ctx context.Context, namespace string) ([]Deployment, error) {
	var cacheKey string
	var ns string

	if namespace == "" || namespace == "all" {
		cacheKey = "deployments-all"
		ns = ""
	} else {
		cacheKey = fmt.Sprintf("deployments-%s", namespace)
		ns = namespace
	}

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if deployments, ok := cached.([]Deployment); ok {
			return deployments, nil
		}
	}

	// Query Kubernetes API
	var deploymentList *appsv1.DeploymentList
	var err error

	if ns == "" {
		deploymentList, err = c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	} else {
		deploymentList, err = c.clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	deployments := make([]Deployment, 0, len(deploymentList.Items))
	for _, d := range deploymentList.Items {
		deployment := convertDeployment(&d)
		deployments = append(deployments, deployment)
	}

	c.cache.set(cacheKey, deployments)
	return deployments, nil
}

// GetServices retrieves services from the specified namespace
func (c *Client) GetServices(ctx context.Context, namespace string) ([]Service, error) {
	var cacheKey string
	var ns string

	if namespace == "" || namespace == "all" {
		cacheKey = "services-all"
		ns = ""
	} else {
		cacheKey = fmt.Sprintf("services-%s", namespace)
		ns = namespace
	}

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if services, ok := cached.([]Service); ok {
			return services, nil
		}
	}

	// Query Kubernetes API
	var serviceList *corev1.ServiceList
	var err error

	if ns == "" {
		serviceList, err = c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	} else {
		serviceList, err = c.clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make([]Service, 0, len(serviceList.Items))
	for _, s := range serviceList.Items {
		service := convertService(&s)
		services = append(services, service)
	}

	c.cache.set(cacheKey, services)
	return services, nil
}

// GetNodes retrieves all nodes in the cluster
func (c *Client) GetNodes(ctx context.Context) ([]Node, error) {
	cacheKey := "nodes-all"

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if nodes, ok := cached.([]Node); ok {
			return nodes, nil
		}
	}

	// Query Kubernetes API
	nodeList, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]Node, 0, len(nodeList.Items))
	for _, n := range nodeList.Items {
		node := convertNode(&n)
		nodes = append(nodes, node)
	}

	c.cache.set(cacheKey, nodes)
	return nodes, nil
}

// GetNamespaces retrieves all namespaces in the cluster
func (c *Client) GetNamespaces(ctx context.Context) ([]Namespace, error) {
	cacheKey := "namespaces-all"

	// Check cache
	if cached, found := c.cache.get(cacheKey); found {
		if namespaces, ok := cached.([]Namespace); ok {
			return namespaces, nil
		}
	}

	// Query Kubernetes API
	namespaceList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]Namespace, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespace := convertNamespace(&ns)
		namespaces = append(namespaces, namespace)
	}

	c.cache.set(cacheKey, namespaces)
	return namespaces, nil
}

// Converter functions

func convertDeployment(d *appsv1.Deployment) Deployment {
	// Extract container images
	images := make([]string, 0)
	for _, container := range d.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}

	labels := d.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	selector := d.Spec.Selector.MatchLabels
	if selector == nil {
		selector = make(map[string]string)
	}

	replicas := int32(0)
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	}

	return Deployment{
		Name:              d.Name,
		Namespace:         d.Namespace,
		Replicas:          replicas,
		ReadyReplicas:     d.Status.ReadyReplicas,
		AvailableReplicas: d.Status.AvailableReplicas,
		Labels:            labels,
		Selector:          selector,
		Images:            images,
		Age:               calculateAge(d.CreationTimestamp.Time),
	}
}

func convertService(s *corev1.Service) Service {
	ports := make([]ServicePort, 0, len(s.Spec.Ports))
	for _, p := range s.Spec.Ports {
		ports = append(ports, ServicePort{
			Name:       p.Name,
			Protocol:   string(p.Protocol),
			Port:       p.Port,
			TargetPort: p.TargetPort.String(),
			NodePort:   p.NodePort,
		})
	}

	labels := s.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	selector := s.Spec.Selector
	if selector == nil {
		selector = make(map[string]string)
	}

	annotations := s.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	externalIP := ""
	if len(s.Status.LoadBalancer.Ingress) > 0 {
		if s.Status.LoadBalancer.Ingress[0].IP != "" {
			externalIP = s.Status.LoadBalancer.Ingress[0].IP
		} else if s.Status.LoadBalancer.Ingress[0].Hostname != "" {
			externalIP = s.Status.LoadBalancer.Ingress[0].Hostname
		}
	}

	return Service{
		Name:        s.Name,
		Namespace:   s.Namespace,
		Type:        string(s.Spec.Type),
		ClusterIP:   s.Spec.ClusterIP,
		ExternalIP:  externalIP,
		Ports:       ports,
		Selector:    selector,
		Labels:      labels,
		Annotations: annotations,
		Age:         calculateAge(s.CreationTimestamp.Time),
	}
}

func convertNode(n *corev1.Node) Node {
	// Determine node status
	status := "NotReady"
	for _, condition := range n.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				status = "Ready"
			}
			break
		}
	}

	// Extract node roles from labels
	roles := []string{}
	for key := range n.Labels {
		if key == "node-role.kubernetes.io/master" || key == "node-role.kubernetes.io/control-plane" {
			roles = append(roles, "control-plane")
		} else if key == "node-role.kubernetes.io/worker" {
			roles = append(roles, "worker")
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker") // Default role
	}

	labels := n.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	return Node{
		Name:               n.Name,
		Status:             status,
		Roles:              roles,
		Version:            n.Status.NodeInfo.KubeletVersion,
		OS:                 n.Status.NodeInfo.OSImage,
		ContainerRuntime:   n.Status.NodeInfo.ContainerRuntimeVersion,
		CPUCapacity:        n.Status.Capacity.Cpu().String(),
		MemoryCapacity:     n.Status.Capacity.Memory().String(),
		PodCapacity:        n.Status.Capacity.Pods().String(),
		CPUAllocatable:     n.Status.Allocatable.Cpu().String(),
		MemoryAllocatable:  n.Status.Allocatable.Memory().String(),
		PodAllocatable:     n.Status.Allocatable.Pods().String(),
		Labels:             labels,
		Age:                calculateAge(n.CreationTimestamp.Time),
	}
}

func convertNamespace(ns *corev1.Namespace) Namespace {
	labels := ns.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	return Namespace{
		Name:   ns.Name,
		Status: string(ns.Status.Phase),
		Labels: labels,
		Age:    calculateAge(ns.CreationTimestamp.Time),
	}
}
