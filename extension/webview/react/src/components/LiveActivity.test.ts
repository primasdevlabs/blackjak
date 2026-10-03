import { describe, expect, it } from 'vitest';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { LiveActivity } from './LiveActivity';

describe('LiveActivity', () => {
  it('returns null when idle with no activity', () => {
    const html = renderToStaticMarkup(
      createElement(LiveActivity, {
        status: 'completed',
        tools: [],
        filesRead: [],
        thought: '',
      }),
    );
    expect(html).toBe('');
  });

  it('uses shimmer status instead of a legacy spinner while running', () => {
    const html = renderToStaticMarkup(
      createElement(LiveActivity, {
        status: 'running',
        tools: [],
        filesRead: [],
        thought: '',
      }),
    );
    expect(html).toContain('live-activity');
    expect(html).toContain('activity-mark-live');
    expect(html).toContain('live-activity-shimmer');
    expect(html).not.toContain('icon-spin');
    expect(html).not.toContain('ArrowPath');
  });
});
