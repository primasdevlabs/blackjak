import { ActivityCategory, ActivityConfig, ActivityState, ActivityStatus } from './activityTypes';
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
      lastUpdated: new Date(),
    };
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
    this.notify();
  }

  public handleAgentEvent(msg: ServerMessage) {
    if (!msg.type) return;

    const data = msg.data;

    switch (msg.type) {
      case 'run.started':
        this.state = 'working';
        this.category = 'orchestrating';
        this.concreteMessage = 'Task started';
        this.rotateAmbientMessage();
        this.scheduleNextRotation();
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
        this.category = 'thinking';
        this.concreteMessage = data?.thought ? `Thinking: ${data.thought}` : undefined;
        this.notify();
        break;

      case 'agent.plan':
        this.state = 'working';
        this.category = 'orchestrating';
        this.concreteMessage = 'Generating step plan';
        this.notify();
        break;

      case 'tool.started':
        this.state = 'working';
        if (data?.tool === 'search') {
          this.category = 'exploring';
          this.concreteMessage = `Searching codebase (${data.step || ''})`;
        } else if (data?.tool === 'filesystem') {
          this.category = 'exploring';
          this.concreteMessage = `Accessing filesystem (${data.step || ''})`;
        } else if (data?.tool === 'test') {
          this.category = 'testing';
          this.concreteMessage = 'Running test suite';
        } else if (data?.tool === 'shell' || data?.tool === 'git') {
          this.category = 'editing';
          this.concreteMessage = `Executing ${data.tool} action`;
        }
        this.notify();
        break;

      case 'file.read':
        this.state = 'working';
        this.category = 'exploring';
        if (data?.path) {
          const name = data.path.split(/[/\\]/).pop();
          this.concreteMessage = `Reading ${name}`;
        }
        this.notify();
        break;

      case 'file.modified':
      case 'file.created':
        this.state = 'working';
        this.category = 'editing';
        if (data?.path) {
          const name = data.path.split(/[/\\]/).pop();
          this.concreteMessage = `Editing ${name}`;
        }
        this.notify();
        break;

      case 'test.started':
        this.state = 'working';
        this.category = 'testing';
        this.concreteMessage = 'Running test suites';
        this.notify();
        break;

      case 'test.failed':
        this.state = 'working';
        this.category = 'debugging';
        this.concreteMessage = 'Test failure detected';
        this.notify();
        break;

      case 'test.passed':
        this.state = 'working';
        this.category = 'finishing';
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
