package workspace

import (
	"path/filepath"
	"testing"
)

func TestWorkspace_IsPathWithinWorkspace(t *testing.T) {
	wsRoot, _ := filepath.Abs("./testdata")
	ws := New(wsRoot)

	insidePath := filepath.Join(wsRoot, "src", "main.go")
	if !ws.IsPathWithinWorkspace(insidePath) {
		t.Errorf("Path inside workspace was rejected: %s", insidePath)
	}

	outsidePath := filepath.Join(wsRoot, "..", "..", "etc", "passwd")
	if ws.IsPathWithinWorkspace(outsidePath) {
		t.Errorf("Path outside workspace was accepted: %s", outsidePath)
	}
}

func TestSandbox_ValidateOperation(t *testing.T) {
	wsRoot, _ := filepath.Abs("./testdata")
	ws := New(wsRoot)
	sb := NewSandbox(ws)

	insidePath := filepath.Join(wsRoot, "config.json")
	if err := sb.ValidateOperation("file", insidePath); err != nil {
		t.Errorf("Unexpected sandbox validation error: %v", err)
	}

	outsidePath := filepath.Join(wsRoot, "..", "secret.txt")
	if err := sb.ValidateOperation("file", outsidePath); err == nil {
		t.Errorf("Expected sandbox validation error for outside path")
	}
}
