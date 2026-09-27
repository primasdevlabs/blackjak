import { AgentEvent, ApprovalRequest, Attachment, FileChange, Plan, PROTOCOL_VERSION, RunDTO, RunStatus, ServerMessage, Subagent, WorkspaceReference } from '../types/events';
import { ConnectionStatus, wsClient } from '../api/websocket';
import { apiClient } from '../api/client';
import { activityManager, describeToolCall } from '../activity/activityManager';
import { agentHost } from '../host';

export interface ToolExecution {
  id: string;
  name: string;
  step: string;
  status: 'running' | 'completed' | 'failed';
  output?: string;
}

export interface TestResult {
  suite: string;
  passed: number;
  failed: number;
  output: string;
}

export interface ChatMessage {
  id: string;
  sender: 'user' | 'agent';
  text: string;
  timestamp: string;
}

export interface CompletionSummary {
  result: string;
  recommendations?: string[];
  durationMs?: number;
  filesChanged?: number;
  subagents?: number;
  steps?: number;
}

/** Lightweight per-tab metadata rendered in the session tab strip. */
export interface SessionMeta {
  id: string;
  title: string;
  runStatus: RunStatus | null;
}

/** Full per-session data — the tab's entire conversation + run state. */
interface Session {
  id: string;
  title: string;
  createdAt: string;
  runId: string | null;
  runStatus: RunStatus | null;
  messages: ChatMessage[];
  plan: Plan | null;
  toolExecutions: ToolExecution[];
  subagents: Subagent[];
  fileChanges: FileChange[];
  pendingApproval: ApprovalRequest | null;
  filesRead: string[];
  thought: string;
  completion: CompletionSummary | null;
}

export interface AgentState {
  backendUrl: string;
  wsUrl: string;
  status: ConnectionStatus;
  runs: RunDTO[];
  sessions: SessionMeta[];
  activeSessionId: string;
  activeRunId: string | null;
  activeRunStatus: RunStatus | null;
  messages: ChatMessage[];
  plan: Plan | null;
  toolExecutions: ToolExecution[];
  subagents: Subagent[];
  selectedSubagentId: string | null;
  fileChanges: FileChange[];
  attachments: Attachment[];
  references: WorkspaceReference[];
  terminalLogs: string[];
  testResults: TestResult[];
  pendingApproval: ApprovalRequest | null;
  filesRead: string[];
  thought: string;
  completion: CompletionSummary | null;
  workspaceRoot: string;
}

type Listener = () => void;

const DEFAULT_TITLE = 'New session';

function makeSession(id: string, title = DEFAULT_TITLE): Session {
  return {
    id,
    title,
    createdAt: new Date().toISOString(),
    runId: null,
    runStatus: null,
    messages: [],
    plan: null,
    toolExecutions: [],
    subagents: [],
    fileChanges: [],
    pendingApproval: null,
    filesRead: [],
    thought: '',
    completion: null,
  };
}

class AgentStore {
  private state: AgentState = {
    backendUrl: 'http://127.0.0.1:47811',
    wsUrl: 'ws://127.0.0.1:47811/ws',
    status: 'disconnected',
    runs: [],
    sessions: [],
    activeSessionId: '',
    activeRunId: null,
    activeRunStatus: null,
    messages: [],
    plan: null,
    toolExecutions: [],
    subagents: [],
    selectedSubagentId: null,
    fileChanges: [],
    attachments: [],
    references: [],
    terminalLogs: [],
    testResults: [],
    pendingApproval: null,
    filesRead: [],
    thought: '',
    completion: null,
    workspaceRoot: '',
  };

  private listeners: Set<Listener> = new Set();

  private replaying = false;

  /** All sessions, keyed by session id — the source of truth for per-tab state. */
  private sessionMap = new Map<string, Session>();
  /** runId → sessionId routing so events land on the owning tab. */
  private runToSession = new Map<string, string>();
  /** Session ids waiting for their run.created event, FIFO. */
  private awaitingRun: string[] = [];

  constructor() {
    const first = makeSession(`sess_${Date.now()}`);
    this.sessionMap.set(first.id, first);
    this.state.activeSessionId = first.id;
    this.publish();

    wsClient.onStatusChange((status) => {
      this.setState({ status });
      if (status === 'connected') {
        void this.reattach();
      }
    });

    wsClient.onMessage((msg) => {
      this.handleServerMessage(msg);
    });
  }

