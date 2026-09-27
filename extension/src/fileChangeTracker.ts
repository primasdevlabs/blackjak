import { IDEHost } from './host/host';

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
  private host: IDEHost;
  private agentTabs: Map<string, AgentTab> = new Map();
  private dirtyWarningShown: Set<string> = new Set();

  constructor(host: IDEHost) {
    this.host = host;
    this.registerEditorListeners();
  }

  private registerEditorListeners() {
    this.host.onDidChangeDocument((filePath, isDirty) => {
      if (isDirty && !this.dirtyWarningShown.has(filePath)) {
        this.dirtyWarningShown.add(filePath);
      }
    });

    this.host.onDidCloseDocument((filePath) => {
      this.agentTabs.delete(filePath);
      this.dirtyWarningShown.delete(filePath);
    });
  }

  public checkUnsavedChanges(filePath: string): boolean {
    if (this.host.isDocumentDirty(filePath)) {
      this.host
        .showNotification(
          `File has unsaved user changes: ${filePath}. Agent edit paused for review.`,
          'warn',
          'Save & Continue',
          'Review Diff',
          'Cancel'
        )
        .then((choice) => {
          if (choice === 'Save & Continue') {
            void this.host.saveDocument(filePath);
          } else if (choice === 'Review Diff') {
            void this.host.openDiff(filePath);
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
