package k8s

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceKey creates a consistent service key from namespace and name
// Format: "namespace/name" or "name" if namespace is empty
func ServiceKey(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", namespace, name)
}

// ParseServiceKey parses a service key into namespace and name
// Returns namespace, name, and whether the key was in namespace/name format
func ParseServiceKey(key string) (namespace, name string, hasNamespace bool) {
	parts := strings.Split(key, "/")
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return "", key, false
}

// TopologyData represents the complete service topology
type TopologyData struct {
	Namespace       string                      `json:"namespace"`
	Services        map[string]ServiceInfo      `json:"services"`
	Connectivity    map[string]ConnectivityInfo `json:"connectivity"`
	Ingress         IngressInfo                 `json:"ingress"`
	Egress          EgressInfo                  `json:"egress"`
	NetworkPolicies []NetworkPolicyInfo         `json:"network_policies"`
	CiliumPolicies  []CiliumNetworkPolicyInfo   `json:"cilium_policies"`
	IstioPolicies   []IstioPolicyInfo           `json:"istio_policies"`
	KyvernoPolicies []KyvernoPolicyInfo         `json:"kyverno_policies"`
	Infrastructure  InfrastructureInfo          `json:"infrastructure"`
	Summary         TopologySummary             `json:"summary"`
	RBAC            RBACData                    `json:"rbac,omitempty"`
	Drift           DriftData                   `json:"drift,omitempty"`
	HubbleEnabled   bool                        `json:"hubble_enabled,omitempty"`
}

// ServiceInfo represents a Kubernetes service with topology metadata
type ServiceInfo struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Type            string            `json:"type"`
	ClusterIP       string            `json:"cluster_ip"`
	Ports           []string          `json:"ports"`
	Labels          map[string]string `json:"labels,omitempty"`
	Pods            []string          `json:"pods"`
	PodCount        int               `json:"pod_count"`
	HealthyPods     int               `json:"healthy_pods"`
	Deployment      string            `json:"deployment,omitempty"`
	HasServiceMesh  bool              `json:"has_service_mesh"`
	ServiceMeshType string            `json:"service_mesh_type,omitempty"` // "istio", "cilium", or ""
	HasCiliumProxy  bool              `json:"has_cilium_proxy"`
	PodSecurity     string            `json:"pod_security,omitempty"` // "privileged", "baseline", "restricted"
	DriftStatus     string            `json:"drift_status,omitempty"` // "Synced", "OutOfSync", "Unknown"
	HasPolicy       bool              `json:"has_policy"`             // True if selected by a NetworkPolicy
	CostStats       *CostStats        `json:"cost_stats,omitempty"`
}

// CostStats represents resource usage and cost insights
type CostStats struct {
	RPS             float64 `json:"rps"`
	CPU             string  `json:"cpu"`
	Memory          string  `json:"memory"`
	PotentialSaving string  `json:"potential_saving"`
	IsZombie        bool    `json:"is_zombie"`
}

// ConnectivityInfo represents connectivity information for a service
type ConnectivityInfo struct {
	Service     string              `json:"service"`
	Connections []ServiceConnection `json:"connections"`
	CanReach    []string            `json:"can_reach"`
	BlockedFrom []string            `json:"blocked_from"`
	IstioRules  []IstioRuleInfo     `json:"istio_rules,omitempty"`
	CiliumRules []CiliumRuleInfo    `json:"cilium_rules,omitempty"`
}

// ServiceConnection represents a connection between services
type ServiceConnection struct {
	Target           string   `json:"target"`
	Allowed          bool     `json:"allowed"`
	Reason           string   `json:"reason"`
	ViaServiceMesh   bool     `json:"via_service_mesh"`
	ServiceMeshType  string   `json:"service_mesh_type,omitempty"`
	BlockedByPolicy  bool     `json:"blocked_by_policy"`
	BlockingPolicies []string `json:"blocking_policies,omitempty"`
	Port             string   `json:"port,omitempty"`
	Protocol         string   `json:"protocol,omitempty"`
}

// NetworkPolicyInfo represents a Kubernetes NetworkPolicy
type NetworkPolicyInfo struct {
	Name      string                      `json:"name"`
	Namespace string                      `json:"namespace"`
	Type      string                      `json:"type"` // "kubernetes" or "cilium"
	YAML      string                      `json:"yaml,omitempty"`
	Policy    *networkingv1.NetworkPolicy `json:"-"` // Internal: actual policy object for evaluation
}

// RBACData represents Kubernetes RBAC information
type RBACData struct {
	RoleBindings        []RoleBindingInfo    `json:"role_bindings"`
	ClusterRoleBindings []RoleBindingInfo    `json:"cluster_role_bindings"`
	ServiceAccounts     []ServiceAccountInfo `json:"service_accounts"`
}

// RoleBindingInfo represents a Kubernetes RoleBinding or ClusterRoleBinding
type RoleBindingInfo struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace,omitempty"`
	RoleName  string        `json:"role_name"`
	RoleKind  string        `json:"role_kind"` // "Role" or "ClusterRole"
	Subjects  []SubjectInfo `json:"subjects"`
}

// SubjectInfo represents an RBAC subject
type SubjectInfo struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// ServiceAccountInfo represents a Kubernetes ServiceAccount
type ServiceAccountInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DriftData represents GitOps drift information
type DriftData struct {
	ArgoEnabled  bool          `json:"argo_enabled"`
	FluxEnabled  bool          `json:"flux_enabled"`
	Applications []ArgoAppInfo `json:"applications,omitempty"`
}

