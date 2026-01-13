'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface Namespace {
  name: string;
  status: string;
  age: string;
}

interface NamespaceSelectorProps {
  currentNamespace: string;
  onNamespaceChange: (namespace: string) => void;
}

export default function NamespaceSelector({ currentNamespace, onNamespaceChange }: NamespaceSelectorProps) {
  const [namespaces, setNamespaces] = useState<Namespace[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadNamespaces();
  }, []);

  const loadNamespaces = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getNamespaces();
      setNamespaces(data.namespaces || []);
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to load namespaces:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSwitchNamespace = (namespaceName: string) => {
    if (namespaceName === currentNamespace) {
      setIsOpen(false);
      return;
    }

    try {
      onNamespaceChange(namespaceName);
      setIsOpen(false);
    } catch (error: any) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to switch namespace:', error);
      }
      alert(`Failed to switch namespace: ${error.message || 'Unknown error'}`);
    }
  };

  if (loading) {
    return (
      <div className="px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">
        <div className="text-xs text-[#71717a] mb-1">Namespace</div>
        <div className="h-4 w-24 bg-[#1a1a24] rounded animate-pulse" />
      </div>
    );
  }

  if (namespaces.length === 0) {
    return (
      <div className="px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">
        <div className="text-xs text-[#71717a] mb-1">Namespace</div>
        <div className="text-sm text-[#3b82f6] font-medium">{currentNamespace || 'default'}</div>
      </div>
    );
  }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] hover:border-[rgba(255,255,255,0.12)] transition-all text-left"
      >
        <div className="text-xs text-[#71717a] mb-1">Namespace</div>
        <div className="text-sm text-[#3b82f6] font-medium truncate">
          {currentNamespace === 'all' || currentNamespace === '' ? '🌐 All Namespaces' : (currentNamespace || 'default')}
        </div>
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute top-full left-0 mt-2 w-64 glass border border-[rgba(255,255,255,0.08)] rounded-md shadow-xl z-20 max-h-96 overflow-y-auto">
            <div className="p-3 border-b border-[rgba(255,255,255,0.08)]">
              <div className="text-xs font-medium text-[#71717a] mb-2">
                Switch Namespace ({namespaces.length} namespaces)
              </div>
            </div>
            <div className="p-2 space-y-1">
              {/* All Namespaces Option */}
              <button
                onClick={() => handleSwitchNamespace('all')}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-all mb-2 ${
                  currentNamespace === 'all' || currentNamespace === ''
                    ? 'bg-[#3b82f6] text-white'
                    : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[#111118]'
                }`}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-medium truncate">🌐 All Namespaces</span>
                  {(currentNamespace === 'all' || currentNamespace === '') && (
                    <Icon name="check" size="sm" className="text-white" />
                  )}
                </div>
                <div className="text-xs opacity-75 truncate">
                  View resources across entire cluster
                </div>
              </button>
              
              <div className="border-t border-[rgba(255,255,255,0.08)] my-2"></div>
              
              {namespaces.map((ns) => (
                <button
                  key={ns.name}
                  onClick={() => handleSwitchNamespace(ns.name)}
                  className={`w-full text-left px-3 py-2 rounded-md text-sm transition-all ${
                    ns.name === currentNamespace
                      ? 'bg-[#3b82f6] text-white'
                      : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[#111118]'
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-medium truncate">{ns.name}</span>
                    {ns.name === currentNamespace && (
                      <Icon name="check" size="sm" className="text-white" />
                    )}
                  </div>
                  <div className="text-xs opacity-75 truncate">
                    Status: {ns.status}
                  </div>
                  <div className="text-xs opacity-60 truncate mt-1">
                    Age: {ns.age}
                  </div>
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
