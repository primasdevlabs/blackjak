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
import {
  GeneralSettings,
  UsageSettings,
  RulesSettings,
  SkillsSettings,
  IndexingSettings,
  CustomizeSettings,
  GitPRSettings,
  BrowserNetworkSettings,
  TabSettingsPanel,
  BetaSettings,
} from './ParitySettings';
import { XMarkIcon, Cog6ToothIcon } from '@heroicons/react/24/outline';

interface SettingsPageProps {
  agentStatus: WSStatus;
  workspacePath: string;
  onClose: () => void;
}

type TabId =
  | 'general'
  | 'usage'
  | 'agents'
  | 'models'
  | 'providers'
  | 'routing'
  | 'rules'
  | 'skills'
  | 'git'
  | 'customize'
  | 'network'
  | 'tab'
  | 'indexing'
  | 'beta'
  | 'connection'
  | 'workspace';

const NAV: { id: TabId; label: string }[] = [
  { id: 'general', label: 'General' },
  { id: 'usage', label: 'Plan & Usage' },
  { id: 'agents', label: 'Agents' },
  { id: 'models', label: 'Models' },
  { id: 'providers', label: 'Providers / BYOK' },
  { id: 'routing', label: 'Routing' },
  { id: 'rules', label: 'Rules & Policies' },
  { id: 'skills', label: 'Skills' },
  { id: 'git', label: 'Git & PRs' },
  { id: 'customize', label: 'Customize' },
  { id: 'network', label: 'Browser & Network' },
  { id: 'tab', label: 'Tab' },
  { id: 'indexing', label: 'Indexing' },
  { id: 'beta', label: 'Beta' },
  { id: 'connection', label: 'Connection' },
  { id: 'workspace', label: 'Workspace' },
];

export const SettingsPage: React.FC<SettingsPageProps> = ({ agentStatus, workspacePath, onClose }) => {
  const [settings, setSettings] = useState<SettingsConfig>(settingsStore.getSettings());
  const [pending, setPending] = useState<Partial<SettingsConfig>>({});
  const [activeTab, setActiveTab] = useState<TabId>('general');
  const [feedback, setFeedback] = useState<string | null>(null);
  const dirty = Object.keys(pending).length > 0;

  useEffect(() => {
    const unsub = settingsStore.subscribe(() => {
      setSettings({ ...settingsStore.getSettings(), ...pending });
    });
    settingsStore.fetchSettings();
    return () => unsub();
  }, [pending]);

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

  const saveable = ['general', 'models', 'routing', 'agents', 'customize', 'git', 'network', 'tab', 'beta'].includes(activeTab);
  const saveBar = saveable ? (
    <div className="settings-save-row">
      <span className={`save-status ${saveMsg ? (saveMsg.ok ? 'ok' : 'err') : ''}`}>
        {saveMsg ? saveMsg.text : dirty ? 'Unsaved changes' : 'No changes'}
      </span>
      <button
        type="button"
        className="btn btn-primary"
        onClick={handleSave}
        disabled={saving || !dirty}
      >
        {saving ? 'Saving…' : 'Save changes'}
      </button>
    </div>
  ) : null;

  const providerStatus = settingsStore.getProviderStatus(settings.activeProvider);

  return (
    <div className="settings-page-modal">
      <div className="settings-page-container settings-shell">
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
          <nav className="settings-nav settings-nav-wide">
            {NAV.map(({ id, label }) => (
              <button
                key={id}
                className={`settings-nav-item ${activeTab === id ? 'active' : ''}`}
                onClick={() => setActiveTab(id)}
              >
                {label}
              </button>
            ))}
          </nav>

          <div className="settings-main">
            <div className="settings-tab-content">
              {activeTab === 'general' && <GeneralSettings settings={settings} onUpdate={handleUpdate} />}
              {activeTab === 'usage' && <UsageSettings />}
              {activeTab === 'agents' && <AgentSettings settings={settings} onUpdate={handleUpdate} />}
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
              {activeTab === 'providers' && (
                <ProviderSettings
                  activeProvider={settings.activeProvider}
                  providers={settings.providers}
                  onChangeProvider={handleSelectProvider}
                  onSaveCredentials={handleSaveCredentials}
                  onTestConnection={handleTestProvider}
                />
              )}
              {activeTab === 'routing' && (
                <ModelRoutingSettings
                  routes={settings.modelRoutes}
                  onChangeRoute={(task, role) => {
                    handleUpdate({ modelRoutes: { ...settings.modelRoutes, [task]: role } });
                  }}
                  onResetDefaults={() => {
                    handleUpdate({
                      modelRoutes: {
                        planning: 'thinking',
                        exploration: 'fast',
                        coding: 'coding',
                        debugging: 'thinking',
                        testing: 'coding',
                        review: 'review',
                        summarization: 'fast',
                      },
                    });
                  }}
                />
              )}
              {activeTab === 'rules' && <RulesSettings />}
              {activeTab === 'skills' && <SkillsSettings />}
              {activeTab === 'git' && <GitPRSettings settings={settings} onUpdate={handleUpdate} />}
              {activeTab === 'customize' && <CustomizeSettings settings={settings} onUpdate={handleUpdate} />}
              {activeTab === 'network' && <BrowserNetworkSettings settings={settings} onUpdate={handleUpdate} />}
              {activeTab === 'tab' && <TabSettingsPanel settings={settings} onUpdate={handleUpdate} />}
              {activeTab === 'indexing' && <IndexingSettings />}
              {activeTab === 'beta' && <BetaSettings settings={settings} onUpdate={handleUpdate} />}
              {activeTab === 'connection' && (
                <ConnectionStatus
                  agentStatus={agentStatus}
                  providerName={settings.activeProvider}
                  providerStatus={providerStatus}
                  onTestProvider={() => handleTestProvider(settings.activeProvider)}
                />
              )}
              {activeTab === 'workspace' && <WorkspaceSettings workspacePath={workspacePath} />}
            </div>
            {saveBar}
          </div>
        </div>
      </div>
    </div>
  );
};
