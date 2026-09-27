import { useState, useEffect } from 'react';
import { agentStore, AgentState } from '../state/agentStore';

export function useAgent(): AgentState & {
  startTask: (prompt: string, workspace?: string) => void;
  cancelTask: () => void;
  respondApproval: (granted: boolean, reason?: string) => void;
  configureBackend: (backendUrl: string, wsUrl: string) => void;
  addAttachment: (path: string, type?: 'file' | 'folder') => void;
  removeAttachment: (id: string) => void;
  selectSubagent: (id: string | null) => void;
  addSystemMessage: (text: string) => void;
  createSession: () => void;
  switchSession: (id: string) => void;
  closeSession: (id: string) => void;
  refreshRuns: () => void;
  loadRunFromHistory: (runId: string) => void;
  deleteRunFromHistory: (runId: string) => void;
  retryLastRun: () => void;
  pauseTask: () => void;
  resumeTask: (prompt?: string) => void;
} {
  const [state, setState] = useState<AgentState>(agentStore.getState());

  useEffect(() => {
    const unsubscribe = agentStore.subscribe(() => {
      setState(agentStore.getState());
    });
    return unsubscribe;
  }, []);

  return {
    ...state,
    startTask: (prompt, workspace) => agentStore.startTask(prompt, workspace),
    cancelTask: () => agentStore.cancelTask(),
    respondApproval: (granted, reason) => agentStore.respondApproval(granted, reason),
    configureBackend: (bUrl, wUrl) => agentStore.configureBackend(bUrl, wUrl),
    addAttachment: (path, type) => agentStore.addAttachment(path, type),
    removeAttachment: (id) => agentStore.removeAttachment(id),
    selectSubagent: (id) => agentStore.selectSubagent(id),
    addSystemMessage: (text) => agentStore.addSystemMessage(text),
    createSession: () => agentStore.createSession(),
    switchSession: (id) => agentStore.switchSession(id),
    closeSession: (id) => agentStore.closeSession(id),
    refreshRuns: () => { void agentStore.refreshRuns(); },
    loadRunFromHistory: (runId) => { void agentStore.loadRunFromHistory(runId); },
    deleteRunFromHistory: (runId) => { void agentStore.deleteRunFromHistory(runId); },
    retryLastRun: () => agentStore.retryLastRun(),
    pauseTask: () => agentStore.pauseTask(),
    resumeTask: (prompt) => agentStore.resumeTask(prompt),
  };
}
