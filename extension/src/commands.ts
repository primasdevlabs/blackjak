import * as vscode from 'vscode';
import { AgentWebviewProvider } from './webviewProvider';
import { AgentManager } from './agentManager';
import { IDEHost } from './host/host';

export function registerAgentCommands(
  context: vscode.ExtensionContext,
  webviewProvider: AgentWebviewProvider,
  agentManager: AgentManager,
  host: IDEHost
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
        void host.showNotification('Active agent task cancelled.', 'info');
      }
    } catch (err: any) {
      void host.showNotification(`Failed to cancel task: ${err.message}`, 'error');
    }
  });

  const restartCmd = vscode.commands.registerCommand('agent.restart', async () => {
    void host.showNotification('Restarting Agent backend...', 'info');
    await agentManager.restart();
    void host.showNotification('Agent backend restarted.', 'info');
  });

  const openSettingsCmd = vscode.commands.registerCommand('agent.openSettings', async () => {
    await agentManager.waitUntilReady();
    webviewProvider.showPanel('settings');
  });

  const openActivityCmd = vscode.commands.registerCommand('agent.openActivity', async () => {
    await agentManager.waitUntilReady();
    webviewProvider.showPanel('activity');
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
    openActivityCmd,
    showLogsCmd
  );
}
