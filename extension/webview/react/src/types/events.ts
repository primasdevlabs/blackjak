export const PROTOCOL_VERSION = "1";

export type RunStatus = 
  | 'pending'
  | 'running'
  | 'waiting'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'paused';

export type SubagentStatus =
  | 'created'
  | 'queued'
  | 'running'
  | 'waiting'
  | 'completed'
  | 'failed'
  | 'cancelled';

export type EventType =
  | 'run.created'
  | 'run.started'
  | 'run.completed'
  | 'run.failed'
  | 'run.cancelled'
  | 'run.paused'
  | 'run.resumed'
  | 'agent.created'
  | 'agent.started'
  | 'agent.waiting'
  | 'agent.completed'
  | 'agent.failed'
  | 'agent.cancelled'
  | 'agent.delegated'
  | 'agent.handoff'
  | 'agent.activity'
  | 'agent.message'
  | 'agent.thinking'
  | 'agent.plan'
  | 'agent.phase'
  | 'tool.started'
  | 'tool.output'
  | 'tool.completed'
  | 'tool.failed'
  | 'file.read'
  | 'file.created'
  | 'file.modified'
  | 'file.deleted'
  | 'file.renamed'
  | 'file.moved'
  | 'file.change.reviewed'
  | 'workspace.reference'
  | 'workspace.attachment'
  | 'workspace.conflict'
  | 'workspace.locked'
  | 'workspace.unlocked'
  | 'workspace.openFile'
  | 'workspace.revealFile'
  | 'workspace.openFolder'
  | 'diff.available'
  | 'command.started'
  | 'command.output'
  | 'command.completed'
  | 'test.started'
  | 'test.output'
  | 'test.passed'
  | 'test.failed'
  | 'approval.required'
  | 'approval.granted'
  | 'approval.denied'
  | 'context.updated'
  | 'memory.updated';

export interface Finding {
  file: string;
  line?: number;
  message: string;
}

export interface Subagent {
  id: string;
  parentRunId: string;
  role: string;
  task: string;
  status: SubagentStatus;
  workspaceScope?: string[];
  startedAt: string;
  finishedAt?: string;
  result?: string;
  error?: string;
  findings?: Finding[];
  activity?: string;
}

export interface Attachment {
  id: string;
  type: 'file' | 'folder' | 'image';
  path: string;
  name: string;
  /** Blob or served URL for image thumbnails / lightbox. */
  previewUrl?: string;
  mime?: string;
}

export interface WorkspaceReference {
  type: 'file' | 'folder';
  path: string;
  line?: number;
  raw: string;
}

export interface FileChange {
  id: string;
  runId: string;
  agentId: string;
  type: 'created' | 'modified' | 'deleted' | 'renamed' | 'moved';
  path: string;
  previousPath?: string;
  diff?: string;
  status?: 'pending' | 'accepted' | 'rejected';
  canRevert?: boolean;
  timestamp: string;
}

export interface AgentEvent {
  id: string;
  runId: string;
  agentId?: string;
  type: EventType;
  timestamp: string;
  data?: any;
}

export interface Plan {
  steps: string[];
  currentStep: number;
}

export interface ApprovalRequest {
  id: string;
  runId: string;
  operation: string;
  description: string;
  details?: any;
}

export interface ApprovalResponse {
  requestId: string;
  granted: boolean;
  reason?: string;
}

export interface RunDTO {
  id: string;
  prompt: string;
  workspace: string;
  status: RunStatus;
  error?: string;
  createdAt: string;
  updatedAt: string;
  plan?: Plan;
  events: AgentEvent[];
  subagents: Subagent[];
  fileChanges: FileChange[];
  attachments?: Attachment[];
  references?: WorkspaceReference[];
  pendingApproval?: ApprovalRequest;
}

export interface ServerMessage {
  protocolVersion: string;
  type: string;
  runId?: string;
  agentId?: string;
  timestamp: string;
  data?: any;
  error?: string;
}

export interface ClientMessage {
  type: 'run.start' | 'run.cancel' | 'approval.respond' | 'workspace.command' | 'ping';
  requestId?: string;
  data?: any;
}

export interface HealthResponse {
  status: string;
  version: string;
  protocolVersion: string;
  workspace: string;
  timestamp: string;
}
