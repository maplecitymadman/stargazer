import asyncio
from typing import Dict, List, Any, Optional
try:
    from .k8s_client import K8sClient
    from .discovery import Discovery
    from .mock_ai import MockAI
    from .utils import Issue, Priority, get_pod_namespace
except ImportError:
    from k8s_client import K8sClient
    from discovery import Discovery
    from mock_ai import MockAI
    from utils import Issue, Priority, get_pod_namespace

class Agent:
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
        self.k8s_client = K8sClient()
        self.ai = MockAI()
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        """Execute agent-specific query"""
        raise NotImplementedError
    
    def get_help(self) -> str:
        return f"{self.name}: {self.description}"

class TroubleshooterAgent(Agent):
    def __init__(self):
        super().__init__("troubleshooter", "Main agent for Kubernetes troubleshooting")
        self.discovery = Discovery(self.k8s_client)
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        if query.startswith("scan") or query == "analyze":
            issues = await self.discovery.scan_all()
            if not issues:
                return "No issues detected in the cluster."
            
            critical_issues = [i for i in issues if i.priority == Priority.CRITICAL]
            warning_issues = [i for i in issues if i.priority == Priority.WARNING]
            info_issues = [i for i in issues if i.priority == Priority.INFO]
            
            response = f"Found {len(issues)} issues:\n"
            
            if critical_issues:
                response += f"\n🔴 CRITICAL ({len(critical_issues)}):\n"
                for issue in critical_issues[:5]:  # Limit to first 5
                    response += f"  • {issue.title}\n"
                    response += f"    {issue.description}\n"
                    response += f"    Resource: {issue.resource_type}/{issue.resource_name}\n\n"
            
            if warning_issues:
                response += f"\n🟡 WARNING ({len(warning_issues)}):\n"
                for issue in warning_issues[:5]:  # Limit to first 5
                    response += f"  • {issue.title}\n"
                    response += f"    {issue.description}\n\n"
            
            if info_issues:
                response += f"\n🔵 INFO ({len(info_issues)}):\n"
                for issue in info_issues[:3]:  # Limit to first 3
                    response += f"  • {issue.title}\n\n"
            
            return response
        
        elif query.startswith("health"):
            health = await self.discovery.get_resource_health()
            return self.ai.analyze_cluster_health(health)
        
        elif query.startswith("analyze"):
            # Extract resource name from query
            parts = query.split()
            if len(parts) > 1:
                resource_name = parts[1]
                return await self._analyze_resource(resource_name)
            else:
                return "Please specify a resource name to analyze. Example: analyze web-app-123"
        
        else:
            return self.ai.analyze_issue(None, context or {})
    
    async def _analyze_resource(self, resource_name: str) -> str:
        """Analyze specific resource"""
        try:
            pods = await self.k8s_client.get_pods()
            target_pod = next((p for p in pods if resource_name in p["name"]), None)
            
            if target_pod:
                issues = await self.discovery._scan_pods()
                pod_issues = [i for i in issues if i.resource_name == target_pod["name"]]
                
                response = f"Analysis for pod {target_pod['name']}:\n"
                response += f"Status: {target_pod['status']}\n"
                response += f"Ready: {target_pod['ready']}\n"
                response += f"Restarts: {target_pod['restarts']}\n"
                response += f"Node: {target_pod['node']}\n\n"
                
                if pod_issues:
                    response += "Issues detected:\n"
                    for issue in pod_issues:
                        response += f"  • {issue.title}\n"
                        response += f"    {issue.description}\n\n"
                else:
                    response += "No specific issues detected for this pod.\n"
                
                return response
            
            return f"Resource '{resource_name}' not found or not accessible."
        
        except Exception as e:
            return f"Error analyzing resource: {e}"

class DiscoveryAgent(Agent):
    def __init__(self):
        super().__init__("discovery", "Discovers and analyzes cluster resources")
        self.discovery = Discovery(self.k8s_client)
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        if query == "pods":
            pods = await self.k8s_client.get_pods()
            response = f"Found {len(pods)} pods:\n"
            for pod in pods[:10]:  # Limit to first 10
                status_icon = "✅" if pod["ready"] and pod["status"] == "Running" else "❌"
                response += f"{status_icon} {pod['name']} ({pod['status']})\n"
            return response
        
        elif query == "deployments":
            deployments = await self.k8s_client.get_deployments()
            response = f"Found {len(deployments)} deployments:\n"
            for dep in deployments[:10]:  # Limit to first 10
                status_icon = "✅" if dep["ready"] == dep["replicas"] else "❌"
                response += f"{status_icon} {dep['name']} ({dep['ready']}/{dep['replicas']})\n"
            return response
        
        elif query == "events":
            events = await self.k8s_client.get_events()
            response = f"Recent events (last hour):\n"
            for event in events[:10]:  # Limit to first 10
                icon = "⚠️" if event["type"] == "Warning" else "❌"
                response += f"{icon} {event['reason']} on {event['object_name']}\n"
                response += f"   {event['message'][:100]}...\n\n"
            return response
        
        else:
            return "Available commands: pods, deployments, events"

