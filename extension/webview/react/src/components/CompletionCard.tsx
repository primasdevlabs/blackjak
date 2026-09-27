import React from 'react';
import { CompletionSummary } from '../state/agentStore';
import { CheckCircleIcon, ArrowRightIcon } from '@heroicons/react/24/outline';

interface CompletionCardProps {
  completion: CompletionSummary;
}

function fmtDuration(ms?: number): string {
  if (!ms || ms <= 0) return '';
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  return `${Math.floor(s / 60)}m ${s % 60}s`;
}

export const CompletionCard: React.FC<CompletionCardProps> = ({ completion }) => {
  const stats: string[] = [];
  const dur = fmtDuration(completion.durationMs);
  if (dur) stats.push(dur);
  if (completion.steps) stats.push(`${completion.steps} tool calls`);
  if (completion.filesChanged) stats.push(`${completion.filesChanged} file${completion.filesChanged === 1 ? '' : 's'} changed`);
  if (completion.subagents) stats.push(`${completion.subagents} subagent${completion.subagents === 1 ? '' : 's'}`);

  return (
    <div className="completion-card">
      <div className="completion-header">
        <CheckCircleIcon className="icon completion-icon" />
        <span className="completion-title">Done</span>
        {stats.length > 0 && <span className="completion-stats">{stats.join(' · ')}</span>}
      </div>

      {completion.result && (
        <div className="completion-summary">{completion.result}</div>
      )}

      {completion.recommendations && completion.recommendations.length > 0 && (
        <div className="completion-recs">
          <div className="completion-recs-title">Suggested next steps</div>
          {completion.recommendations.map((r, i) => (
            <div key={i} className="completion-rec">
              <ArrowRightIcon className="icon-sm" />
              <span>{r}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