// ArgoAppInfo represents an ArgoCD Application status
type ArgoAppInfo struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Status         string `json:"status"` // "Synced", "OutOfSync"
	RepoURL        string `json:"repo_url"`
	TargetRevision string `json:"target_revision"`
}

// CiliumNetworkPolicyInfo represents a Cilium NetworkPolicy
type CiliumNetworkPolicyInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"` // "ciliumnetworkpolicy" or "ciliumclusterwidenetworkpolicy"
	YAML      string `json:"yaml,omitempty"`
}

// IstioPolicyInfo represents Istio policies (VirtualService, DestinationRule, etc.)
type IstioPolicyInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"` // "virtualservice", "destinationrule", "authorizationpolicy", etc.
	YAML      string `json:"yaml,omitempty"`
}

// KyvernoPolicyInfo represents Kyverno policies
type KyvernoPolicyInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"` // "policy" or "clusterpolicy"
	YAML      string `json:"yaml,omitempty"`
}

// IstioRuleInfo represents Istio-specific routing rules
type IstioRuleInfo struct {
	Type        string `json:"type"` // "virtualservice", "destinationrule"
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CiliumRuleInfo represents Cilium-specific network rules
type CiliumRuleInfo struct {
	Type        string `json:"type"` // "networkpolicy", "clusterwidenetworkpolicy"
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InfrastructureInfo represents detected infrastructure components
type InfrastructureInfo struct {
	CNI             string `json:"cni"` // "cilium", "flannel", "calico", etc.
	CiliumEnabled   bool   `json:"cilium_enabled"`
	IstioEnabled    bool   `json:"istio_enabled"`
	KyvernoEnabled  bool   `json:"kyverno_enabled"`
	NetworkPolicies int    `json:"network_policies"`
	CiliumPolicies  int    `json:"cilium_policies"`
	IstioPolicies   int    `json:"istio_policies"`
	KyvernoPolicies int    `json:"kyverno_policies"`
}

// TopologySummary provides summary statistics
type TopologySummary struct {
	TotalServices      int    `json:"total_services"`
	ServicesWithMesh   int    `json:"services_with_mesh"`
	TotalConnections   int    `json:"total_connections"`
	AllowedConnections int    `json:"allowed_connections"`
	BlockedConnections int    `json:"blocked_connections"`
	MeshCoverage       string `json:"mesh_coverage"`
	CiliumCoverage     string `json:"cilium_coverage"`
	IstioCoverage      string `json:"istio_coverage"`
}

// IngressInfo represents ingress gateways and routes
type IngressInfo struct {
	Gateways          []GatewayInfo           `json:"gateways"`
	KubernetesIngress []KubernetesIngressInfo `json:"kubernetes_ingress"`
	Routes            []IngressRoute          `json:"routes"`
	Connections       []IngressConnection     `json:"connections"`
}

// EgressInfo represents egress gateways and external access
type EgressInfo struct {
	Gateways         []GatewayInfo         `json:"gateways"`
	ExternalServices []ExternalServiceInfo `json:"external_services"`
	Connections      []EgressConnection    `json:"connections"`
	HasEgressGateway bool                  `json:"has_egress_gateway"`
	DirectEgress     bool                  `json:"direct_egress"`
}

// GatewayInfo represents an ingress/egress gateway
type GatewayInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"` // "istio", "nginx", "kubernetes"
	Hosts     []string          `json:"hosts"`
	Ports     []string          `json:"ports"`
	Selector  map[string]string `json:"selector,omitempty"`
	Service   string            `json:"service,omitempty"`
}

// KubernetesIngressInfo represents Kubernetes Ingress resources
type KubernetesIngressInfo struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Hosts       []string `json:"hosts"`
	Paths       []string `json:"paths"`
	Backend     string   `json:"backend"`
	BackendPort string   `json:"backend_port,omitempty"`
	TLS         bool     `json:"tls"`
	Class       string   `json:"class,omitempty"`
}

// IngressRoute represents a route from ingress to service
type IngressRoute struct {
	Gateway   string   `json:"gateway"`
	Host      string   `json:"host"`
	Path      string   `json:"path"`
	Service   string   `json:"service"`
	Namespace string   `json:"namespace"`
	Allowed   bool     `json:"allowed"`
	BlockedBy []string `json:"blocked_by,omitempty"`
	Type      string   `json:"type"` // "istio", "nginx", "kubernetes"
}

// IngressConnection represents Ingress → Service connection
type IngressConnection struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Allowed  bool     `json:"allowed"`
	Reason   string   `json:"reason"`
	Policies []string `json:"policies,omitempty"`
	Port     string   `json:"port,omitempty"`
	Protocol string   `json:"protocol,omitempty"`
}

// EgressConnection represents Service → Egress connection
type EgressConnection struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Allowed  bool     `json:"allowed"`
	Reason   string   `json:"reason"`
	Policies []string `json:"policies,omitempty"`
	Port     string   `json:"port,omitempty"`
	Protocol string   `json:"protocol,omitempty"`
}

// ExternalServiceInfo represents external services (ServiceEntry, etc.)
type ExternalServiceInfo struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Hosts     []string `json:"hosts"`
	Ports     []string `json:"ports"`
	Type      string   `json:"type"` // "serviceentry", "direct"
}

