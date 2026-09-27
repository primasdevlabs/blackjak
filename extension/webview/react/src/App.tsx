import React, { useEffect, useState, useRef } from 'react';
import { useAgent } from './hooks/useAgent';
import { settingsStore } from './state/settingsStore';
import { queueStore } from './state/queueStore';
import { agentStore } from './state/agentStore';
import { Chat } from './components/Chat';
import { apiClient } from './api/client';
import { resolveAgentEndpoint } from './api/endpoint';
import { LiveActivity } from './components/LiveActivity';
import { ChangeTracker } from './components/ChangeTracker';
import { CompletionCard } from './components/CompletionCard';
import { AgentResult } from './components/AgentResult';
import { Plan } from './components/Plan';
import { Approval } from './components/Approval';
import { PromptQueue } from './components/queue/PromptQueue';
import { SettingsPage } from './settings/SettingsPage';
import { CommandPalette, CommandOption } from './components/CommandPalette';
import { SlashCommandMenu, CommandItem } from './components/SlashCommandMenu';
import { CompactionModal, CompactionData } from './components/CompactionModal';
import { ContextStatusModal, ContextComposition } from './components/ContextStatusModal';
import { ClearContextModal } from './components/ClearContextModal';
import { TaskSummaryCard, TaskSummaryData } from './components/TaskSummaryCard';
import { DropZone } from './components/DropZone';
import { ReferenceChipList } from './components/ReferenceChip';
import { WorkspaceReference } from './types/events';
import {
  Cog6ToothIcon,
  PlusIcon,
  StopIcon,
  ArrowUpIcon,
  XMarkIcon,
  SparklesIcon,
  ArrowTopRightOnSquareIcon,
  UserGroupIcon,
  ShieldCheckIcon,
  CpuChipIcon,
  ComputerDesktopIcon,
  FolderIcon,
  PaperClipIcon,
  CommandLineIcon,
  ArrowsPointingInIcon,
  ChartBarIcon,
  DocumentTextIcon,
  ChatBubbleLeftIcon,
  ClockIcon,
  TrashIcon,
  ArrowPathIcon,
  DocumentIcon,
  ExclamationTriangleIcon,
  ArrowDownTrayIcon,
  DocumentDuplicateIcon,
  PauseIcon,
  PlayIcon,
} from '@heroicons/react/24/outline';
import './App.css';

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'now';
  if (mins < 60) return `${mins}m`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h`;
  const days = Math.floor(hrs / 24);
  if (days < 7) return `${days}d`;
  return new Date(iso).toLocaleDateString();
}

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
    pendingApproval,
    filesRead,
    thought,
    completion,
    workspaceRoot,
    startTask,
    cancelTask,
    respondApproval,
    configureBackend,
    addAttachment,
    removeAttachment,
    selectSubagent,
    sessions,
    activeSessionId,
    createSession,
    switchSession,
    closeSession,
    runs,
    refreshRuns,
    loadRunFromHistory,
    deleteRunFromHistory,
    retryLastRun,
    pauseTask,
    resumeTask,
  } = useAgent();

  const [settings, setSettings] = useState(settingsStore.getSettings());
  const [showSettings, setShowSettings] = useState(false);
  const [showPalette, setShowPalette] = useState(false);
  const [showPlusMenu, setShowPlusMenu] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [selPopup, setSelPopup] = useState<{ text: string; x: number; y: number } | null>(null);
  const [showAgentsOverlay, setShowAgentsOverlay] = useState(false);
  const [showEffortPicker, setShowEffortPicker] = useState(false);
  const [composerInput, setComposerInput] = useState('');
  const [composerRefs, setComposerRefs] = useState<WorkspaceReference[]>([]);

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

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const sessionRef = useRef<HTMLDivElement>(null);

  // Text selection in the session stream → floating Quote/Copy popup.
  const handleSessionMouseUp = () => {
    const sel = window.getSelection();
    const text = sel?.toString().trim() ?? '';
    if (!text || !sel || sel.rangeCount === 0 || sel.isCollapsed || !sessionRef.current?.contains(sel.anchorNode)) {
      setSelPopup(null);
      return;
    }
    const rect = sel.getRangeAt(0).getBoundingClientRect();
    const host = sessionRef.current.getBoundingClientRect();
    setSelPopup({
      text,
      x: rect.left + rect.width / 2 - host.left,
      y: rect.top - host.top,
    });
  };

  const quoteSelection = () => {
    if (!selPopup) return;
    const quote = selPopup.text.split('\n').map((l) => `> ${l}`).join('\n');
    setComposerInput((prev) => (prev.trim() ? `${prev}\n\n${quote}\n` : `${quote}\n`));
    setSelPopup(null);
    window.getSelection()?.removeAllRanges();
    textareaRef.current?.focus();
  };

  const copySelection = () => {
    if (!selPopup) return;
    void navigator.clipboard.writeText(selPopup.text);
    setSelPopup(null);
    window.getSelection()?.removeAllRanges();
  };

  useEffect(() => {
    const endpoint = resolveAgentEndpoint();
    configureBackend(endpoint.httpUrl, endpoint.wsUrl);

    // Extension host pushes the authoritative backend endpoint once the
    // agent process is ready (covers startup races and port changes).
    const handleHostMessage = (e: MessageEvent) => {
      const msg = e.data;
      if (msg && msg.type === 'agent.endpoint' && msg.data?.port) {
        const httpUrl = `http://${msg.data.host || '127.0.0.1'}:${msg.data.port}`;
        const wsUrl = `ws://${msg.data.host || '127.0.0.1'}:${msg.data.port}/ws`;
        if (wsUrl !== agentStore.getState().wsUrl) {
          configureBackend(httpUrl, wsUrl);
        }
      }
    };
    window.addEventListener('message', handleHostMessage);

    const unsubSettings = settingsStore.subscribe(() => {
      setSettings(settingsStore.getSettings());
    });
    settingsStore.fetchSettings();

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
      window.removeEventListener('message', handleHostMessage);
    };
  }, []);

  const runBackendCommand = async (command: string, args?: Record<string, unknown>): Promise<any> => {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/commands/execute`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command, args }),
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || `/${command} failed`);
    return json.data;
  };

  const executeSlashCommand = async (commandName: string) => {
    const cleanCmd = commandName.replace('/', '').toLowerCase().trim();

    try {
      if (cleanCmd === 'compact') {
        const data = await runBackendCommand('compact');
        if (data.tokensBefore) {
          setCompactionEventNotice({
            before: `${(data.tokensBefore / 1000).toFixed(1)}k`,
            after: data.tokensAfter ? `${(data.tokensAfter / 1000).toFixed(1)}k` : 'pending',
          });
          setCompactionData({
            tokensBefore: data.tokensBefore,
            tokensAfter: data.tokensAfter ?? 0,
            tokensSaved: data.tokensBefore - (data.tokensAfter ?? 0),
            preserved: data.status === 'requested' ? ['System prompt', 'Task objective', 'Recent messages'] : [],
            removed: data.status === 'requested' ? ['Older tool outputs and messages'] : [],
          });
        }
        if (data.message) agentStore.addSystemMessage(data.message);
      } else if (cleanCmd === 'context') {
        const data = await runBackendCommand('context');
        setContextComposition(data);
        setShowContextModal(true);
      } else if (cleanCmd === 'clear') {
        setShowClearModal(true);
      } else if (cleanCmd === 'summarize') {
        const data = await runBackendCommand('summarize');
        setTaskSummary(data);
      } else if (cleanCmd === 'plan') {
        const data = await runBackendCommand('plan');
        if (data?.steps?.length) {
          agentStore.addSystemMessage(`Plan (${data.currentStep + 1}/${data.steps.length}): ${data.steps[data.currentStep] || 'done'}`);
        } else {
          settingsStore.updateSettings({ mode: 'plan' });
          agentStore.addSystemMessage('Switched to plan mode.');
        }
      } else if (cleanCmd === 'code') {
        settingsStore.updateSettings({ mode: 'code' });
        agentStore.addSystemMessage('Switched to code mode.');
      } else if (cleanCmd === 'model') {
        setShowSettings(true);
      } else if (cleanCmd === 'effort') {
        setShowEffortPicker(true);
      } else if (cleanCmd === 'agents') {
        setShowAgentsOverlay(true);
      } else if (cleanCmd === 'queue') {
        const items = await runBackendCommand('queue');
        agentStore.addSystemMessage(
          items?.length
            ? `Queue: ${items.length} item(s)\n${items.map((q: any, i: number) => `${i + 1}. [${q.status}] ${q.prompt}`).join('\n')}`
            : 'Queue is empty.'
        );
      } else if (cleanCmd === 'changes') {
        const data = await runBackendCommand('changes');
        agentStore.addSystemMessage(
          data?.changes?.length
            ? `Changed files:\n${data.changes.map((c: any) => `${c.type === 'created' ? '+' : '~'} ${c.path}`).join('\n')}`
            : 'No file changes in the current run.'
        );
      } else if (cleanCmd === 'files') {
        const data = await runBackendCommand('files');
        const refs = [...(data?.references || []), ...(data?.attachments || [])];
        agentStore.addSystemMessage(
          refs.length
            ? `Context files:\n${refs.map((f: any) => `- ${f.path}`).join('\n')}`
            : 'No files attached to the current run.'
        );
      } else if (cleanCmd === 'undo') {
        const data = await runBackendCommand('undo');
        agentStore.addSystemMessage(`Reverted: ${data.path}`);
      } else if (cleanCmd === 'stop') {
        await runBackendCommand('stop').catch(() => cancelTask());
        cancelTask();
      } else {
        agentStore.addSystemMessage(`Unknown command: /${cleanCmd}`);
      }
    } catch (e: any) {
      agentStore.addSystemMessage(`/${cleanCmd} failed: ${e.message}`);
    }
  };

  // Parse @path references from composer input
  const parseReferencesFromText = (text: string): WorkspaceReference[] => {
    const refPattern = /@(?:(file|folder):)?([a-zA-Z0-9_\-./\\]+(?::\d+)?)/g;
    const refs: WorkspaceReference[] = [];
    let match: RegExpExecArray | null;
    while ((match = refPattern.exec(text)) !== null) {
      const prefix = match[1]; // file | folder | undefined
      const pathPart = match[2];
      let line: number | undefined;
      let cleanPath = pathPart;
      if (cleanPath.includes(':')) {
        const idx = cleanPath.lastIndexOf(':');
        const num = parseInt(cleanPath.substring(idx + 1), 10);
        if (!isNaN(num)) {
          line = num;
          cleanPath = cleanPath.substring(0, idx);
        }
      }
      const isFolder = prefix === 'folder' || (!cleanPath.includes('.') && !line);
      refs.push({
        type: isFolder ? 'folder' : 'file',
        path: cleanPath,
        line,
        raw: match[0],
      });
    }
    return refs;
  };

  const handleRemoveReference = (raw: string) => {
    setComposerInput((prev) => prev.replace(raw, '').replace(/\s{2,}/g, ' ').trim());
    setComposerRefs((prev) => prev.filter((r) => r.raw !== raw));
  };

  const handleComposerInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value;
    setComposerInput(val);
    setComposerRefs(parseReferencesFromText(val));
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
    } else if (activeRunStatus === 'paused' || activeRunStatus === 'cancelled') {
      // Resume from checkpoint with the new instruction appended to context.
      resumeTask(text);
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
  const modelLabel = (settings.coding?.modelId || settings.codingModelId || 'Model').replace(/^claude-/, 'Claude ').replace(/^gpt-/, 'GPT-');
  const workspaceName = workspaceRoot ? workspaceRoot.split(/[/\\]/).filter(Boolean).pop() : '';

  const paletteCommands: CommandOption[] = [
    { id: 'new-task', label: 'New task', shortcut: 'Enter', action: () => textareaRef.current?.focus() },
    { id: 'plan-mode', label: 'Plan mode', action: () => settingsStore.updateSettings({ mode: 'plan' }) },
    { id: 'code-mode', label: 'Code mode', action: () => settingsStore.updateSettings({ mode: 'code' }) },
    { id: 'effort-low', label: 'Effort: Low', action: () => settingsStore.updateSettings({ effort: 'low' }) },
    { id: 'effort-medium', label: 'Effort: Medium', action: () => settingsStore.updateSettings({ effort: 'medium' }) },
    { id: 'effort-high', label: 'Effort: High', action: () => settingsStore.updateSettings({ effort: 'high' }) },
    { id: 'effort-extra', label: 'Effort: Max', action: () => settingsStore.updateSettings({ effort: 'extra_high' }) },
    { id: 'open-settings', label: 'Settings', action: () => setShowSettings(true) },
    { id: 'cancel-task', label: 'Stop task', action: () => cancelTask() },
  ];

  return (
    <DropZone onFileDrop={(path, type) => addAttachment(path, type)}>
    <div className="app-sidebar-container">
      {/* Session tabs — one conversation per tab, IDE-style */}
      <div className="session-tabstrip">
        <div className="session-tabs">
          {sessions.map((s) => (
            <div
              key={s.id}
              className={`session-tab ${s.id === activeSessionId ? 'active' : ''}`}
              onClick={() => switchSession(s.id)}
              title={s.title}
            >
              {(s.runStatus === 'running' || s.runStatus === 'pending') ? (
                <span className="session-tab-dot running" />
              ) : s.runStatus === 'waiting' ? (
                <span className="session-tab-dot waiting" />
              ) : s.runStatus === 'failed' ? (
                <span className="session-tab-dot failed" />
              ) : null}
              <span className="session-tab-title">{s.title}</span>
              {sessions.length > 1 && (
                <button
                  className="session-tab-close"
                  onClick={(e) => { e.stopPropagation(); closeSession(s.id); }}
                  title="Close session"
                >
                  <XMarkIcon className="icon-xs" />
                </button>
              )}
            </div>
          ))}
        </div>
        <button className="session-tab-new" onClick={createSession} title="New session">
          <PlusIcon className="icon-sm" />
        </button>
        <div className="composer-anchored history-anchor">
          <button
            className="session-tab-new"
            title="Session history"
            onClick={() => {
              const next = !showHistory;
              setShowHistory(next);
              if (next) refreshRuns();
            }}
          >
            <ClockIcon className="icon-sm" />
          </button>

          {showHistory && (
            <>
              <div className="dropdown-backdrop" onClick={() => setShowHistory(false)} />
              <div className="history-dropdown">
                <div className="history-header">History</div>
                {runs.length === 0 ? (
                  <div className="history-empty">No past sessions</div>
                ) : (
                  runs.map((r) => (
                    <div key={r.id} className="history-item" title={r.prompt}>
                      <span className={`history-status ${r.status}`} />
                      <span className="history-title">{r.prompt || 'Untitled'}</span>
                      <span className="history-time">{relativeTime(r.createdAt)}</span>
                      <button
                        className="history-action"
                        onClick={() => { setShowHistory(false); loadRunFromHistory(r.id); }}
                        title="Load into new session"
                      >
                        <ArrowDownTrayIcon className="icon-sm" />
                      </button>
                      <button
                        className="history-action danger"
                        onClick={() => deleteRunFromHistory(r.id)}
                        title="Delete from history"
                      >
                        <TrashIcon className="icon-sm" />
                      </button>
                    </div>
                  ))
                )}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Header (~40px) */}
      <header className="sidebar-header">
        <div className="header-brand">
          <SparklesIcon className="icon-sm" />
          <span>BlackJak</span>
        </div>

        <div className="header-status-area">
          {subagents.length > 0 && (
            <button className="chip-btn" onClick={() => setShowAgentsOverlay(true)}>
              <UserGroupIcon className="icon-sm" />
              <span>{subagents.length} agents</span>
            </button>
          )}

          <button className="icon-btn" onClick={() => setShowSettings(true)} title="Settings">
            <Cog6ToothIcon className="icon" />
          </button>
        </div>
      </header>

      {/* Session stream — single column */}
      <div ref={sessionRef} style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column', position: 'relative' }} onMouseUp={handleSessionMouseUp}>
        <section className="conversation-surface">
          {(!settings.providers[settings.activeProvider]?.apiKey && settings.providers[settings.activeProvider]?.storageMode !== 'environment' && messages.length === 0) && (
            <div style={{ padding: '20px', textAlign: 'center', backgroundColor: 'var(--bg-surface)', borderRadius: '6px', border: '1px solid var(--border)', margin: '12px 0' }}>
              <div style={{ fontWeight: 600, fontSize: '13px', color: 'var(--text-primary)' }}>No model provider</div>
              <div style={{ color: 'var(--text-muted)', fontSize: '12px', margin: '8px 0 14px 0' }}>Add an API key to start working.</div>
              <button className="send-btn" style={{ margin: '0 auto' }} onClick={() => setShowSettings(true)}>
                Set up provider
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
                <SparklesIcon className="icon-sm" />
                <span style={{ fontSize: '12px', fontWeight: 500, color: 'var(--text-primary)' }}>
                  Context compacted: {compactionEventNotice.before} to {compactionEventNotice.after}
                </span>
              </div>
              <span style={{ fontSize: '11px', color: 'var(--text-muted)' }}>
                <ArrowTopRightOnSquareIcon className="icon-sm" style={{ display: 'inline', verticalAlign: 'middle' }} />
              </span>
            </div>
          )}

          {taskSummary && <TaskSummaryCard summary={taskSummary} />}

          <Plan plan={plan} />

          <LiveActivity
            status={activeRunStatus}
            tools={toolExecutions}
            filesRead={filesRead}
            thought={thought}
          />

          <ChangeTracker changes={fileChanges} subagents={subagents} />

          {activeRunStatus === 'paused' && (
            <div className="retry-bar">
              <PauseIcon className="icon-sm" />
              <span>Paused — context saved, resume anytime</span>
              <button className="chip-btn" onClick={() => resumeTask()}>
                <PlayIcon className="icon-sm" />
                <span>Resume</span>
              </button>
            </div>
          )}

          {activeRunStatus === 'cancelled' && (
            <div className="retry-bar">
              <StopIcon className="icon-sm" />
              <span>Stopped — context checkpointed</span>
              <button className="chip-btn" onClick={() => resumeTask()}>
                <PlayIcon className="icon-sm" />
                <span>Resume</span>
              </button>
            </div>
          )}

          {activeRunStatus === 'failed' && (
            <div className="retry-bar">
              <ExclamationTriangleIcon className="icon-sm" />
              <span>Run failed</span>
              <button className="chip-btn" onClick={retryLastRun}>
                <ArrowPathIcon className="icon-sm" />
                <span>Retry</span>
              </button>
              <button className="chip-btn" onClick={() => resumeTask()} title="Resume from checkpointed context">
                <PlayIcon className="icon-sm" />
                <span>Resume</span>
              </button>
            </div>
          )}

          {completion && activeRunStatus === 'completed' && (
            <CompletionCard completion={completion} />
          )}

          <PromptQueue />
        </section>

        {selPopup && (
          <div
            className="selection-popup"
            style={{ left: Math.max(selPopup.x, 60), top: Math.max(selPopup.y - 10, 44) }}
          >
            <button className="selection-popup-btn" onMouseDown={(e) => { e.preventDefault(); quoteSelection(); }} title="Quote in composer">
              <ChatBubbleLeftIcon className="icon-sm" />
              <span>Quote</span>
            </button>
            <button className="selection-popup-btn" onMouseDown={(e) => { e.preventDefault(); copySelection(); }} title="Copy">
              <DocumentDuplicateIcon className="icon-sm" />
              <span>Copy</span>
            </button>
          </div>
        )}
      </div>

      {/* Composer */}
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

        <div className="composer-card">
          {(attachments.length > 0 || composerRefs.length > 0) && (
            <div className="composer-chips">
              {attachments.map((att) => (
                <span key={att.id} className="composer-chip" title={att.path}>
                  {att.type === 'folder' ? <FolderIcon className="icon-sm" /> : <DocumentIcon className="icon-sm" />}
                  <span>{att.name}</span>
                  <button className="composer-chip-remove" onClick={() => removeAttachment(att.id)} title="Remove">
                    <XMarkIcon className="icon-xs" />
                  </button>
                </span>
              ))}
              <ReferenceChipList references={composerRefs} onRemove={handleRemoveReference} />
            </div>
          )}

          <textarea
            ref={textareaRef}
            className="composer-input"
            placeholder={status === 'disconnected' ? 'Connecting…' : 'Ask BlackJak — @ to reference files, / for commands'}
            value={composerInput}
            onChange={handleComposerInputChange}
            onKeyDown={handleKeyDownTextarea}
            disabled={status === 'disconnected'}
            rows={3}
          />

          <div className="composer-toolbar">
            <div className="composer-actions">
              <div className="composer-anchored">
                <button className="composer-icon-btn" title="Attach & actions" onClick={() => setShowPlusMenu(!showPlusMenu)}>
                  <PlusIcon className="icon-sm" />
                </button>

                {showPlusMenu && (
                  <>
                    <div className="dropdown-backdrop" onClick={() => setShowPlusMenu(false)} />
                    <div className="composer-dropdown composer-menu">
                      <button className="composer-dropdown-item" onClick={() => {
                        setShowPlusMenu(false);
                        const input = document.createElement('input');
                        input.type = 'file';
                        input.onchange = (e: any) => {
                          const file = e.target?.files?.[0];
                          if (file) addAttachment(file.name);
                        };
                        input.click();
                      }}>
                        <PaperClipIcon className="icon-sm" />
                        <span>Attach file</span>
                      </button>
                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); setShowPalette(true); }}>
                        <CommandLineIcon className="icon-sm" />
                        <span>Commands</span>
                        <kbd>Ctrl+K</kbd>
                      </button>

                      <div className="composer-dropdown-divider" />

                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); executeSlashCommand('/compact'); }}>
                        <ArrowsPointingInIcon className="icon-sm" />
                        <span>Compact context</span>
                      </button>
                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); executeSlashCommand('/context'); }}>
                        <ChartBarIcon className="icon-sm" />
                        <span>Context usage</span>
                      </button>
                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); executeSlashCommand('/summarize'); }}>
                        <DocumentTextIcon className="icon-sm" />
                        <span>Summarize task</span>
                      </button>
                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); setShowAgentsOverlay(true); }}>
                        <UserGroupIcon className="icon-sm" />
                        <span>Agents</span>
                        {subagents.length > 0 && <span className="dropdown-check">{subagents.length}</span>}
                      </button>

                      <div className="composer-dropdown-divider" />

                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); createSession(); }}>
                        <ChatBubbleLeftIcon className="icon-sm" />
                        <span>New session</span>
                      </button>
                      <button className="composer-dropdown-item" onClick={() => { setShowPlusMenu(false); setShowSettings(true); }}>
                        <Cog6ToothIcon className="icon-sm" />
                        <span>Settings</span>
                      </button>
                    </div>
                  </>
                )}
              </div>

              <button
                className="chip-btn"
                onClick={() => settingsStore.updateSettings({ mode: settings.mode === 'plan' ? 'code' : 'plan' })}
                title={`Mode: ${settings.mode === 'plan' ? 'Plan — click for Code' : 'Code — click for Plan'}`}
              >
                <ShieldCheckIcon className="icon-sm" />
                <span>{settings.mode === 'plan' ? 'Plan' : 'Code'}</span>
              </button>

              <div className="composer-anchored">
                <button className="chip-btn" onClick={() => setShowEffortPicker(!showEffortPicker)} title="Model & effort">
                  <span>{modelLabel} · <span style={{ textTransform: 'capitalize' }}>{settings.effort.replace('_', ' ')}</span></span>
                </button>

                {showEffortPicker && (
                  <>
                    <div className="dropdown-backdrop" onClick={() => setShowEffortPicker(false)} />
                    <div className="composer-dropdown">
                      {(['low', 'medium', 'high', 'extra_high'] as const).map((level) => (
                        <button
                          key={level}
                          className={`composer-dropdown-item ${settings.effort === level ? 'active' : ''}`}
                          onClick={() => { settingsStore.updateSettings({ effort: level }); setShowEffortPicker(false); }}
                        >
                          <span style={{ textTransform: 'capitalize' }}>{level.replace('_', ' ')}</span>
                          {settings.effort === level && <span className="dropdown-check">✓</span>}
                        </button>
                      ))}
                    </div>
                  </>
                )}
              </div>
            </div>

            <div className="composer-actions">
              <span className="chip-btn composer-env-chip" title="Environment">
                <CpuChipIcon className="icon-sm" />
                <span>Local</span>
              </span>
              {isWorking ? (
                <>
                  <button className="composer-send-btn" onClick={pauseTask} title="Pause — context stays saved">
                    <PauseIcon className="icon-sm" />
                  </button>
                  <button className="composer-send-btn composer-send-btn-stop" onClick={cancelTask} title="Stop">
                    <StopIcon className="icon-sm" />
                  </button>
                </>
              ) : activeRunStatus === 'paused' ? (
                <button className="composer-send-btn" onClick={() => resumeTask()} title="Resume">
                  <PlayIcon className="icon-sm" />
                </button>
              ) : (
                <button className="composer-send-btn" onClick={handleComposerSubmit} disabled={status === 'disconnected' || !composerInput.trim()} title="Send">
                  <ArrowUpIcon className="icon-sm" />
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Context strip — environment + workspace */}
        <div className="composer-contextbar">
          <span className="context-chip">
            <ComputerDesktopIcon className="icon-sm" />
            <span>Local</span>
          </span>
          {workspaceName && (
            <span className="context-chip" title={workspaceRoot}>
              <FolderIcon className="icon-sm" />
              <span>{workspaceName}</span>
            </span>
          )}
        </div>
      </div>

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
            agentStore.clearConversation();
            runBackendCommand('clear').catch(() => {});
            agentStore.addSystemMessage('Started fresh context.');
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
              <span style={{ fontWeight: 600, fontSize: '12px' }}>AGENTS</span>
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

      <Approval request={pendingApproval} onRespond={respondApproval} />

      <AgentResult subagent={selectedSubagentObj} fileChanges={fileChanges} onClose={() => selectSubagent(null)} />

      {showSettings && (
        <SettingsPage agentStatus={status} workspacePath="" onClose={() => setShowSettings(false)} />
      )}

      <CommandPalette isOpen={showPalette} onClose={() => setShowPalette(false)} commands={paletteCommands} />
    </div>
    </DropZone>
  );
};

export default App;
