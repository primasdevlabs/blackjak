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
  };
}
