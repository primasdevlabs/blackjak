import React, { useMemo, useState } from 'react';
import { FileChange } from '../types/events';
import { ToolExecution } from '../state/agentStore';
import { agentHost } from '../host';
import { DiffView } from './DiffView';
import {
  CommandLineIcon,
  DocumentIcon,
  ChevronRightIcon,
  ChevronDownIcon,
} from '@heroicons/react/24/outline';
import { ActivityIndicator } from './ActivityIndicator';


export type ActivityBlock =
  | { kind: 'status'; id: string; text: string }
  | { kind: 'diff'; id: string; change: FileChange; added: number; removed: number }
  | { kind: 'terminal'; id: string; title: string; output: string; failed?: boolean; running?: boolean }
  | { kind: 'tool'; id: string; tool: ToolExecution };

interface ActivityStreamProps {
  tools: ToolExecution[];
  fileChanges: FileChange[];
  filesRead: string[];
  isWorking: boolean;
  followUpSlot?: React.ReactNode;
}

function diffStats(diff?: string): { added: number; removed: number } {
  if (!diff) return { added: 0, removed: 0 };
  let added = 0;
  let removed = 0;
  for (const line of diff.split('\n')) {
    if (line.startsWith('+') && !line.startsWith('+++')) added++;
    if (line.startsWith('-') && !line.startsWith('---')) removed++;
  }
  return { added, removed };
}

function fileBadge(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() || '';
  if (ext === 'ts' || ext === 'tsx') return 'TS';
  if (ext === 'js' || ext === 'jsx') return 'JS';
  if (ext === 'go') return 'GO';
  if (ext === 'rs') return 'RS';
  if (ext === 'py') return 'PY';
  if (ext === 'css') return 'CSS';
  if (ext === 'md') return 'MD';
  return '•';
}

/** Cursor-like vertical activity feed: diffs, explores, terminals. */
export const ActivityStream: React.FC<ActivityStreamProps> = ({
  tools,
  fileChanges,
  filesRead,
  isWorking,
  followUpSlot,
}) => {
  const [expandedDiffs, setExpandedDiffs] = useState<Record<string, boolean>>({});

  const blocks = useMemo(() => {
    const out: ActivityBlock[] = [];
    if (filesRead.length > 0) {
      out.push({
        kind: 'status',
        id: 'explored',
        text: `Explored ${filesRead.length} file${filesRead.length === 1 ? '' : 's'}`,
      });
    }
    for (const change of fileChanges) {
      const stats = diffStats(change.diff);
      out.push({ kind: 'diff', id: change.id, change, added: stats.added, removed: stats.removed });
    }
    for (const tool of tools) {
      const name = (tool.name || '').toLowerCase();
      const isShell = name === 'shell' || name === 'test' || name === 'git' || /command|terminal/.test(tool.step || '');
      if (isShell) {
        out.push({
          kind: 'terminal',
          id: tool.id,
          title: tool.step || tool.name,
          output: tool.output || '',
          failed: tool.status === 'failed',
          running: tool.status === 'running',
        });
      } else if (tool.status === 'running' || tool.output) {
        out.push({ kind: 'tool', id: tool.id, tool });
      }
    }
    return out;
  }, [tools, fileChanges, filesRead]);

  if (blocks.length === 0 && !isWorking) return null;

  return (
    <div className="activity-stream">
      <div className="activity-stream-feed">
        {blocks.map((b) => {
          if (b.kind === 'status') {
            return (
              <div key={b.id} className="activity-status-line">
                {b.text}
              </div>
            );
          }
          if (b.kind === 'diff') {
            const name = b.change.path.split(/[/\\]/).pop() || b.change.path;
            const open = expandedDiffs[b.id] ?? true;
            return (
              <div key={b.id} className="activity-diff-card">
                <button
                  type="button"
                  className="activity-diff-header"
                  onClick={() => setExpandedDiffs((m) => ({ ...m, [b.id]: !open }))}
                >
                  <span className="activity-file-badge">{fileBadge(b.change.path)}</span>
                  <span
                    className="activity-diff-name"
                    onClick={(e) => {
                      e.stopPropagation();
                      agentHost.openFile(b.change.path);
                    }}
                  >
                    {name}
                  </span>
                  <span className="activity-diff-stats">
                    {b.added > 0 && <span className="add">+{b.added}</span>}
                    {b.removed > 0 && <span className="del">-{b.removed}</span>}
                  </span>
                  {open ? <ChevronDownIcon className="icon-xs" /> : <ChevronRightIcon className="icon-xs" />}
                </button>
                {open && b.change.diff && <DiffView diffText={b.change.diff} />}
                {open && !b.change.diff && (
                  <div className="activity-diff-empty">
                    <DocumentIcon className="icon-sm" />
                    <span>{b.change.type} · no inline diff</span>
                    <button type="button" className="chip-btn" onClick={() => agentHost.openDiff(b.change.path)}>
                      Open diff
                    </button>
                  </div>
                )}
              </div>
            );
          }
          if (b.kind === 'terminal') {
            return (
              <div key={b.id} className={`activity-terminal-card ${b.failed ? 'failed' : ''}`}>
                <div className="activity-terminal-header">
                  <CommandLineIcon className="icon-sm" />
                  <span className={`activity-terminal-title ${b.running ? 'live-activity-shimmer' : ''}`}>
                    {b.title}
                  </span>
                  {b.running && <ActivityIndicator state="live" />}
                  {b.failed && <span className="activity-terminal-dot" />}
                </div>
                {b.output && <pre className="activity-terminal-body">{b.output}</pre>}
              </div>
            );
          }
          return (
            <div key={b.id} className="activity-tool-line">
              <ActivityIndicator
                state={
                  b.tool.status === 'running'
                    ? 'live'
                    : b.tool.status === 'failed'
                      ? 'error'
                      : b.tool.status === 'completed'
                        ? 'done'
                        : 'idle'
                }
              />
              <span className={b.tool.status === 'running' ? 'live-activity-shimmer' : undefined}>
                {b.tool.step || b.tool.name}
              </span>
            </div>
          );
        })}
      </div>

      {followUpSlot}
    </div>
  );
};
