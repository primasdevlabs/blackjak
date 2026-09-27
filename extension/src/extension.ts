import * as vscode from 'vscode';
import { AgentManager } from './agentManager';
import { AgentWebviewProvider } from './webviewProvider';
import { registerAgentCommands } from './commands';
import { VSCodeHost } from './host/vscode';

let agentManager: AgentManager | undefined;

export async function activate(context: vscode.ExtensionContext) {
  // Host adapter boundary: the extension speaks to the IDE only through
  // IDEHost, and negotiates capabilities instead of assuming API coverage.
  const host = await VSCodeHost.create(context);

  agentManager = new AgentManager(context, host);

  const provider = new AgentWebviewProvider(context, agentManager);

  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider('agent.sidebar', provider),
    {
      dispose: () => {
        agentManager?.dispose();
      },
    }
  );

  registerAgentCommands(context, provider, agentManager, host);

  // Zero-configuration startup: Go agent starts automatically on extension activation
  agentManager.start().catch((err) => {
    vscode.window.showErrorMessage(`Agent startup failed: ${err.message}`);
  });
}

export function deactivate() {
  if (agentManager) {
    agentManager.stop().catch(() => {});
  }
}
