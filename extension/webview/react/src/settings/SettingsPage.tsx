import React, { useState, useEffect } from 'react';
import { settingsStore } from '../state/settingsStore';
import { SettingsConfig } from '../api/settings';
import { ConnectionStatus as WSStatus } from '../api/websocket';
import { ConnectionStatus } from './ConnectionStatus';
import { ProviderSettings } from './ProviderSettings';
import { ModelSettings } from './ModelSettings';
import { ModelRoutingSettings } from './ModelRoutingSettings';
import { AgentSettings } from './AgentSettings';
import { WorkspaceSettings } from './WorkspaceSettings';
import { XMarkIcon, Cog6ToothIcon } from '@heroicons/react/24/outline';

interface SettingsPageProps {
  agentStatus: WSStatus;
  workspacePath: string;
  onClose: () => void;
}

export const SettingsPage: React.FC<SettingsPageProps> = ({ agentStatus, workspacePath, onClose }) => {
  const [settings, setSettings] = useState<SettingsConfig>(settingsStore.getSettings());
  const [pending, setPending] = useState<Partial<SettingsConfig>>({});
  const [activeTab, setActiveTab] = useState<'api' | 'providers' | 'models' | 'routing' | 'agent' | 'workspace'>('api');
  const [feedback, setFeedback] = useState<string | null>(null);
  const dirty = Object.keys(pending).length > 0;

  useEffect(() => {
    const unsub = settingsStore.subscribe(() => {
      setSettings((prev) => ({ ...settingsStore.getSettings(), ...pending }));
    });
    settingsStore.fetchSettings();
    return () => unsub();
  }, [pending]);

  // Accumulate edits; only the fields the user actually changed are sent on Save.
  const handleUpdate = (updates: Partial<SettingsConfig>) => {
    setSettings((prev) => ({ ...prev, ...updates }));
    setPending((prev) => ({ ...prev, ...updates }));
  };

  const [saving, setSaving] = useState(false);
  const [saveMsg, setSaveMsg] = useState<{ ok: boolean; text: string } | null>(null);

  const handleSave = async () => {
    setSaving(true);
    setSaveMsg(null);
    try {
      await settingsStore.updateSettings(pending);
      setPending({});
      setSaveMsg({ ok: true, text: 'Saved' });
    } catch {
      setSaveMsg({ ok: false, text: 'Failed to save' });
    } finally {
      setSaving(false);
      setTimeout(() => setSaveMsg(null), 2500);
    }
  };

  // Providers tab: credential save + provider selection persist immediately.
  const handleSaveCredentials = async (name: string, creds: typeof settings.providers[string]) => {
    await settingsStore.updateSettings({ providers: { ...settings.providers, [name]: creds } });
  };

  const handleSelectProvider = async (name: string) => {
    handleUpdate({ activeProvider: name });
    try {
      await settingsStore.updateSettings({ activeProvider: name });
      setPending((prev) => {
        const next = { ...prev };
        delete next.activeProvider;
        return next;
      });
    } catch {
      setFeedback('Failed to save provider selection');
    }
  };

  const handleTestProvider = async (providerId: string) => {
    try {
      const res = await settingsStore.testProvider(providerId);
      setFeedback(res.message);
    } catch (err: any) {
      setFeedback(`Connection failed: ${err.message}`);
    }
  };

  // Editable tabs always show a save row — one status slot: save result, or a hint when clean.
  const saveBar = (activeTab === 'models' || activeTab === 'routing' || activeTab === 'agent') && (
    <div className="settings-save-row">
      <span className={`save-status ${saveMsg ? (saveMsg.ok ? 'ok' : 'err') : ''}`}>
        {saveMsg ? saveMsg.text : dirty ? '' : 'No changes'}
      </span>
      <button className="btn btn-approve" onClick={handleSave} disabled={saving || !dirty}>
        {saving ? 'Saving…' : 'Save changes'}
      </button>
    </div>
  );

  const providerStatus = settingsStore.getProviderStatus(settings.activeProvider);
  const availableModels = settingsStore.getProviderModels(settings.activeProvider);

  return (
    <div className="settings-page-modal">
      <div className="settings-page-container">
        <div className="settings-page-header">
          <div className="title-area">
            <Cog6ToothIcon className="icon header-icon" />
            <span className="page-title">Settings</span>
          </div>
          <button className="close-btn" onClick={onClose}>
            <XMarkIcon className="icon" />
          </button>
        </div>

        {feedback && <div className="settings-feedback-banner">{feedback}</div>}

        <div className="settings-body">
          <nav className="settings-nav">
            {([
              ['api', 'Connection'],
              ['providers', 'Providers'],
              ['models', 'Models'],
              ['routing', 'Routing'],
              ['agent', 'Agent'],
              ['workspace', 'Workspace'],
            ] as const).map(([id, label]) => (
              <button
                key={id}
                className={`settings-nav-item ${activeTab === id ? 'active' : ''}`}
                onClick={() => setActiveTab(id)}
              >
                {label}
              </button>
            ))}
          </nav>

          <div className="settings-tab-content">
          {activeTab === 'api' && (
            <ConnectionStatus
              agentStatus={agentStatus}
              providerName={settings.activeProvider}
              providerStatus={providerStatus}
              onTestProvider={() => handleTestProvider(settings.activeProvider)}
            />
          )}

          {activeTab === 'providers' && (
            <ProviderSettings
              activeProvider={settings.activeProvider}
              providers={settings.providers}
              onChangeProvider={handleSelectProvider}
              onSaveCredentials={handleSaveCredentials}
              onTestConnection={handleTestProvider}
            />
          )}

          {activeTab === 'models' && (
            <ModelSettings
              useSeparateModels={settings.useSeparateModels}
              thinking={settings.thinking || { providerId: 'Anthropic', modelId: settings.thinkingModelId || 'claude-opus-5', role: 'thinking' }}
              coding={settings.coding || { providerId: 'OpenAI', modelId: settings.codingModelId || 'gpt-5.3-codex', role: 'coding' }}
              fast={settings.fast || { providerId: 'Google Gemini', modelId: settings.fastModelId || 'gemini-3.5-flash-lite', role: 'fast' }}
              review={settings.review || { providerId: 'Anthropic', modelId: settings.reviewModelId || 'claude-sonnet-5', role: 'review' }}
              effort={settings.effort}
              availableModelsMap={settingsStore.getAllProviderModels()}
              onChangeSeparate={(val) => handleUpdate({ useSeparateModels: val })}
              onChangeModelConfig={(role, cfg) => {
                if (role === 'thinking') handleUpdate({ thinking: cfg, thinkingModelId: cfg.modelId });
                else if (role === 'coding') handleUpdate({ coding: cfg, codingModelId: cfg.modelId });
                else if (role === 'fast') handleUpdate({ fast: cfg, fastModelId: cfg.modelId });
                else if (role === 'review') handleUpdate({ review: cfg, reviewModelId: cfg.modelId });
              }}
              onChangeEffort={(effort) => handleUpdate({ effort })}
            />
          )}

          {activeTab === 'routing' && (
            <ModelRoutingSettings
              routes={settings.modelRoutes}
              onChangeRoute={(task, role) => {
                const updated = { ...settings.modelRoutes, [task]: role };
                handleUpdate({ modelRoutes: updated });
              }}
              onResetDefaults={() => {
                handleUpdate({
                  modelRoutes: {
                    planning: 'thinking',
                    exploration: 'fast',
                    coding: 'coding',
                    debugging: 'thinking',
                    testing: 'coding',
                    review: 'thinking',
                    summarization: 'fast',
                  },
                });
              }}
            />
          )}

          {activeTab === 'agent' && (
            <AgentSettings settings={settings} onUpdate={handleUpdate} />
          )}

          {activeTab === 'workspace' && (
            <WorkspaceSettings workspacePath={workspacePath} />
          )}

          {saveBar}
          </div>
        </div>
      </div>
    </div>
  );
};
