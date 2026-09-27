import React, { useState, useEffect } from 'react';
import { queueStore } from '../../state/queueStore';
import { TrashIcon, PlayIcon } from '@heroicons/react/24/outline';

export const PromptQueue: React.FC = () => {
  const [queue, setQueue] = useState(queueStore.getQueue());

  useEffect(() => {
    const unsub = queueStore.subscribe(() => {
      setQueue(queueStore.getQueue());
    });
    queueStore.fetchQueue();
    return () => unsub();
  }, []);

  if (queue.length === 0) return null;

  return (
    <div style={{ margin: '8px 0', display: 'flex', flexDirection: 'column', gap: '8px' }}>
      {queue.map((item, idx) => (
        <div key={item.id} className="inline-queued-card">
          <div className="inline-queued-header">
            <span>QUEUED #{idx + 1} ({item.mode.toUpperCase()})</span>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <button
                className="icon-btn"
                title="Run Now"
                onClick={() => queueStore.runItem(item.id)}
              >
                <PlayIcon className="icon-sm" />
              </button>
              <button
                className="icon-btn"
                title="Delete"
                onClick={() => queueStore.removeItem(item.id)}
              >
                <TrashIcon className="icon-sm" />
              </button>
            </div>
          </div>
          <div style={{ color: 'var(--text-primary)', fontSize: '12.5px', lineHeight: '1.4' }}>
            {item.prompt}
          </div>
        </div>
      ))}
    </div>
  );
};