// PathTrace represents a complete path from source to destination
type PathTrace struct {
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Path        []PathHop `json:"path"`
	Allowed     bool      `json:"allowed"`
	BlockedAt   *PathHop  `json:"blocked_at,omitempty"`
	Reason      string    `json:"reason"`
}

// PathHop represents a single hop in a path
type PathHop struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	Type        string   `json:"type"` // "ingress", "service", "egress"
	Allowed     bool     `json:"allowed"`
	Reason      string   `json:"reason"`
	Policies    []string `json:"policies,omitempty"`
	ServiceMesh string   `json:"service_mesh,omitempty"`
}

// NetworkPolicyRule represents a parsed network policy rule
type NetworkPolicyRule struct {
	PolicyName string
	Namespace  string
	Type       string // "ingress" or "egress"
	Allowed    bool
	Ports      []string
	Protocols  []string
	From       []PolicySelector
	To         []PolicySelector
}

// PolicySelector represents a policy selector
type PolicySelector struct {
	NamespaceSelector map[string]string
	PodSelector       map[string]string
	IPBlock           string
}

// GetTopology retrieves complete service topology information
func (c *Client) GetTopology(ctx context.Context, namespace string) (*TopologyData, error) {
	// Determine namespace
	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	// Check cache first
	cacheKey := fmt.Sprintf("topology:%s", ns)
	if data, ok := c.cache.get(cacheKey); ok {
		if topology, ok := data.(*TopologyData); ok {
			return topology, nil
		}
	}

	// Detect infrastructure FIRST (needed for service list)
	infra, err := c.detectInfrastructure(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to detect infrastructure: %v\n", err)
		infra = InfrastructureInfo{}
	}

	var (
		services        map[string]ServiceInfo
		networkPolicies []NetworkPolicyInfo
		ciliumPolicies  []CiliumNetworkPolicyInfo
		istioPolicies   []IstioPolicyInfo
		kyvernoPolicies []KyvernoPolicyInfo
		rbacData        RBACData
		driftData       DriftData
	)

	g, ctx := errgroup.WithContext(ctx)

	// Fetch services
	g.Go(func() error {
		var err error
		services, err = c.getTopologyServices(ctx, ns, infra)
		return err
	})

	// Fetch network policies
	g.Go(func() error {
		var err error
		networkPolicies, err = c.getNetworkPolicies(ctx, ns)
		return err
	})

	// Fetch Cilium policies if Cilium is enabled
	if infra.CiliumEnabled {
		g.Go(func() error {
			var err error
			ciliumPolicies, err = c.getCiliumPolicies(ctx, ns)
			if err != nil {
				fmt.Printf("Warning: Failed to get Cilium policies: %v\n", err)
				return nil // Don't fail the whole topology
			}
			return nil
		})
	}

	// Fetch Istio policies if Istio is enabled
	if infra.IstioEnabled {
		g.Go(func() error {
			var err error
			istioPolicies, err = c.getIstioPolicies(ctx, ns)
			if err != nil {
				fmt.Printf("Warning: Failed to get Istio policies: %v\n", err)
				return nil
			}
			return nil
		})
	}

	// Fetch Kyverno policies if Kyverno is enabled
	if infra.KyvernoEnabled {
		g.Go(func() error {
			var err error
			kyvernoPolicies, err = c.getKyvernoPolicies(ctx, ns)
			if err != nil {
				fmt.Printf("Warning: Failed to get Kyverno policies: %v\n", err)
				return nil // Kyverno is optional
			}
			return nil
		})
	}

	// Fetch RBAC data
	g.Go(func() error {
		var err error
		rbacData, err = c.getRBACData(ctx, ns)
		if err != nil {
			fmt.Printf("Warning: Failed to get RBAC data: %v\n", err)
			return nil
		}
		return nil
	})

	// Fetch Drift data
	g.Go(func() error {
		var err error
		driftData, err = c.getDriftData(ctx, ns)
		if err != nil {
			fmt.Printf("Warning: Failed to get drift data: %v\n", err)
			return nil
		}
		return nil
	})

	// Wait for all fetches to complete
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch topology data: %w", err)
	}

	// Enrich services with policy coverage (must be done after fetching policies)
	for k, svc := range services {
		hasPolicy := false
		for _, np := range networkPolicies {
			if np.Namespace == svc.Namespace {
				// Convert map to selector (np.Policy is populated in getNetworkPolicies)
				if np.Policy != nil {
					npSelector, err := metav1.LabelSelectorAsSelector(&np.Policy.Spec.PodSelector)
					if err == nil && !npSelector.Empty() {
						if npSelector.Matches(labels.Set(svc.Labels)) { // Use svc.Labels (from ServiceInfo) which we populated
							hasPolicy = true
							break
						}
					}
				}
			}
		}
		svc.HasPolicy = hasPolicy
		services[k] = svc
	}

	// Get ingress/egress info (depends on fetched data)
	ingress, err := c.getIngressInfo(ctx, ns, services, infra, networkPolicies, ciliumPolicies, istioPolicies)
	if err != nil {
		fmt.Printf("Warning: Failed to get ingress info: %v\n", err)
		ingress = IngressInfo{}
	}

	egress, err := c.getEgressInfo(ctx, ns, services, infra, networkPolicies, ciliumPolicies, istioPolicies)
	if err != nil {
		fmt.Printf("Warning: Failed to get egress info: %v\n", err)
		egress = EgressInfo{}
	}

	// Detect Hubble
	hubbleEnabled := c.detectHubble(ctx)

	// Map services to drift status
	c.mapServiceToDrift(services, driftData)

	// Build connectivity map
	connectivity := c.buildConnectivityMap(ctx, services, ingress, egress, networkPolicies, ciliumPolicies, istioPolicies, infra)

	// Update infrastructure counts
	infra.NetworkPolicies = len(networkPolicies)
	infra.CiliumPolicies = len(ciliumPolicies)
	infra.IstioPolicies = len(istioPolicies)
	infra.KyvernoPolicies = len(kyvernoPolicies)

	// Calculate summary
	summary := c.calculateTopologySummary(services, connectivity, infra)

	result := &TopologyData{
		Namespace:       ns,
		Services:        services,
		Connectivity:    connectivity,
		Ingress:         ingress,
		Egress:          egress,
		NetworkPolicies: networkPolicies,
		CiliumPolicies:  ciliumPolicies,
		IstioPolicies:   istioPolicies,
		KyvernoPolicies: kyvernoPolicies,
		Infrastructure:  infra,
		Summary:         summary,
		RBAC:            rbacData,
		Drift:           driftData,
		HubbleEnabled:   hubbleEnabled,
	}

	// Cache the result
	c.cache.set(cacheKey, result)

	return result, nil
}

