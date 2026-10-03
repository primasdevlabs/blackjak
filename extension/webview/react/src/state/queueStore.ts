import { QueuedPrompt, settingsApi } from '../api/settings';

type Listener = () => void;

class QueueStore {
  private queue: QueuedPrompt[] = [];
  private listeners: Set<Listener> = new Set();

  getQueue(): QueuedPrompt[] {
    return this.queue;
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  private notify() {
    this.listeners.forEach((l) => l());
  }

  async fetchQueue() {
    try {
      const data = await settingsApi.listQueue();
      this.queue = data;
      this.notify();
    } catch (err) {
      console.warn('Could not fetch queue:', err);
    }
  }

  async addPrompt(prompt: string, mode: QueuedPrompt['mode'] = 'agent', dependencies?: string[]) {
    try {
      const item = await settingsApi.addToQueue(prompt, mode, dependencies);
      this.queue = [...this.queue, item];
      this.notify();
    } catch (err) {
      console.error('Failed to add to queue:', err);
    }
  }

  async removeItem(id: string) {
    this.queue = this.queue.filter((q) => q.id !== id);
    this.notify();
    try {
      await settingsApi.deleteQueueItem(id);
    } catch (err) {
      console.error('Failed to delete queue item:', err);
    }
  }

  async runItem(id: string) {
    this.queue = this.queue.map((q) => (q.id === id ? { ...q, status: 'running' } : q));
    this.notify();
    try {
      await settingsApi.runQueueItem(id);
    } catch (err) {
      console.error('Failed to run queue item:', err);
    }
  }
}

export const queueStore = new QueueStore();
