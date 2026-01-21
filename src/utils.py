import os
import uuid
import hashlib
from dataclasses import dataclass
from enum import Enum
from typing import Optional

class Priority(str, Enum):
    CRITICAL = "critical"
    WARNING = "warning"
    INFO = "info"

@dataclass
class Issue:
    id: str
    title: str
    description: str
    priority: Priority
    resource_type: str
    resource_name: str
    namespace: str

def generate_issue_id(resource_name: str, issue_type: str) -> str:
    """Generate a consistent ID for an issue"""
    content = f"{resource_name}-{issue_type}"
    return hashlib.md5(content.encode()).hexdigest()[:8]

def get_pod_namespace() -> str:
    """
    Get the current pod namespace.
    Returns 'default' if not running in a pod or if config cannot be read.
    """
    # Check env var first
    if ns := os.environ.get("POD_NAMESPACE"):
        return ns

    # Check service account file
    ns_path = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
    if os.path.exists(ns_path):
        try:
            with open(ns_path, "r") as f:
                return f.read().strip()
        except:
            pass

    return "default"
