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

func TestScopedSandbox_ScopeValidation(t *testing.T) {
	wsRoot, _ := filepath.Abs("./testdata")
	ws := New(wsRoot)

	scopes := []string{
		filepath.Join(wsRoot, "src"),
		filepath.Join(wsRoot, "pkg"),
	}

	scopedSb := NewScopedSandbox(ws, scopes)

	// Path inside designated scope -> allowed
	allowedPath := filepath.Join(wsRoot, "src", "components", "App.tsx")
	if err := scopedSb.ValidateOperation("file", allowedPath); err != nil {
		t.Errorf("Expected path inside scope to be allowed, got error: %v", err)
	}

	// Path outside designated scope (even if inside workspace) -> rejected
	outOfScopePath := filepath.Join(wsRoot, "docs", "README.md")
	if err := scopedSb.ValidateOperation("file", outOfScopePath); err == nil {
		t.Errorf("Expected path outside subagent scope to be rejected")
	}

	// Path outside entire workspace -> rejected
	traversalPath := filepath.Join(wsRoot, "..", "etc", "hosts")
	if err := scopedSb.ValidateOperation("file", traversalPath); err == nil {
		t.Errorf("Expected path outside workspace to be rejected")
	}
}
