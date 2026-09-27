package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"blackjak/workspace"
)

// RunStatus represents the current state of an agent execution run.
type RunStatus string

const (
	RunPending   RunStatus = "pending"
	RunRunning   RunStatus = "running"
	RunWaiting   RunStatus = "waiting" // Waiting for human approval
	RunCompleted RunStatus = "completed"
	RunFailed    RunStatus = "failed"
	RunCancelled RunStatus = "cancelled"
)

// ApprovalRequest represents a request for human intervention.
type ApprovalRequest struct {
	ID          string `json:"id"`
	RunID       string `json:"runId"`
	Operation   string `json:"operation"`   // e.g. "command", "file_delete"
	Description string `json:"description"` // e.g. "Execute 'rm -rf node_modules'"
	Details     any    `json:"details,omitempty"`
}

// ApprovalResponse represents the human user's decision on an approval request.
type ApprovalResponse struct {
	RequestID string `json:"requestId"`
	Granted   bool   `json:"granted"`
	Reason    string `json:"reason,omitempty"`
}

// Run represents an individual session/task run of the agent.
type Run struct {
	ID           string                         `json:"id"`
	Prompt       string                         `json:"prompt"`
	Workspace    string                         `json:"workspace"`
	Status       RunStatus                      `json:"status"`
	Error        string                         `json:"error,omitempty"`
	CreatedAt    time.Time                      `json:"createdAt"`
	UpdatedAt    time.Time                      `json:"updatedAt"`
	Plan         *Plan                          `json:"plan,omitempty"`
	Events       []Event                        `json:"events"`
	Subagents    []*Subagent                    `json:"subagents"`
	FileChanges  []workspace.FileChange         `json:"fileChanges"`
	Attachments  []workspace.Attachment         `json:"attachments,omitempty"`
	References   []workspace.WorkspaceReference `json:"references,omitempty"`
	PendingReq   *ApprovalRequest               `json:"pendingApproval,omitempty"`
	Orchestration*Orchestrator                  `json:"-"`
	approvalCh   chan ApprovalResponse          `json:"-"`
	ctx          context.Context                `json:"-"`
	cancel       context.CancelFunc             `json:"-"`
	mu           sync.RWMutex                   `json:"-"`
}

// Context returns the execution context associated with this run.
func (r *Run) Context() context.Context {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctx
}

// GetEvents returns a thread-safe copy of events emitted during this run.
func (r *Run) GetEvents() []Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]Event, len(r.Events))
	copy(res, r.Events)
	return res
}

// AddEvent appends an event to the run history.
func (r *Run) AddEvent(evt Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Events = append(r.Events, evt)
	r.UpdatedAt = time.Now()
}

// AddFileChange appends a workspace file change.
func (r *Run) AddFileChange(fc workspace.FileChange) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.FileChanges = append(r.FileChanges, fc)
	r.UpdatedAt = time.Now()
}

// SetStatus updates the status of the run.
func (r *Run) SetStatus(status RunStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Status = status
	r.UpdatedAt = time.Now()
}

// RunManager manages the creation, storage, cancellation, and retrieval of agent runs.
type RunManager struct {
	mu      sync.RWMutex
	runs    map[string]*Run
	broker  *EventBroker
	tracker *workspace.FileTrackerManager
}

// NewRunManager initializes a RunManager.
func NewRunManager(broker *EventBroker) *RunManager {
	return &RunManager{
		runs:    make(map[string]*Run),
		broker:  broker,
		tracker: workspace.NewFileTrackerManager(),
	}
}

// GetTracker returns the FileTrackerManager associated with the RunManager.
func (m *RunManager) GetTracker() *workspace.FileTrackerManager {
	return m.tracker
}