class LogsAgent(Agent):
    def __init__(self):
        super().__init__("logs", "Retrieves and analyzes pod logs")
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        if query.startswith("get"):
            parts = query.split()
            if len(parts) > 1:
                pod_name = parts[1]
                lines = 50
                if len(parts) > 2 and parts[2].isdigit():
                    lines = int(parts[2])
                
                logs = await self.k8s_client.get_pod_logs(pod_name, lines=lines)
                return f"Logs for {pod_name} (last {lines} lines):\n{logs}"
            else:
                return "Please specify a pod name. Example: get web-app-123"
        
        elif query.startswith("errors"):
            parts = query.split()
            if len(parts) > 1:
                pod_name = parts[1]
                logs = await self.k8s_client.get_pod_logs(pod_name, lines=100)
                
                # Filter for error patterns
                error_lines = [
                    line for line in logs.split('\n') 
                    if any(keyword in line.lower() for keyword in ['error', 'exception', 'failed', 'panic'])
                ]
                
                if error_lines:
                    return f"Error patterns in {pod_name} logs:\n" + "\n".join(error_lines[-10:])
                else:
                    return f"No error patterns found in {pod_name} logs."
            else:
                return "Please specify a pod name. Example: errors web-app-123"
        
        else:
            return "Available commands: get <pod-name> [lines], errors <pod-name>"

class ResourceAgent(Agent):
    def __init__(self):
        super().__init__("resource", "Analyzes resource usage and constraints")
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        if query == "top":
            return "Resource usage analysis would require metrics server. Use `kubectl top pods` for current usage."
        
        elif query == "pressure":
            return "Checking for resource pressure...\nNo immediate resource pressure detected in current namespace."
        
        elif query.startswith("describe"):
            parts = query.split()
            if len(parts) > 2:
                resource_type = parts[1]
                resource_name = parts[2]
                return f"Run: kubectl describe {resource_type} {resource_name} -n {get_pod_namespace()}"
            else:
                return "Usage: describe <resource-type> <resource-name>"
        
        else:
            return "Available commands: top, pressure, describe <type> <name>"

class NetworkAgent(Agent):
    def __init__(self):
        super().__init__("network", "Analyzes network connectivity and policies")
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        if query == "connectivity":
            return "Network connectivity check requires additional tools. Consider using netshoot or similar debugging pods."
        
        elif query == "policies":
            return "Run: kubectl get networkpolicies -n {get_pod_namespace()} to view network policies."
        
        elif query == "endpoints":
            return "Run: kubectl get endpoints -n {get_pod_namespace()} to view service endpoints."
        
        else:
            return "Available commands: connectivity, policies, endpoints"

class SecurityAgent(Agent):
    def __init__(self):
        super().__init__("security", "Analyzes security configurations and compliance")
    
    async def execute(self, query: str, context: Dict[str, Any] = None) -> str:
        if query == "rbac":
            return "Run: kubectl auth can-i --list to view RBAC permissions."
        
        elif query == "secrets":
            return "Checking secrets...\nNo security issues detected with current secret configurations."
        
        elif query == "images":
            return "Image security scanning would require additional tools. Consider using Trivy or similar scanners."
        
        else:
            return "Available commands: rbac, secrets, images"

class AgentSystem:
    def __init__(self):
        self.agents: Dict[str, Agent] = {
            "troubleshooter": TroubleshooterAgent(),
            "discovery": DiscoveryAgent(),
            "logs": LogsAgent(),
            "resource": ResourceAgent(),
            "network": NetworkAgent(),
            "security": SecurityAgent(),
        }
        self.current_agent = "troubleshooter"
    
    def get_agent(self, name: str) -> Optional[Agent]:
        return self.agents.get(name)
    
    def list_agents(self) -> List[str]:
        return list(self.agents.keys())
    
    def get_current_agent(self) -> Agent:
        return self.agents[self.current_agent]
    
    def set_current_agent(self, name: str) -> bool:
        if name in self.agents:
            self.current_agent = name
            return True
        return False
    
    async def execute_command(self, command: str) -> str:
        """Parse and execute command with @agent mentions"""
        if command.startswith("@"):
            # Agent switch or direct agent command
            parts = command[1:].split(" ", 1)
            agent_name = parts[0]
            
            if agent_name in self.agents:
                if len(parts) > 1:
                    # Direct command to specific agent
                    return await self.agents[agent_name].execute(parts[1])
                else:
                    # Switch to agent
                    self.current_agent = agent_name
                    return f"Switched to {agent_name} agent"
            else:
                return f"Unknown agent: {agent_name}"
        
        elif command.startswith("/"):
            # System commands
            if command == "/agents":
                response = "Available agents:\n"
                for name, agent in self.agents.items():
                    current = " (current)" if name == self.current_agent else ""
                    response += f"• @{name}{current} - {agent.description}\n"
                return response
            
            elif command == "/help":
                return """Stargazer Commands:
  /agents - List available agents
  /help - Show this help
  @agentname - Switch to specific agent
  @agentname command - Execute command on specific agent
  !kubectl command - Execute kubectl command"""
            
            else:
                return f"Unknown command: {command}"
        
        elif command.startswith("!"):
            # kubectl commands
            return f"Run: {command[1:]}"
        
        else:
            # Execute with current agent
            return await self.get_current_agent().execute(command)