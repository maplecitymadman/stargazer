'use client';

import { useState, useEffect } from 'react';
import { ClusterHealth, apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface HealthMetricsProps {
  health: ClusterHealth;
  namespace?: string;
  onEventsClick?: () => void;
}

export default function HealthMetrics({ health, namespace, onEventsClick }: HealthMetricsProps) {
  const [topology, setTopology] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTopology();
  }, [namespace]);

  const loadTopology = async () => {
    try {
      const data = await apiClient.getServiceTopology(namespace).catch(() => null);
      setTopology(data);
    } catch (error) {
      console.error('Error loading topology:', error);
    } finally {
      setLoading(false);
    }
  };

  const infra = topology?.infrastructure || {};
  const summary = topology?.summary || {};
  
  // Calculate policy coverage
  const totalServices = summary.total_services || 0;
  const servicesWithPolicies = Object.values(topology?.services || {}).filter((service: any) => {
    // Check if service has any policy coverage
    const hasK8sPolicy = (topology?.network_policies || []).some((np: any) => np.namespace === service.namespace);
    const hasCiliumPolicy = (topology?.cilium_policies || []).some((cp: any) => 
      cp.namespace === service.namespace || cp.namespace === ''
    );
    return hasK8sPolicy || hasCiliumPolicy;
  }).length;
  
  const policyCoverage = totalServices > 0 
    ? Math.round((servicesWithPolicies / totalServices) * 100)
    : 0;

  // Total policies count
  const totalPolicies = (infra.network_policies || 0) + 
                        (infra.cilium_policies || 0) + 
                        (infra.istio_policies || 0) + 
                        (infra.kyverno_policies || 0);

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
      {/* Overall Status */}
      <div className={`rounded-lg p-5 card ${
        health.overall_health === 'healthy' 
          ? 'border-[#10b981]/30 bg-[#10b981]/5' 
          : 'border-[#f59e0b]/30 bg-[#f59e0b]/5'
      }`}>
        <div className="text-xs text-[#71717a] mb-2 tracking-wide">Status</div>
        <div className="text-xl font-semibold mt-2 flex items-center gap-2">
          <Icon 
            name={health.overall_health === 'healthy' ? 'healthy' : 'degraded'} 
            className={health.overall_health === 'healthy' ? 'text-[#10b981]' : 'text-[#f59e0b]'} 
            size="md"
          />
          <span className={health.overall_health === 'healthy' ? 'text-[#10b981]' : 'text-[#f59e0b]'}>
            {health.overall_health === 'healthy' ? 'Operational' : 'Degraded'}
          </span>
        </div>
        <div className="text-xs text-[#71717a] mt-2">
          {health.overall_health === 'healthy' ? 'All systems operational' : 'Issues detected'}
        </div>
      </div>

      {/* Network Policies */}
      <div 
        className="card rounded-lg p-5"
      >
        <div className="text-xs text-[#71717a] mb-2 flex items-center gap-2">
          <Icon name="scan" className="text-[#71717a]" size="sm" />
          <span>Network Policies</span>
        </div>
        <div className="text-2xl font-semibold mt-2 text-[#e4e4e7]">
          {loading ? '...' : totalPolicies}
        </div>
        <div className="text-xs mt-2 text-[#71717a]">
          {loading ? 'Loading...' : (
            <>
              {infra.network_policies || 0} K8s • {infra.cilium_policies || 0} Cilium • {infra.istio_policies || 0} Istio
            </>
          )}
        </div>
        <div className="mt-3 flex gap-1 flex-wrap">
          {!loading && infra.network_policies > 0 && (
            <span className="px-2 py-0.5 rounded text-xs bg-[#3b82f6]/10 text-[#3b82f6] border border-[#3b82f6]/20">
              K8s
            </span>
          )}
          {!loading && infra.cilium_policies > 0 && (
            <span className="px-2 py-0.5 rounded text-xs bg-[#8b5cf6]/10 text-[#8b5cf6] border border-[#8b5cf6]/20">
              Cilium
            </span>
          )}
          {!loading && infra.istio_policies > 0 && (
            <span className="px-2 py-0.5 rounded text-xs bg-[#10b981]/10 text-[#10b981] border border-[#10b981]/20">
              Istio
            </span>
          )}
        </div>
      </div>

      {/* Services Monitored */}
      <div 
        className="card rounded-lg p-5"
      >
        <div className="text-xs text-[#71717a] mb-2 flex items-center gap-2">
          <Icon name="network" className="text-[#71717a]" size="sm" />
          <span>Services</span>
        </div>
        <div className="text-2xl font-semibold mt-2 text-[#e4e4e7]">
          {loading ? '...' : summary.total_services || 0}
        </div>
        <div className={`text-xs mt-2 font-medium ${
          policyCoverage >= 80 
            ? 'text-[#10b981]' 
            : policyCoverage >= 50 
            ? 'text-[#f59e0b]' 
            : 'text-[#ef4444]'
        }`}>
          {loading ? 'Loading...' : `${policyCoverage}% policy coverage`}
        </div>
        <div className="mt-3 h-1 bg-[#1a1a24] rounded-full overflow-hidden">
          <div 
            className={`h-full transition-all duration-500 ${
              policyCoverage >= 80 ? 'bg-[#10b981]' : policyCoverage >= 50 ? 'bg-[#f59e0b]' : 'bg-[#ef4444]'
            }`}
            style={{ width: `${policyCoverage}%` }}
          ></div>
        </div>
      </div>

      {/* Events */}
      <div 
        className="card card-hover rounded-lg p-5 cursor-pointer"
        onClick={onEventsClick || (() => {})}
      >
        <div className="text-xs text-[#71717a] mb-2 flex items-center gap-2">
          <Icon name="events" className="text-[#71717a]" size="sm" />
          <span>Events</span>
        </div>
        <div className="text-2xl font-semibold mt-2 text-[#e4e4e7]">
          {health.events.warnings + health.events.errors}
        </div>
        <div className="text-xs mt-2 text-[#71717a]">
          Total events
        </div>
        <div className="mt-3 flex gap-2 flex-wrap">
          {health.events.warnings > 0 && (
            <span className="px-2 py-1 rounded text-xs font-medium bg-[#f59e0b]/10 text-[#f59e0b] border border-[#f59e0b]/20">
              {health.events.warnings}W
            </span>
          )}
          {health.events.errors > 0 && (
            <span className="px-2 py-1 rounded text-xs font-medium bg-[#ef4444]/10 text-[#ef4444] border border-[#ef4444]/20">
              {health.events.errors}E
            </span>
          )}
          {health.events.warnings === 0 && health.events.errors === 0 && (
            <span className="px-2 py-1 rounded text-xs font-medium bg-[#10b981]/10 text-[#10b981] border border-[#10b981]/20">
              Clear
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
