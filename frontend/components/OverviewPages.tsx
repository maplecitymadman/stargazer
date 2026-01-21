'use client';

import { Icon } from './SpaceshipIcons';
import { apiClient } from '@/lib/api';
import { useState, useEffect } from 'react';

interface OverviewPageProps {
  section: 'traffic-analysis' | 'network-policies' | 'compliance' | 'troubleshooting';
  namespace?: string;
  onNavigate?: (subsection: string) => void;
}

export default function OverviewPage({ section, namespace, onNavigate }: OverviewPageProps) {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [namespace]);

  const loadData = async () => {
    try {
      if (section === 'traffic-analysis' || section === 'troubleshooting') {
        const topology = await apiClient.getServiceTopology(namespace);
        setData(topology);
      } else if (section === 'compliance') {
        const [score, recommendations] = await Promise.all([
          apiClient.getComplianceScore(namespace),
          apiClient.getRecommendations(namespace),
        ]);
        setData({ score, recommendations });
      } else if (section === 'network-policies') {
        const topology = await apiClient.getServiceTopology(namespace);
        setData(topology);
      }
    } catch (error) {
      console.error('Error loading overview data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading overview...</p>
      </div>
    );
  }

  if (section === 'traffic-analysis') {
    const summary = data?.summary || {};
    return (
      <div className="space-y-6">
        <div className="card rounded-lg p-6">
          <h2 className="text-xl font-semibold text-[#e4e4e7] mb-4">Traffic Analysis Overview</h2>
          <p className="text-sm text-[#71717a] mb-6">
            Analyze network traffic flow, service connectivity, and ingress/egress patterns across your cluster.
          </p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Total Services</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{summary.total_services || 0}</div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Connections</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{summary.total_connections || 0}</div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Blocked</div>
              <div className="text-2xl font-bold text-[#ef4444]">{summary.blocked_connections || 0}</div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Mesh Coverage</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{summary.mesh_coverage || '0%'}</div>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button
              onClick={() => onNavigate?.('topology')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
              aria-label="Navigate to Service Topology"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="network" className="text-[#3b82f6]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Service Topology</h3>
              </div>
              <p className="text-sm text-[#71717a]">Interactive graph visualization of service connections and network policies</p>
            </button>
            <button
              onClick={() => onNavigate?.('path-trace')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="network" className="text-[#8b5cf6]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Path Tracer</h3>
              </div>
              <p className="text-sm text-[#71717a]">Debug connection paths between services and identify blocking policies</p>
            </button>
            <button
              onClick={() => onNavigate?.('ingress')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="network" className="text-[#3b82f6]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Ingress Traffic</h3>
              </div>
              <p className="text-sm text-[#71717a]">View incoming traffic routes, gateways, and ingress policies</p>
            </button>
            <button
              onClick={() => onNavigate?.('egress')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="network" className="text-[#8b5cf6]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Egress Traffic</h3>
              </div>
              <p className="text-sm text-[#71717a]">Monitor outbound connections and external service access</p>
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (section === 'network-policies') {
    const infra = data?.infrastructure || {};
    return (
      <div className="space-y-6">
        <div className="card rounded-lg p-6">
          <h2 className="text-xl font-semibold text-[#e4e4e7] mb-4">Network Policies Overview</h2>
          <p className="text-sm text-[#71717a] mb-6">
            View, build, and test network policies for Kubernetes, Cilium, Istio, and Kyverno.
          </p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">K8s Policies</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{infra.network_policies || 0}</div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Cilium Policies</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{infra.cilium_policies || 0}</div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Istio Policies</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{infra.istio_policies || 0}</div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Kyverno Policies</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">{infra.kyverno_policies || 0}</div>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <button
              onClick={() => onNavigate?.('view')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="scan" className="text-[#10b981]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">View Policies</h3>
              </div>
              <p className="text-sm text-[#71717a]">Browse and inspect existing network policies across all namespaces</p>
            </button>
            <button
              onClick={() => onNavigate?.('build')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="execute" className="text-[#3b82f6]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Build Policy</h3>
              </div>
              <p className="text-sm text-[#71717a]">Create new Cilium or Kyverno policies with an intuitive interface</p>
            </button>
            <button
              onClick={() => onNavigate?.('test')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#3b82f6]/30 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="info" className="text-[#8b5cf6]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Test Policy</h3>
              </div>
              <p className="text-sm text-[#71717a]">Apply and test policies before deploying to production</p>
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (section === 'compliance') {
    const score = data?.score || {};
    const recommendations = data?.recommendations || {};
    return (
      <div className="space-y-6">
        <div className="card rounded-lg p-6">
          <h2 className="text-xl font-semibold text-[#e4e4e7] mb-4">Compliance Overview</h2>
          <p className="text-sm text-[#71717a] mb-6">
            Monitor network security compliance and get actionable recommendations to improve your cluster's security posture.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div className="p-6 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-semibold text-[#e4e4e7]">Compliance Score</h3>
                <div className={`text-3xl font-bold ${
                  (score.score || 0) >= 80 ? 'text-[#10b981]' : (score.score || 0) >= 60 ? 'text-[#f59e0b]' : 'text-[#ef4444]'
                }`}>
                  {score.score || 0}%
                </div>
              </div>
              <p className="text-sm text-[#71717a] mb-2">
                {score.passed || 0} of {score.total || 0} checks passed
              </p>
              <button
                onClick={() => onNavigate?.('score')}
                className="mt-4 px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded text-sm transition-all cursor-pointer active:scale-95"
                aria-label="View compliance score details"
              >
                View Details
              </button>
            </div>
            <div className="p-6 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <h3 className="font-semibold text-[#e4e4e7] mb-4">Recommendations</h3>
              <p className="text-sm text-[#71717a] mb-2">
                {recommendations.count || 0} recommendations available
              </p>
              <button
                onClick={() => onNavigate?.('recommendations')}
                className="mt-4 px-4 py-2 bg-[#10b981] hover:bg-[#059669] text-white rounded text-sm transition-all cursor-pointer active:scale-95"
                aria-label="View recommendations"
              >
                View Recommendations
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (section === 'troubleshooting') {
    const summary = data?.summary || {};
    const connectivity = data?.connectivity || {};
    const servicesWithIssues = Object.values(connectivity).filter((conn: any) =>
      conn.blocked_from && conn.blocked_from.length > 0
    ).length;
    return (
      <div className="space-y-6">
        <div className="card rounded-lg p-6">
          <h2 className="text-xl font-semibold text-[#e4e4e7] mb-4">Troubleshooting Overview</h2>
          <p className="text-sm text-[#71717a] mb-6">
            Identify and resolve network connectivity issues, blocked connections, and service problems.
          </p>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6">
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Blocked Connections</div>
              <div className={`text-2xl font-bold ${summary.blocked_connections > 0 ? 'text-[#ef4444]' : 'text-[#e4e4e7]'}`}>
                {summary.blocked_connections || 0}
              </div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Services with Issues</div>
              <div className={`text-2xl font-bold ${servicesWithIssues > 0 ? 'text-[#f59e0b]' : 'text-[#e4e4e7]'}`}>
                {servicesWithIssues}
              </div>
            </div>
            <div className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
              <div className="text-xs text-[#71717a] mb-1">Total Issues</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {(summary.blocked_connections || 0) + servicesWithIssues}
              </div>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button
              onClick={() => onNavigate?.('blocked')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#ef4444]/30 hover:bg-[#252530] transition-all text-left cursor-pointer"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="critical" className="text-[#ef4444]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Blocked Connections</h3>
              </div>
              <p className="text-sm text-[#71717a]">View and debug connections blocked by network policies</p>
            </button>
            <button
              onClick={() => onNavigate?.('services')}
              className="p-4 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)] hover:border-[#f59e0b]/30 hover:bg-[#252530] transition-all text-left cursor-pointer"
            >
              <div className="flex items-center gap-3 mb-2">
                <Icon name="degraded" className="text-[#f59e0b]" size="sm" />
                <h3 className="font-semibold text-[#e4e4e7]">Services with Issues</h3>
              </div>
              <p className="text-sm text-[#71717a]">Identify services experiencing connectivity problems</p>
            </button>
          </div>
        </div>
      </div>
    );
  }

  return null;
}
