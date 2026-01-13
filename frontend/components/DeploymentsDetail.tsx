'use client';

import { useState, useEffect } from 'react';
import apiClient from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface Deployment {
  name: string;
  namespace: string;
  replicas: number;
  ready: number;
  up_to_date: number;
  available: number;
  age: string;
}

interface DeploymentsDetailProps {
  namespace?: string;
  onClose?: () => void;
}

export default function DeploymentsDetail({ namespace, onClose }: DeploymentsDetailProps) {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDeployments();
  }, [namespace]);

  const loadDeployments = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getDeployments(namespace);
      setDeployments(data.deployments);
    } catch (error) {
      // Error loading deployments - will show empty state
    } finally {
      setLoading(false);
    }
  };

  const getHealthStatus = (deployment: Deployment) => {
    if (deployment.ready === deployment.replicas && deployment.available === deployment.replicas) {
      return { status: 'HEALTHY', color: 'text-emerald-400' };
    } else if (deployment.ready === 0) {
      return { status: 'FAILED', color: 'text-red-400' };
    } else {
      return { status: 'DEGRADED', color: 'text-yellow-400' };
    }
  };

  const mainContent = (
    <div className={`card rounded-lg p-5 w-full ${onClose ? 'max-w-6xl max-h-[90vh]' : ''} overflow-hidden flex flex-col`}>
      <div className="flex items-center justify-between mb-5 pb-4 border-b border-[rgba(255,255,255,0.08)]">
        <div className="flex items-center gap-3">
          <Icon name="deployments" className="text-[#71717a]" size="sm" />
          <h2 className="text-2xl font-bold text-[#e4e4e7]">Deployments</h2>
          <span className="px-2.5 py-1 rounded-full bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#71717a] text-xs">
            {deployments.length} TOTAL
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

      <div className="flex-1 overflow-auto space-y-2.5">
        {loading ? (
          <div className="text-center py-10">
            <Icon name="loading" className="text-[#71717a] animate-pulse text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">Loading deployments...</p>
          </div>
        ) : deployments.length === 0 ? (
          <div className="text-center py-10">
            <Icon name="info" className="text-[#e4e4e7]-dim/60 text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">No deployments found</p>
          </div>
        ) : (
          deployments.map(deployment => {
            const health = getHealthStatus(deployment);
            return (
              <div
                key={deployment.name}
                className="card card-hover rounded-lg p-3.5 border border-purple-500/15"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <Icon name="deployments" className="text-[#71717a]" size="sm" />
                      <h3 className="font-bold text-[#e4e4e7] text-base">{deployment.name}</h3>
                      <span className={`px-2 py-0.5 rounded text-xs font-semibold ${health.color}`}>
                        {health.status}
                      </span>
                    </div>
                    <div className="grid grid-cols-2 md:grid-cols-5 gap-2.5 text-xs">
                      <div>
                        <span className="text-[#71717a]">Namespace:</span>
                        <div className="text-[#3b82f6]">{deployment.namespace}</div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Replicas:</span>
                        <div className="text-[#e4e4e7]">{deployment.replicas}</div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Ready:</span>
                        <div className={deployment.ready === deployment.replicas ? 'text-emerald-400' : 'text-yellow-400'}>
                          {deployment.ready}/{deployment.replicas}
                        </div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Available:</span>
                        <div className={deployment.available === deployment.replicas ? 'text-emerald-400' : 'text-yellow-400'}>
                          {deployment.available}/{deployment.replicas}
                        </div>
                      </div>
                      <div>
                        <span className="text-[#71717a]">Age:</span>
                        <div className="text-[#e4e4e7]">{deployment.age}</div>
                      </div>
                    </div>
                    {deployment.ready !== deployment.replicas && (
                      <div className="mt-2 text-xs text-yellow-400">
                        ⚠ Replica mismatch: {deployment.replicas - deployment.ready} pods not ready
                      </div>
                    )}
                  </div>
                </div>
              </div>
            );
          })
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
