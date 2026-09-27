import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

export function getWorkspaceFolder(): string {
  if (vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders.length > 0) {
    return vscode.workspace.workspaceFolders[0].uri.fsPath;
  }
  return process.cwd();
}

export async function openFileInEditor(filePath: string): Promise<void> {
  try {
    const wsRoot = getWorkspaceFolder();
    const targetPath = path.isAbsolute(filePath) ? filePath : path.join(wsRoot, filePath);

    if (fs.existsSync(targetPath)) {
      const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(targetPath));
      await vscode.window.showTextDocument(doc, { preview: false });
    } else {
      vscode.window.showWarningMessage(`File not found: ${targetPath}`);
    }
  } catch (err: any) {
    vscode.window.showErrorMessage(`Failed to open file: ${err.message}`);
  }
}

export async function showNativeDiff(filePath: string, diffText?: string): Promise<void> {
  try {
    const wsRoot = getWorkspaceFolder();
    const targetPath = path.isAbsolute(filePath) ? filePath : path.join(wsRoot, filePath);

    const targetUri = vscode.Uri.file(targetPath);
    // Create temporary file with diff or original content to show side-by-side native VS Code diff
    const originalUri = targetUri.with({ scheme: 'file' });

    await vscode.commands.executeCommand(
      'vscode.diff',
      originalUri,
      targetUri,
      `Agent Diff: ${path.basename(filePath)}`
    );
  } catch (err: any) {
    vscode.window.showErrorMessage(`Failed to show diff: ${err.message}`);
  }
}
