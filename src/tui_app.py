import asyncio
from datetime import datetime
from textual.app import App, ComposeResult
from textual.containers import Container, Horizontal, Vertical
from textual.widgets import Header, Footer, Static, Input, TextArea, Tabs, TabPane, Log, Tree
from textual.reactive import reactive
from textual.binding import Binding
from .k8s_client import K8sClient
from .discovery import Discovery
from .agents import AgentSystem
from .storage import Storage
from .utils import get_pod_namespace, Priority

class StargazerTUI(App):
    """Stargazer TUI Application"""
    
    CSS = """
    .container {
        padding: 1;
    }
    
    .header {
        background: $primary;
        text-align: center;
        padding: 0 1;
    }
    
    .input-container {
        height: 3;
        border: solid $primary;
    }
    
    .output-container {
        height: 1fr;
    }
    
    .sidebar {
        width: 25%;
        border-right: solid $primary;
    }
    
    .main-content {
        width: 75%;
    }
    
    .health-panel {
        height: 10;
        border: solid $secondary;
        margin: 1;
    }
    
    .log-panel {
        height: 1fr;
        margin: 1;
    }
    
    .agent-tab {
        padding: 1;
    }
    
    .issue-item {
        margin: 0 1;
        padding: 1;
        background: $surface;
    }
    
    .critical {
        background: $error 30%;
        border-left: solid $error;
    }
    
    .warning {
        background: $warning 30%;
        border-left: solid $warning;
    }
    
    .info {
        background: $primary 10%;
        border-left: solid $primary;
    }
    """
    
    BINDINGS = [
        Binding("ctrl+c", "quit", "Quit"),
        Binding("ctrl+r", "refresh", "Refresh"),
        Binding("ctrl+h", "health", "Health"),
        Binding("ctrl+s", "scan", "Scan"),
        Binding("tab", "focus_next", "Focus Next"),
        Binding("shift+tab", "focus_previous", "Focus Previous"),
    ]
    
    current_agent: reactive[str] = reactive("troubleshooter")
    
    def __init__(self):
        super().__init__()
        self.k8s_client = K8sClient()
        self.discovery = Discovery(self.k8s_client)
        self.agent_system = AgentSystem()
        self.storage = Storage()
        self.issues = []
        self.health_data = None
    
    def compose(self) -> ComposeResult:
        yield Header()
        
        with Container(id="main-container"):
            with Horizontal():
                # Sidebar
                with Vertical(classes="sidebar"):
                    yield Static("🏥 Cluster Health", classes="header")
                    yield Static("Loading...", id="health-panel", classes="health-panel")
                    
                    yield Static("🤖 Agents", classes="header")
                    with Container():
                        for agent_name in self.agent_system.list_agents():
                            yield Static(f"@{agent_name}", id=f"agent-{agent_name}", classes="agent-item")
                    
                    yield Static("⚡ Quick Actions", classes="header")
                    with Container():
                        yield Static("/help", classes="action-item")
                        yield Static("/agents", classes="action-item")
                        yield Static("!kubectl get pods", classes="action-item")
                
                # Main content
                with Vertical(classes="main-content"):
                    # Tabs for different views
                    with Tabs(id="main-tabs"):
                        with TabPane("Chat", id="chat-tab"):
                            with Vertical(classes="output-container"):
                                yield Log(id="chat-log", classes="log-panel")
                            
                            with Container(classes="input-container"):
                                yield Input(placeholder="Enter command (type /help for commands)", id="command-input")
                        
                        with TabPane("Issues", id="issues-tab"):
                            yield Container(id="issues-container", classes="output-container")
                        
                        with TabPane("Discovery", id="discovery-tab"):
                            with Vertical(classes="output-container"):
                                yield Log(id="discovery-log", classes="log-panel")
                        
                        with TabPane("Resources", id="resources-tab"):
                            with Vertical(classes="output-container"):
                                yield Tree("Cluster Resources", id="resource-tree")
        
        yield Footer()
    
    def on_mount(self) -> None:
        """Initialize the application"""
        self.title = "Stargazer - Kubernetes Troubleshooting"
        self.sub_title = f"Namespace: {get_pod_namespace()}"
        
        # Start background tasks
        self.set_interval(5.0, self.update_health)
        self.set_interval(10.0, self.scan_issues)
        
        # Initial updates
        self.update_health()
        self.scan_issues()
    
    def update_health(self) -> None:
        """Update health information"""
        async def update():
            try:
                health = await self.discovery.get_resource_health()
                self.health_data = health
                
                health_panel = self.query_one("#health-panel", Static)
                
                status_icon = "✅" if health['overall_health'] == "healthy" else "⚠️"
                health_text = (
                    f"{status_icon} {health['overall_health'].upper()}\n"
                    f"Pods: {health['pods']['healthy']}/{health['pods']['total']}\n"
                    f"Deployments: {health['deployments']['healthy']}/{health['deployments']['total']}\n"
                    f"Events: {health['events']['warnings']}W {health['events']['errors']}E"
                )
                
                health_panel.update(health_text)
            except Exception as e:
                self.query_one("#health-panel", Static).update(f"Error: {e}")
        
        asyncio.create_task(update())
    
    def scan_issues(self) -> None:
        """Scan for issues in background"""
        async def scan():
            try:
                issues = await self.discovery.scan_all()
                self.issues = issues
                
                # Update issues tab
                self.update_issues_display()
                
                # Update chat log with new issues count
                if issues:
                    chat_log = self.query_one("#chat-log", Log)
                    chat_log.write_line(f"🔍 Found {len(issues)} issues at {datetime.now().strftime('%H:%M:%S')}")
            except Exception as e:
                chat_log = self.query_one("#chat-log", Log)
                chat_log.write_line(f"❌ Scan error: {e}")
        
        asyncio.create_task(scan())
    
    def update_issues_display(self) -> None:
        """Update the issues display"""
        issues_container = self.query_one("#issues-container", Container)
        issues_container.remove_children()
        
        if not self.issues:
            issues_container.mount(Static("✅ No issues detected"))
            return
        
        for issue in self.issues:
            priority_class = issue.priority.value
            
            issue_widget = Static(
                f"[{priority_class.upper()}] {issue.title}\n"
                f"{issue.description}\n"
                f"📍 {issue.resource_type}/{issue.resource_name} • {issue.timestamp.strftime('%H:%M:%S')}",
                classes=f"issue-item {priority_class}"
            )
            issues_container.mount(issue_widget)
    
    def on_input_submitted(self, event: Input.Submitted) -> None:
        """Handle command input"""
        command = event.value.strip()
        if not command:
            return
        
        chat_log = self.query_one("#chat-log", Log)
        chat_log.write_line(f"@{self.current_agent}> {command}")
        
        # Clear input
        event.input.value = ""
        
        # Execute command
        async def execute():
            try:
                response = await self.agent_system.execute_command(command)
                self.current_agent = self.agent_system.current_agent
                
                # Format response
                for line in response.split('\n'):
                    chat_log.write_line(line)
                
                chat_log.write_line("")  # Add spacing
                
                # Store important commands in storage
                if command.startswith("scan") or "issues" in command:
                    for issue in self.issues:
                        self.storage.append_issue(issue)
            
            except Exception as e:
                chat_log.write_line(f"❌ Error: {e}")
        
        asyncio.create_task(execute())
    
    def action_refresh(self) -> None:
        """Manual refresh"""
        self.scan_issues()
        self.update_health()
        chat_log = self.query_one("#chat-log", Log)
        chat_log.write_line("🔄 Refreshed...")
    
    def action_health(self) -> None:
        """Show health summary"""
        chat_log = self.query_one("#chat-log", Log)
        chat_log.write_line("🏥 Health Summary:")
        
        if self.health_data:
            health = self.health_data
            chat_log.write_line(f"  Overall: {health['overall_health'].upper()}")
            chat_log.write_line(f"  Pods: {health['pods']['healthy']}/{health['pods']['total']}")
            chat_log.write_line(f"  Deployments: {health['deployments']['healthy']}/{health['deployments']['total']}")
            chat_log.write_line(f"  Events: {health['events']['warnings']} warnings, {health['events']['errors']} errors")
        
        chat_log.write_line("")
    
    def action_scan(self) -> None:
        """Manual scan"""
        chat_log = self.query_one("#chat-log", Log)
        chat_log.write_line("🔍 Starting manual scan...")
        self.scan_issues()
    
    def on_static_clicked(self, event: Static.Clicked) -> None:
        """Handle clicks on sidebar items"""
        static_id = event.static.id
        
        if static_id and static_id.startswith("agent-"):
            agent_name = static_id.replace("agent-", "")
            self.current_agent = agent_name
            self.agent_system.set_current_agent(agent_name)
            
            chat_log = self.query_one("#chat-log", Log)
            chat_log.write_line(f"🤖 Switched to @{agent_name} agent")
        
        elif static_id and static_id == "action-help":
            chat_log = self.query_one("#chat-log", Log)
            chat_log.write_line("📖 Stargazer Commands:")
            chat_log.write_line("  /agents - List available agents")
            chat_log.write_line("  /help - Show this help")
            chat_log.write_line("  @agentname - Switch to agent")
            chat_log.write_line("  @agentname command - Execute on agent")
            chat_log.write_line("  !kubectl command - Execute kubectl")
            chat_log.write_line("  scan - Analyze cluster for issues")
            chat_log.write_line("  health - Show cluster health")
            chat_log.write_line("")
        
        elif static_id and static_id == "action-agents":
            self.on_input_submitted(Input.Submitted(self.query_one("#command-input", Input), "/agents"))
        
        elif static_id and static_id == "action-kubectl":
            self.on_input_submitted(Input.Submitted(self.query_one("#command-input", Input), "!kubectl get pods"))

def run_tui():
    """Run the TUI application"""
    app = StargazerTUI()
    app.run()