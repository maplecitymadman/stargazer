'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';
import ReactFlow, {
  Node,
  Edge,
  Controls,
  Background,
  MiniMap,
  Connection,
  useNodesState,
  useEdgesState,
  addEdge,
  MarkerType,
  Position,
  NodeTypes,
  Handle,
} from 'reactflow';
import 'reactflow/dist/style.css';

interface ServiceTopologyProps {
  namespace?: string;
}

interface ServiceConnection {
  target: string;
  allowed: boolean;
  reason: string;
  via_service_mesh: boolean;
  service_mesh_type?: string; // "istio", "cilium", or ""
  blocked_by_policy: boolean;
  blocking_policies?: string[];
  port?: string;
  protocol?: string;
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
  service_mesh_type?: string; // "istio", "cilium", or ""
  has_cilium_proxy: boolean;
}

interface ConnectivityInfo {
  service: string;
  connections: ServiceConnection[];
  can_reach: string[];
  blocked_from: string[];
}

interface InfrastructureInfo {
  cni: string;
  cilium_enabled: boolean;
  istio_enabled: boolean;
  kyverno_enabled: boolean;
  network_policies: number;
  cilium_policies: number;
  istio_policies: number;
  kyverno_policies: number;
}

interface TopologyData {
  namespace: string;
  services: Record<string, ServiceInfo>;
  connectivity: Record<string, ConnectivityInfo>;
  ingress?: {
    gateways: any[];
    kubernetes_ingress: any[];
    routes: any[];
    connections: any[];
  };
  egress?: {
    gateways: any[];
    external_services: any[];
    connections: any[];
    has_egress_gateway: boolean;
    direct_egress: boolean;
  };
  network_policies: NetworkPolicyInfo[];
  cilium_policies: CiliumNetworkPolicyInfo[];
  istio_policies: IstioPolicyInfo[];
  kyverno_policies: KyvernoPolicyInfo[];
  infrastructure: InfrastructureInfo;
  summary: {
    total_services: number;
    services_with_mesh: number;
    total_connections: number;
    allowed_connections: number;
    blocked_connections: number;
    mesh_coverage: string;
    cilium_coverage: string;
    istio_coverage: string;
  };
  hubble_enabled?: boolean;
}

interface NetworkPolicyInfo {
  name: string;
  namespace: string;
  type: string;
  yaml?: string;
}

interface CiliumNetworkPolicyInfo {
  name: string;
  namespace: string;
  type: string;
  yaml?: string;
}

interface IstioPolicyInfo {
  name: string;
  namespace: string;
  type: string;
  yaml?: string;
}

interface KyvernoPolicyInfo {
  name: string;
  namespace: string;
  type: string;
  yaml?: string;
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

