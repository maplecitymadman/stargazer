package k8s

import (
	"context"
	"fmt"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// CiliumNetworkPolicySpec represents a Cilium Network Policy specification
type CiliumNetworkPolicySpec struct {
	Name             string                 `json:"name"`
	Namespace        string                 `json:"namespace"`
	Description      string                 `json:"description,omitempty"`
	EndpointSelector map[string]interface{} `json:"endpoint_selector,omitempty"`
	Ingress          []CiliumIngressRule    `json:"ingress,omitempty"`
	Egress           []CiliumEgressRule     `json:"egress,omitempty"`
}

// CiliumIngressRule represents an ingress rule for Cilium
type CiliumIngressRule struct {
	FromEndpoints []map[string]interface{} `json:"from_endpoints,omitempty"`
	FromCIDR      []string                 `json:"from_cidr,omitempty"`
	FromEntities  []string                 `json:"from_entities,omitempty"`
	ToPorts       []CiliumPortRule         `json:"to_ports,omitempty"`
}

// CiliumEgressRule represents an egress rule for Cilium
type CiliumEgressRule struct {
	ToEndpoints []map[string]interface{} `json:"to_endpoints,omitempty"`
	ToCIDR      []string                 `json:"to_cidr,omitempty"`
	ToEntities  []string                 `json:"to_entities,omitempty"`
	ToPorts     []CiliumPortRule         `json:"to_ports,omitempty"`
	ToServices  []map[string]interface{} `json:"to_services,omitempty"`
}

// CiliumPortRule represents port rules for Cilium
type CiliumPortRule struct {
	Ports []CiliumPortProtocol   `json:"ports,omitempty"`
	Rules map[string]interface{} `json:"rules,omitempty"`
}

// CiliumPortProtocol represents port and protocol
type CiliumPortProtocol struct {
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
}

// KyvernoPolicySpec represents a Kyverno Policy specification
type KyvernoPolicySpec struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace,omitempty"` // Empty for ClusterPolicy
	Type        string        `json:"type"`                // "Policy" or "ClusterPolicy"
	Description string        `json:"description,omitempty"`
	Rules       []KyvernoRule `json:"rules"`
	Validation  bool          `json:"validation,omitempty"`
	Mutation    bool          `json:"mutation,omitempty"`
	Generation  bool          `json:"generation,omitempty"`
}

// KyvernoRule represents a Kyverno policy rule
type KyvernoRule struct {
	Name             string                 `json:"name"`
	MatchResources   map[string]interface{} `json:"match_resources"`
	ExcludeResources map[string]interface{} `json:"exclude_resources,omitempty"`
	Validate         map[string]interface{} `json:"validate,omitempty"`
	Mutate           map[string]interface{} `json:"mutate,omitempty"`
	Generate         map[string]interface{} `json:"generate,omitempty"`
}

// PolicyBuilder handles building and managing policies
type PolicyBuilder struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
}

// NewPolicyBuilder creates a new policy builder
func NewPolicyBuilder(clientset kubernetes.Interface, dynamicClient dynamic.Interface) *PolicyBuilder {
	return &PolicyBuilder{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}
}

