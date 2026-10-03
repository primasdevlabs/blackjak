import * as vscode from 'vscode';
import { IDEHost } from './host/host';

/**
 * Routes webview UI messages to the host adapter.
 *
 * The webview speaks a host-neutral command vocabulary (openFile, openDiff,
 * getHostInfo, ...); this bridge translates each command into IDEHost calls,
 * so the React app never touches IDE APIs directly and stays portable across
 * every VS Code-family host.
 */
export class WebviewBridge {
  private host: IDEHost;
  private outputChannel: vscode.OutputChannel;

  constructor(host: IDEHost, outputChannel: vscode.OutputChannel) {
    this.host = host;
    this.outputChannel = outputChannel;
  }

  public async handleWebviewMessage(message: any, webview: vscode.Webview): Promise<void> {
    const { command, payload = {}, requestId } = message;

    try {
      let result: any = undefined;
      switch (command) {
        case 'openFile':
          await this.host.openFile(payload.path, { line: payload.line });
          break;
        case 'openDiff':
          await this.host.openDiff(payload.leftPath || payload.path, payload.rightPath);
          break;
        case 'revealFile':
          await this.host.revealFile(payload.path);
          break;
        case 'createFile':
          await this.host.createFile(payload.path, payload.content);
          break;
        case 'readFile':
          result = { content: await this.host.readFile(payload.path) };
          break;
        case 'writeFile':
          await this.host.writeFile(payload.path, payload.content ?? '');
          break;
        case 'showNotification':
          await this.host.showNotification(payload.message, payload.type);
          break;
        case 'saveSecret':
          await this.host.storeSecret(payload.key, payload.value);
          result = { success: true };
          break;
        case 'getSecret':
          result = { secret: await this.host.getSecret(payload.key) };
          break;
        case 'getHostInfo':
          result = this.host.info;
          break;
        case 'getWorkspace':
          result = { folders: this.host.getWorkspaceFolders() };
          break;
        case 'getOpenEditors':
          result = { editors: this.host.getOpenEditors(), active: this.host.getActiveEditor() };
          break;
        case 'validatePath': {
          const folders = this.host.getWorkspaceFolders();
          const normalized = payload.path?.replace(/\\/g, '/') || '';
          const withinWorkspace = folders.some((folder: string) => {
            const normalizedFolder = folder.replace(/\\/g, '/');
            return normalized.startsWith(normalizedFolder);
          });
          result = { valid: withinWorkspace, resolvedPath: normalized };
          break;
        }
        case 'revertFile':
          try {
            const uri = vscode.Uri.file(payload.path);
            const terminal = vscode.window.createTerminal({ name: 'Agent: revert' });
            terminal.sendText(`git checkout -- "${payload.path}"`, true);
            // Reopen the file to show reverted content
            setTimeout(async () => {
              const doc = await vscode.workspace.openTextDocument(uri);
              await vscode.window.showTextDocument(doc, { preview: false });
            }, 1000);
          } catch (revertErr: any) {
            this.outputChannel.appendLine(`[WebviewBridge] Revert failed: ${revertErr.message}`);
          }
          break;
        case 'createFolder':
          await vscode.workspace.fs.createDirectory(vscode.Uri.file(payload.path));
          break;
        case 'revealFolder':
          await vscode.commands.executeCommand('revealInExplorer', vscode.Uri.file(payload.path));
          break;
        case 'createTerminal': {
          const terminal = vscode.window.createTerminal({ name: payload.name || 'Agent' });
          terminal.show();
          break;
        }
        case 'runInTerminal': {
          let terminal = vscode.window.terminals.find((t) => t.name === (payload.terminalName || 'Agent'));
          if (!terminal) {
            terminal = vscode.window.createTerminal({ name: payload.terminalName || 'Agent' });
          }
          terminal.show();
          terminal.sendText(payload.command, true);
          break;
        }
        case 'pickFiles':
          result = { paths: await this.host.pickFiles(payload) };
          break;
        case 'openSettingsWindow':
          await this.host.openSettingsWindow();
          break;
        case 'openActivityPanel':
          await this.host.openActivityPanel();
          break;
        case 'closePanel':
          // Dedicated settings/activity panels close via dispose of the active tab.
          await vscode.commands.executeCommand('workbench.action.closeActiveEditor');
          break;
        default:
          this.outputChannel.appendLine(`[WebviewBridge] Unknown command: ${command}`);
      }

      if (requestId) {
        webview.postMessage({ responseId: requestId, success: true, result });
      }
    } catch (err: any) {
      this.outputChannel.appendLine(`[WebviewBridge Error] Command ${command} failed: ${err.message}`);
      if (requestId) {
        webview.postMessage({ responseId: requestId, success: false, error: err.message });
      }
    }
  }
}