  // WebSocket listener for policy changes
  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    const ws = new WebSocket(wsUrl);
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'policy_change') {
          // Reload topology when policy changes
          loadTopology();
        }
      } catch (e) {
        // Ignore parse errors
      }
    };
    
    return () => {
      ws.close();
    };
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

  // Service connectivity metrics available via topology data

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
          <div className="text-xs text-[#71717a] mt-1">
            Istio: {topology.summary.istio_coverage} • Cilium: {topology.summary.cilium_coverage}
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
                          <span className={`text-xs px-2 py-0.5 rounded ${
                            service.service_mesh_type === 'istio' 
                              ? 'bg-blue-500/20 text-blue-300' 
                              : service.service_mesh_type === 'cilium'
                              ? 'bg-purple-500/20 text-purple-300'
                              : 'bg-blue-500/20 text-blue-300'
                          }`}>
                            {service.service_mesh_type === 'istio' ? 'Istio' : 
                             service.service_mesh_type === 'cilium' ? 'Cilium' : 'Mesh'}
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
                              <span className={`text-xs px-2 py-0.5 rounded ${
                                conn.service_mesh_type === 'istio'
                                  ? 'bg-blue-500/20 text-blue-300'
                                  : conn.service_mesh_type === 'cilium'
                                  ? 'bg-purple-500/20 text-purple-300'
                                  : 'bg-blue-500/20 text-blue-300'
                              }`}>
                                via {conn.service_mesh_type === 'istio' ? 'Istio' : 
                                     conn.service_mesh_type === 'cilium' ? 'Cilium' : 'Mesh'}
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

      {/* Enhanced Ingress Info */}
      {topology.ingress && (topology.ingress.gateways?.length > 0 || topology.ingress.kubernetes_ingress?.length > 0) && (
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#3b82f6]" size="sm" />
              Ingress Traffic
            </h3>
            <div className="text-sm text-[#71717a]">
              {topology.ingress.connections?.filter((c: any) => c.allowed).length || 0} / {topology.ingress.connections?.length || 0} allowed
            </div>
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
                      Namespace: {gw.namespace} • Hosts: {gw.hosts?.join(', ') || 'N/A'} • Ports: {gw.ports?.join(', ') || 'N/A'}
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
                    <div className="text-xs text-[#71717a] mb-2">
                      Namespace: {ing.namespace} • Backend: {ing.backend} • TLS: {ing.tls ? 'Yes' : 'No'}
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Hosts: {ing.hosts?.join(', ') || 'N/A'}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Routes */}
          {topology.ingress.routes && topology.ingress.routes.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Routes</h4>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {topology.ingress.routes.map((route: any, idx: number) => (
                  <div key={idx} className={`p-3 rounded border ${
                    route.allowed 
                      ? 'border-[#10b981]/30 bg-[#10b981]/5' 
                      : 'border-[#ef4444]/30 bg-[#ef4444]/5'
                  }`}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">
                        {route.host}{route.path}
                      </span>
                      <span className={`text-xs px-2 py-0.5 rounded ${
                        route.allowed 
                          ? 'bg-[#10b981]/20 text-[#10b981]' 
                          : 'bg-[#ef4444]/20 text-[#ef4444]'
                      }`}>
                        {route.allowed ? 'ALLOWED' : 'BLOCKED'}
                      </span>
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Gateway: {route.gateway} → Service: {route.service} ({route.namespace})
                    </div>
                    {route.blocked_by && route.blocked_by.length > 0 && (
                      <div className="text-xs text-[#f59e0b] mt-1">
                        Blocked by: {route.blocked_by.join(', ')}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Connections */}
          {topology.ingress.connections && topology.ingress.connections.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Connections</h4>
              <div className="space-y-2 max-h-48 overflow-y-auto">
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
                    {conn.policies && conn.policies.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-2">
                        {conn.policies.map((policy: string, pIdx: number) => (
                          <span key={pIdx} className="text-xs px-2 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]">
                            {policy}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Enhanced Egress Info */}
      {topology.egress && (
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#8b5cf6]" size="sm" />
              Egress Traffic
            </h3>
            <div className="text-sm text-[#71717a]">
              {topology.egress.connections?.filter((c: any) => c.allowed).length || 0} / {topology.egress.connections?.length || 0} allowed
            </div>
          </div>

          {/* Gateways */}
          {topology.egress.gateways && topology.egress.gateways.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">Egress Gateways</h4>
              <div className="space-y-2">
                {topology.egress.gateways.map((gw: any, idx: number) => (
                  <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">{gw.name}</span>
                      <span className="text-xs px-2 py-0.5 rounded bg-purple-500/20 text-purple-300">Istio Egress</span>
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Namespace: {gw.namespace} • Type: {gw.type}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* External Services */}
          {topology.egress.external_services && topology.egress.external_services.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-[#e4e4e7] mb-2">External Services</h4>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {topology.egress.external_services.map((ext: any, idx: number) => (
                  <div key={idx} className="p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-sm font-medium text-[#e4e4e7]">{ext.name}</span>
                      <span className="text-xs px-2 py-0.5 rounded bg-purple-500/20 text-purple-300">
                        {ext.type}
                      </span>
                    </div>
                    <div className="text-xs text-[#71717a]">
                      Namespace: {ext.namespace} • Hosts: {ext.hosts?.join(', ') || 'N/A'} • Ports: {ext.ports?.join(', ') || 'N/A'}
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
              <div className="space-y-2 max-h-48 overflow-y-auto">
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
                    {conn.policies && conn.policies.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-2">
                        {conn.policies.map((policy: string, pIdx: number) => (
                          <span key={pIdx} className="text-xs px-2 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]">
                            {policy}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Summary */}
          <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t border-[rgba(255,255,255,0.08)]">
            <div>
              <div className="text-xs text-[#71717a] mb-1">Egress Gateway</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {topology.egress.has_egress_gateway ? 'Yes' : 'No'}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Direct Egress</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {topology.egress.direct_egress ? 'Yes' : 'No'}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">External Services</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {topology.egress.external_services?.length || 0}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Total Connections</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {topology.egress.connections?.length || 0}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Infrastructure Info */}
      <div className="card rounded-lg p-5">
        <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7]">Infrastructure Status</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-4">
          <div className="flex items-center gap-2">
            <Icon name={topology.infrastructure.cni ? "check" : "close"} 
                  className={topology.infrastructure.cni ? "text-green-400" : "text-gray-500"} />
            <div>
              <span className="text-sm text-[#71717a] block">CNI</span>
              <span className="text-xs text-[#a1a1aa]">{topology.infrastructure.cni || "Unknown"}</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon name={topology.infrastructure.cilium_enabled ? "check" : "close"} 
                  className={topology.infrastructure.cilium_enabled ? "text-purple-400" : "text-gray-500"} />
            <div>
              <span className="text-sm text-[#71717a] block">Cilium</span>
              <span className="text-xs text-[#a1a1aa]">{topology.summary.cilium_coverage}</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon name={topology.infrastructure.istio_enabled ? "check" : "close"} 
                  className={topology.infrastructure.istio_enabled ? "text-blue-400" : "text-gray-500"} />
            <div>
              <span className="text-sm text-[#71717a] block">Istio</span>
              <span className="text-xs text-[#a1a1aa]">{topology.summary.istio_coverage}</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon name={topology.infrastructure.kyverno_enabled ? "check" : "close"}
                  className={topology.infrastructure.kyverno_enabled ? "text-green-400" : "text-gray-500"} />
            <div>
              <span className="text-sm text-[#71717a] block">Kyverno</span>
              <span className="text-xs text-[#a1a1aa]">{topology.infrastructure.kyverno_policies} policies</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon name={topology.infrastructure.network_policies > 0 ? "check" : "close"}
                  className={topology.infrastructure.network_policies > 0 ? "text-green-400" : "text-gray-500"} />
            <div>
              <span className="text-sm text-[#71717a] block">K8s Policies</span>
              <span className="text-xs text-[#a1a1aa]">{topology.infrastructure.network_policies}</span>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon name={topology.infrastructure.cilium_policies > 0 ? "check" : "close"}
                  className={topology.infrastructure.cilium_policies > 0 ? "text-purple-400" : "text-gray-500"} />
            <div>
              <span className="text-sm text-[#71717a] block">Cilium Policies</span>
              <span className="text-xs text-[#a1a1aa]">{topology.infrastructure.cilium_policies}</span>
            </div>
          </div>
          {topology.hubble_enabled && (
            <div className="flex items-center gap-2">
              <Icon name="check" className="text-blue-400" />
              <div>
                <span className="text-sm text-[#71717a] block">Hubble</span>
                <span className="text-xs text-[#a1a1aa]">Enabled</span>
              </div>
            </div>
          )}
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

// Custom Service Node Component
function ServiceNode({ data }: { data: any }) {
  const { service, connInfo, hasBlocked, hasIssues } = data;
  
  return (
    <div className={`relative px-4 py-3 rounded-lg border-2 transition-all ${
      hasBlocked 
        ? 'bg-[#ef4444]/10 border-[#ef4444]/50 shadow-lg shadow-[#ef4444]/20' 
        : hasIssues
        ? 'bg-[#f59e0b]/10 border-[#f59e0b]/50'
        : 'bg-[#1a1a24] border-[#3b82f6]/30 hover:border-[#3b82f6]/50'
    }`}>
      <Handle type="target" position={Position.Top} className="!bg-[#71717a]" />
      <Handle type="source" position={Position.Bottom} className="!bg-[#71717a]" />
      
      <div className="flex items-center gap-2 mb-1">
        <Icon 
          name={hasBlocked ? "critical" : "healthy"} 
          className={hasBlocked ? "text-[#ef4444]" : "text-[#10b981]"} 
          size="sm" 
        />
        <div className="font-semibold text-[#e4e4e7] text-sm truncate max-w-[120px]">
          {service.name}
        </div>
        {service.has_service_mesh && (
          <div className={`w-2 h-2 rounded-full ${
            service.service_mesh_type === 'istio' 
              ? 'bg-[#3b82f6]' 
              : service.service_mesh_type === 'cilium'
              ? 'bg-[#8b5cf6]'
              : 'bg-[#3b82f6]'
          }`} title={service.service_mesh_type || 'mesh'} />
        )}
      </div>
      
      <div className="text-xs text-[#71717a] space-y-0.5">
        <div className="flex items-center justify-between">
          <span>{service.namespace}</span>
          <span className="text-[#10b981]">{service.healthy_pods}/{service.pod_count}</span>
        </div>
        {connInfo && (
          <div className="flex items-center gap-1 mt-1">
            <span className="text-[#10b981]">{connInfo.can_reach?.length || 0}</span>
            <span className="text-[#71717a]">→</span>
            <span className="text-[#ef4444]">{connInfo.blocked_from?.length || 0}</span>
          </div>
        )}
      </div>
    </div>
  );
}

// Custom Ingress/Egress Node Component
function GatewayNode({ data }: { data: any }) {
  const { type, name, connections } = data;
  const isIngress = type === 'ingress';
  
  return (
    <div className={`relative px-4 py-3 rounded-lg border-2 bg-gradient-to-br ${
      isIngress 
        ? 'from-[#3b82f6]/20 to-[#3b82f6]/5 border-[#3b82f6]/50' 
        : 'from-[#8b5cf6]/20 to-[#8b5cf6]/5 border-[#8b5cf6]/50'
    }`}>
      <Handle type={isIngress ? "source" : "target"} position={isIngress ? Position.Top : Position.Bottom} className="!bg-[#71717a]" />
      
      <div className="flex items-center gap-2">
        <Icon name="network" className={isIngress ? "text-[#3b82f6]" : "text-[#8b5cf6]"} size="sm" />
        <div className="font-semibold text-[#e4e4e7] text-sm">
          {isIngress ? 'Ingress' : 'Egress'}
        </div>
      </div>
      <div className="text-xs text-[#71717a] mt-1">
        {name} • {connections || 0} routes
      </div>
    </div>
  );
}

const nodeTypes: NodeTypes = {
  service: ServiceNode,
  gateway: GatewayNode,
};

// Topology Graph Component with React Flow
function TopologyGraph({ topology }: { topology: TopologyData | null }) {
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  
  const initialNodes = useMemo(() => {
    if (!topology) return [];
    
    const services = Object.values(topology.services);
    const nodes: Node[] = [];
    const centerX = 500;
    const centerY = 400;
    
    // Create service nodes with circular layout
    services.forEach((service, idx) => {
      const connInfo = topology.connectivity[service.name];
      const hasBlocked = connInfo?.blocked_from && connInfo.blocked_from.length > 0;
      const hasIssues = connInfo && (connInfo.blocked_from?.length > 0 || connInfo.connections?.some((c: ServiceConnection) => !c.allowed));
      
      // Circular layout
      const angle = (idx / services.length) * 2 * Math.PI - Math.PI / 2; // Start from top
      const radius = Math.max(250, Math.sqrt(services.length) * 40);
      const x = centerX + radius * Math.cos(angle);
      const y = centerY + radius * Math.sin(angle);
      
      nodes.push({
        id: service.name,
        type: 'service',
        position: { x, y },
        data: {
          service,
          connInfo,
          hasBlocked,
          hasIssues,
        },
        style: {
          width: 160,
        },
      });
    });
    
    // Add ingress gateway at top center
    if (topology.ingress && (topology.ingress.gateways?.length > 0 || topology.ingress.kubernetes_ingress?.length > 0)) {
      const ingressConnections = topology.ingress.connections?.length || 0;
      nodes.push({
        id: 'ingress-gateway',
        type: 'gateway',
        position: { x: centerX, y: 50 },
        data: {
          type: 'ingress',
          name: 'Ingress Gateway',
          connections: ingressConnections,
        },
        style: {
          width: 160,
        },
      });
    }
    
    // Add egress gateway at bottom center
    if (topology.egress && topology.egress.gateways?.length > 0) {
      const egressConnections = topology.egress.connections?.length || 0;
      nodes.push({
        id: 'egress-gateway',
        type: 'gateway',
        position: { x: centerX, y: 750 },
        data: {
          type: 'egress',
          name: 'Egress Gateway',
          connections: egressConnections,
        },
        style: {
          width: 160,
        },
      });
    }
    
    return nodes;
  }, [topology]);
  
  const initialEdges = useMemo(() => {
    if (!topology) return [];
    
    const services = Object.values(topology.services);
    const edges: Edge[] = [];
    
    // Create edges (connections)
    services.forEach((sourceService) => {
      const sourceConn = topology.connectivity[sourceService.name];
      if (!sourceConn) return;
      
      sourceConn.connections.forEach((conn: ServiceConnection) => {
        edges.push({
          id: `${sourceService.name}-${conn.target}`,
          source: sourceService.name,
          target: conn.target,
          type: 'smoothstep',
          animated: conn.allowed && conn.via_service_mesh,
          style: {
            stroke: conn.allowed ? '#10b981' : '#ef4444',
            strokeWidth: conn.blocked_by_policy ? 3 : 2,
            strokeDasharray: conn.allowed ? '0' : '5,5',
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: conn.allowed ? '#10b981' : '#ef4444',
            width: 20,
            height: 20,
          },
          label: conn.via_service_mesh ? (conn.service_mesh_type || 'mesh') : undefined,
          labelStyle: {
            fill: conn.allowed ? '#10b981' : '#ef4444',
            fontSize: 10,
            fontWeight: 600,
          },
          labelBgStyle: {
            fill: '#0a0a0f',
            fillOpacity: 0.9,
          },
        });
      });
    });
    
    // Add ingress connections
    if (topology.ingress?.connections) {
      topology.ingress.connections.forEach((conn: any, idx: number) => {
        const targetService = services.find(s => s.name === conn.to);
        if (targetService) {
          edges.push({
            id: `ingress-${idx}`,
            source: 'ingress-gateway',
            target: targetService.name,
            type: 'smoothstep',
            style: {
              stroke: conn.allowed ? '#3b82f6' : '#ef4444',
              strokeWidth: 2,
            },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: conn.allowed ? '#3b82f6' : '#ef4444',
              width: 20,
              height: 20,
            },
          });
        }
      });
    }
    
    // Add egress connections
    if (topology.egress?.connections) {
      topology.egress.connections.forEach((conn: any, idx: number) => {
        const sourceService = services.find(s => s.name === conn.from);
        if (sourceService) {
          edges.push({
            id: `egress-${idx}`,
            source: sourceService.name,
            target: 'egress-gateway',
            type: 'smoothstep',
            style: {
              stroke: conn.allowed ? '#8b5cf6' : '#ef4444',
              strokeWidth: 2,
            },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: conn.allowed ? '#8b5cf6' : '#ef4444',
              width: 20,
              height: 20,
            },
          });
        }
      });
    }
    
    return edges;
  }, [topology]);
  
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  
  useEffect(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);
  
  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    setSelectedNode(prev => prev === node.id ? null : node.id);
  }, []);
  
  if (!topology) return null;
  
  return (
    <div className="card rounded-lg p-0 overflow-hidden">
      <div className="p-4 border-b border-[rgba(255,255,255,0.08)] flex items-center justify-between">
        <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
          <Icon name="network" className="text-[#3b82f6]" size="sm" />
          Network Topology
        </h3>
        <div className="flex items-center gap-4 text-xs text-[#71717a]">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-[#10b981]"></div>
            <span>Allowed</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-[#ef4444]"></div>
            <span>Blocked</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-[#3b82f6]"></div>
            <span>Mesh</span>
          </div>
        </div>
      </div>
      
      <div className="h-[600px] bg-[#0a0a0f]">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={onNodeClick}
          nodeTypes={nodeTypes}
          fitView
          minZoom={0.2}
          maxZoom={2}
          defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
        >
          <Background color="#1a1a24" gap={20} size={1} />
          <Controls 
            className="bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-lg"
            showInteractive={false}
          />
          <MiniMap 
            className="bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-lg"
            nodeColor={(node) => {
              if (node.data?.hasBlocked) return '#ef4444';
              if (node.data?.hasIssues) return '#f59e0b';
              return '#10b981';
            }}
            maskColor="rgba(0, 0, 0, 0.6)"
          />
        </ReactFlow>
      </div>
      
      {selectedNode && selectedNode !== 'ingress-gateway' && selectedNode !== 'egress-gateway' && (
        <div className="p-4 border-t border-[rgba(255,255,255,0.08)] bg-[#1a1a24]">
          <div className="text-sm font-semibold text-[#e4e4e7] mb-2">
            {selectedNode} Details
          </div>
          <div className="text-xs text-[#71717a] space-y-1">
            {topology.services[selectedNode] && (
              <>
                <div>Namespace: {topology.services[selectedNode].namespace}</div>
                <div>Type: {topology.services[selectedNode].type}</div>
                <div>Cluster IP: {topology.services[selectedNode].cluster_ip}</div>
                {topology.connectivity[selectedNode] && (
                  <div>
                    Connections: {topology.connectivity[selectedNode].can_reach?.length || 0} allowed, {topology.connectivity[selectedNode].blocked_from?.length || 0} blocked
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
