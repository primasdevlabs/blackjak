package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Sandbox isolates execution environments for security.
// When scope paths are set, file operations are restricted to those directories.
type Sandbox struct {
	ws    *Workspace
	scope []string // Relative scope paths within the workspace (empty = full workspace)
}

// NewSandbox creates a Sandbox bound to a workspace with full workspace access.
func NewSandbox(ws *Workspace) *Sandbox {
	return &Sandbox{
		ws: ws,
	}
}

// NewScopedSandbox creates a Sandbox restricted to specific paths within the workspace.
// If scope is empty or nil, falls back to full workspace access.
func NewScopedSandbox(ws *Workspace, scope []string) *Sandbox {
	return &Sandbox{
		ws:    ws,
		scope: scope,
	}
}

// ValidateOperation checks whether a file or command operation is allowed within the sandbox.
// For file operations, validates against workspace boundary and optional scope restrictions.
func (s *Sandbox) ValidateOperation(operation string, pathOrCommand string) error {
	if s.ws == nil {
		return nil
	}

	if operation == "file" {
		// First: workspace boundary check
		if !s.ws.IsPathWithinWorkspace(pathOrCommand) {
			return fmt.Errorf("sandbox restriction: path '%s' violates workspace isolation", pathOrCommand)
		}

		// Second: scope enforcement (if scope is defined)
		if len(s.scope) > 0 {
			if !s.isWithinScope(pathOrCommand) {
				return fmt.Errorf("sandbox restriction: path '%s' is outside subagent scope %v", pathOrCommand, s.scope)
			}
		}
	}

	return nil
}

// isWithinScope checks whether a path falls within any of the allowed scope paths.
func (s *Sandbox) isWithinScope(targetPath string) bool {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	cleanTarget := filepath.Clean(absTarget)

	for _, scopePath := range s.scope {
		var absScopePath string
		if filepath.IsAbs(scopePath) {
			absScopePath = scopePath
		} else {
			absScopePath = filepath.Join(s.ws.RootPath, scopePath)
		}
		cleanScope := filepath.Clean(absScopePath)

		rel, err := filepath.Rel(cleanScope, cleanTarget)
		if err != nil {
			continue
		}

		// Path is within scope if the relative path doesn't start with ".."
		if !strings.HasPrefix(rel, "..") && rel != ".." {
			return true
		}
	}

	return false
}

// GetScope returns the current scope paths, or nil for unrestricted access.
func (s *Sandbox) GetScope() []string {
	return s.scope
}
