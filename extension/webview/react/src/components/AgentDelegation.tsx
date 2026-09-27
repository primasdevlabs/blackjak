import React from 'react';
import { Subagent } from '../types/events';
import { ArrowRightIcon, ShareIcon } from '@heroicons/react/24/outline';

interface AgentDelegationProps {
  subagents: Subagent[];
}

export const AgentDelegation: React.FC<AgentDelegationProps> = ({ subagents }) => {
  const subagentsWithFindings = subagents.filter((s) => s.findings && s.findings.length > 0);
  if (subagentsWithFindings.length === 0) return null;

  return (
    <div className="card delegation-card">
      <div className="card-header">
        <ShareIcon className="icon header-icon" />
        <span className="card-title">Delegation</span>
      </div>
      <div className="delegation-list">
        {subagentsWithFindings.map((sub) => (
          <div key={sub.id} className="delegation-item">
            <div className="delegation-header">
              <span className="subagent-role">{sub.role.toUpperCase()}</span>
              <ArrowRightIcon className="icon arrow-icon" />
              <span className="delegation-target">Handoff</span>
            </div>
            <div className="findings-list">
              {sub.findings?.map((f, idx) => (
                <div key={idx} className="finding-row">
                  <span className="finding-file">{f.file}{f.line ? `:${f.line}` : ''}</span>
                  <span className="finding-msg">{f.message}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
