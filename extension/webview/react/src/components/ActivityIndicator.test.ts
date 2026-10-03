import { describe, expect, it } from 'vitest';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { ActivityIndicator } from './ActivityIndicator';

describe('ActivityIndicator', () => {
  it('renders live pulse mark by default', () => {
    const html = renderToStaticMarkup(createElement(ActivityIndicator));
    expect(html).toContain('activity-mark-live');
    expect(html).not.toContain('activity-mark-done');
  });

  it('renders done / error / idle states', () => {
    expect(renderToStaticMarkup(createElement(ActivityIndicator, { state: 'done' }))).toContain('activity-mark-done');
    expect(renderToStaticMarkup(createElement(ActivityIndicator, { state: 'error' }))).toContain('activity-mark-error');
    expect(renderToStaticMarkup(createElement(ActivityIndicator, { state: 'idle' }))).toContain('activity-mark-idle');
  });
});