func (c *Client) getRBACData(ctx context.Context, ns string) (RBACData, error) {
	var rbac RBACData

	// Get RoleBindings
	rbList, err := c.clientset.RbacV1().RoleBindings(ns).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, rb := range rbList.Items {
			info := RoleBindingInfo{
				Name:      rb.Name,
				Namespace: rb.Namespace,
				RoleName:  rb.RoleRef.Name,
				RoleKind:  rb.RoleRef.Kind,
			}
			for _, sub := range rb.Subjects {
				info.Subjects = append(info.Subjects, SubjectInfo{
					Kind:      sub.Kind,
					Name:      sub.Name,
					Namespace: sub.Namespace,
				})
			}
			rbac.RoleBindings = append(rbac.RoleBindings, info)
		}
	}

	// Get ClusterRoleBindings
	crbList, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, crb := range crbList.Items {
			info := RoleBindingInfo{
				Name:     crb.Name,
				RoleName: crb.RoleRef.Name,
				RoleKind: crb.RoleRef.Kind,
			}
			for _, sub := range crb.Subjects {
				info.Subjects = append(info.Subjects, SubjectInfo{
					Kind:      sub.Kind,
					Name:      sub.Name,
					Namespace: sub.Namespace,
				})
			}
			rbac.ClusterRoleBindings = append(rbac.ClusterRoleBindings, info)
		}
	}

	// Get ServiceAccounts
	saList, err := c.clientset.CoreV1().ServiceAccounts(ns).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, sa := range saList.Items {
			rbac.ServiceAccounts = append(rbac.ServiceAccounts, ServiceAccountInfo{
				Name:      sa.Name,
				Namespace: sa.Namespace,
			})
		}
	}

	return rbac, nil
}

func (c *Client) evaluatePodSecurity(pod *corev1.Pod) string {
	privileged := false
	hostNetwork := pod.Spec.HostNetwork
	hostPID := pod.Spec.HostPID
	hostIPC := pod.Spec.HostIPC

	for _, container := range pod.Spec.Containers {
		if container.SecurityContext != nil {
			if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				privileged = true
			}
		}
	}

	if privileged || hostNetwork || hostPID || hostIPC {
		return "privileged"
	}

	// Simplified restricted check
	restricted := true
	for _, container := range pod.Spec.Containers {
		sc := container.SecurityContext
		if sc == nil {
			restricted = false
			break
		}
		if sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation {
			restricted = false
			break
		}
		if sc.RunAsNonRoot == nil || !*sc.RunAsNonRoot {
			restricted = false
			break
		}
	}

	if restricted {
		return "restricted"
	}

	return "baseline"
}

// detectInfrastructure detects CNI, service mesh, and policy engines
func (c *Client) detectInfrastructure(ctx context.Context) (InfrastructureInfo, error) {
	infra := InfrastructureInfo{}

	// Detect CNI by checking daemonsets
	daemonsets, err := c.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ds := range daemonsets.Items {
			name := strings.ToLower(ds.Name)
			if strings.Contains(name, "cilium") {
				infra.CNI = "cilium"
				infra.CiliumEnabled = true
				break
			} else if strings.Contains(name, "flannel") {
				infra.CNI = "flannel"
			} else if strings.Contains(name, "calico") {
				infra.CNI = "calico"
			}
		}
	}

	// Detect Istio by checking for istio-system namespace and istiod
	if _, err := c.clientset.CoreV1().Namespaces().Get(ctx, "istio-system", metav1.GetOptions{}); err == nil {
		// Check for istiod deployment
		_, err := c.clientset.AppsV1().Deployments("istio-system").Get(ctx, "istiod", metav1.GetOptions{})
		if err == nil {
			infra.IstioEnabled = true
		}
	}

	// Detect Kyverno by checking for kyverno namespace
	if _, err := c.clientset.CoreV1().Namespaces().Get(ctx, "kyverno", metav1.GetOptions{}); err == nil {
		// Check for kyverno deployment
		_, err := c.clientset.AppsV1().Deployments("kyverno").Get(ctx, "kyverno", metav1.GetOptions{})
		if err == nil {
			infra.KyvernoEnabled = true
		}
	}

	// Count policies
	networkPolicies, _ := c.clientset.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	infra.NetworkPolicies = len(networkPolicies.Items)

	return infra, nil
}

