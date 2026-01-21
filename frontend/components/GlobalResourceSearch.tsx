'use client';

import { useState, useEffect, useRef } from 'react';
import { apiClient, SearchResult } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface GlobalResourceSearchProps {
  onClose: () => void;
  onNavigate: (type: string, name: string, namespace: string) => void;
}

export default function GlobalResourceSearch({ onClose, onNavigate }: GlobalResourceSearchProps) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Focus input on mount
    if (inputRef.current) {
      inputRef.current.focus();
    }

    // Close on escape
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(async () => {
      if (query.trim().length > 1) {
        setLoading(true);
        try {
          const data = await apiClient.searchResources(query);
          setResults(data);
        } catch (error) {
          console.error("Search failed", error);
        } finally {
          setLoading(false);
        }
      } else {
        setResults([]);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query]);

  const handleSelect = (result: SearchResult) => {
    if (result.type === 'service') {
      onNavigate(result.type, result.name, result.namespace);
      // Navigation component handles routing to network-graph
    } else if (result.type === 'pod') {
      onNavigate(result.type, result.name, result.namespace);
      // Navigation component handles routing to resources
    } else {
      onNavigate(result.type, result.name, result.namespace);
    }
    onClose();
  };

  const groupByType = (results: SearchResult[]) => {
    const groups: Record<string, SearchResult[]> = {};
    results.forEach(r => {
      if (!groups[r.type]) groups[r.type] = [];
      groups[r.type].push(r);
    });
    return groups;
  };

  const groups = groupByType(results);

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div
        className="w-full max-w-2xl bg-[#111118] border border-[rgba(255,255,255,0.08)] rounded-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200"
        onClick={e => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 px-4 py-3 border-b border-[rgba(255,255,255,0.08)]">
          <Icon name="scan" className="text-[#a1a1aa]" size="sm" />
          <input
            ref={inputRef}
            type="text"
            className="flex-1 bg-transparent border-none text-[#e4e4e7] placeholder-[#71717a] focus:ring-0 text-lg outline-none"
            placeholder="Search resources (services, pods, nodes)..."
            value={query}
            onChange={e => setQuery(e.target.value)}
          />
          <div className="px-2 py-1 rounded bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[10px] text-[#71717a] font-medium uppercase tracking-wider">
            ESC
          </div>
        </div>

        <div className="max-h-[60vh] overflow-y-auto p-2">
          {loading && (
            <div className="py-8 text-center text-[#71717a]">
              <Icon name="loading" className="inline-block animate-spin mb-2" />
              <div className="text-sm">Searching...</div>
            </div>
          )}

          {!loading && results.length === 0 && query.length > 1 && (
            <div className="py-8 text-center text-[#71717a]">
              <div className="text-sm">No results found for "{query}"</div>
            </div>
          )}

          {!loading && results.length === 0 && query.length <= 1 && (
            <div className="py-12 text-center text-[#71717a]">
              <Icon name="scan" className="inline-block mb-3 opacity-20 text-4xl" />
              <div className="text-sm">Type to search for resources across the cluster</div>
            </div>
          )}

          {Object.entries(groups).map(([type, items]) => (
            <div key={type} className="mb-4 last:mb-0">
              <div className="px-3 py-2 text-xs font-semibold text-[#71717a uppercase tracking-wider flex items-center gap-2">
                 <span className="capitalize">{type}s</span>
                 <span className="w-full h-px bg-[rgba(255,255,255,0.04)]"></span>
              </div>
              <div className="space-y-1">
                {items.map((item, idx) => (
                  <button
                    key={idx}
                    onClick={() => handleSelect(item)}
                    className="w-full text-left px-3 py-2.5 rounded-lg hover:bg-[#3b82f6]/10 hover:text-[#3b82f6] group transition-all flex items-center justify-between"
                  >
                    <div className="flex items-center gap-3">
                       <div className={`p-1.5 rounded bg-[rgba(255,255,255,0.02)] border border-[rgba(255,255,255,0.04)] group-hover:border-[#3b82f6]/30 text-[#71717a] group-hover:text-[#3b82f6]`}>
                         <Icon name={getIconForType(item.type)} size="sm" />
                       </div>
                       <div>
                         <div className="font-medium text-[#e4e4e7] group-hover:text-[#3b82f6]">{item.name}</div>
                         <div className="text-xs text-[#71717a]">{item.namespace}</div>
                       </div>
                    </div>
                    {item.status && (
                       <div className={`px-2 py-0.5 rounded text-[10px] font-medium border ${getStatusColor(item.status)}`}>
                         {item.status}
                       </div>
                    )}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="px-4 py-2 bg-[#0d0d12] border-t border-[rgba(255,255,255,0.08)] text-[10px] text-[#71717a] flex justify-between">
          <div><span className="text-[#e4e4e7]">↑↓</span> to navigate</div>
          <div><span className="text-[#e4e4e7]">↵</span> to select</div>
        </div>
      </div>
    </div>
  );
}

function getIconForType(type: string): any {
  switch (type) {
    case 'service': return 'network';
    case 'pod': return 'pods';
    case 'node': return 'deployments'; // Using available icon
    case 'deployment': return 'deployments';
    default: return 'pods';
  }
}

function getStatusColor(status: string): string {
  const s = status.toLowerCase();
  if (s === 'running' || s === 'ready') return "bg-[#10b981]/10 text-[#10b981] border-[#10b981]/20";
  if (s === 'pending' || s === 'unknown') return "bg-[#f59e0b]/10 text-[#f59e0b] border-[#f59e0b]/20";
  return "bg-[#ef4444]/10 text-[#ef4444] border-[#ef4444]/20";
}
