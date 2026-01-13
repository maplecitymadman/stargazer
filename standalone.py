#!/usr/bin/env python3
"""
Stargazer Standalone Entry Point
For development and testing without external dependencies
"""

import sys
import os
import asyncio
from datetime import datetime

# Add src to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'src'))

def print_banner():
    print("""
    ⭐ STARGAZER - Kubernetes Troubleshooting Tool ⭐
    
    A lightweight, efficient troubleshooting tool that works like
    OpenCode deployed in a cluster.
    
    Features:
    • Auto-discovery scans every 2 seconds
    • Agent system with specialized troubleshooting
    • CLI + TUI interfaces
    • Read-only Kubernetes permissions
    • Persistent issue storage
    
    Usage:
    • Development mode: python standalone.py
    • Docker mode: docker build -t stargazer . && kubectl apply -f kustomization.yaml
    """)

def demo_mode():
    """Run in demo mode without Kubernetes connection"""
    print("\n🚀 Running in Demo Mode")
    print("=" * 50)
    
    print("\n📊 Sample Cluster Health:")
    print("  ✅ Overall: HEALTHY")
    print("  📦 Pods: 12/12 healthy")
    print("  🚀 Deployments: 5/5 healthy")
    print("  📝 Events: 2 warnings, 0 errors")
    
    print("\n🔍 Recent Issues Detected:")
    print("  🟡 WARNING: High restart count for web-app-42")
    print("     Pod has restarted 8 times")
    print("     📍 pod/web-app-42")
    
    print("  🔴 CRITICAL: Deployment api-gateway unavailable")
    print("     Only 1 of 3 replicas available")
    print("     📍 deployment/api-gateway")
    
    print("\n🤖 Available Agents:")
    print("  @troubleshooter - Main agent for Kubernetes troubleshooting")
    print("  @discovery - Discovers and analyzes cluster resources")
    print("  @logs - Retrieves and analyzes pod logs")
    print("  @resource - Analyzes resource usage and constraints")
    print("  @network - Analyzes network connectivity and policies")
    print("  @security - Analyzes security configurations and compliance")
    
    print("\n💡 Example Commands:")
    print("  scan                    - Scan cluster for issues")
    print("  @discovery pods         - List all pods")
    print("  @logs get web-app-42     - Get logs for pod")
    print("  !kubectl get events     - Run kubectl command")
    print("  /agents                 - List all agents")
    print("  /help                   - Show help")

async def interactive_mode():
    """Simple interactive mode for testing"""
    print("\n🎮 Interactive Mode (Type 'exit' to quit)")
    print("=" * 50)
    
    agent_system = None
    try:
        from agents import AgentSystem
        agent_system = AgentSystem()
        print("✅ Agent system loaded")
    except Exception as e:
        print(f"⚠️  Agent system not available: {e}")
        print("Running in simulated mode...")
    
    while True:
        try:
            command = input(f"\n@troubleshooter> ").strip()
            
            if command.lower() in ['exit', 'quit']:
                break
            elif not command:
                continue
            elif command == 'help':
                print("Available commands: scan, health, agents, pods, exit")
            elif command == 'scan':
                print("🔍 Scanning cluster...")
                await asyncio.sleep(1)  # Simulate scan
                print("✅ Scan complete - 2 issues found")
            elif command == 'health':
                print("🏥 Cluster: HEALTHY (12/12 pods, 5/5 deployments)")
            elif command == 'agents':
                print("🤖 Agents: troubleshooter, discovery, logs, resource, network, security")
            elif command == 'pods':
                print("📦 Pods: web-app-42, api-gateway, database, cache, workers (all healthy)")
            elif command.startswith('@'):
                print(f"🤖 Executing: {command}")
                await asyncio.sleep(0.5)
                print("✅ Command executed successfully")
            else:
                print(f"🔍 Processing: {command}")
                await asyncio.sleep(0.3)
                print("💡 Analysis complete - No action required")
        
        except KeyboardInterrupt:
            break
        except EOFError:
            break
    
    print("\n👋 Goodbye!")

def main():
    """Main entry point"""
    print_banner()
    
    if len(sys.argv) > 1:
        mode = sys.argv[1]
        if mode == '--interactive' or mode == '-i':
            asyncio.run(interactive_mode())
        elif mode == '--demo' or mode == '-d':
            demo_mode()
        else:
            print(f"Unknown option: {mode}")
            print("Usage: python standalone.py [--interactive|--demo]")
    else:
        demo_mode()
        print("\n💡 Run 'python standalone.py --interactive' for interactive mode")

if __name__ == '__main__':
    main()