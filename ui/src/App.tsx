import React, { useEffect, useState, useRef } from 'react';
import { useAgent } from './hooks/useAgent';
import { settingsStore } from './state/settingsStore';
import { queueStore } from './state/queueStore';
import { Chat } from './components/Chat';
import { ActivityFeedback } from './components/ActivityFeedback';
import { AgentTree } from './components/AgentTree';
import { AgentResult } from './components/AgentResult';
import { Attachments } from './components/Attachments';
import { Plan } from './components/Plan';
import { ToolCall } from './components/ToolCall';
import { FileChange } from './components/FileChange';
import { Terminal } from './components/Terminal';
import { TestResults } from './components/TestResults';
import { Approval } from './components/Approval';
import { PromptQueue } from './components/queue/PromptQueue';
import { SettingsPage } from './settings/SettingsPage';
import { CommandPalette, CommandOption } from './components/CommandPalette';
import {
  Cog6ToothIcon,
  PlusIcon,
  PaperClipIcon,
  FolderIcon,
  PaperAirplaneIcon,
} from '@heroicons/react/24/outline';
import './App.css';

export const App: React.FC = () => {
  const {
    status,
    activeRunStatus,
    activeRunId,
    messages,
    plan,
    toolExecutions,
    subagents,
    selectedSubagentId,
    fileChanges,
    attachments,
    terminalLogs,
    testResults,
    pendingApproval,
    startTask,
    cancelTask,
    respondApproval,
    configureBackend,
    addAttachment,
    removeAttachment,
    selectSubagent,
  } = useAgent();

  const [settings, setSettings] = useState(settingsStore.getSettings());
  const [showSettings, setShowSettings] = useState(false);
  const [showPalette, setShowPalette] = useState(false);
  const [composerInput, setComposerInput] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const win = window as any;
    const backendPort = win.__AGENT_PORT__ || 8080;
    const host = win.__AGENT_HOST__ || '127.0.0.1';
    configureBackend(`http://${host}:${backendPort}`, `ws://${host}:${backendPort}/ws`);

    const unsubSettings = settingsStore.subscribe(() => {
      setSettings(settingsStore.getSettings());
    });
    settingsStore.fetchSettings();

    // Keybindings: Cmd/Ctrl + K (Palette), Cmd/Ctrl + Enter (Submit)
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setShowPalette((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleGlobalKeyDown);

    return () => {
      unsubSettings();
      window.removeEventListener('keydown', handleGlobalKeyDown);
    };
  }, []);

  const handleComposerSubmit = () => {
    if (!composerInput.trim() || status === 'disconnected') return;
    const text = composerInput.trim();
    if (activeRunStatus === 'running' || activeRunStatus === 'waiting' || activeRunStatus === 'pending') {
      queueStore.addPrompt(text, settings.mode);
    } else {
      startTask(text);
    }
    setComposerInput('');
  };

  const handleKeyDownTextarea = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleComposerSubmit();
    }
  };

  const selectedSubagentObj = subagents.find((s) => s.id === selectedSubagentId) || null;
  const isWorking = activeRunStatus === 'running' || activeRunStatus === 'waiting' || activeRunStatus === 'pending';

  // Command palette entries
  const paletteCommands: CommandOption[] = [
    { id: 'new-task', label: 'Start new task', shortcut: 'Cmd+Enter', action: () => textareaRef.current?.focus() },
    { id: 'plan-mode', label: 'Switch to Plan Mode', action: () => settingsStore.updateSettings({ mode: 'plan' }) },
    { id: 'code-mode', label: 'Switch to Coding Mode', action: () => settingsStore.updateSettings({ mode: 'code' }) },
    { id: 'effort-low', label: 'Set Effort: Low', action: () => settingsStore.updateSettings({ effort: 'low' }) },
    { id: 'effort-medium', label: 'Set Effort: Medium', action: () => settingsStore.updateSettings({ effort: 'medium' }) },
    { id: 'effort-high', label: 'Set Effort: High', action: () => settingsStore.updateSettings({ effort: 'high' }) },
    { id: 'effort-extra', label: 'Set Effort: Extra High', action: () => settingsStore.updateSettings({ effort: 'extra_high' }) },
    { id: 'open-settings', label: 'Open Settings Control Plane', shortcut: 'Cmd+Shift+S', action: () => setShowSettings(true) },
    { id: 'cancel-task', label: 'Cancel Active Task', action: () => cancelTask() },
  ];

  return (
    <div className="app-layout">
      {/* Header (44px) */}
      <header className="top-control-bar">
        <div className="header-brand">
          <span className="brand-mark">✦ AGENT</span>
          <span className="workspace-name">blackjak</span>
        </div>

        <div className="header-status-area">
          <div className="status-indicator">
            <span className={`status-dot ${isWorking ? 'working' : status === 'disconnected' ? 'error' : ''}`} />
            <span>{isWorking ? 'WORKING' : status === 'disconnected' ? 'DISCONNECTED' : 'READY'}</span>
          </div>

          <button className="settings-trigger-btn" onClick={() => setShowSettings(true)} title="Settings (Cmd+Shift+S)">
            <Cog6ToothIcon className="icon" />
            <span>Settings</span>
          </button>
        </div>
      </header>

      {/* Main Three-Column Layout */}
      <main className="three-column-grid">
        {/* Left Column: Navigation & Control */}
        <aside className="column-left">
          <button className="nav-action-btn" onClick={() => textareaRef.current?.focus()}>
            <PlusIcon className="icon" />
            <span>New task</span>
          </button>

          {isWorking && (
            <div style={{ margin: '8px 0' }}>
              <div className="section-label">RUNNING</div>
              <div className="font-mono" style={{ padding: '0 16px', fontSize: '11.5px', color: 'var(--text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                └ {activeRunId || 'Active Task'}
              </div>
            </div>
          )}

          <AgentTree subagents={subagents} selectedId={selectedSubagentId} onSelect={selectSubagent} />

          <PromptQueue />

          <div style={{ marginTop: 'auto', padding: '8px' }}>
            <button className="nav-action-btn" onClick={() => setShowSettings(true)}>
              <Cog6ToothIcon className="icon" />
              <span>Settings</span>
            </button>
          </div>
        </aside>

        {/* Center Column: Conversation & Composer */}
        <section className="column-center">
          {(!settings.providers[settings.activeProvider]?.apiKey && settings.providers[settings.activeProvider]?.storageMode !== 'environment' && messages.length === 0) && (
            <div style={{ margin: '24px', padding: '24px', textAlign: 'center', backgroundColor: 'var(--bg-surface)', borderRadius: '6px', border: '1px solid var(--border)' }}>
              <h3 style={{ margin: 0, fontSize: '14px', color: 'var(--text-primary)', fontWeight: 600 }}>No AI provider configured</h3>
              <p style={{ color: 'var(--text-muted)', margin: '10px 0 16px 0', fontSize: '12px' }}>Connect a provider to start using the agent.</p>
              <button className="composer-send-btn" style={{ margin: '0 auto' }} onClick={() => setShowSettings(true)}>
                Configure Provider
              </button>
            </div>
          )}

          <Chat messages={messages} disabled={status === 'disconnected'} />

          <ActivityFeedback />

          <Attachments attachments={attachments} onAdd={addAttachment} onRemove={removeAttachment} />

          {/* Bottom Command Composer */}
          <div className="bottom-composer-wrapper">
            <textarea
              ref={textareaRef}
              className="composer-input-area"
              placeholder="Describe what you want the agent to do... (Cmd+K for options, Cmd+Enter to send)"
              value={composerInput}
              onChange={(e) => setComposerInput(e.target.value)}
              onKeyDown={handleKeyDownTextarea}
              disabled={status === 'disconnected'}
            />

            <div className="composer-toolbar">
              <div className="composer-actions-left">
                <button className="composer-btn" title="Attach file" onClick={() => {
                  const input = document.createElement('input');
                  input.type = 'file';
                  input.onchange = (e: any) => {
                    const file = e.target?.files?.[0];
                    if (file) addAttachment(file.name);
                  };
                  input.click();
                }}>
                  <PaperClipIcon className="icon-sm" />
                  <span>Attach</span>
                </button>

                <button className="composer-btn" title="Reference folder" onClick={() => setShowPalette(true)}>
                  <FolderIcon className="icon-sm" />
                  <span>Reference</span>
                </button>

                {/* Plan / Code Segmented Control */}
                <div className="segmented-control">
                  <button
                    className={`segmented-btn ${settings.mode === 'plan' ? 'active' : ''}`}
                    onClick={() => settingsStore.updateSettings({ mode: 'plan' })}
                  >
                    PLAN
                  </button>
                  <button
                    className={`segmented-btn ${settings.mode === 'code' ? 'active' : ''}`}
                    onClick={() => settingsStore.updateSettings({ mode: 'code' })}
                  >
                    CODE
                  </button>
                </div>

                {/* Compact Model Selector */}
                <select
                  className="select-input"
                  style={{ fontSize: '11px', padding: '2px 8px' }}
                  value={settings.codingModelId}
                  onChange={(e) => settingsStore.updateSettings({ codingModelId: e.target.value })}
                >
                  <option value="gpt-5.3-codex">GPT-5.3-Codex</option>
                  <option value="claude-opus-5">Claude Opus 5</option>
                  <option value="claude-sonnet-5">Claude Sonnet 5</option>
                  <option value="gemini-3.8-flash">Gemini 3.8 Flash</option>
                </select>

                {/* Compact Effort Selector */}
                <select
                  className="select-input"
                  style={{ fontSize: '11px', padding: '2px 8px', textTransform: 'uppercase' }}
                  value={settings.effort}
                  onChange={(e) => settingsStore.updateSettings({ effort: e.target.value as any })}
                >
                  <option value="low">LOW</option>
                  <option value="medium">MEDIUM</option>
                  <option value="high">HIGH</option>
                  <option value="extra_high">EXTRA HIGH</option>
                </select>
              </div>

              <button
                className="composer-send-btn"
                onClick={handleComposerSubmit}
                disabled={status === 'disconnected' || !composerInput.trim()}
              >
                <span>{isWorking ? 'Queue' : 'Send'}</span>
                <PaperAirplaneIcon className="icon-sm" />
              </button>
            </div>
          </div>
        </section>

        {/* Right Column: Context & Workspace Intelligence */}
        <aside className="column-right">
          <div className="section-label" style={{ paddingLeft: 0 }}>CONTEXT</div>

          <div className="font-mono" style={{ fontSize: '11.5px', margin: '8px 0', display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div>
              <span style={{ color: 'var(--text-muted)' }}>FILES: </span>
              <span>{attachments.length} attached</span>
            </div>

            <div>
              <span style={{ color: 'var(--text-muted)' }}>REFERENCES: </span>
              <span style={{ color: 'var(--text-secondary)' }}>@auth @middleware @tests</span>
            </div>

            <div>
              <span style={{ color: 'var(--text-muted)' }}>TOKENS: </span>
              <span>18.4k / 128k</span>
            </div>

            <div>
              <span style={{ color: 'var(--text-muted)' }}>MODE: </span>
              <span style={{ textTransform: 'uppercase' }}>{settings.mode}</span>
            </div>

            <div>
              <span style={{ color: 'var(--text-muted)' }}>MODEL: </span>
              <span>{settings.codingModelId}</span>
            </div>

            <div>
              <span style={{ color: 'var(--text-muted)' }}>EFFORT: </span>
              <span style={{ textTransform: 'uppercase' }}>{settings.effort}</span>
            </div>
          </div>

          <FileChange changes={fileChanges} />
          <Plan plan={plan} />
          <ToolCall executions={toolExecutions} />
          <TestResults results={testResults} />
          <Terminal logs={terminalLogs} />
        </aside>
      </main>

      <Approval request={pendingApproval} onRespond={respondApproval} />

      <AgentResult subagent={selectedSubagentObj} fileChanges={fileChanges} onClose={() => selectSubagent(null)} />

      {showSettings && (
        <SettingsPage agentStatus={status} workspacePath="" onClose={() => setShowSettings(false)} />
      )}

      <CommandPalette isOpen={showPalette} onClose={() => setShowPalette(false)} commands={paletteCommands} />
    </div>
  );
};

export default App;
