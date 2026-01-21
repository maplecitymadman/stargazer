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

    async def scan_all(self, namespace: str = None) -> List[Issue]:
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace

        # Parallel scanning for efficiency - including advanced infrastructure
        tasks = [
            self._scan_pods(ns),
            self._scan_deployments(ns),
            self._scan_events(ns),
            self._scan_service_mesh_issues(ns),
            self._scan_network_policies(ns),
            self._scan_policy_violations(ns),
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        for result in results:
            if isinstance(result, list):
                issues.extend(result)
            else:
                print(f"Discovery error: {result}")

        return issues

    async def _scan_pods(self, namespace: str = None) -> List[Issue]:
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace or self.namespace

        try:
            pods = await self.k8s_client.get_pods(ns)
        except Exception as e:
            print(f"Error scanning pods: {e}")
            return issues

        for pod in pods:
            # Check pod status
            if pod["status"] not in ["Running", "Succeeded"]:
                priority = Priority.CRITICAL if pod["status"] in ["Failed", "CrashLoopBackOff", "Error"] else Priority.WARNING
                status_hint = ""
                if pod["status"] in ["CrashLoopBackOff", "Error"]:
                    status_hint = f" Container is crashing. Check logs with: kubectl logs {pod['name']} -n {pod['namespace']} --previous"
                elif pod["status"] == "Pending":
                    status_hint = f" Pod cannot be scheduled. Check events with: kubectl get events -n {pod['namespace']} --field-selector involvedObject.name={pod['name']}"
                elif pod["status"] == "Failed":
                    status_hint = f" Pod has failed. Check describe output: kubectl describe pod {pod['name']} -n {pod['namespace']}"

                issues.append(Issue(
                    id=generate_issue_id(pod["name"], "status"),
                    title=f"Pod {pod['name']} in {pod['status']} state",
                    description=f"Pod is in {pod['status']} state instead of Running.{status_hint}",
                    priority=priority,
                    resource_type="pod",
                    resource_name=pod["name"],
                    namespace=pod["namespace"]
                ))

            # Check restart count
            if pod["restarts"] > 5:
                pod_name = pod["name"]
                pod_ns = pod["namespace"]
                issues.append(Issue(
                    id=generate_issue_id(pod_name, "restarts"),
                    title=f"High restart count for {pod_name}",
                    description=f"Pod has restarted {pod['restarts']} times. Check crash logs: kubectl logs {pod_name} -n {pod_ns} --previous. Review events: kubectl get events -n {pod_ns} --field-selector involvedObject.name={pod_name}",
                    priority=Priority.WARNING,
                    resource_type="pod",
                    resource_name=pod_name,
                    namespace=pod_ns
                ))

            # Check readiness
            if not pod["ready"] and pod["status"] == "Running":
                pod_name = pod["name"]
                pod_ns = pod["namespace"]
                issues.append(Issue(
                    id=generate_issue_id(pod_name, "readiness"),
                    title=f"Pod {pod_name} not ready",
                    description=f"Pod is running but readiness probe is failing. Check logs and probe configuration. Run: kubectl describe pod {pod_name} -n {pod_ns} to see probe details.",
                    priority=Priority.WARNING,
                    resource_type="pod",
                    resource_name=pod_name,
                    namespace=pod_ns
                ))

        return issues

    async def _scan_deployments(self, namespace: str = None) -> List[Issue]:
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace or self.namespace

        try:
            deployments = await self.k8s_client.get_deployments(ns)
        except Exception as e:
            print(f"Error scanning deployments: {e}")
            return issues

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

    async def _scan_events(self, namespace: str = None) -> List[Issue]:
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace or self.namespace

        try:
            events = await self.k8s_client.get_events(ns)
        except Exception as e:
            print(f"Error scanning events: {e}")
            return issues

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

    async def get_resource_health(self, namespace: str = None) -> Dict[str, Any]:
        """Get overall cluster health summary"""
        # Support "all" or empty string for all namespaces
        if namespace == "all" or namespace == "":
            ns = None
        elif namespace is None:
            # If no namespace specified, use all namespaces by default
            ns = None
        else:
            ns = namespace

        try:
            pods = await self.k8s_client.get_pods(ns)
        except Exception as e:
            print(f"Error getting pods for health check: {e}")
            pods = []

        try:
            deployments = await self.k8s_client.get_deployments(ns)
        except Exception as e:
            print(f"Error getting deployments for health check: {e}")
            deployments = []

        try:
            events = await self.k8s_client.get_events(ns)
        except Exception as e:
            print(f"Error getting events for health check: {e}")
            events = []

        total_pods = len(pods)
        healthy_pods = len([p for p in pods if p.get("status") == "Running" and p.get("ready")])

        total_deployments = len(deployments)
        healthy_deployments = len([d for d in deployments if d.get("ready") == d.get("replicas")])

        warning_events = len([e for e in events if e.get("type") == "Warning"])
        error_events = len([e for e in events if "Error" in str(e.get("reason", "")) or "Fail" in str(e.get("reason", ""))])

        return {
            "pods": {"total": total_pods, "healthy": healthy_pods},
            "deployments": {"total": total_deployments, "healthy": healthy_deployments},
            "events": {"warnings": warning_events, "errors": error_events},
            "overall_health": "healthy" if healthy_pods == total_pods and healthy_deployments == total_deployments and total_pods > 0 else "degraded"
        }

    async def _scan_service_mesh_issues(self, namespace: str = None) -> List[Issue]:
        """Detect service mesh (Istio) related issues"""
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace or self.namespace

        try:
            pods = await self.k8s_client.get_pods(ns)

            for pod in pods:
                labels = pod.get("labels", {})
                pod_name = pod["name"]
                pod_ns = pod["namespace"]

                # Check for Istio sidecar issues
                if "istio-injection" in str(labels) or any("istio" in k.lower() for k in labels.keys()):
                    # Check if pod is running but sidecar might be failing
                    if pod["status"] == "Running" and not pod.get("ready"):
                        # Could be sidecar not ready
                        issues.append(Issue(
                            id=generate_issue_id(pod_name, "istio-sidecar"),
                            title=f"Istio sidecar issue in {pod_name}",
                            description=f"Pod appears to have Istio injection but may have sidecar issues. Check: kubectl logs {pod_name} -n {pod_ns} -c istio-proxy. Verify sidecar: kubectl get pod {pod_name} -n {pod_ns} -o jsonpath='{{.spec.containers[*].name}}'",
                            priority=Priority.WARNING,
                            resource_type="pod",
                            resource_name=pod_name,
                            namespace=pod_ns
                        ))
        except Exception as e:
            # Silently fail for service mesh scanning - not all clusters have it
            pass

        return issues

    async def _scan_network_policies(self, namespace: str = None) -> List[Issue]:
        """Detect network policy (Cilium/NetworkPolicy) related issues"""
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace or self.namespace

        try:
            # Check for pods that might be blocked by network policies
            pods = await self.k8s_client.get_pods(ns)
            events = await self.k8s_client.get_events(ns)

            # Look for network-related errors in events
            network_errors = [e for e in events if any(keyword in e.get("message", "").lower()
                for keyword in ["network", "policy", "forbidden", "denied", "cilium"])]

            for event in network_errors[-5:]:  # Last 5 network errors
                if "network" in event.get("message", "").lower() or "policy" in event.get("message", "").lower():
                    issues.append(Issue(
                        id=generate_issue_id(event["object_name"], "network-policy"),
                        title=f"Network policy issue: {event['reason']}",
                        description=f"{event['message']}. Check network policies: kubectl get networkpolicies -n {event['namespace']}. For Cilium: kubectl get cnp,ccnp -A. Debug: kubectl exec -n {event['namespace']} <pod> -- curl <target>",
                        priority=Priority.WARNING,
                        resource_type=event.get("object_kind", "pod").lower(),
                        resource_name=event["object_name"],
                        namespace=event["namespace"]
                    ))
        except Exception as e:
            # Silently fail for network policy scanning - not all clusters have it
            pass

        return issues

    async def _scan_policy_violations(self, namespace: str = None) -> List[Issue]:
        """Detect policy engine (Kyverno) violations"""
        issues = []
        # Support "all" or empty string or None for all namespaces
        if namespace == "all" or namespace == "" or namespace is None:
            ns = None
        else:
            ns = namespace or self.namespace

        try:
            events = await self.k8s_client.get_events(ns)

            # Look for Kyverno policy violations
            kyverno_events = [e for e in events if "kyverno" in e.get("message", "").lower()
                or "policy" in e.get("reason", "").lower() and "violation" in e.get("message", "").lower()]

            for event in kyverno_events[-5:]:  # Last 5 policy violations
                issues.append(Issue(
                    id=generate_issue_id(event["object_name"], "policy-violation"),
                    title=f"Policy violation: {event['reason']}",
                    description=f"{event['message']}. Check Kyverno policies: kubectl get policyreport -n {event['namespace']}. View violations: kubectl get policyreport -n {event['namespace']} -o yaml",
                    priority=Priority.WARNING,
                    resource_type=event.get("object_kind", "pod").lower(),
                    resource_name=event["object_name"],
                    namespace=event["namespace"]
                ))
        except Exception as e:
            # Silently fail for policy violation scanning - not all clusters have it
            pass

        return issues
