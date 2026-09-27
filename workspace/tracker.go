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
	Timeout    time.Duration `json:"-"` // Lease expires after this duration (0 = no timeout)
}

// DefaultLeaseTimeout is the default expiration time for file leases.
const DefaultLeaseTimeout = 30 * time.Second

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
// Expired leases are automatically released before checking for conflicts.
func (tm *FileTrackerManager) AcquireLease(path, agentID string) (bool, string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	existing, exists := tm.leases[path]
	if exists {
		if existing.AgentID == agentID {
			return true, agentID
		}
		// Check if the existing lease has expired
		if existing.Timeout > 0 && time.Since(existing.AcquiredAt) > existing.Timeout {
			// Lease expired — allow takeover
			delete(tm.leases, path)
		} else {
			// Conflict: file is locked by another subagent
			return false, existing.AgentID
		}
	}

	tm.leases[path] = FileLease{
		Path:       path,
		AgentID:    agentID,
		AcquiredAt: time.Now(),
		Timeout:    DefaultLeaseTimeout,
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

// CleanupAgentLeases releases all leases held by a specific agent.
// Called when a subagent fails, is cancelled, or crashes to prevent lease deadlocks.
func (tm *FileTrackerManager) CleanupAgentLeases(agentID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for path, lease := range tm.leases {
		if lease.AgentID == agentID {
			delete(tm.leases, path)
		}
	}
}

// CleanupExpiredLeases removes all leases that have exceeded their timeout.
// Can be called periodically by a background goroutine.
func (tm *FileTrackerManager) CleanupExpiredLeases() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	cleaned := 0
	for path, lease := range tm.leases {
		if lease.Timeout > 0 && time.Since(lease.AcquiredAt) > lease.Timeout {
			delete(tm.leases, path)
			cleaned++
		}
	}
	return cleaned
}

// GetAllLeases returns a snapshot of all active leases (for diagnostics/UI).
func (tm *FileTrackerManager) GetAllLeases() []FileLease {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]FileLease, 0, len(tm.leases))
	for _, lease := range tm.leases {
		result = append(result, lease)
	}
	return result
}
