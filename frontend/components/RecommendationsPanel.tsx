'use client';

import { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface RecommendationsPanelProps {
  namespace?: string;
}

interface Recommendation {
  id: string;
  title: string;
  description: string;
  category: string;
  severity: string;
  service?: string;
  namespace?: string;
  fix: {
    type: string;
    template?: string;
    command?: string;
    manual_steps?: string[];
  };
  impact: string;
}

interface ComplianceScore {
  score: number;
  passed: number;
  total: number;
  details: Record<string, { name: string; passed: boolean }>;
  recommendations_count: number;
}

export default function RecommendationsPanel({ namespace }: RecommendationsPanelProps) {
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [score, setScore] = useState<ComplianceScore | null>(null);
  const [loading, setLoading] = useState(true);
  const [expandedRec, setExpandedRec] = useState<string | null>(null);
  const [applying, setApplying] = useState<string | null>(null);

  useEffect(() => {
    loadRecommendations();
  }, [namespace]);

  const handleApply = async (rec: Recommendation) => {
    if (!rec.fix.template) return;

    try {
      setApplying(rec.id);
      if (rec.fix.type === 'networkpolicy') {
        await apiClient.applyNetworkPolicy(rec.fix.template, rec.namespace);
      } else if (rec.fix.type === 'cilium') {
        await apiClient.applyCiliumPolicy(rec.fix.template, rec.namespace);
      }
      // Reload to show updated state
      await loadRecommendations();
    } catch (error) {
      console.error('Error applying fix:', error);
    } finally {
      setApplying(null);
    }
  };

  const loadRecommendations = async () => {
    try {
      setLoading(true);
      const [recsData, scoreData] = await Promise.all([
        apiClient.getRecommendations(namespace),
        apiClient.getComplianceScore(namespace),
      ]);
      setRecommendations(recsData.recommendations || []);
      setScore(scoreData);
    } catch (error) {
      console.error('Error loading recommendations:', error);
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'text-[#ef4444] bg-[#ef4444]/10 border-[#ef4444]/30';
      case 'high':
        return 'text-[#f59e0b] bg-[#f59e0b]/10 border-[#f59e0b]/30';
      case 'medium':
        return 'text-[#3b82f6] bg-[#3b82f6]/10 border-[#3b82f6]/30';
      case 'low':
        return 'text-[#71717a] bg-[#71717a]/10 border-[#71717a]/30';
      default:
        return 'text-[#71717a] bg-[#71717a]/10 border-[#71717a]/30';
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'security':
        return 'critical';
      case 'performance':
        return 'healthy';
      case 'observability':
        return 'info';
      case 'resilience':
        return 'degraded';
      default:
        return 'info';
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    // Could add a toast notification here
  };

  if (loading) {
    return (
      <div className="card rounded-lg p-8 text-center">
        <Icon name="loading" className="text-[#3b82f6] animate-pulse text-4xl mb-4" />
        <p className="text-[#71717a]">Loading recommendations...</p>
      </div>
    );
  }

  const criticalRecs = recommendations.filter(r => r.severity === 'critical');
  const highRecs = recommendations.filter(r => r.severity === 'high');
  const mediumRecs = recommendations.filter(r => r.severity === 'medium');
  const lowRecs = recommendations.filter(r => r.severity === 'low');

  return (
    <div className="space-y-6">
      {/* Compliance Score Card */}
      {score && (
        <div className={`card rounded-lg p-5 border-l-4 ${
          score.score >= 80 ? 'border-[#10b981]' :
          score.score >= 60 ? 'border-[#f59e0b]' :
          'border-[#ef4444]'
        }`}>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-lg font-semibold text-[#e4e4e7] mb-1">
                Networking Compliance Score
              </h3>
              <p className="text-sm text-[#71717a]">
                Based on {score.total} best practice checks
              </p>
            </div>
            <div className="text-right">
              <div className={`text-4xl font-bold ${
                score.score >= 80 ? 'text-[#10b981]' :
                score.score >= 60 ? 'text-[#f59e0b]' :
                'text-[#ef4444]'
              }`}>
                {score.score}%
              </div>
              <div className="text-xs text-[#71717a] mt-1">
                {score.passed}/{score.total} passed
              </div>
            </div>
          </div>
          <div className="mt-4 h-2 bg-[#1a1a24] rounded-full overflow-hidden">
            <div
              className={`h-full transition-all ${
                score.score >= 80 ? 'bg-[#10b981]' :
                score.score >= 60 ? 'bg-[#f59e0b]' :
                'bg-[#ef4444]'
              }`}
              style={{ width: `${score.score}%` }}
            ></div>
          </div>
          {score.recommendations_count > 0 && (
            <div className="mt-4 text-sm text-[#71717a]">
              {score.recommendations_count} recommendation{score.recommendations_count !== 1 ? 's' : ''} available
            </div>
          )}
        </div>
      )}

      {/* Recommendations by Severity */}
      {recommendations.length === 0 ? (
        <div className="card rounded-lg p-8 text-center">
          <Icon name="healthy" className="text-[#10b981] text-4xl mb-4" />
          <h3 className="text-lg font-semibold text-[#e4e4e7] mb-2">All Best Practices Met!</h3>
          <p className="text-sm text-[#71717a]">
            Your cluster networking configuration follows all best practices.
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {/* Critical */}
          {criticalRecs.length > 0 && (
            <div>
              <h3 className="text-md font-semibold text-[#e4e4e7] mb-3 flex items-center gap-2">
                <Icon name="critical" className="text-[#ef4444]" size="sm" />
                Critical ({criticalRecs.length})
              </h3>
              <div className="space-y-3">
                {criticalRecs.map((rec) => (
                  <RecommendationCard
                    key={rec.id}
                    recommendation={rec}
                    expanded={expandedRec === rec.id}
                    applying={applying === rec.id}
                    onToggle={() => setExpandedRec(expandedRec === rec.id ? null : rec.id)}
                    onApply={() => handleApply(rec)}
                    getSeverityColor={getSeverityColor}
                    getCategoryIcon={getCategoryIcon}
                    copyToClipboard={copyToClipboard}
                  />
                ))}
              </div>
            </div>
          )}

          {/* High */}
          {highRecs.length > 0 && (
            <div>
              <h3 className="text-md font-semibold text-[#e4e4e7] mb-3 flex items-center gap-2">
                <Icon name="degraded" className="text-[#f59e0b]" size="sm" />
                High ({highRecs.length})
              </h3>
              <div className="space-y-3">
                {highRecs.map((rec) => (
                  <RecommendationCard
                    key={rec.id}
                    recommendation={rec}
                    expanded={expandedRec === rec.id}
                    applying={applying === rec.id}
                    onToggle={() => setExpandedRec(expandedRec === rec.id ? null : rec.id)}
                    onApply={() => handleApply(rec)}
                    getSeverityColor={getSeverityColor}
                    getCategoryIcon={getCategoryIcon}
                    copyToClipboard={copyToClipboard}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Medium */}
          {mediumRecs.length > 0 && (
            <div>
              <h3 className="text-md font-semibold text-[#e4e4e7] mb-3 flex items-center gap-2">
                <Icon name="info" className="text-[#3b82f6]" size="sm" />
                Medium ({mediumRecs.length})
              </h3>
              <div className="space-y-3">
                {mediumRecs.map((rec) => (
                  <RecommendationCard
                    key={rec.id}
                    recommendation={rec}
                    expanded={expandedRec === rec.id}
                    applying={applying === rec.id}
                    onToggle={() => setExpandedRec(expandedRec === rec.id ? null : rec.id)}
                    onApply={() => handleApply(rec)}
                    getSeverityColor={getSeverityColor}
                    getCategoryIcon={getCategoryIcon}
                    copyToClipboard={copyToClipboard}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Low */}
          {lowRecs.length > 0 && (
            <div>
              <h3 className="text-md font-semibold text-[#e4e4e7] mb-3 flex items-center gap-2">
                <Icon name="info" className="text-[#71717a]" size="sm" />
                Low ({lowRecs.length})
              </h3>
              <div className="space-y-3">
                {lowRecs.map((rec) => (
                  <RecommendationCard
                    key={rec.id}
                    recommendation={rec}
                    expanded={expandedRec === rec.id}
                    applying={applying === rec.id}
                    onToggle={() => setExpandedRec(expandedRec === rec.id ? null : rec.id)}
                    onApply={() => handleApply(rec)}
                    getSeverityColor={getSeverityColor}
                    getCategoryIcon={getCategoryIcon}
                    copyToClipboard={copyToClipboard}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

interface RecommendationCardProps {
  recommendation: Recommendation;
  expanded: boolean;
  applying: boolean;
  onToggle: () => void;
  onApply: () => void;
  getSeverityColor: (severity: string) => string;
  getCategoryIcon: (category: string) => string;
  copyToClipboard: (text: string) => void;
}

function RecommendationCard({
  recommendation,
  expanded,
  applying,
  onToggle,
  onApply,
  getSeverityColor,
  getCategoryIcon,
  copyToClipboard,
}: RecommendationCardProps) {
  return (
    <div className={`card rounded-lg p-4 border ${getSeverityColor(recommendation.severity)}`}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <Icon name={getCategoryIcon(recommendation.category) as any} size="sm" />
            <h4 className="font-semibold text-[#e4e4e7]">{recommendation.title}</h4>
            <span className={`text-xs px-2 py-0.5 rounded ${getSeverityColor(recommendation.severity)}`}>
              {recommendation.severity.toUpperCase()}
            </span>
          </div>
          <p className="text-sm text-[#71717a] mb-2">{recommendation.description}</p>
          {recommendation.service && (
            <p className="text-xs text-[#71717a] mb-2">
              Service: {recommendation.service}
            </p>
          )}
          <p className="text-xs text-[#10b981] mb-2">
            <strong>Impact:</strong> {recommendation.impact}
          </p>
        </div>
        <div className="flex flex-col gap-2">
          <button
            onClick={onToggle}
            className="ml-4 text-[#71717a] hover:text-[#e4e4e7] transition-colors"
          >
            <Icon name={expanded ? "close" : "info"} size="sm" />
          </button>

          {recommendation.fix.template && (
            <button
              onClick={onApply}
              disabled={applying}
              className={`ml-4 px-3 py-1 text-xs font-semibold rounded transition-all active:scale-95 ${
                applying
                  ? 'bg-[#3b82f6]/20 text-[#3b82f6] cursor-wait'
                  : 'bg-[#3b82f6] text-white hover:bg-[#2563eb] cursor-pointer'
              }`}
            >
              {applying ? 'Applying...' : 'Apply Fix'}
            </button>
          )}
        </div>
      </div>

      {expanded && (
        <div className="mt-4 pt-4 border-t border-[rgba(255,255,255,0.08)]">
          <div className="space-y-4">
            {recommendation.fix.template && (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h5 className="text-sm font-semibold text-[#e4e4e7]">Fix Template</h5>
                  <button
                    onClick={() => copyToClipboard(recommendation.fix.template || '')}
                    className="text-xs text-[#3b82f6] hover:text-[#2563eb]"
                  >
                    Copy YAML
                  </button>
                </div>
                <pre className="bg-[#0a0a0f] p-3 rounded text-xs text-[#e4e4e7] overflow-x-auto">
                  {recommendation.fix.template}
                </pre>
              </div>
            )}

            {recommendation.fix.command && (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h5 className="text-sm font-semibold text-[#e4e4e7]">Apply Command</h5>
                  <button
                    onClick={() => copyToClipboard(recommendation.fix.command || '')}
                    className="text-xs text-[#3b82f6] hover:text-[#2563eb]"
                  >
                    Copy
                  </button>
                </div>
                <code className="block bg-[#0a0a0f] p-3 rounded text-xs text-[#e4e4e7]">
                  {recommendation.fix.command}
                </code>
              </div>
            )}

            {recommendation.fix.manual_steps && recommendation.fix.manual_steps.length > 0 && (
              <div>
                <h5 className="text-sm font-semibold text-[#e4e4e7] mb-2">Manual Steps</h5>
                <ol className="list-decimal list-inside space-y-1 text-sm text-[#71717a]">
                  {recommendation.fix.manual_steps.map((step, idx) => (
                    <li key={idx}>{step}</li>
                  ))}
                </ol>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
