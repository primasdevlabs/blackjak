import * as vscode from 'vscode';

/**
 * Normalized host identifiers shared with the Go agent's protocol package.
 */
export const HOST_IDS = {
  VSCode: 'vscode',
  Cursor: 'cursor',
  Windsurf: 'windsurf',
  VSCodium: 'vscodium',
  Theia: 'theia',
  Unknown: 'unknown',
} as const;

/**
 * Maps the host application's self-reported name onto a normalized host id.
 *
 * Detection is informational only — feature code must branch on
 * HostCapabilities, never on this name.
 */
export function detectHostName(): string {
  const appName = (vscode.env.appName || '').toLowerCase();
  const appHost = (vscode.env.appHost || '').toLowerCase();

  if (appName.includes('cursor')) {
    return HOST_IDS.Cursor;
  }
  if (appName.includes('windsurf')) {
    return HOST_IDS.Windsurf;
  }
  if (appName.includes('vscodium')) {
    return HOST_IDS.VSCodium;
  }
  if (appName.includes('theia') || appHost.includes('theia')) {
    return HOST_IDS.Theia;
  }
  if (appName.includes('code') || appName.includes('vscode')) {
    return HOST_IDS.VSCode;
  }
  return HOST_IDS.Unknown;
}
