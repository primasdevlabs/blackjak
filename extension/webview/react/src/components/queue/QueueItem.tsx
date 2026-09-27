import React from 'react';
import { QueuedPrompt } from '../../api/settings';
import { PlayIcon, TrashIcon, ClockIcon, ArrowPathIcon } from '@heroicons/react/24/outline';

interface QueueItemProps {
  item: QueuedPrompt;
  onRun: (id: string) => void;
  onDelete: (id: string) => void;
}

export const QueueItem: React.FC<QueueItemProps> = ({ item, onRun, onDelete }) => {
  return (
    <div className={`queue-item status-${item.status}`}>
      <span className="queue-status-badge">
        {item.status === 'running' && <ArrowPathIcon className="icon icon-spin" />}
        {item.status === 'queued' && <ClockIcon className="icon icon-muted" />}
        <span>{item.status.toUpperCase()}</span>
      </span>

      <span className="queue-mode-badge">{item.mode.toUpperCase()}</span>

      <div className="queue-prompt-text">{item.prompt}</div>

      <div className="queue-actions">
        {item.status === 'queued' && (
          <button className="queue-btn btn-run" onClick={() => onRun(item.id)} title="Run task now">
            <PlayIcon className="icon" />
          </button>
        )}
        <button className="queue-btn btn-delete" onClick={() => onDelete(item.id)} title="Delete task">
          <TrashIcon className="icon" />
        </button>
      </div>
    </div>
  );
};
