import React from 'react';
import { Subagent } from '../types/events';
import { ClockIcon } from '@heroicons/react/24/outline';

interface AgentTimelineProps {
  subagents: Subagent[];
}

export const AgentTimeline: React.FC<AgentTimelineProps> = ({ subagents }) => {
  if (subagents.length === 0) return null;

  return (
    <div className="card agent-timeline-card">
      <div className="card-header">
        <ClockIcon className="icon header-icon" />
        <span className="card-title">Timeline</span>
      </div>
      <div className="timeline-list">
        {subagents.map((sub) => (
          <div key={sub.id} className="timeline-item">
            <span className="timeline-badge">{sub.role}</span>
            <span className="timeline-task">{sub.task}</span>
            <span className={`timeline-status status-${sub.status}`}>{sub.status}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
