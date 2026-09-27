import { AgentEvent, ApprovalRequest, Attachment, FileChange, Plan, RunDTO, RunStatus, ServerMessage, Subagent, WorkspaceReference } from '../types/events';
import { ConnectionStatus, wsClient } from '../api/websocket';
import { apiClient } from '../api/client';
import { activityManager } from '../activity/activityManager';

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

export interface AgentState {
  backendUrl: string;
  wsUrl: string;
  status: ConnectionStatus;
  runs: RunDTO[];
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
}

type Listener = () => void;

class AgentStore {
  private state: AgentState = {
    backendUrl: 'http://127.0.0.1:8080',
    wsUrl: 'ws://127.0.0.1:8080/ws',
    status: 'disconnected',
    runs: [],
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
  };

  private listeners: Set<Listener> = new Set();

  constructor() {
    wsClient.onStatusChange((status) => {
      this.setState({ status });
    });

    wsClient.onMessage((msg) => {
      this.handleServerMessage(msg);
    });
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

  startTask(prompt: string, workspace?: string) {
    const userMsg: ChatMessage = {
      id: `msg_${Date.now()}`,
      sender: 'user',
      text: prompt,
      timestamp: new Date().toISOString(),
    };

    const currentAttachments = [...this.state.attachments];

    this.setState({
      messages: [...this.state.messages, userMsg],
      plan: null,
      toolExecutions: [],
      subagents: [],
      selectedSubagentId: null,
      fileChanges: [],
      terminalLogs: [],
      testResults: [],
      pendingApproval: null,
      activeRunStatus: 'pending',
      attachments: [], // clear attachments after sending
    });

    wsClient.startRun(prompt, workspace, currentAttachments);
  }

  cancelTask() {
    if (this.state.activeRunId) {
      wsClient.cancelRun(this.state.activeRunId);
    }
  }

  respondApproval(granted: boolean, reason?: string) {
    if (this.state.activeRunId && this.state.pendingApproval) {
      wsClient.respondApproval(
        this.state.activeRunId,
        this.state.pendingApproval.id,
        granted,
        reason
      );
      this.setState({ pendingApproval: null });
    }
  }

  private handleServerMessage(msg: ServerMessage) {
    if (!msg.type) return;

    activityManager.handleAgentEvent(msg);

    if (msg.runId) {
      this.setState({ activeRunId: msg.runId });
    }

    const data = msg.data;

    switch (msg.type) {
      case 'run.started':
        this.setState({ activeRunStatus: 'running' });
        break;

      case 'run.completed':
        this.setState({ activeRunStatus: 'completed' });
        break;

      case 'run.failed':
        this.setState({ activeRunStatus: 'failed' });
        break;

      case 'run.cancelled':
        this.setState({ activeRunStatus: 'cancelled' });
        break;

      case 'agent.created':
      case 'agent.started':
        if (data?.id) {
          const existingIdx = this.state.subagents.findIndex((s) => s.id === data.id);
          const newSub: Subagent = {
            id: data.id,
            parentRunId: msg.runId || '',
            role: data.role || 'worker',
            task: data.task || '',
            status: msg.type === 'agent.started' ? 'running' : 'created',
            workspaceScope: data.scope,
            startedAt: msg.timestamp || new Date().toISOString(),
          };

          if (existingIdx !== -1) {
            const updated = [...this.state.subagents];
            updated[existingIdx] = { ...updated[existingIdx], status: newSub.status };
            this.setState({ subagents: updated });
          } else {
            this.setState({ subagents: [...this.state.subagents, newSub] });
          }
        }
        break;

      case 'agent.completed':
      case 'agent.failed':
        if (data?.id) {
          this.setState({
            subagents: this.state.subagents.map((s) => {
              if (s.id === data.id) {
                return {
                  ...s,
                  status: msg.type === 'agent.completed' ? 'completed' : 'failed',
                  result: data.result,
                  findings: data.findings,
                  error: data.error,
                  finishedAt: msg.timestamp || new Date().toISOString(),
                };
              }
              return s;
            }),
          });
        }
        break;

      case 'agent.message':
        if (data?.content) {
          const agentMsg: ChatMessage = {
            id: `msg_${Date.now()}`,
            sender: 'agent',
            text: data.content,
            timestamp: msg.timestamp || new Date().toISOString(),
          };
          this.setState({ messages: [...this.state.messages, agentMsg] });
        }
        break;

      case 'agent.plan':
        if (data && Array.isArray(data.steps)) {
          this.setState({ plan: data as Plan });
        }
        break;

      case 'tool.started':
        if (data?.tool) {
          const toolExec: ToolExecution = {
            id: `tool_${Date.now()}`,
            name: data.tool,
            step: data.step || '',
            status: 'running',
          };
          this.setState({ toolExecutions: [...this.state.toolExecutions, toolExec] });
        }
        break;

      case 'tool.completed':
        this.setState({
          toolExecutions: this.state.toolExecutions.map((t) =>
            t.status === 'running' ? { ...t, status: 'completed' } : t
          ),
        });
        break;

      case 'file.read':
      case 'file.created':
      case 'file.modified':
      case 'file.deleted':
        if (data?.path) {
          const type = (msg.type.split('.')[1] as any) || 'modified';
          const fileChange: FileChange = {
            id: data.id || `fc_${Date.now()}`,
            runId: msg.runId || '',
            agentId: msg.agentId || 'orchestrator',
            path: data.path,
            type,
            diff: data.diff,
            timestamp: msg.timestamp || new Date().toISOString(),
          };
          this.setState({ fileChanges: [...this.state.fileChanges, fileChange] });

          // Handle new file auto-open command to VS Code extension
          if (msg.type === 'file.created' && typeof (window as any).vscode !== 'undefined') {
            (window as any).vscode.postMessage({
              command: 'openFile',
              path: data.path,
            });
          }
        }
        break;

      case 'command.output':
        if (data?.output) {
          this.setState({ terminalLogs: [...this.state.terminalLogs, data.output] });
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
          this.setState({
            pendingApproval: data as ApprovalRequest,
            activeRunStatus: 'waiting',
          });
        }
        break;
    }
  }
}

export const agentStore = new AgentStore();
