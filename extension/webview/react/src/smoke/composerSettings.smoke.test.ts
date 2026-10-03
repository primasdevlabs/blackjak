import { describe, expect, it } from 'vitest';
import { normalizeMode } from '../api/settings';
import { modeFromSettings, normalizeModeUI } from '../utils/mode';
import { DEFAULT_COMMANDS } from '../components/SlashCommandMenu';

describe('composer + settings smoke', () => {
  it('normalizes legacy code mode to agent', () => {
    expect(normalizeMode('code')).toBe('agent');
    expect(normalizeMode('ask')).toBe('ask');
    expect(normalizeMode('plan')).toBe('plan');
    expect(normalizeModeUI('code')).toBe('agent');
  });

  it('maps settings mode for composer pill', () => {
    expect(modeFromSettings({ mode: 'code' } as any)).toBe('agent');
    expect(modeFromSettings({ mode: 'plan' } as any)).toBe('plan');
    expect(modeFromSettings({ mode: 'ask' } as any)).toBe('ask');
  });

  it('exposes core slash commands including skills entry points', () => {
    const names = DEFAULT_COMMANDS.map((c) => c.name);
    expect(names).toContain('/compact');
    expect(names).toContain('/ask');
    expect(names).toContain('/agent');
    expect(names).toContain('/pr');
    expect(names).toContain('/files');
  });

  it('settings modules export cursor-like panels', async () => {
    // Import parity panels directly (avoids SettingsPage → ConnectionStatus → host).
    const parity = await import('../settings/ParitySettings');
    expect(typeof parity.GeneralSettings).toBe('function');
    expect(typeof parity.RulesSettings).toBe('function');
    expect(typeof parity.SkillsSettings).toBe('function');
    expect(typeof parity.IndexingSettings).toBe('function');
    expect(typeof parity.BetaSettings).toBe('function');
    expect(typeof parity.UsageSettings).toBe('function');
    expect(typeof parity.GitPRSettings).toBe('function');
  });
});
