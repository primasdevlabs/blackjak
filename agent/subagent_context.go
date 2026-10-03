package agent

import (
	"blackjak/workspace"
)

// SubagentContext provides an isolated execution environment for a subagent.
// Each subagent receives its own context containing scoped workspace access,
// sandbox enforcement, tool permission checks, a message inbox, and an event stream.
type SubagentContext struct {
	// Subagent is the subagent this context belongs to.
	Subagent *Subagent

	// Workspace provides workspace boundary validation.
	Workspace *workspace.Workspace

	// Sandbox enforces scope restrictions for file operations.
	Sandbox *workspace.Sandbox

	// ToolPermissions lists tools this subagent is allowed to invoke.
	// An empty slice means all tools are allowed.
	ToolPermissions []string

	// Inbox receives structured messages from the orchestrator or other subagents.
	Inbox chan AgentMessage

	// EventStream receives events relevant to this subagent's execution.
	EventStream chan Event
}

// DefaultToolPermissions returns the default tool set for a given role.
// Names match registered tools (filesystem, search, shell, git, test, memory).
// Roles not in this map get full tool access (nil).
func DefaultToolPermissions(role string) []string {
	return RoleToolAllowlist(role)
}

// HasToolPermission checks whether the subagent context allows a specific tool.
func (sc *SubagentContext) HasToolPermission(toolName string) bool {
	if sc.ToolPermissions == nil || len(sc.ToolPermissions) == 0 {
		return true
	}
	for _, allowed := range sc.ToolPermissions {
		if allowed == toolName {
			return true
		}
	}
	return false
}

// NewSubagentContext creates an isolated context for a subagent.
func NewSubagentContext(sub *Subagent, ws *workspace.Workspace) *SubagentContext {
	sandbox := workspace.NewScopedSandbox(ws, sub.WorkspaceScope)
	perms := DefaultToolPermissions(sub.Role)

	return &SubagentContext{
		Subagent:        sub,
		Workspace:       ws,
		Sandbox:         sandbox,
		ToolPermissions: perms,
		Inbox:           make(chan AgentMessage, 64),
		EventStream:     make(chan Event, 256),
	}
}
