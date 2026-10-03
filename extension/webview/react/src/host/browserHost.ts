import {
  AgentHost,
  HostInfo,
  NotificationLevel,
  OpenEditorsState,
  OpenFileOptions,
  PickFilesOptions,
  WorkspaceState,
} from './agentHost';

/**
 * BrowserAgentHost is the development fallback used when the UI runs outside
 * an IDE webview (e.g. `npm run dev` on localhost). Host operations become
 * no-ops so the UI remains usable against a standalone agent server.
 */
export class BrowserAgentHost implements AgentHost {
  readonly isEmbedded = false;
  private secrets: Map<string, string> = new Map();

  postMessage(message: unknown): void {
    console.debug('[BrowserHost] postMessage (no IDE host attached):', message);
  }

  openFile(path: string, options?: OpenFileOptions): void {
    console.debug('[BrowserHost] openFile:', path, options);
  }

  openDiff(path: string, _rightPath?: string): void {
    void (async () => {
      try {
        const { apiClient } = await import('../api/client');
        const rel = path.replace(/\\/g, '/');
        const res = await fetch(`${apiClient.getBaseUrl()}/api/diffs/${encodeURIComponent(rel)}`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${rel.split('/').pop() || 'file'}.diff.txt`;
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(url);
        this.showNotification(`Downloaded diff for ${rel}`, 'info');
      } catch (err: any) {
        this.showNotification(`Diff download failed: ${err?.message || err}`, 'error');
      }
    })();
  }

  revealFile(path: string): void {
    console.debug('[BrowserHost] revealFile:', path);
  }

  createFile(path: string, content = ''): void {
    console.debug('[BrowserHost] createFile:', path, `${content.length} bytes`);
  }

  createFolder(path: string): void {
    console.debug('[BrowserHost] createFolder:', path);
  }

  revealFolder(path: string): void {
    console.debug('[BrowserHost] revealFolder:', path);
  }

  createTerminal(name?: string): void {
    console.debug('[BrowserHost] createTerminal:', name);
  }

  runInTerminal(command: string, terminalName?: string): void {
    console.debug('[BrowserHost] runInTerminal:', command, terminalName);
  }

  showNotification(message: string, level: NotificationLevel = 'info'): void {
    const w = window as any;
    const invoke = w.__TAURI__?.core?.invoke || w.__TAURI_INTERNALS__?.invoke;
    if (typeof invoke === 'function') {
      void invoke('notify', {
        title: level === 'error' ? 'BlackJak Error' : 'BlackJak',
        body: message,
      }).catch(() => undefined);
    }
    console.log(`[BrowserHost] ${level.toUpperCase()}: ${message}`);
  }

  async getHostInfo(): Promise<HostInfo | undefined> {
    return {
      name: 'browser',
      displayName: 'Browser (standalone)',
      version: '',
      apiVersion: '',
      extensionVersion: '',
      capabilities: {
        webview: false,
        editorTabs: false,
        fileWatcher: false,
        terminal: false,
        scm: false,
        diffEditor: false,
        secrets: true,
        filePicker: true,
        multiPanel: false,
        settingsWindow: true,
        activityPanel: true,
      },
    };
  }

  async getWorkspaceState(): Promise<WorkspaceState> {
    const w = window as any;
    const invoke = w.__TAURI__?.core?.invoke || w.__TAURI_INTERNALS__?.invoke;
    if (typeof invoke === 'function') {
      try {
        const ws = await invoke('agent_workspace');
        if (typeof ws === 'string' && ws) {
          return { folders: [ws] };
        }
      } catch {
        /* fall through */
      }
    }
    const stored = localStorage.getItem('blackjak.workspaceFolders');
    return { folders: stored ? JSON.parse(stored) : [] };
  }

  async getOpenEditors(): Promise<OpenEditorsState> {
    return { editors: [] };
  }

  async saveSecret(key: string, value: string): Promise<void> {
    this.secrets.set(key, value);
    try {
      localStorage.setItem(`blackjak.secret.${key}`, value);
    } catch {
      /* quota */
    }
  }

  async getSecret(key: string): Promise<string | undefined> {
    if (this.secrets.has(key)) return this.secrets.get(key);
    return localStorage.getItem(`blackjak.secret.${key}`) ?? undefined;
  }

  async pickFiles(_options?: PickFilesOptions): Promise<string[]> {
    const w = window as any;
    const invoke = w.__TAURI__?.core?.invoke || w.__TAURI_INTERNALS__?.invoke;
    if (typeof invoke === 'function') {
      try {
        const paths = await invoke('pick_files', {
          multiple: _options?.canSelectMany ?? true,
          title: _options?.title || 'Attach files',
        });
        if (Array.isArray(paths)) return paths as string[];
      } catch (err) {
        console.debug('[BrowserHost] Tauri pick_files failed, falling back:', err);
      }
    }

    const { apiClient } = await import('../api/client');
    const { rememberPreview } = await import('../utils/attachmentPreview');
    return new Promise((resolve) => {
      const input = document.createElement('input');
      input.type = 'file';
      input.multiple = _options?.canSelectMany ?? true;
      if (_options?.filters) {
        input.accept = Object.values(_options.filters)
          .flat()
          .map((e) => (e.startsWith('.') ? e : `.${e}`))
          .join(',');
      }
      input.onchange = async () => {
        const files = Array.from(input.files ?? []);
        const paths: string[] = [];
        for (const file of files) {
          const localPreview = file.type.startsWith('image/') ? URL.createObjectURL(file) : undefined;
          try {
            const body = new FormData();
            body.append('file', file);
            const res = await fetch(`${apiClient.getBaseUrl()}/api/uploads`, {
              method: 'POST',
              body,
            });
            const json = await res.json();
            if (json.success && json.data?.path) {
              paths.push(json.data.path);
              if (localPreview) rememberPreview(json.data.path, localPreview);
            } else {
              paths.push(file.name);
              if (localPreview) rememberPreview(file.name, localPreview);
            }
          } catch {
            paths.push(file.name);
            if (localPreview) rememberPreview(file.name, localPreview);
          }
        }
        resolve(paths);
      };
      input.click();
    });
  }

  openSettingsWindow(): void {
    const w = window as any;
    const invoke = w.__TAURI__?.core?.invoke || w.__TAURI_INTERNALS__?.invoke;
    if (typeof invoke === 'function') {
      void invoke('open_settings_window').catch(() => {
        window.dispatchEvent(new CustomEvent('blackjak:open-settings'));
      });
      return;
    }
    window.dispatchEvent(new CustomEvent('blackjak:open-settings'));
  }

  openActivityPanel(): void {
    const w = window as any;
    const invoke = w.__TAURI__?.core?.invoke || w.__TAURI_INTERNALS__?.invoke;
    if (typeof invoke === 'function') {
      void invoke('open_activity_window').catch(() => {
        window.dispatchEvent(new CustomEvent('blackjak:open-activity'));
      });
      return;
    }
    window.dispatchEvent(new CustomEvent('blackjak:open-activity'));
  }
}
