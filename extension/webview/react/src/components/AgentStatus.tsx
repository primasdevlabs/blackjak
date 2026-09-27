import React from 'react';
import { Subagent } from '../types/events';

interface AgentStatusProps {
  subagents: Subagent[];
}

export const AgentStatus: React.FC<AgentStatusProps> = ({ subagents }) => {
  if (subagents.length === 0) return null;

  const runningCount = subagents.filter((s) => s.status === 'running').length;
  const completedCount = subagents.filter((s) => s.status === 'completed').length;
  const failedCount = subagents.filter((s) => s.status === 'failed').length;

  return (
    <div className="agent-status-bar">
      <div className="status-stat">
        <span className="stat-label">Subagents</span>
        <span className="stat-value">{subagents.length}</span>
      </div>
      <div className="status-stat">
        <span className="stat-label">Running:</span>
        <span className="stat-value running">{runningCount}</span>
      </div>
      <div className="status-stat">
        <span className="stat-label">Completed:</span>
        <span className="stat-value completed">{completedCount}</span>
      </div>
      {failedCount > 0 && (
        <div className="status-stat">
          <span className="stat-label">Failed:</span>
          <span className="stat-value failed">{failedCount}</span>
        </div>
      )}
    </div>
  );
};
