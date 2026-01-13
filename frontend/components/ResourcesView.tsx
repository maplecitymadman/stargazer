'use client';

import { useState, useEffect } from 'react';
import apiClient from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface ResourcesViewProps {
  namespace?: string;
}

const resourceIcons: Record<string, string> = {
  pods: 'pods',
  deployments: 'deployments',
  services: 'events',
  configmaps: 'info',
  secrets: 'scan',
  ingresses: 'events',
  statefulsets: 'deployments',
  daemonsets: 'deployments',
  jobs: 'terminal',
  cronjobs: 'terminal',
  persistentvolumeclaims: 'info',
};

const resourceLabels: Record<string, string> = {
  pods: 'PODS',
  deployments: 'DEPLOYMENTS',
  services: 'SERVICES',
  configmaps: 'CONFIGMAPS',
  secrets: 'SECRETS',
  ingresses: 'INGRESSES',
  statefulsets: 'STATEFULSETS',
  daemonsets: 'DAEMONSETS',
  jobs: 'JOBS',
  cronjobs: 'CRONJOBS',
  persistentvolumeclaims: 'PERSISTENT VOLUME CLAIMS',
};

export default function ResourcesView({ namespace }: ResourcesViewProps) {
  const [resources, setResources] = useState<any>({});
  const [loading, setLoading] = useState(true);
  const [selectedType, setSelectedType] = useState<string | null>(null);

  useEffect(() => {
    loadResources();
  }, [namespace]);

  const loadResources = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getAllResources(namespace);
      setResources(data.resources || {});
    } catch (error) {
      // Error loading resources - will show empty state
    } finally {
      setLoading(false);
    }
  };

  const renderResourceValue = (key: string, value: any) => {
    if (Array.isArray(value)) {
      return value.length > 0 ? value.join(', ') : 'None';
    }
    if (typeof value === 'object' && value !== null) {
      return Object.keys(value).length > 0 ? `${Object.keys(value).length} keys` : 'Empty';
    }
    return String(value || 'N/A');
  };

  const renderResourceTable = (resourceType: string, items: any[]) => {
    if (items.length === 0) {
      return (
        <div className="text-center py-8 text-[#71717a] text-sm">
          No {resourceLabels[resourceType].toLowerCase()} found
        </div>
      );
    }

    // Get all unique keys from all items
    const allKeys = new Set<string>();
    items.forEach(item => {
      Object.keys(item).forEach(key => {
        if (key !== 'labels' && key !== 'namespace') {
          allKeys.add(key);
        }
      });
    });
    const columns = Array.from(allKeys).slice(0, 6); // Limit to 6 columns

    return (
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[rgba(255,255,255,0.08)]">
              <th className="text-left py-2 px-3 text-[#71717a] font-semibold">Name</th>
              {columns.map(col => (
                <th key={col} className="text-left py-2 px-3 text-[#71717a] font-semibold">
                  {col.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {items.map((item, idx) => (
              <tr
                key={item.name || idx}
                className="border-b border-[rgba(255,255,255,0.08)] hover:bg-[#1a1a24] transition-colors"
              >
                <td className="py-2 px-3 text-[#3b82f6] font-semibold">{item.name}</td>
                {columns.map(col => (
                  <td key={col} className="py-2 px-3 text-[#e4e4e7]">
                    {renderResourceValue(col, item[col])}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="card rounded-lg p-8">
        <div className="text-center py-10">
          <Icon name="loading" className="text-[#71717a] animate-pulse text-3xl" />
          <p className="text-[#71717a] mt-2 text-sm">Loading resources...</p>
        </div>
      </div>
    );
  }

  const resourceTypes = Object.keys(resources).filter(key => resources[key].length > 0);
  const totalResources = Object.values(resources).reduce((sum: number, arr: any) => sum + arr.length, 0);

  return (
    <div className="card rounded-lg p-5">
      <div className="flex items-center justify-between mb-5 pb-4 border-b border-[rgba(255,255,255,0.08)]">
        <div className="flex items-center gap-3">
          <Icon name="scan" className="text-[#71717a]" size="sm" />
          <h2 className="text-2xl font-bold text-[#e4e4e7]">All Resources</h2>
          <span className="px-2.5 py-1 rounded-full bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#71717a] text-xs">
            {totalResources} TOTAL
          </span>
        </div>
        <button
          onClick={loadResources}
          className="px-3 py-1.5 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-sm transition-all flex items-center gap-2"
        >
          <Icon name="refresh" className="text-white" size="sm" />
          REFRESH
        </button>
      </div>

      {resourceTypes.length === 0 ? (
        <div className="text-center py-10">
          <Icon name="info" className="text-space-text-dim/60 text-3xl" />
          <p className="text-[#71717a] mt-2 text-sm">No resources found</p>
        </div>
      ) : (
        <div className="space-y-4">
          {resourceTypes.map(resourceType => (
            <div key={resourceType} className="card rounded-lg p-4 border border-[rgba(255,255,255,0.08)]">
              <button
                onClick={() => setSelectedType(selectedType === resourceType ? null : resourceType)}
                className="w-full flex items-center justify-between mb-3 hover:opacity-80 transition-opacity"
              >
                <div className="flex items-center gap-2">
                  <Icon name={resourceIcons[resourceType] as any || 'info'} className="text-[#71717a]" size="sm" />
                  <h3 className="font-bold text-[#e4e4e7] text-base">
                    {resourceLabels[resourceType] || resourceType.replace(/\b\w/g, l => l.toUpperCase())}
                  </h3>
                  <span className="px-2 py-0.5 rounded-full bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#71717a] text-xs">
                    {resources[resourceType].length}
                  </span>
                </div>
                <Icon
                  name="info"
                  className={`text-[#71717a] transition-transform ${selectedType === resourceType ? 'rotate-90' : ''}`}
                  size="sm"
                />
              </button>
              
              {selectedType === resourceType && (
                <div className="mt-3 border-t border-[rgba(255,255,255,0.08)] pt-3">
                  {renderResourceTable(resourceType, resources[resourceType])}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
