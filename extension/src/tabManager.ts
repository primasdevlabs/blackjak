import { FileChangeTracker, FileChangeEventPayload } from './fileChangeTracker';
import { IDEHost } from './host/host';

export interface TabPolicySettings {
  autoOpenThreshold: number; // default: 5
  openLocation: 'beside' | 'active' | 'newGroup'; // default: 'beside'
  preserveFocus: boolean; // default: true
  openCreatedFiles: boolean; // default: true
  openModifiedFiles: boolean; // default: true
  openDeletedFiles: boolean; // default: false
}

export class TabManager {
  private tracker: FileChangeTracker;
  private host: IDEHost;
  private settings: TabPolicySettings;
  private touchedCount: Map<string, number> = new Map(); // taskId -> count

  constructor(tracker: FileChangeTracker, host: IDEHost) {
    this.tracker = tracker;
    this.host = host;
    this.settings = {
      autoOpenThreshold: 5,
      openLocation: 'beside',
      preserveFocus: true,
      openCreatedFiles: true,
      openModifiedFiles: true,
      openDeletedFiles: false,
    };
    this.registerFocusListener();
  }

  public updateSettings(newSettings: Partial<TabPolicySettings>) {
    this.settings = { ...this.settings, ...newSettings };
  }

  private registerFocusListener() {
    this.host.onDidChangeActiveEditor((_filePath) => {
      // Optionally notify webview or client of active editor focus
    });
  }

  public async handleFileChangeEvent(evt: FileChangeEventPayload) {
    const { taskId, path: filePath, changeType, role, importance } = evt;

    if (changeType === 'deleted' && !this.settings.openDeletedFiles) {
      return;
    }

    // Check unsaved changes safety
    const isDirty = this.tracker.checkUnsavedChanges(filePath);
    if (isDirty) {
      return;
    }

    const currentCount = (this.touchedCount.get(taskId) || 0) + 1;
    this.touchedCount.set(taskId, currentCount);

    this.tracker.trackAgentTab({
      uri: filePath,
      taskId,
      openedByAgent: true,
      touched: true,
      created: changeType === 'created',
      role,
      importance,
    });

    if (currentCount > this.settings.autoOpenThreshold) {
      if (importance && importance >= 4) {
        await this.openDocument(filePath);
      } else {
        void this.host.showNotification(
          `Agent changed ${currentCount} files in task. Primary files opened in tabs.`,
          'info',
          'Review All Changes'
        );
      }
      return;
    }

    await this.openDocument(filePath);
  }

  private async openDocument(filePath: string) {
    try {
      // Check if file is already open in any visible or background editor group
      if (this.host.isDocumentOpen(filePath)) {
        // Already open, no duplicate tab creation needed
        return;
      }

      await this.host.openFile(filePath, {
        viewColumn: this.settings.openLocation,
        preview: false, // Persistent tabs for agent created/modified files
        preserveFocus: this.settings.preserveFocus,
      });
    } catch (e) {
      // File may be deleted or inaccessible
    }
  }

  public async openTaskChangesGroup(filePaths: string[]) {
    for (let i = 0; i < Math.min(filePaths.length, 10); i++) {
      await this.openDocument(filePaths[i]);
    }
  }
}
