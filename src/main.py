import asyncio
import sys
import time
from datetime import datetime
import click
from .k8s_client import K8sClient
from .discovery import Discovery
from .agents import AgentSystem
from .storage import Storage
from .utils import get_pod_namespace, Priority

try:
    from .tui_app import run_tui
    TUI_AVAILABLE = True
except ImportError:
    TUI_AVAILABLE = False

@click.group()
@click.option('--namespace', '-n', default=None, help='Kubernetes namespace to operate in')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose output')
@click.pass_context
def cli(ctx, namespace, verbose):
    """Stargazer - Lightweight Kubernetes Troubleshooting Tool"""
    ctx.ensure_object(dict)
    ctx.obj['namespace'] = namespace or get_pod_namespace()
    ctx.obj['verbose'] = verbose
    ctx.obj['k8s_client'] = K8sClient(ctx.obj['namespace'])
    ctx.obj['discovery'] = Discovery(ctx.obj['k8s_client'])
    ctx.obj['agent_system'] = AgentSystem()
    ctx.obj['storage'] = Storage()

@cli.command()
@click.option('--continuous', '-c', is_flag=True, help='Run continuous scanning')
@click.option('--interval', '-i', default=2, help='Scan interval in seconds (default: 2)')
@click.option('--namespace', '-n', default=None, help='Namespace to scan (use "all" for all namespaces, default: current namespace)')
@click.pass_context
def scan(ctx, continuous, interval, namespace):
    """Scan cluster for issues"""
    async def run_scan():
        discovery = ctx.obj['discovery']
        storage = ctx.obj['storage']
        agent_system = ctx.obj['agent_system']

        while True:
            try:
                ns_label = namespace or "current namespace" if namespace != "all" else "all namespaces"
                click.echo(f"\n🔍 Scanning {ns_label} at {datetime.now().strftime('%H:%M:%S')}...")

                issues = await discovery.scan_all(namespace)

                if not issues:
                    click.echo("✅ No issues detected in the cluster")
                else:
                    click.echo(f"⚠️  Found {len(issues)} issues:")

                    # Group by priority
                    critical = [i for i in issues if i.priority == Priority.CRITICAL]
                    warning = [i for i in issues if i.priority == Priority.WARNING]
                    info = [i for i in issues if i.priority == Priority.INFO]

                    if critical:
                        click.echo(f"\n🔴 CRITICAL ({len(critical)}):")
                        for issue in critical[:5]:
                            click.echo(f"  • {issue.title}")
                            click.echo(f"    {issue.description}")
                            click.echo(f"    📍 {issue.resource_type}/{issue.resource_name}")

                    if warning:
                        click.echo(f"\n🟡 WARNING ({len(warning)}):")
                        for issue in warning[:5]:
                            click.echo(f"  • {issue.title}")
                            click.echo(f"    📍 {issue.resource_type}/{issue.resource_name}")

                    if info:
                        click.echo(f"\n🔵 INFO ({len(info)}):")
                        for issue in info[:3]:
                            click.echo(f"  • {issue.title}")
                            click.echo(f"    📍 {issue.resource_type}/{issue.resource_name}")

                    # Store issues
                    for issue in issues:
                        storage.append_issue(issue)

                if not continuous:
                    break

                await asyncio.sleep(interval)

            except KeyboardInterrupt:
                click.echo("\n👋 Stopping scan...")
                break
            except Exception as e:
                click.echo(f"❌ Error during scan: {e}")
                if ctx.obj['verbose']:
                    import traceback
                    traceback.print_exc()
                if not continuous:
                    break
                await asyncio.sleep(interval)

    asyncio.run(run_scan())

@cli.command()
@click.argument('query', default='')
@click.pass_context
def ask(ctx, query):
    """Ask the AI troubleshooter about issues"""
    async def run_ask():
        agent_system = ctx.obj['agent_system']

        if not query:
            # Interactive mode
            click.echo("🤖 Stargazer AI Troubleshooter (Type 'exit' to quit)")
            while True:
                try:
                    command = input(f"@{agent_system.current_agent}> ").strip()
                    if command.lower() in ['exit', 'quit']:
                        break
                    elif not command:
                        continue

                    response = await agent_system.execute_command(command)
                    click.echo(response)

                except KeyboardInterrupt:
                    break
                except Exception as e:
                    click.echo(f"❌ Error: {e}")
        else:
            response = await agent_system.execute_command(query)
            click.echo(response)

    asyncio.run(run_ask())

@cli.command()
@click.option('--namespace', '-n', default=None, help='Namespace to check (use "all" for all namespaces, default: current namespace)')
@click.pass_context
def health(ctx, namespace):
    """Get cluster health summary"""
    async def run_health():
        discovery = ctx.obj['discovery']
        health = await discovery.get_resource_health(namespace)

        click.echo("🏥 Cluster Health Summary:")
        click.echo(f"  Pods: {health['pods']['healthy']}/{health['pods']['total']} healthy")
        click.echo(f"  Deployments: {health['deployments']['healthy']}/{health['deployments']['total']} healthy")
        click.echo(f"  Warning Events: {health['events']['warnings']}")
        click.echo(f"  Error Events: {health['events']['errors']}")

        status_icon = "✅" if health['overall_health'] == "healthy" else "⚠️"
        click.echo(f"\n{status_icon} Overall Status: {health['overall_health'].upper()}")

    asyncio.run(run_health())

@cli.command()
@click.option('--lines', '-l', default=50, help='Number of log lines to fetch')
@click.argument('pod_name')
@click.pass_context
def logs(ctx, pod_name, lines):
    """Get logs for a specific pod"""
    async def run_logs():
        k8s_client = ctx.obj['k8s_client']

        try:
            logs = await k8s_client.get_pod_logs(pod_name, lines=lines)
            click.echo(f"📋 Logs for {pod_name} (last {lines} lines):")
            click.echo("-" * 50)
            click.echo(logs)
        except Exception as e:
            click.echo(f"❌ Error getting logs: {e}")

    asyncio.run(run_logs())

@cli.command()
@click.argument('query')
@click.pass_context
def exec(ctx, query):
    """Execute agent command"""
    async def run_exec():
        agent_system = ctx.obj['agent_system']

        try:
            response = await agent_system.execute_command(query)
            click.echo(response)
        except Exception as e:
            click.echo(f"❌ Error executing command: {e}")

    asyncio.run(run_exec())

@cli.command()
@click.option('--mode', default='tui', help='Interface mode (cli/tui)')
@click.pass_context
def start(ctx, mode):
    """Start Stargazer with specified interface mode"""
    if mode == 'tui':
        if TUI_AVAILABLE:
            run_tui()
        else:
            click.echo("❌ TUI not available. Install textual package.")
    elif mode == 'cli':
        click.echo("🚀 Stargazer CLI mode ready. Use 'stargazer scan', 'stargazer ask', etc.")
    else:
        click.echo(f"❌ Unknown mode: {mode}")

if __name__ == '__main__':
    cli()
