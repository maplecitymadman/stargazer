'use client';

import { Issue } from '@/lib/api';
import { useState } from 'react';
import { Icon } from './SpaceshipIcons';
import IssueTroubleshooter from './IssueTroubleshooter';

interface IssuesListProps {
  issues: Issue[];
}

export default function IssuesList({ issues }: IssuesListProps) {
  const [selectedTab, setSelectedTab] = useState<'all' | 'critical' | 'warning' | 'info'>('all');
  const [selectedIssue, setSelectedIssue] = useState<Issue | null>(null);

  const critical = issues.filter(i => i.priority === 'critical');
  const warning = issues.filter(i => i.priority === 'warning');
  const info = issues.filter(i => i.priority === 'info');

  const displayIssues = selectedTab === 'all' 
    ? issues 
    : selectedTab === 'critical' 
    ? critical 
    : selectedTab === 'warning' 
    ? warning 
    : info;

  const getPriorityStyles = (priority: string) => {
    switch (priority) {
      case 'critical':
        return {
          card: 'border-l-4 border-[#ef4444]',
          badge: 'bg-[#ef4444] text-white',
          icon: 'critical' as const,
        };
      case 'warning':
        return {
          card: 'border-l-4 border-[#f59e0b]',
          badge: 'bg-[#f59e0b] text-white',
          icon: 'warning' as const,
        };
      case 'info':
        return {
          card: 'border-l-4 border-[#3b82f6]',
          badge: 'bg-[#3b82f6] text-white',
          icon: 'info' as const,
        };
      default:
        return {
          card: 'border-l-4 border-[#71717a]',
          badge: 'bg-[#71717a] text-white',
          icon: 'info' as const,
        };
    }
  };

  if (issues.length === 0) {
    return (
      <div className="card rounded-lg p-10 text-center border border-[#10b981]/20">
        <div className="text-4xl mb-4">
          <Icon name="healthy" className="text-[#10b981]" size="lg" />
        </div>
        <div className="text-[#10b981] font-semibold text-lg mb-2">All Systems Clear</div>
        <div className="text-[#71717a] text-sm">No issues detected</div>
      </div>
    );
  }

  return (
    <div>
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
        <div className="card card-hover rounded-lg p-5 border border-[#ef4444]/20 bg-[#ef4444]/5">
          <div className="text-3xl font-semibold text-[#ef4444] mb-1.5">{critical.length}</div>
          <div className="text-xs text-[#71717a] mb-1">Critical Issues</div>
          {critical.length > 0 && (
            <div className="text-xs text-[#ef4444]/70 mt-1.5">Requires immediate attention</div>
          )}
        </div>
        <div className="card card-hover rounded-lg p-5 border border-[#f59e0b]/20 bg-[#f59e0b]/5">
          <div className="text-3xl font-semibold text-[#f59e0b] mb-1.5">{warning.length}</div>
          <div className="text-xs text-[#71717a] mb-1">Warnings</div>
          {warning.length > 0 && (
            <div className="text-xs text-[#f59e0b]/70 mt-1.5">Monitor closely</div>
          )}
        </div>
        <div className="card card-hover rounded-lg p-5 border border-[#3b82f6]/20 bg-[#3b82f6]/5">
          <div className="text-3xl font-semibold text-[#3b82f6] mb-1.5">{info.length}</div>
          <div className="text-xs text-[#71717a] mb-1">Info Items</div>
          {info.length > 0 && (
            <div className="text-xs text-[#3b82f6]/70 mt-1.5">Informational only</div>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="card rounded-lg p-1 mb-4">
        <nav className="flex gap-1">
          {[
            { id: 'all', label: 'All', count: issues.length, icon: 'all' as const },
            { id: 'critical', label: 'Critical', count: critical.length, icon: 'critical' as const },
            { id: 'warning', label: 'Warnings', count: warning.length, icon: 'warning' as const },
            { id: 'info', label: 'Info', count: info.length, icon: 'info' as const },
          ].map(tab => (
            <button
              key={tab.id}
              onClick={() => setSelectedTab(tab.id as any)}
              className={`flex-1 py-2.5 px-4 rounded-md font-medium text-sm transition-all ${
                selectedTab === tab.id
                  ? 'bg-[#3b82f6] text-white'
                  : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[#111118]'
              }`}
            >
              <span className="mr-1.5"><Icon name={tab.icon} size="sm" /></span>
              {tab.label} ({tab.count})
            </button>
          ))}
        </nav>
      </div>

      {/* Issues List */}
      <div className="space-y-3">
            {displayIssues.map(issue => {
          const styles = getPriorityStyles(issue.priority);
          return (
            <div
              key={issue.id}
              className={`${styles.card} card card-hover rounded-lg p-4`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2.5 mb-2.5">
                    <Icon name={styles.icon} className="text-lg" />
                    <span className={`${styles.badge} px-2.5 py-0.5 rounded text-xs font-semibold`}>
                      {issue.priority.toUpperCase()}
                    </span>
                    <h3 className="font-semibold text-[#e4e4e7] text-base">{issue.title}</h3>
                  </div>
                  <div className="flex flex-wrap items-center gap-2 mb-2.5 text-xs">
                    <span className="px-2.5 py-1 rounded bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#a1a1aa]">
                      {issue.resource_type}/{issue.resource_name}
                    </span>
                    <span className="text-[#71717a]">in</span>
                    <span className="px-2.5 py-1 rounded bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#3b82f6]">
                      {issue.namespace}
                    </span>
                  </div>
                  <p className="text-[#a1a1aa] mb-2.5 leading-relaxed text-sm">{issue.description}</p>
                  <div className="flex items-center gap-2 mt-3">
                    <button
                      onClick={() => setSelectedIssue(issue)}
                      className="px-3 py-1.5 bg-[#3b82f6] text-white rounded-md text-xs transition-all flex items-center gap-1.5 hover:bg-[#2563eb]"
                    >
                      <Icon name="scan" className="text-white" size="sm" />
                      AI Troubleshoot
                    </button>
                    <div className="flex items-center gap-1.5 text-xs text-[#71717a]">
                      <Icon name="time" className="text-[#71717a]" size="sm" />
                      <span>{new Date(issue.timestamp).toLocaleString()}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Troubleshooting Modal */}
      {selectedIssue && (
        <IssueTroubleshooter
          issue={selectedIssue}
          onClose={() => setSelectedIssue(null)}
        />
      )}
    </div>
  );
}
