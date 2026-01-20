'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

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
    { id: 'view', label: 'View Policies', icon: 'scan' as const },
    { id: 'build', label: 'Build Policy', icon: 'execute' as const },
    { id: 'test', label: 'Test Policy', icon: 'info' as const },
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
            className={`px-4 py-2 text-sm font-medium transition-all border-b-2 whitespace-nowrap ${
              activeTab === tab.id
                ? 'text-[#3b82f6] border-[#3b82f6]'
                : 'text-[#71717a] border-transparent hover:text-[#e4e4e7]'
            }`}
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
  if (!topology) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="info" className="text-[#71717a] text-4xl mb-4" />
        <p className="text-[#71717a]">No policy data available</p>
      </div>
    );
  }

  const infra = topology.infrastructure || {};

  return (
    <div className="space-y-4">
      {/* Policy Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">K8s NetworkPolicies</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{infra.network_policies || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Cilium Policies</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{infra.cilium_policies || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Istio Policies</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{infra.istio_policies || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Kyverno Policies</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{infra.kyverno_policies || 0}</div>
        </div>
      </div>

      {/* Policy Lists */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* K8s NetworkPolicies */}
        {topology.network_policies && topology.network_policies.length > 0 && (
          <div className="card rounded-lg p-5">
            <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7]">Kubernetes NetworkPolicies</h3>
            <div className="space-y-2">
              {topology.network_policies.map((policy: any, idx: number) => (
                <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                  <div className="text-sm font-medium text-[#e4e4e7]">{policy.name}</div>
                  <div className="text-xs text-[#71717a]">Namespace: {policy.namespace}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Cilium Policies */}
        {topology.cilium_policies && topology.cilium_policies.length > 0 && (
          <div className="card rounded-lg p-5">
            <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7]">Cilium NetworkPolicies</h3>
            <div className="space-y-2">
              {topology.cilium_policies.map((policy: any, idx: number) => (
                <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                  <div className="text-sm font-medium text-[#e4e4e7]">{policy.name}</div>
                  <div className="text-xs text-[#71717a]">Namespace: {policy.namespace}</div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Build Policy Tab (placeholder for now)
function BuildPolicy({ namespace }: { namespace?: string }) {
  return (
    <div className="card rounded-lg p-8 text-center">
      <Icon name="execute" className="text-[#3b82f6] text-4xl mb-4" />
      <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">Policy Builder</h3>
      <p className="text-sm text-[#71717a] mb-4">
        Build Cilium and Kyverno policies with an intuitive interface
      </p>
      <p className="text-xs text-[#71717a]">Coming soon - will integrate with existing policy building features</p>
    </div>
  );
}

// Test Policy Tab (placeholder for now)
function TestPolicy({ namespace }: { namespace?: string }) {
  return (
    <div className="card rounded-lg p-8 text-center">
      <Icon name="info" className="text-[#3b82f6] text-4xl mb-4" />
      <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">Policy Testing</h3>
      <p className="text-sm text-[#71717a] mb-4">
        Test policies before applying them to production
      </p>
      <p className="text-xs text-[#71717a]">Coming soon - will integrate with existing policy testing features</p>
    </div>
  );
}
