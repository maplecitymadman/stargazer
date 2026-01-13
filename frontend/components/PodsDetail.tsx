'use client';

import { useState, useEffect } from 'react';
import { apiClient, Pod } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface PodsDetailProps {
  namespace?: string;
  onClose?: () => void;
}

export default function PodsDetail({ namespace, onClose }: PodsDetailProps) {
  const [pods, setPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'running' | 'pending' | 'failed'>('all');

  useEffect(() => {
    loadPods();
  }, [namespace]);

  const loadPods = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getPods(namespace);
      setPods(data);
    } catch (error) {
      // Error loading pods - will show empty state
    } finally {
      setLoading(false);
    }
  };

  const filteredPods = pods.filter(pod => {
    if (filter === 'all') return true;
    return pod.status.toLowerCase() === filter;
  });

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'running':
        return 'text-emerald-400';
      case 'pending':
        return 'text-yellow-400';
      case 'failed':
      case 'error':
        return 'text-red-400';
      default:
        return 'text-[#e4e4e7]-dim';
    }
  };

  const mainContent = (
    <div className={`card rounded-lg p-5 w-full ${onClose ? 'max-w-6xl max-h-[90vh]' : ''} overflow-hidden flex flex-col`}>
      <div className="flex items-center justify-between mb-5 pb-4 border-b border-[rgba(255,255,255,0.08)]">
        <div className="flex items-center gap-3">
          <Icon name="pods" className="text-[#71717a]" size="sm" />
          <h2 className="text-2xl font-bold text-[#e4e4e7]">Pods</h2>
          <span className="px-2.5 py-1 rounded-full bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#71717a] text-xs">
            {pods.length} TOTAL
          </span>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="px-3 py-1.5 bg-red-600/90 hover:bg-red-700 text-white rounded-lg text-sm transition-all"
          >
            CLOSE
          </button>
        )}
      </div>

      <div className="flex gap-2 mb-4">
        {[
          { id: 'all', label: 'All' },
          { id: 'running', label: 'Running' },
          { id: 'pending', label: 'Pending' },
          { id: 'failed', label: 'Failed' },
        ].map(f => (
          <button
            key={f.id}
            onClick={() => setFilter(f.id as any)}
            className={`px-4 py-2 rounded-lg text-sm transition-all ${
              filter === f.id
                ? 'bg-[#3b82f6] text-white'
                : 'bg-[#1a1a24] text-[#71717a] hover:bg-[#27272a]'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-auto space-y-3">
        {loading ? (
          <div className="text-center py-10">
            <Icon name="loading" className="text-[#71717a] animate-pulse text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">Loading pods...</p>
          </div>
        ) : filteredPods.length === 0 ? (
          <div className="text-center py-10">
            <Icon name="info" className="text-[#e4e4e7]-dim/60 text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">No pods found</p>
          </div>
        ) : (
            filteredPods.map(pod => (
              <div
                key={pod.name}
                className="card card-hover rounded-lg p-3.5 border border-purple-500/15"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <Icon name="pods" className="text-[#71717a]" size="sm" />
                      <h3 className="font-bold text-[#e4e4e7] text-base">{pod.name}</h3>
                      <span className={`px-2 py-0.5 rounded text-xs font-semibold ${getStatusColor(pod.status)}`}>
                        {pod.status.toUpperCase()}
                      </span>
                      {pod.ready && (
                        <span className="px-2 py-0.5 rounded text-xs bg-emerald-900/20 text-emerald-400/90 border border-emerald-500/20">
                          READY
                        </span>
                      )}
                    </div>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-2.5 text-xs">
                      <div>
                        <span className="text-[#71717a]">Namespace:</span>
                        <div className="text-[#3b82f6]">{pod.namespace}</div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Node:</span>
                        <div className="text-[#e4e4e7]">{pod.node || 'N/A'}</div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Restarts:</span>
                        <div className={pod.restarts > 5 ? 'text-red-400' : 'text-[#e4e4e7]'}>
                          {pod.restarts}
                        </div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Age:</span>
                        <div className="text-[#e4e4e7]">{pod.age}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))
        )}
      </div>
    </div>
  );

  if (onClose) {
    return (
      <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        {mainContent}
      </div>
    );
  }

  return mainContent;
}
