/**
 * IDEHost is the stable boundary between the extension and the concrete IDE.
 *
 * Everything the agent or the webview UI needs from an editor (files, tabs,
 * diffs, watchers, notifications, secrets) is expressed through this
 * interface. Host adapters (VSCodeHost today; a Theia or fork-specific
 * adapter only if a host genuinely diverges) implement it against the
 * public extension API — never against IDE internals.
 */

export type NotificationLevel = 'info' | 'warn' | 'error';

export interface HostDisposable {
  dispose(): void;
}

export type ViewColumnOption = 'active' | 'beside' | 'newGroup';

export interface OpenFileOptions {
  line?: number;
  preview?: boolean;
  preserveFocus?: boolean;
  viewColumn?: ViewColumnOption;
}

export interface EditorState {
  path: string;
  isDirty: boolean;
}

export type FileChangeKind = 'created' | 'changed' | 'deleted';
export type FileChangeHandler = (path: string, kind: FileChangeKind) => void;

export interface HostCapabilities {
  webview: boolean;
  editorTabs: boolean;
  fileWatcher: boolean;
  terminal: boolean;
  scm: boolean;
  diffEditor: boolean;
  secrets: boolean;
}

export interface HostInfo {
  /** Normalized host id: vscode | cursor | windsurf | vscodium | theia | unknown */
  name: string;
  /** Raw host application name, e.g. "Visual Studio Code" */
  displayName: string;
  /** Host application version (empty when not exposed) */
  version: string;
  /** Supported VS Code extension API level (vscode.version) */
  apiVersion: string;
  /** Installed extension version */
  extensionVersion: string;
  capabilities: HostCapabilities;
}

export interface IDEHost {
  readonly info: HostInfo;

  // Editor / files
  openFile(filePath: string, options?: OpenFileOptions): Promise<void>;
  openDiff(leftPath: string, rightPath?: string, title?: string): Promise<void>;
  revealFile(filePath: string): Promise<void>;
  createFile(filePath: string, content?: string): Promise<void>;
  readFile(filePath: string): Promise<string>;
  writeFile(filePath: string, content: string): Promise<void>;
  fileExists(filePath: string): boolean;
  isDocumentOpen(filePath: string): boolean;
  isDocumentDirty(filePath: string): boolean;
  saveDocument(filePath: string): Promise<boolean>;

  // Workspace / editor state
  getWorkspaceFolders(): string[];
  getActiveEditor(): EditorState | undefined;
  getOpenEditors(): EditorState[];

  // Events
  onDidChangeDocument(handler: (path: string, isDirty: boolean) => void): HostDisposable;
  onDidCloseDocument(handler: (path: string) => void): HostDisposable;
  onDidChangeActiveEditor(handler: (path: string | undefined) => void): HostDisposable;
  watchFiles(glob: string, handler: FileChangeHandler): HostDisposable;

  // UX
  showNotification(
    message: string,
    level: NotificationLevel,
    ...actions: string[]
  ): Promise<string | undefined>;
  executeCommand<T = unknown>(command: string, ...args: unknown[]): Promise<T | undefined>;

  // Secrets
  storeSecret(key: string, value: string): Promise<void>;
  getSecret(key: string): Promise<string | undefined>;
}
