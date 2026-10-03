import { describe, expect, it } from 'vitest';
import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { ActivityStream } from './ActivityStream';
import { FilesReviewBar } from './FilesReviewBar';

describe('ActivityStream', () => {
  it('returns null when idle with no blocks', () => {
    const html = renderToStaticMarkup(
      createElement(ActivityStream, {
        tools: [],
        fileChanges: [],
        filesRead: [],
        isWorking: false,
      }),
    );
    expect(html).toBe('');
  });

  it('renders explored status without Files/Review (those sit above composer)', () => {
    const html = renderToStaticMarkup(
      createElement(ActivityStream, {
        tools: [],
        fileChanges: [],
        filesRead: ['a.go', 'b.go'],
        isWorking: true,
      }),
    );
    expect(html).toContain('activity-stream');
    expect(html).toContain('Explored 2 files');
    expect(html).not.toContain('Review');
    expect(html).not.toContain('Files');
  });

  it('renders terminal cards for shell tools', () => {
    const html = renderToStaticMarkup(
      createElement(ActivityStream, {
        tools: [
          {
            id: 't1',
            name: 'shell',
            step: 'go test ./...',
            status: 'running',
            output: 'ok',
          } as any,
        ],
        fileChanges: [],
        filesRead: [],
        isWorking: true,
      }),
    );
    expect(html).toContain('activity-terminal-card');
    expect(html).toContain('go test');
    expect(html).toContain('activity-mark-live');
  });
});

describe('FilesReviewBar', () => {
  it('renders file count and Review above the composer', () => {
    const html = renderToStaticMarkup(
      createElement(FilesReviewBar, {
        filesRead: ['a.go'],
        filePaths: ['b.go'],
        onReview: () => undefined,
      }),
    );
    expect(html).toContain('composer-files-bar');
    expect(html).toContain('2 Files');
    expect(html).toContain('Review');
  });
});
