import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { AgentManager } from './agentManager';

export class AgentWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'agent.sidebar';
  private view?: vscode.WebviewView;
  private context: vscode.ExtensionContext;
  private agentManager: AgentManager;

  constructor(context: vscode.ExtensionContext, agentManager: AgentManager) {
    this.context = context;
    this.agentManager = agentManager;
  }

  public async resolveWebviewView(
    webviewView: vscode.WebviewView,
    context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ) {
    this.view = webviewView;

    const extensionUri = this.context.extensionUri;
    const webviewDistPath = path.join(extensionUri.fsPath, 'webview', 'react', 'dist');

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [extensionUri, vscode.Uri.file(webviewDistPath)],
    };

    // Give the backend a moment to allocate a port so the injected
    // __AGENT_PORT__ is correct; the live endpoint push below covers
    // the case where it isn't ready within the grace period.
    await Promise.race([
      this.agentManager.waitUntilReady(),
      new Promise((r) => setTimeout(r, 5000)),
    ]);

    webviewView.webview.html = this.getHtmlForWebview(webviewView.webview);

    // Forward webview messages to the host-backed bridge
    webviewView.webview.onDidReceiveMessage((message) => {
      this.agentManager.getWebviewBridge().handleWebviewMessage(message, webviewView.webview);
    });

    // Forward backend WebSocket events to React Webview
    const connection = this.agentManager.getConnection();
    const eventSub = connection.onEvent((evt) => {
      webviewView.webview.postMessage(evt);
    });

    // Push the authoritative endpoint whenever the connection reaches
    // 'ready' — covers webview resolving before the backend finished
    // starting, and backend restarts that allocate a different port.
    const pushEndpoint = () => {
      const port = connection.getPort();
      if (port > 0) {
        webviewView.webview.postMessage({
          type: 'agent.endpoint',
          data: { host: '127.0.0.1', port },
        });
      }
    };
    const stateSub = connection.onStateChange((state) => {
      if (state === 'ready') pushEndpoint();
    });
    if (connection.getState() === 'ready') pushEndpoint();

    webviewView.onDidDispose(() => {
      eventSub.dispose();
      stateSub.dispose();
    });
  }

  public showPanel() {
    const extensionUri = this.context.extensionUri;
    const webviewDistPath = path.join(extensionUri.fsPath, 'webview', 'react', 'dist');

    const panel = vscode.window.createWebviewPanel(
      'agentPanel',
      'BlackJak AI Agent',
      vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [extensionUri, vscode.Uri.file(webviewDistPath)],
      }
    );

    panel.webview.html = this.getHtmlForWebview(panel.webview);

    panel.webview.onDidReceiveMessage((message) => {
      this.agentManager.getWebviewBridge().handleWebviewMessage(message, panel.webview);
    });

    const connection = this.agentManager.getConnection();
    const eventSub = connection.onEvent((evt) => {
      panel.webview.postMessage(evt);
    });

    panel.onDidDispose(() => {
      eventSub.dispose();
    });
  }

  private getHtmlForWebview(webview: vscode.Webview): string {
    const extensionUri = this.context.extensionUri;
    const distPath = path.join(extensionUri.fsPath, 'webview', 'react', 'dist');
    const indexPath = path.join(distPath, 'index.html');

    const port = this.agentManager.getConnection().getPort();
    const host = '127.0.0.1';

    if (fs.existsSync(indexPath)) {
      let html = fs.readFileSync(indexPath, 'utf8');
      const baseUri = webview.asWebviewUri(vscode.Uri.file(distPath)).toString();

      // Convert relative asset links
      html = html.replace(/(src|href)="(\.\/|\/)?([^"]+)"/g, (match, p1, p2, p3) => {
        if (p3.startsWith('http') || p3.startsWith('//')) return match;
        return `${p1}="${baseUri}/${p3}"`;
      });

      // Inject server config & VS Code API bridge
      const scriptInjection = `<script>
        window.__AGENT_HOST__ = "${host}";
        window.__AGENT_PORT__ = ${port};
        if (typeof acquireVsCodeApi !== 'undefined') {
          window.vscode = acquireVsCodeApi();
        }
      </script>`;

      return html.replace('<head>', `<head>${scriptInjection}`);
    }

    // Dev fallback if dist files not yet compiled
    return `<!DOCTYPE html>
      <html lang="en">
      <head>
        <meta charset="UTF-8">
        <title>AI Agent</title>
        <script>
          window.__AGENT_HOST__ = "${host}";
          window.__AGENT_PORT__ = ${port};
          if (typeof acquireVsCodeApi !== 'undefined') {
            window.vscode = acquireVsCodeApi();
          }
        </script>
      </head>
      <body style="padding: 20px; font-family: sans-serif; color: #ccc; background-color: #1e1e1e;">
        <h2>BlackJak AI Agent</h2>
        <p>Connecting to backend at http://${host}:${port}...</p>
        <iframe src="http://localhost:5173" style="width: 100%; height: 600px; border: none;"></iframe>
      </body>
      </html>`;
  }
}
