// Endpoint resolution for the agent backend.
// Priority: ?port URL param > localStorage override > extension-injected port > default.
// The localStorage key lets another local app pin a shared backend instance.

const DEFAULT_PORT = 47811;
const DEFAULT_HOST = '127.0.0.1';
const PORT_KEY = 'blackjak.agentPort';
const HOST_KEY = 'blackjak.agentHost';

export interface AgentEndpoint {
  host: string;
  port: number;
  httpUrl: string;
  wsUrl: string;
}

function readStored(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function resolveAgentEndpoint(): AgentEndpoint {
  const win = window as any;
  const params = new URLSearchParams(window.location.search || '');

  const rawPort =
    params.get('port') ??
    readStored(PORT_KEY) ??
    (win.__AGENT_PORT__ ? String(win.__AGENT_PORT__) : null);
  const rawHost =
    params.get('host') ??
    readStored(HOST_KEY) ??
    win.__AGENT_HOST__;

  const portNum = parseInt(rawPort || '', 10);
  const port = Number.isFinite(portNum) && portNum > 0 && portNum < 65536 ? portNum : DEFAULT_PORT;
  const host = rawHost || DEFAULT_HOST;

  return {
    host,
    port,
    httpUrl: `http://${host}:${port}`,
    wsUrl: `ws://${host}:${port}/ws`,
  };
}

/** Persist a port override so a manually-started backend survives reloads. */
export function setAgentPortOverride(port: number | null) {
  try {
    if (port === null) {
      window.localStorage.removeItem(PORT_KEY);
    } else {
      window.localStorage.setItem(PORT_KEY, String(port));
    }
  } catch {
    // storage unavailable (restricted webview) — ignore
  }
}
