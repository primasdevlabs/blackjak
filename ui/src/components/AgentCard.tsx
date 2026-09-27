import React from 'react';
import { Subagent } from '../types/events';
import { UserIcon, CheckCircleIcon, ArrowPathIcon, XCircleIcon, ClockIcon } from '@heroicons/react/24/outline';

interface AgentCardProps {
  subagent: Subagent;
  onClick: () => void;
}

export const AgentCard: React.FC<AgentCardProps> = ({ subagent, onClick }) => {
  return (
    <div className={`subagent-card card-status-${subagent.status}`} onClick={onClick}>
      <div className="subagent-card-header">
        <UserIcon className="icon role-icon" />
        <span className="role-name">{subagent.role.toUpperCase()}</span>
        <span className="status-badge">
          {subagent.status === 'completed' && <CheckCircleIcon className="icon icon-success" />}
          {subagent.status === 'running' && <ArrowPathIcon className="icon icon-spin" />}
          {subagent.status === 'failed' && <XCircleIcon className="icon icon-error" />}
          {(subagent.status === 'created' || subagent.status === 'queued' || subagent.status === 'waiting') && (
            <ClockIcon className="icon icon-muted" />
          )}
          <span>{subagent.status}</span>
        </span>
      </div>

      <div className="subagent-card-body">
        <p className="subagent-task">{subagent.task}</p>
        {subagent.activity && <p className="subagent-activity">✦ {subagent.activity}</p>}
        {subagent.result && <p className="subagent-result">Result: {subagent.result}</p>}
      </div>
    </div>
  );
};
