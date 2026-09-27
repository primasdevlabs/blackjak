package tools

import (
	"fmt"
	"path"
	"strings"
)

// PolicyMode controls how much autonomy the agent is allowed.
type PolicyMode string

const (
	// PolicySupervised allows reads, writes, and commands but prompts before
	// risky operations (destructive commands, git commit).
	PolicySupervised PolicyMode = "supervised"
	// PolicyReadOnly permits inspection only — no file writes, no git
	// mutations, and only a safe allowlist of shell commands.
	PolicyReadOnly PolicyMode = "readonly"
	// PolicyAutonomous never prompts for approval. Deny-listed commands and
	// protected paths still apply — autonomy does not override hard blocks.
	PolicyAutonomous PolicyMode = "autonomous"
)

// Policy enforces runtime guardrails inside tools. It is the last line of
// defense — prompts can ask the model nicely, the policy is what actually
// stops a bad call.
type Policy struct {
	Mode             PolicyMode
	ShellAllowed     bool     // master switch for the shell tool
	ApproveAllShell  bool     // prompt for every command, not just destructive ones
	RequireApproval  bool     // prompt for destructive-pattern commands (supervised mode)
	CommandsDeny     []string // substrings — always blocked, even in autonomous mode
	ProtectedPaths   []string // workspace-relative globs that can never be written
	SubagentsAllowed bool     // permit `delegate` calls
	MaxSubagents     int      // total subagents per run (0 = unlimited)
	MaxSteps         int      // tool-loop iteration cap (0 = use default)
}

// DefaultPolicy returns the supervised baseline.
func DefaultPolicy() Policy {
	return Policy{
		Mode:            PolicySupervised,
		ShellAllowed:    true,
		RequireApproval: true,
		CommandsDeny: []string{
			"rm -rf /", "rm -rf ~", "rm -rf c:\\", "mkfs", "dd if=/dev/",
			":(){", "shutdown", "reboot", "format c:", "format d:",
			"rd /s /q c:\\", "del /s /q c:\\", "| sh", "| bash",
		},
		ProtectedPaths: []string{
			".git", ".env", ".env.*", "*.pem", "*.key", "id_rsa*",
			"*.pfx", "*.p12", ".blackjak",
		},
		SubagentsAllowed: true,
		MaxSubagents:     4,
		MaxSteps:         60,
	}
}

// readOnlyCmdPrefixes are commands considered side-effect-free enough for
// read-only mode. Build/test commands are included — their outputs land in
// caches/dist dirs, not in tracked sources.
var readOnlyCmdPrefixes = []string{
	"ls", "dir", "cat", "type", "pwd", "cd", "echo", "head", "tail", "wc",
	"rg", "grep", "find", "findstr", "where", "which", "tree",
	"git status", "git diff", "git log", "git show", "git branch",
	"git remote", "git blame", "git ls-files", "git rev-parse",
	"go build", "go test", "go vet", "go list", "go version", "go env",
	"go doc", "gofmt -l", "go run",
	"node --version", "npm --version", "npm list", "npm outdated",
	"npm test", "npm run test", "npm run lint", "npm run typecheck", "npm run build",
	"npx tsc --noEmit", "npx vitest", "npx jest", "tsc --noEmit",
	"python --version", "python3 --version", "python -m compileall", "python3 -m compileall",
	"python -m pytest", "python3 -m pytest", "pytest",
	"cargo check", "cargo build", "cargo test", "rustc --version",
	"javac -version", "mvn -version", "mvn test", "gradle -v",
	"make -n", "make test",
	"dotnet test", "dotnet build",
	"whoami", "hostname", "date", "systeminfo", "uname",
	"netstat", "tasklist", "ipconfig",
}

// isReadOnlyCommand reports whether a command only observes state.
func isReadOnlyCommand(cmd string) bool {
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, prefix := range readOnlyCmdPrefixes {
		if lower == prefix || strings.HasPrefix(lower, prefix+" ") || strings.HasPrefix(lower, prefix+"\t") {
			return true
		}
	}
	return false
}

