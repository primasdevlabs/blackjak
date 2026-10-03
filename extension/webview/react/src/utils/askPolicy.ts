import type { AskPolicy } from '../api/settings';

/** Normalize ask-policy values from settings / patches. */
export function normalizeAskPolicy(value: unknown): AskPolicy {
  return value === 'accept_all' ? 'accept_all' : 'standard';
}

export function askPolicyLabel(policy: AskPolicy): string {
  return policy === 'accept_all' ? 'Accept all' : 'Reject certain asks';
}
