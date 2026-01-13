'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface ServiceTopologyProps {
  namespace?: string;
}

interface ServiceConnection {
  target: string;
  allowed: boolean;
  reason: string;
  via_service_mesh: boolean;
  blocked_by_policy: boolean;
  blocking_policies?: string[];
}

interface ServiceInfo {
  name: string;
  namespace: string;
  type: string;
  cluster_ip: string;
  ports: string[];
  pods: string[];
  pod_count: number;
  healthy_pods: number;
  deployment: string | null;
  has_service_mesh: boolean;
}

interface ConnectivityInfo {
  service: string;
  connections: ServiceConnection[];
  can_reach: string[];
  blocked_from: string[];
}

interface TopologyData {
  namespace: string;
  services: Record<string, ServiceInfo>;
  connectivity: Record<string, ConnectivityInfo>;
  network_policies: number;
  istio_enabled: boolean;
  summary: {
    total_services: number;
    services_with_mesh: number;
    total_connections: number;
    allowed_connections: number;
    blocked_connections: number;
    mesh_coverage: string;
  };
}

export default function ServiceTopology({ namespace }: ServiceTopologyProps) {
  const [topology, setTopology] = useState<TopologyData | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedService, setSelectedService] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'allowed' | 'blocked'>('all');
  const [viewMode, setViewMode] = useState<'list' | 'graph'>('list');
  const [yamlModal, setYamlModal] = useState<{ open: boolean; policyName: string | null; yaml: string | null }>({
    open: false,
    policyName: null,
    yaml: null
  });

  useEffect(() => {
    loadTopology();
  }, [namespace]);

  // Auto-select first service with blocked connections if none selected
  useEffect(() => {
    if (topology && !selectedService) {
      const serviceWithBlocked = Object.values(topology.services).find(service => {
        const connInfo = topology.connectivity[service.name];
        return connInfo && connInfo.blocked_from && connInfo.blocked_from.length > 0;
      });
      if (serviceWithBlocked) {
        setSelectedService(serviceWithBlocked.name);
      }
    }
  }, [topology, selectedService]);

  const loadTopology = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getServiceTopology(namespace);
      if (process.env.NODE_ENV === 'development') {
        console.log('Topology data loaded');
      }
      setTopology(data);
    } catch (error) {
      console.error('Error loading topology:', error);
    } finally {
      setLoading(false);
    }
  };

  const viewPolicyYaml = async (policyName: string) => {
    try {
      const data = await apiClient.getNetworkPolicyYaml(policyName, namespace);
      setYamlModal({
        open: true,
        policyName: data.name,
        yaml: data.yaml
      });
    } catch (error) {
      console.error('Error loading NetworkPolicy YAML:', error);
      alert(`Failed to load NetworkPolicy YAML: ${error}`);
    }
  };

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading service topology...</p>
      </div>
    );
  }

  if (!topology) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="critical" className="text-red-400 text-4xl mb-4" />
        <p className="text-[#71717a]">Failed to load service topology</p>
      </div>
    );
  }

  const services = Object.values(topology.services);
  const selectedConnections = selectedService 
    ? topology.connectivity[selectedService]?.connections || []
    : [];

  const filteredConnections = selectedConnections.filter(conn => {
    if (filter === 'allowed') return conn.allowed;
    if (filter === 'blocked') return !conn.allowed;
    return true;
  });

  // Debug logging (development only)
  if (process.env.NODE_ENV === 'development' && selectedService === 'web-app') {
    console.log('Web-app connectivity:', {
      total: selectedConnections.length,
      allowed: selectedConnections.filter(c => c.allowed).length,
      blocked: selectedConnections.filter(c => !c.allowed).length,
    });
  }

  return (
    <div className="space-y-6">
      {/* Header with Refresh and View Toggle */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-[#e4e4e7]">Service Topology</h2>
        <div className="flex items-center gap-3">
          <div className="flex gap-2 bg-[#1a1a24] rounded-lg p-1">
            <button
              onClick={() => setViewMode('list')}
              className={`px-3 py-1 rounded text-sm transition-all ${
                viewMode === 'list' ? 'bg-[#3b82f6] text-white' : 'text-[#71717a] hover:text-[#e4e4e7]'
              }`}
            >
              List
            </button>
            <button
              onClick={() => setViewMode('graph')}
              className={`px-3 py-1 rounded text-sm transition-all ${
                viewMode === 'graph' ? 'bg-[#3b82f6] text-white' : 'text-[#71717a] hover:text-[#e4e4e7]'
              }`}
            >
              Graph
            </button>
          </div>
          <button
            onClick={loadTopology}
            className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-sm transition-all flex items-center gap-2"
          >
            <Icon name="refresh" className="text-white" size="sm" />
            Refresh
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <div className="card rounded-lg p-4">
          <div className="text-sm text-[#71717a] mb-1">Total Services</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.summary.total_services}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-sm text-[#71717a] mb-1">Total Connections</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">{topology.summary.total_connections}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-sm text-[#71717a] mb-1">Service Mesh</div>
          <div className="text-2xl font-bold text-[#e4e4e7]">
            {topology.summary.services_with_mesh}
            <span className="text-sm text-[#71717a] ml-2">({topology.summary.mesh_coverage})</span>
          </div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-sm text-[#71717a] mb-1">Allowed Connections</div>
          <div className="text-2xl font-bold text-green-400">{topology.summary.allowed_connections}</div>
        </div>
        <div className="card rounded-lg p-4">
          <div className="text-sm text-[#71717a] mb-1">Blocked Connections</div>
          <div className="text-2xl font-bold text-red-400">{topology.summary.blocked_connections}</div>
        </div>
      </div>

      {viewMode === 'graph' ? (
        <TopologyGraph topology={topology} />
      ) : (
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Service List */}
        <div className="lg:col-span-1">
          <div className="card rounded-lg p-5">
            <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#71717a]" size="sm" />
              <span>Services</span>
            </h3>
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {services.map((service) => {
                const connInfo = topology.connectivity[service.name];
                const isSelected = selectedService === service.name;
                
                return (
                  <button
                    key={service.name}
                    onClick={() => setSelectedService(service.name)}
                    className={`w-full text-left p-3 rounded-md transition-all ${
                      isSelected
                        ? 'bg-[#3b82f6] text-white'
                        : connInfo && connInfo.blocked_from && connInfo.blocked_from.length > 0
                        ? 'bg-[#1a1a24] hover:bg-[#252530] text-[#e4e4e7] border border-red-500/30'
                        : 'bg-[#1a1a24] hover:bg-[#252530] text-[#e4e4e7]'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{service.name}</span>
                        {service.has_service_mesh && (
                          <span className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-300">
                            Istio
                          </span>
                        )}
                      </div>
                      <div className="text-xs opacity-75">
                        <span className="text-green-400">{connInfo?.can_reach.length || 0}</span> → <span className="text-red-400">{connInfo?.blocked_from.length || 0}</span> ✗
                      </div>
                    </div>
                    <div className="text-xs mt-1 opacity-75">
                      {service.healthy_pods}/{service.pod_count} pods
                    </div>
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        {/* Connection Details */}
        <div className="lg:col-span-2">
          <div className="card rounded-lg p-5">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
                <Icon name="connections" className="text-[#71717a]" size="sm" />
                <span>
                  {selectedService ? `Connections: ${selectedService}` : 'Select a service'}
                </span>
              </h3>
              {selectedService && (
                <div className="flex gap-2">
                  <button
                    onClick={() => setFilter('all')}
                    className={`px-3 py-1 rounded text-sm ${
                      filter === 'all' ? 'bg-[#3b82f6] text-white' : 'bg-[#1a1a24] text-[#71717a]'
                    }`}
                  >
                    All
                  </button>
                  <button
                    onClick={() => setFilter('allowed')}
                    className={`px-3 py-1 rounded text-sm ${
                      filter === 'allowed' ? 'bg-green-500 text-white' : 'bg-[#1a1a24] text-[#71717a]'
                    }`}
                  >
                    Allowed
                  </button>
                  <button
                    onClick={() => setFilter('blocked')}
                    className={`px-3 py-1 rounded text-sm ${
                      filter === 'blocked' ? 'bg-red-500 text-white' : 'bg-[#1a1a24] text-[#71717a]'
                    }`}
                  >
                    Blocked
                  </button>
                </div>
              )}
            </div>

            {selectedService ? (
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {filteredConnections.length === 0 ? (
                  <div className="text-center py-8 text-[#71717a]">
                    No connections found
                  </div>
                ) : (
                  filteredConnections.map((conn) => {
                    const targetService = topology.services[conn.target];
                    return (
                      <div
                        key={conn.target}
                        className={`p-4 rounded-lg border ${
                          conn.allowed
                            ? 'border-green-500/30 bg-green-500/5'
                            : 'border-red-500/30 bg-red-500/5'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center gap-3">
                            <Icon
                              name={conn.allowed ? "check" : "critical"}
                              className={conn.allowed ? "text-green-400" : "text-red-400"}
                              size="sm"
                            />
                            <span className="font-medium text-[#e4e4e7]">{conn.target}</span>
                            {conn.via_service_mesh && (
                              <span className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-300">
                                via Istio
                              </span>
                            )}
                            {conn.blocked_by_policy && conn.blocking_policies && conn.blocking_policies.length > 0 && (
                              <div className="flex items-center gap-1">
                                {conn.blocking_policies.map((policyName, idx) => (
                                  <button
                                    key={idx}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      viewPolicyYaml(policyName);
                                    }}
                                    className="text-xs px-2 py-0.5 rounded bg-red-500/20 text-red-300 hover:bg-red-500/30 cursor-pointer transition-all"
                                    title={`View ${policyName} YAML`}
                                  >
                                    {policyName}
                                  </button>
                                ))}
                              </div>
                            )}
                          </div>
                          <span
                            className={`text-xs px-2 py-1 rounded ${
                              conn.allowed
                                ? 'bg-green-500/20 text-green-300'
                                : 'bg-red-500/20 text-red-300'
                            }`}
                          >
                            {conn.allowed ? 'ALLOWED' : 'BLOCKED'}
                          </span>
                        </div>
                        <div className="text-sm text-[#71717a] mt-2">
                          {conn.reason}
                        </div>
                        {targetService && (
                          <div className="text-xs text-[#71717a] mt-2">
                            {targetService.healthy_pods}/{targetService.pod_count} pods • {targetService.cluster_ip}
                          </div>
                        )}
                      </div>
                    );
                  })
                )}
              </div>
            ) : (
              <div className="text-center py-12 text-[#71717a]">
                <Icon name="network" className="text-4xl mb-4 opacity-50" />
                <p className="mb-2">Select a service to view its connections</p>
                <p className="text-xs opacity-75">
                  Services with blocked connections are highlighted in the list
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
      )}

      {/* Infrastructure Info */}
      <div className="card rounded-lg p-5">
        <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7]">Infrastructure Status</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="flex items-center gap-2">
            <Icon name={topology.istio_enabled ? "check" : "close"} 
                  className={topology.istio_enabled ? "text-green-400" : "text-gray-500"} />
            <span className="text-sm text-[#71717a]">Istio Service Mesh</span>
          </div>
          <div className="flex items-center gap-2">
            <Icon name={topology.network_policies > 0 ? "check" : "close"}
                  className={topology.network_policies > 0 ? "text-green-400" : "text-gray-500"} />
            <span className="text-sm text-[#71717a]">
              Network Policies ({topology.network_policies})
            </span>
          </div>
        </div>
      </div>

      {/* YAML Viewer Modal */}
      {yamlModal.open && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setYamlModal({ open: false, policyName: null, yaml: null })}>
          <div className="bg-[#1a1a24] rounded-lg p-6 max-w-4xl w-full max-h-[90vh] overflow-auto border border-[#252530]" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-xl font-bold text-[#e4e4e7]">NetworkPolicy: {yamlModal.policyName}</h3>
              <button
                onClick={() => setYamlModal({ open: false, policyName: null, yaml: null })}
                className="text-[#71717a] hover:text-[#e4e4e7]"
              >
                <Icon name="close" size="sm" />
              </button>
            </div>
            <pre className="bg-[#0a0a0f] p-4 rounded-lg overflow-x-auto text-sm text-[#e4e4e7] font-mono">
              {yamlModal.yaml}
            </pre>
          </div>
        </div>
      )}
    </div>
  );
}

