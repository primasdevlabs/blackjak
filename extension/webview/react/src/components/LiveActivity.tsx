import React, { useEffect, useState } from 'react';
import { ToolExecution } from '../state/agentStore';
import { RunStatus } from '../types/events';
import { activityManager } from '../activity/activityManager';
import { ActivityStatus, formatElapsed, phaseLabel } from '../activity/activityTypes';
import {
  ArrowPathIcon,
  CheckCircleIcon,
  XCircleIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  DocumentTextIcon,
} from '@heroicons/react/24/outline';

interface LiveActivityProps {
  status: RunStatus | null;
  tools: ToolExecution[];
  filesRead: string[];
  thought: string;
}

export const LiveActivity: React.FC<LiveActivityProps> = ({ status, tools, filesRead, thought }) => {
  const [expanded, setExpanded] = useState(false);
  const [showFiles, setShowFiles] = useState(false);
  const [activity, setActivity] = useState<ActivityStatus>(() => activityManager.getStatus());
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => activityManager.subscribe(setActivity), []);

  const isLive = status === 'running' || status === 'waiting' || status === 'pending';

  // Tick once per second while live so the elapsed-time label stays current.
  useEffect(() => {
    if (!isLive) return;
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, [isLive]);

  const hasActivity = tools.length > 0 || filesRead.length > 0;
  if (!isLive && !hasActivity) return null;

  const running = tools.filter((t) => t.status === 'running');
  const done = tools.filter((t) => t.status === 'completed');
  const failed = tools.filter((t) => t.status === 'failed');
  const current = running[running.length - 1];
  const recentDone = done.slice(-4);
  const concrete = current?.step || activity.concreteMessage;
  const phaseElapsed = activity.categorySince ? now - activity.categorySince : 0;
  const totalElapsed = activity.workingSince ? now - activity.workingSince : 0;

  return (
    <div className="live-activity">
      {/* Current action line */}
      <div className="live-activity-current" onClick={() => setExpanded((v) => !v)} role="button">
        {isLive ? (
          <ArrowPathIcon className="icon-sm icon-spin live-activity-spinner" />
        ) : failed.length > 0 ? (
          <XCircleIcon className="icon-sm live-activity-icon-error" />
        ) : (
          <CheckCircleIcon className="icon-sm live-activity-icon-done" />
        )}
        <span className="live-activity-label">
          {status === 'waiting'
            ? 'Waiting for approval'
            : isLive
              ? concrete || activity.ambientMessage || 'Thinking…'
              : `${done.length} steps · ${filesRead.length} files scanned`}
        </span>
        {isLive && totalElapsed > 0 && (
          <span className="live-activity-elapsed">{formatElapsed(totalElapsed)}</span>
        )}
        {hasActivity && (
          expanded
            ? <ChevronDownIcon className="icon-sm live-activity-chevron" />
            : <ChevronRightIcon className="icon-sm live-activity-chevron" />
        )}
      </div>

      {/* Rotating ambient status — only when a concrete action is already shown above */}
      {isLive && activity.ambientMessage && concrete && activity.ambientMessage !== concrete && (
        <div className="live-activity-ambient">
          {activity.ambientMessage}
          {phaseElapsed > 1000 && ` · ${formatElapsed(phaseElapsed)}`}
        </div>
      )}

      {/* Thought preview */}
      {isLive && thought && (
        <div className="live-activity-thought">{thought}</div>
      )}

      {/* Expanded detail */}
      {expanded && (
        <div className="live-activity-detail">
          {activity.phases.length > 0 && (
            <div className="live-activity-phases">
              {activity.phases.map((p, i) => (
                <span key={i} className="live-activity-phase">
                  {phaseLabel(p.category, p.message)} · {formatElapsed(p.durationMs)}
                </span>
              ))}
            </div>
          )}
          {recentDone.map((t) => (
            <div key={t.id} className="live-activity-row">
              <CheckCircleIcon className="icon-sm live-activity-icon-done" />
              <span>{t.step || t.name}</span>
            </div>
          ))}
          {running.map((t) => (
            <div key={t.id} className="live-activity-row live-activity-row-active">
              <ArrowPathIcon className="icon-sm icon-spin" />
              <span>{t.step || t.name}</span>
            </div>
          ))}
          {failed.slice(-2).map((t) => (
            <div key={t.id} className="live-activity-row">
              <XCircleIcon className="icon-sm live-activity-icon-error" />
              <span>{t.step || t.name}{t.output ? ` — ${t.output}` : ''}</span>
            </div>
          ))}

          {filesRead.length > 0 && (
            <div className="live-activity-row live-activity-files" onClick={() => setShowFiles((v) => !v)} role="button">
              <DocumentTextIcon className="icon-sm" />
              <span>Scanned {filesRead.length} file{filesRead.length === 1 ? '' : 's'}</span>
              {showFiles
                ? <ChevronDownIcon className="icon-sm live-activity-chevron" />
                : <ChevronRightIcon className="icon-sm live-activity-chevron" />}
            </div>
          )}
          {showFiles && filesRead.slice(-12).map((p) => (
            <div key={p} className="live-activity-row live-activity-file">
              <span className="live-activity-file-path">{p.split(/[/\\]/).slice(-2).join('/')}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
