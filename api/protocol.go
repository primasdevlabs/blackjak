package api

import (
	"time"

	"blackjak/agent"
	"blackjak/workspace"
)

// ProtocolVersion specifies the wire protocol schema version.
const ProtocolVersion = "1"

// ClientMessageType defines messages sent from client (UI/IDE) to server.
type ClientMessageType string

const (
	ClientMsgRunStart        ClientMessageType = "run.start"
	ClientMsgRunCancel       ClientMessageType = "run.cancel"
	ClientMsgApprovalRespond  ClientMessageType = "approval.respond"
	ClientMsgWorkspaceCommand ClientMessageType = "workspace.command"
	ClientMsgPing            ClientMessageType = "ping"
)

// ClientMessage represents an incoming message over WebSocket.
type ClientMessage struct {
	Type      ClientMessageType `json:"type"`
	RequestID string            `json:"requestId,omitempty"`
	Data      map[string]any    `json:"data,omitempty"`
}

// ServerMessage represents an outgoing frame over WebSocket.
type ServerMessage struct {
	ProtocolVersion string    `json:"protocolVersion"`
	Type            string    `json:"type"`
	RunID           string    `json:"runId,omitempty"`
	AgentID         string    `json:"agentId,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	Data            any       `json:"data,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// CreateRunPayload represents POST /api/runs request body.
type CreateRunPayload struct {
	Prompt      string                 `json:"prompt"`
	Workspace   string                 `json:"workspace,omitempty"`
	Attachments []workspace.Attachment `json:"attachments,omitempty"`
}

// ApprovalPayload represents POST /api/runs/:id/approval request body.
type ApprovalPayload struct {
	RequestID string `json:"requestId"`
	Granted   bool   `json:"granted"`
	Reason    string `json:"reason,omitempty"`
}

// APIResponse wraps standardized JSON API outputs.
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse returned by GET /health.
type HealthResponse struct {
	Status          string    `json:"status"`
	Version         string    `json:"version"`
	ProtocolVersion string    `json:"protocolVersion"`
	Workspace       string    `json:"workspace"`
	Timestamp       time.Time `json:"timestamp"`
}

// RunDTO represents serialized run state for responses.
type RunDTO struct {
	ID          string                         `json:"id"`
	Prompt      string                         `json:"prompt"`
	Workspace   string                         `json:"workspace"`
	Status      agent.RunStatus                `json:"status"`
	Error       string                         `json:"error,omitempty"`
	CreatedAt   time.Time                      `json:"createdAt"`
	UpdatedAt   time.Time                      `json:"updatedAt"`
	Plan        *agent.Plan                    `json:"plan,omitempty"`
	Events      []agent.Event                  `json:"events"`
	Subagents   []*agent.Subagent              `json:"subagents"`
	FileChanges []workspace.FileChange         `json:"fileChanges"`
	Attachments []workspace.Attachment         `json:"attachments,omitempty"`
	References  []workspace.WorkspaceReference `json:"references,omitempty"`
	PendingReq  *agent.ApprovalRequest         `json:"pendingApproval,omitempty"`
}
