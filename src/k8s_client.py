import asyncio
import os
from datetime import datetime, timezone, timedelta
from typing import List, Dict, Any, Optional
from kubernetes import client, config
from kubernetes.client.rest import ApiException

class K8sClient:
    def __init__(self, namespace: str = None):
        self.namespace = namespace or os.getenv("POD_NAMESPACE", "default")
        self.cache = {}
        self.cache_ttl = 30  # seconds
        
        try:
            config.load_incluster_config()
        except:
            try:
                config.load_kube_config()
            except:
                raise Exception("Could not load Kubernetes configuration")
        
        self.v1 = client.CoreV1Api()
        self.apps_v1 = client.AppsV1Api()
        self.batch_v1 = client.BatchV1Api()
        self.networking_v1 = client.NetworkingV1Api()
    
    def _is_cache_valid(self, cache_key: str) -> bool:
        if cache_key not in self.cache:
            return False
        
        cached_time = self.cache[cache_key]["timestamp"]
        return datetime.now(timezone.utc) - cached_time < timedelta(seconds=self.cache_ttl)
    
    def _get_from_cache(self, cache_key: str) -> Optional[Any]:
        if self._is_cache_valid(cache_key):
            return self.cache[cache_key]["data"]
        return None
    
    def _set_cache(self, cache_key: str, data: Any) -> None:
        self.cache[cache_key] = {
            "data": data,
            "timestamp": datetime.now(timezone.utc)
        }
    
    async def get_pods(self, namespace: str = None) -> List[Dict[str, Any]]:
        ns = namespace or self.namespace
        cache_key = f"pods-{ns}"
        
        cached = self._get_from_cache(cache_key)
        if cached:
            return cached
        
        try:
            pods = self.v1.list_namespaced_pod(ns)
            pod_list = [
                {
                    "name": pod.metadata.name,
                    "namespace": pod.metadata.namespace,
                    "status": pod.status.phase,
                    "node": pod.spec.node_name,
                    "ready": self._is_pod_ready(pod),
                    "restarts": sum(container.restart_count for container in pod.status.container_statuses or []),
                    "age": self._calculate_age(pod.metadata.creation_timestamp),
                    "labels": pod.metadata.labels or {},
                    "events": []
                }
                for pod in pods.items
            ]
            
            self._set_cache(cache_key, pod_list)
            return pod_list
        
        except ApiException as e:
            print(f"Error getting pods: {e}")
            return []
    
    async def get_events(self, namespace: str = None) -> List[Dict[str, Any]]:
        ns = namespace or self.namespace
        cache_key = f"events-{ns}"
        
        cached = self._get_from_cache(cache_key)
        if cached:
            return cached
        
        try:
            events = self.v1.list_namespaced_event(ns, field_selector="type!=Normal")
            event_list = [
                {
                    "name": event.metadata.name,
                    "namespace": event.metadata.namespace,
                    "type": event.type,
                    "reason": event.reason,
                    "message": event.message,
                    "object_kind": event.involved_object.kind,
                    "object_name": event.involved_object.name,
                    "timestamp": event.last_timestamp,
                    "age": self._calculate_age(event.first_timestamp)
                }
                for event in events.items
            ]
            
            self._set_cache(cache_key, event_list)
            return event_list
        
        except ApiException as e:
            print(f"Error getting events: {e}")
            return []
    
    async def get_deployments(self, namespace: str = None) -> List[Dict[str, Any]]:
        ns = namespace or self.namespace
        cache_key = f"deployments-{ns}"
        
        cached = self._get_from_cache(cache_key)
        if cached:
            return cached
        
        try:
            deployments = self.apps_v1.list_namespaced_deployment(ns)
            deployment_list = [
                {
                    "name": deployment.metadata.name,
                    "namespace": deployment.metadata.namespace,
                    "replicas": deployment.spec.replicas,
                    "ready": deployment.status.ready_replicas or 0,
                    "up_to_date": deployment.status.updated_replicas or 0,
                    "available": deployment.status.available_replicas or 0,
                    "age": self._calculate_age(deployment.metadata.creation_timestamp),
                    "labels": deployment.metadata.labels or {}
                }
                for deployment in deployments.items
            ]
            
            self._set_cache(cache_key, deployment_list)
            return deployment_list
        
        except ApiException as e:
            print(f"Error getting deployments: {e}")
            return []
    
    def _is_pod_ready(self, pod) -> bool:
        if not pod.status.container_statuses:
            return False
        return all(container.ready for container in pod.status.container_statuses)
    
    def _calculate_age(self, timestamp) -> str:
        if not timestamp:
            return "Unknown"
        
        now = datetime.now(timezone.utc)
        age = now - timestamp
        
        if age.days > 0:
            return f"{age.days}d"
        elif age.seconds > 3600:
            hours = age.seconds // 3600
            return f"{hours}h"
        elif age.seconds > 60:
            minutes = age.seconds // 60
            return f"{minutes}m"
        else:
            return f"{age.seconds}s"
    
    async def get_pod_logs(self, pod_name: str, namespace: str = None, lines: int = 50) -> str:
        ns = namespace or self.namespace
        try:
            logs = self.v1.read_namespaced_pod_log(
                name=pod_name,
                namespace=ns,
                tail_lines=lines
            )
            return logs
        except ApiException as e:
            return f"Error getting logs: {e}"
    
    def clear_cache(self) -> None:
        self.cache.clear()