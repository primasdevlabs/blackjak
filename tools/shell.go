package tools

import (
	"context"
)

// ShellTool enables executing terminal commands.
type ShellTool struct{}

// NewShellTool initializes a new ShellTool.
func NewShellTool() *ShellTool {
	return &ShellTool{}
}

func (s *ShellTool) Name() string { return "shell" }

func (s *ShellTool) Description() string {
	return "Execute shell commands inside the workspace."
}

func (s *ShellTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return nil, nil
}
