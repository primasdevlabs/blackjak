import React, { useState, useEffect } from 'react';
import { queueStore } from '../../state/queueStore';
import { QueueItem } from './QueueItem';
import { ListBulletIcon, PlusIcon } from '@heroicons/react/24/outline';

export const PromptQueue: React.FC = () => {
  const [queue, setQueue] = useState(queueStore.getQueue());
  const [showAdd, setShowAdd] = useState(false);
  const [newPrompt, setNewPrompt] = useState('');
  const [newMode, setNewMode] = useState<'plan' | 'code'>('code');

  useEffect(() => {
    const unsub = queueStore.subscribe(() => {
      setQueue(queueStore.getQueue());
    });
    queueStore.fetchQueue();
    return () => unsub();
  }, []);

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPrompt.trim()) return;
    queueStore.addPrompt(newPrompt.trim(), newMode);
    setNewPrompt('');
    setShowAdd(false);
  };

  return (
    <div className="card prompt-queue-card">
      <div className="card-header">
        <ListBulletIcon className="icon header-icon" />
        <span className="card-title">Prompt Queue ({queue.length})</span>
        <button className="add-attach-btn" onClick={() => setShowAdd(!showAdd)}>
          <PlusIcon className="icon btn-icon-sm" />
          <span>Queue Task</span>
        </button>
      </div>

      {showAdd && (
        <form className="queue-input-form" onSubmit={handleAdd}>
          <input
            type="text"
            className="queue-text-input"
            placeholder="Enter task prompt to queue..."
            value={newPrompt}
            onChange={(e) => setNewPrompt(e.target.value)}
          />
          <select
            className="select-input mode-select"
            value={newMode}
            onChange={(e) => setNewMode(e.target.value as any)}
          >
            <option value="code">Code</option>
            <option value="plan">Plan</option>
          </select>
          <button type="submit" className="btn btn-approve btn-sm">Add</button>
        </form>
      )}

      <div className="queue-items-list">
        {queue.length === 0 ? (
          <p className="empty-queue-msg">No tasks queued.</p>
        ) : (
          queue.map((item) => (
            <QueueItem
              key={item.id}
              item={item}
              onRun={(id) => queueStore.runItem(id)}
              onDelete={(id) => queueStore.removeItem(id)}
            />
          ))
        )}
      </div>
    </div>
  );
};
