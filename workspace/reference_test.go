package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspace_ParseWorkspaceReferences(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ws_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	authDir := filepath.Join(tempDir, "internal", "auth")
	_ = os.MkdirAll(authDir, 0755)
	_ = os.WriteFile(filepath.Join(authDir, "middleware.go"), []byte("package auth"), 0644)

	ws := New(tempDir)

	prompt := "Fix the bug in @internal/auth/middleware.go:84 and inspect @internal/auth"
	refs, _ := ws.ParseWorkspaceReferences(prompt)

	if len(refs) != 2 {
		t.Fatalf("Expected 2 references, got %d", len(refs))
	}

	if refs[0].Path != "internal/auth/middleware.go" || refs[0].Line != 84 {
		t.Errorf("Unexpected first ref: %+v", refs[0])
	}

	if refs[1].Path != "internal/auth" || refs[1].Type != RefTypeFolder {
		t.Errorf("Unexpected second ref: %+v", refs[1])
	}
}
