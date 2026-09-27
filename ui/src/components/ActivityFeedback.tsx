import React, { useEffect, useState } from 'react';
import { activityManager } from '../activity/activityManager';
import { ActivityStatus } from '../activity/activityTypes';
import { SparklesIcon, ArrowPathIcon, CheckCircleIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';

export const ActivityFeedback: React.FC = () => {
  const [status, setStatus] = useState<ActivityStatus>(activityManager.getStatus());

  useEffect(() => {
    const unsubscribe = activityManager.subscribe((newStatus) => {
      setStatus(newStatus);
    });
    return () => unsubscribe();
  }, []);

  if (status.state === 'idle') return null;

  return (
    <div className="card activity-feedback-card">
      <div className="activity-status-header">
        <span className={`status-indicator status-${status.state}`}>
          {status.state === 'working' && <ArrowPathIcon className="icon icon-spin" />}
          {status.state === 'completed' && <CheckCircleIcon className="icon icon-success" />}
          {status.state === 'error' && <ExclamationTriangleIcon className="icon icon-error" />}
          {status.state === 'approval' && <ExclamationTriangleIcon className="icon icon-warning" />}
          <span className="state-text">{status.state.toUpperCase()}</span>
        </span>
      </div>

      {status.state === 'working' && status.ambientMessage && (
        <div className="ambient-message-row">
          <SparklesIcon className="icon icon-sparkle icon-pulse" />
          <span className="ambient-text">{status.ambientMessage}</span>
        </div>
      )}

      {status.concreteMessage && (
        <div className="concrete-message-row">
          <span className="concrete-dot">●</span>
          <span className="concrete-text">{status.concreteMessage}</span>
        </div>
      )}
    </div>
  );
};
