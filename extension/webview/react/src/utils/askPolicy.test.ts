import { describe, expect, it } from 'vitest';
import { askPolicyLabel, normalizeAskPolicy } from './askPolicy';

describe('askPolicy', () => {
  it('normalizes unknown values to standard', () => {
    expect(normalizeAskPolicy(undefined)).toBe('standard');
    expect(normalizeAskPolicy('')).toBe('standard');
    expect(normalizeAskPolicy('nope')).toBe('standard');
    expect(normalizeAskPolicy('standard')).toBe('standard');
    expect(normalizeAskPolicy('accept_all')).toBe('accept_all');
  });

  it('labels policies for settings UI', () => {
    expect(askPolicyLabel('standard')).toBe('Reject certain asks');
    expect(askPolicyLabel('accept_all')).toBe('Accept all');
  });
});
