'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';
import RecommendationsPanel from './RecommendationsPanel';

interface DashboardProps {
  namespace?: string;
  onNavigateToTopology?: () => void;
  onNavigateToService?: (serviceName: string) => void;
}

interface ServiceIssue {
  name: string;
  namespace: string;
  blockedConnections: number;
  totalConnections: number;
  blockingPolicies: string[];
  canReach: string[];
  blockedFrom: string[];
}

interface BlockedConnection {
  from: string;
  to: string;
  reason: string;
  policies: string[];
  type: 'ingress' | 'egress' | 'service';
}

export default function Dashboard({ namespace, onNavigateToTopology, onNavigateToService }: DashboardProps) {
  const [topology, setTopology] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [servicesWithIssues, setServicesWithIssues] = useState<ServiceIssue[]>([]);
  const [blockedConnections, setBlockedConnections] = useState<BlockedConnection[]>([]);
  const [autoRefresh, setAutoRefresh] = useState(true);

  useEffect(() => {
    loadDashboardData();
    
    // Auto-refresh every 10 seconds if enabled
    if (autoRefresh) {
      const interval = setInterval(loadDashboardData, 10000);
      return () => clearInterval(interval);
    }
  }, [namespace, autoRefresh]);

  const loadDashboardData = async () => {
    try {
      const data = await apiClient.getServiceTopology(namespace);
      setTopology(data);
      
      // Extract services with issues
      const issues: ServiceIssue[] = [];
      const blocked: BlockedConnection[] = [];
      
      // Service-to-service blocked connections
      Object.entries(data.connectivity || {}).forEach(([serviceKey, connInfo]: [string, any]) => {
        if (connInfo.blocked_from && connInfo.blocked_from.length > 0) {
          const [ns, name] = serviceKey.includes('/') ? serviceKey.split('/') : [namespace || 'default', serviceKey];
          issues.push({
            name,
            namespace: ns,
            blockedConnections: connInfo.blocked_from.length,
            totalConnections: connInfo.connections?.length || 0,
            blockingPolicies: Array.from(new Set(
              connInfo.connections
                ?.filter((c: any) => !c.allowed && c.blocking_policies)
                .flatMap((c: any) => c.blocking_policies || []) || []
            )),
            canReach: connInfo.can_reach || [],
            blockedFrom: connInfo.blocked_from || [],
          });

          // Add blocked connections to list
          connInfo.connections?.forEach((conn: any) => {
            if (!conn.allowed) {
              blocked.push({
                from: serviceKey,
                to: conn.target,
                reason: conn.reason,
                policies: conn.blocking_policies || [],
                type: 'service',
              });
            }
          });
        }
      });

      // Ingress blocked connections
      if (data.ingress?.connections) {
        data.ingress.connections.forEach((conn: any) => {
          if (!conn.allowed) {
            blocked.push({
              from: conn.from || 'ingress-gateway',
              to: conn.to,
              reason: conn.reason,
              policies: conn.policies || [],
              type: 'ingress',
            });
          }
        });
      }

      // Egress blocked connections
      if (data.egress?.connections) {
        data.egress.connections.forEach((conn: any) => {
          if (!conn.allowed) {
            blocked.push({
              from: conn.from,
              to: conn.to || 'external',
              reason: conn.reason,
              policies: conn.policies || [],
              type: 'egress',
            });
          }
        });
      }

      setServicesWithIssues(issues);
      setBlockedConnections(blocked);
    } catch (error) {
      console.error('Error loading dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading network status...</p>
      </div>
    );
  }

  if (!topology) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="critical" className="text-red-400 text-4xl mb-4" />
        <p className="text-[#71717a]">Failed to load network data</p>
        <button
          onClick={loadDashboardData}
          className="mt-4 px-4 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  const summary = topology.summary || {
    total_services: 0,
    total_connections: 0,
    allowed_connections: 0,
    blocked_connections: 0,
  };

  const infra = topology.infrastructure || {};
  const connectionHealthPercent = summary.total_connections > 0
    ? Math.round((summary.allowed_connections / summary.total_connections) * 100)
    : 100;

  const hasIssues = servicesWithIssues.length > 0 || blockedConnections.length > 0;
  const ingressBlocked = topology.ingress?.connections?.filter((c: any) => !c.allowed).length || 0;
  const egressBlocked = topology.egress?.connections?.filter((c: any) => !c.allowed).length || 0;

  return (
    <div className="space-y-6">
      {/* Alert Banner */}
      {hasIssues ? (
        <div className={`card rounded-lg p-4 border-l-4 ${
          connectionHealthPercent < 50 ? 'border-[#ef4444] bg-[#ef4444]/5' :
          connectionHealthPercent < 80 ? 'border-[#f59e0b] bg-[#f59e0b]/5' :
          'border-[#10b981] bg-[#10b981]/5'
        }`}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Icon 
                name={connectionHealthPercent < 50 ? "critical" : connectionHealthPercent < 80 ? "degraded" : "healthy"}
                className={connectionHealthPercent < 50 ? "text-[#ef4444]" : connectionHealthPercent < 80 ? "text-[#f59e0b]" : "text-[#10b981]"}
                size="md"
              />
              <div>
                <div className="font-semibold text-[#e4e4e7]">
                  {connectionHealthPercent < 50 ? 'Critical Network Issues Detected' :
                   connectionHealthPercent < 80 ? 'Network Degradation Detected' :
                   'Network Health Issues'}
                </div>
                <div className="text-sm text-[#71717a]">
                  {servicesWithIssues.length} service{servicesWithIssues.length !== 1 ? 's' : ''} with blocked connections • {blockedConnections.length} total blocked connection{blockedConnections.length !== 1 ? 's' : ''}
                </div>
              </div>
            </div>
            <button
              onClick={() => onNavigateToTopology?.()}
              className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-md text-sm font-medium transition-colors flex items-center gap-2"
            >
              <Icon name="network" size="sm" />
              Investigate
            </button>
          </div>
        </div>
      ) : (
        <div className="card rounded-lg p-4 border-l-4 border-[#10b981] bg-[#10b981]/5">
          <div className="flex items-center gap-3">
            <Icon name="healthy" className="text-[#10b981]" size="md" />
            <div>
              <div className="font-semibold text-[#e4e4e7]">Network Health: Operational</div>
              <div className="text-sm text-[#71717a]">
                All connections are allowed • {summary.total_services} services healthy
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Key Metrics - Top Row */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {/* Connection Health */}
        <div className={`card rounded-lg p-5 ${
          connectionHealthPercent === 100 ? 'border-[#10b981]/30' :
          connectionHealthPercent >= 80 ? 'border-[#f59e0b]/30' :
          'border-[#ef4444]/30'
        }`}>
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Connection Health</div>
            <Icon 
              name={connectionHealthPercent === 100 ? "healthy" : connectionHealthPercent >= 80 ? "degraded" : "critical"} 
              className={connectionHealthPercent === 100 ? "text-[#10b981]" : connectionHealthPercent >= 80 ? "text-[#f59e0b]" : "text-[#ef4444]"} 
              size="sm" 
            />
          </div>
          <div className={`text-3xl font-bold ${
            connectionHealthPercent === 100 ? 'text-[#10b981]' :
            connectionHealthPercent >= 80 ? 'text-[#f59e0b]' :
            'text-[#ef4444]'
          }`}>
            {connectionHealthPercent}%
          </div>
          <div className="mt-2 h-2 bg-[#1a1a24] rounded-full overflow-hidden">
            <div 
              className={`h-full transition-all ${
                connectionHealthPercent === 100 ? 'bg-[#10b981]' : 
                connectionHealthPercent >= 80 ? 'bg-[#f59e0b]' : 'bg-[#ef4444]'
              }`}
              style={{ width: `${connectionHealthPercent}%` }}
            ></div>
          </div>
          <div className="text-xs text-[#71717a] mt-2">
            {summary.allowed_connections} / {summary.total_connections} allowed
          </div>
        </div>

        {/* Blocked Connections */}
        <div className={`card rounded-lg p-5 ${
          summary.blocked_connections > 0 ? 'border-[#ef4444]/30 bg-[#ef4444]/5' : ''
        }`}>
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Blocked Connections</div>
            <Icon name="critical" className={summary.blocked_connections > 0 ? "text-[#ef4444]" : "text-[#71717a]"} size="sm" />
          </div>
          <div className={`text-3xl font-bold ${summary.blocked_connections > 0 ? 'text-[#ef4444]' : 'text-[#e4e4e7]'}`}>
            {summary.blocked_connections}
          </div>
          <div className="text-xs text-[#71717a] mt-2">
            {summary.blocked_connections > 0 ? 'Requires attention' : 'All connections allowed'}
          </div>
        </div>

        {/* Services with Issues */}
        <div className={`card rounded-lg p-5 ${
          servicesWithIssues.length > 0 ? 'border-[#f59e0b]/30 bg-[#f59e0b]/5' : ''
        }`}>
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Services with Issues</div>
            <Icon name="degraded" className={servicesWithIssues.length > 0 ? "text-[#f59e0b]" : "text-[#71717a]"} size="sm" />
          </div>
          <div className={`text-3xl font-bold ${servicesWithIssues.length > 0 ? 'text-[#f59e0b]' : 'text-[#e4e4e7]'}`}>
            {servicesWithIssues.length}
          </div>
          <div className="text-xs text-[#71717a] mt-2">
            {servicesWithIssues.length > 0 ? 'Needs investigation' : 'All services healthy'}
          </div>
        </div>

        {/* Total Services */}
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Total Services</div>
            <Icon name="network" className="text-[#71717a]" size="sm" />
          </div>
          <div className="text-3xl font-bold text-[#e4e4e7]">{summary.total_services}</div>
          <div className="text-xs text-[#71717a] mt-2">
            {summary.services_with_mesh || 0} with service mesh
          </div>
        </div>
      </div>

      {/* Services with Issues - Detailed List */}
      {servicesWithIssues.length > 0 && (
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="critical" className="text-[#ef4444]" size="sm" />
              Services with Blocked Connections
            </h3>
            <button
              onClick={() => onNavigateToTopology?.()}
              className="text-sm text-[#3b82f6] hover:text-[#2563eb] flex items-center gap-1"
            >
              View All →
            </button>
          </div>
          <div className="space-y-3">
            {servicesWithIssues.slice(0, 10).map((service, idx) => (
              <div
                key={idx}
                className="p-4 bg-[#1a1a24] rounded-lg border border-[rgba(255,255,255,0.08)] hover:border-[#ef4444]/30 transition-all cursor-pointer"
                onClick={() => {
                  const serviceKey = service.namespace ? `${service.namespace}/${service.name}` : service.name;
                  onNavigateToService?.(serviceKey);
                }}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-3">
                    <Icon name="critical" className="text-[#ef4444]" size="sm" />
                    <div>
                      <div className="font-semibold text-[#e4e4e7]">
                        {service.namespace ? `${service.namespace}/${service.name}` : service.name}
                      </div>
                      <div className="text-xs text-[#71717a]">
                        {service.blockedConnections} blocked • {service.totalConnections} total connections
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs px-2 py-1 rounded bg-[#ef4444]/20 text-[#ef4444]">
                      {service.blockedConnections} blocked
                    </span>
                    <span className="text-[#71717a]">→</span>
                  </div>
                </div>
                {service.blockingPolicies.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-2">
                    {service.blockingPolicies.slice(0, 3).map((policy, pIdx) => (
                      <span key={pIdx} className="text-xs px-2 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]">
                        {policy}
                      </span>
                    ))}
                    {service.blockingPolicies.length > 3 && (
                      <span className="text-xs px-2 py-0.5 rounded bg-[#71717a]/20 text-[#71717a]">
                        +{service.blockingPolicies.length - 3} more
                      </span>
                    )}
                  </div>
                )}
              </div>
            ))}
            {servicesWithIssues.length > 10 && (
              <div className="text-center pt-2">
                <button
                  onClick={() => onNavigateToTopology?.()}
                  className="text-sm text-[#3b82f6] hover:text-[#2563eb]"
                >
                  View all {servicesWithIssues.length} services with issues →
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Blocked Connections - Detailed List */}
      {blockedConnections.length > 0 && (
        <div className="card rounded-lg p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="critical" className="text-[#ef4444]" size="sm" />
              Blocked Connections
            </h3>
            <div className="flex items-center gap-2">
              <label className="flex items-center gap-2 cursor-pointer text-sm text-[#71717a]">
                <input
                  type="checkbox"
                  checked={autoRefresh}
                  onChange={(e) => setAutoRefresh(e.target.checked)}
                  className="w-4 h-4 rounded border-[rgba(255,255,255,0.2)] bg-[#1a1a24] text-[#3b82f6]"
                />
                Auto-refresh
              </label>
              <button
                onClick={loadDashboardData}
                className="px-3 py-1.5 text-sm bg-[#1a1a24] hover:bg-[#252530] text-[#71717a] rounded-md transition-all flex items-center gap-2"
              >
                <Icon name="refresh" size="sm" />
                Refresh
              </button>
            </div>
          </div>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {blockedConnections.slice(0, 20).map((conn, idx) => (
              <div
                key={idx}
                className="p-3 bg-[#1a1a24] rounded border border-[#ef4444]/30"
              >
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <Icon name="critical" className="text-[#ef4444]" size="sm" />
                    <span className="text-sm font-medium text-[#e4e4e7]">
                      {conn.from} → {conn.to}
                    </span>
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      conn.type === 'ingress' ? 'bg-blue-500/20 text-blue-300' :
                      conn.type === 'egress' ? 'bg-purple-500/20 text-purple-300' :
                      'bg-[#71717a]/20 text-[#71717a]'
                    }`}>
                      {conn.type}
                    </span>
                  </div>
                  <span className="text-xs px-2 py-0.5 rounded bg-[#ef4444]/20 text-[#ef4444]">
                    BLOCKED
                  </span>
                </div>
                <div className="text-xs text-[#71717a] mb-2">{conn.reason}</div>
                {conn.policies.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {conn.policies.map((policy, pIdx) => (
                      <span key={pIdx} className="text-xs px-2 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]">
                        {policy}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
            {blockedConnections.length > 20 && (
              <div className="text-center pt-2 text-sm text-[#71717a]">
                Showing 20 of {blockedConnections.length} blocked connections
              </div>
            )}
          </div>
        </div>
      )}

      {/* Traffic Flow Summary */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Ingress */}
        <div className={`card rounded-lg p-5 ${
          ingressBlocked > 0 ? 'border-[#ef4444]/30' : ''
        }`}>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#3b82f6]" size="sm" />
              Ingress Traffic
            </h3>
            {ingressBlocked > 0 && (
              <span className="text-xs px-2 py-1 rounded bg-[#ef4444]/20 text-[#ef4444]">
                {ingressBlocked} blocked
              </span>
            )}
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="text-xs text-[#71717a] mb-1">Gateways</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {(topology.ingress?.gateways?.length || 0) + (topology.ingress?.kubernetes_ingress?.length || 0)}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Routes</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {topology.ingress?.routes?.length || 0}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Connections</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {topology.ingress?.connections?.length || 0}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Blocked</div>
              <div className={`text-2xl font-bold ${ingressBlocked > 0 ? 'text-[#ef4444]' : 'text-[#10b981]'}`}>
                {ingressBlocked}
              </div>
            </div>
          </div>
        </div>

        {/* Egress */}
        <div className={`card rounded-lg p-5 ${
          egressBlocked > 0 ? 'border-[#ef4444]/30' : ''
        }`}>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
              <Icon name="network" className="text-[#8b5cf6]" size="sm" />
              Egress Traffic
            </h3>
            {egressBlocked > 0 && (
              <span className="text-xs px-2 py-1 rounded bg-[#ef4444]/20 text-[#ef4444]">
                {egressBlocked} blocked
              </span>
            )}
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <div className="text-xs text-[#71717a] mb-1">Gateways</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {topology.egress?.gateways?.length || 0}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">External Services</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {topology.egress?.external_services?.length || 0}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Connections</div>
              <div className="text-2xl font-bold text-[#e4e4e7]">
                {topology.egress?.connections?.length || 0}
              </div>
            </div>
            <div>
              <div className="text-xs text-[#71717a] mb-1">Blocked</div>
              <div className={`text-2xl font-bold ${egressBlocked > 0 ? 'text-[#ef4444]' : 'text-[#10b981]'}`}>
                {egressBlocked}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card rounded-lg p-5">
        <h3 className="text-lg font-semibold mb-4 text-[#e4e4e7] flex items-center gap-2">
          <Icon name="info" className="text-[#71717a]" size="sm" />
          Quick Actions
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <button
            onClick={() => onNavigateToTopology?.()}
            className="p-4 bg-[#1a1a24] hover:bg-[#252530] rounded-lg border border-[rgba(255,255,255,0.08)] transition-all text-left flex items-center gap-3"
          >
            <Icon name="network" className="text-[#3b82f6]" size="md" />
            <div>
              <div className="text-sm font-semibold text-[#e4e4e7]">View Topology</div>
              <div className="text-xs text-[#71717a]">Explore all service connections</div>
            </div>
          </button>
          <button
            onClick={() => onNavigateToTopology?.()}
            className="p-4 bg-[#1a1a24] hover:bg-[#252530] rounded-lg border border-[rgba(255,255,255,0.08)] transition-all text-left flex items-center gap-3"
          >
            <Icon name="network" className="text-[#8b5cf6]" size="md" />
            <div>
              <div className="text-sm font-semibold text-[#e4e4e7]">Trace Path</div>
              <div className="text-xs text-[#71717a]">Debug connection paths</div>
            </div>
          </button>
          <button
            onClick={() => {
              const event = new CustomEvent('navigate', { detail: 'events' });
              window.dispatchEvent(event);
            }}
            className="p-4 bg-[#1a1a24] hover:bg-[#252530] rounded-lg border border-[rgba(255,255,255,0.08)] transition-all text-left flex items-center gap-3"
          >
            <Icon name="events" className="text-[#f59e0b]" size="md" />
            <div>
              <div className="text-sm font-semibold text-[#e4e4e7]">View Events</div>
              <div className="text-xs text-[#71717a]">Check cluster events</div>
            </div>
          </button>
        </div>
      </div>

      {/* Recommendations Panel */}
      <div className="card rounded-lg p-5">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
            <Icon name="scan" className="text-[#3b82f6]" size="sm" />
            Networking Recommendations
          </h3>
          <button
            onClick={loadDashboardData}
            className="px-3 py-1.5 text-sm bg-[#1a1a24] hover:bg-[#252530] text-[#71717a] rounded-md transition-all flex items-center gap-2"
          >
            <Icon name="refresh" size="sm" />
            Refresh
          </button>
        </div>
        <RecommendationsPanel namespace={namespace} />
      </div>

      {/* Infrastructure Status - Collapsed by default */}
      <details className="card rounded-lg p-5">
        <summary className="cursor-pointer text-lg font-semibold text-[#e4e4e7] flex items-center gap-2 list-none">
          <Icon name="info" className="text-[#71717a]" size="sm" />
          Infrastructure Status
        </summary>
        <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <div className="text-xs text-[#71717a] mb-1">CNI</div>
            <div className="text-sm font-semibold text-[#e4e4e7]">{infra.cni || 'Unknown'}</div>
          </div>
          <div className="flex items-center gap-2">
            <Icon 
              name={infra.cilium_enabled ? "check" : "close"} 
              className={infra.cilium_enabled ? "text-[#8b5cf6]" : "text-[#71717a]"} 
            />
            <div>
              <div className="text-xs text-[#71717a]">Cilium</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {infra.cilium_enabled ? 'Enabled' : 'Disabled'}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon 
              name={infra.istio_enabled ? "check" : "close"} 
              className={infra.istio_enabled ? "text-[#3b82f6]" : "text-[#71717a]"} 
            />
            <div>
              <div className="text-xs text-[#71717a]">Istio</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {infra.istio_enabled ? 'Enabled' : 'Disabled'}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Icon 
              name={infra.kyverno_enabled ? "check" : "close"} 
              className={infra.kyverno_enabled ? "text-[#10b981]" : "text-[#71717a]"} 
            />
            <div>
              <div className="text-xs text-[#71717a]">Kyverno</div>
              <div className="text-sm font-semibold text-[#e4e4e7]">
                {infra.kyverno_enabled ? 'Enabled' : 'Disabled'}
              </div>
            </div>
          </div>
        </div>
        <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <div className="text-xs text-[#71717a] mb-1">K8s NetworkPolicies</div>
            <div className="text-sm font-semibold text-[#e4e4e7]">{infra.network_policies || 0}</div>
          </div>
          <div>
            <div className="text-xs text-[#71717a] mb-1">Cilium Policies</div>
            <div className="text-sm font-semibold text-[#e4e4e7]">{infra.cilium_policies || 0}</div>
          </div>
          <div>
            <div className="text-xs text-[#71717a] mb-1">Istio Policies</div>
            <div className="text-sm font-semibold text-[#e4e4e7]">{infra.istio_policies || 0}</div>
          </div>
          <div>
            <div className="text-xs text-[#71717a] mb-1">Kyverno Policies</div>
            <div className="text-sm font-semibold text-[#e4e4e7]">{infra.kyverno_policies || 0}</div>
          </div>
        </div>
      </details>
    </div>
  );
}
