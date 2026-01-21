/**
 * API client for Stargazer backend
 */
import axios from "axios";

// In Wails desktop app, API runs on localhost:8000
// In browser dev mode, use empty string for same origin
const API_BASE =
  typeof window !== "undefined"
    ? window.location.hostname === "localhost" &&
      window.location.port !== "8000"
      ? "http://localhost:8000"
      : ""
    : "http://localhost:8000";

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 30000, // 30 second timeout
  validateStatus: (status: number) => status < 500, // Don't throw on 4xx
});

export interface ClusterHealth {
  pods: { total: number; healthy: number };
  deployments: { total: number; healthy: number };
  events: { warnings: number; errors: number };
  overall_health: "healthy" | "degraded";
}

export interface Issue {
  id: string;
  title: string;
  description: string;
  priority: "critical" | "high" | "warning" | "info";
  resource_type: string;
  resource_name: string;
  namespace: string;
  timestamp?: string;
}

export interface AnalysisResult {
  resource: string;
  type: string;
  namespace: string;
  status: string;
  issues: Issue[];
  recommendations: string[];
  events?: any[];
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

export interface IngressInfo {
  gateways: GatewayInfo[];
  kubernetes_ingress: KubernetesIngressInfo[];
  routes: IngressRoute[];
  connections: IngressConnection[];
}

export interface EgressInfo {
  gateways: GatewayInfo[];
  external_services: ExternalServiceInfo[];
  connections: EgressConnection[];
  has_egress_gateway: boolean;
  direct_egress: boolean;
}

export interface GatewayInfo {
  name: string;
  namespace: string;
  type: "istio" | "nginx" | "kubernetes";
  hosts: string[];
  ports: string[];
  selector?: Record<string, string>;
  service?: string;
}

export interface KubernetesIngressInfo {
  name: string;
  namespace: string;
  hosts: string[];
  paths: string[];
  backend: string;
  backend_port?: string;
  tls: boolean;
  class?: string;
}

export interface IngressRoute {
  gateway: string;
  host: string;
  path: string;
  service: string;
  namespace: string;
  allowed: boolean;
  blocked_by?: string[];
  type: "istio" | "nginx" | "kubernetes";
}

export interface IngressConnection {
  from: string;
  to: string;
  allowed: boolean;
  reason: string;
  policies?: string[];
  port?: string;
  protocol?: string;
}

export interface EgressConnection {
  from: string;
  to: string;
  allowed: boolean;
  reason: string;
  policies?: string[];
  port?: string;
  protocol?: string;
}

export interface ExternalServiceInfo {
  name: string;
  namespace: string;
  hosts: string[];
  ports: string[];
  type: "serviceentry" | "direct";
}

export interface PathTrace {
  source: string;
  destination: string;
  path: PathHop[];
  allowed: boolean;
  blocked_at?: PathHop;
  reason: string;
}

export interface PathHop {
  from: string;
  to: string;
  type: "ingress" | "service" | "egress";
  allowed: boolean;
  reason: string;
  policies?: string[];
  service_mesh?: string;
}

export interface SearchResult {
  type: string;
  name: string;
  namespace: string;
  status: string;
}

export const apiClient = {
  // ... existing methods
  async searchResources(query: string): Promise<SearchResult[]> {
    const response = await api.get<SearchResult[]>("/api/search", {
      params: { query },
    });
    return response.data;
  },
  // ... rest of apiClient
  async getHealth(namespace?: string): Promise<ClusterHealth> {
    try {
      const response = await api.get<ClusterHealth>("/api/cluster/health", {
        params: namespace !== undefined ? { namespace } : {},
      });
      return response.data;
    } catch (error: any) {
      // Return default health on error
      return {
        pods: { total: 0, healthy: 0 },
        deployments: { total: 0, healthy: 0 },
        events: { warnings: 0, errors: 0 },
        overall_health: "degraded" as const,
      };
    }
  },

  async getIssues(namespace?: string): Promise<Issue[]> {
    // Stub - endpoint removed
    return [];
  },

  async getPods(namespace?: string): Promise<Pod[]> {
    // Stub - endpoint removed
    return [];
  },

  async getAgents(): Promise<Agent> {
    const response = await api.get<Agent>("/api/agents");
    return response.data;
  },

  async executeCommand(command: string, agent?: string): Promise<string> {
    try {
      const response = await api.post<{ result: string; agent: string }>(
        "/api/agents/execute",
        { command, agent },
      );
      if (!response.data || !("result" in response.data)) {
        throw new Error("Invalid response format: missing result field");
      }
      return response.data.result;
    } catch (error: any) {
      throw error;
    }
  },

  async getNamespace(): Promise<string> {
    try {
      const response = await api.get<{ namespace: string }>("/api/namespace");
      return response.data.namespace;
    } catch (error: any) {
      return "default";
    }
  },

  async getDeployments(namespace?: string): Promise<any> {
    // Stub - endpoint removed
    return { deployments: [], namespace: namespace || "all", count: 0 };
  },

  async getEvents(namespace?: string): Promise<any> {
    const response = await api.get<{
      events: any[];
      namespace: string;
      count: number;
    }>("/api/events", {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },


  async getServiceTopology(namespace?: string): Promise<any> {
    const response = await api.get("/api/topology", {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async tracePath(
    source: string,
    destination: string,
    namespace?: string,
  ): Promise<PathTrace> {
    const response = await api.get("/api/topology/trace", {
      params: {
        source,
        destination,
        namespace: namespace || "all",
      },
    });
    return response.data;
  },

  async getServiceConnections(
    serviceName: string,
    namespace?: string,
  ): Promise<any> {
    const response = await api.get(`/api/topology/${serviceName}`, {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getNetworkPolicyYaml(
    policyName: string,
    namespace?: string,
  ): Promise<any> {
    const response = await api.get(`/api/networkpolicy/${policyName}`, {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },


  async getContexts(): Promise<any> {
    const response = await api.get("/api/contexts");
    return response.data;
  },

  async getCurrentContext(): Promise<any> {
    const response = await api.get("/api/context/current");
    return response.data;
  },

  async switchContext(context: string): Promise<any> {
    const response = await api.post("/api/context/switch", { context });
    return response.data;
  },

  async getNamespaces(): Promise<any> {
    const response = await api.get("/api/namespaces");
    return response.data;
  },

  async getServices(namespace?: string): Promise<any> {
    const response = await api.get("/api/services", {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getRecommendations(namespace?: string): Promise<any> {
    const response = await api.get("/api/recommendations", {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getComplianceScore(namespace?: string): Promise<any> {
    const response = await api.get("/api/recommendations/score", {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async getProvidersConfig(): Promise<any> {
    const response = await api.get("/api/config/providers");
    return response.data;
  },

  async setProviderModel(provider: string, model: string): Promise<any> {
    const response = await api.post(`/api/config/providers/${provider}/model`, {
      model,
    });
    return response.data;
  },

  async enableProvider(provider: string, enabled: boolean): Promise<any> {
    const response = await api.post(
      `/api/config/providers/${provider}/enable`,
      { enabled },
    );
    return response.data;
  },

  async setProviderApiKey(provider: string, apiKey: string): Promise<any> {
    const response = await api.post(
      `/api/config/providers/${provider}/api-key`,
      { api_key: apiKey },
    );
    return response.data;
  },

  async getKubeconfigStatus(): Promise<any> {
    const response = await api.get("/api/config/kubeconfig/status");
    return response.data;
  },

  async setKubeconfig(path: string, context?: string): Promise<any> {
    const response = await api.post("/api/config/kubeconfig", {
      path,
      context,
    });
    return response.data;
  },

  // Policy Building and Testing
  async buildCiliumPolicy(spec: any): Promise<any> {
    const response = await api.post("/api/policies/cilium/build", spec);
    return response.data;
  },

  async applyCiliumPolicy(yaml: string, namespace?: string): Promise<any> {
    const response = await api.post("/api/policies/cilium/apply", {
      yaml,
      namespace,
    });
    return response.data;
  },

  async applyNetworkPolicy(yaml: string, namespace?: string): Promise<any> {
    const response = await api.post("/api/policies/network/apply", {
      yaml,
      namespace,
    });
    return response.data;
  },

  async exportCiliumPolicy(yaml: string): Promise<Blob> {
    const response = await api.post(
      "/api/policies/cilium/export",
      { yaml },
      {
        responseType: "blob",
      },
    );
    return response.data;
  },

  async deleteCiliumPolicy(name: string, namespace?: string): Promise<any> {
    const response = await api.delete(`/api/policies/cilium/${name}`, {
      params: namespace ? { namespace } : {},
    });
    return response.data;
  },

  async buildKyvernoPolicy(spec: any): Promise<any> {
    const response = await api.post("/api/policies/kyverno/build", spec);
    return response.data;
  },

  async applyKyvernoPolicy(yaml: string, namespace?: string): Promise<any> {
    const response = await api.post("/api/policies/kyverno/apply", {
      yaml,
      namespace,
    });
    return response.data;
  },

  async exportKyvernoPolicy(yaml: string, type?: string): Promise<Blob> {
    const response = await api.post(
      "/api/policies/kyverno/export",
      { yaml, type },
      {
        responseType: "blob",
      },
    );
    return response.data;
  },

  async deleteKyvernoPolicy(
    name: string,
    namespace?: string,
    isClusterPolicy?: boolean,
  ): Promise<any> {
    const response = await api.delete(`/api/policies/kyverno/${name}`, {
      params: {
        ...(namespace ? { namespace } : {}),
        ...(isClusterPolicy ? { cluster_policy: "true" } : {}),
      },
    });
    return response.data;
  },

  async troubleshoot(
    type: string,
    name: string,
    namespace?: string,
  ): Promise<AnalysisResult> {
    const response = await api.get<AnalysisResult>("/api/troubleshoot", {
      params: { type, name, namespace },
    });
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
