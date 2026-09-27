import React, { useEffect, useState } from 'react';
import { useAgent } from './hooks/useAgent';
import { settingsStore } from './state/settingsStore';
import { queueStore } from './state/queueStore';
import { StatusBar } from './components/StatusBar';
import { Chat } from './components/Chat';
import { ActivityFeedback } from './components/ActivityFeedback';
import { AgentTree } from './components/AgentTree';
import { AgentStatus } from './components/AgentStatus';
import { AgentTimeline } from './components/AgentTimeline';
import { AgentDelegation } from './components/AgentDelegation';
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
import { Cog6ToothIcon, CpuChipIcon, BoltIcon } from '@heroicons/react/24/outline';
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

  useEffect(() => {
    const win = window as any;
    const backendPort = win.__AGENT_PORT__ || 8080;
    const host = win.__AGENT_HOST__ || '127.0.0.1';
    configureBackend(`http://${host}:${backendPort}`, `ws://${host}:${backendPort}/ws`);

    const unsubSettings = settingsStore.subscribe(() => {
      setSettings(settingsStore.getSettings());
    });
    settingsStore.fetchSettings();

    return () => unsubSettings();
  }, []);

  const handlePromptSubmit = (promptText: string) => {
    if (activeRunStatus === 'running' || activeRunStatus === 'waiting' || activeRunStatus === 'pending') {
      // Agent is busy: add prompt to queue
      queueStore.addPrompt(promptText, settings.mode);
    } else {
      startTask(promptText);
    }
  };

  const selectedSubagentObj = subagents.find((s) => s.id === selectedSubagentId) || null;
  const isBusy = activeRunStatus === 'running' || activeRunStatus === 'waiting' || activeRunStatus === 'pending';

  return (
    <div className="app-layout">
      <header className="top-control-bar">
        <StatusBar
          connectionStatus={status}
          runStatus={activeRunStatus}
          activeRunId={activeRunId}
          onCancel={cancelTask}
        />
        <button className="settings-trigger-btn" onClick={() => setShowSettings(true)} title="Open Settings Control Plane">
          <Cog6ToothIcon className="icon" />
          <span>Settings</span>
        </button>
      </header>

      <AgentStatus subagents={subagents} />

      <main className="three-column-grid">
        {/* Left Column: Agent Tree, Subagents & Prompt Queue */}
        <aside className="column-left">
          <AgentTree
            subagents={subagents}
            selectedId={selectedSubagentId}
            onSelect={selectSubagent}
          />
          <PromptQueue />
          <AgentDelegation subagents={subagents} />
          <AgentTimeline subagents={subagents} />
        </aside>

        {/* Center Column: Chat Conversation, Activity Feedback & Attachments */}
        <section className="column-center">
          <Chat
            messages={messages}
            onSubmit={handlePromptSubmit}
            disabled={status === 'disconnected'}
          />

          <ActivityFeedback />

          <Attachments attachments={attachments} onAdd={addAttachment} onRemove={removeAttachment} />
        </section>

        {/* Right Column: Context & Workspace Intelligence */}
        <aside className="column-right">
          <Plan plan={plan} />
          <ToolCall executions={toolExecutions} />
          <FileChange changes={fileChanges} />
          <TestResults results={testResults} />
          <Terminal logs={terminalLogs} />
        </aside>
      </main>

      {/* Bottom Control Plane Bar */}
      <footer className="bottom-control-plane">
        <div className="mode-toggle-group">
          <button
            className={`mode-btn ${settings.mode === 'plan' ? 'active' : ''}`}
            onClick={() => settingsStore.updateSettings({ mode: 'plan' })}
          >
            PLAN MODE
          </button>
          <button
            className={`mode-btn ${settings.mode === 'code' ? 'active' : ''}`}
            onClick={() => settingsStore.updateSettings({ mode: 'code' })}
          >
            CODING MODE
          </button>
        </div>

        <div className="effort-selector-group">
          <BoltIcon className="icon" />
          <span className="effort-label">Effort:</span>
          <select
            className="select-input effort-select"
            value={settings.effort}
            onChange={(e) => settingsStore.updateSettings({ effort: e.target.value as any })}
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="extra_high">Extra High</option>
          </select>
        </div>

        <div className="provider-models-badge">
          <CpuChipIcon className="icon" />
          <span>{settings.activeProvider} ({settings.useSeparateModels ? `${settings.thinkingModelId} / ${settings.codingModelId}` : settings.codingModelId})</span>
        </div>

        {isBusy && <span className="busy-queue-hint">[Prompt Queueing Enabled]</span>}
      </footer>

      <Approval request={pendingApproval} onRespond={respondApproval} />

      <AgentResult
        subagent={selectedSubagentObj}
        fileChanges={fileChanges}
        onClose={() => selectSubagent(null)}
      />

      {showSettings && (
        <SettingsPage
          agentStatus={status}
          workspacePath={""}
          onClose={() => setShowSettings(false)}
        />
      )}
    </div>
  );
};

export default App;
