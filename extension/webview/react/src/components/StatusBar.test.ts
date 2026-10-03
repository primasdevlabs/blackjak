import { describe, expect, it, vi } from 'vitest';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { StatusBar } from './StatusBar';

describe('StatusBar', () => {
  it('hides when idle and connected', () => {
    const html = renderToStaticMarkup(
      createElement(StatusBar, {
        connectionStatus: 'connected',
        runStatus: 'completed',
        activeRunId: null,
        onCancel: () => undefined,
      }),
    );
    expect(html).toBe('');
  });

  it('shows Working + Stop above composer while running', () => {
    const onCancel = vi.fn();
    const html = renderToStaticMarkup(
      createElement(StatusBar, {
        connectionStatus: 'connected',
        runStatus: 'running',
        activeRunId: 'run_1',
        onCancel,
      }),
    );
    expect(html).toContain('composer-run-bar');
    expect(html).toContain('Working');
    expect(html).toContain('Stop');
    expect(html).toContain('activity-mark-live');
  });

  it('shows Waiting label when awaiting approval', () => {
    const html = renderToStaticMarkup(
      createElement(StatusBar, {
        connectionStatus: 'connected',
        runStatus: 'waiting',
        activeRunId: 'run_1',
        onCancel: () => undefined,
      }),
    );
    expect(html).toContain('Waiting');
  });

  it('surfaces connection issues without a run', () => {
    const html = renderToStaticMarkup(
      createElement(StatusBar, {
        connectionStatus: 'disconnected',
        runStatus: null,
        activeRunId: null,
        onCancel: () => undefined,
      }),
    );
    expect(html).toContain('Disconnected');
    expect(html).not.toContain('Working');
  });
});
