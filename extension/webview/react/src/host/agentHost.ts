/**
 * AgentHost is the host-neutral contract between the React UI and the
 * surrounding IDE. The UI never references vscode.* or any host-specific
 * global — it only speaks this vocabulary. The extension host layer
 * translates these calls into concrete IDE operations, so the same UI runs
 * unchanged in VS Code, Cursor, Windsurf, VSCodium, Theia, or a plain
 * browser during development.
 */

export type NotificationLevel = 'info' | 'warn' | 'error';

export interface HostCapabilities {
  webview: boolean;
  editorTabs: boolean;
  fileWatcher: boolean;
  terminal: boolean;
  scm: boolean;
  diffEditor: boolean;
  secrets: boolean;
  filePicker: boolean;
  multiPanel: boolean;
  settingsWindow: boolean;
  activityPanel: boolean;
}

export interface PickFilesOptions {
  canSelectMany?: boolean;
  canSelectFolders?: boolean;
  title?: string;
  /** Map of label → extensions, e.g. { Images: ['png','jpg'] } */
  filters?: Record<string, string[]>;
}

export interface HostInfo {
  /** Normalized host id: vscode | cursor | windsurf | vscodium | theia | unknown */
  name: string;
  displayName: string;
  version: string;
  apiVersion: string;
  extensionVersion: string;
  capabilities: HostCapabilities;
}

export interface OpenFileOptions {
  line?: number;
}

export interface WorkspaceState {
  folders: string[];
}

export interface EditorState {
  path: string;
  isDirty: boolean;
}

export interface OpenEditorsState {
  editors: EditorState[];
  active?: EditorState;
}

export interface AgentHost {
  /** True when running inside a real IDE webview. */
  readonly isEmbedded: boolean;

  // Editor / files
  openFile(path: string, options?: OpenFileOptions): void;
  openDiff(path: string, rightPath?: string): void;
  revealFile(path: string): void;
  createFile(path: string, content?: string): void;
  createFolder(path: string): void;
  revealFolder(path: string): void;

  // Terminal
  createTerminal(name?: string): void;
  runInTerminal(command: string, terminalName?: string): void;

  // UX
  showNotification(message: string, level?: NotificationLevel): void;

  // State queries
  getHostInfo(): Promise<HostInfo | undefined>;
  getWorkspaceState(): Promise<WorkspaceState>;
  getOpenEditors(): Promise<OpenEditorsState>;

  // Secrets
  saveSecret(key: string, value: string): Promise<void>;
  getSecret(key: string): Promise<string | undefined>;

  // Panels / pickers
  pickFiles(options?: PickFilesOptions): Promise<string[]>;
  openSettingsWindow(): void;
  openActivityPanel(): void;

  /** Raw channel for host commands not covered by the typed surface. */
  postMessage(message: unknown): void;
}
