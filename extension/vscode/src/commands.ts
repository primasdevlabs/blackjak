import * as vscode from 'vscode';
import { AgentWebviewProvider } from './webviewProvider';
import { AgentServerManager } from './agentClient';

export function registerAgentCommands(
  context: vscode.ExtensionContext,
  webviewProvider: AgentWebviewProvider,
  serverManager: AgentServerManager
) {
  const openCmd = vscode.commands.registerCommand('agent.open', async () => {
    await serverManager.ensureServerRunning(context.extensionPath);
    webviewProvider.showPanel();
  });

  const newTaskCmd = vscode.commands.registerCommand('agent.newTask', async () => {
    await serverManager.ensureServerRunning(context.extensionPath);
    webviewProvider.showPanel();
  });

  const cancelCmd = vscode.commands.registerCommand('agent.cancelTask', async () => {
    const port = serverManager.getPort();
    const host = serverManager.getHost();
    try {
      // Fetch active runs and cancel
      const res = await fetch(`http://${host}:${port}/api/runs`);
      const json: any = await res.json();
      if (json.success && Array.isArray(json.data)) {
        for (const run of json.data) {
          if (run.status === 'running' || run.status === 'waiting' || run.status === 'pending') {
            await fetch(`http://${host}:${port}/api/runs/${run.id}/cancel`, { method: 'POST' });
          }
        }
        vscode.window.showInformationMessage('Active agent task cancelled.');
      }
    } catch (err: any) {
      vscode.window.showErrorMessage(`Failed to cancel task: ${err.message}`);
    }
  });

  context.subscriptions.push(openCmd, newTaskCmd, cancelCmd);
}
