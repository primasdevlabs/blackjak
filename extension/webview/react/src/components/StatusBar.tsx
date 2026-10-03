import React from 'react';
import { ConnectionStatus } from '../api/websocket';
import { RunStatus } from '../types/events';
import { StopIcon } from '@heroicons/react/24/outline';
import { ActivityIndicator } from './ActivityIndicator';

interface StatusBarProps {
  connectionStatus: ConnectionStatus;
  runStatus: RunStatus | null;
  activeRunId: string | null;
  onCancel: () => void;
}

/**
 * Composer-adjacent run controls: Working / Stop sit above the input,
 * not in the top chrome. Connection problems share the same strip.
 */
export const StatusBar: React.FC<StatusBarProps> = ({
  connectionStatus,
  runStatus,
  activeRunId,
  onCancel,
}) => {
  const isRunning = runStatus === 'running' || runStatus === 'waiting' || runStatus === 'pending';
  const connectionIssue =
    connectionStatus === 'disconnected' ||
    connectionStatus === 'connecting' ||
    connectionStatus === 'reconnecting';

  if (!connectionIssue && !isRunning) {
    return null;
  }

  return (
    <div className="composer-run-bar" role="status">
      <div className="composer-run-bar-left">
        {connectionIssue && (
          <div className={`composer-run-conn connection-${connectionStatus}`} title="Agent connection">
            <span className="dot" />
            <span>
              {connectionStatus === 'disconnected'
                ? 'Disconnected'
                : connectionStatus === 'reconnecting'
                  ? 'Reconnecting…'
                  : 'Connecting…'}
            </span>
          </div>
        )}

        {isRunning && (
          <div className="composer-run-status">
            <ActivityIndicator state="live" />
            <span className="live-activity-shimmer">
              {runStatus === 'waiting' ? 'Waiting' : 'Working'}
            </span>
          </div>
        )}
      </div>

      {isRunning && activeRunId && (
        <button type="button" className="composer-run-stop" onClick={onCancel} title="Stop (Ctrl+Shift+X)">
          <StopIcon className="icon-sm" />
          <span>Stop</span>
        </button>
      )}
    </div>
  );
};
