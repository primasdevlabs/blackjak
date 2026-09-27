import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { openFileInEditor, showNativeDiff } from './workspace';

export class AgentWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'agent.view';
  private view?: vscode.WebviewView;
  private extensionUri: vscode.Uri;
  private port: number;
  private host: string;

  constructor(extensionUri: vscode.Uri, host: string, port: number) {
    this.extensionUri = extensionUri;
    this.host = host;
    this.port = port;
  }

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ) {
    this.view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [
        this.extensionUri,
        vscode.Uri.file(path.join(this.extensionUri.fsPath, '..', '..', 'ui', 'dist')),
      ],
    };

    webviewView.webview.html = this.getHtmlForWebview(webviewView.webview);

    webviewView.webview.onDidReceiveMessage(async (message) => {
      switch (message.command) {
        case 'openFile':
          if (message.path) {
            await openFileInEditor(message.path);
          }
          break;

        case 'showDiff':
          if (message.path) {
            await showNativeDiff(message.path, message.diff);
          }
          break;
      }
    });
  }

  public showPanel() {
    const panel = vscode.window.createWebviewPanel(
      'agentPanel',
      'AI Agent',
      vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [
          this.extensionUri,
          vscode.Uri.file(path.join(this.extensionUri.fsPath, '..', '..', 'ui', 'dist')),
        ],
      }
    );

    panel.webview.html = this.getHtmlForWebview(panel.webview);

    panel.webview.onDidReceiveMessage(async (message) => {
      switch (message.command) {
        case 'openFile':
          if (message.path) {
            await openFileInEditor(message.path);
          }
          break;

        case 'showDiff':
          if (message.path) {
            await showNativeDiff(message.path, message.diff);
          }
          break;
      }
    });
  }

  private getHtmlForWebview(webview: vscode.Webview): string {
    const uiDistPath = path.join(this.extensionUri.fsPath, '..', '..', 'ui', 'dist');
    const indexPath = path.join(uiDistPath, 'index.html');

    if (fs.existsSync(indexPath)) {
      let html = fs.readFileSync(indexPath, 'utf8');

      // Convert asset references to webview URIs
      const baseUri = webview.asWebviewUri(vscode.Uri.file(uiDistPath)).toString();
      html = html.replace(/(src|href)="(\.\/|\/)?([^"]+)"/g, (match, p1, p2, p3) => {
        if (p3.startsWith('http') || p3.startsWith('//')) return match;
        return `${p1}="${baseUri}/${p3}"`;
      });

      // Inject server host & port into window object
      const scriptInjection = `<script>
        window.__AGENT_HOST__ = "${this.host}";
        window.__AGENT_PORT__ = ${this.port};
        if (typeof acquireVsCodeApi !== 'undefined') {
          window.vscode = acquireVsCodeApi();
        }
      </script>`;

      return html.replace('<head>', `<head>${scriptInjection}`);
    }

    // Fallback if ui/dist has not been built yet
    return `<!DOCTYPE html>
      <html lang="en">
      <head>
        <meta charset="UTF-8">
        <title>AI Agent</title>
        <script>
          window.__AGENT_HOST__ = "${this.host}";
          window.__AGENT_PORT__ = ${this.port};
          if (typeof acquireVsCodeApi !== 'undefined') {
            window.vscode = acquireVsCodeApi();
          }
        </script>
      </head>
      <body style="padding: 20px; font-family: sans-serif; color: #ccc; background-color: #1e1e1e;">
        <h2>AI Agent (Development Mode)</h2>
        <p>Connecting to backend at http://${this.host}:${this.port}...</p>
        <iframe src="http://localhost:5173" style="width: 100%; height: 600px; border: none;"></iframe>
      </body>
      </html>`;
  }
}