// getTopologyServices retrieves services with topology metadata
func (c *Client) getTopologyServices(ctx context.Context, namespace string, infra InfrastructureInfo) (map[string]ServiceInfo, error) {
	services := make(map[string]ServiceInfo)

	// Get all services
	var serviceList *corev1.ServiceList
	var err error
	if namespace == "" {
		serviceList, err = c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	} else {
		serviceList, err = c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}

	// Get all pods for service mesh detection
	var podList *corev1.PodList
	if namespace == "" {
		podList, err = c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		podList, err = c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}

	// Optimization: Index pods by namespace to avoid O(N*M) complexity
	podsByNamespace := make(map[string][]*corev1.Pod)
	for i := range podList.Items {
		pod := &podList.Items[i]
		podsByNamespace[pod.Namespace] = append(podsByNamespace[pod.Namespace], pod)
	}

	// Fetch cost stats (zombie services)
	costStats, _ := c.detectZombieServices(ctx, namespace)

	// Build service info
	for _, svc := range serviceList.Items {
		serviceInfo := ServiceInfo{
			Name:        svc.Name,
			Namespace:   svc.Namespace,
			Type:        string(svc.Spec.Type),
			ClusterIP:   svc.Spec.ClusterIP,
			Labels:      svc.Labels,
			PodSecurity: "baseline", // Default
			HasPolicy:   false,
		}

		// Find pods matching service selector
		selector := labels.Set(svc.Spec.Selector).AsSelector()
		var matchingPods []string
		var healthyPods int
		var hasIstio bool
		var hasCiliumProxy bool

		// Optimized: Only iterate pods in the same namespace
		if namespacePods, ok := podsByNamespace[svc.Namespace]; ok {
			for _, pod := range namespacePods {
				if selector.Matches(labels.Set(pod.Labels)) {
					matchingPods = append(matchingPods, pod.Name)
					if pod.Status.Phase == corev1.PodRunning {
						healthyPods++
					}

					// Check for Istio sidecar
					if infra.IstioEnabled {
						for _, container := range pod.Spec.Containers {
							if strings.Contains(container.Image, "istio/proxy") || strings.Contains(container.Image, "istio") {
								hasIstio = true
								break
							}
						}
						// Check for Istio annotations
						if _, ok := pod.Annotations["sidecar.istio.io/status"]; ok {
							hasIstio = true
						}
					}

					// Evaluate Pod Security
					security := c.evaluatePodSecurity(pod)
					if serviceInfo.PodSecurity == "" || security == "privileged" {
						serviceInfo.PodSecurity = security
					}

					// Check for Cilium proxy (eBPF)
					if infra.CiliumEnabled {
						// Cilium uses eBPF, but we can check for Cilium annotations
						if val, ok := pod.Annotations["io.cilium.k8s.policy.name"]; ok && val != "" {
							hasCiliumProxy = true
						}
					}
				}
			}
		}

		// Format ports
		var portStrings []string
		for _, port := range svc.Spec.Ports {
			portStr := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
			if port.Name != "" {
				portStr = fmt.Sprintf("%s:%s", port.Name, portStr)
			}
			portStrings = append(portStrings, portStr)
		}

		// Determine service mesh type
		meshType := ""
		if hasIstio {
			meshType = "istio"
		} else if hasCiliumProxy && infra.CiliumEnabled {
			meshType = "cilium"
		}

		serviceInfo.Ports = portStrings
		serviceInfo.Pods = matchingPods
		serviceInfo.PodCount = len(matchingPods)
		serviceInfo.HealthyPods = healthyPods
		serviceInfo.HasServiceMesh = hasIstio || hasCiliumProxy
		serviceInfo.ServiceMeshType = meshType
		serviceInfo.HasCiliumProxy = hasCiliumProxy

		// Try to find associated deployment
		if len(matchingPods) > 0 {
			// Get first pod to find deployment
			for _, podName := range matchingPods {
				for _, pod := range podList.Items {
					if pod.Name == podName && pod.Namespace == svc.Namespace {
						if deployName, ok := pod.Labels["app"]; ok {
							serviceInfo.Deployment = deployName
							break
						}
					}
				}
				if serviceInfo.Deployment != "" {
					break
				}
			}
		}

		key := fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)

		// Calculate resource requests from the first matching pod (assuming identical replicas)
		var totalCPU int64    // millicores
		var totalMemory int64 // MiB

		if len(matchingPods) > 0 {
			// Find the actual pod object
			var firstPod *corev1.Pod
			if namespacePods, ok := podsByNamespace[svc.Namespace]; ok {
				for _, p := range namespacePods {
					if p.Name == matchingPods[0] {
						firstPod = p
						break
					}
				}
			}

			if firstPod != nil {
				for _, container := range firstPod.Spec.Containers {
					cpu := container.Resources.Requests.Cpu().MilliValue()
					mem := container.Resources.Requests.Memory().Value() / (1024 * 1024)
					totalCPU += cpu
					totalMemory += totalMemory + mem
				}
			}
		}

		if stats, ok := costStats[key]; ok {
			// Use real resource data if available, otherwise defaults (0)
			stats.CPU = fmt.Sprintf("%dm", totalCPU)
			stats.Memory = fmt.Sprintf("%dMi", totalMemory)

			// Calculate potential monthly savings if it's a zombie service
			// Estimation: ~$30/vCPU/mo, ~$4/GB/mo
			if stats.IsZombie && (totalCPU > 0 || totalMemory > 0) {
				cpuCost := float64(totalCPU) / 1000.0 * 30.0
				memCost := float64(totalMemory) / 1024.0 * 4.0
				stats.PotentialSaving = fmt.Sprintf("$%.2f/mo", cpuCost+memCost)
			} else {
				stats.PotentialSaving = "$0.00/mo"
			}

			serviceInfo.CostStats = &stats
		}
		services[key] = serviceInfo
	}

	return services, nil
}

