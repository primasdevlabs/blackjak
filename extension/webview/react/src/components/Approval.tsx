import React from 'react';
import { ApprovalRequest } from '../types/events';
import { ShieldExclamationIcon } from '@heroicons/react/24/outline';

interface ApprovalProps {
  request: ApprovalRequest | null;
  onRespond: (granted: boolean, reason?: string) => void;
}

export const Approval: React.FC<ApprovalProps> = ({ request, onRespond }) => {
  if (!request) return null;

  return (
    <div className="approval-modal">
      <div className="approval-content">
        <div className="approval-header">
          <ShieldExclamationIcon className="icon warning-icon" />
          <span className="approval-title">Approval Required</span>
        </div>
        <p className="approval-desc">{request.description}</p>

        {request.details && (
          <pre className="approval-details">
            {JSON.stringify(request.details, null, 2)}
          </pre>
        )}

        <div className="approval-actions">
          <button className="btn btn-deny" onClick={() => onRespond(false, 'User denied operation')}>
            Deny
          </button>
          <button className="btn btn-approve" onClick={() => onRespond(true)}>
            Approve & Continue
          </button>
        </div>
      </div>
    </div>
  );
};
