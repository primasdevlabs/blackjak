import React, { useState } from 'react';
import { FileChange, Subagent } from '../types/events';
import { agentHost } from '../host';
import { agentStore } from '../state/agentStore';
import { DiffView } from './DiffView';
import {
  CheckIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';

interface ChangeTrackerProps {
  changes: FileChange[];
  subagents: Subagent[];
}

/**
 * Groups file changes by owning agent and renders them with per-change
 * approve / decline review actions, matching the Devin-style workflow:
 * changes are written live so the agent keeps working, and declining a
 * change reverts the file on disk (delete for created, restore for edits).
 */
export const ChangeTracker: React.FC<ChangeTrackerProps> = ({ changes, subagents }) => {
  const [expandedDiff, setExpandedDiff] = useState<string | null>(null);

  if (!changes || changes.length === 0) return null;

  // Group changes by agentId
  const groupedChanges = new Map<string, FileChange[]>();
  for (const change of changes) {
    const key = change.agentId || 'orchestrator';
    if (!groupedChanges.has(key)) {
      groupedChanges.set(key, []);
    }
    groupedChanges.get(key)!.push(change);
  }

  const pendingCount = changes.filter((c) => !c.status || c.status === 'pending').length;

  // Resolve agent role names
  const getAgentLabel = (agentId: string): string => {
    const sub = subagents.find((s) => s.id === agentId);
    return sub ? sub.role.charAt(0).toUpperCase() + sub.role.slice(1) : 'Coordinator';
  };

  const getTypeLetter = (type: string): string => {
    switch (type.toLowerCase()) {
      case 'modified': return 'M';
      case 'created': return 'A';
      case 'deleted': return 'D';
      case 'renamed': return 'R';
      case 'moved': return 'V';
      default: return 'M';
    }
  };

  const getDiffStats = (diff?: string): { added: number; removed: number } => {
    if (!diff) return { added: 0, removed: 0 };
    const lines = diff.split('\n');
    let added = 0;
    let removed = 0;
    for (const line of lines) {
      if (line.startsWith('+') && !line.startsWith('+++')) added++;
      if (line.startsWith('-') && !line.startsWith('---')) removed++;
    }
    return { added, removed };
  };

  const handleOpenFile = (path: string) => {
    agentHost.openFile(path);
  };

  const handleViewDiff = (path: string) => {
    agentHost.openDiff(path);
  };

  const review = (change: FileChange, action: 'accept' | 'reject') => {
    agentStore.reviewFileChange(change.id, action);
  };

  return (
    <div className="change-tracker">
      <div className="change-tracker-header">
        <div className="section-label">
          CHANGES · {changes.length}
          {pendingCount > 0 && <span className="change-pending-count">{pendingCount} pending</span>}
        </div>
        {pendingCount > 0 && (
          <div className="change-bulk-actions">
            <button
              className="change-action-btn change-btn-accept"
              onClick={() => agentStore.reviewAllFileChanges('accept')}
              title="Keep all pending changes"
            >
              <CheckIcon className="icon-sm" /> Approve all
            </button>
            <button
              className="change-action-btn change-btn-decline"
              onClick={() => agentStore.reviewAllFileChanges('reject')}
              title="Revert all pending changes"
            >
              <XMarkIcon className="icon-sm" /> Decline all
            </button>
          </div>
        )}
      </div>

      {Array.from(groupedChanges.entries()).map(([agentId, agentChanges]) => (
        <div key={agentId} className="change-group">
          <div className="change-group-header">
            <span className="change-group-label">{getAgentLabel(agentId)}</span>
            <span className="change-group-count">{agentChanges.length} file{agentChanges.length !== 1 ? 's' : ''}</span>
          </div>

          <div className="change-group-files">
            {agentChanges.map((change) => {
              const letter = getTypeLetter(change.type);
              const stats = getDiffStats(change.diff);
              const isExpanded = expandedDiff === change.id;
              const status = change.status || 'pending';

              return (
                <div key={change.id} className={`change-file-entry ${status === 'rejected' ? 'change-rejected' : ''}`}>
                  <div className="change-file-row">
                    <span
                      className={`change-badge ${letter}`}
                      onClick={() => handleViewDiff(change.path)}
                      title="View diff in editor"
                    >
                      {letter}
                    </span>
                    <span
                      className="change-file-path"
                      onClick={() => handleOpenFile(change.path)}
                      title="Open file"
                    >
                      {change.path}
                    </span>

                    <span className="change-stats">
                      {stats.added > 0 && <span className="stat-added">+{stats.added}</span>}
                      {stats.removed > 0 && <span className="stat-removed">-{stats.removed}</span>}
                    </span>

                    <div className="change-actions">
                      {change.diff && (
                        <button
                          className="change-action-btn"
                          onClick={() => setExpandedDiff(isExpanded ? null : change.id)}
                          title="Toggle inline diff"
                        >
                          {isExpanded ? 'Hide' : 'Diff'}
                        </button>
                      )}

                      {status === 'pending' && (
                        <>
                          <button
                            className="change-action-btn change-btn-accept"
                            onClick={() => review(change, 'accept')}
                            title="Keep this change"
                          >
                            <CheckIcon className="icon-sm" />
                          </button>
                          <button
                            className="change-action-btn change-btn-decline"
                            onClick={() => review(change, 'reject')}
                            title={change.canRevert === false ? 'Cannot revert (no snapshot)' : 'Revert this change'}
                            disabled={change.canRevert === false}
                          >
                            <XMarkIcon className="icon-sm" />
                          </button>
                        </>
                      )}
                      {status === 'accepted' && <span className="change-status accepted">Approved</span>}
                      {status === 'rejected' && <span className="change-status rejected">Declined · reverted</span>}
                    </div>
                  </div>

                  {isExpanded && change.diff && (
                    <DiffView
                      diffText={change.diff}
                      filePath={change.path}
                      onClose={() => setExpandedDiff(null)}
                    />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
};
