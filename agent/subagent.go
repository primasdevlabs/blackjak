package agent

import (
	"sync"
	"time"
)

// SubagentPaused indicates the subagent has been temporarily suspended by the orchestrator.
const SubagentPaused SubagentStatus = "paused"

// SubagentStatus defines the lifecycle states of a subagent.
type SubagentStatus string

const (
	SubagentCreated   SubagentStatus = "created"
	SubagentQueued    SubagentStatus = "queued"
	SubagentRunning   SubagentStatus = "running"
	SubagentWaiting   SubagentStatus = "waiting"
	SubagentCompleted SubagentStatus = "completed"
	SubagentFailed    SubagentStatus = "failed"
	SubagentCancelled SubagentStatus = "cancelled"
)

// Finding represents an insight or issue discovered by a subagent.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

// MessageType defines the nature of subagent communication.
type MessageType string

const (
	MsgTypeTask     MessageType = "task"
	MsgTypeResult   MessageType = "result"
	MsgTypeFinding  MessageType = "finding"
	MsgTypeQuestion MessageType = "question"
	MsgTypeRequest  MessageType = "request"
	MsgTypeHandoff  MessageType = "handoff"
	MsgTypeWarning  MessageType = "warning"
	MsgTypeError    MessageType = "error"
)

// AgentMessage represents structured communication between agents.
type AgentMessage struct {
	FromAgentID string      `json:"fromAgentId"`
	ToAgentID   string      `json:"toAgentId"`
	Type        MessageType `json:"type"`
	Data        any         `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Subagent represents a first-class specialized worker subagent.
type Subagent struct {
	ID              string         `json:"id"`
	ParentRunID     string         `json:"parentRunId"`
	Role            string         `json:"role"`
	Task            string         `json:"task"`
	Status          SubagentStatus `json:"status"`
	WorkspaceScope  []string       `json:"workspaceScope,omitempty"`
	ToolPermissions []string       `json:"toolPermissions,omitempty"`
	StartedAt       time.Time      `json:"startedAt"`
	FinishedAt      *time.Time     `json:"finishedAt,omitempty"`
	Result          string         `json:"result,omitempty"`
	Error           string         `json:"error,omitempty"`
	Findings        []Finding      `json:"findings,omitempty"`
	Activity        string         `json:"activity,omitempty"`
	Inbox           chan AgentMessage `json:"-"`
	mu              sync.RWMutex   `json:"-"`
	pauseMu         sync.Mutex     `json:"-"`
	pauseCond       *sync.Cond     `json:"-"`
	paused          bool           `json:"-"`
}

// NewSubagent initializes a new Subagent instance.
func NewSubagent(id, parentRunID, role, task string, scope []string) *Subagent {
	s := &Subagent{
		ID:              id,
		ParentRunID:     parentRunID,
		Role:            role,
		Task:            task,
		Status:          SubagentCreated,
		WorkspaceScope:  scope,
		ToolPermissions: DefaultToolPermissions(role),
		StartedAt:       time.Now(),
		Findings:        make([]Finding, 0),
		Inbox:           make(chan AgentMessage, 64),
	}
	s.pauseCond = sync.NewCond(&s.pauseMu)
	return s
}

// GetStatus returns the thread-safe status of the subagent.
func (s *Subagent) GetStatus() SubagentStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// SetStatus updates the status and handles timestamping upon completion.
func (s *Subagent) SetStatus(status SubagentStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
	if status == SubagentCompleted || status == SubagentFailed || status == SubagentCancelled {
		now := time.Now()
		s.FinishedAt = &now
	}
}

// SetActivity updates the ambient activity string for this subagent.
func (s *Subagent) SetActivity(activity string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Activity = activity
}

// AddFinding appends a discovered finding to the subagent.
func (s *Subagent) AddFinding(f Finding) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Findings = append(s.Findings, f)
}

// SetResult records the outcome of the subagent task execution.
func (s *Subagent) SetResult(result string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Result = result
}

// SetError records an execution failure error.
func (s *Subagent) SetError(errStr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Error = errStr
	s.Status = SubagentFailed
}

// Pause suspends the subagent. Execution functions should call WaitIfPaused()
// at safe checkpoints to honour pause requests.
func (s *Subagent) Pause() {
	s.pauseMu.Lock()
	s.paused = true
	s.pauseMu.Unlock()
	s.SetStatus(SubagentPaused)
}

// Resume unpauses a paused subagent and wakes its execution goroutine.
func (s *Subagent) Resume() {
	s.pauseMu.Lock()
	s.paused = false
	s.pauseCond.Broadcast()
	s.pauseMu.Unlock()
	s.SetStatus(SubagentRunning)
}

// WaitIfPaused blocks the calling goroutine while the subagent is paused.
// Call this at safe checkpoints inside subagent execution functions.
func (s *Subagent) WaitIfPaused() {
	s.pauseMu.Lock()
	for s.paused {
		s.pauseCond.Wait()
	}
	s.pauseMu.Unlock()
}

// IsPaused returns whether the subagent is currently paused.
func (s *Subagent) IsPaused() bool {
	s.pauseMu.Lock()
	defer s.pauseMu.Unlock()
	return s.paused
}

// SendMessage delivers a message to the subagent's inbox (non-blocking).
func (s *Subagent) SendMessage(msg AgentMessage) {
	select {
	case s.Inbox <- msg:
	default:
		// Inbox full — drop message to avoid blocking the sender
	}
}
