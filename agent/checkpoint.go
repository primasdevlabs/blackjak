package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blackjak/llm"
	"blackjak/workspace"
)

// RunCheckpoint is the durable snapshot of a run's conversation state.
// It is written after every message mutation so a run can be paused,
// stopped, or interrupted and later resumed with full context — including
// across backend restarts.
type RunCheckpoint struct {
	RunID       string                 `json:"runId"`
	Prompt      string                 `json:"prompt"`
	Workspace   string                 `json:"workspace"`
	Status      RunStatus              `json:"status"`
	Messages    []llm.Message          `json:"messages"`
	Plan        *Plan                  `json:"plan,omitempty"`
	FileChanges []workspace.FileChange `json:"fileChanges,omitempty"`
	Subagents   []*Subagent            `json:"subagents,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// checkpointDir returns <workspace>/.blackjak/runs.
func checkpointDir(workspacePath string) string {
	return filepath.Join(workspacePath, ".blackjak", "runs")
}

// SaveCheckpoint persists the run's current message state to disk.
// Writes are atomic (tmp file + rename) so a crash mid-write can't leave
// a corrupt checkpoint.
func (m *RunManager) SaveCheckpoint(run *Run, messages []llm.Message) {
	if run == nil || run.Workspace == "" {
		return
	}
	run.mu.RLock()
	cp := RunCheckpoint{
		RunID:       run.ID,
		Prompt:      run.Prompt,
		Workspace:   run.Workspace,
		Status:      run.Status,
		Plan:        run.Plan,
		CreatedAt:   run.CreatedAt,
		UpdatedAt:   time.Now(),
		FileChanges: run.FileChanges,
	}
	run.mu.RUnlock()
	if run.Orchestration != nil {
		cp.Subagents = run.Orchestration.ListSubagents()
	} else {
		cp.Subagents = run.Subagents
	}

	cp.Messages = make([]llm.Message, len(messages))
	copy(cp.Messages, messages)

	dir := checkpointDir(run.Workspace)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	data, err := json.Marshal(cp)
	if err != nil {
		return
	}
	tmp := filepath.Join(dir, run.ID+".tmp")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, filepath.Join(dir, run.ID+".json"))
}

// LoadCheckpoint reads a run's checkpoint from its workspace.
func (m *RunManager) LoadCheckpoint(runID, workspacePath string) (*RunCheckpoint, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("no workspace recorded for run '%s'", runID)
	}
	data, err := os.ReadFile(filepath.Join(checkpointDir(workspacePath), runID+".json"))
	if err != nil {
		return nil, err
	}
	var cp RunCheckpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("corrupt checkpoint for run '%s': %w", runID, err)
	}
	return &cp, nil
}

// DeleteCheckpoint removes a run's persisted checkpoint.
func (m *RunManager) DeleteCheckpoint(run *Run) {
	if run == nil || run.Workspace == "" {
		return
	}
	_ = os.Remove(filepath.Join(checkpointDir(run.Workspace), run.ID+".json"))
}

// RestoreFromWorkspace scans <root>/.blackjak/runs and rehydrates runs that
// aren't already in memory. Any checkpoint left in a live state (running,
// pending, waiting) is restored as paused — nothing survives a restart.
// Returns the number of runs restored.
func (m *RunManager) RestoreFromWorkspace(root string) int {
	entries, err := os.ReadDir(checkpointDir(root))
	if err != nil {
		return 0
	}
	restored := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		runID := strings.TrimSuffix(e.Name(), ".json")
		if _, exists := m.GetRun(runID); exists {
			continue
		}
		data, err := os.ReadFile(filepath.Join(checkpointDir(root), e.Name()))
		if err != nil {
			continue
		}
		var cp RunCheckpoint
		if err := json.Unmarshal(data, &cp); err != nil {
			continue
		}
		m.restoreRun(&cp)
		restored++
	}
	return restored
}

// restoreRun rebuilds a Run from a checkpoint. Snapshots inside file
// changes were never persisted, so modified changes are no longer
// revertible after a restart — created files still are (revert = delete).
func (m *RunManager) restoreRun(cp *RunCheckpoint) *Run {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := cp.Status
	if status == RunRunning || status == RunPending || status == RunWaiting {
		status = RunPaused
	}

	// Reverts for 'modified'/'deleted' changes need the pre-change snapshot
	// which is intentionally never serialized — mark them unrevertible.
	for i := range cp.FileChanges {
		if cp.FileChanges[i].Type != workspace.ChangeCreated {
			cp.FileChanges[i].CanRevert = false
		}
		if cp.FileChanges[i].Status == "" {
			cp.FileChanges[i].Status = workspace.ReviewPending
		}
	}

	orch := NewOrchestrator(cp.RunID, m.broker, m.tracker)
	for _, s := range cp.Subagents {
		orch.Adopt(s)
	}

	ctx, cancel := context.WithCancel(context.Background())
	run := &Run{
		ID:           cp.RunID,
		Prompt:       cp.Prompt,
		Workspace:    cp.Workspace,
		Status:       status,
		CreatedAt:    cp.CreatedAt,
		UpdatedAt:    cp.UpdatedAt,
		Plan:         cp.Plan,
		Events:       make([]Event, 0),
		Subagents:    cp.Subagents,
		FileChanges:  cp.FileChanges,
		Orchestration: orch,
		approvalCh:   make(chan ApprovalResponse, 1),
		compactCh:    make(chan struct{}, 1),
		messages:     cp.Messages,
		ctx:          ctx,
		cancel:       cancel,
	}
	if run.Subagents == nil {
		run.Subagents = make([]*Subagent, 0)
	}
	m.runs[cp.RunID] = run
	return run
}

// PauseRun requests a graceful pause: the flag is set and the context is
// cancelled so an in-flight model call aborts; the loop then records the
// pause instead of treating it as a hard cancel.
func (m *RunManager) PauseRun(id string) bool {
	run, ok := m.GetRun(id)
	if !ok {
		return false
	}
	run.mu.Lock()
	live := run.Status == RunRunning || run.Status == RunWaiting || run.Status == RunPending
	run.mu.Unlock()
	if !live {
		return false
	}
	run.RequestPause()
	return true
}

// ResumeRun prepares a paused/cancelled/failed run for re-execution.
// Runs missing from memory (e.g. after a backend restart) are rebuilt
// from their on-disk checkpoint. If prompt is non-empty it is appended
// as a new user message so the user can redirect the resumed run.
func (m *RunManager) ResumeRun(id string, prompt string) (*Run, error) {
	run, ok := m.GetRun(id)
	if !ok {
		return nil, fmt.Errorf("run '%s' not found", id)
	}

	run.mu.Lock()
	switch run.Status {
	case RunRunning, RunWaiting, RunPending:
		run.mu.Unlock()
		return nil, fmt.Errorf("run is still active")
	}
	msgs := make([]llm.Message, len(run.messages))
	copy(msgs, run.messages)
	run.mu.Unlock()

	if len(msgs) == 0 {
		cp, err := m.LoadCheckpoint(run.ID, run.Workspace)
		if err == nil && len(cp.Messages) > 0 {
			msgs = cp.Messages
		}
	}
	if len(msgs) == 0 {
		return nil, fmt.Errorf("run has no checkpointed context to resume from")
	}

	if prompt != "" {
		msgs = append(msgs, llm.Message{Role: llm.RoleUser, Content: prompt})
	}

	run.ResetContext()
	run.SetMessages(msgs)
	run.SetStatus(RunPending)
	return run, nil
}
