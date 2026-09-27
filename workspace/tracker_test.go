package workspace

import (
	"testing"
)

func TestFileTrackerManager_LeaseAndConflict(t *testing.T) {
	tm := NewFileTrackerManager()

	path := "src/auth.ts"
	acquired, owner := tm.AcquireLease(path, "coder_1")
	if !acquired || owner != "coder_1" {
		t.Fatalf("Expected coder_1 to acquire lease, got acquired=%v owner=%s", acquired, owner)
	}

	// Conflict check: second agent tries to acquire same file
	acquired2, owner2 := tm.AcquireLease(path, "coder_2")
	if acquired2 || owner2 != "coder_1" {
		t.Errorf("Expected conflict with owner coder_1, got acquired2=%v owner2=%s", acquired2, owner2)
	}

	// Release lease
	tm.ReleaseLease(path, "coder_1")

	// Now coder_2 can acquire
	acquired3, owner3 := tm.AcquireLease(path, "coder_2")
	if !acquired3 || owner3 != "coder_2" {
		t.Errorf("Expected coder_2 to acquire lease after release, got acquired3=%v owner3=%s", acquired3, owner3)
	}
}

func TestFileTrackerManager_TrackChange(t *testing.T) {
	tm := NewFileTrackerManager()

	fc := tm.TrackChange("run_1", "coder_1", ChangeModified, "src/auth.ts", "", "+ diff")
	if fc.RunID != "run_1" || fc.AgentID != "coder_1" {
		t.Errorf("Unexpected FileChange payload: %+v", fc)
	}

	changes := tm.GetChanges("run_1")
	if len(changes) != 1 || changes[0].Path != "src/auth.ts" {
		t.Errorf("Unexpected GetChanges output: %+v", changes)
	}
}
