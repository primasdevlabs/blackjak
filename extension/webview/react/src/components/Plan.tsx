import React from 'react';
import { Plan as PlanType } from '../types/events';
import {
  CheckIcon,
  PlayIcon,
  ShareIcon,
} from '@heroicons/react/24/solid';

interface PlanProps {
  plan: PlanType | null;
  /** Overall task title shown after "Build". */
  title?: string;
  /** Left label in the header pill — Cursor uses "Build". */
  modeLabel?: string;
}

type StepStatus = 'done' | 'active' | 'pending';

function stepStatus(index: number, current: number, total: number): StepStatus {
  if (current >= total) return 'done';
  if (index < current) return 'done';
  if (index === current) return 'active';
  return 'pending';
}

/**
 * Cursor-style Build todo list: header pill + vertical status rail
 * (check / play / hollow) for plan steps.
 */
export const Plan: React.FC<PlanProps> = ({
  plan,
  title,
  modeLabel = 'Build',
}) => {
  if (!plan || !plan.steps || plan.steps.length === 0) return null;

  const heading = title?.trim() || 'Implementation plan';
  const current = Math.max(0, Math.min(plan.currentStep, plan.steps.length));

  return (
    <div className="plan-todo" role="list" aria-label={`${modeLabel}: ${heading}`}>
      <div className="plan-todo-header">
        <span className="plan-todo-mode">{modeLabel}</span>
        <ShareIcon className="plan-todo-tree-icon" aria-hidden />
        <span className="plan-todo-title">{heading}</span>
      </div>

      <div className="plan-todo-list">
        {plan.steps.map((step, idx) => {
          const status = stepStatus(idx, current, plan.steps.length);
          return (
            <div key={idx} className={`plan-todo-item plan-todo-${status}`} role="listitem">
              <div className="plan-todo-rail" aria-hidden>
                {idx < plan.steps.length - 1 && <span className="plan-todo-connector" />}
                <span className={`plan-todo-marker plan-todo-marker-${status}`}>
                  {status === 'done' && <CheckIcon className="plan-todo-check" />}
                  {status === 'active' && <PlayIcon className="plan-todo-play" />}
                </span>
              </div>
              <span className="plan-todo-text">{step}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
