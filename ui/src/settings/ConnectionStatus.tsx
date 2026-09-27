import React from 'react';
import { ConnectionStatus as WSStatus } from '../api/websocket';
import { CheckCircleIcon, XCircleIcon, ArrowPathIcon } from '@heroicons/react/24/outline';

interface ConnectionStatusProps {
  agentStatus: WSStatus;
  providerName: string;
  providerStatus: 'connected' | 'disconnected' | 'testing';
  onTestProvider: () => void;
}

export const ConnectionStatus: React.FC<ConnectionStatusProps> = ({
  agentStatus,
  providerName,
  providerStatus,
  onTestProvider,
}) => {
  return (
    <div className="card connection-status-card">
      <div className="card-header">
        <span className="card-title">Connections</span>
      </div>

      <div className="connection-rows">
        <div className="conn-row">
          <span className="conn-name">Agent API Backend:</span>
          <span className={`conn-badge status-${agentStatus}`}>
            {agentStatus === 'connected' && <CheckCircleIcon className="icon icon-success" />}
            {agentStatus === 'connecting' && <ArrowPathIcon className="icon icon-spin" />}
            {agentStatus === 'disconnected' && <XCircleIcon className="icon icon-error" />}
            <span>{agentStatus.toUpperCase()}</span>
          </span>
        </div>

        <div className="conn-row">
          <span className="conn-name">LLM Provider ({providerName}):</span>
          <span className={`conn-badge status-${providerStatus}`}>
            {providerStatus === 'connected' && <CheckCircleIcon className="icon icon-success" />}
            {providerStatus === 'testing' && <ArrowPathIcon className="icon icon-spin" />}
            {providerStatus === 'disconnected' && <XCircleIcon className="icon icon-error" />}
            <span>{providerStatus.toUpperCase()}</span>
          </span>
          <button className="btn btn-sm btn-approve" onClick={onTestProvider} disabled={providerStatus === 'testing'}>
            Test Connection
          </button>
        </div>
      </div>
    </div>
  );
};
