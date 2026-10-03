import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { AgentManager } from './agentManager';

type PanelKind = 'agent' | 'settings' | 'activity';

export class AgentWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'agent.sidebar';
  private view?: vscode.WebviewView;
  private context: vscode.ExtensionContext;
  private agentManager: AgentManager;
  private panel: vscode.WebviewPanel | undefined;
  private settingsPanel: vscode.WebviewPanel | undefined;
  private activityPanel: vscode.WebviewPanel | undefined;

  constructor(context: vscode.ExtensionContext, agentManager: AgentManager) {
    this.context = context;
    this.agentManager = agentManager;
  }

  public async resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ) {
    this.view = webviewView;

    const extensionUri = this.context.extensionUri;
    const webviewDistPath = path.join(extensionUri.fsPath, 'webview', 'react', 'dist');

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [extensionUri, vscode.Uri.file(webviewDistPath)],
    };

    await Promise.race([
      this.agentManager.waitUntilReady(),
      new Promise((r) => setTimeout(r, 5000)),
    ]);

    webviewView.webview.html = this.getHtmlForWebview(webviewView.webview, 'agent');

    webviewView.webview.onDidReceiveMessage((message) => {
      this.agentManager.getWebviewBridge().handleWebviewMessage(message, webviewView.webview);
    });

    const connection = this.agentManager.getConnection();
    const eventSub = connection.onEvent((evt) => {
      webviewView.webview.postMessage(evt);
    });

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

  public showPanel(focus?: 'settings' | 'activity') {
    if (focus === 'settings') {
      this.openDedicatedPanel('settings');
      return;
    }
    if (focus === 'activity') {
      this.openDedicatedPanel('activity');
      return;
    }
    this.openOrRevealAgentPanel();
  }

  private openOrRevealAgentPanel() {
    if (!this.panel) {
      this.panel = this.createWiredPanel('agentPanel', 'BlackJak AI Agent', 'agent', () => {
        this.panel = undefined;
      });
    } else {
      this.panel.reveal(vscode.ViewColumn.One);
    }
  }

  private openDedicatedPanel(kind: 'settings' | 'activity') {
    const existing = kind === 'settings' ? this.settingsPanel : this.activityPanel;
    if (existing) {
      existing.reveal(vscode.ViewColumn.Beside);
      return;
    }

    const title = kind === 'settings' ? 'BlackJak Settings' : 'BlackJak Activity';
    const viewType = kind === 'settings' ? 'agentSettingsPanel' : 'agentActivityPanel';
    const panel = this.createWiredPanel(viewType, title, kind, () => {
      if (kind === 'settings') this.settingsPanel = undefined;
      else this.activityPanel = undefined;
    });
    if (kind === 'settings') this.settingsPanel = panel;
    else this.activityPanel = panel;
    panel.reveal(vscode.ViewColumn.Beside);
  }

  private createWiredPanel(
    viewType: string,
    title: string,
    kind: PanelKind,
    onDispose: () => void
  ): vscode.WebviewPanel {
    const extensionUri = this.context.extensionUri;
    const webviewDistPath = path.join(extensionUri.fsPath, 'webview', 'react', 'dist');
    const panel = vscode.window.createWebviewPanel(viewType, title, vscode.ViewColumn.Beside, {
      enableScripts: true,
      retainContextWhenHidden: true,
      localResourceRoots: [extensionUri, vscode.Uri.file(webviewDistPath)],
    });
    panel.webview.html = this.getHtmlForWebview(panel.webview, kind);
    panel.webview.onDidReceiveMessage((message) => {
      this.agentManager.getWebviewBridge().handleWebviewMessage(message, panel.webview);
    });
    const connection = this.agentManager.getConnection();
    const eventSub = connection.onEvent((evt) => {
      panel.webview.postMessage(evt);
    });
    const pushEndpoint = () => {
      const port = connection.getPort();
      if (port > 0) {
        panel.webview.postMessage({
          type: 'agent.endpoint',
          data: { host: '127.0.0.1', port },
        });
      }
    };
    const stateSub = connection.onStateChange((state) => {
      if (state === 'ready') pushEndpoint();
    });
    if (connection.getState() === 'ready') pushEndpoint();
    panel.onDidDispose(() => {
      eventSub.dispose();
      stateSub.dispose();
      onDispose();
    });
    return panel;
  }

  private getHtmlForWebview(webview: vscode.Webview, kind: PanelKind): string {
    const extensionUri = this.context.extensionUri;
    const distPath = path.join(extensionUri.fsPath, 'webview', 'react', 'dist');
    const indexPath = path.join(distPath, 'index.html');

    const port = this.agentManager.getConnection().getPort();
    const host = '127.0.0.1';
    const route = kind === 'settings' ? 'settings' : kind === 'activity' ? 'activity' : 'chat';

    if (fs.existsSync(indexPath)) {
      let html = fs.readFileSync(indexPath, 'utf8');
      const baseUri = webview.asWebviewUri(vscode.Uri.file(distPath)).toString();

      html = html.replace(/(src|href)="(\.\/|\/)?([^"]+)"/g, (match, p1, p2, p3) => {
        if (p3.startsWith('http') || p3.startsWith('//')) return match;
        return `${p1}="${baseUri}/${p3}"`;
      });

      const scriptInjection = `<script>
        window.__AGENT_HOST__ = "${host}";
        window.__AGENT_PORT__ = ${port};
        window.__BLACKJAK_ROUTE__ = "${route}";
        if (typeof acquireVsCodeApi !== 'undefined') {
          window.vscode = acquireVsCodeApi();
        }
      </script>`;

      return html.replace('<head>', `<head>${scriptInjection}`);
    }

    return `<!DOCTYPE html>
      <html lang="en">
      <head>
        <meta charset="UTF-8">
        <title>AI Agent</title>
        <script>
          window.__AGENT_HOST__ = "${host}";
          window.__AGENT_PORT__ = ${port};
          window.__BLACKJAK_ROUTE__ = "${route}";
          if (typeof acquireVsCodeApi !== 'undefined') {
            window.vscode = acquireVsCodeApi();
          }
        </script>
      </head>
      <body style="padding: 20px; font-family: sans-serif; color: #ccc; background-color: #1e1e1e;">
        <h2>BlackJak AI Agent</h2>
        <p>Connecting to backend at http://${host}:${port}...</p>
        <iframe src="http://localhost:5173/#/${route === 'chat' ? '' : route}" style="width: 100%; height: 600px; border: none;"></iframe>
      </body>
      </html>`;
  }
}
