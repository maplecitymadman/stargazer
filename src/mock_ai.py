import re
from typing import Dict, List, Any
try:
    from .utils import Issue, Priority
except ImportError:
    from utils import Issue, Priority

class MockAI:
    def __init__(self):
        self.patterns = {
            "crash": [
                "Container appears to be crashing. Check resource limits and application logs.",
                "High restart count detected. Consider increasing memory/CPU limits or fixing application errors.",
                "Pod is in CrashLoopBackOff. Review container configuration and application startup."
            ],
            "pending": [
                "Pod is pending. Possible causes: insufficient resources, taints/tolerations, or image pull issues.",
                "Pod scheduling failure. Check node resources and pod requirements.",
                "Pending state detected. Verify resource requests and node availability."
            ],
            "image": [
                "Image pull issue detected. Check image name, tag, and registry access.",
                "Container image problem. Verify image exists and registry is accessible.",
                "Image pull failure. Check credentials and image repository."
            ],
            "network": [
                "Network connectivity issue detected. Check service configuration and network policies.",
                "Communication problem between services. Review endpoints and service discovery.",
                "Network policy might be blocking traffic. Verify ingress/egress rules."
            ],
            "resource": [
                "Resource exhaustion detected. Check CPU/memory usage and limits.",
                "Insufficient resources. Consider scaling up or optimizing resource usage.",
                "Resource pressure on node. Monitor node utilization and pod distribution."
            ],
            "general": [
                "Issue detected in the cluster. Run `kubectl describe` for more details.",
                "Kubernetes resource needs attention. Check logs and events for context.",
                "Cluster resource state requires investigation. Use OpenCode commands for deeper analysis."
            ]
        }
    
    def analyze_issue(self, issue: Issue, context: Dict[str, Any] = None) -> str:
        """Generate AI-like response for an issue"""
        issue_text = f"{issue.title} {issue.description}".lower()
        
        # Determine which pattern to use
        response_pattern = None
        
        for keyword, patterns in self.patterns.items():
            if keyword != "general" and keyword in issue_text:
                response_pattern = patterns
                break
        
        if not response_pattern:
            response_pattern = self.patterns["general"]
        
        # Select response based on priority
        if issue.priority == Priority.CRITICAL:
            return response_pattern[0]  # Most urgent response
        elif issue.priority == Priority.WARNING:
            return response_pattern[1]  # Medium urgency
        else:
            return response_pattern[2]  # General advice
    
    def suggest_commands(self, issue: Issue) -> List[str]:
        """Suggest relevant kubectl commands"""
        commands = []
        
        if issue.resource_type == "pod":
            commands = [
                f"kubectl describe pod {issue.resource_name} -n {issue.namespace}",
                f"kubectl logs {issue.resource_name} -n {issue.namespace}",
                f"kubectl get events -n {issue.namespace} --field-selector involvedObject.name={issue.resource_name}"
            ]
        elif issue.resource_type == "deployment":
            commands = [
                f"kubectl describe deployment {issue.resource_name} -n {issue.namespace}",
                f"kubectl rollout status deployment/{issue.resource_name} -n {issue.namespace}",
                f"kubectl get pods -l app={issue.resource_name} -n {issue.namespace}"
            ]
        
        # Add general commands
        commands.extend([
            f"kubectl get events -n {issue.namespace} --sort-by='.lastTimestamp'",
            f"kubectl top pods -n {issue.namespace}"
        ])
        
        return commands
    
    def analyze_cluster_health(self, health_summary: Dict[str, Any]) -> str:
        """Generate cluster health analysis"""
        overall = health_summary.get("overall_health", "unknown")
        pods = health_summary.get("pods", {})
        deployments = health_summary.get("deployments", {})
        events = health_summary.get("events", {})
        
        if overall == "healthy":
            return "Cluster appears healthy. All pods and deployments are running as expected."
        
        issues = []
        
        if pods.get("healthy", 0) < pods.get("total", 0):
            issues.append(f"{pods['total'] - pods['healthy']} pods are unhealthy")
        
        if deployments.get("healthy", 0) < deployments.get("total", 0):
            issues.append(f"{deployments['total'] - deployments['healthy']} deployments have issues")
        
        if events.get("errors", 0) > 0:
            issues.append(f"{events['errors']} error events detected")
        
        if events.get("warnings", 0) > 5:
            issues.append(f"{events['warnings']} warning events (high)")
        
        if issues:
            return f"Cluster health degraded: {', '.join(issues)}. Investigate critical issues first."
        else:
            return "Cluster has some concerns but no immediate critical issues."