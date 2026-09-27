import React, { useEffect, useState, useRef } from 'react';
import { useAgent } from './hooks/useAgent';
import { settingsStore } from './state/settingsStore';
import { queueStore } from './state/queueStore';
import { Chat } from './components/Chat';
import { ActivityFeedback } from './components/ActivityFeedback';
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
import { SlashCommandMenu, CommandItem } from './components/SlashCommandMenu';
import { CompactionModal, CompactionData } from './components/CompactionModal';
import { ContextStatusModal, ContextComposition } from './components/ContextStatusModal';
import { ClearContextModal } from './components/ClearContextModal';
import { TaskSummaryCard, TaskSummaryData } from './components/TaskSummaryCard';
import {
  Cog6ToothIcon,
  EllipsisHorizontalIcon,
  PlusIcon,
  PaperClipIcon,
  StopIcon,
  ArrowUpRightIcon,
  XMarkIcon,
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
  const [showOverflowMenu, setShowOverflowMenu] = useState(false);
  const [showAgentsOverlay, setShowAgentsOverlay] = useState(false);
  const [showModePicker, setShowModePicker] = useState(false);
  const [showEffortPicker, setShowEffortPicker] = useState(false);
  const [composerInput, setComposerInput] = useState('');
  const [containerWidth, setContainerWidth] = useState<number>(window.innerWidth);

  const [showSlashMenu, setShowSlashMenu] = useState(false);
  const [showCompactionModal, setShowCompactionModal] = useState(false);
  const [compactionData, setCompactionData] = useState<CompactionData | null>(null);
  const [showContextModal, setShowContextModal] = useState(false);
  const [contextComposition, setContextComposition] = useState<ContextComposition>({
    system: 4200,
    task: 1100,
    conversation: 3700,
    files: 6400,
    toolResults: 2100,
    agentSummaries: 1300,
    total: 18800,
    limit: 128000,
    pressure: 15,
  });
  const [showClearModal, setShowClearModal] = useState(false);
  const [taskSummary, setTaskSummary] = useState<TaskSummaryData | null>(null);
  const [compactionEventNotice, setCompactionEventNotice] = useState<{ before: string; after: string } | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
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

    const handleResize = () => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.clientWidth);
      } else {
        setContainerWidth(window.innerWidth);
      }
    };
    window.addEventListener('resize', handleResize);
    handleResize();

    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setShowPalette((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleGlobalKeyDown);

    return () => {
      unsubSettings();
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('keydown', handleGlobalKeyDown);
    };
  }, []);

  const executeSlashCommand = async (commandName: string) => {
    const cleanCmd = commandName.replace('/', '').toLowerCase().trim();

    if (cleanCmd === 'compact') {
      try {
        const win = window as any;
        const port = win.__AGENT_PORT__ || 8080;
        const res = await fetch(`http://127.0.0.1:${port}/api/commands/execute`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ command: 'compact' }),
        });
        const json = await res.json();
        if (json.success && json.data) {
          setCompactionData(json.data);
          setCompactionEventNotice({
            before: json.data.tokensBefore ? `${(json.data.tokensBefore / 1000).toFixed(1)}k` : '42.1k',
            after: json.data.tokensAfter ? `${(json.data.tokensAfter / 1000).toFixed(1)}k` : '11.7k',
          });
        }
      } catch (e) {
        setCompactionData({
          tokensBefore: 42100,
          tokensAfter: 11700,
          tokensSaved: 30400,
          preserved: ['Task objective', 'Implementation decisions', 'Modified files state', 'Test results', 'Subagent findings'],
          removed: ['12 tool outputs', '8 duplicate search results', '4 superseded plans'],
        });
        setCompactionEventNotice({ before: '42.1k', after: '11.7k' });
      }
    } else if (cleanCmd === 'context') {
      try {
        const win = window as any;
        const port = win.__AGENT_PORT__ || 8080;
        const res = await fetch(`http://127.0.0.1:${port}/api/commands/execute`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ command: 'context' }),
        });
        const json = await res.json();
        if (json.success && json.data) {
          setContextComposition(json.data);
        }
      } catch (e) {
        // Fallback default
      }
      setShowContextModal(true);
    } else if (cleanCmd === 'clear') {
      setShowClearModal(true);
    } else if (cleanCmd === 'summarize') {
      try {
        const win = window as any;
        const port = win.__AGENT_PORT__ || 8080;
        const res = await fetch(`http://127.0.0.1:${port}/api/commands/execute`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ command: 'summarize' }),
        });
        const json = await res.json();
        if (json.success && json.data) {
          setTaskSummary(json.data);
        }
      } catch (e) {
        setTaskSummary({
          objective: 'Fix token expiration handling',
          workspace: 'internal/auth/',
          completed: ['Located token validation', 'Updated expiration comparison', 'Added middleware handling'],
          remaining: ['Add regression tests', 'Run auth test suite'],
          files: [
            { path: 'internal/auth/token.go', status: 'M' },
            { path: 'internal/auth/middleware.go', status: 'M' },
            { path: 'internal/auth/token_test.go', status: 'A' },
          ],
        });
      }
    } else if (cleanCmd === 'plan') {
      settingsStore.updateSettings({ mode: 'plan' });
    } else if (cleanCmd === 'code') {
      settingsStore.updateSettings({ mode: 'code' });
    } else if (cleanCmd === 'model') {
      setShowSettings(true);
    } else if (cleanCmd === 'effort') {
      setShowEffortPicker(true);
    } else if (cleanCmd === 'agents') {
      setShowAgentsOverlay(true);
    } else if (cleanCmd === 'stop') {
      cancelTask();
    }
  };

  const handleComposerInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value;
    setComposerInput(val);
    if (val.startsWith('/')) {
      setShowSlashMenu(true);
    } else {
      setShowSlashMenu(false);
    }
  };

  const handleComposerSubmit = () => {
    if (!composerInput.trim() || status === 'disconnected') return;
    const text = composerInput.trim();
    if (text.startsWith('/')) {
      executeSlashCommand(text);
      setComposerInput('');
      setShowSlashMenu(false);
      return;
    }
    if (activeRunStatus === 'running' || activeRunStatus === 'waiting' || activeRunStatus === 'pending') {
      queueStore.addPrompt(text, settings.mode);
    } else {
      startTask(text);
    }
    setComposerInput('');
    setShowSlashMenu(false);
  };

  const handleKeyDownTextarea = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleComposerSubmit();
    }
  };

  const selectedSubagentObj = subagents.find((s) => s.id === selectedSubagentId) || null;
  const isWorking = activeRunStatus === 'running' || activeRunStatus === 'waiting' || activeRunStatus === 'pending';
  const isExpandedView = containerWidth >= 500;

  const paletteCommands: CommandOption[] = [
    { id: 'new-task', label: 'Start new task', shortcut: 'Enter', action: () => textareaRef.current?.focus() },
    { id: 'plan-mode', label: 'Switch to Plan Mode', action: () => settingsStore.updateSettings({ mode: 'plan' }) },
    { id: 'code-mode', label: 'Switch to Coding Mode', action: () => settingsStore.updateSettings({ mode: 'code' }) },
    { id: 'effort-low', label: 'Set Effort: Low', action: () => settingsStore.updateSettings({ effort: 'low' }) },
    { id: 'effort-medium', label: 'Set Effort: Medium', action: () => settingsStore.updateSettings({ effort: 'medium' }) },
    { id: 'effort-high', label: 'Set Effort: High', action: () => settingsStore.updateSettings({ effort: 'high' }) },
    { id: 'effort-extra', label: 'Set Effort: Extra High', action: () => settingsStore.updateSettings({ effort: 'extra_high' }) },
    { id: 'open-settings', label: 'Open Settings Control Plane', action: () => setShowSettings(true) },
    { id: 'cancel-task', label: 'Cancel Active Task', action: () => cancelTask() },
  ];

  return (
    <div ref={containerRef} className="app-sidebar-container">
      {/* Header (~40px) */}
      <header className="sidebar-header">
        <div className="header-brand">
          <span>✦ Agent</span>
        </div>

        <div className="header-status-area">
          <div className="status-indicator">
            <span className={`status-dot ${isWorking ? 'working' : status === 'disconnected' ? 'error' : ''}`} />
            <span>{isWorking ? 'Working' : status === 'disconnected' ? 'Disconnected' : 'Ready'}</span>
          </div>

          {subagents.length > 0 && (
            <button className="chip-btn" onClick={() => setShowAgentsOverlay(true)}>
              <span>✦ {subagents.length} agents</span>
            </button>
          )}

          <button className="icon-btn" onClick={() => setShowOverflowMenu(!showOverflowMenu)} title="Menu (⋯)">
            <EllipsisHorizontalIcon className="icon" />
          </button>

          <button className="icon-btn" onClick={() => setShowSettings(true)} title="Settings (⚙)">
            <Cog6ToothIcon className="icon" />
          </button>
        </div>
      </header>

      {/* Responsive Body */}
      <div className={isExpandedView ? 'expanded-grid' : ''} style={{ flex: 1, overflow: 'hidden', display: isExpandedView ? 'grid' : 'flex', flexDirection: 'column' }}>
        {/* Conversation Surface + Inline Queue & Tool Calls */}
        <section className="conversation-surface">
          {(!settings.providers[settings.activeProvider]?.apiKey && settings.providers[settings.activeProvider]?.storageMode !== 'environment' && messages.length === 0) && (
            <div style={{ padding: '20px', textAlign: 'center', backgroundColor: 'var(--bg-surface)', borderRadius: '6px', border: '1px solid var(--border)', margin: '12px 0' }}>
              <div style={{ fontWeight: 600, fontSize: '13px', color: 'var(--text-primary)' }}>No AI provider configured</div>
              <div style={{ color: 'var(--text-muted)', fontSize: '12px', margin: '8px 0 14px 0' }}>Connect a provider to start using the agent.</div>
              <button className="send-btn" style={{ margin: '0 auto' }} onClick={() => setShowSettings(true)}>
                Configure Provider
              </button>
            </div>
          )}

          <Chat messages={messages} disabled={status === 'disconnected'} />

          {compactionEventNotice && (
            <div
              className="activity-transient"
              style={{
                cursor: 'pointer',
                padding: '6px 10px',
                backgroundColor: 'var(--bg-surface)',
                border: '1px solid var(--border)',
                borderRadius: '4px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
              onClick={() => setShowCompactionModal(true)}
              title="Click for compaction details"
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <span>✦</span>
                <span style={{ fontSize: '12px', fontWeight: 500, color: 'var(--text-primary)' }}>
                  Context compacted · {compactionEventNotice.before} → {compactionEventNotice.after}
                </span>
              </div>
              <span style={{ fontSize: '11px', color: 'var(--text-muted)' }}>Details ↗</span>
            </div>
          )}

          {taskSummary && <TaskSummaryCard summary={taskSummary} />}

          <ActivityFeedback />

          <PromptQueue />

          <Attachments attachments={attachments} onAdd={addAttachment} onRemove={removeAttachment} />
        </section>

        {/* Expanded View Details Panel (when width >= 500px) */}
        {isExpandedView && (
          <aside className="details-panel">
            <div style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text-muted)', letterSpacing: '0.06em' }}>CONTEXT & DETAILS</div>
            <FileChange changes={fileChanges} />
            <Plan plan={plan} />
            <ToolCall executions={toolExecutions} />
            <TestResults results={testResults} />
            <Terminal logs={terminalLogs} />
          </aside>
        )}
      </div>

      {/* Bottom Compact Composer */}
      <div className="composer-container" style={{ position: 'relative' }}>
        {showSlashMenu && (
          <SlashCommandMenu
            query={composerInput}
            onSelect={(cmd) => {
              executeSlashCommand(cmd.name);
              setComposerInput('');
              setShowSlashMenu(false);
            }}
            onClose={() => setShowSlashMenu(false)}
          />
        )}

        <div className="composer-toolbar">
          <div className="composer-actions">
            <button className="icon-btn" title="Add task instruction" onClick={() => textareaRef.current?.focus()}>
              <PlusIcon className="icon-sm" />
            </button>

            <button className="chip-btn" title="Attach file" onClick={() => {
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

            <button className="chip-btn" onClick={() => setShowModePicker(!showModePicker)}>
              <span>{settings.mode === 'plan' ? 'Plan' : 'Code'} ▾</span>
            </button>

            <button className="chip-btn" onClick={() => setShowEffortPicker(!showEffortPicker)}>
              <span style={{ textTransform: 'capitalize' }}>{settings.effort.replace('_', ' ')} ▾</span>
            </button>
          </div>

          {isWorking ? (
            <button className="stop-btn" onClick={cancelTask}>
              <StopIcon className="icon-sm" />
              <span>Stop</span>
            </button>
          ) : (
            <button className="send-btn" onClick={handleComposerSubmit} disabled={status === 'disconnected' || !composerInput.trim()}>
              <span>Send</span>
              <ArrowUpRightIcon className="icon-sm" />
            </button>
          )}
        </div>

        <textarea
          ref={textareaRef}
          className="composer-input"
          placeholder="Describe what to build or type / for commands..."
          value={composerInput}
          onChange={handleComposerInputChange}
          onKeyDown={handleKeyDownTextarea}
          disabled={status === 'disconnected'}
        />
      </div>

      {/* Overflow Menu Popover */}
      {showOverflowMenu && (
        <div className="popover-overlay" onClick={() => setShowOverflowMenu(false)}>
          <div className="popover-card" onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px', borderBottom: '1px solid var(--border)', paddingBottom: '8px' }}>
              <span style={{ fontWeight: 600, fontSize: '12px' }}>ACTIONS & OVERFLOW</span>
              <button className="icon-btn" onClick={() => setShowOverflowMenu(false)}><XMarkIcon className="icon-sm" /></button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
              <button className="chip-btn" onClick={() => { setComposerInput(''); textareaRef.current?.focus(); setShowOverflowMenu(false); }}>✦ New task</button>
              <button className="chip-btn" onClick={() => { executeSlashCommand('/compact'); setShowOverflowMenu(false); }}>✦ Compact Context (/compact)</button>
              <button className="chip-btn" onClick={() => { executeSlashCommand('/context'); setShowOverflowMenu(false); }}>✦ Context Status (/context)</button>
              <button className="chip-btn" onClick={() => { executeSlashCommand('/summarize'); setShowOverflowMenu(false); }}>✦ Task Summary (/summarize)</button>
              <button className="chip-btn" onClick={() => { setShowAgentsOverlay(true); setShowOverflowMenu(false); }}>✦ Active Agents ({subagents.length})</button>
              <button className="chip-btn" onClick={() => { setShowSettings(true); setShowOverflowMenu(false); }}>⚙ Settings Control Plane</button>
              <button className="chip-btn" onClick={() => { setShowPalette(true); setShowOverflowMenu(false); }}>⌘ Command Palette (Cmd+K)</button>
            </div>
          </div>
        </div>
      )}

      {/* Context Compaction Modal */}
      {showCompactionModal && compactionData && (
        <CompactionModal data={compactionData} onClose={() => setShowCompactionModal(false)} />
      )}

      {/* Context Composition Status Modal */}
      {showContextModal && (
        <ContextStatusModal
          composition={contextComposition}
          onClose={() => setShowContextModal(false)}
          onCompactNow={() => executeSlashCommand('/compact')}
        />
      )}

      {/* Clear Context Modal */}
      {showClearModal && (
        <ClearContextModal
          onConfirm={() => {
            setCompactionEventNotice(null);
            setTaskSummary(null);
          }}
          onClose={() => setShowClearModal(false)}
        />
      )}

      {/* Active Agents Popover Modal */}
      {showAgentsOverlay && (
        <div className="popover-overlay" onClick={() => setShowAgentsOverlay(false)}>
          <div className="popover-card" onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px', borderBottom: '1px solid var(--border)', paddingBottom: '8px' }}>
              <span style={{ fontWeight: 600, fontSize: '12px' }}>ACTIVE AGENTS</span>
              <button className="icon-btn" onClick={() => setShowAgentsOverlay(false)}><XMarkIcon className="icon-sm" /></button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
              {subagents.length === 0 ? (
                <div style={{ color: 'var(--text-muted)', fontSize: '12px' }}>No active subagents.</div>
              ) : (
                subagents.map((sub) => (
                  <div key={sub.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: '12px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <span className={`status-dot ${sub.status === 'running' ? 'working' : ''}`} />
                      <span style={{ fontWeight: 600 }}>{sub.role.toUpperCase()}</span>
                    </div>
                    <span style={{ color: 'var(--text-muted)' }}>{sub.task || sub.status}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}

      {/* Mode Picker Popover */}
      {showModePicker && (
        <div className="popover-overlay" onClick={() => setShowModePicker(false)}>
          <div className="popover-card" style={{ width: '240px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ fontWeight: 600, fontSize: '12px', marginBottom: '10px' }}>MODE</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
              <button className="chip-btn" onClick={() => { settingsStore.updateSettings({ mode: 'plan' }); setShowModePicker(false); }}>
                Plan - Analyze and propose
              </button>
              <button className="chip-btn" onClick={() => { settingsStore.updateSettings({ mode: 'code' }); setShowModePicker(false); }}>
                Code - Edit, test and verify
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Effort Picker Popover */}
      {showEffortPicker && (
        <div className="popover-overlay" onClick={() => setShowEffortPicker(false)}>
          <div className="popover-card" style={{ width: '220px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ fontWeight: 600, fontSize: '12px', marginBottom: '10px' }}>EFFORT</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
              {(['low', 'medium', 'high', 'extra_high'] as const).map((level) => (
                <button
                  key={level}
                  className="chip-btn"
                  onClick={() => { settingsStore.updateSettings({ effort: level }); setShowEffortPicker(false); }}
                >
                  <span style={{ textTransform: 'capitalize' }}>{level.replace('_', ' ')}</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

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
