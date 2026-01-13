import os
import json
import asyncio
from datetime import datetime, timezone
from enum import Enum
from typing import Dict, List, Optional, Any

class Priority(Enum):
    CRITICAL = "critical"
    WARNING = "warning"
    INFO = "info"

class Issue:
    def __init__(self, 
                 id: str,
                 title: str,
                 description: str,
                 priority: Priority,
                 resource_type: str,
                 resource_name: str,
                 namespace: str = "default",
                 timestamp: Optional[datetime] = None):
        self.id = id
        self.title = title
        self.description = description
        self.priority = priority
        self.resource_type = resource_type
        self.resource_name = resource_name
        self.namespace = namespace
        self.timestamp = timestamp or datetime.now(timezone.utc)
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "title": self.title,
            "description": self.description,
            "priority": self.priority.value,
            "resource_type": self.resource_type,
            "resource_name": self.resource_name,
            "namespace": self.namespace,
            "timestamp": self.timestamp.isoformat()
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Issue':
        return cls(
            id=data["id"],
            title=data["title"],
            description=data["description"],
            priority=Priority(data["priority"]),
            resource_type=data["resource_type"],
            resource_name=data["resource_name"],
            namespace=data.get("namespace", "default"),
            timestamp=datetime.fromisoformat(data["timestamp"])
        )

def generate_issue_id(resource_name: str, issue_type: str) -> str:
    return f"{resource_name}-{issue_type}-{int(datetime.now().timestamp())}"

def get_pod_namespace() -> str:
    return os.getenv("POD_NAMESPACE", "default")

def get_pod_name() -> str:
    return os.getenv("POD_NAME", "stargazer")