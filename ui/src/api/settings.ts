import { apiClient } from './client';

export interface ProviderCredentials {
  apiKey?: string;
  baseUrl?: string;
  orgId?: string;
  modelId?: string;
  storageMode: 'environment' | 'stored' | 'session';
}

export interface SettingsConfig {
  activeProvider: string;
  providers: Record<string, ProviderCredentials>;
  thinkingModelId: string;
  codingModelId: string;
  fastModelId: string;
  useSeparateModels: boolean;
  effort: 'minimal' | 'low' | 'medium' | 'high' | 'maximum';
  mode: 'plan' | 'code';
  parallelSubagents: boolean;
  maxSubagents: number;
  promptQueueBehavior: 'sequential' | 'parallel' | 'ask';
  autoOpenFile: boolean;
  autoOpenDiff: boolean;
  askDestructiveOps: boolean;
  modelRoutes: Record<string, string>;
}

export interface Model {
  id: string;
  name: string;
  provider: string;
  supportsTools: boolean;
  supportsVision: boolean;
  supportsReasoning: boolean;
  contextWindow: number;
  defaultMaxTokens: number;
}

export interface QueuedPrompt {
  id: string;
  runId?: string;
  prompt: string;
  mode: 'plan' | 'code';
  createdAt: string;
  status: 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
  dependencies?: string[];
}

export class AgentSettingsApi {
  async getSettings(): Promise<SettingsConfig> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/settings`);
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to fetch settings');
    return json.data;
  }

  async updateSettings(settings: Partial<SettingsConfig>): Promise<SettingsConfig> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/settings`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings),
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to update settings');
    return json.data;
  }

  async testProvider(providerId: string): Promise<{ message: string; models: Model[] }> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/providers/${encodeURIComponent(providerId)}/test`, {
      method: 'POST',
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to test provider connection');
    return json.data;
  }

  async listQueue(): Promise<QueuedPrompt[]> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/queue`);
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to fetch queue');
    return json.data;
  }

  async addToQueue(prompt: string, mode: 'plan' | 'code' = 'code', dependencies?: string[]): Promise<QueuedPrompt> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/queue`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt, mode, dependencies }),
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to add to queue');
    return json.data;
  }

  async deleteQueueItem(id: string): Promise<void> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/queue/${id}`, {
      method: 'DELETE',
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to delete queue item');
  }

  async runQueueItem(id: string): Promise<void> {
    const res = await fetch(`${apiClient.getBaseUrl()}/api/queue/${id}/run`, {
      method: 'POST',
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || 'Failed to run queue item');
  }
}

export const settingsApi = new AgentSettingsApi();
