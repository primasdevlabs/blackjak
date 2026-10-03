import React from 'react';
import { Subagent } from '../types/events';
import { CheckCircleIcon, ArrowPathIcon, XCircleIcon, MinusCircleIcon } from '@heroicons/react/24/outline';

interface AgentTreeProps {
  subagents: Subagent[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

const getStatusIcon = (status: string): React.ReactNode => {
  switch (status) {
    case 'completed':
      return <CheckCircleIcon className="tree-status-icon" />;
    case 'running':
      return <ArrowPathIcon className="tree-status-icon icon-spin" />;
    case 'failed':
      return <XCircleIcon className="tree-status-icon" style={{ color: 'var(--error)' }} />;
    default:
      return <MinusCircleIcon className="tree-status-icon" />;
  }
};

/** Only render when there are real subagents — never show an idle Coordinator stub. */
export const AgentTree: React.FC<AgentTreeProps> = ({ subagents, selectedId, onSelect }) => {
  if (!subagents.length) return null;

  return (
    <div className="agent-tree">
      <div className="section-label">Agents</div>
      <div className="agent-tree-sidebar">
        {subagents.map((sub, idx) => {
          const isLast = idx === subagents.length - 1;
          const prefix = isLast ? '\u2514\u2500 ' : '\u251C\u2500 ';
          const isSelected = sub.id === selectedId;

          return (
            <React.Fragment key={sub.id}>
              <div
                className={`agent-tree-node${isSelected ? ' selected' : ''}`}
                onClick={() => onSelect(sub.id)}
              >
                <span className="tree-prefix">{prefix}</span>
                <span className="tree-role">{sub.role}</span>
                {getStatusIcon(sub.status)}
              </div>
              {sub.activity && sub.status === 'running' && (
                <div className="tree-activity">{sub.activity}</div>
              )}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
};
