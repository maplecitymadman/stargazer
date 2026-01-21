'use client';

import { useState, useEffect, useMemo } from 'react';
import apiClient, { PathTrace } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface PathTracerProps {
  namespace?: string;
}

interface ServiceOption {
  name: string;
  namespace: string;
  type: string;
  display: string;
}

export default function PathTracer({ namespace }: PathTracerProps) {
  const [source, setSource] = useState('');
  const [destination, setDestination] = useState('');
  const [sourceSearch, setSourceSearch] = useState('');
  const [destinationSearch, setDestinationSearch] = useState('');
  const [services, setServices] = useState<ServiceOption[]>([]);
  const [loadingServices, setLoadingServices] = useState(false);
  const [tracing, setTracing] = useState(false);
  const [traceResult, setTraceResult] = useState<PathTrace | null>(null);
  const [showSourceDropdown, setShowSourceDropdown] = useState(false);
  const [showDestDropdown, setShowDestDropdown] = useState(false);

  // Load services on mount
  useEffect(() => {
    loadServices();
  }, [namespace]);

  const loadServices = async () => {
    setLoadingServices(true);
    try {
      const data = await apiClient.getServices(namespace || 'all');
      if (data.services) {
        setServices(data.services);
      }
    } catch (error) {
      console.error('Error loading services:', error);
    } finally {
      setLoadingServices(false);
    }
  };

  // Filter services based on search
  const filteredSourceServices = useMemo(() => {
    if (!sourceSearch) return services;
    const search = sourceSearch.toLowerCase();
    return services.filter(s =>
      s.display.toLowerCase().includes(search) ||
      s.name.toLowerCase().includes(search) ||
      s.type.toLowerCase().includes(search)
    );
  }, [services, sourceSearch]);

  const filteredDestServices = useMemo(() => {
    if (!destinationSearch) return services;
    const search = destinationSearch.toLowerCase();
    return services.filter(s =>
      s.display.toLowerCase().includes(search) ||
      s.name.toLowerCase().includes(search) ||
      s.type.toLowerCase().includes(search)
    );
  }, [services, destinationSearch]);

  const handleSourceSelect = (service: ServiceOption) => {
    if (service.type === 'ingress' || service.type === 'egress' || service.type === 'external') {
      setSource(service.name);
    } else {
      setSource(service.namespace ? `${service.namespace}/${service.name}` : service.name);
    }
    setSourceSearch('');
    setShowSourceDropdown(false);
  };

  const handleDestSelect = (service: ServiceOption) => {
    if (service.type === 'ingress' || service.type === 'egress' || service.type === 'external') {
      setDestination(service.name);
    } else {
      setDestination(service.namespace ? `${service.namespace}/${service.name}` : service.name);
    }
    setDestinationSearch('');
    setShowDestDropdown(false);
  };

  const handleTrace = async () => {
    if (!source || !destination) return;

    setTracing(true);
    try {
      const result = await apiClient.tracePath(source, destination, namespace);
      setTraceResult(result);
    } catch (error) {
      console.error('Error tracing path:', error);
      setTraceResult({
        source,
        destination,
        path: [],
        allowed: false,
        reason: 'Failed to trace path'
      });
    } finally {
      setTracing(false);
    }
  };

  return (
    <div className="card rounded-lg p-5">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-[#e4e4e7]">Connection Path Tracer</h2>
        <button
          onClick={loadServices}
          disabled={loadingServices}
          className="px-3 py-1.5 text-sm bg-[#1a1a24] hover:bg-[#252530] text-[#71717a] rounded-md disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2 cursor-pointer active:scale-95"
          aria-label="Refresh services list"
        >
          <Icon name={loadingServices ? "loading" : "refresh"} className="text-[#71717a]" size="sm" />
          Refresh Services
        </button>
      </div>

      <div className="space-y-4">
        {/* Source Selection */}
        <div className="relative">
          <label className="block text-sm text-[#71717a] mb-2">Source</label>
          <div className="relative">
            <input
              type="text"
              value={sourceSearch || source}
              onChange={(e) => {
                setSourceSearch(e.target.value);
                setShowSourceDropdown(true);
                if (!e.target.value) {
                  setSource('');
                }
              }}
              onFocus={() => setShowSourceDropdown(true)}
              placeholder="Search or type: ingress-gateway, namespace/service, or egress-gateway"
              className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7] focus:outline-none focus:ring-2 focus:ring-[#3b82f6]"
            />
            {source && (
              <button
                onClick={() => {
                  setSource('');
                  setSourceSearch('');
                }}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-[#71717a] hover:text-[#e4e4e7]"
              >
                <Icon name="close" size="sm" />
              </button>
            )}
            {showSourceDropdown && filteredSourceServices.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md shadow-lg max-h-60 overflow-y-auto">
                {filteredSourceServices.map((service, idx) => (
                  <button
                    key={idx}
                    onClick={() => handleSourceSelect(service)}
                    className="w-full text-left px-3 py-2 hover:bg-[#252530] text-[#e4e4e7] transition-colors flex items-center justify-between"
                  >
                    <span>{service.display}</span>
                    <span className="text-xs text-[#71717a] px-2 py-0.5 rounded bg-[#0a0a0f]">
                      {service.type}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Destination Selection */}
        <div className="relative">
          <label className="block text-sm text-[#71717a] mb-2">Destination</label>
          <div className="relative">
            <input
              type="text"
              value={destinationSearch || destination}
              onChange={(e) => {
                setDestinationSearch(e.target.value);
                setShowDestDropdown(true);
                if (!e.target.value) {
                  setDestination('');
                }
              }}
              onFocus={() => setShowDestDropdown(true)}
              placeholder="Search or type: egress-gateway, external, or namespace/service"
              className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7] focus:outline-none focus:ring-2 focus:ring-[#3b82f6]"
            />
            {destination && (
              <button
                onClick={() => {
                  setDestination('');
                  setDestinationSearch('');
                }}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-[#71717a] hover:text-[#e4e4e7]"
              >
                <Icon name="close" size="sm" />
              </button>
            )}
            {showDestDropdown && filteredDestServices.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md shadow-lg max-h-60 overflow-y-auto">
                {filteredDestServices.map((service, idx) => (
                  <button
                    key={idx}
                    onClick={() => handleDestSelect(service)}
                    className="w-full text-left px-3 py-2 hover:bg-[#252530] text-[#e4e4e7] transition-colors flex items-center justify-between"
                  >
                    <span>{service.display}</span>
                    <span className="text-xs text-[#71717a] px-2 py-0.5 rounded bg-[#0a0a0f]">
                      {service.type}
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        <button
          onClick={handleTrace}
          disabled={tracing || !source || !destination}
          className="w-full px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-md disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2 cursor-pointer active:scale-95"
          aria-label="Trace connection path"
        >
          {tracing ? (
            <>
              <Icon name="loading" className="animate-spin" size="sm" />
              <span>Tracing...</span>
            </>
          ) : (
            <>
              <Icon name="network" size="sm" />
              <span>Trace Path</span>
            </>
          )}
        </button>
      </div>

      {traceResult && (
        <div className="mt-6 p-4 bg-[#1a1a24] rounded-lg border border-[rgba(255,255,255,0.08)]">
          <div className={`flex items-center gap-2 mb-3 ${traceResult.allowed ? 'text-[#10b981]' : 'text-[#ef4444]'}`}>
            <Icon name={traceResult.allowed ? 'healthy' : 'degraded'} />
            <span className="font-bold">
              {traceResult.allowed ? 'Path Allowed' : 'Path Blocked'}
            </span>
          </div>

          {traceResult.reason && (
            <p className="text-sm text-[#71717a] mb-4">{traceResult.reason}</p>
          )}

          {traceResult.path && traceResult.path.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-sm font-semibold text-[#e4e4e7]">Path Hops:</h3>
              <div className="space-y-2">
                {traceResult.path.map((hop, idx) => (
                  <div key={idx} className="flex items-start gap-3">
                    {idx > 0 && (
                      <div className="flex flex-col items-center pt-2">
                        <div className="w-0.5 h-4 bg-[#71717a]"></div>
                        <span className="text-[#71717a]">↓</span>
                        <div className="w-0.5 h-4 bg-[#71717a]"></div>
                      </div>
                    )}
                    <div
                      className={`flex-1 p-3 rounded border ${
                        hop.allowed
                          ? 'border-[#10b981]/30 bg-[#10b981]/5'
                          : 'border-[#ef4444]/30 bg-[#ef4444]/5'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <div className="flex items-center gap-2">
                          <Icon
                            name={hop.allowed ? "check" : "critical"}
                            className={hop.allowed ? "text-[#10b981]" : "text-[#ef4444]"}
                            size="sm"
                          />
                          <span className="text-sm font-medium text-[#e4e4e7]">
                            {hop.from} → {hop.to}
                          </span>
                          {hop.service_mesh && (
                            <span className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-300">
                              {hop.service_mesh}
                            </span>
                          )}
                        </div>
                        <span className={`text-xs px-2 py-1 rounded ${
                          hop.allowed
                            ? 'bg-[#10b981]/20 text-[#10b981]'
                            : 'bg-[#ef4444]/20 text-[#ef4444]'
                        }`}>
                          {hop.allowed ? 'ALLOWED' : 'BLOCKED'}
                        </span>
                      </div>
                      {hop.reason && (
                        <p className="text-xs text-[#71717a] mt-1">{hop.reason}</p>
                      )}
                      {hop.policies && hop.policies.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          {hop.policies.map((policy, pIdx) => (
                            <span key={pIdx} className="text-xs px-2 py-0.5 rounded bg-[#f59e0b]/20 text-[#f59e0b]">
                              {policy}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
