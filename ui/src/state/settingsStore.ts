import { SettingsConfig, settingsApi, Model } from '../api/settings';

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
    thinkingModelId: 'gpt-4o',
    codingModelId: 'claude-3-5-sonnet-20241022',
    fastModelId: 'gpt-4o-mini',
    useSeparateModels: true,
    effort: 'medium',
    mode: 'code',
    parallelSubagents: true,
    maxSubagents: 4,
    promptQueueBehavior: 'sequential',
    autoOpenFile: true,
    autoOpenDiff: true,
    askDestructiveOps: true,
    modelRoutes: {
      planning: 'thinking',
      exploration: 'fast',
      coding: 'coding',
      debugging: 'thinking',
      testing: 'coding',
      review: 'thinking',
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
}

export const settingsStore = new SettingsStore();
