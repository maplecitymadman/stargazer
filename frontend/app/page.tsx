'use client';

import { useState, useEffect, useCallback } from 'react';
import { apiClient, ClusterHealth, Issue } from '@/lib/api';
import HealthMetrics from '@/components/HealthMetrics';
import IssuesList from '@/components/IssuesList';
import Navigation from '@/components/Navigation';
import Breadcrumbs from '@/components/Breadcrumbs';
import PodsDetail from '@/components/PodsDetail';
import DeploymentsDetail from '@/components/DeploymentsDetail';
import EventsDetail from '@/components/EventsDetail';
import ResourcesView from '@/components/ResourcesView';
import ServiceTopology from '@/components/ServiceTopology';
import StargazerLogo from '@/components/StargazerLogo';
import Settings from '@/components/Settings';
import { Icon } from '@/components/SpaceshipIcons';

export default function Home() {
  const [health, setHealth] = useState<ClusterHealth | null>(null);
  const [issues, setIssues] = useState<Issue[]>([]);
  const [namespace, setNamespace] = useState<string>('all');
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [currentSection, setCurrentSection] = useState<string>('dashboard');
  const [showDetailModal, setShowDetailModal] = useState<'pods' | 'deployments' | 'events' | null>(null);

  useEffect(() => {
    // Load theme from localStorage
    const savedTheme = localStorage.getItem('stargazer-theme') || 'dark';
    applyTheme(savedTheme as 'dark' | 'light' | 'auto');
    
    // Listen for theme changes from Settings
    const handleThemeChange = (e: CustomEvent) => {
      applyTheme(e.detail);
    };
    window.addEventListener('theme-change', handleThemeChange as EventListener);
    
    // Listen for system theme changes when in auto mode
    const setupAutoTheme = () => {
      const currentTheme = localStorage.getItem('stargazer-theme') || 'dark';
      if (currentTheme === 'auto') {
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        const handleChange = (e: MediaQueryListEvent) => {
          const root = document.documentElement;
          root.classList.toggle('light', !e.matches);
          root.classList.toggle('dark', e.matches);
        };
        mediaQuery.addEventListener('change', handleChange);
        return () => mediaQuery.removeEventListener('change', handleChange);
      }
      return () => {};
    };
    const cleanupAuto = setupAutoTheme();
    
    loadInitialData();
    setupWebSocket();

    return () => {
      window.removeEventListener('theme-change', handleThemeChange as EventListener);
      cleanupAuto();
      if (ws) {
        ws.close();
      }
    };
  }, []);

  const applyTheme = (theme: 'dark' | 'light' | 'auto') => {
    const root = document.documentElement;
    if (theme === 'auto') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      root.classList.toggle('light', !prefersDark);
      root.classList.toggle('dark', prefersDark);
    } else {
      root.classList.remove('light', 'dark');
      root.classList.add(theme);
    }
  };

  useEffect(() => {
    // Only use auto-refresh if WebSocket is not connected
    // WebSocket provides real-time updates, so we don't need polling
    if (autoRefresh && !ws) {
      const interval = setInterval(() => {
        loadData();
      }, 5000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, ws, namespace]); // Include namespace in dependencies

  // Reload issues when namespace changes (if issues are empty)
  useEffect(() => {
    if (namespace && issues.length === 0 && !loading) {
      // Namespace was set but issues are empty, reload them
      loadData();
    }
  }, [namespace, issues.length, loading, loadData]); // Include all dependencies

  // Track if initial load is complete
  const [initialLoadComplete, setInitialLoadComplete] = useState(false);

  // Reload data when switching sections (but not on initial load)
  useEffect(() => {
    // Only reload if initial load is complete and we're not currently loading
    if (initialLoadComplete && currentSection && !loading) {
      loadData();
    }
  }, [currentSection, initialLoadComplete, loading, loadData]); // Include all dependencies

  const loadInitialData = async (preserveNamespace: boolean = false) => {
    try {
      setLoading(true);
      
      // Use 'all' as default namespace
      let currentNs = preserveNamespace ? namespace : 'all';
      if (!preserveNamespace) {
        // If namespace is empty or not set, default to 'all'
        if (!namespace || namespace === '') {
          setNamespace('all');
          currentNs = 'all';
        } else {
          currentNs = namespace;
        }
      }
      
      // Now load health and issues with the correct namespace
      const ns = currentNs === 'all' ? 'all' : currentNs || 'all';
      const [healthData, issuesData] = await Promise.all([
        apiClient.getHealth(ns).catch(err => {
          if (process.env.NODE_ENV === 'development') {
            console.error('Error loading health:', err);
          }
          // Return default health on error
          return {
            pods: { total: 0, healthy: 0 },
            deployments: { total: 0, healthy: 0 },
            events: { warnings: 0, errors: 0 },
            overall_health: 'degraded' as const
          };
        }),
        apiClient.getIssues(ns).catch(err => {
          if (process.env.NODE_ENV === 'development') {
            console.error('Error loading issues:', err);
          }
          return [];
        }),
      ]);
      
      setHealth(healthData);
      setIssues(issuesData);
      setInitialLoadComplete(true);
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Error loading initial data:', error);
      }
      // Set default states on error
      setHealth({
        pods: { total: 0, healthy: 0 },
        deployments: { total: 0, healthy: 0 },
        events: { warnings: 0, errors: 0 },
        overall_health: 'degraded'
      });
      setIssues([]);
      setInitialLoadComplete(true);
    } finally {
      setLoading(false);
    }
  };

  const handleContextChange = () => {
    // Reload all data when context changes
    loadInitialData();
  };

  const loadData = async () => {
    try {
      const ns = namespace === 'all' ? 'all' : namespace || undefined;
      const [healthData, issuesData] = await Promise.all([
        apiClient.getHealth(ns).catch(err => {
          console.error('Error loading health:', err);
          // Return default health on error, but don't overwrite existing data
          return null;
        }),
        apiClient.getIssues(ns).catch(err => {
          console.error('Error loading issues:', err);
          return null; // Return null instead of empty array to preserve existing data
        }),
      ]);
      // Only update if we got valid data
      if (healthData) {
        setHealth(healthData);
      }
      // Only update issues if we got valid data (not null)
      if (issuesData !== null) {
        setIssues(issuesData);
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Error loading data:', error);
      }
      // Don't clear existing data on error
    }
  };

  const setupWebSocket = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    try {
      const websocket = new WebSocket(wsUrl);
      
      websocket.onopen = () => {
        setWs(websocket);
      };
      
      websocket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'initial' || data.type === 'update') {
            // Only update if we have valid data
            if (data.health) {
              setHealth(data.health);
            }
            // Only update issues if we have valid data (array)
            // Don't overwrite with empty array if we already have issues (might be a transient error)
            if (Array.isArray(data.issues)) {
              // Only update if we got issues, or if current issues list is empty
              setIssues(prevIssues => {
                // If new data has issues, always update
                if (data.issues.length > 0) {
                  return data.issues;
                }
                // If new data is empty but we have existing issues, keep existing (might be transient)
                if (prevIssues.length > 0 && data.issues.length === 0) {
                  return prevIssues;
                }
                // Otherwise update (both empty, or initial load)
                return data.issues;
              });
            }
          }
        } catch (e) {
          // Invalid JSON - ignore
        }
      };
      
      websocket.onerror = () => {
        setWs(null);
      };
      
      websocket.onclose = () => {
        setWs(null);
        setTimeout(setupWebSocket, 5000);
      };
    } catch (error) {
      // Silently handle WebSocket setup errors
      setWs(null);
    }
  };

  const handleRefresh = async () => {
    setLoading(true);
    try {
      await loadData();
    } finally {
      setLoading(false);
    }
  };

  const handleSectionChange = (section: string) => {
    setCurrentSection(section);
    setShowDetailModal(null);
    // Reload data when switching to issues section to ensure it's populated
    if (section === 'issues' && issues.length === 0) {
      loadData();
    }
  };

  const getBreadcrumbs = () => {
    const base = [{ label: 'DASHBOARD', onClick: () => setCurrentSection('dashboard') }];
    
    switch (currentSection) {
      case 'topology':
        return [...base, { label: 'SERVICE TOPOLOGY' }];
      case 'resources':
        return [...base, { label: 'RESOURCES' }];
      case 'pods':
        return [...base, { label: 'PODS' }];
      case 'deployments':
        return [...base, { label: 'DEPLOYMENTS' }];
      case 'events':
        return [...base, { label: 'EVENTS' }];
      case 'issues':
        return [...base, { label: 'ISSUES' }];
      case 'settings':
        return [...base, { label: 'SETTINGS' }];
      default:
        return base;
    }
  };

  if (loading && !health) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#0a0a0f]">
          <div className="text-center">
            <div className="mb-6">
              <StargazerLogo size="lg" />
            </div>
            <div className="text-xl font-medium text-[#e4e4e7] mb-2 flex items-center justify-center gap-2">
              <Icon name="loading" className="text-[#3b82f6] animate-pulse" />
              <span>Initializing...</span>
            </div>
            <div className="text-sm text-[#71717a]">Scanning cluster</div>
          </div>
        </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#0a0a0f] flex">
      {/* Sidebar Navigation */}
      <Navigation
        currentSection={currentSection}
        onSectionChange={handleSectionChange}
        namespace={namespace}
        onNamespaceChange={(ns) => {
          setNamespace(ns);
          loadInitialData(true); // Preserve the selected namespace
        }}
        onContextChange={handleContextChange}
      />

      {/* Main Content */}
      <div className="flex-1 lg:ml-64">
        {/* Top Bar */}
        <header className="glass border-b border-[rgba(255,255,255,0.08)] sticky top-0 z-20">
          <div className="px-6 py-4">
            <div className="flex items-center justify-between">
              <div>
                <Breadcrumbs items={getBreadcrumbs()} />
                <h1 className="text-xl font-semibold text-[#e4e4e7] tracking-tight mt-1">
                  {currentSection === 'dashboard' && 'Dashboard'}
                  {currentSection === 'topology' && 'Service Topology'}
                  {currentSection === 'resources' && 'Resources'}
                  {currentSection === 'pods' && 'Pods'}
                  {currentSection === 'deployments' && 'Deployments'}
                  {currentSection === 'events' && 'Events'}
                  {currentSection === 'issues' && 'Issues'}
                  {currentSection === 'settings' && 'Settings'}
                </h1>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={handleRefresh}
                  disabled={loading}
                  className="px-4 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] disabled:opacity-50 transition-all flex items-center gap-2 text-sm font-medium"
                >
                  <Icon name={loading ? "loading" : "refresh"} className="text-white" size="sm" />
                  Refresh
                </button>
                <label className="flex items-center gap-2 cursor-pointer px-3 py-2 rounded-md hover:bg-[#111118] transition-colors">
                  <input
                    type="checkbox"
                    checked={autoRefresh}
                    onChange={(e) => setAutoRefresh(e.target.checked)}
                    className="w-4 h-4 rounded border-[rgba(255,255,255,0.2)] bg-[#1a1a24] text-[#3b82f6] focus:ring-[#3b82f6]"
                  />
                  <span className="text-sm text-[#a1a1aa]">Auto</span>
                </label>
                {ws && (
                  <span className="text-xs text-[#10b981] flex items-center gap-2 px-3 py-1.5 rounded-md bg-[#10b981]/10 border border-[#10b981]/20">
                    <span className="w-1.5 h-1.5 bg-[#10b981] rounded-full animate-pulse"></span>
                    Live
                  </span>
                )}
              </div>
            </div>
          </div>
        </header>

        {/* Content Area */}
        <main className="px-6 py-6">
          {currentSection === 'dashboard' && (
            <>
              {loading && (
                <div className="space-y-6">
                  {/* Health Metrics Skeleton */}
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="card rounded-lg p-5">
                        <div className="h-3 bg-[#1a1a24] rounded w-1/4 mb-3 animate-pulse" />
                        <div className="h-8 bg-[#1a1a24] rounded w-1/3 mb-2 animate-pulse" />
                        <div className="h-3 bg-[#1a1a24] rounded w-1/2 mb-3 animate-pulse" />
                        <div className="h-1 bg-[#1a1a24] rounded-full animate-pulse" />
                      </div>
                    ))}
                  </div>
                  {/* Issues List Skeleton */}
                  <div className="card rounded-lg p-5">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
                      {Array.from({ length: 3 }).map((_, i) => (
                        <div key={i} className="h-24 bg-[#1a1a24] rounded-lg animate-pulse" />
                      ))}
                    </div>
                    <div className="space-y-3">
                      {Array.from({ length: 3 }).map((_, i) => (
                        <div key={i} className="h-20 bg-[#1a1a24] rounded-lg animate-pulse" />
                      ))}
                    </div>
                  </div>
                </div>
              )}
              {!loading && health && (
                <>
                  <div className="mb-6">
                    <HealthMetrics 
                      health={health} 
                      namespace={namespace}
                      onPodsClick={() => {
                        setCurrentSection('pods');
                        setShowDetailModal(null);
                      }}
                      onDeploymentsClick={() => {
                        setCurrentSection('deployments');
                        setShowDetailModal(null);
                      }}
                      onEventsClick={() => {
                        setCurrentSection('events');
                        setShowDetailModal(null);
                      }}
                    />
                  </div>
                  <div className="mb-6">
                    <IssuesList issues={issues} />
                  </div>
                </>
              )}
              {!loading && !health && (
                <div className="card p-8 text-center">
                  <Icon name="info" className="text-[#71717a] mx-auto mb-4" size="md" />
                  <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">Unable to load dashboard</h3>
                  <p className="text-sm text-[#71717a] mb-4">
                    {health === null || (health && typeof health === 'object' && 'error' in health)
                      ? "Kubeconfig not configured. Please configure it in Settings → Cluster."
                      : "Check your Kubernetes connection and try refreshing."}
                  </p>
                  <div className="flex gap-3 justify-center">
                    <button
                      onClick={() => {
                        setCurrentSection('settings');
                        // Set cluster tab active after a brief delay to ensure Settings component is mounted
                        setTimeout(() => {
                          const event = new CustomEvent('settings-tab-change', { detail: 'cluster' });
                          window.dispatchEvent(event);
                        }, 100);
                      }}
                      className="px-4 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] transition-colors cursor-pointer"
                    >
                      Configure
                    </button>
                    <button
                      onClick={handleRefresh}
                      className="px-4 py-2 border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] rounded-md hover:bg-[#111118] transition-colors cursor-pointer"
                    >
                      Retry
                    </button>
                  </div>
                </div>
              )}
            </>
          )}

          {currentSection === 'topology' && (
            <ServiceTopology namespace={namespace} />
          )}

          {currentSection === 'resources' && (
            <ResourcesView namespace={namespace} />
          )}

          {currentSection === 'issues' && (
            <div className="mb-6">
              {loading && issues.length === 0 && (
                <div className="space-y-3">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="card rounded-lg p-4">
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex-1 space-y-2">
                          <div className="flex items-center gap-2">
                            <div className="h-5 w-5 bg-[#1a1a24] rounded animate-pulse" />
                            <div className="h-6 w-20 bg-[#1a1a24] rounded animate-pulse" />
                            <div className="h-5 w-32 bg-[#1a1a24] rounded animate-pulse" />
                          </div>
                          <div className="h-4 w-full bg-[#1a1a24] rounded animate-pulse" />
                          <div className="h-4 w-3/4 bg-[#1a1a24] rounded animate-pulse" />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {!loading && issues.length === 0 && (
                <div className="card p-8 text-center">
                  <Icon name="healthy" className="text-[#10b981] mx-auto mb-4" size="md" />
                  <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">No Issues Found</h3>
                  <p className="text-sm text-[#71717a] mb-4">
                    Your cluster appears to be healthy, or there was an error loading issues.
                  </p>
                  <button
                    onClick={handleRefresh}
                    className="px-4 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] transition-colors"
                  >
                    Refresh
                  </button>
                </div>
              )}
              {issues.length > 0 && <IssuesList issues={issues} />}
            </div>
          )}


          {/* Detail views in main content when navigating via sidebar */}
          {currentSection === 'pods' && (
            <PodsDetail namespace={namespace} />
          )}

          {currentSection === 'deployments' && (
            <DeploymentsDetail namespace={namespace} />
          )}

          {currentSection === 'events' && (
            <EventsDetail namespace={namespace} />
          )}

          {currentSection === 'settings' && (
            <Settings />
          )}
        </main>
      </div>

      {/* Detail Modals */}
      {showDetailModal === 'pods' && (
        <PodsDetail namespace={namespace} onClose={() => setShowDetailModal(null)} />
      )}
      {showDetailModal === 'deployments' && (
        <DeploymentsDetail namespace={namespace} onClose={() => setShowDetailModal(null)} />
      )}
      {showDetailModal === 'events' && (
        <EventsDetail namespace={namespace} onClose={() => setShowDetailModal(null)} />
      )}
    </div>
  );
}
