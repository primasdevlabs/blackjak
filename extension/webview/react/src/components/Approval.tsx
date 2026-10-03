import React from 'react';
import { ApprovalRequest } from '../types/events';
import { ShieldExclamationIcon } from '@heroicons/react/24/outline';
import { WalkthroughQuestion } from './WalkthroughQuestion';

interface ApprovalProps {
  request: ApprovalRequest | null;
  onRespond: (granted: boolean, reason?: string) => void;
}

function parseWalkthrough(request: ApprovalRequest): { options: string[]; allowOther: boolean } | null {
  if (request.operation !== 'ask_user') return null;
  const details = request.details || {};
  const options = Array.isArray(details.options)
    ? details.options.filter((o: unknown): o is string => typeof o === 'string' && o.trim() !== '')
    : [];
  if (options.length === 0) return null;
  return {
    options,
    allowOther: details.allowOther !== false,
  };
}

export const Approval: React.FC<ApprovalProps> = ({ request, onRespond }) => {
  if (!request) return null;

  const walkthrough = parseWalkthrough(request);
  if (walkthrough) {
    return (
      <div className="approval-modal walkthrough-modal">
        <div className="approval-content walkthrough-content">
          <WalkthroughQuestion
            question={request.description}
            options={walkthrough.options}
            allowOther={walkthrough.allowOther}
            onAnswer={(answer) => onRespond(true, answer)}
            onDismiss={() => onRespond(false, 'dismissed')}
          />
        </div>
      </div>
    );
  }

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
