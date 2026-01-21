'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import ServiceTopology from './ServiceTopology';
import PathTracer from './PathTracer';
import { Icon } from './SpaceshipIcons';

interface TrafficAnalysisPageProps {
  subsection?: string;
  namespace?: string;
}

export default function TrafficAnalysisPage({ subsection, namespace }: TrafficAnalysisPageProps) {
  const [activeTab, setActiveTab] = useState(subsection || 'topology');

  useEffect(() => {
    if (subsection) {
      setActiveTab(subsection);
    }
  }, [subsection]);

  const tabs = [
    { id: 'topology', label: 'Service Topology', icon: 'network' as const },
    { id: 'path-trace', label: 'Path Tracer', icon: 'network' as const },
    { id: 'ingress', label: 'Ingress Traffic', icon: 'network' as const },
    { id: 'egress', label: 'Egress Traffic', icon: 'network' as const },
  ];

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
        {activeTab === 'topology' && <ServiceTopology namespace={namespace} />}
        {activeTab === 'path-trace' && <PathTracer namespace={namespace} />}
        {activeTab === 'ingress' && <IngressView namespace={namespace} />}
        {activeTab === 'egress' && <EgressView namespace={namespace} />}
      </div>
    </div>
  );
}

// Ingress View - shows detailed ingress info from ServiceTopology
function IngressView({ namespace }: { namespace?: string }) {
  const [topology, setTopology] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [namespace]);

  const loadData = async () => {
    try {
      const data = await apiClient.getServiceTopology(namespace);
      setTopology(data);
    } catch (error) {
      console.error('Error loading ingress data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="card rounded-lg p-4">
              <div className="h-3 bg-[#1a1a24] rounded w-1/4 mb-3 animate-pulse" />
              <div className="h-8 bg-[#1a1a24] rounded w-1/3 mb-2 animate-pulse" />
            </div>
          ))}
        </div>
        <div className="card rounded-lg p-8">
          <div className="h-6 bg-[#1a1a24] rounded w-1/3 mb-4 animate-pulse" />
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-16 bg-[#1a1a24] rounded animate-pulse" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (!topology?.ingress) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="info" className="text-[#71717a] text-4xl mb-4" />
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">No Ingress Configuration Found</h3>
        <p className="text-sm text-[#71717a] mb-4">No ingress gateways or routes detected in this namespace</p>
        <div className="text-xs text-[#71717a] space-y-1">
          <p>• Check if Istio Gateway or Kubernetes Ingress is configured</p>
          <p>• Verify you're viewing the correct namespace</p>
          <p>• Ingress may be configured in a different namespace</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Gateways</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">
            {(topology.ingress.gateways?.length || 0) + (topology.ingress.kubernetes_ingress?.length || 0)}
          </div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Routes</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.ingress.routes?.length || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Connections</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.ingress.connections?.length || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Blocked</div>
          <div className={`text-2xl font-bold ${
            (topology.ingress.connections?.filter((c: any) => !c.allowed).length || 0) > 0
              ? 'text-[#ef4444]'
              : 'text-[#10b981]'
          }`}>
            {topology.ingress.connections?.filter((c: any) => !c.allowed).length || 0}
          </div>
        </div>
      </div>

      {/* Detailed Ingress Info - reuse from ServiceTopology */}
      {topology.ingress && (topology.ingress.gateways?.length > 0 || topology.ingress.kubernetes_ingress?.length > 0) && (
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#3b82f6]" size="sm" />
              Ingress Traffic Details
            </h3>
          </div>

          {/* Gateways */}
          {topology.ingress.gateways && topology.ingress.gateways.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Istio Gateways</h4>
              <div className="space-y-2">
                {topology.ingress.gateways.map((gw: any, idx: number) => (
                  <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">{gw.name}</span>
                      <span className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-300">Istio</span>
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Namespace: {gw.namespace} • Hosts: {gw.hosts?.join(', ') || 'N/A'}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Kubernetes Ingress */}
          {topology.ingress.kubernetes_ingress && topology.ingress.kubernetes_ingress.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Kubernetes Ingress</h4>
              <div className="space-y-2">
                {topology.ingress.kubernetes_ingress.map((ing: any, idx: number) => (
                  <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">{ing.name}</span>
                      <span className="text-xs px-2 py-0.5 rounded bg-green-500/20 text-green-300">
                        {ing.class || 'nginx'}
                      </span>
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Namespace: {ing.namespace} • Backend: {ing.backend} • TLS: {ing.tls ? 'Yes' : 'No'}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Connections */}
          {topology.ingress.connections && topology.ingress.connections.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Connections</h4>
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {topology.ingress.connections.map((conn: any, idx: number) => (
                  <div key={idx} className={`p-3 rounded border ${
                    conn.allowed
                      ? 'border-[#10b981]/30 bg-[#10b981]/5'
                      : 'border-[#ef4444]/30 bg-[#ef4444]/5'
                  }`}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">
                        {conn.from} → {conn.to}
                      </span>
                      <span className={`text-xs px-2 py-0.5 rounded ${
                        conn.allowed
                          ? 'bg-[#10b981]/20 text-[#10b981]'
                          : 'bg-[#ef4444]/20 text-[#ef4444]'
                      }`}>
                        {conn.allowed ? 'ALLOWED' : 'BLOCKED'}
                      </span>
                    </div>
                    <div className="text-xs text-[#71717a]">{conn.reason}</div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// Egress View
function EgressView({ namespace }: { namespace?: string }) {
  const [topology, setTopology] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [namespace]);

  const loadData = async () => {
    try {
      const data = await apiClient.getServiceTopology(namespace);
      setTopology(data);
    } catch (error) {
      console.error('Error loading egress data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="card rounded-lg p-4">
              <div className="h-3 bg-[#1a1a24] rounded w-1/4 mb-3 animate-pulse" />
              <div className="h-8 bg-[#1a1a24] rounded w-1/3 mb-2 animate-pulse" />
            </div>
          ))}
        </div>
        <div className="card rounded-lg p-8">
          <div className="h-6 bg-[#1a1a24] rounded w-1/3 mb-4 animate-pulse" />
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="h-16 bg-[#1a1a24] rounded animate-pulse" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (!topology?.egress) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="info" className="text-[#71717a] text-4xl mb-4" />
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">No Egress Configuration Found</h3>
        <p className="text-sm text-[#71717a] mb-4">No egress gateways or external service entries detected</p>
        <div className="text-xs text-[#71717a] space-y-1">
          <p>• Services may be accessing external resources directly</p>
          <p>• Check if Istio EgressGateway is configured</p>
          <p>• Verify ServiceEntry resources exist for external services</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Gateways</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.egress.gateways?.length || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">External Services</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.egress.external_services?.length || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Connections</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.egress.connections?.length || 0}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-xs text-[#71717a] mb-1">Blocked</div>
          <div className={`text-2xl font-bold ${
            (topology.egress.connections?.filter((c: any) => !c.allowed).length || 0) > 0
              ? 'text-[#ef4444]'
              : 'text-[#10b981]'
          }`}>
            {topology.egress.connections?.filter((c: any) => !c.allowed).length || 0}
          </div>
        </div>
      </div>

      {/* Detailed Egress Info */}
      {topology.egress && (
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#8b5cf6]" size="sm" />
              Egress Traffic Details
            </h3>
          </div>

          {/* External Services */}
          {topology.egress.external_services && topology.egress.external_services.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">External Services</h4>
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {topology.egress.external_services.map((ext: any, idx: number) => (
                  <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">{ext.name}</span>
                      <span className="text-xs px-2 py-0.5 rounded bg-purple-500/20 text-purple-300">
                        {ext.type}
                      </span>
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Namespace: {ext.namespace} • Hosts: {ext.hosts?.join(', ') || 'N/A'}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Connections */}
          {topology.egress.connections && topology.egress.connections.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Egress Connections</h4>
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {topology.egress.connections.map((conn: any, idx: number) => (
                  <div key={idx} className={`p-3 rounded border ${
                    conn.allowed
                      ? 'border-[#10b981]/30 bg-[#10b981]/5'
                      : 'border-[#ef4444]/30 bg-[#ef4444]/5'
                  }`}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">
                        {conn.from} → {conn.to}
                      </span>
                      <span className={`text-xs px-2 py-0.5 rounded ${
                        conn.allowed
                          ? 'bg-[#10b981]/20 text-[#10b981]'
                          : 'bg-[#ef4444]/20 text-[#ef4444]'
                      }`}>
                        {conn.allowed ? 'ALLOWED' : 'BLOCKED'}
                      </span>
                    </div>
                    <div className="text-xs text-[#71717a]">{conn.reason}</div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
