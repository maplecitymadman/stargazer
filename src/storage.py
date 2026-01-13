import json
import os
from typing import List, Dict, Any
try:
    from .utils import Issue, get_pod_namespace
except ImportError:
    from utils import Issue, get_pod_namespace

class Storage:
    def __init__(self, data_dir: str = "/data"):
        self.data_dir = data_dir
        self.issues_file = os.path.join(data_dir, "issues.json")
        os.makedirs(data_dir, exist_ok=True)
    
    def save_issues(self, issues: List[Issue]) -> None:
        try:
            issues_data = [issue.to_dict() for issue in issues]
            with open(self.issues_file, 'w') as f:
                json.dump(issues_data, f, indent=2)
        except Exception as e:
            print(f"Error saving issues: {e}")
    
    def load_issues(self) -> List[Issue]:
        try:
            if not os.path.exists(self.issues_file):
                return []
            
            with open(self.issues_file, 'r') as f:
                issues_data = json.load(f)
            
            return [Issue.from_dict(data) for data in issues_data]
        except Exception as e:
            print(f"Error loading issues: {e}")
            return []
    
    def append_issue(self, issue: Issue) -> None:
        issues = self.load_issues()
        issues.append(issue)
        # Keep only last 1000 issues to prevent file growth
        if len(issues) > 1000:
            issues = issues[-1000:]
        self.save_issues(issues)
    
    def clear_old_issues(self, hours: int = 24) -> None:
        from datetime import datetime, timezone, timedelta
        
        issues = self.load_issues()
        cutoff = datetime.now(timezone.utc) - timedelta(hours=hours)
        
        filtered_issues = [
            issue for issue in issues 
            if issue.timestamp > cutoff
        ]
        
        self.save_issues(filtered_issues)