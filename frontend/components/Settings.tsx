'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

// Kubeconfig Settings Component
function KubeconfigSettings({ config, onUpdate }: { config: any; onUpdate: () => void }) {
  const [kubeconfigPath, setKubeconfigPath] = useState('');
  const [kubeconfigStatus, setKubeconfigStatus] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    loadKubeconfigStatus();
  }, []);

  const loadKubeconfigStatus = async () => {
    try {
      setLoading(true);
      const status = await apiClient.getKubeconfigStatus();
      setKubeconfigStatus(status);
      setKubeconfigPath(status.path || '~/.kube/config');
    } catch (err: any) {
      console.error('Failed to load kubeconfig status:', err);
      setKubeconfigStatus({
        configured: false,
        found: false,
        connected: false,
        error: 'Failed to load kubeconfig status',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSaveKubeconfig = async () => {
    if (!kubeconfigPath.trim()) {
      setError('Kubeconfig path is required');
      return;
    }

    try {
      setSaving(true);
      setError(null);
      setSuccess(false);

      await apiClient.setKubeconfig(kubeconfigPath.trim());

      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);

      // Reload status
      await loadKubeconfigStatus();
      onUpdate();

      // Reload page to reconnect
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to set kubeconfig');
    } finally {
      setSaving(false);
    }
  };

  const handleAutoDetect = async () => {
    // Try to auto-detect by calling the API
    try {
      const status = await apiClient.getKubeconfigStatus();
      if (status.path && status.found) {
        setKubeconfigPath(status.path);
        setError(null);
      } else {
        // Fallback to default
        setKubeconfigPath('~/.kube/config');
      }
    } catch (err) {
      // Fallback to default
      setKubeconfigPath('~/.kube/config');
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="card p-6">
          <div className="text-[#71717a]">Loading kubeconfig status...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="card p-6">
        <h2 className="text-lg font-semibold text-[#e4e4e7] mb-4">Kubernetes Configuration</h2>

        {/* Status Display */}
        {kubeconfigStatus && (
          <div className={`mb-4 p-3 rounded-md border ${
            kubeconfigStatus.connected
              ? 'bg-[#10b981]/10 border-[#10b981]/20'
              : kubeconfigStatus.found
              ? 'bg-[#f59e0b]/10 border-[#f59e0b]/20'
              : 'bg-[#ef4444]/10 border-[#ef4444]/20'
          }`}>
            <div className="flex items-center gap-2 mb-2">
              {kubeconfigStatus.connected ? (
                <>
                  <Icon name="healthy" className="text-[#10b981]" size="sm" />
                  <span className="text-sm font-medium text-[#10b981]">Connected</span>
                </>
              ) : kubeconfigStatus.found ? (
                <>
                  <Icon name="degraded" className="text-[#f59e0b]" size="sm" />
                  <span className="text-sm font-medium text-[#f59e0b]">File Found (Not Connected)</span>
                </>
              ) : (
                <>
                  <Icon name="critical" className="text-[#ef4444]" size="sm" />
                  <span className="text-sm font-medium text-[#ef4444]">Not Configured</span>
                </>
              )}
            </div>
            {kubeconfigStatus.path && (
              <div className="text-xs text-[#71717a] mb-1">
                Path: {kubeconfigStatus.path}
              </div>
            )}
            {kubeconfigStatus.context && (
              <div className="text-xs text-[#71717a] mb-1">
                Context: {kubeconfigStatus.context}
              </div>
            )}
            {kubeconfigStatus.error && (
              <div className="text-xs text-[#ef4444] mt-1">
                {kubeconfigStatus.error}
              </div>
            )}
            {kubeconfigStatus.connection_error && (
              <div className="text-xs text-[#ef4444] mt-1">
                Connection error: {kubeconfigStatus.connection_error}
              </div>
            )}
          </div>
        )}

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-[#a1a1aa] mb-2">
              Kubeconfig Path
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={kubeconfigPath}
                onChange={(e) => setKubeconfigPath(e.target.value)}
                placeholder="~/.kube/config or /path/to/kubeconfig"
                className="flex-1 px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#e4e4e7] focus:border-[#3b82f6] focus:outline-none"
              />
              <button
                onClick={handleAutoDetect}
                className="px-3 py-2 text-xs font-medium text-[#3b82f6] hover:text-[#60a5fa] transition-colors cursor-pointer"
              >
                Auto
              </button>
            </div>
            <p className="text-xs text-[#71717a] mt-1">
              Enter the path to your kubeconfig file. Supports ~ for home directory and environment variables.
            </p>
          </div>

          {error && (
            <div className="p-3 rounded-md bg-[#ef4444]/10 border border-[#ef4444]/20">
              <p className="text-sm text-[#ef4444]">{error}</p>
            </div>
          )}

          {success && (
            <div className="p-3 rounded-md bg-[#10b981]/10 border border-[#10b981]/20">
              <p className="text-sm text-[#10b981]">Kubeconfig configured successfully! Reloading...</p>
            </div>
          )}

          <div className="flex gap-3 flex-wrap">
            <button
              onClick={handleSaveKubeconfig}
              disabled={saving}
              className="px-4 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              {saving ? (
                <>
                  <Icon name="loading" className="animate-spin" size="sm" />
                  <span>Saving...</span>
                </>
              ) : (
                <>
                  <Icon name="check" size="sm" />
                  <span>Save & Connect</span>
                </>
              )}
            </button>
            <button
              onClick={() => {
                setKubeconfigPath('~/.kube/config');
                setError(null);
              }}
              className="px-4 py-2 border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] rounded-md hover:bg-[#111118] transition-colors"
              title="Set to standard location: ~/.kube/config"
            >
              Use Default
            </button>
            <button
              onClick={handleAutoDetect}
              className="px-4 py-2 border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] rounded-md hover:bg-[#111118] transition-colors"
            >
              Auto Detect
            </button>
            <button
              onClick={loadKubeconfigStatus}
              className="px-4 py-2 border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] rounded-md hover:bg-[#111118] transition-colors"
            >
              Refresh Status
            </button>
          </div>

          <div className="p-3 rounded-md bg-[#3b82f6]/10 border border-[#3b82f6]/20">
            <p className="text-xs text-[#3b82f6]">
              💡 Stargazer automatically checks these locations if no path is set:
              <br />• ~/.kube/config (standard location)
              <br />• ~/.kube/kubeconfig, ~/kubeconfig (alternatives)
              <br />• KUBECONFIG environment variable (supports multiple paths)
              <br />• Current directory (kubeconfig, .kube/config)
              <br />• /etc/kubernetes/admin.conf (kubeadm setups)
              {kubeconfigStatus?.auto_found && (
                <span className="block mt-2 text-[#10b981]">
                  ✓ Auto-discovered kubeconfig at: {kubeconfigStatus.path}
                </span>
              )}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

// Provider configuration component
function ProviderConfig({ provider, config, onUpdate }: { provider: any; config: any; onUpdate: () => void }) {
  const providerConfig = config?.providers?.[provider.key] || {};
  const isEnabled = providerConfig.enabled || false;
  const hasApiKey = providerConfig.has_key || false;
  const currentModel = providerConfig.model || provider.models[0];
  const [selectedModel, setSelectedModel] = useState(currentModel || provider.models[0]);
  const [apiKey, setApiKey] = useState('');
  const [showApiKeyInput, setShowApiKeyInput] = useState(!hasApiKey && isEnabled);
  const [savingApiKey, setSavingApiKey] = useState(false);
  const [sopsAvailable] = useState(config?.sops_available || false);

  // Update selectedModel when currentModel changes
  useEffect(() => {
    if (currentModel) {
      setSelectedModel(currentModel);
    } else {
      setSelectedModel(provider.models[0]);
    }
  }, [currentModel, provider.models]);

  // Show API key input when enabling if no key is set
  useEffect(() => {
    if (isEnabled && !hasApiKey) {
      setShowApiKeyInput(true);
    }
  }, [isEnabled, hasApiKey]);

  return (
    <div className="p-4 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)]">
      <div className="flex items-center justify-between mb-3">
        <div>
          <div className="text-sm font-medium text-[#a1a1aa]">{provider.name}</div>
          <div className="text-xs text-[#71717a] mt-1">
            {providerConfig.has_key ? '✅ API key configured' : '⚠️ API key not set'}
          </div>
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={isEnabled}
            onChange={async (e) => {
              const willBeEnabled = e.target.checked;

              try {
                await apiClient.enableProvider(provider.key, willBeEnabled);

                // If enabling, show API key input if no key is set
                if (willBeEnabled && !hasApiKey) {
                  setShowApiKeyInput(true);
                }

                // If disabling, hide API key input
                if (!willBeEnabled) {
                  setShowApiKeyInput(false);
                }

                onUpdate();
              } catch (error) {
                if (process.env.NODE_ENV === 'development') {
                  console.error('Failed to enable provider:', error);
                }
                alert('Failed to update provider');
              }
            }}
            className="w-5 h-5 rounded border-[rgba(255,255,255,0.2)] bg-[#1a1a24] text-[#3b82f6] focus:ring-[#3b82f6]"
          />
          <span className="text-xs text-[#71717a]">Enabled</span>
        </label>
      </div>

      {isEnabled && (
        <div className="mt-3 pt-3 border-t border-[rgba(255,255,255,0.08)] space-y-4">
          {/* API Key Configuration */}
          <div>
            <label className="block text-xs font-medium text-[#71717a] mb-2">
              API Key {!hasApiKey && <span className="text-[#f59e0b]">(Required)</span>}
              {sopsAvailable && hasApiKey && (
                <span className="ml-2 text-[#10b981] text-xs">🔒 Encrypted with SOPS</span>
              )}
            </label>
            {showApiKeyInput ? (
              <div className="space-y-2">
                <input
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  placeholder="Enter API key..."
                  className="w-full px-3 py-2 rounded-md bg-[#0a0a0f] border border-[rgba(255,255,255,0.08)] text-[#e4e4e7] text-sm focus:border-[#3b82f6] focus:outline-none"
                />
                <div className="flex gap-2">
                  <button
                    onClick={async () => {
                      if (!apiKey.trim()) {
                        alert('Please enter an API key');
                        return;
                      }
                      setSavingApiKey(true);
                      try {
                        await apiClient.setProviderApiKey(provider.key, apiKey);
                        setApiKey('');
                        setShowApiKeyInput(false);
                        onUpdate();
                      } catch (error) {
                        if (process.env.NODE_ENV === 'development') {
                          console.error('Failed to save API key:', error);
                        }
                        alert('Failed to save API key');
                      } finally {
                        setSavingApiKey(false);
                      }
                    }}
                    disabled={savingApiKey}
                    className="flex-1 px-3 py-2 bg-[#3b82f6] text-white rounded-md text-xs hover:bg-[#2563eb] transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {savingApiKey ? 'Saving...' : 'Save API Key'}
                  </button>
                  {hasApiKey && (
                    <button
                      onClick={() => {
                        setShowApiKeyInput(false);
                        setApiKey('');
                      }}
                      className="px-3 py-2 border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] rounded-md text-xs hover:bg-[#111118] transition-colors"
                    >
                      Cancel
                    </button>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-between">
                <span className="text-sm text-[#71717a]">
                  {hasApiKey ? '••••••••••••••••' : 'No API key set'}
                </span>
                <button
                  onClick={() => setShowApiKeyInput(true)}
                  className="px-3 py-1.5 text-xs font-medium text-[#3b82f6] hover:text-[#60a5fa] transition-colors"
                >
                  {hasApiKey ? 'Change' : 'Set API Key'}
                </button>
              </div>
            )}
          </div>

          {/* Model Configuration */}
          <div>
            <label className="block text-xs font-medium text-[#71717a] mb-2">
              Model Configuration {!currentModel && <span className="text-[#f59e0b]">(Required)</span>}
            </label>
            <div className="space-y-2">
              <select
                value={selectedModel}
                onChange={(e) => setSelectedModel(e.target.value)}
                className="w-full px-3 py-2 rounded-md bg-[#0a0a0f] border border-[rgba(255,255,255,0.08)] text-[#e4e4e7] text-sm focus:border-[#3b82f6] focus:outline-none"
              >
                {provider.models.map((model: string) => (
                  <option key={model} value={model}>{model}</option>
                ))}
              </select>
              <div className="flex gap-2">
                <button
                  onClick={async () => {
                    if (!selectedModel) {
                      alert('Please select a model');
                      return;
                    }
                    try {
                      await apiClient.setProviderModel(provider.key, selectedModel);
                      onUpdate();
                    } catch (error) {
                      if (process.env.NODE_ENV === 'development') {
                        console.error('Failed to set model:', error);
                      }
                      alert('Failed to update model');
                    }
                  }}
                  className="flex-1 px-3 py-2 bg-[#3b82f6] text-white rounded-md text-xs hover:bg-[#2563eb] transition-colors font-medium"
                >
                  {currentModel === selectedModel ? '✓ Saved' : 'Save Model'}
                </button>
              </div>
              {currentModel && currentModel !== selectedModel && (
                <p className="text-xs text-[#71717a]">
                  Current: <span className="text-[#a1a1aa]">{currentModel}</span>
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

interface SettingsProps {
  onClose?: () => void;
}

type Theme = 'dark' | 'light' | 'auto';

export default function Settings({ onClose }: SettingsProps) {
  const [theme, setTheme] = useState<Theme>('dark');
  const [config, setConfig] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<'general' | 'ai' | 'cluster' | 'advanced'>('general');

  useEffect(() => {
    loadSettings();
    loadTheme();

    // Listen for tab change events from other components (e.g., when clicking "Configure" from dashboard)
    const handleTabChange = (e: CustomEvent) => {
      if (e.detail === 'cluster') {
        setActiveTab('cluster');
      }
    };
    window.addEventListener('settings-tab-change', handleTabChange as EventListener);
    return () => {
      window.removeEventListener('settings-tab-change', handleTabChange as EventListener);
    };
  }, []);

  const loadTheme = () => {
    const savedTheme = localStorage.getItem('stargazer-theme') as Theme || 'dark';
    setTheme(savedTheme);
    applyTheme(savedTheme);
  };

  const applyTheme = (newTheme: Theme) => {
    const root = document.documentElement;

    if (newTheme === 'auto') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      root.classList.toggle('light', !prefersDark);
      root.classList.toggle('dark', prefersDark);
    } else {
      root.classList.remove('light', 'dark');
      root.classList.add(newTheme);
    }

    localStorage.setItem('stargazer-theme', newTheme);
  };

  const handleThemeChange = (newTheme: Theme) => {
    setTheme(newTheme);
    applyTheme(newTheme);

    // Dispatch custom event so main app can update
    window.dispatchEvent(new CustomEvent('theme-change', { detail: newTheme }));
  };

  const loadSettings = async () => {
    try {
      setLoading(true);
      // Load config from API
      const providersConfig = await apiClient.getProvidersConfig();
      setConfig(providersConfig);

      // Also load from local storage for UI preferences
      const savedConfig = localStorage.getItem('stargazer-config');
      if (savedConfig) {
        const localConfig = JSON.parse(savedConfig);
        setConfig({ ...providersConfig, ...localConfig });
      }
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to load settings:', error);
      }
      // Fallback to local storage
      const savedConfig = localStorage.getItem('stargazer-config');
      if (savedConfig) {
        setConfig(JSON.parse(savedConfig));
      }
    } finally {
      setLoading(false);
    }
  };

  const saveSettings = async () => {
    try {
      setSaving(true);
      if (config) {
        localStorage.setItem('stargazer-config', JSON.stringify(config));
      }
      // Show success message
      alert('Settings saved successfully!');
    } catch (error) {
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to save settings:', error);
      }
      alert('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-[#71717a]">Loading settings...</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-[#e4e4e7] mb-2">Settings</h1>
        <p className="text-sm text-[#71717a]">Configure Stargazer preferences and options</p>
      </div>

      {/* Tabs */}
      <div className="mb-6 border-b border-[rgba(255,255,255,0.08)]">
        <div className="flex gap-4">
          {[
            { id: 'general', label: 'General', icon: 'info' as const },
            { id: 'ai', label: 'AI Providers', icon: 'scan' as const },
            { id: 'cluster', label: 'Cluster', icon: 'network' as const },
            { id: 'advanced', label: 'Advanced', icon: 'terminal' as const },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors cursor-pointer ${
                activeTab === tab.id
                  ? 'border-[#3b82f6] text-[#3b82f6]'
                  : 'border-transparent text-[#71717a] hover:text-[#a1a1aa]'
              }`}
            >
              <div className="flex items-center gap-2">
                <Icon name={tab.icon} size="sm" />
                <span>{tab.label}</span>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* General Settings */}
      {activeTab === 'general' && (
        <div className="space-y-6">
          <div className="card p-6">
            <h2 className="text-lg font-semibold text-[#e4e4e7] mb-4">Appearance</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-[#a1a1aa] mb-2">
                  Theme
                </label>
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { value: 'dark', label: 'Dark', icon: '🌙' },
                    { value: 'light', label: 'Light', icon: '☀️' },
                    { value: 'auto', label: 'Auto', icon: '🔄' },
                  ].map((option) => (
                    <button
                      key={option.value}
                      onClick={() => handleThemeChange(option.value as Theme)}
                      className={`px-4 py-3 rounded-md border transition-all ${
                        theme === option.value
                          ? 'border-[#3b82f6] bg-[#3b82f6]/10 text-[#3b82f6]'
                          : 'border-[rgba(255,255,255,0.08)] text-[#a1a1aa] hover:border-[rgba(255,255,255,0.12)]'
                      }`}
                    >
                      <div className="text-2xl mb-1">{option.icon}</div>
                      <div className="text-sm font-medium">{option.label}</div>
                    </button>
                  ))}
                </div>
                <p className="text-xs text-[#71717a] mt-2">
                  Auto mode follows your system preference
                </p>
              </div>
            </div>
          </div>

          <div className="card p-6">
            <h2 className="text-lg font-semibold text-[#e4e4e7] mb-4">Dashboard</h2>

            <div className="space-y-4">
              <label className="flex items-center justify-between">
                <div>
                  <div className="text-sm font-medium text-[#a1a1aa]">Auto-refresh</div>
                  <div className="text-xs text-[#71717a]">Automatically refresh dashboard data</div>
                </div>
                <input
                  type="checkbox"
                  defaultChecked={true}
                  className="w-5 h-5 rounded border-[rgba(255,255,255,0.2)] bg-[#1a1a24] text-[#3b82f6] focus:ring-[#3b82f6]"
                />
              </label>

              <label className="flex items-center justify-between">
                <div>
                  <div className="text-sm font-medium text-[#a1a1aa]">Show notifications</div>
                  <div className="text-xs text-[#71717a]">Display browser notifications for issues</div>
                </div>
                <input
                  type="checkbox"
                  defaultChecked={false}
                  className="w-5 h-5 rounded border-[rgba(255,255,255,0.2)] bg-[#1a1a24] text-[#3b82f6] focus:ring-[#3b82f6]"
                />
              </label>
            </div>
          </div>
        </div>
      )}

      {/* AI Settings */}
      {activeTab === 'ai' && (
        <div className="space-y-6">
          <div className="card p-6">
            <h2 className="text-lg font-semibold text-[#e4e4e7] mb-4">AI Provider Configuration</h2>
            <p className="text-sm text-[#71717a] mb-4">
              Configure AI providers and models for troubleshooting. API keys are set via environment variables.
            </p>

            <div className="space-y-4">
              {[
                { name: 'OpenAI', key: 'openai', models: ['gpt-4o-mini', 'gpt-4o', 'gpt-4-turbo', 'gpt-3.5-turbo'] },
                { name: 'Anthropic', key: 'anthropic', models: ['claude-3-haiku-20240307', 'claude-3-sonnet-20240229', 'claude-3-opus-20240229'] },
                { name: 'Google Gemini', key: 'gemini', models: ['gemini-pro', 'gemini-pro-vision'] },
                { name: 'Cohere', key: 'cohere', models: ['command-r-plus', 'command-r', 'command'] },
                { name: 'Mistral AI', key: 'mistral', models: ['mistral-large-latest', 'mistral-medium-latest', 'mistral-small-latest'] },
                { name: 'Groq', key: 'groq', models: ['llama-3.1-70b-versatile', 'llama-3.1-8b-instant', 'mixtral-8x7b-32768'] },
                { name: 'Together AI', key: 'together', models: ['meta-llama/Llama-3-70b-chat-hf', 'mistralai/Mixtral-8x7B-Instruct-v0.1'] },
              ].map((provider) => (
                <ProviderConfig
                  key={provider.key}
                  provider={provider}
                  config={config}
                  onUpdate={loadSettings}
                />
              ))}
            </div>

            <div className="mt-4 p-3 rounded-md bg-[#3b82f6]/10 border border-[#3b82f6]/20">
              <p className="text-xs text-[#3b82f6]">
                💡 Tip: Set API keys via environment variables (e.g., OPENAI_API_KEY) or run <code className="bg-[#1a1a24] px-1 rounded">stargazer setup</code> in terminal
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Cluster Settings */}
      {activeTab === 'cluster' && (
        <KubeconfigSettings config={config} onUpdate={loadSettings} />
      )}

      {/* Advanced Settings */}
      {activeTab === 'advanced' && (
        <div className="space-y-6">
          <div className="card p-6">
            <h2 className="text-lg font-semibold text-[#e4e4e7] mb-4">Advanced Options</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-[#a1a1aa] mb-2">
                  Cache TTL (seconds)
                </label>
                <input
                  type="number"
                  defaultValue={30}
                  min={5}
                  max={300}
                  className="w-full px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#e4e4e7] focus:border-[#3b82f6] focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-[#a1a1aa] mb-2">
                  API Timeout (seconds)
                </label>
                <input
                  type="number"
                  defaultValue={30}
                  min={5}
                  max={120}
                  className="w-full px-3 py-2 rounded-md bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#e4e4e7] focus:border-[#3b82f6] focus:outline-none"
                />
              </div>

              <div className="p-3 rounded-md bg-[#f59e0b]/10 border border-[#f59e0b]/20">
                <p className="text-xs text-[#f59e0b]">
                  ⚠️ Advanced settings may affect performance. Modify with caution.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Save Button */}
      <div className="mt-6 flex justify-end gap-3">
        {onClose && (
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-md border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] hover:bg-[#111118] transition-colors"
          >
            Cancel
          </button>
        )}
        <button
          onClick={saveSettings}
          disabled={saving}
          className="px-4 py-2 rounded-md bg-[#3b82f6] text-white hover:bg-[#2563eb] disabled:opacity-50 transition-colors flex items-center gap-2"
        >
          {saving ? (
            <>
              <Icon name="loading" className="animate-spin" size="sm" />
              <span>Saving...</span>
            </>
          ) : (
            <>
              <Icon name="check" size="sm" />
              <span>Save Settings</span>
            </>
          )}
        </button>
      </div>
    </div>
  );
}
