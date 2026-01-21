import json
import os
from typing import List, Any
from datetime import datetime

class Storage:
    def __init__(self, storage_path: str = "/data/issues.json"):
        self.storage_path = os.environ.get("STORAGE_PATH", storage_path)
        self.ensure_storage_dir()

    def ensure_storage_dir(self):
        directory = os.path.dirname(self.storage_path)
        if directory and not os.path.exists(directory):
            try:
                os.makedirs(directory, exist_ok=True)
            except OSError:
                # Fallback to local directory if permission denied
                self.storage_path = "issues.json"

    def append_issue(self, issue: Any):
        """Append an issue to storage"""
        issues = self.load_issues()

        # Convert issue object to dict if possible
        if hasattr(issue, "__dict__"):
            issue_data = issue.__dict__
        else:
            issue_data = str(issue)

        # Add timestamp
        if isinstance(issue_data, dict):
            issue_data["timestamp"] = datetime.now().isoformat()

        issues.append(issue_data)
        self.save_issues(issues)

    def load_issues(self) -> List[Any]:
        if not os.path.exists(self.storage_path):
            return []

        try:
            with open(self.storage_path, 'r') as f:
                return json.load(f)
        except:
            return []

    def save_issues(self, issues: List[Any]):
        try:
            with open(self.storage_path, 'w') as f:
                json.dump(issues, f, indent=2)
        except Exception as e:
            print(f"Failed to save issues: {e}")
