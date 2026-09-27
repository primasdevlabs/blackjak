import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

export class IDEBridge {
  private context: vscode.ExtensionContext;
  private outputChannel: vscode.OutputChannel;

  constructor(context: vscode.ExtensionContext, outputChannel: vscode.OutputChannel) {
    this.context = context;
    this.outputChannel = outputChannel;
  }

  public async handleWebviewMessage(message: any, webview: vscode.Webview): Promise<void> {
    const { command, payload, requestId } = message;

    try {
      let result: any = null;
      switch (command) {
        case 'openFile':
          await this.openFile(payload.path, payload.line);
          break;
        case 'openDiff':
          await this.openDiff(payload.leftPath || payload.path, payload.rightPath);
          break;
        case 'revealFile':
          await this.revealFile(payload.path);
          break;
        case 'createFile':
          await this.createFile(payload.path, payload.content);
          break;
        case 'showNotification':
          this.showNotification(payload.message, payload.type);
          break;
        case 'saveSecret':
          await this.context.secrets.store(payload.key, payload.value);
          result = { success: true };
          break;
        case 'getSecret':
          const secret = await this.context.secrets.get(payload.key);
          result = { secret };
          break;
        default:
          this.outputChannel.appendLine(`[IDEBridge] Unknown command: ${command}`);
      }

      if (requestId) {
        webview.postMessage({ responseId: requestId, success: true, result });
      }
    } catch (err: any) {
      this.outputChannel.appendLine(`[IDEBridge Error] Command ${command} failed: ${err.message}`);
      if (requestId) {
        webview.postMessage({ responseId: requestId, success: false, error: err.message });
      }
    }
  }

  public async openFile(filePath: string, line?: number): Promise<void> {
    if (!filePath) return;
    const doc = await vscode.workspace.openTextDocument(filePath);
    const editor = await vscode.window.showTextDocument(doc, vscode.ViewColumn.One);

    if (line && line > 0) {
      const pos = new vscode.Position(line - 1, 0);
      editor.selection = new vscode.Selection(pos, pos);
      editor.revealRange(new vscode.Range(pos, pos), vscode.TextEditorRevealType.InCenter);
    }
  }

  public async openDiff(leftPath: string, rightPath?: string): Promise<void> {
    const leftUri = vscode.Uri.file(leftPath);
    if (rightPath) {
      const rightUri = vscode.Uri.file(rightPath);
      await vscode.commands.executeCommand('vscode.diff', leftUri, rightUri, `Diff: ${path.basename(leftPath)} ↔ ${path.basename(rightPath)}`);
    } else {
      // Git diff against HEAD if rightPath not specified
      await vscode.commands.executeCommand('vscode.diff', leftUri, leftUri, `Diff: ${path.basename(leftPath)}`);
    }
  }

  public async revealFile(filePath: string): Promise<void> {
    const uri = vscode.Uri.file(filePath);
    await vscode.commands.executeCommand('revealInExplorer', uri);
  }

  public async createFile(filePath: string, content: string = ''): Promise<void> {
    const dir = path.dirname(filePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    fs.writeFileSync(filePath, content, 'utf8');
    await this.openFile(filePath);
  }

  public showNotification(message: string, type: 'info' | 'warn' | 'error' = 'info'): void {
    switch (type) {
      case 'error':
        vscode.window.showErrorMessage(message);
        break;
      case 'warn':
        vscode.window.showWarningMessage(message);
        break;
      default:
        vscode.window.showInformationMessage(message);
    }
  }
}
