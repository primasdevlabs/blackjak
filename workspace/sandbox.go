package workspace

import (
	"fmt"
)

// Sandbox isolates execution environments for security.
type Sandbox struct {
	ws *Workspace
}

// NewSandbox creates a Sandbox bound to a workspace.
func NewSandbox(ws *Workspace) *Sandbox {
	return &Sandbox{
		ws: ws,
	}
}

// ValidateOperation checks whether a file or command operation is allowed within the sandbox.
func (s *Sandbox) ValidateOperation(operation string, pathOrCommand string) error {
	if s.ws != nil && operation == "file" {
		if !s.ws.IsPathWithinWorkspace(pathOrCommand) {
			return fmt.Errorf("sandbox restriction: path '%s' violates workspace isolation", pathOrCommand)
		}
	}
	return nil
}
