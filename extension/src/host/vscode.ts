import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import {
  EditorState,
  FileChangeHandler,
  HostDisposable,
  HostInfo,
  IDEHost,
  NotificationLevel,
  OpenFileOptions,
} from './host';
import { detectHostName } from './compatibility';
import { detectCapabilities } from './capabilities';

/**
 * VSCodeHost implements IDEHost against the public VS Code extension API.
 *
 * Cursor, Windsurf and VSCodium expose the same extension surface, so they
 * are served by this adapter. Hosts with genuinely different surfaces (e.g.
 * a future Theia-specific adapter) get their own IDEHost implementation —
 * never branching inside agent logic.
 */
export class VSCodeHost implements IDEHost {
  readonly info: HostInfo;
  private context: vscode.ExtensionContext;

  private constructor(context: vscode.ExtensionContext, info: HostInfo) {
    this.context = context;
    this.info = info;
  }

  static async create(context: vscode.ExtensionContext): Promise<VSCodeHost> {
    const capabilities = await detectCapabilities(context);
    const info: HostInfo = {
      name: detectHostName(),
      displayName: vscode.env.appName || 'unknown',
      version: '',
      apiVersion: vscode.version,
      extensionVersion: String(context.extension?.packageJSON?.version ?? ''),
      capabilities,
    };
    return new VSCodeHost(context, info);
  }

  async openFile(filePath: string, options?: OpenFileOptions): Promise<void> {
    if (!filePath) {
      return;
    }
    const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(filePath));
    const editor = await vscode.window.showTextDocument(doc, {
      viewColumn: this.toViewColumn(options?.viewColumn),
      preview: options?.preview ?? true,
      preserveFocus: options?.preserveFocus ?? false,
    });

    if (options?.line && options.line > 0) {
      const pos = new vscode.Position(options.line - 1, 0);
      editor.selection = new vscode.Selection(pos, pos);
      editor.revealRange(new vscode.Range(pos, pos), vscode.TextEditorRevealType.InCenter);
    }
  }

  async openDiff(leftPath: string, rightPath?: string, title?: string): Promise<void> {
    if (!this.info.capabilities.diffEditor) {
      await this.openFile(leftPath);
      return;
    }
    const leftUri = vscode.Uri.file(leftPath);
    const rightUri = rightPath ? vscode.Uri.file(rightPath) : leftUri;
    const label =
      title ??
      (rightPath
        ? `Diff: ${path.basename(leftPath)} ↔ ${path.basename(rightPath)}`
        : `Diff: ${path.basename(leftPath)}`);
    await vscode.commands.executeCommand('vscode.diff', leftUri, rightUri, label);
  }

  async revealFile(filePath: string): Promise<void> {
    await vscode.commands.executeCommand('revealInExplorer', vscode.Uri.file(filePath));
  }

  async createFile(filePath: string, content = ''): Promise<void> {
    const dir = path.dirname(filePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    fs.writeFileSync(filePath, content, 'utf8');
    await this.openFile(filePath);
  }

  async readFile(filePath: string): Promise<string> {
    const bytes = await vscode.workspace.fs.readFile(vscode.Uri.file(filePath));
    return Buffer.from(bytes).toString('utf8');
  }

  async writeFile(filePath: string, content: string): Promise<void> {
    const dir = path.dirname(filePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    await vscode.workspace.fs.writeFile(vscode.Uri.file(filePath), Buffer.from(content, 'utf8'));
  }

  fileExists(filePath: string): boolean {
    return fs.existsSync(filePath);
  }

  isDocumentOpen(filePath: string): boolean {
    return vscode.workspace.textDocuments.some((d) => d.uri.fsPath === filePath);
  }

  isDocumentDirty(filePath: string): boolean {
    const doc = vscode.workspace.textDocuments.find((d) => d.uri.fsPath === filePath);
    return !!doc?.isDirty;
  }

  async saveDocument(filePath: string): Promise<boolean> {
    const doc = vscode.workspace.textDocuments.find((d) => d.uri.fsPath === filePath);
    if (!doc) {
      return false;
    }
    return doc.save();
  }

  getWorkspaceFolders(): string[] {
    const folders = vscode.workspace.workspaceFolders;
    if (!folders || folders.length === 0) {
      return [];
    }
    return folders.map((f) => f.uri.fsPath);
  }

  getActiveEditor(): EditorState | undefined {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.uri.scheme !== 'file') {
      return undefined;
    }
    return { path: editor.document.uri.fsPath, isDirty: editor.document.isDirty };
  }

  getOpenEditors(): EditorState[] {
    return vscode.workspace.textDocuments
      .filter((d) => d.uri.scheme === 'file')
      .map((d) => ({ path: d.uri.fsPath, isDirty: d.isDirty }));
  }

  onDidChangeDocument(handler: (path: string, isDirty: boolean) => void): HostDisposable {
    return vscode.workspace.onDidChangeTextDocument((e) => {
      if (e.document.uri.scheme === 'file') {
        handler(e.document.uri.fsPath, e.document.isDirty);
      }
    });
  }

  onDidCloseDocument(handler: (path: string) => void): HostDisposable {
    return vscode.workspace.onDidCloseTextDocument((doc) => {
      if (doc.uri.scheme === 'file') {
        handler(doc.uri.fsPath);
      }
    });
  }

  onDidChangeActiveEditor(handler: (path: string | undefined) => void): HostDisposable {
    return vscode.window.onDidChangeActiveTextEditor((editor) => {
      handler(editor && editor.document.uri.scheme === 'file' ? editor.document.uri.fsPath : undefined);
    });
  }

  watchFiles(glob: string, handler: FileChangeHandler): HostDisposable {
    if (!this.info.capabilities.fileWatcher) {
      return { dispose: () => {} };
    }
    const watcher = vscode.workspace.createFileSystemWatcher(glob);
    watcher.onDidCreate((uri) => handler(uri.fsPath, 'created'));
    watcher.onDidChange((uri) => handler(uri.fsPath, 'changed'));
    watcher.onDidDelete((uri) => handler(uri.fsPath, 'deleted'));
    return watcher;
  }

  async showNotification(
    message: string,
    level: NotificationLevel = 'info',
    ...actions: string[]
  ): Promise<string | undefined> {
    switch (level) {
      case 'error':
        return vscode.window.showErrorMessage(message, ...actions);
      case 'warn':
        return vscode.window.showWarningMessage(message, ...actions);
      default:
        return vscode.window.showInformationMessage(message, ...actions);
    }
  }

  async executeCommand<T = unknown>(command: string, ...args: unknown[]): Promise<T | undefined> {
    return vscode.commands.executeCommand<T>(command, ...args);
  }

  async storeSecret(key: string, value: string): Promise<void> {
    if (!this.info.capabilities.secrets) {
      throw new Error('Secret storage is not supported by this host');
    }
    await this.context.secrets.store(key, value);
  }

  async getSecret(key: string): Promise<string | undefined> {
    if (!this.info.capabilities.secrets) {
      return undefined;
    }
    return this.context.secrets.get(key);
  }

  private toViewColumn(option?: OpenFileOptions['viewColumn']): vscode.ViewColumn {
    switch (option) {
      case 'active':
        return vscode.ViewColumn.Active;
      case 'newGroup':
        return vscode.ViewColumn.Three;
      case 'beside':
      default:
        return vscode.ViewColumn.Beside;
    }
  }
}
