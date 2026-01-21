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
      <div className="glass rounded-xl p-8 text-center border border-[rgba(255,255,255,0.08)]">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a] font-medium tracking-wide">INITIALIZING DASHBOARD SYSTEM...</p>
      </div>
    );
  }

  const summary = topology?.summary || {};
  const connectionHealth = summary.total_connections > 0
    ? Math.round((summary.allowed_connections / summary.total_connections) * 100)
    : null;

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
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {/* Key Metrics - Compact Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {/* Policy Permissiveness */}
        <button
          onClick={() => onNavigate?.('traffic-analysis')}
          className={`group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-5 border transition-all text-left cursor-pointer active:scale-[0.98] relative overflow-hidden ${
            connectionHealth !== null && connectionHealth < 80
              ? 'border-amber-500/20 hover:border-amber-500/40 hover:shadow-[0_0_20px_-5px_rgba(245,158,11,0.2)]'
              : connectionHealth !== null && connectionHealth < 100
                ? 'border-[rgba(255,255,255,0.08)] hover:border-blue-500/30'
                : 'border-[rgba(255,255,255,0.08)] hover:border-purple-500/30'
          }`}
        >
          <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
            <Icon name="network" size="lg" />
          </div>
          <div className="flex items-center justify-between mb-3 relative z-10">
            <div className="text-xs text-[#a1a1aa] uppercase tracking-wider font-medium">Permissiveness</div>
            <Icon
              name={connectionHealth === 100 ? "info" : "check"}
              className={connectionHealth === 100 ? "text-purple-400" : "text-amber-400"}
              size="sm"
            />
          </div>
          <div className={`text-3xl font-bold tracking-tight mb-1 text-glow relative z-10 ${
            connectionHealth === 100 ? 'text-purple-400' : 'text-[#e4e4e7]'
          }`}>
            {connectionHealth !== null ? `${connectionHealth}%` : "N/A"}
          </div>
          <div className="text-xs text-[#71717a] relative z-10 font-mono">
            {connectionHealth !== null ? `${summary.allowed_connections || 0}/${summary.total_connections || 0} FLOWS ALLOWED` : "NO TRAFFIC"}
          </div>
        </button>

        {/* Service Mesh Connections */}
        <button
          onClick={() => onNavigate?.('traffic-analysis')}
          className={`group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-5 border transition-all text-left cursor-pointer active:scale-[0.98] relative overflow-hidden ${
            meshCoveragePercent >= 80
              ? 'border-[rgba(255,255,255,0.08)] hover:border-blue-500/30 hover:shadow-[0_0_20px_-5px_rgba(59,130,246,0.2)]'
              : meshCoveragePercent >= 50
                ? 'border-amber-500/20 hover:border-amber-500/40'
                : 'border-[rgba(255,255,255,0.08)] hover:border-[rgba(255,255,255,0.15)]'
          }`}
        >
          <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
            <Icon name="check" size="lg" />
          </div>
          <div className="flex items-center justify-between mb-3 relative z-10">
            <div className="text-xs text-[#a1a1aa] uppercase tracking-wider font-medium">Service Mesh</div>
            <Icon
              name={meshCoveragePercent >= 80 ? "healthy" : meshCoveragePercent >= 50 ? "degraded" : "info"}
              className={meshCoveragePercent >= 80 ? "text-blue-400" : meshCoveragePercent >= 50 ? "text-amber-400" : "text-[#71717a]"}
              size="sm"
            />
          </div>
          <div className={`text-3xl font-bold tracking-tight mb-1 text-glow relative z-10 ${
            meshCoveragePercent >= 80 ? 'text-blue-400' : meshCoveragePercent >= 50 ? 'text-amber-400' : 'text-[#e4e4e7]'
          }`}>
            {meshConnections}
          </div>
          <div className="text-xs text-[#71717a] relative z-10 font-mono">
           {meshCoveragePercent}% COVERAGE
          </div>
        </button>

        {/* Policy Enforcement Rate */}
        <button
          onClick={() => onNavigate?.('network-policies')}
          className={`group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-5 border transition-all text-left cursor-pointer active:scale-[0.98] relative overflow-hidden ${
             connectionHealth !== null && connectionHealth >= 95
              ? 'border-[rgba(255,255,255,0.08)] hover:border-cyan-500/30 hover:shadow-[0_0_20px_-5px_rgba(6,182,212,0.2)]'
              : 'border-amber-500/20 hover:border-amber-500/40'
          }`}
        >
          <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
            <Icon name="scan" size="lg" />
          </div>
          <div className="flex items-center justify-between mb-3 relative z-10">
            <div className="text-xs text-[#a1a1aa] uppercase tracking-wider font-medium">Enforcement</div>
            <Icon
              name={connectionHealth !== null && connectionHealth >= 95 ? "healthy" : connectionHealth !== null && connectionHealth >= 80 ? "degraded" : "critical"}
              className={connectionHealth !== null && connectionHealth >= 95 ? "text-cyan-400" : connectionHealth !== null && connectionHealth >= 80 ? "text-amber-400" : "text-red-400"}
              size="sm"
            />
          </div>
          <div className={`text-3xl font-bold tracking-tight mb-1 text-glow relative z-10 ${
            connectionHealth !== null && connectionHealth >= 95 ? 'text-cyan-400' : connectionHealth !== null && connectionHealth >= 80 ? 'text-amber-400' : 'text-red-400'
          }`}>
            {connectionHealth !== null ? `${connectionHealth}%` : "N/A"}
          </div>
          <div className="text-xs text-[#71717a] relative z-10 font-mono">
            POLICY ACTIVE
          </div>
        </button>

        {/* Compliance Score */}
        <button
          onClick={() => onNavigate?.('compliance')}
           className={`group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-5 border transition-all text-left cursor-pointer active:scale-[0.98] relative overflow-hidden ${
            (score?.score || 0) < 60
              ? 'border-red-500/20 hover:border-red-500/40'
              : (score?.score || 0) < 80
                ? 'border-amber-500/20 hover:border-amber-500/40'
                : 'border-[rgba(255,255,255,0.08)] hover:border-purple-500/30 hover:shadow-[0_0_20px_-5px_rgba(168,85,247,0.2)]'
          }`}
        >
           <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
            <Icon name="check" size="lg" />
          </div>
          <div className="flex items-center justify-between mb-3 relative z-10">
            <div className="text-xs text-[#a1a1aa] uppercase tracking-wider font-medium">Compliance</div>
            <Icon
              name={(score?.score || 0) >= 80 ? "healthy" : (score?.score || 0) >= 60 ? "degraded" : "critical"}
              className={(score?.score || 0) >= 80 ? "text-purple-400" : (score?.score || 0) >= 60 ? "text-amber-400" : "text-red-400"}
              size="sm"
            />
          </div>
          <div className={`text-3xl font-bold tracking-tight mb-1 text-glow relative z-10 ${
            (score?.score || 0) >= 80 ? 'text-purple-400' : (score?.score || 0) >= 60 ? 'text-amber-400' : 'text-red-400'
          }`}>
            {score?.score || 0}%
          </div>
          <div className="text-xs text-[#71717a] relative z-10 font-mono">
            {score?.passed || 0}/{score?.total || 0} CHECKS PASSED
          </div>
        </button>
      </div>

      {/* Quick Actions - Compact */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: 'Topology', icon: 'network' as const, color: 'text-blue-400', section: 'traffic-analysis', subsection: 'topology', desc: `${summary.total_services || 0} SVCS • ${summary.total_connections || 0} LINKS` },
          { label: 'Path Tracer', icon: 'network' as const, color: 'text-purple-400', section: 'traffic-analysis', subsection: 'path-trace', desc: 'DEBUG PATHS' },
          { label: 'Policies', icon: 'scan' as const, color: 'text-emerald-400', section: 'network-policies', desc: `${(topology?.infrastructure?.network_policies || 0) + (topology?.infrastructure?.cilium_policies || 0)} ACTIVE` },
          { label: 'Compliance', icon: 'scan' as const, color: 'text-cyan-400', section: 'compliance', desc: `${score?.recommendations_count || 0} ACTIONS` }
        ].map((action, idx) => (
           <button
            key={idx}
            onClick={() => onNavigate?.(action.section, action.subsection)}
            className="group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-4 border border-[rgba(255,255,255,0.08)] hover:bg-[rgba(255,255,255,0.03)] hover:border-[rgba(255,255,255,0.15)] transition-all text-left cursor-pointer active:scale-[0.98] flex items-center justify-between"
          >
            <div>
              <div className="flex items-center gap-2 mb-1">
                <Icon name={action.icon} className={action.color} size="sm" />
                <div className="text-sm font-semibold text-[#e4e4e7] tracking-wide">{action.label}</div>
              </div>
              <div className="text-[10px] text-[#71717a] font-mono uppercase pl-6">
                {action.desc}
              </div>
            </div>
            <div className="opacity-0 group-hover:opacity-100 transition-opacity text-[#71717a] -translate-x-2 group-hover:translate-x-0 duration-200">
               <Icon name="info" size="sm" />
            </div>
          </button>
        ))}
      </div>

      {/* Traffic Summary - Compact */}
      <div className="grid grid-cols-2 gap-4">
        <button
          onClick={() => onNavigate?.('traffic-analysis', 'ingress')}
          className={`group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-5 border transition-all text-left cursor-pointer active:scale-[0.98] ${
            ingressBlocked > 0
            ? 'border-red-500/20 hover:border-red-500/40'
            : 'border-[rgba(255,255,255,0.08)] hover:border-blue-500/30'
          }`}
        >
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400">
                 <Icon name="network" size="sm" />
              </div>
              <div>
                <div className="text-sm font-semibold text-[#e4e4e7] tracking-wide">Ingress Traffic</div>
                <div className="text-[10px] text-[#71717a] font-mono">INBOUND TRAFFIC FLOW</div>
              </div>
            </div>
            {ingressBlocked > 0 && (
              <span className="text-[10px] px-2 py-0.5 rounded-full bg-red-500/10 text-red-500 border border-red-500/20 font-mono">
                {ingressBlocked} BLOCKED
              </span>
            )}
          </div>
          <div className="text-xs text-[#a1a1aa] pl-[44px]">
            {topology?.ingress?.routes?.length || 0} Routes configured • {topology?.ingress?.connections?.length || 0} Active connections
          </div>
        </button>

        <button
          onClick={() => onNavigate?.('traffic-analysis', 'egress')}
          className={`group bg-[#0a0a0f]/40 backdrop-blur-md rounded-xl p-5 border transition-all text-left cursor-pointer active:scale-[0.98] ${
            egressBlocked > 0
            ? 'border-red-500/20 hover:border-red-500/40'
            : 'border-[rgba(255,255,255,0.08)] hover:border-purple-500/30'
          }`}
        >
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-purple-500/10 text-purple-400">
                 <Icon name="network" size="sm" />
              </div>
              <div>
                <div className="text-sm font-semibold text-[#e4e4e7] tracking-wide">Egress Traffic</div>
                <div className="text-[10px] text-[#71717a] font-mono">OUTBOUND TRAFFIC FLOW</div>
              </div>
            </div>
            {egressBlocked > 0 && (
              <span className="text-[10px] px-2 py-0.5 rounded-full bg-red-500/10 text-red-500 border border-red-500/20 font-mono">
                {egressBlocked} BLOCKED
              </span>
            )}
          </div>
          <div className="text-xs text-[#a1a1aa] pl-[44px]">
            {topology?.egress?.external_services?.length || 0} External services • {topology?.egress?.connections?.length || 0} Active connections
          </div>
        </button>
      </div>
    </div>
  );
}
