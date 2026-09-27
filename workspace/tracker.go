package workspace

import (
	"fmt"
	"sync"
	"time"
)

// ChangeType defines the type of workspace file modification.
type ChangeType string

const (
	ChangeCreated  ChangeType = "created"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
	ChangeRenamed  ChangeType = "renamed"
	ChangeMoved    ChangeType = "moved"
)

// FileChange tracks modifications made to a file by a specific agent run.
type FileChange struct {
	ID           string     `json:"id"`
	RunID        string     `json:"runId"`
	AgentID      string     `json:"agentId"`
	Type         ChangeType `json:"type"`
	Path         string     `json:"path"`
	PreviousPath string     `json:"previousPath,omitempty"`
	Diff         string     `json:"diff,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
}

// FileLease represents a temporary exclusive edit lock on a file path.
type FileLease struct {
	Path       string    `json:"path"`
	AgentID    string    `json:"agentId"`
	AcquiredAt time.Time `json:"acquiredAt"`
}

// FileTrackerManager tracks workspace file changes and manages agent file leases.
type FileTrackerManager struct {
	mu      sync.RWMutex
	changes map[string][]FileChange // runID -> []FileChange
	leases  map[string]FileLease    // path -> FileLease
}

// NewFileTrackerManager initializes a new FileTrackerManager.
func NewFileTrackerManager() *FileTrackerManager {
	return &FileTrackerManager{
		changes: make(map[string][]FileChange),
		leases:  make(map[string]FileLease),
	}
}

// TrackChange registers a file change event with attribution.
func (tm *FileTrackerManager) TrackChange(runID, agentID string, cType ChangeType, path, prevPath, diff string) FileChange {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	fc := FileChange{
		ID:           fmt.Sprintf("fc_%d", time.Now().UnixNano()),
		RunID:        runID,
		AgentID:      agentID,
		Type:         cType,
		Path:         path,
		PreviousPath: prevPath,
		Diff:         diff,
		Timestamp:    time.Now(),
	}

	tm.changes[runID] = append(tm.changes[runID], fc)
	return fc
}

// GetChanges returns all tracked file changes for a given runID.
func (tm *FileTrackerManager) GetChanges(runID string) []FileChange {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	list, ok := tm.changes[runID]
	if !ok {
		return nil
	}
	res := make([]FileChange, len(list))
	copy(res, list)
	return res
}

// AcquireLease attempts to grant an exclusive edit lease to an agent.
func (tm *FileTrackerManager) AcquireLease(path, agentID string) (bool, string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	existing, exists := tm.leases[path]
	if exists {
		if existing.AgentID == agentID {
			return true, agentID
		}
		// Conflict: file is locked by another subagent
		return false, existing.AgentID
	}

	tm.leases[path] = FileLease{
		Path:       path,
		AgentID:    agentID,
		AcquiredAt: time.Now(),
	}

	return true, agentID
}

// ReleaseLease unlocks a file path held by an agent.
func (tm *FileTrackerManager) ReleaseLease(path, agentID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if existing, exists := tm.leases[path]; exists && existing.AgentID == agentID {
		delete(tm.leases, path)
	}
}

// ForceReleaseLease enables the orchestrator to override lease ownership.
func (tm *FileTrackerManager) ForceReleaseLease(path string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.leases, path)
}
