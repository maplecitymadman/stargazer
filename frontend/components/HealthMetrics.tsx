'use client';

import { ClusterHealth } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface HealthMetricsProps {
  health: ClusterHealth;
  namespace?: string;
  onPodsClick?: () => void;
  onDeploymentsClick?: () => void;
  onEventsClick?: () => void;
}

export default function HealthMetrics({ health, namespace, onPodsClick, onDeploymentsClick, onEventsClick }: HealthMetricsProps) {
  const podsPercent = health.pods.total > 0 
    ? Math.round((health.pods.healthy / health.pods.total) * 100) 
    : 0;
  const depsPercent = health.deployments.total > 0
    ? Math.round((health.deployments.healthy / health.deployments.total) * 100)
    : 0;

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

      {/* Pods */}
      <div 
        className="card card-hover rounded-lg p-5 cursor-pointer"
        onClick={onPodsClick || (() => {})}
      >
        <div className="text-xs text-[#71717a] mb-2 flex items-center gap-2">
          <Icon name="pods" className="text-[#71717a]" size="sm" />
          <span>Pods</span>
        </div>
        <div className="text-2xl font-semibold mt-2 text-[#e4e4e7]">
          {health.pods.healthy}/{health.pods.total}
        </div>
        <div className={`text-xs mt-2 font-medium ${
          podsPercent === 100 
            ? 'text-[#10b981]' 
            : podsPercent >= 80 
            ? 'text-[#f59e0b]' 
            : 'text-[#ef4444]'
        }`}>
          {podsPercent}% healthy
        </div>
        <div className="mt-3 h-1 bg-[#1a1a24] rounded-full overflow-hidden">
          <div 
            className={`h-full transition-all duration-500 ${
              podsPercent === 100 ? 'bg-[#10b981]' : podsPercent >= 80 ? 'bg-[#f59e0b]' : 'bg-[#ef4444]'
            }`}
            style={{ width: `${podsPercent}%` }}
          ></div>
        </div>
      </div>

      {/* Deployments */}
      <div 
        className="card card-hover rounded-lg p-5 cursor-pointer"
        onClick={onDeploymentsClick || (() => {})}
      >
        <div className="text-xs text-[#71717a] mb-2 flex items-center gap-2">
          <Icon name="deployments" className="text-[#71717a]" size="sm" />
          <span>Deployments</span>
        </div>
        <div className="text-2xl font-semibold mt-2 text-[#e4e4e7]">
          {health.deployments.healthy}/{health.deployments.total}
        </div>
        <div className={`text-xs mt-2 font-medium ${
          depsPercent === 100 
            ? 'text-[#10b981]' 
            : depsPercent >= 80 
            ? 'text-[#f59e0b]' 
            : 'text-[#ef4444]'
        }`}>
          {depsPercent}% healthy
        </div>
        <div className="mt-3 h-1 bg-[#1a1a24] rounded-full overflow-hidden">
          <div 
            className={`h-full transition-all duration-500 ${
              depsPercent === 100 ? 'bg-[#10b981]' : depsPercent >= 80 ? 'bg-[#f59e0b]' : 'bg-[#ef4444]'
            }`}
            style={{ width: `${depsPercent}%` }}
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
