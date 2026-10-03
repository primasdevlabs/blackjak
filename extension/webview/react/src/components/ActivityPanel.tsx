import React, { useEffect, useState } from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { apiClient } from '../api/client';
import { useAgent } from '../hooks/useAgent';
import { ActivityStream } from './ActivityStream';
import { ActivityIndicator } from './ActivityIndicator';

interface ActivityEntry {
  id: string;
  runId?: string;
  type: string;
  message: string;
  timestamp: string;
}

export interface ActivityPanelProps {
  runId?: string | null;
  onClose: () => void;
}

/** Dedicated activity window: live Cursor-style stream + persisted activity log. */
export function ActivityPanel({ runId, onClose }: ActivityPanelProps) {
  const {
    toolExecutions,
    fileChanges,
    filesRead,
    activeRunStatus,
    cancelTask,
  } = useAgent();
  const [entries, setEntries] = useState<ActivityEntry[]>([]);
  const [filterRun, setFilterRun] = useState(!!runId);
  const [tab, setTab] = useState<'stream' | 'log'>('stream');

  const load = async () => {
    try {
      const q = new URLSearchParams({ limit: '200' });
      if (filterRun && runId) q.set('runId', runId);
      const res = await fetch(`${apiClient.getBaseUrl()}/api/activity?${q}`);
      const json = await res.json();
      if (json.success) setEntries(json.data || []);
    } catch {
      setEntries([]);
    }
  };

  useEffect(() => {
    void load();
    const t = setInterval(() => {
      void load();
    }, 2000);
    return () => clearInterval(t);
  }, [runId, filterRun]);

  const isWorking =
    activeRunStatus === 'running' ||
    activeRunStatus === 'waiting' ||
    activeRunStatus === 'pending';

  return (
    <div className="settings-page-modal activity-panel-modal">
      <div className="settings-page-container activity-panel activity-panel-wide">
        <div className="settings-page-header">
          <div className="title-area">
            {isWorking ? (
              <ActivityIndicator state="live" />
            ) : (
              <ActivityIndicator state="idle" />
            )}
            <span className={`page-title ${isWorking ? 'live-activity-shimmer' : ''}`}>
              {isWorking ? 'Working' : 'Activity'}
            </span>
          </div>
          <button className="close-btn" onClick={onClose} title="Close">
            <XMarkIcon className="icon" />
          </button>
        </div>
        <div className="activity-toolbar">
          <div className="activity-tabs">
            <button
              type="button"
              className={`activity-tab ${tab === 'stream' ? 'active' : ''}`}
              onClick={() => setTab('stream')}
            >
              Stream
            </button>
            <button
              type="button"
              className={`activity-tab ${tab === 'log' ? 'active' : ''}`}
              onClick={() => setTab('log')}
            >
              Log
            </button>
          </div>
          {tab === 'log' && (
            <label className="activity-filter">
              <input
                type="checkbox"
                checked={filterRun}
                onChange={(e) => setFilterRun(e.target.checked)}
              />
              Current run
            </label>
          )}
        </div>

        {tab === 'stream' ? (
          <div className="activity-panel-stream">
            <ActivityStream
              tools={toolExecutions}
              fileChanges={fileChanges}
              filesRead={filesRead}
              isWorking={isWorking}
            />
            {!isWorking && toolExecutions.length === 0 && fileChanges.length === 0 && (
              <div className="mention-picker-empty">
                No live activity — switch to Log for history
              </div>
            )}
          </div>
        ) : (
          <div className="activity-list">
            {entries.length === 0 && (
              <div className="mention-picker-empty">No activity yet</div>
            )}
            {entries.map((e) => (
              <div key={e.id} className="activity-row">
                <span className="activity-time">
                  {new Date(e.timestamp).toLocaleTimeString()}
                </span>
                <span className="activity-type">{e.type}</span>
                <span className="activity-msg">{e.message || e.type}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default ActivityPanel;
