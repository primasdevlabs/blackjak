import React from 'react';

interface ActivityIndicatorProps {
  className?: string;
  /** done | error | idle suppress the pulse */
  state?: 'live' | 'done' | 'error' | 'idle';
}

/** Cursor-like status mark — pulse dot, not a legacy refresh spinner. */
export const ActivityIndicator: React.FC<ActivityIndicatorProps> = ({
  className = '',
  state = 'live',
}) => {
  if (state === 'done') {
    return <span className={`activity-mark activity-mark-done ${className}`} aria-hidden />;
  }
  if (state === 'error') {
    return <span className={`activity-mark activity-mark-error ${className}`} aria-hidden />;
  }
  if (state === 'idle') {
    return <span className={`activity-mark activity-mark-idle ${className}`} aria-hidden />;
  }
  return <span className={`activity-mark activity-mark-live ${className}`} aria-hidden />;
};
