/**
 * API client for Stargazer backend
 */
import axios from 'axios';

// In Wails desktop app, API runs on localhost:8000
// In browser dev mode, use empty string for same origin
const API_BASE = typeof window !== 'undefined' 
	? (window.location.hostname === 'localhost' && window.location.port !== '8000' ? 'http://localhost:8000' : '')
	: 'http://localhost:8000';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000, // 30 second timeout
  validateStatus: (status: number) => status < 500, // Don't throw on 4xx
});

export interface ClusterHealth {
  pods: { total: number; healthy: number };
  deployments: { total: number; healthy: number };
  events: { warnings: number; errors: number };
  overall_health: 'healthy' | 'degraded';
}

export interface Issue {
  id: string;
  title: string;
  description: string;
  priority: 'critical' | 'warning' | 'info';
  resource_type: string;
  resource_name: string;
  namespace: string;
  timestamp: string;
}

export interface Pod {
  name: string;
  namespace: string;
  status: string;
  node: string;
  ready: boolean;
  restarts: number;
  age: string;
}

export interface Agent {
  agents: string[];
  current: string;
}

export const apiClient = {
  async getHealth(namespace?: string): Promise<ClusterHealth> {
    try {
      const response = await api.get<ClusterHealth>('/api/cluster/health', {
        params: namespace !== undefined ? { namespace } : {},
      });
      return response.data;
    } catch (error: any) {
      // Return default health on error
      return {
        pods: { total: 0, healthy: 0 },
        deployments: { total: 0, healthy: 0 },
        events: { warnings: 0, errors: 0 },
        overall_health: 'degraded' as const,
      };
    }
  },

  async getIssues(namespace?: string): Promise<Issue[]> {
    try {
      const response = await api.get<{ issues: Issue[]; count: number }>('/api/cluster/issues', {
        params: namespace !== undefined ? { namespace } : {},
      });
      return response.data.issues;
    } catch (error: any) {
      console.error('Error fetching issues:', error);
      return [];
    }
  },

  async getPods(namespace?: string): Promise<Pod[]> {
    const response = await api.get<{ pods: Pod[]; namespace: string; count: number }>('/api/pods', {
      params: { namespace },
    });
    return response.data.pods;
  },

  async getAgents(): Promise<Agent> {
    const response = await api.get<Agent>('/api/agents');
    return response.data;
  },

  async executeCommand(command: string, agent?: string): Promise<string> {
    try {
      const response = await api.post<{ result: string; agent: string }>(
        '/api/agents/execute',
        { command, agent }
      );
      if (!response.data || !('result' in response.data)) {
        throw new Error('Invalid response format: missing result field');
      }
      return response.data.result;
    } catch (error: any) {
      throw error;
    }
  },

  async getNamespace(): Promise<string> {
    try {
      const response = await api.get<{ namespace: string }>('/api/namespace');
      return response.data.namespace;
    } catch (error: any) {
      return 'default';
    }
  },

  async getDeployments(namespace?: string): Promise<any> {
    const response = await api.get<{ deployments: any[]; namespace: string; count: number }>('/api/deployments', {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getEvents(namespace?: string): Promise<any> {
    const response = await api.get<{ events: any[]; namespace: string; count: number }>('/api/events', {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getAllResources(namespace?: string): Promise<any> {
    const response = await api.get<{ resources: any; namespace: string }>('/api/resources', {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getIssueRecommendations(issueId: string): Promise<any> {
    const response = await api.get(`/api/issues/${issueId}/recommendations`);
    return response.data;
  },

  async executeFix(issueId: string, command: string): Promise<any> {
    try {
      const response = await api.post(`/api/issues/${issueId}/execute-fix`, { command });
      return response.data;
    } catch (error: any) {
      if (error.response?.status === 403) {
        throw new Error('Command not approved. Only AI-recommended commands can be executed.');
      }
      throw error;
    }
  },

  async getProgressiveHelp(issueId: string): Promise<any> {
    const response = await api.get(`/api/issues/${issueId}/progressive-help`);
    return response.data;
  },

  async getServiceTopology(namespace?: string): Promise<any> {
    const response = await api.get('/api/topology', {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getServiceConnections(serviceName: string, namespace?: string): Promise<any> {
    const response = await api.get(`/api/topology/${serviceName}`, {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getNetworkPolicyYaml(policyName: string, namespace?: string): Promise<any> {
    const response = await api.get(`/api/networkpolicy/${policyName}`, {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getResourceYaml(resourceType: string, resourceName: string, namespace?: string): Promise<any> {
    const response = await api.get(`/api/resources/${resourceType}/${resourceName}/yaml`, {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getContexts(): Promise<any> {
    const response = await api.get('/api/contexts');
    return response.data;
  },

  async getCurrentContext(): Promise<any> {
    const response = await api.get('/api/context/current');
    return response.data;
  },

  async switchContext(context: string): Promise<any> {
    const response = await api.post('/api/context/switch', { context });
    return response.data;
  },

  async getNamespaces(): Promise<any> {
    const response = await api.get('/api/namespaces');
    return response.data;
  },

  async getProvidersConfig(): Promise<any> {
    const response = await api.get('/api/config/providers');
    return response.data;
  },

  async setProviderModel(provider: string, model: string): Promise<any> {
    const response = await api.post(`/api/config/providers/${provider}/model`, { model });
    return response.data;
  },

  async enableProvider(provider: string, enabled: boolean): Promise<any> {
    const response = await api.post(`/api/config/providers/${provider}/enable`, { enabled });
    return response.data;
  },

  async setProviderApiKey(provider: string, apiKey: string): Promise<any> {
    const response = await api.post(`/api/config/providers/${provider}/api-key`, { api_key: apiKey });
    return response.data;
  },

  async getKubeconfigStatus(): Promise<any> {
    const response = await api.get('/api/config/kubeconfig/status');
    return response.data;
  },

  async setKubeconfig(path: string, context?: string): Promise<any> {
    const response = await api.post('/api/config/kubeconfig', { path, context });
    return response.data;
  },
};

export interface KubernetesContext {
  name: string;
  cluster: string;
  user: string;
  namespace: string;
  server: string;
  cloud_provider: string;
  is_current: boolean;
}

export default apiClient;
