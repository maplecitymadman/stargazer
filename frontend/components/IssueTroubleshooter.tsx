'use client';

import { useState, useEffect } from 'react';
import { Issue } from '@/lib/api';
import { Icon } from './SpaceshipIcons';
import apiClient from '@/lib/api';

interface IssueTroubleshooterProps {
  issue: Issue;
  onClose: () => void;
}

interface Recommendation {
  priority: string;
  title: string;
  description: string;
  steps: string[];
}

interface FixCommand {
  label: string;
  command: string;
  description: string;
  type: 'fix' | 'diagnostic';
  requires_confirmation?: boolean;
  interactive?: boolean;
}

interface TroubleshootingAnalysis {
  issue_id: string;
  analysis: string;
  recommendations: Recommendation[];
  fix_commands: FixCommand[];
  diagnostic_commands: string[];
  estimated_severity: string;
}

export default function IssueTroubleshooter({ issue, onClose }: IssueTroubleshooterProps) {
  const [analysis, setAnalysis] = useState<TroubleshootingAnalysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [executing, setExecuting] = useState<string | null>(null);
  const [executionResult, setExecutionResult] = useState<string | null>(null);
  const [selectedCommand, setSelectedCommand] = useState<FixCommand | null>(null);
  const [showConfirm, setShowConfirm] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);
  const [nextFix, setNextFix] = useState<FixCommand | null>(null);

  useEffect(() => {
    loadAnalysis();
  }, [issue.id]);

  const loadAnalysis = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getIssueRecommendations(issue.id);
      setAnalysis(data);
    } catch (error) {
      // Error loading analysis - will show empty state
    } finally {
      setLoading(false);
    }
  };

  const executeCommand = async (command: FixCommand) => {
    if (command.requires_confirmation && !showConfirm) {
      setSelectedCommand(command);
      setShowConfirm(true);
      return;
    }

    try {
      setExecuting(command.label);
      setExecutionResult(null);
      
      const response = await apiClient.executeFix(issue.id, command.command);
      
      setExecutionResult(response.result);
      
      // If issue is resolved, show success
      if (response.issue_resolved) {
        setNextFix(null);
        // Reload analysis to get updated status
        await loadAnalysis();
        return;
      }
      
      // If there's a next fix recommendation from AI analysis
      if (response.analysis && response.analysis.next_fix) {
        setAnalyzing(true);
        // Simulate AI thinking time
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        const nextFixCmd: FixCommand = {
          label: response.analysis.next_fix.label || "Next AI-Recommended Fix",
          command: response.analysis.next_fix.command,
          description: response.analysis.next_fix.description || "AI analyzed the result and recommends this next step",
          type: response.analysis.next_fix.type || "fix",
          requires_confirmation: response.analysis.next_fix.requires_confirmation || false
        };
        
        setNextFix(nextFixCmd);
        setAnalyzing(false);
        
        // Add to approved commands list
        setAnalysis(prev => {
          if (!prev) return prev;
          return {
            ...prev,
            fix_commands: [
              ...prev.fix_commands,
              nextFixCmd
            ]
          };
        });
      } else if (response.next_fix) {
        // Fallback to direct next_fix
        const nextFixCmd: FixCommand = {
          label: response.next_fix.label || "Next Recommended Fix",
          command: response.next_fix.command,
          description: response.next_fix.description || "AI-recommended next step",
          type: response.next_fix.type || "fix",
          requires_confirmation: response.next_fix.requires_confirmation || false
        };
        setNextFix(nextFixCmd);
        setAnalysis(prev => {
          if (!prev) return prev;
          return {
            ...prev,
            fix_commands: [
              ...prev.fix_commands,
              nextFixCmd
            ]
          };
        });
      }
      
      // Reload analysis to get updated status
      await loadAnalysis();
    } catch (error: any) {
      const errorMsg = error.response?.data?.detail || error.message || 'Unknown error';
      setExecutionResult(`Error: ${errorMsg}`);
      
      // If it's a 403 (not approved), show helpful message
      if (error.response?.status === 403) {
        setExecutionResult(`Security Error: ${errorMsg}\n\nOnly AI-recommended commands can be executed for security. Please use one of the recommended fix commands above.`);
      }
    } finally {
      setExecuting(null);
      setShowConfirm(false);
      setSelectedCommand(null);
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority.toLowerCase()) {
      case 'critical':
        return 'text-red-400';
      case 'high':
        return 'text-yellow-400';
      default:
        return 'text-blue-400';
    }
  };

  const getPriorityBg = (priority: string) => {
    switch (priority.toLowerCase()) {
      case 'critical':
        return 'bg-red-900/20 border-red-500/30';
      case 'high':
        return 'bg-yellow-900/20 border-yellow-500/30';
      default:
        return 'bg-blue-900/20 border-blue-500/30';
    }
  };

  if (loading) {
    return (
      <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <div className="card rounded-xl p-8 max-w-4xl w-full">
          <div className="text-center py-10">
            <Icon name="loading" className="text-[#71717a] animate-pulse text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">Analyzing issue...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 overflow-y-auto">
      <div className="card rounded-xl p-6 max-w-5xl w-full max-h-[90vh] overflow-y-auto my-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6 pb-4 border-b border-[rgba(255,255,255,0.08)]">
          <div className="flex items-center gap-3">
            <Icon name="scan" className="text-[#3b82f6]" size="sm" />
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-2xl font-bold text-[#e4e4e7]">AI Troubleshooting</h2>
                <span className="px-2 py-0.5 rounded text-xs font-medium bg-[#3b82f6]/10 text-[#3b82f6] border border-[#3b82f6]/20">
                  AI-POWERED
                </span>
              </div>
              <p className="text-sm text-[#71717a] mt-1">{issue.title}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="px-3 py-1.5 bg-red-600/90 hover:bg-red-700 text-white rounded-lg text-sm transition-all"
          >
            Close
          </button>
        </div>

        {/* Root Cause Analysis */}
        {analysis && (
          <>
            <div className="mb-6 card rounded-lg p-4 border border-[rgba(255,255,255,0.08)]">
              <div className="flex items-center gap-2 mb-3">
                <Icon name="info" className="text-[#3b82f6]" size="sm" />
                <h3 className="font-bold text-[#e4e4e7]">AI Root Cause Analysis</h3>
                <span className="px-1.5 py-0.5 rounded text-xs bg-[#3b82f6]/10 text-[#3b82f6] border border-[#3b82f6]/20">
                  AI
                </span>
              </div>
              <p className="text-[#e4e4e7] text-sm leading-relaxed">{analysis.analysis || "Analysis in progress..."}</p>
            </div>

            {/* Recommendations */}
            <div className="mb-6">
              <h3 className="font-bold text-[#e4e4e7] mb-4 flex items-center gap-2">
                <Icon name="scan" className="text-[#3b82f6]" size="sm" />
                AI Recommendations
                <span className="px-1.5 py-0.5 rounded text-xs bg-[#3b82f6]/10 text-[#3b82f6] border border-[#3b82f6]/20">
                  AI
                </span>
              </h3>
              {(analysis.recommendations || []).length === 0 ? (
                <div className="card rounded-lg p-4 border border-[rgba(255,255,255,0.08)] text-center text-[#71717a]">
                  No recommendations available yet. AI is analyzing...
                </div>
              ) : (
              <div className="space-y-3">
                {analysis.recommendations.map((rec, idx) => (
                  <div key={idx} className={`card rounded-lg p-4 border ${getPriorityBg(rec.priority)}`}>
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <span className={`px-2 py-0.5 rounded text-xs font-semibold ${getPriorityColor(rec.priority)}`}>
                          {rec.priority}
                        </span>
                        <h4 className="font-bold text-[#e4e4e7]">{rec.title}</h4>
                      </div>
                    </div>
                    <p className="text-[#71717a] text-sm mb-3">{rec.description}</p>
                    <div className="space-y-1.5">
                      {(rec.steps || []).map((step, stepIdx) => (
                        <div key={stepIdx} className="flex items-start gap-2 text-xs text-[#71717a]">
                          <span className="text-[#3b82f6] mt-0.5">▸</span>
                          <span>{step}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
              )}
            </div>

            {/* Fix Commands */}
            <div className="mb-6">
              <h3 className="font-bold text-[#e4e4e7] mb-4 flex items-center gap-2">
                <Icon name="terminal" className="text-[#3b82f6]" size="sm" />
                AI-Generated Fix Commands
                <span className="px-1.5 py-0.5 rounded text-xs bg-[#3b82f6]/10 text-[#3b82f6] border border-[#3b82f6]/20">
                  AI
                </span>
              </h3>
              <p className="text-xs text-[#71717a] mb-4">
                These commands were generated by AI. Only AI-recommended commands can be executed for security.
              </p>
              {(analysis.fix_commands || []).length === 0 ? (
                <div className="card rounded-lg p-4 border border-[rgba(255,255,255,0.08)] text-center text-[#71717a]">
                  No fix commands available yet. AI is generating commands...
                </div>
              ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                {analysis.fix_commands.map((cmd, idx) => (
                  <div
                    key={idx}
                    className={`card rounded-lg p-4 border ${
                      cmd.type === 'fix' ? 'border-[rgba(255,255,255,0.08)]' : 'border-[rgba(255,255,255,0.08)]'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <Icon
                            name={cmd.type === 'fix' ? 'execute' : 'info'}
                            className="text-[#71717a]"
                            size="sm"
                          />
                          <h4 className="font-semibold text-[#e4e4e7] text-sm">{cmd.label}</h4>
                          {cmd.type === 'fix' && (
                            <span className="px-1.5 py-0.5 rounded text-xs font-semibold bg-[#1a1a24] text-[#71717a] border border-[rgba(255,255,255,0.08)]">
                              Fix
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-[#71717a] mb-2">{cmd.description}</p>
                        <code className="block text-xs bg-[#1a1a24] p-2 rounded border border-[rgba(255,255,255,0.08)] text-[#3b82f6] break-all">
                          {cmd.command}
                        </code>
                      </div>
                    </div>
                    <button
                      onClick={() => executeCommand(cmd)}
                      disabled={executing === cmd.label}
                      className={`mt-3 w-full px-3 py-2 rounded-lg text-xs transition-all ${
                        cmd.type === 'fix'
                          ? 'bg-[#3b82f6] hover:bg-[#2563eb] text-white'
                          : 'bg-[#1a1a24] text-[#71717a] border border-[rgba(255,255,255,0.08)] hover:bg-[#27272a]'
                      } disabled:opacity-50`}
                    >
                      {executing === cmd.label ? (
                        <span className="flex items-center justify-center gap-2">
                          <Icon name="loading" className="animate-pulse" size="sm" />
                          Executing...
                        </span>
                      ) : (
                        cmd.type === 'fix' ? 'Execute Fix' : 'Run Diagnostic'
                      )}
                    </button>
                  </div>
                ))}
              </div>
              )}
            </div>

            {/* Execution Result */}
            {executionResult && (
              <div className="mb-6 card rounded-lg p-4 border border-[rgba(255,255,255,0.08)]">
                <div className="flex items-center gap-2 mb-2">
                  <Icon name="output" className="text-[#71717a]" size="sm" />
                  <h3 className="font-bold text-[#e4e4e7]">Execution Result</h3>
                </div>
                <pre className="bg-[#1a1a24] p-3 rounded border border-[rgba(255,255,255,0.08)] text-xs text-[#3b82f6] overflow-x-auto whitespace-pre-wrap">
                  {executionResult}
                </pre>
                
                {/* AI Analyzing Next Steps */}
                {analyzing && (
                  <div className="mt-4 p-3 bg-[#3b82f6]/10 border border-[#3b82f6]/20 rounded-lg">
                    <div className="flex items-center gap-2 text-sm text-[#3b82f6]">
                      <Icon name="loading" className="animate-pulse" size="sm" />
                      <span>AI is analyzing the result and generating next steps...</span>
                    </div>
                  </div>
                )}
                
                {/* Next Fix Recommendation */}
                {nextFix && !analyzing && (
                  <div className="mt-4 p-4 bg-[#3b82f6]/10 border border-[#3b82f6]/30 rounded-lg">
                    <div className="flex items-center gap-2 mb-2">
                      <Icon name="scan" className="text-[#3b82f6]" size="sm" />
                      <h4 className="font-bold text-[#e4e4e7] text-sm">AI Recommended Next Fix</h4>
                      <span className="px-1.5 py-0.5 rounded text-xs bg-[#3b82f6]/20 text-[#3b82f6] border border-[#3b82f6]/30">
                        NEW
                      </span>
                    </div>
                    <p className="text-xs text-[#71717a] mb-3">{nextFix.description}</p>
                    <code className="block text-xs bg-[#1a1a24] p-2 rounded border border-[rgba(255,255,255,0.08)] text-[#3b82f6] break-all mb-3">
                      {nextFix.command}
                    </code>
                    <button
                      onClick={() => {
                        setNextFix(null);
                        executeCommand(nextFix);
                      }}
                      className="w-full px-3 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-xs transition-all"
                    >
                      Execute This Fix
                    </button>
                  </div>
                )}
              </div>
            )}

            {/* Confirmation Dialog */}
            {showConfirm && selectedCommand && (
              <div className="fixed inset-0 bg-black/90 backdrop-blur-sm z-60 flex items-center justify-center p-4">
                <div className="card rounded-xl p-6 max-w-md w-full">
                  <h3 className="font-bold text-[#e4e4e7] mb-3">Confirm Action</h3>
                  <p className="text-[#71717a] text-sm mb-4">
                    This will execute: <code className="text-[#3b82f6]">{selectedCommand.command}</code>
                  </p>
                  <p className="text-yellow-400 text-xs mb-6">
                    ⚠ This action may modify cluster resources. Are you sure?
                  </p>
                  <div className="flex gap-3">
                    <button
                      onClick={() => {
                        setShowConfirm(false);
                        setSelectedCommand(null);
                      }}
                      className="flex-1 px-4 py-2 bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#e4e4e7] rounded-lg text-sm hover:bg-[#27272a]"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={() => executeCommand(selectedCommand)}
                      className="flex-1 px-4 py-2 bg-red-600/90 hover:bg-red-700 text-white rounded-lg text-sm"
                    >
                      Confirm
                    </button>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
