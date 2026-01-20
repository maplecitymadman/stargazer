'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import RecommendationsPanel from './RecommendationsPanel';
import { Icon } from './SpaceshipIcons';

interface TroubleshootingPageProps {
  subsection?: string;
  namespace?: string;
}

export default function TroubleshootingPage({ subsection, namespace }: TroubleshootingPageProps) {
  const [activeTab, setActiveTab] = useState(subsection || 'blocked');
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
      console.error('Error loading troubleshooting data:', error);
    } finally {
      setLoading(false);
    }
  };

  const tabs = [
    { id: 'blocked', label: 'Blocked Connections', icon: 'critical' as const },
    { id: 'services', label: 'Services with Issues', icon: 'degraded' as const },
    { id: 'recommendations', label: 'Recommendations', icon: 'scan' as const },
  ];

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading troubleshooting data...</p>
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
        {activeTab === 'blocked' && <BlockedConnections topology={topology} namespace={namespace} />}
        {activeTab === 'services' && <ServicesWithIssues topology={topology} namespace={namespace} />}
        {activeTab === 'recommendations' && <RecommendationsPanel namespace={namespace} />}
      </div>
    </div>
  );
}

// Blocked Connections Tab
function BlockedConnections({ topology, namespace }: { topology: any; namespace?: string }) {
  const [blockedConnections, setBlockedConnections] = useState<any[]>([]);

  useEffect(() => {
    if (!topology) return;

    const blocked: any[] = [];

    // Service-to-service blocked connections
    Object.entries(topology.connectivity || {}).forEach(([serviceKey, connInfo]: [string, any]) => {
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
    });

    // Ingress blocked
    topology.ingress?.connections?.forEach((conn: any) => {
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

    // Egress blocked
    topology.egress?.connections?.forEach((conn: any) => {
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

    setBlockedConnections(blocked);
  }, [topology]);

  if (blockedConnections.length === 0) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="healthy" className="text-[#10b981] text-4xl mb-4" />
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">No Blocked Connections</h3>
        <p className="text-sm text-[#71717a]">All connections are allowed</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="card rounded-lg p-5">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
            <Icon name="critical" className="text-[#ef4444]" size="sm" />
            Blocked Connections ({blockedConnections.length})
          </h3>
        </div>
        <div className="space-y-2 max-h-[600px] overflow-y-auto">
          {blockedConnections.map((conn, idx) => (
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
    </div>
  );
}

// Services with Issues Tab
function ServicesWithIssues({ topology, namespace }: { topology: any; namespace?: string }) {
  const [servicesWithIssues, setServicesWithIssues] = useState<any[]>([]);

  useEffect(() => {
    if (!topology) return;

    const issues: any[] = [];
    Object.entries(topology.connectivity || {}).forEach(([serviceKey, connInfo]: [string, any]) => {
      if (connInfo.blocked_from && connInfo.blocked_from.length > 0) {
        const [ns, name] = serviceKey.includes('/') ? serviceKey.split('/') : [namespace || 'default', serviceKey];
        issues.push({
          name,
          namespace: ns,
          serviceKey,
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
      }
    });

    setServicesWithIssues(issues);
  }, [topology, namespace]);

  if (servicesWithIssues.length === 0) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="healthy" className="text-[#10b981] text-4xl mb-4" />
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">All Services Healthy</h3>
        <p className="text-sm text-[#71717a]">No services with blocked connections</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="card rounded-lg p-5">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
            <Icon name="critical" className="text-[#ef4444]" size="sm" />
            Services with Blocked Connections ({servicesWithIssues.length})
          </h3>
        </div>
        <div className="space-y-3">
          {servicesWithIssues.map((service, idx) => (
            <div
              key={idx}
              className="p-4 bg-[#1a1a24] rounded-lg border border-[rgba(255,255,255,0.08)] hover:border-[#ef4444]/30 transition-all"
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
                <span className="text-xs px-2 py-1 rounded bg-[#ef4444]/20 text-[#ef4444]">
                  {service.blockedConnections} blocked
                </span>
              </div>
              {service.blockingPolicies.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {service.blockingPolicies.slice(0, 5).map((policy: string, pIdx: number) => (
                    <span key={pIdx} className="text-xs px-2 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]">
                      {policy}
                    </span>
                  ))}
                  {service.blockingPolicies.length > 5 && (
                    <span className="text-xs px-2 py-0.5 rounded bg-[#71717a]/20 text-[#71717a]">
                      +{service.blockingPolicies.length - 5} more
                    </span>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
