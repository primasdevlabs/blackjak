package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"blackjak/llm"
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
	RunPaused    RunStatus = "paused"
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
	compactCh    chan struct{}                  `json:"-"`
	messages     []llm.Message                  `json:"-"`
	pauseRequested bool                         `json:"-"`
	ctx          context.Context                `json:"-"`
	cancel       context.CancelFunc             `json:"-"`
	mu           sync.RWMutex                   `json:"-"`
}

// RequestPause marks the run for a graceful pause: the loop stops at the
// next boundary and the context is cancelled to abort an in-flight call.
func (r *Run) RequestPause() {
	r.mu.Lock()
	r.pauseRequested = true
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Unlock()
}

// IsPauseRequested reports whether a pause has been signalled.
func (r *Run) IsPauseRequested() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pauseRequested
}

// ResetContext gives a resumed run a fresh cancellable context and clears
// the pause flag. The caller must hold no lock; used before re-executing.
func (r *Run) ResetContext() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.pauseRequested = false
	r.PendingReq = nil
}

// SetMessages stores the loop's current conversation so command handlers can
// inspect or compact it.
func (r *Run) SetMessages(msgs []llm.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = msgs
}

// GetMessages returns a copy of the loop's current conversation.
func (r *Run) GetMessages() []llm.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]llm.Message, len(r.messages))
	copy(out, r.messages)
	return out
}

// RequestCompaction asks the running loop to compact its context at the next
// iteration. Returns false if there is no active loop or a request is pending.
func (r *Run) RequestCompaction() bool {
	select {
	case r.compactCh <- struct{}{}:
		return true
	default:
		return false
	}
}

// TakeCompactionRequest reports and clears a pending compaction request.
// Called by the agent loop between iterations.
func (r *Run) TakeCompactionRequest() bool {
	select {
	case <-r.compactCh:
		return true
	default:
		return false
	}
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

// ReviewFileChange applies a human review decision to a tracked file change.
// "accept" keeps the change on disk; "reject" reverts it — deleting created
// files or restoring the pre-change snapshot for modified ones.
func (m *RunManager) ReviewFileChange(runID, changeID, action string) (*workspace.FileChange, error) {
	if action != "accept" && action != "reject" {
		return nil, fmt.Errorf("invalid action %q — expected accept or reject", action)
	}

	run, ok := m.GetRun(runID)
	if !ok {
		return nil, fmt.Errorf("run '%s' not found", runID)
	}

	run.mu.Lock()
	defer run.mu.Unlock()

	idx := -1
	for i := range run.FileChanges {
		if run.FileChanges[i].ID == changeID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("change '%s' not found", changeID)
	}
	fc := run.FileChanges[idx]
	if fc.Status == workspace.ReviewRejected {
		return nil, fmt.Errorf("change '%s' was already declined and its content reverted", changeID)
	}
	if fc.Status == workspace.ReviewAccepted && action == "accept" {
		return nil, fmt.Errorf("change '%s' was already accepted", changeID)
	}

	abs := fc.Path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(run.Workspace, abs)
	}
	rel, err := filepath.Rel(run.Workspace, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("change path escapes the workspace")
	}

	var status workspace.ReviewStatus
	if action == "accept" {
		status = workspace.ReviewAccepted
	} else {
		if !fc.CanRevert {
			return nil, fmt.Errorf("change '%s' cannot be reverted (no snapshot)", changeID)
		}
		if err := revertFileChange(fc, abs); err != nil {
			return nil, fmt.Errorf("revert failed: %w", err)
		}
		status = workspace.ReviewRejected
	}

	run.FileChanges[idx].Status = status
	run.UpdatedAt = time.Now()
	updated := run.FileChanges[idx]

	m.tracker.SetStatus(runID, changeID, status)

	m.broker.Publish(Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		RunID:     runID,
		Type:      EventFileChangeReviewed,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"changeId": changeID,
			"path":     fc.Path,
			"action":   action,
			"status":   string(status),
		},
	})

	return &updated, nil
}

// revertFileChange restores a file to its pre-change state.
func revertFileChange(fc workspace.FileChange, abs string) error {
	switch fc.Type {
	case workspace.ChangeCreated:
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	case workspace.ChangeModified, workspace.ChangeDeleted:
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		return os.WriteFile(abs, []byte(fc.PreviousContent), 0o644)
	default:
		return fmt.Errorf("revert not supported for change type %q", fc.Type)
	}
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
		compactCh:     make(chan struct{}, 1),
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

// LatestRun returns the most recently created run, preferring a live one.
func (m *RunManager) LatestRun() *Run {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var live, newest *Run
	for _, r := range m.runs {
		r.mu.RLock()
		status, created := r.Status, r.CreatedAt
		r.mu.RUnlock()
		if status == RunRunning || status == RunWaiting || status == RunPending {
			if live == nil || created.After(live.CreatedAt) {
				live = r
			}
			continue
		}
		if newest == nil || created.After(newest.CreatedAt) {
			newest = r
		}
	}
	if live != nil {
		return live
	}
	return newest
}

// DeleteRun removes a run from history. A live run is cancelled first.
func (m *RunManager) DeleteRun(id string) bool {
	m.mu.Lock()
	run, ok := m.runs[id]
	if !ok {
		m.mu.Unlock()
		return false
	}
	delete(m.runs, id)
	m.mu.Unlock()

	run.mu.Lock()
	live := run.Status == RunRunning || run.Status == RunWaiting || run.Status == RunPending
	if live {
		run.Status = RunCancelled
		run.UpdatedAt = time.Now()
		if run.cancel != nil {
			run.cancel()
		}
	}
	run.mu.Unlock()
	m.DeleteCheckpoint(run)
	return true
}

// ClearFinished removes completed/failed/cancelled runs from history.
// Returns the number of runs removed.
func (m *RunManager) ClearFinished() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	removed := 0
	for id, r := range m.runs {
		r.mu.RLock()
		status := r.Status
		r.mu.RUnlock()
		if status == RunCompleted || status == RunFailed || status == RunCancelled {
			delete(m.runs, id)
			removed++
		}
	}
	return removed
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
