import { SettingsConfig } from '../api/settings';

export type AgentModeUI = 'ask' | 'plan' | 'agent';

export function normalizeModeUI(mode: string): AgentModeUI {
  if (mode === 'ask' || mode === 'plan' || mode === 'agent') return mode;
  if (mode === 'code') return 'agent';
  return 'agent';
}

export function modeFromSettings(settings: Pick<SettingsConfig, 'mode'>): AgentModeUI {
  return normalizeModeUI(settings.mode);
}
