import React, { useState } from 'react';
import { ProviderCredentials } from '../api/settings';

interface ProviderSettingsProps {
  activeProvider: string;
  providers: Record<string, ProviderCredentials>;
  onChangeProvider: (name: string) => void;
  onUpdateCredentials: (name: string, creds: ProviderCredentials) => void;
  onTestConnection: (name: string) => void;
}

export const ProviderSettings: React.FC<ProviderSettingsProps> = ({
  activeProvider,
  providers,
  onChangeProvider,
  onUpdateCredentials,
  onTestConnection,
}) => {
  const currentCreds = providers[activeProvider] || { storageMode: 'environment' };
  const [apiKey, setApiKey] = useState(currentCreds.apiKey || '');
  const [baseUrl, setBaseUrl] = useState(currentCreds.baseUrl || '');
  const [orgId, setOrgId] = useState(currentCreds.orgId || '');
  const [modelId, setModelId] = useState(currentCreds.modelId || '');
  const [storageMode, setStorageMode] = useState<'environment' | 'stored' | 'session'>(currentCreds.storageMode || 'environment');

  const handleSave = () => {
    onUpdateCredentials(activeProvider, {
      apiKey,
      baseUrl,
      orgId,
      modelId,
      storageMode,
    });
  };

  return (
    <div className="card provider-settings-card">
      <div className="card-header">
        <span className="card-title">LLM Provider Configuration</span>
      </div>

      <div className="provider-selector-radios">
        {['OpenAI', 'Google Gemini', 'Anthropic', 'OpenAI-compatible'].map((name) => (
          <label key={name} className={`radio-pill ${activeProvider === name ? 'active' : ''}`}>
            <input
              type="radio"
              name="provider"
              value={name}
              checked={activeProvider === name}
              onChange={() => onChangeProvider(name)}
            />
            <span>{name}</span>
          </label>
        ))}
      </div>

      <div className="form-group-list">
        <div className="form-row">
          <label>Storage Mode:</label>
          <select
            className="select-input"
            value={storageMode}
            onChange={(e) => setStorageMode(e.target.value as any)}
          >
            <option value="environment">Environment Variable (OPENAI_API_KEY, etc.)</option>
            <option value="stored">Stored Credential (Backend Encrypted)</option>
            <option value="session">Session Only (Disappears when server exits)</option>
          </select>
        </div>

        {storageMode !== 'environment' && (
          <div className="form-row">
            <label>API Key:</label>
            <input
              type="password"
              className="text-input"
              placeholder="sk-••••••••••••••••"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
            />
          </div>
        )}

        {(activeProvider === 'OpenAI' || activeProvider === 'Anthropic' || activeProvider === 'OpenAI-compatible') && (
          <div className="form-row">
            <label>Base URL:</label>
            <input
              type="text"
              className="text-input"
              placeholder={activeProvider === 'OpenAI-compatible' ? 'http://localhost:11434/v1' : 'Default endpoint'}
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
            />
          </div>
        )}

        {activeProvider === 'OpenAI' && (
          <div className="form-row">
            <label>Organization ID (Optional):</label>
            <input
              type="text"
              className="text-input"
              placeholder="org-••••••••"
              value={orgId}
              onChange={(e) => setOrgId(e.target.value)}
            />
          </div>
        )}

        {activeProvider === 'OpenAI-compatible' && (
          <div className="form-row">
            <label>Custom Model ID:</label>
            <input
              type="text"
              className="text-input"
              placeholder="llama3.2, mistral, etc."
              value={modelId}
              onChange={(e) => setModelId(e.target.value)}
            />
          </div>
        )}

        <div className="form-actions">
          <button className="btn btn-approve" onClick={handleSave}>
            Save Provider Credentials
          </button>
          <button className="btn btn-deny" onClick={() => onTestConnection(activeProvider)}>
            Test Connection
          </button>
        </div>
      </div>
    </div>
  );
};