// CheckCommand validates a shell command against the policy. A non-nil error
// is a hard block — no approval flow can override it.
func (p Policy) CheckCommand(cmd string) error {
	if !p.ShellAllowed {
		return fmt.Errorf("shell commands are disabled by guardrails")
	}

	lower := strings.ToLower(cmd)
	for _, pattern := range p.CommandsDeny {
		deny := strings.ToLower(strings.TrimSpace(pattern))
		if deny != "" && strings.Contains(lower, deny) {
			return fmt.Errorf("command blocked by guardrail rule %q", pattern)
		}
	}

	if p.Mode == PolicyReadOnly && !isReadOnlyCommand(cmd) {
		return fmt.Errorf("read-only mode: command not on the safe allowlist — %q", cmd)
	}
	return nil
}

// CommandNeedsApproval reports whether a command must prompt the user first.
// Called after CheckCommand — anything already blocked never reaches this.
func (p Policy) CommandNeedsApproval(cmd string) bool {
	if p.Mode != PolicySupervised {
		return false
	}
	if p.ApproveAllShell {
		return true
	}
	return p.RequireApproval && destructivePattern.MatchString(cmd)
}

// GitNeedsApproval reports whether a mutating git op must prompt.
func (p Policy) GitNeedsApproval(op string) bool {
	return p.Mode == PolicySupervised && p.RequireApproval && op == "commit"
}

// CheckGitOp validates a git operation. Read-only mode permits only
// inspection operations; commit stays approval-gated in supervised mode.
func (p Policy) CheckGitOp(op string) error {
	if p.Mode == PolicyReadOnly {
		switch op {
		case "status", "diff", "log", "show":
			return nil
		case "branch":
			return nil // bare `branch` lists; creation args are checked by the caller
		default:
			return fmt.Errorf("read-only mode: git %s is not permitted", op)
		}
	}
	return nil
}

// CheckWrite validates a write/edit against the policy.
func (p Policy) CheckWrite(relPath string) error {
	if p.Mode == PolicyReadOnly {
		return fmt.Errorf("file writes are disabled by read-only mode")
	}
	rel := strings.TrimPrefix(filepathToSlash(relPath), "./")
	for _, pattern := range p.ProtectedPaths {
		if matchProtectedPath(pattern, rel) {
			return fmt.Errorf("%q is a protected path", rel)
		}
	}
	return nil
}

// matchProtectedPath matches a protected pattern against a workspace-relative
// slash path. Bare names (".git", ".env") match the path itself or anything
// beneath it; glob patterns ("*.pem", ".env.*") match via path.Match.
func matchProtectedPath(pattern, rel string) bool {
	p := strings.Trim(strings.TrimSpace(pattern), "/")
	if p == "" {
		return false
	}
	if !strings.ContainsAny(p, "*?[") {
		return rel == p || strings.HasPrefix(rel, p+"/") || strings.HasPrefix(rel, p+".")
	}
	if ok, _ := path.Match(p, rel); ok {
		return true
	}
	// Patterns like ".env.*" should also match nested ".env.local" files.
	if ok, _ := path.Match(p, path.Base(rel)); ok && !strings.Contains(p, "/") {
		return true
	}
	return false
}

// CheckDelegate validates a subagent spawn against the policy.
func (p Policy) CheckDelegate(currentCount int) error {
	if !p.SubagentsAllowed {
		return fmt.Errorf("subagent delegation is disabled by guardrails")
	}
	if p.MaxSubagents > 0 && currentCount >= p.MaxSubagents {
		return fmt.Errorf("subagent limit reached (%d)", p.MaxSubagents)
	}
	return nil
}

// StepsLimit resolves the effective loop cap.
func (p Policy) StepsLimit() int {
	if p.MaxSteps > 0 {
		return p.MaxSteps
	}
	return maxStepsDefault
}

const maxStepsDefault = 60

var filepathToSlash = strings.NewReplacer(`\`, "/").Replace
