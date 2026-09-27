import React, { useState } from 'react';
import { FileChange, Subagent } from '../types/events';
import { agentHost } from '../host';
import { DiffView } from './DiffView';

interface ChangeTrackerProps {
  changes: FileChange[];
  subagents: Subagent[];
}

/**
 * Groups file changes by owning subagent and renders them in an expandable tree.
 * Shows per-file additions/deletions summary and View Diff / Revert actions.
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

  // Resolve agent role names
  const getAgentLabel = (agentId: string): string => {
    const sub = subagents.find((s) => s.id === agentId);
    return sub ? sub.role.charAt(0).toUpperCase() + sub.role.slice(1) : 'Orchestrator';
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

  const handleRevert = (path: string) => {
    agentHost.postMessage({ command: 'revertFile', payload: { path } });
  };

  return (
    <div className="change-tracker">
      <div className="section-label" style={{ marginBottom: '8px' }}>
        CHANGES · {changes.length}
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

              return (
                <div key={change.id} className="change-file-entry">
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
                      <button
                        className="change-action-btn"
                        onClick={() => handleRevert(change.path)}
                        title="Revert this change"
                      >
                        Revert
                      </button>
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
