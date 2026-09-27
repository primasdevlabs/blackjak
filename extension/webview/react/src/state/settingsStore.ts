import { SettingsConfig, settingsApi, Model, EffortLevel } from '../api/settings';

type Listener = () => void;

class SettingsStore {
  private settings: SettingsConfig = {
    activeProvider: 'OpenAI',
    providers: {
      'OpenAI': { baseUrl: 'https://api.openai.com/v1', storageMode: 'environment' },
      'Google Gemini': { storageMode: 'environment' },
      'Anthropic': { baseUrl: 'https://api.anthropic.com/v1', storageMode: 'environment' },
      'OpenAI-compatible': { baseUrl: 'http://localhost:11434/v1', modelId: 'llama3.2', storageMode: 'stored' },
    },
    thinking: { providerId: 'Anthropic', modelId: 'claude-opus-5', role: 'thinking' },
    coding: { providerId: 'OpenAI', modelId: 'gpt-5.3-codex', role: 'coding' },
    fast: { providerId: 'Google Gemini', modelId: 'gemini-3.5-flash-lite', role: 'fast' },
    review: { providerId: 'Anthropic', modelId: 'claude-sonnet-5', role: 'review' },
    thinkingModelId: 'claude-opus-5',
    codingModelId: 'gpt-5.3-codex',
    fastModelId: 'gemini-3.5-flash-lite',
    reviewModelId: 'claude-sonnet-5',
    useSeparateModels: true,
    effort: 'medium',
    mode: 'code',
    parallelSubagents: true,
    maxSubagents: 4,
    promptQueueBehavior: 'sequential',
    autoOpenFile: true,
    autoOpenDiff: true,
    askDestructiveOps: true,
    guardrails: {
      mode: 'supervised',
      shellAllowed: true,
      approveAllShell: false,
      denyCommands: [],
      protectedPaths: [],
      subagentsAllowed: true,
      maxSteps: 60,
    },
    modelRoutes: {
      planning: 'thinking',
      exploration: 'fast',
      coding: 'coding',
      debugging: 'thinking',
      testing: 'coding',
      review: 'review',
      summarization: 'fast',
    },
  };

  private providerModels: Record<string, Model[]> = {};
  private providerStatus: Record<string, 'connected' | 'disconnected' | 'testing'> = {};
  private listeners: Set<Listener> = new Set();

  getSettings(): SettingsConfig {
    return this.settings;
  }

  getProviderModels(providerId: string): Model[] {
    return this.providerModels[providerId] || [];
  }

  getAllProviderModels(): Record<string, Model[]> {
    return this.providerModels;
  }

  getProviderStatus(providerId: string): 'connected' | 'disconnected' | 'testing' {
    return this.providerStatus[providerId] || 'disconnected';
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  private notify() {
    this.listeners.forEach((l) => l());
  }

  async fetchSettings() {
    try {
      const data = await settingsApi.getSettings();
      this.settings = data;
      this.notify();
    } catch (err) {
      console.warn('Could not fetch backend settings:', err);
    }
  }

  async updateSettings(updates: Partial<SettingsConfig>) {
    this.settings = { ...this.settings, ...updates };
    this.notify();
    try {
      const updated = await settingsApi.updateSettings(updates);
      this.settings = updated;
      this.notify();
    } catch (err) {
      console.error('Failed to update settings:', err);
    }
  }

  async testProvider(providerId: string) {
    this.providerStatus[providerId] = 'testing';
    this.notify();
    try {
      const res = await settingsApi.testProvider(providerId);
      this.providerStatus[providerId] = 'connected';
      this.providerModels[providerId] = res.models || [];
      this.notify();
      return res;
    } catch (err: any) {
      this.providerStatus[providerId] = 'disconnected';
      this.notify();
      throw err;
    }
  }

  async refreshModels(providerId?: string) {
    const targets = providerId
      ? [providerId]
      : ['OpenAI', 'Google Gemini', 'Anthropic', 'OpenAI-compatible'];

    const results = await Promise.allSettled(
      targets.map((p) => settingsApi.refreshModels(p))
    );
    results.forEach((res, i) => {
      if (res.status === 'fulfilled' && Array.isArray(res.value)) {
        this.providerModels[targets[i]] = res.value;
      }
    });
    this.notify();
    return providerId ? this.providerModels[providerId] : this.providerModels;
  }
}

export const settingsStore = new SettingsStore();