// BuildCiliumNetworkPolicy builds a Cilium Network Policy from spec
func (pb *PolicyBuilder) BuildCiliumNetworkPolicy(ctx context.Context, spec CiliumNetworkPolicySpec) (string, error) {
	policy := map[string]interface{}{
		"apiVersion": "cilium.io/v2",
		"kind":       "CiliumNetworkPolicy",
		"metadata": map[string]interface{}{
			"name": spec.Name,
		},
		"spec": map[string]interface{}{},
	}

	// Add namespace if specified
	if spec.Namespace != "" {
		metadata := policy["metadata"].(map[string]interface{})
		metadata["namespace"] = spec.Namespace
	}

	specMap := policy["spec"].(map[string]interface{})

	// Add endpoint selector
	if spec.EndpointSelector != nil {
		specMap["endpointSelector"] = spec.EndpointSelector
	}

	// Add ingress rules
	if len(spec.Ingress) > 0 {
		ingressRules := make([]interface{}, 0, len(spec.Ingress))
		for _, rule := range spec.Ingress {
			ruleMap := make(map[string]interface{})
			if len(rule.FromEndpoints) > 0 {
				ruleMap["fromEndpoints"] = rule.FromEndpoints
			}
			if len(rule.FromCIDR) > 0 {
				ruleMap["fromCIDR"] = rule.FromCIDR
			}
			if len(rule.FromEntities) > 0 {
				ruleMap["fromEntities"] = rule.FromEntities
			}
			if len(rule.ToPorts) > 0 {
				toPorts := make([]interface{}, 0, len(rule.ToPorts))
				for _, portRule := range rule.ToPorts {
					portRuleMap := make(map[string]interface{})
					if len(portRule.Ports) > 0 {
						ports := make([]interface{}, 0, len(portRule.Ports))
						for _, pp := range portRule.Ports {
							ports = append(ports, map[string]interface{}{
								"port":     pp.Port,
								"protocol": pp.Protocol,
							})
						}
						portRuleMap["ports"] = ports
					}
					if portRule.Rules != nil {
						portRuleMap["rules"] = portRule.Rules
					}
					toPorts = append(toPorts, portRuleMap)
				}
				ruleMap["toPorts"] = toPorts
			}
			ingressRules = append(ingressRules, ruleMap)
		}
		specMap["ingress"] = ingressRules
	}

	// Add egress rules
	if len(spec.Egress) > 0 {
		egressRules := make([]interface{}, 0, len(spec.Egress))
		for _, rule := range spec.Egress {
			ruleMap := make(map[string]interface{})
			if len(rule.ToEndpoints) > 0 {
				ruleMap["toEndpoints"] = rule.ToEndpoints
			}
			if len(rule.ToCIDR) > 0 {
				ruleMap["toCIDR"] = rule.ToCIDR
			}
			if len(rule.ToEntities) > 0 {
				ruleMap["toEntities"] = rule.ToEntities
			}
			if len(rule.ToPorts) > 0 {
				toPorts := make([]interface{}, 0, len(rule.ToPorts))
				for _, portRule := range rule.ToPorts {
					portRuleMap := make(map[string]interface{})
					if len(portRule.Ports) > 0 {
						ports := make([]interface{}, 0, len(portRule.Ports))
						for _, pp := range portRule.Ports {
							ports = append(ports, map[string]interface{}{
								"port":     pp.Port,
								"protocol": pp.Protocol,
							})
						}
						portRuleMap["ports"] = ports
					}
					if portRule.Rules != nil {
						portRuleMap["rules"] = portRule.Rules
					}
					toPorts = append(toPorts, portRuleMap)
				}
				ruleMap["toPorts"] = toPorts
			}
			if len(rule.ToServices) > 0 {
				ruleMap["toServices"] = rule.ToServices
			}
			egressRules = append(egressRules, ruleMap)
		}
		specMap["egress"] = egressRules
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// BuildKyvernoPolicy builds a Kyverno Policy from spec
func (pb *PolicyBuilder) BuildKyvernoPolicy(ctx context.Context, spec KyvernoPolicySpec) (string, error) {
	kind := "Policy"
	if spec.Type == "ClusterPolicy" {
		kind = "ClusterPolicy"
	}

	policy := map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": spec.Name,
		},
		"spec": map[string]interface{}{
			"rules": make([]interface{}, 0, len(spec.Rules)),
		},
	}

	// Add namespace for Policy (not ClusterPolicy)
	if spec.Namespace != "" && kind == "Policy" {
		metadata := policy["metadata"].(map[string]interface{})
		metadata["namespace"] = spec.Namespace
	}

	// Add description if provided
	if spec.Description != "" {
		metadata := policy["metadata"].(map[string]interface{})
		metadata["annotations"] = map[string]interface{}{
			"description": spec.Description,
		}
	}

	specMap := policy["spec"].(map[string]interface{})

	// Build rules
	rules := make([]interface{}, 0, len(spec.Rules))
	for _, rule := range spec.Rules {
		ruleMap := make(map[string]interface{})
		ruleMap["name"] = rule.Name
		ruleMap["match"] = rule.MatchResources
		if rule.ExcludeResources != nil {
			ruleMap["exclude"] = rule.ExcludeResources
		}
		if rule.Validate != nil {
			ruleMap["validate"] = rule.Validate
		}
		if rule.Mutate != nil {
			ruleMap["mutate"] = rule.Mutate
		}
		if rule.Generate != nil {
			ruleMap["generate"] = rule.Generate
		}
		rules = append(rules, ruleMap)
	}
	specMap["rules"] = rules

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// ApplyCiliumNetworkPolicy applies a Cilium Network Policy to the cluster
func (pb *PolicyBuilder) ApplyCiliumNetworkPolicy(ctx context.Context, yamlContent string, namespace string) error {
	if pb.dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	// Parse YAML
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set GVR for CiliumNetworkPolicy
	gvr := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnetworkpolicies",
	}

	// Determine namespace
	ns := namespace
	if ns == "" {
		ns = obj.GetNamespace()
	}
	if ns == "" {
		ns = "default"
	}

	// Apply the policy
	_, err := pb.dynamicClient.Resource(gvr).Namespace(ns).Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		// Try to update if already exists
		if strings.Contains(err.Error(), "already exists") {
			_, err = pb.dynamicClient.Resource(gvr).Namespace(ns).Update(ctx, &obj, metav1.UpdateOptions{})
		}
		return err
	}

	return nil
}

