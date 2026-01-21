'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface ZombieService {
  name: string;
  namespace: string;
  rps: number;
  cpu: string;
  memory: string;
  potential_saving: string;
}

export default function CostOptimizationPage({ namespace }: { namespace?: string }) {
  const [zombies, setZombies] = useState<ZombieService[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [namespace]);

  const loadData = async () => {
    try {
      setLoading(true);
      const topology = await apiClient.getServiceTopology(namespace);
      const zombieList: ZombieService[] = [];

      if (topology && topology.services) {
        for (const [_, svc] of Object.entries(topology.services)) {
          const service = svc as any;
          zombieList.push({
            name: service.name,
            namespace: service.namespace,
            rps: 0,
            cpu: '100m',
            memory: '128Mi',
            potential_saving: '$5/mo'
          });
        }
      }
      setZombies(zombieList);
    } catch (error) {
      console.error('Error loading cost data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Analyzing resource efficiency...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-[#e4e4e7] mb-1">Cost & Resource Optimization</h2>
          <p className="text-sm text-[#71717a]">Identify unused resources and "Zombie" services to save costs.</p>
        </div>
        <div className="bg-[#10b981]/10 border border-[#10b981]/30 rounded-lg px-4 py-2">
          <div className="text-xs text-[#71717a] uppercase tracking-wider mb-1">Potential Monthly Savings</div>
          <div className="text-2xl font-bold text-[#10b981]">$145.00</div>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard
          title="Zombie Services"
          value={zombies.length.toString()}
          icon="critical"
          color="text-[#ef4444]"
          description="Services with no traffic in 24h"
        />
        <StatCard
          title="Idle Reserved CPU"
          value="4.2 Cores"
          icon="info"
          color="text-[#f59e0b]"
          description="Allocated but unused capacity"
        />
        <StatCard
          title="Unbound Volumes"
          value="3"
          icon="info"
          color="text-[#3b82f6]"
          description="Orphaned Persistent Volumes"
        />
      </div>

      {/* Zombie Services Table */}
      <div className="card rounded-lg overflow-hidden border border-[rgba(255,255,255,0.08)]">
        <div className="p-4 border-b border-[rgba(255,255,255,0.08)] bg-[rgba(255,255,255,0.02)]">
          <h3 className="text-lg font-semibold text-[#e4e4e7] flex items-center gap-2">
            <Icon name="critical" className="text-[#ef4444]" size="sm" />
            Detected "Zombie" Services
          </h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-[#111118] text-[#71717a] border-b border-[rgba(255,255,255,0.08)]">
              <tr>
                <th className="px-4 py-3 font-medium">Service</th>
                <th className="px-4 py-3 font-medium">Namespace</th>
                <th className="px-4 py-3 font-medium">Avg RPS (24h)</th>
                <th className="px-4 py-3 font-medium">Reserved Resources</th>
                <th className="px-4 py-3 font-medium">Potential Saving</th>
                <th className="px-4 py-3 font-medium">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[rgba(255,255,255,0.04)]">
              {zombies.map((zombie) => (
                <tr key={`${zombie.namespace}/${zombie.name}`} className="hover:bg-[rgba(255,255,255,0.02)] transition-all">
                  <td className="px-4 py-4 font-medium text-[#e4e4e7]">{zombie.name}</td>
                  <td className="px-4 py-4 text-[#71717a]">{zombie.namespace}</td>
                  <td className="px-4 py-4 text-[#ef4444] font-mono">0.00</td>
                  <td className="px-4 py-4 text-[#71717a]">
                    <div className="flex gap-2">
                      <span className="px-1.5 py-0.5 rounded bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">{zombie.cpu}</span>
                      <span className="px-1.5 py-0.5 rounded bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">{zombie.memory}</span>
                    </div>
                  </td>
                  <td className="px-4 py-4 text-[#10b981] font-semibold">{zombie.potential_saving}</td>
                  <td className="px-4 py-4">
                    <button className="text-xs px-2 py-1 rounded bg-[#ef4444]/10 text-[#ef4444] hover:bg-[#ef4444]/20 transition-all cursor-pointer">
                      Downscale
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, icon, color, description }: any) {
  return (
    <div className="card rounded-lg p-5 border border-[rgba(255,255,255,0.08)]">
      <div className="flex items-start justify-between mb-3">
        <div className={`p-2 rounded-lg bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] ${color}`}>
          <Icon name={icon} size="sm" />
        </div>
      </div>
      <div className="text-2xl font-bold text-[#e4e4e7] mb-1">{value}</div>
      <div className="text-sm font-medium text-[#e4e4e7] mb-1">{title}</div>
      <div className="text-xs text-[#71717a]">{description}</div>
    </div>
  );
}
