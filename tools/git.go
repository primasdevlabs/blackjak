package tools

import (
	"context"
	"fmt"
	"strings"

	"blackjak/workspace"
)

// GitTool provides safe git operations inside the workspace.
type GitTool struct {
	ws      *workspace.Workspace
	approve ApprovalFunc
	policy  Policy
}

func (g *GitTool) Name() string { return "git" }

func (g *GitTool) Description() string {
	return "Run git operations: status, diff, log, add, commit, branch, show. Write operations that rewrite history or push are not permitted."
}

func (g *GitTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"status", "diff", "log", "add", "commit", "branch", "show"},
				"description": "Git operation",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Extra arguments (e.g. file paths for diff/add, message via -m for commit)",
			},
		},
		"required": []string{"operation"},
	}
}

func (g *GitTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	op := argString(args, "operation")
	extra := argString(args, "args")

	if err := g.policy.CheckGitOp(op); err != nil {
		return nil, err
	}
	// `branch` with args creates/mutates — read-only allows bare listing only.
	if g.policy.Mode == PolicyReadOnly && op == "branch" && extra != "" {
		return nil, fmt.Errorf("read-only mode: git branch with arguments is not permitted")
	}

	var gitArgs []string
	switch op {
	case "status":
		gitArgs = []string{"status", "--short", "--branch"}
	case "diff":
		gitArgs = []string{"diff", "--no-color"}
		if extra != "" {
			gitArgs = append(gitArgs, strings.Fields(extra)...)
		}
	case "log":
		gitArgs = []string{"log", "--oneline", "-20"}
		if extra != "" {
			gitArgs = append(gitArgs, strings.Fields(extra)...)
		}
	case "show":
		gitArgs = []string{"show", "--stat", "--no-color"}
		if extra != "" {
			gitArgs = append(gitArgs, strings.Fields(extra)...)
		}
	case "add":
		if extra == "" {
			return nil, fmt.Errorf("add requires file args")
		}
		gitArgs = append([]string{"add"}, strings.Fields(extra)...)
	case "commit":
		if extra == "" {
			return nil, fmt.Errorf("commit requires args (e.g. -m \"message\")")
		}
		gitArgs = append([]string{"commit"}, strings.Fields(extra)...)
		if g.policy.GitNeedsApproval(op) {
			if g.approve == nil {
				return nil, fmt.Errorf("git commit requires approval but no approver is configured")
			}
			ok, err := g.approve(ctx, "git_commit", fmt.Sprintf("git commit %s", extra))
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("commit denied by user")
			}
		}
	case "branch":
		gitArgs = append([]string{"branch"}, strings.Fields(extra)...)
	default:
		return nil, fmt.Errorf("unsupported git operation: %s", op)
	}

	shell := &ShellTool{ws: g.ws, approve: g.approve, policy: g.policy}
	return shell.Execute(ctx, map[string]interface{}{
		"command":         "git " + strings.Join(gitArgs, " "),
		"timeout_seconds": 30,
	})
}
