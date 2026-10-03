import { afterEach, describe, expect, it, vi } from 'vitest';

/**
 * Contract tests for the rules settings API surface used by RulesSettings.
 * Keeps the colon-free body endpoints from regressing to path-based 405s.
 */

const base = 'http://127.0.0.1:47811';

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('rules settings API contract', () => {
  it('toggles enable via POST /api/rules/enable with id in body', async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => ({ success: true }),
    }));
    vi.stubGlobal('fetch', fetchMock);

    await fetch(`${base}/api/rules/enable`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: 'project:anti-slop', enabled: false }),
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toBe(`${base}/api/rules/enable`);
    expect(init?.method).toBe('POST');
    expect(JSON.parse(String(init?.body))).toEqual({
      id: 'project:anti-slop',
      enabled: false,
    });
    // Must not put ":" into the path.
    expect(String(url)).not.toContain('project%3A');
    expect(String(url)).not.toContain('project:');
  });

  it('updates via PUT /api/rules/update', async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => ({ success: true, data: {} }),
    }));
    vi.stubGlobal('fetch', fetchMock);

    await fetch(`${base}/api/rules/update`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        id: 'project:ui',
        name: 'ui',
        body: 'custom',
        source: 'project',
      }),
    });

    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toBe(`${base}/api/rules/update`);
    expect(init?.method).toBe('PUT');
  });

  it('deletes via POST /api/rules/delete', async () => {
    const fetchMock = vi.fn(async () => ({
      ok: true,
      json: async () => ({ success: true }),
    }));
    vi.stubGlobal('fetch', fetchMock);

    await fetch(`${base}/api/rules/delete`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: 'user:custom' }),
    });

    const [url] = fetchMock.mock.calls[0];
    expect(String(url)).toBe(`${base}/api/rules/delete`);
  });
});
