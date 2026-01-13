'use client';

import { useState, useEffect, useRef } from 'react';
import { apiClient, Agent } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface CommandHistory {
  time: Date;
  agent: string;
  command: string;
  response: string;
}

export default function CommandTerminal() {
  const [agents, setAgents] = useState<Agent | null>(null);
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [command, setCommand] = useState('');
  const [response, setResponse] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [history, setHistory] = useState<CommandHistory[]>([]);
  const responseRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadAgents();
  }, []);

  useEffect(() => {
    if (responseRef.current) {
      responseRef.current.scrollTop = responseRef.current.scrollHeight;
    }
  }, [response, history]);

  const loadAgents = async () => {
    try {
      const agentData = await apiClient.getAgents();
      setAgents(agentData);
      setSelectedAgent(agentData.current);
    } catch (error) {
      // Error loading agents - will show empty state
    }
  };

  const executeCommand = async () => {
    if (!command.trim()) return;

    setLoading(true);
    try {
      const result = await apiClient.executeCommand(command, selectedAgent);
      setResponse(result);
      
      // Add to history
      const newHistory: CommandHistory = {
        time: new Date(),
        agent: selectedAgent,
        command,
        response: result,
      };
      setHistory(prev => [...prev.slice(-9), newHistory]);
      
      // Reload agents to get updated current agent
      await loadAgents();
      
      // Clear command
      setCommand('');
    } catch (error: any) {
      setResponse(`Error: ${error.message || 'Failed to execute command'}`);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      executeCommand();
    }
  };

  return (
    <div className="card rounded-lg p-5">
      <h2 className="text-lg font-semibold mb-4 text-[#e4e4e7] flex items-center gap-2">
        <Icon name="terminal" className="text-[#71717a]" size="sm" />
        <span>Command Terminal</span>
      </h2>
      
      {/* Agent and Command Input */}
      <div className="flex gap-2 mb-4">
        <select
          value={selectedAgent}
          onChange={(e) => setSelectedAgent(e.target.value)}
          className="px-3 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7] focus:ring-2 focus:ring-[#3b82f6] focus:border-[#3b82f6] text-sm"
        >
          {agents?.agents.map(agent => (
            <option key={agent} value={agent} className="bg-[#1a1a24]">
              @{agent}
            </option>
          ))}
        </select>
        
        <input
          type="text"
          value={command}
          onChange={(e) => setCommand(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder={`Enter command for @${selectedAgent}...`}
          className="flex-1 px-4 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7] placeholder-[#71717a] focus:ring-2 focus:ring-[#3b82f6] focus:border-[#3b82f6] text-sm"
        />
        
        <button
          onClick={executeCommand}
          disabled={loading || !command.trim()}
          className="px-5 py-2 bg-[#3b82f6] text-white rounded-md hover:bg-[#2563eb] disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-all flex items-center gap-2 text-sm"
        >
          <Icon name={loading ? "loading" : "execute"} className="text-white" size="sm" />
          <span>Execute</span>
        </button>
      </div>

      {/* Response */}
      {response && (
        <div className="mb-4">
          <div className="text-xs font-medium text-[#71717a] mb-2 flex items-center gap-1.5">
            <Icon name="output" className="text-[#71717a]" size="sm" />
            <span>Response</span>
          </div>
          <div
            ref={responseRef}
            className="bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] p-3 rounded-md text-xs overflow-auto max-h-64"
          >
            {response}
          </div>
        </div>
      )}

      {/* History */}
      {history.length > 0 && (
        <div>
          <div className="text-xs font-medium text-[#71717a] mb-2.5 flex items-center gap-1.5">
            <Icon name="history" className="text-[#71717a]" size="sm" />
            <span>Command History (Last 10)</span>
          </div>
          <div className="space-y-2 max-h-64 overflow-auto">
            {history.slice().reverse().map((cmd, idx) => (
              <div key={idx} className="card rounded-md p-3 border border-[rgba(255,255,255,0.08)]">
                <div className="text-xs text-[#71717a] mb-1.5 flex items-center gap-1.5">
                  <Icon name="time" className="text-[#71717a]" size="sm" />
                  <span>{cmd.time.toLocaleTimeString()}</span>
                  <span className="text-[#71717a]">•</span>
                  <code className="bg-[#1a1a24] px-2 py-0.5 rounded border border-[rgba(255,255,255,0.08)] text-[#3b82f6]">
                    @{cmd.agent} &gt; {cmd.command}
                  </code>
                </div>
                <div className="bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#a1a1aa] p-2.5 rounded text-xs overflow-auto">
                  {cmd.response}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
