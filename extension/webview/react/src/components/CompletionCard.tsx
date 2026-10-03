import React from 'react';
import { CompletionSummary } from '../state/agentStore';
import { CheckCircleIcon, ArrowRightIcon } from '@heroicons/react/24/outline';

interface CompletionCardProps {
  completion: CompletionSummary;
  /** Suppress repeating the last assistant reply already shown in chat. */
  hideResult?: boolean;
}

function fmtDuration(ms?: number): string {
  if (!ms || ms <= 0) return '';
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  return `${Math.floor(s / 60)}m ${s % 60}s`;
}

/** True when there is something worth showing beyond a bare "Done". */
export function completionIsMeaningful(
  completion: CompletionSummary | null | undefined,
  lastAgentText?: string
): boolean {
  if (!completion) return false;
  if (completion.recommendations && completion.recommendations.length > 0) return true;
  if ((completion.filesChanged || 0) > 0) return true;
  if ((completion.subagents || 0) > 0) return true;
  if ((completion.steps || 0) > 1) return true;
  const result = (completion.result || '').trim();
  if (!result || /^task completed$/i.test(result)) return false;
  if (lastAgentText && result === lastAgentText.trim()) {
    // Only a duration badge for a reply already in chat — skip the card.
    return false;
  }
  return result.length > 0;
}

export const CompletionCard: React.FC<CompletionCardProps> = ({
  completion,
  hideResult = false,
}) => {
  const stats: string[] = [];
  const dur = fmtDuration(completion.durationMs);
  if (dur) stats.push(dur);
  if (completion.steps) stats.push(`${completion.steps} tool call${completion.steps === 1 ? '' : 's'}`);
  if (completion.filesChanged) {
    stats.push(`${completion.filesChanged} file${completion.filesChanged === 1 ? '' : 's'} changed`);
  }
  if (completion.subagents) {
    stats.push(`${completion.subagents} subagent${completion.subagents === 1 ? '' : 's'}`);
  }

  const showResult =
    !hideResult &&
    !!completion.result &&
    !/^task completed$/i.test(completion.result.trim());

  return (
    <div className="completion-card">
      <div className="completion-header">
        <CheckCircleIcon className="icon completion-icon" />
        <span className="completion-title">Completed</span>
        {stats.length > 0 && <span className="completion-stats">{stats.join(' · ')}</span>}
      </div>

      {showResult && <div className="completion-summary">{completion.result}</div>}

      {completion.recommendations && completion.recommendations.length > 0 && (
        <div className="completion-recs">
          <div className="completion-recs-title">Next steps</div>
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
