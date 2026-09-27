import * as vscode from 'vscode';
import { HostCapabilities } from './host';

/**
 * Detects host capabilities through feature detection rather than host name.
 * Forks and Theia may stub or omit parts of the VS Code API surface, so each
 * capability is probed independently and the agent degrades gracefully when
 * one is reported false.
 */
export async function detectCapabilities(context: vscode.ExtensionContext): Promise<HostCapabilities> {
  const fn = (v: unknown) => typeof v === 'function';

  let diffEditor = false;
  try {
    const commands = await vscode.commands.getCommands(true);
    diffEditor = commands.includes('vscode.diff');
  } catch {
    // Command enumeration unsupported — assume the base diff command exists.
    diffEditor = fn(vscode.commands.executeCommand);
  }

  return {
    // The extension only runs once a webview view resolves, but report the
    // actual surface anyway for completeness.
    webview: fn(vscode.window.registerWebviewViewProvider),
    editorTabs: fn(vscode.window.showTextDocument) && !!(vscode.window as { tabGroups?: unknown }).tabGroups,
    fileWatcher: fn(vscode.workspace.createFileSystemWatcher),
    terminal: fn(vscode.window.createTerminal),
    scm:
      !!vscode.extensions.getExtension('vscode.git') ||
      fn((vscode as unknown as { scm?: { createSourceControl?: unknown } }).scm?.createSourceControl),
    diffEditor,
    secrets: !!context.secrets,
  };
}