// CreateRun initializes and registers a new agent run with optional attachments.
func (m *RunManager) CreateRun(prompt string, workspacePath string, attachments ...workspace.Attachment) *Run {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("run_%d", time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())

	wsObj := workspace.New(workspacePath)
	refs, _ := wsObj.ParseWorkspaceReferences(prompt)
	validAtts, _ := wsObj.ValidateAttachments(attachments)

	now := time.Now()
	orch := NewOrchestrator(id, m.broker, m.tracker)

	run := &Run{
		ID:            id,
		Prompt:        prompt,
		Workspace:     workspacePath,
		Status:        RunPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		Events:        make([]Event, 0),
		Subagents:     make([]*Subagent, 0),
		FileChanges:   make([]workspace.FileChange, 0),
		Attachments:   validAtts,
		References:    refs,
		Orchestration: orch,
		approvalCh:    make(chan ApprovalResponse, 1),
		ctx:           ctx,
		cancel:        cancel,
	}

	m.runs[id] = run

	evt := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		RunID:     id,
		Type:      EventRunCreated,
		Timestamp: now,
		Data: map[string]interface{}{
			"prompt":      prompt,
			"workspace":   workspacePath,
			"attachments": validAtts,
			"references":  refs,
		},
	}
	run.AddEvent(evt)
	if m.broker != nil {
		m.broker.Publish(evt)
	}

	return run
}

// GetRun retrieves a run by ID.
func (m *RunManager) GetRun(id string) (*Run, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	run, ok := m.runs[id]
	return run, ok
}

// ListRuns returns all runs in reverse chronological order.
func (m *RunManager) ListRuns() []*Run {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Run, 0, len(m.runs))
	for _, r := range m.runs {
		list = append(list, r)
	}
	return list
}

// CancelRun signals context cancellation for a running task.
func (m *RunManager) CancelRun(id string) bool {
	m.mu.Lock()
	run, ok := m.runs[id]
	m.mu.Unlock()

	if !ok {
		return false
	}

	run.mu.Lock()
	if run.Status == RunCompleted || run.Status == RunFailed || run.Status == RunCancelled {
		run.mu.Unlock()
		return false
	}
	run.Status = RunCancelled
	run.UpdatedAt = time.Now()
	if run.cancel != nil {
		run.cancel()
	}
	run.mu.Unlock()

	evt := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		RunID:     id,
		Type:      EventRunCancelled,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"reason": "Cancelled by user"},
	}
	run.AddEvent(evt)
	if m.broker != nil {
		m.broker.Publish(evt)
	}

	return true
}

// SubmitApproval sends an approval response for a waiting run.
func (m *RunManager) SubmitApproval(runID string, response ApprovalResponse) error {
	m.mu.RLock()
	run, ok := m.runs[runID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("run not found: %s", runID)
	}

	run.mu.Lock()
	if run.Status != RunWaiting {
		run.mu.Unlock()
		return fmt.Errorf("run is not waiting for approval")
	}
	run.PendingReq = nil
	run.mu.Unlock()

	select {
	case run.approvalCh <- response:
		evtType := EventApprovalGranted
		if !response.Granted {
			evtType = EventApprovalDenied
		}
		evt := Event{
			ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
			RunID:     runID,
			Type:      evtType,
			Timestamp: time.Now(),
			Data:      response,
		}
		run.AddEvent(evt)
		if m.broker != nil {
			m.broker.Publish(evt)
		}
		return nil
	default:
		return fmt.Errorf("approval response channel full or closed")
	}
}

// RequestApproval triggers an approval request event and blocks until response or context done.
func (m *RunManager) RequestApproval(ctx context.Context, run *Run, req ApprovalRequest) (bool, string, error) {
	run.mu.Lock()
	run.Status = RunWaiting
	run.PendingReq = &req
	run.UpdatedAt = time.Now()
	run.mu.Unlock()

	evt := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		RunID:     run.ID,
		Type:      EventApprovalRequired,
		Timestamp: time.Now(),
		Data:      req,
	}
	run.AddEvent(evt)
	if m.broker != nil {
		m.broker.Publish(evt)
	}

	select {
	case <-ctx.Done():
		run.SetStatus(RunCancelled)
		return false, "context cancelled", ctx.Err()
	case resp := <-run.approvalCh:
		run.SetStatus(RunRunning)
		return resp.Granted, resp.Reason, nil
	}
}
