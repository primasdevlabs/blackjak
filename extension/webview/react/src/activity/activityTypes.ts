export type ActivityCategory =
  | 'orchestrating'
  | 'exploring'
  | 'thinking'
  | 'debugging'
  | 'testing'
  | 'editing'
  | 'finishing';

export type ActivityState =
  | 'idle'
  | 'working'
  | 'waiting'
  | 'approval'
  | 'completed'
  | 'error';

export interface ActivityConfig {
  minIntervalMs: number;
  maxIntervalMs: number;
  enabled: boolean;
}

/** A finished phase with its elapsed duration, e.g. "Reviewing project structure — 2m 40s". */
export interface ActivityPhase {
  category: ActivityCategory;
  /** The ambient message that was showing when the phase ended. */
  message: string;
  durationMs: number;
}

export interface ActivityStatus {
  state: ActivityState;
  category: ActivityCategory;
  ambientMessage: string;
  concreteMessage?: string;
  /** Epoch ms when the current run started working. */
  workingSince?: number;
  /** Epoch ms when the current category/phase started. */
  categorySince?: number;
  /** Completed phases in order, oldest first. */
  phases: ActivityPhase[];
  lastUpdated: Date;
}

/** Format an elapsed duration compactly: "5s", "2m 40s", "1h 12m". */
export function formatElapsed(ms: number): string {
  if (ms < 0) ms = 0;
  const s = Math.floor(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  if (m < 60) {
    const rem = s % 60;
    return rem > 0 ? `${m}m ${rem}s` : `${m}m`;
  }
  const h = Math.floor(m / 60);
  const remM = m % 60;
  return remM > 0 ? `${h}h ${remM}m` : `${h}h`;
}

/** Past-tense label for a finished phase, e.g. exploring -> "Explored". */
export function phaseLabel(cat: ActivityCategory, message?: string): string {
  if (message) return message.replace(/…+$/, '').replace(/\.+$/, '');
  switch (cat) {
    case 'exploring': return 'Explored';
    case 'editing': return 'Edited';
    case 'thinking': return 'Analyzed';
    case 'testing': return 'Validated';
    case 'debugging': return 'Debugged';
    case 'finishing': return 'Completed';
    default: return 'Coordinated';
  }
}
