import React, { useEffect, useState } from 'react';
import { activityManager } from '../activity/activityManager';
import { ActivityStatus } from '../activity/activityTypes';

export const ActivityFeedback: React.FC = () => {
  const [status, setStatus] = useState<ActivityStatus>(activityManager.getStatus());

  useEffect(() => {
    const unsubscribe = activityManager.subscribe((newStatus) => {
      setStatus(newStatus);
    });
    return () => unsubscribe();
  }, []);

  if (status.state === 'idle') return null;

  const displayMsg = status.ambientMessage || status.concreteMessage;
  if (!displayMsg) return null;

  return (
    <div className="activity-transient" style={{ padding: '4px 32px' }}>
      <span>✦</span>
      <span style={{ fontSize: '12px' }}>{displayMsg}</span>
    </div>
  );
};
