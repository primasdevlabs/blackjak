import React from 'react';
import { ConnectionStatus } from '../api/websocket';
import { RunStatus } from '../types/events';
import { CpuChipIcon, XMarkIcon } from '@heroicons/react/24/outline';

interface StatusBarProps {
  connectionStatus: ConnectionStatus;
  runStatus: RunStatus | null;
  activeRunId: string | null;
  onCancel: () => void;
}

export const StatusBar: React.FC<StatusBarProps> = ({
  connectionStatus,
  runStatus,
  activeRunId,
  onCancel,
}) => {
  const isRunning = runStatus === 'running' || runStatus === 'waiting' || runStatus === 'pending';

  return (
    <header className="status-bar">
      <div className="status-title">
        <CpuChipIcon className="icon title-icon" />
        <span className="title-text">AI Agent</span>
      </div>

      <div className="status-badges">
        <div className={`badge connection-${connectionStatus}`}>
          <span className="dot"></span>
          <span>{connectionStatus}</span>
        </div>

        {runStatus && (
          <div className={`badge run-${runStatus}`}>
            <span>{runStatus.toUpperCase()}</span>
          </div>
        )}

        {isRunning && activeRunId && (
          <button className="cancel-btn" onClick={onCancel}>
            <XMarkIcon className="icon btn-icon" />
            <span>Stop</span>
          </button>
        )}
      </div>
    </header>
  );
};