// getNetworkPolicies retrieves Kubernetes NetworkPolicies
func (c *Client) getNetworkPolicies(ctx context.Context, namespace string) ([]NetworkPolicyInfo, error) {
	var policyList *networkingv1.NetworkPolicyList
	var err error

	if namespace == "" {
		policyList, err = c.clientset.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	} else {
		policyList, err = c.clientset.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}

	policies := make([]NetworkPolicyInfo, 0, len(policyList.Items))
	for _, policy := range policyList.Items {
		// Create a copy of the policy for evaluation
		policyCopy := policy
		policies = append(policies, NetworkPolicyInfo{
			Name:      policy.Name,
			Namespace: policy.Namespace,
			Type:      "kubernetes",
			Policy:    &policyCopy,
		})
	}

	return policies, nil
}

// getCiliumPolicies retrieves Cilium NetworkPolicies (CRDs)
func (c *Client) getCiliumPolicies(ctx context.Context, namespace string) ([]CiliumNetworkPolicyInfo, error) {
	policies := []CiliumNetworkPolicyInfo{}

	if c.dynamicClient == nil {
		return policies, nil
	}

	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	// CiliumNetworkPolicy
	cnpGVR := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnetworkpolicies",
	}

	var cnpList *unstructured.UnstructuredList
	var err error
	if ns == "" {
		cnpList, err = c.dynamicClient.Resource(cnpGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		cnpList, err = c.dynamicClient.Resource(cnpGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		for _, item := range cnpList.Items {
			policies = append(policies, CiliumNetworkPolicyInfo{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Type:      "ciliumnetworkpolicy",
			})
		}
	}

	// CiliumClusterwideNetworkPolicy
	ccnpGVR := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumclusterwidenetworkpolicies",
	}

	ccnpList, err := c.dynamicClient.Resource(ccnpGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range ccnpList.Items {
			policies = append(policies, CiliumNetworkPolicyInfo{
				Name:      item.GetName(),
				Namespace: "", // Cluster-wide
				Type:      "ciliumclusterwidenetworkpolicy",
			})
		}
	}

	return policies, nil
}

