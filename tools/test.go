package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"blackjak/workspace"
)

// TestTool detects and runs the project test suite.
type TestTool struct {
	ws     *workspace.Workspace
	policy Policy
}

func (t *TestTool) Name() string { return "test" }

func (t *TestTool) Description() string {
	return "Run the project test suite. Auto-detects go test, npm test, pytest, or cargo test when no command is given."
}

func (t *TestTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Explicit test command (optional — auto-detected otherwise)",
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default 180)",
			},
		},
	}
}

func (t *TestTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	cmd := argString(args, "command")
	if cmd == "" {
		cmd = t.detect()
		if cmd == "" {
			return nil, fmt.Errorf("could not detect test command; specify one explicitly")
		}
	}
	timeout := argInt(args, "timeout_seconds", 180)

	shell := &ShellTool{ws: t.ws, policy: t.policy}
	res, err := shell.Execute(ctx, map[string]interface{}{
		"command":         cmd,
		"timeout_seconds": timeout,
	})
	if err != nil {
		return nil, err
	}
	if m, ok := res.(map[string]interface{}); ok {
		m["command"] = cmd
		m["passed"] = m["exitCode"] == 0
	}
	return res, nil
}

func (t *TestTool) detect() string {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(t.ws.RootPath, name))
		return err == nil
	}
	switch {
	case exists("go.mod"):
		return "go test ./..."
	case exists("package.json"):
		return "npm test -- --watch=false"
	case exists("pytest.ini"), exists("pyproject.toml"), exists("setup.cfg"):
		return "python -m pytest"
	case exists("Cargo.toml"):
		return "cargo test"
	case exists("Makefile"):
		return "make test"
	}
	return ""
}
