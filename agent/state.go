package agent

// Status represents the operational status of the agent.
type Status string

const (
	StatusIdle      Status = "idle"
	StatusPlanning  Status = "planning"
	StatusExecuting Status = "executing"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// AgentState holds execution state for the agent lifecycle.
type AgentState struct {
	ID     string
	Status Status
}

// NewAgentState initializes a new AgentState instance.
func NewAgentState(id string) *AgentState {
	return &AgentState{
		ID:     id,
		Status: StatusIdle,
	}
}