// getIstioPolicies retrieves Istio policies (VirtualService, DestinationRule, etc.)
func (c *Client) getIstioPolicies(ctx context.Context, namespace string) ([]IstioPolicyInfo, error) {
	policies := []IstioPolicyInfo{}

	if c.dynamicClient == nil {
		return policies, nil
	}

	ns := namespace
	if namespace == "all" || namespace == "" {
		ns = ""
	}

	// Detect Istio API version
	istioVersion, err := c.detectIstioAPIVersion(ctx)
	if err != nil {
		return policies, nil // Istio not available
	}

	// VirtualService
	vsGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  istioVersion,
		Resource: "virtualservices",
	}

	var vsList *unstructured.UnstructuredList
	if ns == "" {
		vsList, err = c.dynamicClient.Resource(vsGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		vsList, err = c.dynamicClient.Resource(vsGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		for _, item := range vsList.Items {
			policies = append(policies, IstioPolicyInfo{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Type:      "virtualservice",
			})
		}
	}

	// DestinationRule
	drGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  istioVersion,
		Resource: "destinationrules",
	}

	var drList *unstructured.UnstructuredList
	if ns == "" {
		drList, err = c.dynamicClient.Resource(drGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		drList, err = c.dynamicClient.Resource(drGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	}

	if err == nil {
		for _, item := range drList.Items {
			policies = append(policies, IstioPolicyInfo{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Type:      "destinationrule",
			})
		}
	}

	// AuthorizationPolicy (uses security.istio.io, which may have different version)
	// Try v1 first (most common), then fall back to v1beta1 (older Istio versions)
	apGVR := schema.GroupVersionResource{
		Group:    "security.istio.io",
		Version:  "v1",
		Resource: "authorizationpolicies",
	}

	var apList *unstructured.UnstructuredList
	if ns == "" {
		apList, err = c.dynamicClient.Resource(apGVR).Namespace("").List(ctx, metav1.ListOptions{})
	} else {
		apList, err = c.dynamicClient.Resource(apGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
	}

	// If v1 fails, try v1beta1 (older Istio versions)
	if err != nil {
		apGVR.Version = "v1beta1"
		if ns == "" {
			apList, err = c.dynamicClient.Resource(apGVR).Namespace("").List(ctx, metav1.ListOptions{})
		} else {
			apList, err = c.dynamicClient.Resource(apGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
		}
	}

	// Log error but don't fail completely - AuthorizationPolicies are optional
	if err != nil {
		// Check if it's a "not found" error (CRD not installed) vs actual error
		if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "the server could not find") {
			// Only log non-CRD errors
			fmt.Printf("Warning: Failed to list AuthorizationPolicies: %v\n", err)
		}
	} else {
		for _, item := range apList.Items {
			policies = append(policies, IstioPolicyInfo{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
				Type:      "authorizationpolicy",
			})
		}
	}

	return policies, nil
}

// getKyvernoPolicies retrieves Kyverno policies
func (c *Client) getKyvernoPolicies(ctx context.Context, namespace string) ([]KyvernoPolicyInfo, error) {
	// Note: This requires Kyverno CRDs
	// For now, return empty - this will be enhanced when we add dynamic client support
	return []KyvernoPolicyInfo{}, nil
}

// buildConnectivityMap builds the connectivity information between services (includes ingress/egress)
func (c *Client) buildConnectivityMap(ctx context.Context, services map[string]ServiceInfo, ingress IngressInfo, egress EgressInfo, networkPolicies []NetworkPolicyInfo, ciliumPolicies []CiliumNetworkPolicyInfo, istioPolicies []IstioPolicyInfo, infra InfrastructureInfo) map[string]ConnectivityInfo {
	connectivity := make(map[string]ConnectivityInfo)

	for serviceKey, service := range services {
		connInfo := ConnectivityInfo{
			Service:     service.Name,
			Connections: []ServiceConnection{},
			CanReach:    []string{},
			BlockedFrom: []string{},
		}

		// Build connections to other services
		for targetKey, targetService := range services {
			if serviceKey == targetKey {
				continue
			}
			// Skip if different namespace and we're filtering by namespace
			if service.Namespace != "" && targetService.Namespace != service.Namespace {
				continue
			}

			// Check if connection is allowed (improved policy evaluation)
			allowed := true
			reason := "No policy blocking"
			blockedByPolicy := false
			var blockingPolicies []string

			// Evaluate network policies for the source service namespace
			hasPolicies := false
			for _, policyInfo := range networkPolicies {
				if policyInfo.Namespace == service.Namespace {
					hasPolicies = true
					if policyInfo.Policy != nil {
						// Check if this is a default-deny policy (empty ingress/egress rules)
						policy := policyInfo.Policy
						hasIngressType := false
						hasEgressType := false
						for _, pt := range policy.Spec.PolicyTypes {
							if pt == networkingv1.PolicyTypeIngress {
								hasIngressType = true
							}
							if pt == networkingv1.PolicyTypeEgress {
								hasEgressType = true
							}
						}

						// Default deny: policy type specified but no rules
						if hasIngressType && len(policy.Spec.Ingress) == 0 {
							blockedByPolicy = true
							blockingPolicies = append(blockingPolicies, policy.Name)
							allowed = false
							reason = fmt.Sprintf("Blocked by default-deny NetworkPolicy: %s (no ingress rules)", policy.Name)
						}
						if hasEgressType && len(policy.Spec.Egress) == 0 {
							blockedByPolicy = true
							blockingPolicies = append(blockingPolicies, policy.Name)
							allowed = false
							reason = fmt.Sprintf("Blocked by default-deny NetworkPolicy: %s (no egress rules)", policy.Name)
						}

						// If policy has rules, check if they might allow this connection
						// This is a simplified check - full evaluation would require pod selector matching
						if hasIngressType && len(policy.Spec.Ingress) > 0 {
							// Policy has allow rules, so connection might be allowed
							// In a full implementation, we'd check pod selectors and namespace selectors
						}
					} else {
						// Policy object not available, use conservative approach
						// If policy name suggests default-deny, assume blocking
						if strings.Contains(strings.ToLower(policyInfo.Name), "deny") ||
							strings.Contains(strings.ToLower(policyInfo.Name), "block") {
							blockedByPolicy = true
							blockingPolicies = append(blockingPolicies, policyInfo.Name)
							allowed = false
							reason = fmt.Sprintf("Potentially blocked by NetworkPolicy: %s", policyInfo.Name)
						}
					}
				}
			}

			// If no policies exist, allow (default Kubernetes behavior)
			if !hasPolicies {
				allowed = true
				reason = "No NetworkPolicies in namespace"
			}

			connection := ServiceConnection{
				Target:           targetService.Name,
				Allowed:          allowed,
				Reason:           reason,
				ViaServiceMesh:   service.HasServiceMesh,
				ServiceMeshType:  service.ServiceMeshType,
				BlockedByPolicy:  blockedByPolicy,
				BlockingPolicies: blockingPolicies,
			}

			connInfo.Connections = append(connInfo.Connections, connection)

			if allowed {
				connInfo.CanReach = append(connInfo.CanReach, targetService.Name)
			} else {
				connInfo.BlockedFrom = append(connInfo.BlockedFrom, targetService.Name)
			}
		}

		connectivity[serviceKey] = connInfo
	}

	// Add ingress gateway as special node
	if len(ingress.Gateways) > 0 || len(ingress.KubernetesIngress) > 0 {
		ingressConn := ConnectivityInfo{
			Service:     "ingress-gateway",
			Connections: []ServiceConnection{},
			CanReach:    []string{},
			BlockedFrom: []string{},
		}

		for _, conn := range ingress.Connections {
			ingressConn.Connections = append(ingressConn.Connections, ServiceConnection{
				Target:           conn.To,
				Allowed:          conn.Allowed,
				Reason:           conn.Reason,
				BlockedByPolicy:  !conn.Allowed,
				BlockingPolicies: conn.Policies,
			})
			if conn.Allowed {
				ingressConn.CanReach = append(ingressConn.CanReach, conn.To)
			} else {
				ingressConn.BlockedFrom = append(ingressConn.BlockedFrom, conn.To)
			}
		}
		connectivity["ingress-gateway"] = ingressConn
	}

	// Add egress gateway as special node
	if egress.HasEgressGateway || len(egress.ExternalServices) > 0 {
		egressConn := ConnectivityInfo{
			Service:     "egress-gateway",
			Connections: []ServiceConnection{},
			CanReach:    []string{},
			BlockedFrom: []string{},
		}

		for _, conn := range egress.Connections {
			egressConn.Connections = append(egressConn.Connections, ServiceConnection{
				Target:           conn.To,
				Allowed:          conn.Allowed,
				Reason:           conn.Reason,
				BlockedByPolicy:  !conn.Allowed,
				BlockingPolicies: conn.Policies,
			})
		}
		connectivity["egress-gateway"] = egressConn
	}

	// Add reverse connections (services → ingress, services → egress)
	for serviceKey, service := range services {
		connInfo := connectivity[serviceKey]

		// Check if this service can reach egress
		for _, egressConn := range egress.Connections {
			if egressConn.From == service.Name && egressConn.Allowed {
				connInfo.Connections = append(connInfo.Connections, ServiceConnection{
					Target:  "egress-gateway",
					Allowed: true,
					Reason:  egressConn.Reason,
				})
				connInfo.CanReach = append(connInfo.CanReach, "egress-gateway")
			}
		}

		connectivity[serviceKey] = connInfo
	}

	return connectivity
}

// calculateTopologySummary calculates summary statistics
func (c *Client) calculateTopologySummary(services map[string]ServiceInfo, connectivity map[string]ConnectivityInfo, infra InfrastructureInfo) TopologySummary {
	totalServices := len(services)
	servicesWithMesh := 0
	totalConnections := 0
	allowedConnections := 0
	blockedConnections := 0

	for _, service := range services {
		if service.HasServiceMesh {
			servicesWithMesh++
		}
	}

	for _, connInfo := range connectivity {
		for _, conn := range connInfo.Connections {
			totalConnections++
			if conn.Allowed && !conn.BlockedByPolicy {
				allowedConnections++
			} else {
				// Count as blocked if it's explicitly blocked by policy or just not allowed
				blockedConnections++
			}
		}
	}

	meshCoverage := "0%"
	if totalServices > 0 {
		meshCoverage = fmt.Sprintf("%.0f%%", float64(servicesWithMesh)/float64(totalServices)*100)
	}

	ciliumCoverage := "0%"
	istioCoverage := "0%"
	if totalServices > 0 {
		ciliumCount := 0
		istioCount := 0
		for _, service := range services {
			if service.ServiceMeshType == "cilium" {
				ciliumCount++
			} else if service.ServiceMeshType == "istio" {
				istioCount++
			}
		}
		ciliumCoverage = fmt.Sprintf("%.0f%%", float64(ciliumCount)/float64(totalServices)*100)
		istioCoverage = fmt.Sprintf("%.0f%%", float64(istioCount)/float64(totalServices)*100)
	}

	return TopologySummary{
		TotalServices:      totalServices,
		ServicesWithMesh:   servicesWithMesh,
		TotalConnections:   totalConnections,
		AllowedConnections: allowedConnections,
		BlockedConnections: blockedConnections,
		MeshCoverage:       meshCoverage,
		CiliumCoverage:     ciliumCoverage,
		IstioCoverage:      istioCoverage,
	}
}

// detectIstioAPIVersion detects which Istio API version to use
func (c *Client) detectIstioAPIVersion(ctx context.Context) (string, error) {
	if c.dynamicClient == nil {
		return "", fmt.Errorf("dynamic client not available")
	}

	// Try v1beta1 first (Istio 1.10+)
	v1beta1GVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
		Resource: "gateways",
	}

	_, err := c.dynamicClient.Resource(v1beta1GVR).Namespace("").List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil {
		return "v1beta1", nil
	}

	// Fall back to v1alpha3 (older Istio)
	v1alpha3GVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "gateways",
	}

	_, err = c.dynamicClient.Resource(v1alpha3GVR).Namespace("").List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil {
		return "v1alpha3", nil
	}

	// Try v1 (newest)
	v1GVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1",
		Resource: "gateways",
	}

	_, err = c.dynamicClient.Resource(v1GVR).Namespace("").List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil {
		return "v1", nil
	}

	return "", fmt.Errorf("Istio API not detected")
}

// detectHubble checks if Hubble is installed (Cilium observability)
func (c *Client) detectHubble(ctx context.Context) bool {
	// Check for Hubble deployment or service
	deployments, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app=hubble",
	})
	if err == nil && len(deployments.Items) > 0 {
		return true
	}

	// Check for Hubble service
	services, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app=hubble",
	})
	if err == nil && len(services.Items) > 0 {
		return true
	}

	return false
}
