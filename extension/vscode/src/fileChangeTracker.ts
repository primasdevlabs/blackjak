import * as vscode from 'vscode';

export interface AgentTab {
  uri: string;
  taskId: string;
  openedByAgent: boolean;
  touched: boolean;
  created: boolean;
  role?: string;
  importance?: number;
}

export interface FileChangeEventPayload {
  taskId: string;
  agentId: string;
  path: string;
  previousPath?: string;
  changeType: 'created' | 'modified' | 'deleted' | 'renamed' | 'moved';
  state?: string;
  role?: 'primary' | 'related' | 'test';
  importance?: number;
  isPreExisting?: boolean;
}

export class FileChangeTracker {
  private agentTabs: Map<string, AgentTab> = new Map();
  private dirtyWarningShown: Set<string> = new Set();

  constructor() {
    this.registerEditorListeners();
  }

  private registerEditorListeners() {
    vscode.workspace.onDidChangeTextDocument((e) => {
      if (e.document.isDirty) {
        const filePath = e.document.uri.fsPath;
        if (!this.dirtyWarningShown.has(filePath)) {
          this.dirtyWarningShown.add(filePath);
        }
      }
    });

    vscode.workspace.onDidCloseTextDocument((doc) => {
      this.agentTabs.delete(doc.uri.fsPath);
      this.dirtyWarningShown.delete(doc.uri.fsPath);
    });
  }

  public checkUnsavedChanges(filePath: string): boolean {
    const doc = vscode.workspace.textDocuments.find(d => d.uri.fsPath === filePath);
    if (doc && doc.isDirty) {
      vscode.window.showWarningMessage(
        `File has unsaved user changes: ${vscode.workspace.asRelativePath(filePath)}. Agent edit paused for review.`,
        'Save & Continue',
        'Review Diff',
        'Cancel'
      ).then(choice => {
        if (choice === 'Save & Continue') {
          doc.save();
        }
      });
      return true; // Is dirty
    }
    return false;
  }

  public trackAgentTab(tab: AgentTab) {
    this.agentTabs.set(tab.uri, tab);
  }

  public getAgentTabs(): AgentTab[] {
    return Array.from(this.agentTabs.values());
  }

  public clearTaskTabs(taskId: string) {
    for (const [key, tab] of this.agentTabs.entries()) {
      if (tab.taskId === taskId) {
        this.agentTabs.delete(key);
      }
    }
  }
}

export const fileChangeTracker = new FileChangeTracker();
