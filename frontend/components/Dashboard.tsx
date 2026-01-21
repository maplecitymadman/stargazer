'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface DashboardProps {
  namespace?: string;
  onNavigate?: (section: string, subsection?: string) => void;
}

export default function Dashboard({ namespace, onNavigate }: DashboardProps) {
  const [topology, setTopology] = useState<any>(null);
  const [health, setHealth] = useState<any>(null);
  const [score, setScore] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
    const interval = setInterval(loadDashboardData, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, [namespace]);

  const loadDashboardData = async () => {
    try {
      const [topologyData, healthData, scoreData] = await Promise.all([
        apiClient.getServiceTopology(namespace).catch(() => null),
        apiClient.getHealth(namespace).catch(() => null),
        apiClient.getComplianceScore(namespace).catch(() => null),
      ]);
      setTopology(topologyData);
      setHealth(healthData);
      setScore(scoreData);
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
        <p className="text-[#71717a]">Loading dashboard...</p>
      </div>
    );
  }

  const summary = topology?.summary || {};
  const connectionHealth = summary.total_connections > 0
    ? Math.round((summary.allowed_connections / summary.total_connections) * 100)
    : 100;

  const servicesWithIssues = topology ? Object.values(topology.connectivity || {}).filter((conn: any) =>
    conn.blocked_from && conn.blocked_from.length > 0
  ).length : 0;

  // Service mesh connections - count connections that go through service mesh
  const meshConnections = topology ? Object.values(topology.connectivity || {}).reduce((total: number, conn: any) => {
    const meshCount = (conn.connections || []).filter((c: any) => c.via_service_mesh).length;
    return total + meshCount;
  }, 0) : 0;

  // Calculate mTLS coverage - services with mesh that have strict mTLS
  const servicesWithMesh = summary.services_with_mesh || 0;
  const totalServices = summary.total_services || 0;
  const meshCoveragePercent = totalServices > 0 ? Math.round((servicesWithMesh / totalServices) * 100) : 0;

  const ingressBlocked = topology?.ingress?.connections?.filter((c: any) => !c.allowed).length || 0;
  const egressBlocked = topology?.egress?.connections?.filter((c: any) => !c.allowed).length || 0;

  return (
    <div className="space-y-4">
      {/* Key Metrics - Compact Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {/* Connection Health */}
        <button
          onClick={() => onNavigate?.('traffic-analysis')}
          className={`card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98] ${
            connectionHealth < 80 ? 'border-[#ef4444]/30' : connectionHealth < 100 ? 'border-[#f59e0b]/30' : ''
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Connection Health</div>
            <Icon
              name={connectionHealth === 100 ? "healthy" : connectionHealth >= 80 ? "degraded" : "critical"}
              className={connectionHealth === 100 ? "text-[#10b981]" : connectionHealth >= 80 ? "text-[#f59e0b]" : "text-[#ef4444]"}
              size="sm"
            />
          </div>
          <div className={`text-2xl font-bold ${
            connectionHealth === 100 ? 'text-[#10b981]' : connectionHealth >= 80 ? 'text-[#f59e0b]' : 'text-[#ef4444]'
          }`}>
            {connectionHealth}%
          </div>
          <div className="text-xs text-[#71717a] mt-1">
            {summary.allowed_connections || 0} / {summary.total_connections || 0} allowed
          </div>
        </button>

        {/* Service Mesh Connections */}
        <button
          onClick={() => onNavigate?.('traffic-analysis', 'topology')}
          className={`card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98] ${
            meshCoveragePercent >= 80 ? 'border-[#10b981]/30' : meshCoveragePercent >= 50 ? 'border-[#f59e0b]/30' : ''
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Service Mesh</div>
            <Icon
              name={meshCoveragePercent >= 80 ? "healthy" : meshCoveragePercent >= 50 ? "degraded" : "info"}
              className={meshCoveragePercent >= 80 ? "text-[#10b981]" : meshCoveragePercent >= 50 ? "text-[#f59e0b]" : "text-[#71717a]"}
              size="sm"
            />
          </div>
          <div className={`text-2xl font-bold ${
            meshCoveragePercent >= 80 ? 'text-[#10b981]' : meshCoveragePercent >= 50 ? 'text-[#f59e0b]' : 'text-[#e4e4e7]'
          }`}>
            {meshConnections}
          </div>
          <div className="text-xs text-[#71717a] mt-1">
            {meshCoveragePercent}% coverage
          </div>
        </button>

        {/* Policy Enforcement Rate */}
        <button
          onClick={() => onNavigate?.('network-policies')}
          className={`card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98] ${
            connectionHealth >= 95 ? 'border-[#10b981]/30' : connectionHealth >= 80 ? 'border-[#f59e0b]/30' : ''
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Policy Enforcement</div>
            <Icon
              name={connectionHealth >= 95 ? "healthy" : connectionHealth >= 80 ? "degraded" : "critical"}
              className={connectionHealth >= 95 ? "text-[#10b981]" : connectionHealth >= 80 ? "text-[#f59e0b]" : "text-[#ef4444]"}
              size="sm"
            />
          </div>
          <div className={`text-2xl font-bold ${
            connectionHealth >= 95 ? 'text-[#10b981]' : connectionHealth >= 80 ? 'text-[#f59e0b]' : 'text-[#ef4444]'
          }`}>
            {connectionHealth}%
          </div>
          <div className="text-xs text-[#71717a] mt-1">
            {summary.allowed_connections || 0} / {summary.total_connections || 0} allowed
          </div>
        </button>

        {/* Compliance Score */}
        <button
          onClick={() => onNavigate?.('compliance', 'score')}
          className={`card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98] ${
            (score?.score || 0) < 60 ? 'border-[#ef4444]/30' : (score?.score || 0) < 80 ? 'border-[#f59e0b]/30' : ''
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs text-[#71717a]">Compliance</div>
            <Icon
              name={(score?.score || 0) >= 80 ? "healthy" : (score?.score || 0) >= 60 ? "degraded" : "critical"}
              className={(score?.score || 0) >= 80 ? "text-[#10b981]" : (score?.score || 0) >= 60 ? "text-[#f59e0b]" : "text-[#ef4444]"}
              size="sm"
            />
          </div>
          <div className={`text-2xl font-bold ${
            (score?.score || 0) >= 80 ? 'text-[#10b981]' : (score?.score || 0) >= 60 ? 'text-[#f59e0b]' : 'text-[#ef4444]'
          }`}>
            {score?.score || 0}%
          </div>
          <div className="text-xs text-[#71717a] mt-1">
            {score?.passed || 0} / {score?.total || 0} checks
          </div>
        </button>
      </div>

      {/* Quick Actions - Compact */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <button
          onClick={() => onNavigate?.('traffic-analysis', 'topology')}
          className="card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
        >
          <div className="flex items-center gap-2 mb-2">
            <Icon name="network" className="text-[#3b82f6]" size="sm" />
            <div className="text-sm font-semibold text-[#e4e4e7]">Topology</div>
          </div>
          <div className="text-xs text-[#71717a]">
            {summary.total_services || 0} services • {summary.total_connections || 0} connections
          </div>
        </button>

        <button
          onClick={() => onNavigate?.('traffic-analysis', 'path-trace')}
          className="card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
        >
          <div className="flex items-center gap-2 mb-2">
            <Icon name="network" className="text-[#8b5cf6]" size="sm" />
            <div className="text-sm font-semibold text-[#e4e4e7]">Path Tracer</div>
          </div>
          <div className="text-xs text-[#71717a]">Debug connection paths</div>
        </button>

        <button
          onClick={() => onNavigate?.('network-policies')}
          className="card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
        >
          <div className="flex items-center gap-2 mb-2">
            <Icon name="scan" className="text-[#10b981]" size="sm" />
            <div className="text-sm font-semibold text-[#e4e4e7]">Policies</div>
          </div>
          <div className="text-xs text-[#71717a]">
            {topology?.infrastructure?.network_policies || 0} K8s • {topology?.infrastructure?.cilium_policies || 0} Cilium
          </div>
        </button>

        <button
          onClick={() => onNavigate?.('compliance')}
          className="card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98]"
        >
          <div className="flex items-center gap-2 mb-2">
            <Icon name="scan" className="text-[#3b82f6]" size="sm" />
            <div className="text-sm font-semibold text-[#e4e4e7]">Compliance</div>
          </div>
          <div className="text-xs text-[#71717a]">
            {score?.recommendations_count || 0} recommendations
          </div>
        </button>
      </div>

      {/* Traffic Summary - Compact */}
      <div className="grid grid-cols-2 gap-3">
        <button
          onClick={() => onNavigate?.('traffic-analysis', 'ingress')}
          className={`card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98] ${
            ingressBlocked > 0 ? 'border-[#ef4444]/30' : ''
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <Icon name="network" className="text-[#3b82f6]" size="sm" />
              <div className="text-sm font-semibold text-[#e4e4e7]">Ingress</div>
            </div>
            {ingressBlocked > 0 && (
              <span className="text-xs px-2 py-0.5 rounded bg-[#ef4444]/20 text-[#ef4444]">
                {ingressBlocked} blocked
              </span>
            )}
          </div>
          <div className="text-xs text-[#71717a]">
            {topology?.ingress?.routes?.length || 0} routes • {topology?.ingress?.connections?.length || 0} connections
          </div>
        </button>

        <button
          onClick={() => onNavigate?.('traffic-analysis', 'egress')}
          className={`card rounded-lg p-4 hover:bg-[#252530] transition-all text-left cursor-pointer active:scale-[0.98] ${
            egressBlocked > 0 ? 'border-[#ef4444]/30' : ''
          }`}
        >
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <Icon name="network" className="text-[#8b5cf6]" size="sm" />
              <div className="text-sm font-semibold text-[#e4e4e7]">Egress</div>
            </div>
            {egressBlocked > 0 && (
              <span className="text-xs px-2 py-0.5 rounded bg-[#ef4444]/20 text-[#ef4444]">
                {egressBlocked} blocked
              </span>
            )}
          </div>
          <div className="text-xs text-[#71717a]">
            {topology?.egress?.external_services?.length || 0} external • {topology?.egress?.connections?.length || 0} connections
          </div>
        </button>
      </div>
    </div>
  );
}
