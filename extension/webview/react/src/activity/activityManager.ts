import { ActivityCategory, ActivityConfig, ActivityPhase, ActivityState, ActivityStatus } from './activityTypes';
import { ACTIVITY_MESSAGES } from './messages';
import { ServerMessage } from '../types/events';

type Listener = (status: ActivityStatus) => void;

export class ActivityManager {
  private config: ActivityConfig;
  private state: ActivityState = 'idle';
  private category: ActivityCategory = 'orchestrating';
  private ambientMessage: string = '';
  private concreteMessage?: string;
  private lastMessage: string = '';
  private timer: any = null;
  private listeners: Set<Listener> = new Set();
  private workingSince?: number;
  private categorySince?: number;
  private phases: ActivityPhase[] = [];

  constructor(config?: Partial<ActivityConfig>) {
    this.config = {
      minIntervalMs: 3000,
      maxIntervalMs: 8000,
      enabled: true,
      ...config,
    };
  }

  public updateConfig(newConfig: Partial<ActivityConfig>) {
    this.config = { ...this.config, ...newConfig };
    if (!this.config.enabled) {
      this.clearTimer();
    } else if (this.state === 'working' && !this.timer) {
      this.scheduleNextRotation();
    }
  }

  public getStatus(): ActivityStatus {
    return {
      state: this.state,
      category: this.category,
      ambientMessage: this.ambientMessage,
      concreteMessage: this.concreteMessage,
      workingSince: this.workingSince,
      categorySince: this.categorySince,
      phases: [...this.phases],
      lastUpdated: new Date(),
    };
  }

  /** Switch the active category, recording how long the previous one ran. */
  private setCategory(cat: ActivityCategory) {
    if (cat === this.category || this.state !== 'working') {
      if (cat === this.category) return;
      this.category = cat;
      this.categorySince = Date.now();
      return;
    }
    if (this.categorySince) {
      const durationMs = Date.now() - this.categorySince;
      if (durationMs > 500) {
        this.phases.push({ category: this.category, message: this.ambientMessage, durationMs });
        if (this.phases.length > 8) this.phases = this.phases.slice(-8);
      }
    }
    this.category = cat;
    this.categorySince = Date.now();
  }

