import React, { useEffect, useState } from 'react';
import { ConnectionStatus as WSStatus } from '../api/websocket';
import { agentHost, HostInfo } from '../host';
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
  const [hostInfo, setHostInfo] = useState<HostInfo | undefined>(undefined);

  useEffect(() => {
    agentHost.getHostInfo().then(setHostInfo).catch(() => setHostInfo(undefined));
  }, []);

  const hostLabel = hostInfo
    ? `${hostInfo.displayName || hostInfo.name}${hostInfo.apiVersion ? ` (API ${hostInfo.apiVersion})` : ''}`
    : 'Standalone';

  return (
    <div className="card connection-status-card">
      <div className="card-header">
        <span className="card-title">Connections</span>
      </div>

      <div className="connection-rows">
        <div className="conn-row">
          <span className="conn-name">IDE Host:</span>
          <span className="conn-badge status-connected">
            <span>{hostLabel}</span>
          </span>
        </div>

        <div className="conn-row">
          <span className="conn-name">Agent API Backend:</span>
          <span className={`conn-badge status-${agentStatus}`}>
            {agentStatus === 'connected' && <CheckCircleIcon className="icon icon-success" />}
            {agentStatus === 'connecting' && <ArrowPathIcon className="icon icon-spin" />}
            {agentStatus === 'disconnected' && <XCircleIcon className="icon icon-error" />}
            <span style={{ textTransform: 'capitalize' }}>{agentStatus}</span>
          </span>
        </div>

        <div className="conn-row">
          <span className="conn-name">LLM Provider ({providerName}):</span>
          <span className={`conn-badge status-${providerStatus}`}>
            {providerStatus === 'connected' && <CheckCircleIcon className="icon icon-success" />}
            {providerStatus === 'testing' && <ArrowPathIcon className="icon icon-spin" />}
            {providerStatus === 'disconnected' && <XCircleIcon className="icon icon-error" />}
            <span style={{ textTransform: 'capitalize' }}>{providerStatus}</span>
          </span>
          <button className="btn btn-sm btn-approve" onClick={onTestProvider} disabled={providerStatus === 'testing'}>
            Test
          </button>
        </div>
      </div>
    </div>
  );
};
