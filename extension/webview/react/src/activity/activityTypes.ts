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

export interface ActivityStatus {
  state: ActivityState;
  category: ActivityCategory;
  ambientMessage: string;
  concreteMessage?: string;
  lastUpdated: Date;
}
