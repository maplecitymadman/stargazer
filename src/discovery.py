import asyncio
import re
from datetime import datetime, timezone
from typing import List, Dict, Any
try:
    from .k8s_client import K8sClient
    from .utils import Issue, Priority, generate_issue_id, get_pod_namespace
except ImportError:
    from k8s_client import K8sClient
    from utils import Issue, Priority, generate_issue_id, get_pod_namespace

class Discovery:
    def __init__(self, k8s_client: K8sClient):
        self.k8s_client = k8s_client
        self.namespace = get_pod_namespace()
    
    async def scan_all(self) -> List[Issue]:
        issues = []
        
        # Parallel scanning for efficiency
        tasks = [
            self._scan_pods(),
            self._scan_deployments(),
            self._scan_events(),
        ]
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        for result in results:
            if isinstance(result, list):
                issues.extend(result)
            else:
                print(f"Discovery error: {result}")
        
        return issues
    
    async def _scan_pods(self) -> List[Issue]:
        issues = []
        pods = await self.k8s_client.get_pods(self.namespace)
        
        for pod in pods:
            # Check pod status
            if pod["status"] not in ["Running", "Succeeded"]:
                priority = Priority.CRITICAL if pod["status"] == "Failed" else Priority.WARNING
                issues.append(Issue(
                    id=generate_issue_id(pod["name"], "status"),
                    title=f"Pod {pod['name']} in {pod['status']} state",
                    description=f"Pod is in {pod['status']} state instead of Running",
                    priority=priority,
                    resource_type="pod",
                    resource_name=pod["name"],
                    namespace=pod["namespace"]
                ))
            
            # Check restart count
            if pod["restarts"] > 5:
                issues.append(Issue(
                    id=generate_issue_id(pod["name"], "restarts"),
                    title=f"High restart count for {pod['name']}",
                    description=f"Pod has restarted {pod['restarts']} times",
                    priority=Priority.WARNING,
                    resource_type="pod",
                    resource_name=pod["name"],
                    namespace=pod["namespace"]
                ))
            
            # Check readiness
            if not pod["ready"] and pod["status"] == "Running":
                issues.append(Issue(
                    id=generate_issue_id(pod["name"], "readiness"),
                    title=f"Pod {pod['name']} not ready",
                    description="Pod is running but not ready",
                    priority=Priority.WARNING,
                    resource_type="pod",
                    resource_name=pod["name"],
                    namespace=pod["namespace"]
                ))
        
        return issues
    
    async def _scan_deployments(self) -> List[Issue]:
        issues = []
        deployments = await self.k8s_client.get_deployments(self.namespace)
        
        for deployment in deployments:
            # Check replica mismatches
            if deployment["replicas"] != deployment["ready"]:
                issues.append(Issue(
                    id=generate_issue_id(deployment["name"], "replicas"),
                    title=f"Replica mismatch in {deployment['name']}",
                    description=f"Expected {deployment['replicas']} replicas, {deployment['ready']} ready",
                    priority=Priority.WARNING,
                    resource_type="deployment",
                    resource_name=deployment["name"],
                    namespace=deployment["namespace"]
                ))
            
            # Check availability
            if deployment["available"] < deployment["replicas"]:
                issues.append(Issue(
                    id=generate_issue_id(deployment["name"], "availability"),
                    title=f"Deployment {deployment['name']} unavailable",
                    description=f"Only {deployment['available']} of {deployment['replicas']} replicas available",
                    priority=Priority.CRITICAL,
                    resource_type="deployment",
                    resource_name=deployment["name"],
                    namespace=deployment["namespace"]
                ))
        
        return issues
    
    async def _scan_events(self) -> List[Issue]:
        issues = []
        events = await self.k8s_client.get_events(self.namespace)
        
        # Group events by object and type
        event_groups = {}
        for event in events:
            key = f"{event['object_kind']}-{event['object_name']}-{event['reason']}"
            if key not in event_groups:
                event_groups[key] = []
            event_groups[key].append(event)
        
        for key, group in event_groups.items():
            # Focus on recent warning events
            recent_events = [e for e in group if self._is_recent(e["timestamp"])]
            
            if recent_events:
                latest_event = recent_events[0]  # Events are sorted by timestamp
                
                # Priority based on event type
                if "Error" in latest_event["reason"] or "Failed" in latest_event["reason"]:
                    priority = Priority.CRITICAL
                elif latest_event["type"] == "Warning":
                    priority = Priority.WARNING
                else:
                    priority = Priority.INFO
                
                # Clean up message
                message = latest_event["message"]
                if len(message) > 200:
                    message = message[:197] + "..."
                
                issues.append(Issue(
                    id=generate_issue_id(latest_event["object_name"], "event"),
                    title=f"{latest_event['reason']} on {latest_event['object_kind']}/{latest_event['object_name']}",
                    description=message,
                    priority=priority,
                    resource_type=latest_event["object_kind"].lower(),
                    resource_name=latest_event["object_name"],
                    namespace=latest_event["namespace"]
                ))
        
        return issues
    
    def _is_recent(self, timestamp) -> bool:
        if not timestamp:
            return False
        
        age = datetime.now(timezone.utc) - timestamp
        return age.total_seconds() < 3600  # Last hour
    
    async def get_resource_health(self) -> Dict[str, Any]:
        """Get overall cluster health summary"""
        pods = await self.k8s_client.get_pods(self.namespace)
        deployments = await self.k8s_client.get_deployments(self.namespace)
        events = await self.k8s_client.get_events(self.namespace)
        
        total_pods = len(pods)
        healthy_pods = len([p for p in pods if p["status"] == "Running" and p["ready"]])
        
        total_deployments = len(deployments)
        healthy_deployments = len([d for d in deployments if d["ready"] == d["replicas"]])
        
        warning_events = len([e for e in events if e["type"] == "Warning"])
        error_events = len([e for e in events if "Error" in e["reason"] or "Fail" in e["reason"]])
        
        return {
            "pods": {"total": total_pods, "healthy": healthy_pods},
            "deployments": {"total": total_deployments, "healthy": healthy_deployments},
            "events": {"warnings": warning_events, "errors": error_events},
            "overall_health": "healthy" if healthy_pods == total_pods and healthy_deployments == total_deployments else "degraded"
        }