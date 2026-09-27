package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"blackjak/workspace"
)

// FileChangeTracker manages live file change tracking across tasks, subagents, and pre-existing workspace states.
type FileChangeTracker struct {
	mu           sync.RWMutex
	broker       *EventBroker
	changeSets   map[string]*workspace.TaskChangeSet // taskID -> TaskChangeSet
	initialState map[string]string                   // path -> hash before task start
}

func NewFileChangeTracker(broker *EventBroker) *FileChangeTracker {
	return &FileChangeTracker{
		broker:       broker,
		changeSets:   make(map[string]*workspace.TaskChangeSet),
		initialState: make(map[string]string),
	}
}

// SnapshotTaskState captures pre-existing file and git checksum states at the start of a task.
func (fct *FileChangeTracker) SnapshotTaskState(taskID string, rootPath string) {
	fct.mu.Lock()
	defer fct.mu.Unlock()

	fct.changeSets[taskID] = &workspace.TaskChangeSet{
		TaskID:             taskID,
		PreExistingChanges: make([]workspace.ExtendedFileChange, 0),
		TaskChanges:        make([]workspace.ExtendedFileChange, 0),
		Files:              make([]workspace.ExtendedFileChange, 0),
	}

	if rootPath == "" {
		return
	}

	_ = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(rootPath, path)
		if isIgnoredPath(rel) {
			return nil
		}
		hash := workspace.CalculateFileHash(path)
		fct.initialState[path] = hash
		return nil
	})
}

// TrackChange registers a file modification, created, deleted, or renamed operation from any agent or subagent.
func (fct *FileChangeTracker) TrackChange(taskID, agentID string, cType workspace.ChangeType, path, prevPath, diff string) workspace.ExtendedFileChange {
	fct.mu.Lock()
	defer fct.mu.Unlock()

	cs, ok := fct.changeSets[taskID]
	if !ok {
		cs = &workspace.TaskChangeSet{
			TaskID:             taskID,
			PreExistingChanges: make([]workspace.ExtendedFileChange, 0),
			TaskChanges:        make([]workspace.ExtendedFileChange, 0),
			Files:              make([]workspace.ExtendedFileChange, 0),
		}
		fct.changeSets[taskID] = cs
	}

	beforeHash := fct.initialState[path]
	afterHash := workspace.CalculateFileHash(path)
	if afterHash == "" && diff != "" {
		afterHash = workspace.CalculateContentHash(diff)
	}

	role := workspace.ClassifyRole(path)
	importance := 3
	if role == workspace.RolePrimary {
		importance = 5
	} else if role == workspace.RoleTest {
		importance = 4
	}

	fcState := workspace.Modified
	if cType == workspace.ChangeCreated {
		fcState = workspace.Created
	} else if cType == workspace.ChangeDeleted {
		fcState = workspace.Deleted
	} else if cType == workspace.ChangeRenamed {
		fcState = workspace.Renamed
	}

	change := workspace.ExtendedFileChange{
		ID:            fmt.Sprintf("ext_fc_%d", time.Now().UnixNano()),
		RunID:         taskID,
		TaskID:        taskID,
		AgentID:       agentID,
		Type:          cType,
		State:         fcState,
		Role:          role,
		Importance:    importance,
		Path:          path,
		PreviousPath:  prevPath,
		BeforeHash:    beforeHash,
		AfterHash:     afterHash,
		Diff:          diff,
		IsPreExisting: false,
		Timestamp:     time.Now(),
	}

	// Update Unified Task ChangeSet
	cs.TaskChanges = append(cs.TaskChanges, change)
	cs.Files = append(cs.Files, change)

	// Broadcast file.changed event
	if fct.broker != nil {
		fct.broker.Publish(Event{
			ID:        change.ID,
			RunID:     taskID,
			AgentID:   agentID,
			Type:      EventType("file.changed"),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"taskId":        taskID,
				"agentId":       agentID,
				"path":          path,
				"previousPath":  prevPath,
				"changeType":    string(cType),
				"state":         string(fcState),
				"role":          string(role),
				"importance":    importance,
				"isPreExisting": false,
			},
		})
	}

	return change
}

// GetTaskChanges returns unified task change set for orchestrator or UI views.
func (fct *FileChangeTracker) GetTaskChanges(taskID string) *workspace.TaskChangeSet {
	fct.mu.RLock()
	defer fct.mu.RUnlock()

	cs, ok := fct.changeSets[taskID]
	if !ok {
		return &workspace.TaskChangeSet{TaskID: taskID}
	}
	return cs
}

func isIgnoredPath(rel string) bool {
	return filepath.HasPrefix(rel, ".git") || filepath.HasPrefix(rel, "node_modules") || filepath.HasPrefix(rel, "dist") || filepath.HasPrefix(rel, "bin")
}
