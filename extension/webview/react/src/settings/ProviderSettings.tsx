import React, { useState, useEffect } from 'react';
import { ProviderCredentials } from '../api/settings';

interface ProviderSettingsProps {
  activeProvider: string;
  providers: Record<string, ProviderCredentials>;
  onChangeProvider: (name: string) => void;
  onSaveCredentials: (name: string, creds: ProviderCredentials) => Promise<void>;
  onTestConnection: (name: string) => void;
}

export const ProviderSettings: React.FC<ProviderSettingsProps> = ({
  activeProvider,
  providers,
  onChangeProvider,
  onSaveCredentials,
  onTestConnection,
}) => {
  const currentCreds = providers[activeProvider] || { storageMode: 'environment' };
  const [apiKey, setApiKey] = useState(currentCreds.apiKey || '');
  const [baseUrl, setBaseUrl] = useState(currentCreds.baseUrl || '');
  const [orgId, setOrgId] = useState(currentCreds.orgId || '');
  const [modelId, setModelId] = useState(currentCreds.modelId || '');
  const [storageMode, setStorageMode] = useState<'environment' | 'stored' | 'session'>(currentCreds.storageMode || 'environment');
  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState<{ ok: boolean; text: string } | null>(null);

  // Re-sync the form when switching providers so one provider's key/URL
  // never bleeds into another provider's credentials.
  useEffect(() => {
    const creds = providers[activeProvider] || { storageMode: 'environment' as const };
    setApiKey(creds.apiKey || '');
    setBaseUrl(creds.baseUrl || '');
    setOrgId(creds.orgId || '');
    setModelId(creds.modelId || '');
    setStorageMode(creds.storageMode || 'environment');
  }, [activeProvider]);

  const handleSave = async () => {
    setSaving(true);
    setSaveMsg(null);
    try {
      await onSaveCredentials(activeProvider, { apiKey, baseUrl, orgId, modelId, storageMode });
      setSaveMsg({ ok: true, text: 'Saved' });
    } catch (err: any) {
      setSaveMsg({ ok: false, text: err?.message || 'Failed to save' });
    } finally {
      setSaving(false);
      setTimeout(() => setSaveMsg(null), 2500);
    }
  };

  return (
    <div className="card provider-settings-card">
      <div className="card-header">
        <span className="card-title">Providers</span>
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
          <label>Storage</label>
          <select
            className="select-input"
            value={storageMode}
            onChange={(e) => setStorageMode(e.target.value as any)}
          >
            <option value="environment">Environment variable</option>
            <option value="stored">Stored in backend</option>
            <option value="session">Session only</option>
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
              placeholder={activeProvider === 'OpenAI-compatible' ? 'https://api.example.com/v1' : 'Default endpoint'}
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
            />
            {activeProvider === 'OpenAI-compatible' && (
              <p className="field-hint">
                Use the API root ending in <code>/v1</code>, not <code>/chat/completions</code>.
              </p>
            )}
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
            <label>Model ID</label>
            <input
              type="text"
              className="text-input"
              placeholder="notrack-uncensored, llama3.2, …"
              value={modelId}
              onChange={(e) => setModelId(e.target.value)}
            />
          </div>
        )}

        <div className="form-actions settings-action-row">
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? 'Saving…' : 'Save'}
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => onTestConnection(activeProvider)}
            disabled={saving}
          >
            Test connection
          </button>
          {saveMsg && (
            <span className={`save-status ${saveMsg.ok ? 'ok' : 'err'}`}>{saveMsg.text}</span>
          )}
        </div>
      </div>
    </div>
  );
};
