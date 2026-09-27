import React from 'react';
import { Plan as PlanType } from '../types/events';
import { ClipboardDocumentListIcon, CheckCircleIcon, PlayIcon, ClockIcon } from '@heroicons/react/24/outline';

interface PlanProps {
  plan: PlanType | null;
}

export const Plan: React.FC<PlanProps> = ({ plan }) => {
  if (!plan || !plan.steps || plan.steps.length === 0) return null;

  return (
    <div className="card plan-card">
      <div className="card-header">
        <ClipboardDocumentListIcon className="icon header-icon" />
        <span className="card-title">Execution Plan</span>
      </div>
      <div className="plan-steps">
        {plan.steps.map((step, idx) => {
          let statusClass = 'pending';
          let StatusIcon = ClockIcon;

          if (idx < plan.currentStep) {
            statusClass = 'done';
            StatusIcon = CheckCircleIcon;
          } else if (idx === plan.currentStep) {
            statusClass = 'active';
            StatusIcon = PlayIcon;
          }

          return (
            <div key={idx} className={`plan-step step-${statusClass}`}>
              <StatusIcon className="step-icon" />
              <span className="step-text">{step}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
