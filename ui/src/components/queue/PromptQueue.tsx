import React, { useState, useEffect } from 'react';
import { queueStore } from '../../state/queueStore';

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
    <div style={{ margin: '12px 0' }}>
      <div className="section-label">QUEUE</div>
      <div className="font-mono" style={{ padding: '0 8px', fontSize: '11.5px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
        {queue.map((item, idx) => {
          const numStr = String(idx + 1).padStart(2, '0');
          return (
            <div key={item.id} style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-disabled)', width: '20px' }}>{numStr}</span>
                <span style={{ color: 'var(--text-primary)', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {item.prompt}
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingLeft: '20px', fontSize: '10px', color: 'var(--text-muted)' }}>
                <span>{item.status}</span>
                <span style={{ cursor: 'pointer', color: 'var(--text-disabled)' }} onClick={() => queueStore.removeItem(item.id)}>remove</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
