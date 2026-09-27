import { HealthResponse, RunDTO, ApprovalResponse } from '../types/events';

export class AgentApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://127.0.0.1:47811') {
    this.baseUrl = baseUrl.replace(/\/+$/, '');
  }

  setBaseUrl(url: string) {
    this.baseUrl = url.replace(/\/+$/, '');
  }

  getBaseUrl(): string {
    return this.baseUrl;
  }

  async getHealth(): Promise<HealthResponse> {
    const res = await fetch(`${this.baseUrl}/health`);
    if (!res.ok) {
      throw new Error(`Health check failed with status ${res.status}`);
    }
    return res.json();
  }

  async createRun(prompt: string, workspace?: string): Promise<RunDTO> {
    const res = await fetch(`${this.baseUrl}/api/runs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt, workspace }),
    });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to create run');
    }
    return json.data;
  }

  async listRuns(): Promise<RunDTO[]> {
    const res = await fetch(`${this.baseUrl}/api/runs`);
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to list runs');
    }
    return json.data;
  }

  async getRun(id: string): Promise<RunDTO> {
    const res = await fetch(`${this.baseUrl}/api/runs/${id}`);
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to get run');
    }
    return json.data;
  }

  async deleteRun(id: string): Promise<void> {
    const res = await fetch(`${this.baseUrl}/api/runs/${id}`, {
      method: 'DELETE',
    });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to delete run');
    }
  }

  async cancelRun(id: string): Promise<RunDTO> {
    const res = await fetch(`${this.baseUrl}/api/runs/${id}/cancel`, {
      method: 'POST',
    });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to cancel run');
    }
    return json.data;
  }

  async pauseRun(id: string): Promise<void> {
    const res = await fetch(`${this.baseUrl}/api/runs/${id}/pause`, { method: 'POST' });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to pause run');
    }
  }

  async resumeRun(id: string, prompt?: string): Promise<RunDTO> {
    const res = await fetch(`${this.baseUrl}/api/runs/${id}/resume`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt: prompt || '' }),
    });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to resume run');
    }
    return json.data;
  }

  async reviewFileChange(runId: string, changeId: string, action: 'accept' | 'reject'): Promise<RunDTO> {
    const res = await fetch(`${this.baseUrl}/api/runs/${runId}/changes/${changeId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action }),
    });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to review change');
    }
    return json.data;
  }

  async submitApproval(runId: string, response: ApprovalResponse): Promise<RunDTO> {
    const res = await fetch(`${this.baseUrl}/api/runs/${runId}/approval`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(response),
    });
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error || 'Failed to submit approval');
    }
    return json.data;
  }
}

export const apiClient = new AgentApiClient();
