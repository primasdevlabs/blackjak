import * as vscode from 'vscode';
import { getWorkspaceFolder } from './workspace';
import { AgentServerManager } from './agentClient';
import { AgentWebviewProvider } from './webviewProvider';
import { registerAgentCommands } from './commands';

let serverManager: AgentServerManager | undefined;

export async function activate(context: vscode.ExtensionContext) {
  const wsPath = getWorkspaceFolder();
  const port = 8080;
  const host = '127.0.0.1';

  serverManager = new AgentServerManager(wsPath, port);

  const provider = new AgentWebviewProvider(context.extensionUri, host, port);

  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(AgentWebviewProvider.viewType, provider)
  );

  registerAgentCommands(context, provider, serverManager);

  // Auto-check/start server on extension activation
  await serverManager.ensureServerRunning(context.extensionPath);
}

export function deactivate() {
  if (serverManager) {
    serverManager.stopServer();
  }
}
