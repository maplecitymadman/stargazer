'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';
import toast from 'react-hot-toast';

interface NetworkPoliciesPageProps {
  subsection?: string;
  namespace?: string;
}

export default function NetworkPoliciesPage({ subsection, namespace }: NetworkPoliciesPageProps) {
  const [activeTab, setActiveTab] = useState(subsection || 'view');
  const [topology, setTopology] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (subsection) {
      setActiveTab(subsection);
    }
    loadData();
  }, [subsection, namespace]);

  const loadData = async () => {
    try {
      const data = await apiClient.getServiceTopology(namespace);
      setTopology(data);
    } catch (error) {
      console.error('Error loading policies:', error);
    } finally {
      setLoading(false);
    }
  };

  const tabs = [
    { id: 'view', label: 'Active Policies', icon: 'scan' as const },
    { id: 'build', label: 'Policy Builder', icon: 'execute' as const },
    { id: 'test', label: 'Policy Playground', icon: 'info' as const },
  ];

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading policies...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Tab Navigation */}
      <div className="flex gap-2 border-b border-[rgba(255,255,255,0.08)] overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-medium transition-all border-b-2 whitespace-nowrap cursor-pointer active:scale-[0.98] ${
              activeTab === tab.id
                ? 'text-[#3b82f6] border-[#3b82f6]'
                : 'text-[#71717a] border-transparent hover:text-[#e4e4e7]'
            }`}
            aria-label={`Switch to ${tab.label} tab`}
          >
            <div className="flex items-center gap-2">
              <Icon name={tab.icon} size="sm" />
              {tab.label}
            </div>
          </button>
        ))}
      </div>

      {/* Content */}
      <div>
        {activeTab === 'view' && <ViewPolicies topology={topology} namespace={namespace} />}
        {activeTab === 'build' && <BuildPolicy namespace={namespace} />}
        {activeTab === 'test' && <TestPolicy namespace={namespace} />}
      </div>
    </div>
  );
}

// View Policies Tab
function ViewPolicies({ topology, namespace }: { topology: any; namespace?: string }) {
  const [selectedPolicy, setSelectedPolicy] = useState<{name: string, type: string} | null>(null);
  const [policyYaml, setPolicyYaml] = useState<string>('');
  const [loadingYaml, setLoadingYaml] = useState(false);

  useEffect(() => {
    if (selectedPolicy) {
      loadPolicyYaml();
    } else {
      setPolicyYaml('');
    }
  }, [selectedPolicy]);

  const loadPolicyYaml = async () => {
    if (!selectedPolicy) return;
    setLoadingYaml(true);
    try {
      // Currently only supporting standard NetworkPolicies for YAML retrieval in this demo
      // In a real app, we'd have endpoints for Cilium/Kyverno YAMLs too
      const data = await apiClient.getNetworkPolicyYaml(selectedPolicy.name, namespace);
      setPolicyYaml(data.yaml || JSON.stringify(data, null, 2));
    } catch (error) {
       setPolicyYaml('Error loading policy details. Context might be missing or policy not found.');
    } finally {
      setLoadingYaml(false);
    }
  };

  if (!topology) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="info" className="text-[#71717a] text-4xl mb-4" />
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">No Policy Data Available</h3>
        <p className="text-sm text-[#71717a] mb-4">Unable to load network policy information</p>
        <div className="text-xs text-[#71717a] space-y-1">
          <p>• Check your Kubernetes connection</p>
          <p>• Verify RBAC permissions are configured</p>
          <p>• Try refreshing the page</p>
        </div>
      </div>
    );
  }

  const infra = topology.infrastructure || {};

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-2 space-y-6">
        {/* Policy Summary */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="card rounded-lg p-4 bg-[#1a1a24]/50 border border-[rgba(255,255,255,0.05)]">
            <div className="text-xs text-[#71717a] mb-1 font-medium uppercase tracking-wider">K8s Native</div>
            <div className="text-2xl font-bold text-[#e4e4e7]">{infra.network_policies || 0}</div>
          </div>
          <div className="card rounded-lg p-4 bg-[#1a1a24]/50 border border-[rgba(255,255,255,0.05)]">
            <div className="text-xs text-[#71717a] mb-1 font-medium uppercase tracking-wider">Cilium</div>
            <div className="text-2xl font-bold text-[#e4e4e7]">{infra.cilium_policies || 0}</div>
          </div>
          <div className="card rounded-lg p-4 bg-[#1a1a24]/50 border border-[rgba(255,255,255,0.05)]">
            <div className="text-xs text-[#71717a] mb-1 font-medium uppercase tracking-wider">Istio</div>
            <div className="text-2xl font-bold text-[#e4e4e7]">{infra.istio_policies || 0}</div>
          </div>
          <div className="card rounded-lg p-4 bg-[#1a1a24]/50 border border-[rgba(255,255,255,0.05)]">
            <div className="text-xs text-[#71717a] mb-1 font-medium uppercase tracking-wider">Kyverno</div>
            <div className="text-2xl font-bold text-[#e4e4e7]">{infra.kyverno_policies || 0}</div>
          </div>
        </div>

        {/* Policy Lists */}
        <div className="space-y-6">
          {/* K8s NetworkPolicies */}
          <div className="card rounded-lg p-5">
            <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7] flex items-center gap-2">
              <Icon name="scan" className="text-blue-400" size="sm" />
              Kubernetes NetworkPolicies
            </h3>
            {topology.network_policies && topology.network_policies.length > 0 ? (
              <div className="space-y-2">
                {topology.network_policies.map((policy: any, idx: number) => (
                  <button
                    key={idx}
                    onClick={() => setSelectedPolicy({ name: policy.name, type: 'k8s' })}
                    className={`w-full text-left p-3 rounded border transition-all flex items-center justify-between group ${
                      selectedPolicy?.name === policy.name
                        ? 'bg-[#3b82f6]/10 border-[#3b82f6]/50'
                        : 'bg-[#1a1a24] border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30'
                    }`}
                  >
                    <div>
                      <div className="text-sm font-medium text-[#e4e4e7] group-hover:text-[#3b82f6] transition-colors">
                        {policy.name}
                      </div>
                      <div className="text-xs text-[#71717a]">Namespace: {policy.namespace}</div>
                    </div>
                    <Icon name="arrowRight" className="text-[#71717a] group-hover:text-[#3b82f6] opacity-0 group-hover:opacity-100 transition-all" size="sm" />
                  </button>
                ))}
              </div>
            ) : (
              <p className="text-sm text-[#71717a] italic">No standard network policies found.</p>
            )}
          </div>

          {/* Cilium Policies */}
          <div className="card rounded-lg p-5">
            <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7] flex items-center gap-2">
              <Icon name="scan" className="text-purple-400" size="sm" />
              Cilium NetworkPolicies
            </h3>
            {topology.cilium_policies && topology.cilium_policies.length > 0 ? (
              <div className="space-y-2">
                {topology.cilium_policies.map((policy: any, idx: number) => (
                  <button
                     key={idx}
                    onClick={() => setSelectedPolicy({ name: policy.name, type: 'cilium' })}
                    className={`w-full text-left p-3 rounded border transition-all flex items-center justify-between group ${
                      selectedPolicy?.name === policy.name
                        ? 'bg-[#3b82f6]/10 border-[#3b82f6]/50'
                        : 'bg-[#1a1a24] border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30'
                    }`}
                  >
                    <div>
                      <div className="text-sm font-medium text-[#e4e4e7] group-hover:text-[#3b82f6] transition-colors">{policy.name}</div>
                      <div className="text-xs text-[#71717a]">Namespace: {policy.namespace}</div>
                    </div>
                    <Icon name="arrowRight" className="text-[#71717a] group-hover:text-[#3b82f6] opacity-0 group-hover:opacity-100 transition-all" size="sm" />
                  </button>
                ))}
              </div>
            ) : (
                <p className="text-sm text-[#71717a] italic">No Cilium policies found.</p>
            )}
          </div>
        </div>
      </div>

      {/* Details Panel */}
      <div className="lg:col-span-1">
        <div className="card rounded-lg p-5 sticky top-24 h-[calc(100vh-8rem)] flex flex-col">
          <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7]">Policy Details</h3>
          {selectedPolicy ? (
            <div className="flex-1 flex flex-col min-h-0">
              <div className="mb-4">
                 <div className="text-xs text-[#71717a] uppercase mb-1">SELECTED POLICY</div>
                 <div className="text-base font-bold text-[#e4e4e7] break-all">{selectedPolicy.name}</div>
                 <div className="text-xs text-[#3b82f6] font-mono mt-1">{selectedPolicy.type.toUpperCase()}</div>
              </div>

               <div className="flex-1 bg-[#0a0a0f] rounded border border-[rgba(255,255,255,0.08)] p-3 overflow-auto">
                 {loadingYaml ? (
                   <div className="h-full flex flex-col items-center justify-center text-[#71717a]">
                      <Icon name="loading" className="animate-pulse mb-2 text-[#3b82f6]" size="md" />
                      <span className="text-xs">Fetching policy definition...</span>
                   </div>
                 ) : (
                   <pre className="text-xs text-green-400 font-mono whitespace-pre-wrap break-all">
                     {policyYaml || "No details available."}
                   </pre>
                 )}
               </div>
            </div>
          ) : (
            <div className="h-full flex flex-col items-center justify-center text-center text-[#71717a] border-2 border-dashed border-[rgba(255,255,255,0.05)] rounded-lg">
              <Icon name="info" className="mb-3 opacity-50" size="lg" />
              <p className="text-sm font-medium">Select a policy</p>
              <p className="text-xs opacity-70 mt-1 max-w-[200px]">Click on a policy from the list to view its configuration.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Build Policy Tab
function BuildPolicy({ namespace }: { namespace?: string }) {
  const [policyType, setPolicyType] = useState<'cilium' | 'kyverno'>('cilium');
  const [topology, setTopology] = useState<any>(null);
  const [services, setServices] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [building, setBuilding] = useState(false);
  const [generatedYaml, setGeneratedYaml] = useState<string>('');
  const [policyName, setPolicyName] = useState('');
  const [policyNamespace, setPolicyNamespace] = useState(namespace || 'default');
  const [selectedServices, setSelectedServices] = useState<string[]>([]);
  const [selectedBlocked, setSelectedBlocked] = useState<string[]>([]);
  const [description, setDescription] = useState('');

  useEffect(() => {
    loadServices();
  }, [namespace]);

  const loadServices = async () => {
    setLoading(true);
    try {
      const [topologyData, servicesData] = await Promise.all([
        apiClient.getServiceTopology(namespace),
        apiClient.getServices(namespace),
      ]);
      setTopology(topologyData);
      setServices(servicesData.services || []);
    } catch (error) {
      console.error('Error loading services:', error);
    } finally {
      setLoading(false);
    }
  };

  const toggleService = (serviceKey: string, isBlocked: boolean) => {
    if (isBlocked) {
      setSelectedBlocked(prev =>
        prev.includes(serviceKey)
          ? prev.filter(s => s !== serviceKey)
          : [...prev, serviceKey]
      );
    } else {
      setSelectedServices(prev =>
        prev.includes(serviceKey)
          ? prev.filter(s => s !== serviceKey)
          : [...prev, serviceKey]
      );
    }
  };

  const buildPolicy = async () => {
    if (!policyName) {
      toast.error('Please enter a policy name');
      return;
    }

    setBuilding(true);
    try {
      if (policyType === 'cilium') {
        // Build Cilium Network Policy
        const spec = {
          name: policyName,
          namespace: policyNamespace,
          description: description,
          endpoint_selector: {
            matchLabels: {},
          },
          ingress: selectedServices.length > 0 ? [{
            from_endpoints: selectedServices.map(serviceKey => {
              const [ns, name] = serviceKey.includes('/') ? serviceKey.split('/') : [policyNamespace, serviceKey];
              return {
                matchLabels: {
                  'k8s:io.kubernetes.pod.namespace': ns,
                  'k8s:k8s-app': name,
                },
              };
            }),
          }] : [],
          egress: selectedServices.length > 0 ? [{
            to_endpoints: selectedServices.map(serviceKey => {
              const [ns, name] = serviceKey.includes('/') ? serviceKey.split('/') : [policyNamespace, serviceKey];
              return {
                matchLabels: {
                  'k8s:io.kubernetes.pod.namespace': ns,
                  'k8s:k8s-app': name,
                },
              };
            }),
          }] : [],
        };

        const result = await apiClient.buildCiliumPolicy(spec);
        setGeneratedYaml(result.yaml);
      } else {
        // Build Kyverno Policy
        const spec = {
          name: policyName,
          namespace: policyNamespace,
          type: 'Policy',
          description: description,
          rules: [
            {
              name: 'network-policy-rule',
              match_resources: {
                kinds: ['Pod'],
                namespaces: [policyNamespace],
              },
              validate: {
                message: 'Network policy validation',
                pattern: {},
              },
            },
          ],
        };

        const result = await apiClient.buildKyvernoPolicy(spec);
        setGeneratedYaml(result.yaml);
        toast.success('Policy built successfully');
      }
    } catch (error: any) {
      toast.error(`Failed to build policy: ${error.response?.data?.error || error.message}`);
    } finally {
      setBuilding(false);
    }
  };

  const exportYaml = async () => {
    if (!generatedYaml) return;

    try {
      let blob: Blob;
      if (policyType === 'cilium') {
        blob = await apiClient.exportCiliumPolicy(generatedYaml);
      } else {
        blob = await apiClient.exportKyvernoPolicy(generatedYaml, 'Policy');
      }

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${policyName}-${policyType}-policy.yaml`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      toast.success('Policy exported successfully');
    } catch (error: any) {
      toast.error(`Failed to export: ${error.response?.data?.error || error.message}`);
    }
  };

  const serviceList = services.map(s => {
    const key = s.namespace ? `${s.namespace}/${s.name}` : s.name;
    return { key, name: s.name, namespace: s.namespace || 'default', display: s.display || key };
  });

  return (
    <div className="space-y-4">
      {/* Policy Type Selector */}
      <div className="card rounded-lg p-4">
        <div className="flex items-center gap-4 mb-4">
          <label className="text-sm font-semibold text-[#e4e4e7]">Policy Type:</label>
          <div className="flex gap-2">
            <button
              onClick={() => setPolicyType('cilium')}
              className={`px-4 py-2 rounded text-sm transition-all cursor-pointer active:scale-95 ${
                policyType === 'cilium'
                  ? 'bg-[#3b82f6] text-white'
                  : 'bg-[#1a1a24] text-[#71717a] hover:text-[#e4e4e7]'
              }`}
              aria-label="Select Cilium Network Policy"
            >
              Cilium Network Policy
            </button>
            <button
              onClick={() => setPolicyType('kyverno')}
              className={`px-4 py-2 rounded text-sm transition-all ${
                policyType === 'kyverno'
                  ? 'bg-[#3b82f6] text-white'
                  : 'bg-[#1a1a24] text-[#71717a] hover:text-[#e4e4e7]'
              }`}
            >
              Kyverno Policy
            </button>
          </div>
        </div>

        {/* Basic Info */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="block text-xs text-[#71717a] mb-1">Policy Name</label>
            <input
              type="text"
              value={policyName}
              onChange={(e) => setPolicyName(e.target.value)}
              className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded text-[#e4e4e7] text-sm"
              placeholder="my-network-policy"
            />
          </div>
          <div>
            <label className="block text-xs text-[#71717a] mb-1">Namespace</label>
            <input
              type="text"
              value={policyNamespace}
              onChange={(e) => setPolicyNamespace(e.target.value)}
              className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded text-[#e4e4e7] text-sm"
              placeholder="default"
            />
          </div>
        </div>

        <div className="mb-4">
          <label className="block text-xs text-[#71717a] mb-1">Description (optional)</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded text-[#e4e4e7] text-sm"
            rows={2}
            placeholder="Policy description..."
          />
        </div>
      </div>

      {/* Service Selection */}
      <div className="card rounded-lg p-4">
        <h3 className="text-sm font-semibold text-[#e4e4e7] mb-4">Select Services</h3>
        {loading ? (
          <div className="text-center py-8">
            <Icon name="loading" className="text-[#3b82f6] animate-pulse text-2xl mb-2" />
            <p className="text-xs text-[#71717a]">Loading services...</p>
          </div>
        ) : (
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {serviceList.length === 0 ? (
              <p className="text-sm text-[#71717a] text-center py-4">No services found</p>
            ) : (
              serviceList.map((service) => (
                <label
                  key={service.key}
                  className="flex items-center gap-3 p-2 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={selectedServices.includes(service.key)}
                    onChange={() => toggleService(service.key, false)}
                    className="w-4 h-4 text-[#3b82f6] bg-[#1a1a24] border-[rgba(255,255,255,0.08)] rounded focus:ring-[#3b82f6]"
                  />
                  <div className="flex-1">
                    <div className="text-sm text-[#e4e4e7]">{service.display}</div>
                    <div className="text-xs text-[#71717a]">{service.namespace}</div>
                  </div>
                </label>
              ))
            )}
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex gap-3">
        <button
          onClick={buildPolicy}
          disabled={building || !policyName}
          className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] disabled:bg-[#1a1a24] disabled:text-[#71717a] disabled:cursor-not-allowed disabled:opacity-50 text-white rounded text-sm transition-all flex items-center gap-2 cursor-pointer active:scale-95"
        >
          <Icon name="execute" size="sm" />
          {building ? 'Building...' : 'Build Policy'}
        </button>
        {generatedYaml && (
          <button
            onClick={exportYaml}
            className="px-4 py-2 bg-[#10b981] hover:bg-[#059669] text-white rounded text-sm transition-all flex items-center gap-2 cursor-pointer active:scale-95"
            aria-label="Export policy YAML"
          >
            <Icon name="output" size="sm" />
            Export YAML
          </button>
        )}
      </div>

      {/* Generated YAML */}
      {generatedYaml && (
        <div className="card rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-sm font-semibold text-[#e4e4e7]">Generated Policy YAML</h3>
            <button
              onClick={() => {
                navigator.clipboard.writeText(generatedYaml);
                toast.success('YAML copied to clipboard');
              }}
              className="text-xs text-[#3b82f6] hover:text-[#2563eb] cursor-pointer transition-all"
              aria-label="Copy YAML to clipboard"
            >
              Copy
            </button>
          </div>
          <pre className="p-3 bg-[#0a0a0f] rounded text-xs text-[#e4e4e7] overflow-x-auto max-h-96 overflow-y-auto">
            {generatedYaml}
          </pre>
        </div>
      )}
    </div>
  );
}

// Test Policy Tab
function TestPolicy({ namespace }: { namespace?: string }) {
  const [policyType, setPolicyType] = useState<'cilium' | 'kyverno'>('cilium');
  const [yamlInput, setYamlInput] = useState('');
  const [testNamespace, setTestNamespace] = useState(namespace || 'default');
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<any>(null);
  const [error, setError] = useState('');

  const testPolicy = async () => {
    if (!yamlInput.trim()) {
      setError('Please enter policy YAML');
      return;
    }

    setTesting(true);
    setError('');
    setTestResult(null);

    try {
      let result;
      if (policyType === 'cilium') {
        result = await apiClient.applyCiliumPolicy(yamlInput, testNamespace);
      } else {
        result = await apiClient.applyKyvernoPolicy(yamlInput, testNamespace);
      }

      setTestResult(result);
      toast.success('Policy applied successfully for testing');
    } catch (error: any) {
      const errorMsg = error.response?.data?.error || error.message || 'Failed to apply policy';
      setError(errorMsg);
      toast.error(errorMsg);
    } finally {
      setTesting(false);
    }
  };

  const deleteTestPolicy = async () => {
    if (!testResult?.name) return;

    try {
      if (policyType === 'cilium') {
        await apiClient.deleteCiliumPolicy(testResult.name, testNamespace);
      } else {
        await apiClient.deleteKyvernoPolicy(testResult.name, testNamespace, false);
      }
      setTestResult(null);
      setYamlInput('');
      setError('');
      toast.success('Test policy deleted successfully');
    } catch (error: any) {
      toast.error(`Failed to delete: ${error.response?.data?.error || error.message}`);
    }
  };

  return (
    <div className="space-y-4">
      <div className="card rounded-lg p-4">
        <div className="flex items-center gap-4 mb-4">
          <label className="text-sm font-semibold text-[#e4e4e7]">Policy Type:</label>
          <div className="flex gap-2">
            <button
              onClick={() => setPolicyType('cilium')}
              className={`px-4 py-2 rounded text-sm transition-all cursor-pointer active:scale-95 ${
                policyType === 'cilium'
                  ? 'bg-[#3b82f6] text-white'
                  : 'bg-[#1a1a24] text-[#71717a] hover:text-[#e4e4e7]'
              }`}
              aria-label="Select Cilium Network Policy"
            >
              Cilium Network Policy
            </button>
            <button
              onClick={() => setPolicyType('kyverno')}
              className={`px-4 py-2 rounded text-sm transition-all ${
                policyType === 'kyverno'
                  ? 'bg-[#3b82f6] text-white'
                  : 'bg-[#1a1a24] text-[#71717a] hover:text-[#e4e4e7]'
              }`}
            >
              Kyverno Policy
            </button>
          </div>
        </div>

        <div className="mb-4">
          <label className="block text-xs text-[#71717a] mb-1">Test Namespace</label>
          <input
            type="text"
            value={testNamespace}
            onChange={(e) => setTestNamespace(e.target.value)}
            className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded text-[#e4e4e7] text-sm"
            placeholder="default"
          />
        </div>

        <div className="mb-4">
          <label className="block text-xs text-[#71717a] mb-1">Policy YAML</label>
          <textarea
            value={yamlInput}
            onChange={(e) => setYamlInput(e.target.value)}
            className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded text-[#e4e4e7] text-sm font-mono"
            rows={15}
            placeholder="Paste your policy YAML here..."
          />
        </div>

        <div className="flex gap-3">
          <button
            onClick={testPolicy}
            disabled={testing || !yamlInput.trim()}
            className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] disabled:bg-[#1a1a24] disabled:text-[#71717a] disabled:cursor-not-allowed disabled:opacity-50 text-white rounded text-sm transition-all flex items-center gap-2 cursor-pointer active:scale-95"
          >
            <Icon name="execute" size="sm" />
            {testing ? 'Applying...' : 'Apply & Test'}
          </button>
          {testResult && (
            <button
              onClick={deleteTestPolicy}
              className="px-4 py-2 bg-[#ef4444] hover:bg-[#dc2626] text-white rounded text-sm transition-all flex items-center gap-2 cursor-pointer active:scale-95"
              aria-label="Delete test policy"
            >
              <Icon name="critical" size="sm" />
              Delete Test Policy
            </button>
          )}
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="card rounded-lg p-4 bg-[#ef4444]/10 border border-[#ef4444]/30">
          <div className="flex items-center gap-2 mb-2">
            <Icon name="critical" className="text-[#ef4444]" size="sm" />
            <h3 className="text-sm font-semibold text-[#ef4444]">Error</h3>
          </div>
          <p className="text-sm text-[#e4e4e7]">{error}</p>
        </div>
      )}

      {/* Test Result */}
      {testResult && (
        <div className="card rounded-lg p-4 bg-[#10b981]/10 border border-[#10b981]/30">
          <div className="flex items-center gap-2 mb-2">
            <Icon name="healthy" className="text-[#10b981]" size="sm" />
            <h3 className="text-sm font-semibold text-[#10b981]">Policy Applied Successfully</h3>
          </div>
          <div className="text-sm text-[#e4e4e7] space-y-1">
            <p><span className="text-[#71717a]">Message:</span> {testResult.message}</p>
            <p><span className="text-[#71717a]">Namespace:</span> {testResult.namespace}</p>
            {testResult.name && (
              <p><span className="text-[#71717a]">Policy Name:</span> {testResult.name}</p>
            )}
          </div>
          <div className="mt-4 p-3 bg-[#1a1a24] rounded text-xs text-[#71717a]">
            <p className="mb-1">✅ Policy has been applied to the cluster for testing.</p>
            <p className="mb-1">⚠️ Remember to delete the test policy when done.</p>
            <p>💡 Use the "Delete Test Policy" button above to remove it.</p>
          </div>
        </div>
      )}
    </div>
  );
}
