import { AgentHost } from './agentHost';
import { WebviewAgentHost } from './webviewHost';
import { BrowserAgentHost } from './browserHost';

function createAgentHost(): AgentHost {
  // The extension injects window.vscode (acquireVsCodeApi) before the bundle
  // loads. Any VS Code-family host (Code, Cursor, Windsurf, VSCodium, Theia)
  // supplies the same global, so this single check is sufficient.
  const w = window as any;
  if (typeof w.vscode !== 'undefined' && typeof w.vscode.postMessage === 'function') {
    return new WebviewAgentHost(w.vscode);
  }
  return new BrowserAgentHost();
}

/** Shared host accessor — the only entry point UI code should use. */
export const agentHost: AgentHost = createAgentHost();

export * from './agentHost';