// ApplyKyvernoPolicy applies a Kyverno Policy to the cluster
func (pb *PolicyBuilder) ApplyKyvernoPolicy(ctx context.Context, yamlContent string, namespace string) error {
	if pb.dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	// Parse YAML
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Determine if it's a Policy or ClusterPolicy
	kind := obj.GetKind()
	var gvr schema.GroupVersionResource
	var ns string

	if kind == "ClusterPolicy" {
		gvr = schema.GroupVersionResource{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "clusterpolicies",
		}
		ns = "" // ClusterPolicy is cluster-scoped
	} else {
		gvr = schema.GroupVersionResource{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "policies",
		}
		ns = namespace
		if ns == "" {
			ns = obj.GetNamespace()
		}
		if ns == "" {
			ns = "default"
		}
	}

	// Apply the policy
	var err error
	if ns == "" {
		// Cluster-scoped
		_, err = pb.dynamicClient.Resource(gvr).Create(ctx, &obj, metav1.CreateOptions{})
		if err != nil && strings.Contains(err.Error(), "already exists") {
			_, err = pb.dynamicClient.Resource(gvr).Update(ctx, &obj, metav1.UpdateOptions{})
		}
	} else {
		// Namespace-scoped
		_, err = pb.dynamicClient.Resource(gvr).Namespace(ns).Create(ctx, &obj, metav1.CreateOptions{})
		if err != nil && strings.Contains(err.Error(), "already exists") {
			_, err = pb.dynamicClient.Resource(gvr).Namespace(ns).Update(ctx, &obj, metav1.UpdateOptions{})
		}
	}

	return err
}

// ApplyNetworkPolicy applies a standard Kubernetes NetworkPolicy to the cluster
func (pb *PolicyBuilder) ApplyNetworkPolicy(ctx context.Context, yamlContent string, namespace string) error {
	// Parse YAML
	var policy networkingv1.NetworkPolicy
	if err := yaml.Unmarshal([]byte(yamlContent), &policy); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Determine namespace
	ns := namespace
	if ns == "" {
		ns = policy.Namespace
	}
	if ns == "" {
		ns = "default"
	}

	// Apply the policy
	_, err := pb.clientset.NetworkingV1().NetworkPolicies(ns).Create(ctx, &policy, metav1.CreateOptions{})
	if err != nil {
		// Try to update if already exists
		if strings.Contains(err.Error(), "already exists") {
			_, err = pb.clientset.NetworkingV1().NetworkPolicies(ns).Update(ctx, &policy, metav1.UpdateOptions{})
		}
		return err
	}

	return nil
}

// DeleteCiliumNetworkPolicy deletes a Cilium Network Policy
func (pb *PolicyBuilder) DeleteCiliumNetworkPolicy(ctx context.Context, name, namespace string) error {
	if pb.dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	gvr := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnetworkpolicies",
	}

	return pb.dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// DeleteKyvernoPolicy deletes a Kyverno Policy
func (pb *PolicyBuilder) DeleteKyvernoPolicy(ctx context.Context, name, namespace string, isClusterPolicy bool) error {
	if pb.dynamicClient == nil {
		return fmt.Errorf("dynamic client not available")
	}

	var gvr schema.GroupVersionResource
	if isClusterPolicy {
		gvr = schema.GroupVersionResource{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "clusterpolicies",
		}
		return pb.dynamicClient.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		gvr = schema.GroupVersionResource{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "policies",
		}
		return pb.dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}
}
