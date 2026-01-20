'use client';

import { useState } from 'react';
import apiClient, { PathTrace } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface PathTracerProps {
  namespace?: string;
}

export default function PathTracer({ namespace }: PathTracerProps) {
  const [source, setSource] = useState('');
  const [destination, setDestination] = useState('');
  const [tracing, setTracing] = useState(false);
  const [traceResult, setTraceResult] = useState<PathTrace | null>(null);

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
      <h2 className="text-xl font-bold text-[#e4e4e7] mb-4">Connection Path Tracer</h2>
      
      <div className="space-y-4">
        <div>
          <label className="block text-sm text-[#71717a] mb-2">Source</label>
          <input
            type="text"
            value={source}
            onChange={(e) => setSource(e.target.value)}
            placeholder="ingress-gateway or namespace/service"
            className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7]"
          />
        </div>
        
        <div>
          <label className="block text-sm text-[#71717a] mb-2">Destination</label>
          <input
            type="text"
            value={destination}
            onChange={(e) => setDestination(e.target.value)}
            placeholder="egress-gateway, external, or namespace/service"
            className="w-full px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7]"
          />
        </div>
        
        <button
          onClick={handleTrace}
          disabled={tracing || !source || !destination}
          className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-md disabled:opacity-50 transition-all"
        >
          {tracing ? 'Tracing...' : 'Trace Path'}
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
              {traceResult.path.map((hop, idx) => (
                <div
                  key={idx}
                  className={`p-3 rounded border ${
                    hop.allowed
                      ? 'border-[#10b981]/30 bg-[#10b981]/5'
                      : 'border-[#ef4444]/30 bg-[#ef4444]/5'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-[#e4e4e7]">
                      {hop.from} → {hop.to}
                    </span>
                    <span className={`text-xs ${hop.allowed ? 'text-[#10b981]' : 'text-[#ef4444]'}`}>
                      {hop.allowed ? '✓' : '✗'}
                    </span>
                  </div>
                  {hop.reason && (
                    <p className="text-xs text-[#71717a] mt-1">{hop.reason}</p>
                  )}
                  {hop.policies && hop.policies.length > 0 && (
                    <p className="text-xs text-[#f59e0b] mt-1">
                      Policies: {hop.policies.join(', ')}
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
