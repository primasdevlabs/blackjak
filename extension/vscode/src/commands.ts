import * as vscode from 'vscode';
import { AgentWebviewProvider } from './webviewProvider';
import { AgentManager } from './agentManager';

export function registerAgentCommands(
  context: vscode.ExtensionContext,
  webviewProvider: AgentWebviewProvider,
  agentManager: AgentManager
) {
  const openCmd = vscode.commands.registerCommand('agent.open', async () => {
    await agentManager.waitUntilReady();
    webviewProvider.showPanel();
  });

  const openSidebarCmd = vscode.commands.registerCommand('agent.openSidebar', async () => {
    await vscode.commands.executeCommand('agent.sidebar.focus');
  });

  const newTaskCmd = vscode.commands.registerCommand('agent.newTask', async () => {
    await agentManager.waitUntilReady();
    webviewProvider.showPanel();
  });

  const planTaskCmd = vscode.commands.registerCommand('agent.planTask', async () => {
    await agentManager.waitUntilReady();
    webviewProvider.showPanel();
  });

  const cancelCmd = vscode.commands.registerCommand('agent.cancelTask', async () => {
    const port = agentManager.getConnection().getPort();
    try {
      const res = await fetch(`http://127.0.0.1:${port}/api/runs`);
      const json: any = await res.json();
      if (json.success && Array.isArray(json.data)) {
        for (const run of json.data) {
          if (run.status === 'running' || run.status === 'waiting' || run.status === 'pending') {
            await fetch(`http://127.0.0.1:${port}/api/runs/${run.id}/cancel`, { method: 'POST' });
          }
        }
        vscode.window.showInformationMessage('Active agent task cancelled.');
      }
    } catch (err: any) {
      vscode.window.showErrorMessage(`Failed to cancel task: ${err.message}`);
    }
  });

  const restartCmd = vscode.commands.registerCommand('agent.restart', async () => {
    vscode.window.showInformationMessage('Restarting Agent backend...');
    await agentManager.restart();
    vscode.window.showInformationMessage('Agent backend restarted.');
  });

  const openSettingsCmd = vscode.commands.registerCommand('agent.openSettings', async () => {
    await agentManager.waitUntilReady();
    webviewProvider.showPanel();
  });

  const showLogsCmd = vscode.commands.registerCommand('agent.showLogs', () => {
    agentManager.getOutputChannel().show(true);
  });

  context.subscriptions.push(
    openCmd,
    openSidebarCmd,
    newTaskCmd,
    planTaskCmd,
    cancelCmd,
    restartCmd,
    openSettingsCmd,
    showLogsCmd
  );
}
