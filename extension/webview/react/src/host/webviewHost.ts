import {
  AgentHost,
  HostInfo,
  NotificationLevel,
  OpenEditorsState,
  OpenFileOptions,
  WorkspaceState,
} from './agentHost';

interface VsCodeWebviewApi {
  postMessage(message: unknown): void;
  getState(): unknown;
  setState(state: unknown): void;
}

interface PendingRequest {
  resolve: (value: any) => void;
  reject: (error: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

/**
 * WebviewAgentHost implements AgentHost over the VS Code-family webview
 * bridge (acquireVsCodeApi().postMessage). Requests that need a response are
 * correlated via requestId/responseId pairs handled by the extension's
 * WebviewBridge.
 */
export class WebviewAgentHost implements AgentHost {
  readonly isEmbedded = true;
  private api: VsCodeWebviewApi;
  private pending: Map<string, PendingRequest> = new Map();
  private counter = 0;

  constructor(api: VsCodeWebviewApi) {
    this.api = api;
    window.addEventListener('message', (event) => this.handleMessage(event.data));
  }

  private handleMessage(data: any) {
    if (data && typeof data.responseId === 'string' && this.pending.has(data.responseId)) {
      const { resolve, reject, timer } = this.pending.get(data.responseId)!;
      this.pending.delete(data.responseId);
      clearTimeout(timer);
      if (data.success) {
        resolve(data.result);
      } else {
        reject(new Error(data.error || 'Host command failed'));
      }
    }
  }

  private request<T>(command: string, payload?: Record<string, unknown>): Promise<T> {
    const requestId = `req_${++this.counter}_${Date.now()}`;
    return new Promise<T>((resolve, reject) => {
      const timer = setTimeout(() => {
        if (this.pending.delete(requestId)) {
          reject(new Error(`Host command "${command}" timed out`));
        }
      }, 10000);
      this.pending.set(requestId, { resolve, reject, timer });
      this.api.postMessage({ command, payload, requestId });
    });
  }

  postMessage(message: unknown): void {
    this.api.postMessage(message);
  }

  openFile(path: string, options?: OpenFileOptions): void {
    this.postMessage({ command: 'openFile', payload: { path, line: options?.line } });
  }

  openDiff(path: string, rightPath?: string): void {
    this.postMessage({ command: 'openDiff', payload: { path, rightPath } });
  }

  revealFile(path: string): void {
    this.postMessage({ command: 'revealFile', payload: { path } });
  }

  createFile(path: string, content = ''): void {
    this.postMessage({ command: 'createFile', payload: { path, content } });
  }

  createFolder(path: string): void {
    this.postMessage({ command: 'createFolder', payload: { path } });
  }

  revealFolder(path: string): void {
    this.postMessage({ command: 'revealFolder', payload: { path } });
  }

  createTerminal(name?: string): void {
    this.postMessage({ command: 'createTerminal', payload: { name } });
  }

  runInTerminal(command: string, terminalName?: string): void {
    this.postMessage({ command: 'runInTerminal', payload: { command, terminalName } });
  }

  showNotification(message: string, level: NotificationLevel = 'info'): void {
    this.postMessage({ command: 'showNotification', payload: { message, type: level } });
  }

  getHostInfo(): Promise<HostInfo | undefined> {
    return this.request<HostInfo>('getHostInfo');
  }

  getWorkspaceState(): Promise<WorkspaceState> {
    return this.request<WorkspaceState>('getWorkspace');
  }

  getOpenEditors(): Promise<OpenEditorsState> {
    return this.request<OpenEditorsState>('getOpenEditors');
  }

  async saveSecret(key: string, value: string): Promise<void> {
    await this.request('saveSecret', { key, value });
  }

  async getSecret(key: string): Promise<string | undefined> {
    const res = await this.request<{ secret?: string }>('getSecret', { key });
    return res?.secret;
  }
}
