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
  const [activeTab, setActiveTab] = useState<'api' | 'providers' | 'models' | 'routing' | 'agent' | 'workspace'>('api');
  const [feedback, setFeedback] = useState<string | null>(null);

  useEffect(() => {
    const unsub = settingsStore.subscribe(() => {
      setSettings(settingsStore.getSettings());
    });
    settingsStore.fetchSettings();
    return () => unsub();
  }, []);

  const handleUpdate = (updates: Partial<SettingsConfig>) => {
    settingsStore.updateSettings(updates);
    setFeedback('Settings updated successfully');
    setTimeout(() => setFeedback(null), 3000);
  };

  const handleTestProvider = async (providerId: string) => {
    try {
      const res = await settingsStore.testProvider(providerId);
      setFeedback(res.message);
    } catch (err: any) {
      setFeedback(`Connection failed: ${err.message}`);
    }
  };

  const providerStatus = settingsStore.getProviderStatus(settings.activeProvider);
  const availableModels = settingsStore.getProviderModels(settings.activeProvider);

  return (
    <div className="settings-page-modal">
      <div className="settings-page-container">
        <div className="settings-page-header">
          <div className="title-area">
            <Cog6ToothIcon className="icon header-icon" />
            <span className="page-title">Agent Control Plane & Runtime Settings</span>
          </div>
          <button className="close-btn" onClick={onClose}>
            <XMarkIcon className="icon" />
          </button>
        </div>

        {feedback && <div className="settings-feedback-banner">{feedback}</div>}

        <div className="settings-tabs">
          <button className={`tab-btn ${activeTab === 'api' ? 'active' : ''}`} onClick={() => setActiveTab('api')}>
            Connections & API
          </button>
          <button className={`tab-btn ${activeTab === 'providers' ? 'active' : ''}`} onClick={() => setActiveTab('providers')}>
            LLM Providers
          </button>
          <button className={`tab-btn ${activeTab === 'models' ? 'active' : ''}`} onClick={() => setActiveTab('models')}>
            Model Selection
          </button>
          <button className={`tab-btn ${activeTab === 'routing' ? 'active' : ''}`} onClick={() => setActiveTab('routing')}>
            Model Routing
          </button>
          <button className={`tab-btn ${activeTab === 'agent' ? 'active' : ''}`} onClick={() => setActiveTab('agent')}>
            Agent & Orchestration
          </button>
          <button className={`tab-btn ${activeTab === 'workspace' ? 'active' : ''}`} onClick={() => setActiveTab('workspace')}>
            Workspace
          </button>
        </div>

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
              onChangeProvider={(name) => handleUpdate({ activeProvider: name })}
              onUpdateCredentials={(name, creds) => {
                const updated = { ...settings.providers, [name]: creds };
                handleUpdate({ providers: updated });
              }}
              onTestConnection={handleTestProvider}
            />
          )}

          {activeTab === 'models' && (
            <ModelSettings
              useSeparateModels={settings.useSeparateModels}
              thinkingModelId={settings.thinkingModelId}
              codingModelId={settings.codingModelId}
              fastModelId={settings.fastModelId}
              availableModels={availableModels}
              onChangeSeparate={(val) => handleUpdate({ useSeparateModels: val })}
              onChangeThinking={(id) => handleUpdate({ thinkingModelId: id })}
              onChangeCoding={(id) => handleUpdate({ codingModelId: id })}
              onChangeFast={(id) => handleUpdate({ fastModelId: id })}
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
        </div>
      </div>
    </div>
  );
};
