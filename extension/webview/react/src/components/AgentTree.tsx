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

export const AgentTree: React.FC<AgentTreeProps> = ({ subagents, selectedId, onSelect }) => {
  return (
    <div style={{ margin: '8px 0' }}>
      <div className="section-label">AGENTS</div>
      <div className="agent-tree-sidebar">
        <div className="agent-tree-node" style={{ fontWeight: 600, color: 'var(--text-primary)' }}>
          <span className="status-dot working" style={{ width: '6px', height: '6px' }} />
          <span>Orchestrator</span>
        </div>

        {subagents.length === 0 ? (
          <div style={{ paddingLeft: '22px', color: 'var(--text-muted)', fontSize: '11px', marginTop: '2px' }}>
            Idle
          </div>
        ) : (
          subagents.map((sub, idx) => {
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
          })
        )}
      </div>
    </div>
  );
};
