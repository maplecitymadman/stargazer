'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import RecommendationsPanel from './RecommendationsPanel';
import { Icon } from './SpaceshipIcons';

interface CompliancePageProps {
  subsection?: string;
  namespace?: string;
}

export default function CompliancePage({ subsection, namespace }: CompliancePageProps) {
  const [activeTab, setActiveTab] = useState(subsection || 'score');
  const [score, setScore] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (subsection) {
      setActiveTab(subsection);
    }
    loadData();
  }, [subsection, namespace]);

  const loadData = async () => {
    try {
      const scoreData = await apiClient.getComplianceScore(namespace);
      setScore(scoreData);
    } catch (error) {
      console.error('Error loading compliance data:', error);
    } finally {
      setLoading(false);
    }
  };

  const tabs = [
    { id: 'score', label: 'Compliance Score', icon: 'scan' as const },
    { id: 'recommendations', label: 'Recommendations', icon: 'scan' as const },
  ];

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading compliance data...</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Tab Navigation */}
      <div className="flex gap-2 border-b border-[rgba(255,255,255,0.08)] overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-medium transition-all border-b-2 whitespace-nowrap cursor-pointer active:scale-[0.98] ${
              activeTab === tab.id
                ? 'text-[#3b82f6] border-[#3b82f6]'
                : 'text-[#71717a] border-transparent hover:text-[#e4e4e7]'
            }`}
            aria-label={`Switch to ${tab.label} tab`}
          >
            <div className="flex items-center gap-2">
              <Icon name={tab.icon} size="sm" />
              {tab.label}
            </div>
          </button>
        ))}
      </div>

      {/* Content */}
      <div>
        {activeTab === 'score' && <ComplianceScore score={score} namespace={namespace} />}
        {activeTab === 'recommendations' && <RecommendationsPanel namespace={namespace} />}
      </div>
    </div>
  );
}

// Compliance Score Tab
function ComplianceScore({ score, namespace }: { score: any; namespace?: string }) {
  if (!score) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="degraded" className="text-[#f59e0b] text-4xl mb-4" />
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">No Compliance Data</h3>
        <p className="text-sm text-[#71717a]">Unable to load compliance score</p>
      </div>
    );
  }

  const scoreValue = score.score || 0;
  const passed = score.passed || 0;
  const total = score.total || 0;
  const checkDetails = score.check_details || {};

  const getScoreColor = (score: number) => {
    if (score >= 80) return 'text-[#10b981]';
    if (score >= 60) return 'text-[#f59e0b]';
    return 'text-[#ef4444]';
  };

  const getScoreBg = (score: number) => {
    if (score >= 80) return 'bg-[#10b981]/10 border-[#10b981]/30';
    if (score >= 60) return 'bg-[#f59e0b]/10 border-[#f59e0b]/30';
    return 'bg-[#ef4444]/10 border-[#ef4444]/30';
  };

  return (
    <div className="space-y-4">
      {/* Overall Score Card */}
      <div className={`card rounded-lg p-6 border-2 ${getScoreBg(scoreValue)}`}>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold text-[#e4e4e7] mb-1">Networking Compliance Score</h3>
            <p className="text-sm text-[#71717a]">
              {passed} of {total} checks passed
            </p>
          </div>
          <div className={`text-5xl font-bold ${getScoreColor(scoreValue)}`}>
            {scoreValue}%
          </div>
        </div>
        
        {/* Progress Bar */}
        <div className="mt-4 h-3 bg-[#1a1a24] rounded-full overflow-hidden">
          <div
            className={`h-full transition-all duration-500 ${
              scoreValue >= 80 ? 'bg-[#10b981]' : scoreValue >= 60 ? 'bg-[#f59e0b]' : 'bg-[#ef4444]'
            }`}
            style={{ width: `${scoreValue}%` }}
          />
        </div>
      </div>

      {/* Check Details */}
      <div className="card rounded-lg p-5">
        <h3 className="text-lg font-semibold text-[#e4e4e7] mb-4 flex items-center gap-2">
          <Icon name="scan" className="text-[#3b82f6]" size="sm" />
          Compliance Checks
        </h3>
        <div className="space-y-2">
          {Object.entries(checkDetails).map(([id, detail]: [string, any]) => (
            <div
              key={id}
              className="flex items-center justify-between p-3 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]"
            >
              <div className="flex items-center gap-3">
                <Icon
                  name={detail.passed ? 'healthy' : 'critical'}
                  className={detail.passed ? 'text-[#10b981]' : 'text-[#ef4444]'}
                  size="sm"
                />
                <span className="text-sm text-[#e4e4e7]">{detail.name}</span>
              </div>
              <span
                className={`text-xs px-2 py-1 rounded ${
                  detail.passed
                    ? 'bg-[#10b981]/20 text-[#10b981]'
                    : 'bg-[#ef4444]/20 text-[#ef4444]'
                }`}
              >
                {detail.passed ? 'PASSED' : 'FAILED'}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Recommendations Count */}
      {score.recommendations_count > 0 && (
        <div className="card rounded-lg p-4 bg-[#f59e0b]/5 border border-[#f59e0b]/20">
          <div className="flex items-center gap-2">
            <Icon name="warning" className="text-[#f59e0b]" size="sm" />
            <span className="text-sm text-[#e4e4e7]">
              <span className="font-semibold">{score.recommendations_count}</span> recommendations available to improve compliance
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
