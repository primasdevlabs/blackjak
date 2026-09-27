import {
  AgentHost,
  HostInfo,
  NotificationLevel,
  OpenEditorsState,
  OpenFileOptions,
  WorkspaceState,
} from './agentHost';

/**
 * BrowserAgentHost is the development fallback used when the UI runs outside
 * an IDE webview (e.g. `npm run dev` on localhost). Host operations become
 * no-ops so the UI remains usable against a standalone agent server.
 */
export class BrowserAgentHost implements AgentHost {
  readonly isEmbedded = false;
  private secrets: Map<string, string> = new Map();

  postMessage(message: unknown): void {
    console.debug('[BrowserHost] postMessage (no IDE host attached):', message);
  }

  openFile(path: string, options?: OpenFileOptions): void {
    console.debug('[BrowserHost] openFile:', path, options);
  }

  openDiff(path: string, rightPath?: string): void {
    console.debug('[BrowserHost] openDiff:', path, rightPath);
  }

  revealFile(path: string): void {
    console.debug('[BrowserHost] revealFile:', path);
  }

  createFile(path: string, content = ''): void {
    console.debug('[BrowserHost] createFile:', path, `${content.length} bytes`);
  }

  createFolder(path: string): void {
    console.debug('[BrowserHost] createFolder:', path);
  }

  revealFolder(path: string): void {
    console.debug('[BrowserHost] revealFolder:', path);
  }

  createTerminal(name?: string): void {
    console.debug('[BrowserHost] createTerminal:', name);
  }

  runInTerminal(command: string, terminalName?: string): void {
    console.debug('[BrowserHost] runInTerminal:', command, terminalName);
  }

  showNotification(message: string, level: NotificationLevel = 'info'): void {
    console.log(`[BrowserHost] ${level.toUpperCase()}: ${message}`);
  }

  async getHostInfo(): Promise<HostInfo | undefined> {
    return {
      name: 'browser',
      displayName: 'Browser (dev)',
      version: '',
      apiVersion: '',
      extensionVersion: '',
      capabilities: {
        webview: false,
        editorTabs: false,
        fileWatcher: false,
        terminal: false,
        scm: false,
        diffEditor: false,
        secrets: false,
      },
    };
  }

  async getWorkspaceState(): Promise<WorkspaceState> {
    return { folders: [] };
  }

  async getOpenEditors(): Promise<OpenEditorsState> {
    return { editors: [] };
  }

  async saveSecret(key: string, value: string): Promise<void> {
    this.secrets.set(key, value);
  }

  async getSecret(key: string): Promise<string | undefined> {
    return this.secrets.get(key);
  }
}
