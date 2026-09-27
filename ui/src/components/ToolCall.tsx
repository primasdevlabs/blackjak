import React from 'react';
import { ToolExecution } from '../state/agentStore';
import { WrenchScrewdriverIcon, CheckCircleIcon, ArrowPathIcon, XCircleIcon } from '@heroicons/react/24/outline';

interface ToolCallProps {
  executions: ToolExecution[];
}

export const ToolCall: React.FC<ToolCallProps> = ({ executions }) => {
  if (executions.length === 0) return null;

  return (
    <div className="card tool-card">
      <div className="card-header">
        <WrenchScrewdriverIcon className="icon header-icon" />
        <span className="card-title">Tools Execution</span>
      </div>
      <div className="tool-list">
        {executions.map((tool) => (
          <div key={tool.id} className={`tool-item tool-${tool.status}`}>
            <span className="tool-status-icon">
              {tool.status === 'completed' && <CheckCircleIcon className="icon icon-success" />}
              {tool.status === 'running' && <ArrowPathIcon className="icon icon-spin" />}
              {tool.status === 'failed' && <XCircleIcon className="icon icon-error" />}
            </span>
            <span className="tool-name">{tool.name}</span>
            <span className="tool-step">{tool.step}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
