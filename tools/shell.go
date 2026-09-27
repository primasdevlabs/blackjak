package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"blackjak/workspace"
)

const (
	defaultTimeout = 60 * time.Second
	maxOutput      = 32 * 1024
)

// ShellTool executes terminal commands inside the workspace.
type ShellTool struct {
	ws      *workspace.Workspace
	approve ApprovalFunc
	policy  Policy
}

func (s *ShellTool) Name() string { return "shell" }

func (s *ShellTool) Description() string {
	return "Run a shell command in the workspace. Returns stdout, stderr, and exit code. Destructive commands require user approval."
}

func (s *ShellTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Shell command to execute",
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default 60, max 600)",
			},
		},
		"required": []string{"command"},
	}
}

// destructivePattern matches commands that delete data, kill processes, or
// mutate system/git state irreversibly.
var destructivePattern = regexp.MustCompile(`(?i)(\brm\s+-[a-z]*r|\brm\s+-[a-z]*f|\bdel\s+/[fs]|\brmdir\s+/s|git\s+(push\s+.*--force|reset\s+--hard|clean\s+-[a-z]*f)|\bmkfs\b|\bdd\s+if=|\bformat\b|>\s*/dev/|shutdown|reboot|kill\s+-9)`)

func (s *ShellTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	cmd := argString(args, "command")
	if cmd == "" {
		return nil, fmt.Errorf("command is required")
	}
	timeout := time.Duration(argInt(args, "timeout_seconds", 60)) * time.Second
	if timeout <= 0 || timeout > 10*time.Minute {
		timeout = defaultTimeout
	}

	// Hard blocks first — no approval flow can override these.
	if err := s.policy.CheckCommand(cmd); err != nil {
		return nil, err
	}

	if s.policy.CommandNeedsApproval(cmd) {
		if s.approve == nil {
			return nil, fmt.Errorf("command requires approval but no approver is configured")
		}
		ok, err := s.approve(ctx, "command", fmt.Sprintf("Execute command: %s", cmd))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("command denied by user")
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(execCtx, "cmd", "/C", cmd)
	} else {
		c = exec.CommandContext(execCtx, "sh", "-c", cmd)
	}
	c.Dir = s.ws.RootPath

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	runErr := c.Run()
	exitCode := 0
	if c.ProcessState != nil {
		exitCode = c.ProcessState.ExitCode()
	}
	if execCtx.Err() == context.DeadlineExceeded {
		return map[string]interface{}{
			"stdout":   truncateOutput(stdout.String()),
			"stderr":   truncateOutput(stderr.String()),
			"exitCode": -1,
			"error":    fmt.Sprintf("timed out after %s", timeout),
		}, nil
	}
	result := map[string]interface{}{
		"stdout":   truncateOutput(stdout.String()),
		"stderr":   truncateOutput(stderr.String()),
		"exitCode": exitCode,
	}
	if runErr != nil && exitCode == 0 {
		return nil, fmt.Errorf("command failed to start: %w", runErr)
	}
	if runErr != nil {
		result["error"] = runErr.Error()
	}
	return result, nil
}

func truncateOutput(s string) string {
	if len(s) > maxOutput {
		return s[:maxOutput] + fmt.Sprintf("\n… [%d bytes truncated]", len(s)-maxOutput)
	}
	return s
}
