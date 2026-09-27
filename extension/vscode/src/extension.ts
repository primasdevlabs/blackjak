import * as vscode from 'vscode';
import { AgentManager } from './agentManager';
import { AgentWebviewProvider } from './webviewProvider';
import { registerAgentCommands } from './commands';

let agentManager: AgentManager | undefined;

export async function activate(context: vscode.ExtensionContext) {
  agentManager = new AgentManager(context);

  const provider = new AgentWebviewProvider(context, agentManager);

  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider('agent.sidebar', provider),
    {
      dispose: () => {
        agentManager?.dispose();
      },
    }
  );

  registerAgentCommands(context, provider, agentManager);

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