// Topology Graph Component
function TopologyGraph({ topology }: { topology: TopologyData | null }) {
  if (!topology) return null;

  const services = Object.values(topology.services);
  const nodeSize = 80;
  const spacing = 200;

  // Simple grid layout for now (can be enhanced with force-directed layout later)
  const rows = Math.ceil(Math.sqrt(services.length));
  const cols = Math.ceil(services.length / rows);

  return (
    <div className="card rounded-lg p-6">
      <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7]">Service Topology Graph</h3>
      <div className="relative" style={{ minHeight: `${rows * spacing}px` }}>
        <svg width="100%" height={rows * spacing} className="overflow-visible">
          {/* Draw connections */}
          {services.map((sourceService, idx) => {
            const sourceConn = topology.connectivity[sourceService.name];
            if (!sourceConn) return null;

            const sourceRow = Math.floor(idx / cols);
            const sourceCol = idx % cols;
            const sourceX = (sourceCol + 1) * spacing;
            const sourceY = (sourceRow + 1) * spacing;

            return sourceConn.connections.map((conn) => {
              const targetIdx = services.findIndex(s => s.name === conn.target);
              if (targetIdx === -1) return null;

              const targetRow = Math.floor(targetIdx / cols);
              const targetCol = targetIdx % cols;
              const targetX = (targetCol + 1) * spacing;
              const targetY = (targetRow + 1) * spacing;

              return (
                <line
                  key={`${sourceService.name}-${conn.target}`}
                  x1={sourceX}
                  y1={sourceY}
                  x2={targetX}
                  y2={targetY}
                  stroke={conn.allowed ? '#22c55e' : '#ef4444'}
                  strokeWidth={2}
                  strokeOpacity={0.3}
                  markerEnd={conn.allowed ? undefined : "url(#arrow-red)"}
                />
              );
            });
          })}

          {/* Draw service nodes */}
          {services.map((service, idx) => {
            const row = Math.floor(idx / cols);
            const col = idx % cols;
            const x = (col + 1) * spacing;
            const y = (row + 1) * spacing;
            const conn = topology.connectivity[service.name];
            const hasBlocked = conn && conn.blocked_from && conn.blocked_from.length > 0;

            return (
              <g key={service.name}>
                <circle
                  cx={x}
                  cy={y}
                  r={nodeSize / 2}
                  fill={hasBlocked ? '#ef4444' : '#22c55e'}
                  fillOpacity={0.2}
                  stroke={hasBlocked ? '#ef4444' : '#22c55e'}
                  strokeWidth={2}
                />
                <text
                  x={x}
                  y={y}
                  textAnchor="middle"
                  dominantBaseline="middle"
                  className="text-xs fill-[#e4e4e7] font-medium"
                >
                  {service.name.length > 12 ? service.name.substring(0, 10) + '...' : service.name}
                </text>
                {service.has_service_mesh && (
                  <circle cx={x + nodeSize/2 - 5} cy={y - nodeSize/2 + 5} r={4} fill="#3b82f6" />
                )}
              </g>
            );
          })}

          {/* Arrow markers */}
          <defs>
            <marker id="arrow-red" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto" markerUnits="strokeWidth">
              <path d="M0,0 L0,6 L9,3 z" fill="#ef4444" />
            </marker>
          </defs>
        </svg>
      </div>
      <div className="mt-4 flex items-center gap-4 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-green-500/20 border border-green-500 rounded"></div>
          <span className="text-[#71717a]">Allowed Connection</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-red-500/20 border border-red-500 rounded"></div>
          <span className="text-[#71717a]">Blocked Connection</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-green-500/20 border border-green-500 rounded"></div>
          <span className="text-[#71717a]">Service Node</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
          <span className="text-[#71717a]">Service Mesh</span>
        </div>
      </div>
    </div>
  );
}