  /**
   * Reattach to a run already in flight — restores conversation, plan, tool
   * activity, file changes, and pending approval from the run's event log.
   */
  private async reattach() {
    try {
      // Pick up workspace + host info for the composer context bar.
      try {
        const health = await apiClient.getHealth();
        if (health.workspace) this.setState({ workspaceRoot: health.workspace });
      } catch { /* health endpoint optional */ }

      const runs = await apiClient.listRuns();
      this.setState({ runs });
      const live = runs
        .filter((r) => r.status === 'running' || r.status === 'waiting' || r.status === 'pending' || r.status === 'paused')
        .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())[0];
      if (!live) return;

      const run = await apiClient.getRun(live.id);

      // Bind the live run to the active session.
      const sess = this.activeSession();
      if (!sess) return;
      this.loadRunIntoSession(run, sess);
    } catch {
      // backend unreachable or no run history — stay idle
    }
  }

  /** Hydrate a session from a run record: bind runId, seed fields, replay events. */
  private loadRunIntoSession(run: RunDTO, sess: Session) {
    sess.runId = run.id;
    sess.runStatus = run.status;
    sess.title = truncateTitle(run.prompt);
    sess.messages = [{
      id: `msg_${run.id}`,
      sender: 'user',
      text: run.prompt,
      timestamp: run.createdAt,
    }];
    sess.plan = run.plan ?? null;
    sess.subagents = run.subagents ?? [];
    sess.fileChanges = run.fileChanges ?? [];
    sess.pendingApproval = run.pendingApproval ?? null;
    sess.toolExecutions = [];
    sess.filesRead = [];
    sess.thought = '';
    sess.completion = null;
    this.setState({ terminalLogs: [], testResults: [] });
    this.runToSession.set(run.id, sess.id);
    this.publish();

    this.replaying = true;
    try {
      for (const evt of run.events ?? []) {
        this.handleServerMessage({
          protocolVersion: PROTOCOL_VERSION,
          type: evt.type,
          runId: evt.runId,
          agentId: evt.agentId,
          timestamp: evt.timestamp,
          data: evt.data,
        });
      }
    } finally {
      this.replaying = false;
    }
  }

  /** Refresh the run list — used by the session-history dropdown. */
  async refreshRuns() {
    try {
      const runs = await apiClient.listRuns();
      this.setState({ runs });
    } catch {
      // backend unreachable
    }
  }

  /** Open a past run in a new session tab and replay its transcript. */
  async loadRunFromHistory(runId: string) {
    try {
      const run = await apiClient.getRun(runId);
      const sess = makeSession(`sess_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`);
      this.sessionMap.set(sess.id, sess);
      this.state.activeSessionId = sess.id;
      this.loadRunIntoSession(run, sess);
    } catch {
      this.addSystemMessage('Failed to load that run — it may have been deleted.');
    }
  }

  /** Delete a run from backend history and unbind any session showing it. */
  async deleteRunFromHistory(runId: string) {
    try {
      await apiClient.deleteRun(runId);
    } catch {
      this.addSystemMessage('Failed to delete that run.');
      return;
    }
    const sessId = this.runToSession.get(runId);
    if (sessId) {
      this.runToSession.delete(runId);
      const sess = this.sessionMap.get(sessId);
      if (sess && sess.runId === runId) {
        sess.runId = null;
        sess.runStatus = null;
      }
    }
    this.setState({ runs: this.state.runs.filter((r) => r.id !== runId) });
    this.publish();
  }

  getState(): AgentState {
    return this.state;
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  private setState(partial: Partial<AgentState>) {
    this.state = { ...this.state, ...partial };
    this.listeners.forEach((l) => l());
  }

  private activeSession(): Session | undefined {
    return this.sessionMap.get(this.state.activeSessionId);
  }

  /** Copy the active session's fields into the flat projection + refresh tab list. */
  private publish() {
    const sess = this.activeSession();
    this.setState({
      sessions: [...this.sessionMap.values()].map((s) => ({
        id: s.id,
        title: s.title,
        runStatus: s.runStatus,
      })),
      activeRunId: sess?.runId ?? null,
      activeRunStatus: sess?.runStatus ?? null,
      messages: sess?.messages ?? [],
      plan: sess?.plan ?? null,
      toolExecutions: sess?.toolExecutions ?? [],
      subagents: sess?.subagents ?? [],
      fileChanges: sess?.fileChanges ?? [],
      pendingApproval: sess?.pendingApproval ?? null,
      filesRead: sess?.filesRead ?? [],
      thought: sess?.thought ?? '',
      completion: sess?.completion ?? null,
    });
  }

  // --- Session management ---

  createSession() {
    const sess = makeSession(`sess_${Date.now()}_${Math.random().toString(36).slice(2, 6)}`);
    this.sessionMap.set(sess.id, sess);
    this.state.activeSessionId = sess.id;
    this.publish();
  }

  switchSession(id: string) {
    if (!this.sessionMap.has(id) || id === this.state.activeSessionId) return;
    this.state.activeSessionId = id;
    this.publish();
  }

  closeSession(id: string) {
    const sess = this.sessionMap.get(id);
    if (!sess) return;

    // Cancel a live run before dropping the tab.
    if (sess.runId && (sess.runStatus === 'running' || sess.runStatus === 'waiting' || sess.runStatus === 'pending')) {
      wsClient.cancelRun(sess.runId);
    }
    if (sess.runId) this.runToSession.delete(sess.runId);
    this.sessionMap.delete(id);

    if (this.sessionMap.size === 0) {
      const fresh = makeSession(`sess_${Date.now()}`);
      this.sessionMap.set(fresh.id, fresh);
    }
    if (this.state.activeSessionId === id) {
      this.state.activeSessionId = [...this.sessionMap.keys()][0];
    }
    this.publish();
  }

  // ---

  configureBackend(backendUrl: string, wsUrl: string) {
    apiClient.setBaseUrl(backendUrl);
    wsClient.setUrl(wsUrl);
    this.setState({ backendUrl, wsUrl });
    wsClient.connect();
  }

  addAttachment(path: string, type: 'file' | 'folder' = 'file') {
    const name = path.split(/[/\\]/).pop() || path;
    const att: Attachment = {
      id: `att_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
      type,
      path,
      name,
    };
    this.setState({ attachments: [...this.state.attachments, att] });
  }

  removeAttachment(id: string) {
    this.setState({
      attachments: this.state.attachments.filter((a) => a.id !== id),
    });
  }

  selectSubagent(subagentId: string | null) {
    this.setState({ selectedSubagentId: subagentId });
  }

  /** Reset the visible conversation — used by /clear. */
  clearConversation() {
    const sess = this.activeSession();
    if (!sess) return;
    const fresh = makeSession(sess.id, sess.title);
    fresh.runId = sess.runId;
    fresh.runStatus = sess.runStatus;
    this.sessionMap.set(sess.id, fresh);
    this.publish();
  }

  /** Append a local system/agent message — used for slash-command results. */
  addSystemMessage(text: string) {
    const sess = this.activeSession();
    if (!sess) return;
    sess.messages = [
      ...sess.messages,
      {
        id: `sys_${Date.now()}`,
        sender: 'agent',
        text,
        timestamp: new Date().toISOString(),
      },
    ];
    this.publish();
  }

  startTask(prompt: string, workspace?: string) {
    const sess = this.activeSession();
    if (!sess) return;

    const userMsg: ChatMessage = {
      id: `msg_${Date.now()}`,
      sender: 'user',
      text: prompt,
      timestamp: new Date().toISOString(),
    };

    // First prompt names the tab.
    if (sess.title === DEFAULT_TITLE || sess.messages.length === 0) {
      sess.title = truncateTitle(prompt);
    }

    sess.messages = [...sess.messages, userMsg];
    sess.plan = null;
    sess.toolExecutions = [];
    sess.subagents = [];
    sess.fileChanges = [];
    sess.pendingApproval = null;
    sess.filesRead = [];
    sess.thought = '';
    sess.completion = null;
    sess.runStatus = 'pending';
    this.awaitingRun.push(sess.id);

    const currentAttachments = [...this.state.attachments];
    this.setState({
      terminalLogs: [],
      testResults: [],
      attachments: [], // clear attachments after sending
    });
    this.publish();

    wsClient.startRun(prompt, workspace, currentAttachments);
  }

  /** Resend the last user prompt in the active session — used after a failed run. */
  retryLastRun() {
    const sess = this.activeSession();
    if (!sess) return;
    const lastUser = [...sess.messages].reverse().find((m) => m.sender === 'user');
    if (!lastUser) return;
    this.startTask(lastUser.text);
  }

  cancelTask() {
    const runId = this.activeSession()?.runId;
    if (runId) {
      wsClient.cancelRun(runId);
    }
  }

  /** Gracefully pause the running task — context stays checkpointed. */
  pauseTask() {
    const sess = this.activeSession();
    if (!sess?.runId) return;
    apiClient.pauseRun(sess.runId).catch(() => {});
  }

  /** Resume a paused/cancelled/failed run, optionally steering it with a
   *  new instruction appended to the checkpointed conversation. */
  resumeTask(prompt?: string) {
    const sess = this.activeSession();
    if (!sess?.runId) return;
    if (prompt) {
      sess.messages = [
        ...sess.messages,
        { id: `msg_${Date.now()}`, sender: 'user', text: prompt, timestamp: new Date().toISOString() },
      ];
    }
    sess.runStatus = 'running';
    sess.pendingApproval = null;
    this.publish();
    apiClient.resumeRun(sess.runId, prompt).catch((err) => {
      sess.runStatus = 'failed';
      this.publish();
      console.error('resume failed:', err);
    });
  }

  respondApproval(granted: boolean, reason?: string) {
    const sess = this.activeSession();
    if (sess?.runId && sess.pendingApproval) {
      wsClient.respondApproval(sess.runId, sess.pendingApproval.id, granted, reason);
      sess.pendingApproval = null;
      this.publish();
    }
  }

  /** Approve (keep) or decline (revert) a single pending file change. */
  reviewFileChange(changeId: string, action: 'accept' | 'reject') {
    const sess = this.activeSession();
    if (!sess?.runId) return;
    // Optimistic update — the file.change.reviewed broadcast confirms it.
    sess.fileChanges = sess.fileChanges.map((c) =>
      c.id === changeId ? { ...c, status: action === 'accept' ? 'accepted' : 'rejected' } : c
    );
    this.publish();
    apiClient.reviewFileChange(sess.runId, changeId, action).catch(() => {
      // Revert the optimistic flip on failure.
      sess.fileChanges = sess.fileChanges.map((c) =>
        c.id === changeId ? { ...c, status: 'pending' } : c
      );
      this.publish();
    });
  }

  /** Apply an action to every pending change in the active session. */
  reviewAllFileChanges(action: 'accept' | 'reject') {
    const sess = this.activeSession();
    if (!sess?.runId) return;
    const pending = sess.fileChanges.filter((c) => !c.status || c.status === 'pending');
    if (pending.length === 0) return;
    const status = action === 'accept' ? 'accepted' : 'rejected';
    sess.fileChanges = sess.fileChanges.map((c) =>
      pending.some((p) => p.id === c.id) ? { ...c, status } : c
    );
    this.publish();
    apiClient.reviewFileChange(sess.runId, 'all', action).catch(() => {
      pending.forEach((p) => {
        sess.fileChanges = sess.fileChanges.map((c) => (c.id === p.id ? { ...c, status: 'pending' } : c));
      });
      this.publish();
    });
  }

  private handleServerMessage(msg: ServerMessage) {
    if (!msg.type) return;

    activityManager.handleAgentEvent(msg);

    // run.created binds a runId to the session that started it.
    if (msg.type === 'run.created' && msg.runId) {
      const owner = this.awaitingRun.shift() ?? this.state.activeSessionId;
      const sess = this.sessionMap.get(owner);
      if (sess) {
        sess.runId = msg.runId;
        sess.runStatus = 'pending';
        this.runToSession.set(msg.runId, sess.id);
      }
      this.publish();
      return;
    }

    // Route to the owning session; unknown runIds land on the active tab.
    const sess = (msg.runId && this.sessionMap.get(this.runToSession.get(msg.runId) ?? '')) || this.activeSession();
    if (!sess) return;

    this.applyEvent(sess, msg);
    this.publish();
  }

  private applyEvent(sess: Session, msg: ServerMessage) {
    const data = msg.data;
    const now = msg.timestamp || new Date().toISOString();

    switch (msg.type) {
      case 'run.started':
        sess.runStatus = 'running';
        break;

      case 'run.completed':
        sess.runStatus = 'completed';
        sess.completion = {
          result: data?.result || 'Task completed',
          recommendations: data?.recommendations,
          durationMs: data?.durationMs,
          filesChanged: data?.filesChanged,
          subagents: data?.subagents,
          steps: data?.steps,
        };
        break;

      case 'run.failed':
        sess.runStatus = 'failed';
        if (data?.error) {
          sess.messages = [
            ...sess.messages,
            {
              id: `err_${Date.now()}`,
              sender: 'agent',
              text: `Run failed — ${data.error}`,
              timestamp: now,
            },
          ];
        }
        break;

      case 'run.paused':
        sess.runStatus = 'paused';
        sess.pendingApproval = null;
        break;

      case 'run.resumed':
        sess.runStatus = 'running';
        break;

      case 'run.cancelled':
        sess.runStatus = 'cancelled';
        sess.messages = [
          ...sess.messages,
          {
            id: `cancel_${Date.now()}`,
            sender: 'agent',
            text: 'Run cancelled.',
            timestamp: now,
          },
        ];
        break;

      case 'agent.created':
      case 'agent.started':
        if (data?.id) {
          const existingIdx = sess.subagents.findIndex((s) => s.id === data.id);
          if (existingIdx !== -1) {
            sess.subagents = sess.subagents.map((s, i) =>
              i === existingIdx ? { ...s, status: msg.type === 'agent.started' ? 'running' : 'created' } : s
            );
          } else {
            sess.subagents = [
              ...sess.subagents,
              {
                id: data.id,
                parentRunId: msg.runId || '',
                role: data.role || 'worker',
                task: data.task || '',
                status: msg.type === 'agent.started' ? 'running' : 'created',
                workspaceScope: data.scope,
                startedAt: now,
              },
            ];
          }
        }
        break;

      case 'agent.completed':
      case 'agent.failed':
        if (data?.id) {
          sess.subagents = sess.subagents.map((s) =>
            s.id === data.id
              ? {
                  ...s,
                  status: msg.type === 'agent.completed' ? 'completed' : 'failed',
                  result: data.result,
                  findings: data.findings,
                  error: data.error,
                  finishedAt: now,
                }
              : s
          );
        }
        break;

      case 'agent.thinking':
        sess.thought = data?.thought || '';
        break;

      case 'agent.message':
        if (data?.content) {
          sess.thought = String(data.content).split('\n')[0].slice(0, 140);
          sess.messages = [
            ...sess.messages,
            {
              id: `msg_${Date.now()}`,
              sender: 'agent',
              text: data.content,
              timestamp: now,
            },
          ];
        }
        break;

      case 'agent.plan':
        if (data && Array.isArray(data.steps)) {
          sess.plan = data as Plan;
        }
        break;

      case 'tool.started':
        if (data?.tool) {
          sess.toolExecutions = [
            ...sess.toolExecutions,
            {
              id: `tool_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`,
              name: data.tool,
              step: describeToolCall(data.tool, data.args),
              status: 'running',
            },
          ];
        }
        break;

      case 'tool.completed':
        sess.toolExecutions = sess.toolExecutions.map((t) =>
          t.status === 'running' ? { ...t, status: 'completed' } : t
        );
        break;

      case 'tool.failed':
        sess.toolExecutions = sess.toolExecutions.map((t) =>
          t.status === 'running' ? { ...t, status: 'failed', output: data?.error } : t
        );
        break;

      case 'file.read':
        if (data?.path) {
          const p = String(data.path);
          if (!sess.filesRead.includes(p)) {
            sess.filesRead = [...sess.filesRead, p];
          }
        }
        break;

      case 'file.created':
      case 'file.modified':
      case 'file.deleted':
        if (data?.path) {
          const type = (msg.type.split('.')[1] as any) || 'modified';
          sess.fileChanges = [
            ...sess.fileChanges,
            {
              id: data.id || `fc_${Date.now()}`,
              runId: msg.runId || '',
              agentId: msg.agentId || 'orchestrator',
              path: data.path,
              type,
              diff: data.diff,
              status: data.status || 'pending',
              canRevert: data.canRevert !== false,
              timestamp: now,
            },
          ];

          // Handle new file auto-open through the host adapter
          if (msg.type === 'file.created' && !this.replaying) {
            agentHost.openFile(data.path);
          }
        }
        break;

      case 'file.change.reviewed':
        if (data?.changeId) {
          sess.fileChanges = sess.fileChanges.map((c) =>
            c.id === data.changeId ? { ...c, status: data.status } : c
          );
        }
        break;

      case 'command.output':
        if (data?.output) {
          this.setState({ terminalLogs: [...this.state.terminalLogs, data.output] });
        }
        break;

      case 'command.completed':
        if (data?.stdout) {
          this.setState({ terminalLogs: [...this.state.terminalLogs, data.stdout] });
        }
        break;

      case 'test.output':
      case 'test.passed':
      case 'test.failed':
        if (data?.output || data?.passed !== undefined) {
          const testRes: TestResult = {
            suite: data.suite || 'unit',
            passed: data.passed || 0,
            failed: data.failed || 0,
            output: data.output || '',
          };
          this.setState({ testResults: [...this.state.testResults, testRes] });
        }
        break;

      case 'approval.required':
        if (data?.id) {
          sess.pendingApproval = data as ApprovalRequest;
          sess.runStatus = 'waiting';
        }
        break;
    }
  }
}

function truncateTitle(prompt: string): string {
  const clean = prompt.replace(/\s+/g, ' ').trim();
  return clean.length > 28 ? clean.slice(0, 28) + '…' : clean || DEFAULT_TITLE;
}

export const agentStore = new AgentStore();
