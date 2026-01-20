'use client';

import { useState, useEffect } from 'react';
import { apiClient, KubernetesContext } from '@/lib/api';
import { Icon } from './SpaceshipIcons';
import toast from 'react-hot-toast';

interface ContextSelectorProps {
  onContextChange?: () => void;
}

export default function ContextSelector({ onContextChange }: ContextSelectorProps) {
  const [contexts, setContexts] = useState<KubernetesContext[]>([]);
  const [currentContext, setCurrentContext] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [pendingContext, setPendingContext] = useState<string | null>(null);

  useEffect(() => {
    loadContexts();
  }, []);

  const loadContexts = async () => {
    try {
      setLoading(true);
      const [contextsData, currentData] = await Promise.all([
        apiClient.getContexts(),
        apiClient.getCurrentContext(),
      ]);
      setContexts(contextsData.contexts || []);
      setCurrentContext(currentData.context || null);
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to load contexts:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSwitchContext = async (contextName: string) => {
    if (contextName === currentContext) {
      setIsOpen(false);
      return;
    }

    // Show confirmation modal
    setPendingContext(contextName);
  };

  const confirmSwitchContext = async () => {
    if (!pendingContext) return;

    try {
      await apiClient.switchContext(pendingContext);
      setCurrentContext(pendingContext);
      setIsOpen(false);
      setPendingContext(null);
      toast.success(`Switched to cluster: ${pendingContext}`);
      
      // Reload page to refresh all data
      if (onContextChange) {
        onContextChange();
      } else {
        window.location.reload();
      }
    } catch (error: any) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to switch context:', error);
      }
      toast.error(`Failed to switch context: ${error.message || 'Unknown error'}`);
      setPendingContext(null);
    }
  };

  const cancelSwitchContext = () => {
    setPendingContext(null);
  };

  const currentContextInfo = contexts.find(c => c.name === currentContext);

  if (loading) {
    return (
      <div className="px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">
        <div className="text-xs text-[#71717a] mb-1">Cluster</div>
        <div className="h-4 w-32 bg-[#1a1a24] rounded animate-pulse" />
      </div>
    );
  }

  if (contexts.length === 0) {
    return (
      <div className="px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">
        <div className="text-xs text-[#71717a] mb-1">Cluster</div>
        <div className="text-sm text-[#71717a]">No contexts found</div>
      </div>
    );
  }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] hover:border-[rgba(255,255,255,0.12)] transition-all text-left cursor-pointer"
        aria-label="Select cluster context"
      >
        <div className="text-xs text-[#71717a] mb-1">Cluster</div>
        <div className="text-sm text-[#3b82f6] font-medium truncate">
          {currentContextInfo?.cluster || currentContext || 'Unknown'}
        </div>
        {currentContextInfo && (
          <div className="text-xs text-[#71717a] mt-1 truncate">
            {currentContextInfo.cloud_provider}
          </div>
        )}
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute top-full left-0 mt-2 w-80 glass border border-[rgba(255,255,255,0.08)] rounded-md shadow-xl z-20 max-h-96 overflow-y-auto">
            <div className="p-3 border-b border-[rgba(255,255,255,0.08)]">
              <div className="text-xs font-medium text-[#71717a] mb-2">
                Switch Cluster Context ({contexts.length} available)
              </div>
            </div>
            <div className="p-2 space-y-1">
              {contexts.map((context) => (
                <button
                  key={context.name}
                  onClick={() => handleSwitchContext(context.name)}
                  className={`w-full text-left px-3 py-2 rounded-md text-sm transition-all cursor-pointer active:scale-[0.98] ${
                    context.name === currentContext
                      ? 'bg-[#3b82f6] text-white'
                      : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[#111118]'
                  }`}
                  aria-label={`Switch to ${context.name}`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-medium truncate">{context.name}</span>
                    {context.name === currentContext && (
                      <Icon name="check" size="sm" className="text-white" />
                    )}
                  </div>
                  <div className="text-xs opacity-75 truncate">
                    {context.cloud_provider}
                  </div>
                  <div className="text-xs opacity-60 truncate mt-1">
                    {context.cluster}
                  </div>
                </button>
              ))}
            </div>
          </div>
        </>
      )}

      {/* Confirmation Modal */}
      {pendingContext && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="glass border border-[rgba(255,255,255,0.08)] rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">Switch Cluster Context?</h3>
            <p className="text-sm text-[#71717a] mb-4">
              You are about to switch from <span className="text-[#e4e4e7]">{currentContext}</span> to <span className="text-[#e4e4e7]">{pendingContext}</span>.
              This will reload all data and may take a moment.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={cancelSwitchContext}
                className="px-4 py-2 border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] rounded-md hover:bg-[#111118] transition-all cursor-pointer active:scale-95"
              >
                Cancel
              </button>
              <button
                onClick={confirmSwitchContext}
                className="px-4 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] transition-all cursor-pointer active:scale-95"
              >
                Switch
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
