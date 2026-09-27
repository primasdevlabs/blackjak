import React from 'react';
import { Subagent } from '../types/events';
import { CpuChipIcon, CheckCircleIcon, ArrowPathIcon, ClockIcon, XCircleIcon } from '@heroicons/react/24/outline';

interface AgentTreeProps {
  subagents: Subagent[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export const AgentTree: React.FC<AgentTreeProps> = ({ subagents, selectedId, onSelect }) => {
  if (subagents.length === 0) return null;

  return (
    <div className="card agent-tree-card">
      <div className="card-header">
        <CpuChipIcon className="icon header-icon" />
        <span className="card-title">Orchestrator Agent Hierarchy</span>
      </div>

      <div className="agent-tree-list">
        <div className="tree-orchestrator-node">
          <span className="tree-role">ORCHESTRATOR</span>
          <span className="tree-sub-count">({subagents.length} subagents)</span>
        </div>

        <div className="tree-children">
          {subagents.map((sub, idx) => {
            const isLast = idx === subagents.length - 1;
            const prefix = isLast ? '└─ ' : '├─ ';
            const isSelected = sub.id === selectedId;

            return (
              <div
                key={sub.id}
                className={`tree-subagent-node ${isSelected ? 'selected' : ''}`}
                onClick={() => onSelect(sub.id)}
              >
                <span className="tree-branch">{prefix}</span>
                <span className="subagent-role">{sub.role.toUpperCase()}</span>

                <span className="subagent-status-icon">
                  {sub.status === 'completed' && <CheckCircleIcon className="icon icon-success" />}
                  {sub.status === 'running' && <ArrowPathIcon className="icon icon-spin" />}
                  {sub.status === 'failed' && <XCircleIcon className="icon icon-error" />}
                  {(sub.status === 'created' || sub.status === 'queued' || sub.status === 'waiting') && (
                    <ClockIcon className="icon icon-muted" />
                  )}
                </span>

                <div className="subagent-task-preview">
                  <span className="task-desc">{sub.task}</span>
                  {sub.result && <span className="task-result">✓ {sub.result}</span>}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};
