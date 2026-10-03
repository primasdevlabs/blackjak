import { describe, expect, it } from 'vitest';

/** Mirrors App composer routing: failed/paused/cancelled resume; others start fresh. */
function shouldResumeRun(status: string | null | undefined): boolean {
  return status === 'paused' || status === 'cancelled' || status === 'failed';
}

describe('run continuity routing', () => {
  it('resumes failed, paused, and cancelled runs', () => {
    expect(shouldResumeRun('failed')).toBe(true);
    expect(shouldResumeRun('paused')).toBe(true);
    expect(shouldResumeRun('cancelled')).toBe(true);
  });

  it('starts a new task when idle/completed', () => {
    expect(shouldResumeRun('completed')).toBe(false);
    expect(shouldResumeRun(null)).toBe(false);
    expect(shouldResumeRun('running')).toBe(false);
  });
});
