package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"blackjak/workspace"
)

// TestTool detects and runs the project test suite.
type TestTool struct {
	ws      *workspace.Workspace
	approve ApprovalFunc
	policy  Policy
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
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Optional package/directory hint for auto-detect (e.g. ./api or extension/webview/react)",
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
		cmd = t.detect(argString(args, "path"))
		if cmd == "" {
			return nil, fmt.Errorf("could not detect test command; specify one explicitly")
		}
	}
	timeout := argInt(args, "timeout_seconds", 180)

	shell := &ShellTool{ws: t.ws, approve: t.approve, policy: t.policy}
	res, err := shell.Execute(ctx, map[string]interface{}{
		"command":         cmd,
		"timeout_seconds": timeout,
	})
	if err != nil {
		return nil, err
	}
	if m, ok := res.(map[string]interface{}); ok {
		m["command"] = cmd
		passed := false
		switch v := m["exitCode"].(type) {
		case int:
			passed = v == 0
		case float64:
			passed = v == 0
		case int64:
			passed = v == 0
		}
		m["passed"] = passed
	}
	return res, nil
}

func (t *TestTool) detect(hint string) string {
	root := t.ws.RootPath
	if hint != "" {
		if abs, err := t.ws.ResolvePath(hint); err == nil {
			root = abs
		}
	}
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	relHint := strings.TrimSpace(hint)
	switch {
	case exists("go.mod"):
		if relHint != "" && relHint != "." && relHint != "./" {
			return "go test " + filepath.ToSlash(relHint)
		}
		return "go test ./..."
	case exists("package.json"):
		if relHint != "" && relHint != "." && relHint != "./" {
			return "npm test --prefix " + filepath.ToSlash(relHint) + " -- --watch=false"
		}
		return "npm test -- --watch=false"
	case exists("pytest.ini"), exists("pyproject.toml"), exists("setup.cfg"):
		return "python -m pytest"
	case exists("Cargo.toml"):
		return "cargo test"
	case exists("Makefile"):
		return "make test"
	}
	// Polyglot fallback: prefer Go at workspace root even if hint missed.
	if _, err := os.Stat(filepath.Join(t.ws.RootPath, "go.mod")); err == nil {
		return "go test ./..."
	}
	return ""
}
