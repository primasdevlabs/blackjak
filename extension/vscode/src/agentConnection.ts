import WebSocket from 'ws';
import * as http from 'http';
import * as vscode from 'vscode';

export type ConnectionState =
  | 'starting'
  | 'starting_backend'
  | 'connecting'
  | 'initializing'
  | 'ready'
  | 'reconnecting'
  | 'stopping'
  | 'stopped'
  | 'error';

export interface InitialStateData {
  type: string;
  data: {
    workspace: { root: string };
    settings: any;
    providers: any;
    models: any;
    runs: any[];
    queue: any[];
    capabilities: any;
  };
}

export class AgentConnection {
  private ws: WebSocket | null = null;
  private port: number = 0;
  private state: ConnectionState = 'stopped';
  private outputChannel: vscode.OutputChannel;
  private listeners: Set<(state: ConnectionState) => void> = new Set();
  private eventListeners: Set<(event: any) => void> = new Set();
  private reconnectAttempt: number = 0;
  private maxReconnectDelayMs: number = 16000;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private isIntentionallyStopped: boolean = false;

  constructor(outputChannel: vscode.OutputChannel) {
    this.outputChannel = outputChannel;
  }

  public getState(): ConnectionState {
    return this.state;
  }

  public getPort(): number {
    return this.port;
  }

  public setState(newState: ConnectionState) {
    this.state = newState;
    this.outputChannel.appendLine(`[AgentConnection] State -> ${newState}`);
    this.listeners.forEach((l) => l(newState));
  }

  public onStateChange(listener: (state: ConnectionState) => void): vscode.Disposable {
    this.listeners.add(listener);
    return new vscode.Disposable(() => this.listeners.delete(listener));
  }

  public onEvent(listener: (event: any) => void): vscode.Disposable {
    this.eventListeners.add(listener);
    return new vscode.Disposable(() => this.eventListeners.delete(listener));
  }

  public async connect(port: number): Promise<void> {
    this.port = port;
    this.isIntentionallyStopped = false;
    this.setState('connecting');

    return new Promise((resolve, reject) => {
      const wsUrl = `ws://127.0.0.1:${port}/ws`;
      this.outputChannel.appendLine(`[AgentConnection] Connecting WebSocket to ${wsUrl}`);

      const socket = new WebSocket(wsUrl);
      this.ws = socket;

      socket.on('open', async () => {
        this.outputChannel.appendLine('[AgentConnection] WebSocket connected!');
        this.reconnectAttempt = 0;
        this.setState('initializing');

        try {
          // Single Initial State Snapshot Handshake
          const snapshot = await this.fetchInitialState();
          this.setState('ready');
          this.broadcastEvent(snapshot);
          resolve();
        } catch (err: any) {
          this.outputChannel.appendLine(`[AgentConnection] Initial state fetch error: ${err.message}`);
          this.setState('ready'); // Fallback to ready
          resolve();
        }
      });

      socket.on('message', (data: Buffer) => {
        try {
          const parsed = JSON.parse(data.toString());
          this.broadcastEvent(parsed);
        } catch (_) {}
      });

      socket.on('error', (err: any) => {
        this.outputChannel.appendLine(`[AgentConnection Error] ${err.message}`);
        if (this.state === 'connecting') {
          reject(err);
        }
      });

      socket.on('close', () => {
        this.outputChannel.appendLine('[AgentConnection] WebSocket closed');
        this.ws = null;
        if (!this.isIntentionallyStopped) {
          this.scheduleReconnect();
        }
      });
    });
  }

  public disconnect(): void {
    this.isIntentionallyStopped = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.setState('stopped');
  }

  private scheduleReconnect(): void {
    this.setState('reconnecting');
    this.reconnectAttempt++;
    const backoffMs = Math.min(Math.pow(2, this.reconnectAttempt - 1) * 1000, this.maxReconnectDelayMs);

    this.outputChannel.appendLine(`[AgentConnection] Reconnecting in ${backoffMs}ms (Attempt ${this.reconnectAttempt})...`);
    this.reconnectTimer = setTimeout(() => {
      if (!this.isIntentionallyStopped && this.port > 0) {
        this.connect(this.port).catch(() => {});
      }
    }, backoffMs);
  }

  public async fetchInitialState(): Promise<InitialStateData> {
    return new Promise((resolve, reject) => {
      const req = http.get(`http://127.0.0.1:${this.port}/api/initial-state`, (res) => {
        let body = '';
        res.on('data', (c) => (body += c));
        res.on('end', () => {
          try {
            const json = JSON.parse(body);
            resolve(json.data || json);
          } catch (e) {
            reject(e);
          }
        });
      });
      req.on('error', (e) => reject(e));
    });
  }

  private broadcastEvent(event: any): void {
    this.eventListeners.forEach((l) => l(event));
  }
}
