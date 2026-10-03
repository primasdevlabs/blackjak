import React, { useState } from 'react';
import { Plan as PlanType } from '../types/events';
import { ChevronDownIcon } from '@heroicons/react/24/outline';
import { Plan } from './Plan';

interface PlanCreatedCardProps {
  title?: string;
  description?: string;
  plan: PlanType | null;
  onBuild: () => void;
  onView?: () => void;
}

/** Cursor-style "Created Plan" card with View Plan / Build actions. */
export const PlanCreatedCard: React.FC<PlanCreatedCardProps> = ({
  title = 'Implementation Plan',
  description,
  plan,
  onBuild,
  onView,
}) => {
  const [expanded, setExpanded] = useState(false);
  if (!plan?.steps?.length) return null;

  const summary =
    description ||
    plan.steps.slice(0, 2).join(' · ') + (plan.steps.length > 2 ? ` · +${plan.steps.length - 2} more` : '');

  return (
    <div className="plan-created-card">
      <div className="plan-created-eyebrow">Created Plan</div>
      <div className="plan-created-title">{title}</div>
      <p className="plan-created-desc">{summary}</p>
      {expanded && <Plan plan={plan} title={title} modeLabel="Build" />}
      <div className="plan-created-actions">
        <button
          type="button"
          className="plan-created-link"
          onClick={() => {
            setExpanded((v) => !v);
            onView?.();
          }}
        >
          {expanded ? 'Hide Plan' : 'View Plan'}
        </button>
        <button type="button" className="plan-created-build" onClick={onBuild} title="Build (Ctrl+Enter)">
          Build
          <kbd>Ctrl+↵</kbd>
          <ChevronDownIcon className="icon-xs" />
        </button>
      </div>
    </div>
  );
};
