package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"
)

type FileChangeState string

const (
	Unchanged FileChangeState = "unchanged"
	Created   FileChangeState = "created"
	Modified  FileChangeState = "modified"
	Deleted   FileChangeState = "deleted"
	Renamed   FileChangeState = "renamed"
	Reviewed  FileChangeState = "reviewed"
)

type FileRole string

const (
	RolePrimary FileRole = "primary"
	RoleRelated FileRole = "related"
	RoleTest    FileRole = "test"
)

type FileVersion struct {
	AgentHash      string `json:"agentHash"`
	FilesystemHash string `json:"filesystemHash"`
	EditorHash     string `json:"editorHash,omitempty"`
}

type ExtendedFileChange struct {
	ID             string          `json:"id"`
	RunID          string          `json:"runId"`
	TaskID         string          `json:"taskId"`
	AgentID        string          `json:"agentId"`
	Type           ChangeType      `json:"type"`
	State          FileChangeState `json:"state"`
	Role           FileRole        `json:"role"`
	Importance     int             `json:"importance"` // 1 (low) to 5 (high/primary)
	Path           string          `json:"path"`
	PreviousPath   string          `json:"previousPath,omitempty"`
	BeforeHash     string          `json:"beforeHash,omitempty"`
	AfterHash      string          `json:"afterHash,omitempty"`
	Diff           string          `json:"diff,omitempty"`
	IsPreExisting  bool            `json:"isPreExisting"`
	Timestamp      time.Time       `json:"timestamp"`
}

type TaskChangeSet struct {
	TaskID             string               `json:"taskId"`
	PreExistingChanges []ExtendedFileChange `json:"preExistingChanges"`
	TaskChanges        []ExtendedFileChange `json:"taskChanges"`
	Files              []ExtendedFileChange `json:"files"`
}

// CalculateFileHash computes sha256 checksum of file contents.
func CalculateFileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CalculateContentHash computes sha256 checksum of raw string or bytes.
func CalculateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// ClassifyRole infers file role based on path conventions.
func ClassifyRole(path string) FileRole {
	if containsSubstring(path, []string{"test", "spec", "_test.go", ".test.ts", ".spec.ts", ".test.js"}) {
		return RoleTest
	}
	if containsSubstring(path, []string{"config", "setting", "env", "pkg", "utils"}) {
		return RoleRelated
	}
	return RolePrimary
}

func containsSubstring(s string, subs []string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