  public subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    listener(this.getStatus());
    return () => this.listeners.delete(listener);
  }

  private notify() {
    const status = this.getStatus();
    this.listeners.forEach((l) => l(status));
  }

  public reset() {
    this.clearTimer();
    this.state = 'idle';
    this.category = 'orchestrating';
    this.ambientMessage = '';
    this.concreteMessage = undefined;
    this.lastMessage = '';
    this.workingSince = undefined;
    this.categorySince = undefined;
    this.phases = [];
    this.notify();
  }

  public handleAgentEvent(msg: ServerMessage) {
    if (!msg.type) return;

    const data = msg.data;

    switch (msg.type) {
      case 'run.started':
      case 'run.resumed':
        this.state = 'working';
        this.workingSince = Date.now();
        this.phases = [];
        this.category = 'orchestrating';
        this.categorySince = Date.now();
        this.concreteMessage = msg.type === 'run.resumed' ? 'Resumed from checkpoint' : 'Task started';
        this.rotateAmbientMessage();
        this.scheduleNextRotation();
        break;

      case 'run.paused':
        this.state = 'idle';
        this.concreteMessage = 'Paused — context checkpointed';
        this.clearTimer();
        this.notify();
        break;

      case 'run.completed':
        this.state = 'completed';
        this.concreteMessage = 'Task completed successfully';
        this.clearTimer();
        this.notify();
        break;

      case 'run.failed':
        this.state = 'error';
        this.concreteMessage = data?.error || 'Task execution failed';
        this.clearTimer();
        this.notify();
        break;

      case 'run.cancelled':
        this.state = 'idle';
        this.concreteMessage = 'Task cancelled';
        this.clearTimer();
        this.notify();
        break;

      case 'approval.required':
        this.state = 'approval';
        this.concreteMessage = data?.description || 'Approval required for operation';
        this.clearTimer();
        this.notify();
        break;

      case 'approval.granted':
      case 'approval.denied':
        if (this.state === 'approval') {
          this.state = 'working';
          this.scheduleNextRotation();
        }
        break;

      case 'agent.thinking':
        this.state = 'working';
        this.setCategory('thinking');
        this.concreteMessage = data?.thought ? `Thinking: ${data.thought}` : 'Thinking…';
        this.notify();
        break;

      case 'agent.plan':
        this.state = 'working';
        this.setCategory('orchestrating');
        this.concreteMessage = 'Generating step plan';
        this.notify();
        break;

      case 'tool.started':
        this.state = 'working';
        this.concreteMessage = describeToolCall(data?.tool || '', data?.args);
        if (data?.tool === 'search' || (data?.tool === 'filesystem' && ['read', 'list', 'exists'].includes(data?.args?.operation))) {
          this.setCategory('exploring');
        } else if (data?.tool === 'test') {
          this.setCategory('testing');
        } else if (data?.tool === 'filesystem' || data?.tool === 'shell' || data?.tool === 'git') {
          this.setCategory('editing');
        }
        this.notify();
        break;

      case 'file.read':
        this.state = 'working';
        this.setCategory('exploring');
        if (data?.path) {
          const name = data.path.split(/[/\\]/).pop();
          this.concreteMessage = `Reading ${name}`;
        }
        this.notify();
        break;

      case 'file.modified':
      case 'file.created':
        this.state = 'working';
        this.setCategory('editing');
        if (data?.path) {
          const name = data.path.split(/[/\\]/).pop();
          this.concreteMessage = `Editing ${name}`;
        }
        this.notify();
        break;

      case 'test.started':
        this.state = 'working';
        this.setCategory('testing');
        this.concreteMessage = 'Running test suites';
        this.notify();
        break;

      case 'test.failed':
        this.state = 'working';
        this.setCategory('debugging');
        this.concreteMessage = 'Test failure detected';
        this.notify();
        break;

      case 'test.passed':
        this.state = 'working';
        this.setCategory('finishing');
        this.concreteMessage = 'Test passed cleanly';
        this.notify();
        break;
    }
  }

  private rotateAmbientMessage() {
    if (this.state !== 'working' || !this.config.enabled) {
      return;
    }

    const categoryPool = ACTIVITY_MESSAGES[this.category] || ACTIVITY_MESSAGES.orchestrating;

    let available = categoryPool.filter((m) => m !== this.lastMessage);
    if (available.length === 0) {
      available = categoryPool;
    }

    // Pseudo-random non-sequential selection
    const index = Math.floor(Math.random() * available.length);
    const selected = available[index];

    this.lastMessage = selected;
    this.ambientMessage = selected;
    this.notify();
  }

  private scheduleNextRotation() {
    this.clearTimer();
    if (this.state !== 'working' || !this.config.enabled) {
      return;
    }

    const { minIntervalMs, maxIntervalMs } = this.config;
    const delay = Math.floor(Math.random() * (maxIntervalMs - minIntervalMs + 1)) + minIntervalMs;

    this.timer = setTimeout(() => {
      this.rotateAmbientMessage();
      this.scheduleNextRotation();
    }, delay);
  }

  private clearTimer() {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }

  public dispose() {
    this.clearTimer();
    this.listeners.clear();
  }
}

export const activityManager = new ActivityManager();

/** Human-readable one-line description of a tool call, e.g. "Reading auth.ts". */
export function describeToolCall(tool: string, args?: any): string {
  const a = args || {};
  const shortPath = (p?: string) => {
    if (!p) return '';
    const parts = p.split(/[/\\]/);
    return parts.length > 2 ? `…/${parts.slice(-2).join('/')}` : p;
  };
  const trunc = (s: string, n: number) => (s.length > n ? s.slice(0, n - 1) + '…' : s);

  switch (tool) {
    case 'filesystem': {
      const op = a.operation || 'read';
      const verb = op === 'read' ? 'Reading' : op === 'write' ? 'Writing' : op === 'edit' ? 'Editing' : op === 'list' ? 'Listing' : 'Checking';
      return `${verb} ${shortPath(a.path) || 'files'}`;
    }
    case 'shell':
      return `Running \`${trunc(a.command || 'command', 56)}\``;
    case 'search':
      return `Searching for \`${trunc(a.pattern || '', 40)}\``;
    case 'git':
      return `git ${a.operation || 'status'}${a.args ? ` ${trunc(a.args, 30)}` : ''}`;
    case 'test':
      return `Running ${a.command ? trunc(a.command, 40) : 'test suite'}`;
    case 'memory':
      return `Memory ${a.operation || 'get'}${a.key ? `: ${a.key}` : ''}`;
    case 'delegate':
      return `Delegating to ${a.role || 'subagent'}`;
    case 'update_plan':
      return 'Updating plan';
    case 'task_complete':
      return 'Completing task';
    case 'ask_user':
      return 'Waiting for input';
    default:
      return `Running ${tool}`;
  }
}
