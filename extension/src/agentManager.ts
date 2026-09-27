import * as vscode from 'vscode';
import { AgentProcessManager } from './agentProcess';
import { HealthCheckProbe } from './healthCheck';
import { AgentConnection, ConnectionState } from './agentConnection';
import { WebviewBridge } from './webviewBridge';
import { IDEHost } from './host/host';

import { FileChangeTracker } from './fileChangeTracker';
import { TabManager } from './tabManager';

export class AgentManager {
  private context: vscode.ExtensionContext;
  private host: IDEHost;
  private processManager: AgentProcessManager;
  private connection: AgentConnection;
  private webviewBridge: WebviewBridge;
  private fileChangeTracker: FileChangeTracker;
  private tabManager: TabManager;
  private outputChannel: vscode.OutputChannel;
  private statusBarItem: vscode.StatusBarItem;
  private readyPromise: Promise<void> | null = null;
  private readyResolve: (() => void) | null = null;

  constructor(context: vscode.ExtensionContext, host: IDEHost) {
    this.context = context;
    this.host = host;
    this.outputChannel = vscode.window.createOutputChannel('AI Agent');
    this.processManager = new AgentProcessManager(this.outputChannel);
    this.connection = new AgentConnection(this.outputChannel);
    this.webviewBridge = new WebviewBridge(host, this.outputChannel);
    this.fileChangeTracker = new FileChangeTracker(host);
    this.tabManager = new TabManager(this.fileChangeTracker, host);

    this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    this.statusBarItem.command = 'agent.openSidebar';
    this.updateStatusBar('starting');

    this.connection.onStateChange((state) => {
      this.updateStatusBar(state);
    });

    this.connection.onEvent((evt: any) => {
      if (evt && (evt.type === 'file.changed' || evt.type === 'file.created') && evt.data) {
        this.tabManager.handleFileChangeEvent({
          taskId: evt.data.taskId || evt.runId || 'default_task',
          agentId: evt.data.agentId || '',
          path: evt.data.path || '',
          previousPath: evt.data.previousPath,
          changeType: evt.data.changeType || 'modified',
          state: evt.data.state,
          role: evt.data.role,
          importance: evt.data.importance,
          isPreExisting: evt.data.isPreExisting,
        });
      }
    });

    this.resetReadyPromise();
  }

  private resetReadyPromise() {
    this.readyPromise = new Promise((resolve) => {
      this.readyResolve = resolve;
    });
  }

  public getOutputChannel(): vscode.OutputChannel {
    return this.outputChannel;
  }

  public getConnection(): AgentConnection {
    return this.connection;
  }

  public getHost(): IDEHost {
    return this.host;
  }

  public getWebviewBridge(): WebviewBridge {
    return this.webviewBridge;
  }

  public async start(): Promise<void> {
    const workspaceRoots = this.host.getWorkspaceFolders();
    const primaryWorkspace = workspaceRoots.length > 0 ? workspaceRoots[0] : process.cwd();

    this.outputChannel.appendLine('========================================');
    this.outputChannel.appendLine('  Starting BlackJak AI Agent Lifecycle');
    this.outputChannel.appendLine(`  Host: ${this.host.info.displayName} (${this.host.info.name})`);
    this.outputChannel.appendLine(`  API:  ${this.host.info.apiVersion}`);
    this.outputChannel.appendLine(`  Workspace: ${primaryWorkspace}`);
    this.outputChannel.appendLine('========================================');

    this.connection.setState('starting_backend');

    try {
      // 1. Spawn backend binary on dynamic free port
      const { port } = await this.processManager.startProcess(primaryWorkspace, this.context.extensionPath);
      this.outputChannel.appendLine(`[AgentManager] Backend allocated port ${port}`);

      // 2. Poll health check until ready == true
      this.connection.setState('connecting');
      await HealthCheckProbe.pollUntilReady(port, 15000);
      this.outputChannel.appendLine('[AgentManager] Backend health check PASSED (ready == true)');

      // 3. Connect WebSocket & Handshake
      await this.connection.connect(port);
      this.outputChannel.appendLine('[AgentManager] Agent Control Plane READY');

      // 4. Capability negotiation: advertise host identity & features so the
      //    agent can degrade gracefully on hosts with partial API support.
      this.connection.send({ type: 'host.hello', data: this.host.info });

      if (this.readyResolve) {
        this.readyResolve();
      }
    } catch (err: any) {
      this.outputChannel.appendLine(`[AgentManager ERROR] Startup failed: ${err.message}`);
      this.connection.setState('error');
      void this.host.showNotification(`Agent backend failed to start: ${err.message}`, 'error');
      throw err;
    }
  }

  public async stop(): Promise<void> {
    this.connection.setState('stopping');
    this.connection.disconnect();
    await this.processManager.stopProcess();
    this.connection.setState('stopped');
    this.resetReadyPromise();
  }

  public async restart(): Promise<void> {
    this.outputChannel.appendLine('[AgentManager] Restarting Agent...');
    await this.stop();
    await this.start();
  }

  public async waitUntilReady(): Promise<void> {
    if (this.connection.getState() === 'ready') {
      return;
    }
    if (this.readyPromise) {
      await this.readyPromise;
    }
  }

  private updateStatusBar(state: ConnectionState) {
    this.statusBarItem.show();
    switch (state) {
      case 'starting':
      case 'starting_backend':
        this.statusBarItem.text = '$(sync~spin) Agent: Starting';
        this.statusBarItem.tooltip = 'Agent backend is starting up...';
        break;
      case 'connecting':
      case 'initializing':
        this.statusBarItem.text = '$(sync~spin) Agent: Connecting';
        this.statusBarItem.tooltip = 'Connecting to Agent API server...';
        break;
      case 'ready':
        this.statusBarItem.text = '$(check) Agent: Ready';
        this.statusBarItem.tooltip = 'Agent is connected and ready.';
        break;
      case 'reconnecting':
        this.statusBarItem.text = '$(warning) Agent: Reconnecting';
        this.statusBarItem.tooltip = 'Reconnecting to agent backend...';
        break;
      case 'error':
        this.statusBarItem.text = '$(error) Agent: Error';
        this.statusBarItem.tooltip = 'Agent server error. Click to view logs.';
        break;
      case 'stopped':
      case 'stopping':
        this.statusBarItem.text = '$(circle-slash) Agent: Stopped';
        this.statusBarItem.tooltip = 'Agent server is stopped.';
        break;
    }
  }

  public dispose() {
    this.statusBarItem.dispose();
    this.stop().catch(() => {});
  }
}
