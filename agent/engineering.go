package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Phase is a step in Blackjak's engineering loop.
type Phase string

const (
	PhaseUnderstand Phase = "understand"
	PhaseInspect    Phase = "inspect"
	PhaseReason     Phase = "reason"
	PhasePlan       Phase = "plan"
	PhaseImplement  Phase = "implement"
	PhaseVerify     Phase = "verify"
	PhaseReview     Phase = "review"
	PhaseRefine     Phase = "refine"
	PhaseDone       Phase = "done"
)

// TaskClass distinguishes trivial tasks (gates skipped) from non-trivial ones.
type TaskClass string

const (
	TaskTrivial    TaskClass = "trivial"
	TaskNonTrivial TaskClass = "nonTrivial"
)

// Evidence tracks what the agent has proven during a run.
type Evidence struct {
	Inspect bool `json:"inspect"`
	Plan    bool `json:"plan"`
	Implement bool `json:"implement"`
	Verify  bool `json:"verify"`
	Review  bool `json:"review"`
}

// ReviewResult is the structured self-review verdict.
type ReviewResult struct {
	SolvesProblem         bool     `json:"solvesProblem"`
	UnnecessaryComplexity bool     `json:"unnecessaryComplexity"`
	SecurityIssues        []string `json:"securityIssues,omitempty"`
	Races                 []string `json:"races,omitempty"`
	Leaks                 []string `json:"leaks,omitempty"`
	ContractBreaks        []string `json:"contractBreaks,omitempty"`
	AISlopRisk            bool     `json:"aiSlopRisk"`
	Findings              []string `json:"findings,omitempty"`
	Verdict               string   `json:"verdict"` // pass | revise
}

// EngineeringState tracks phases, evidence, and gates for a run.
type EngineeringState struct {
	mu            sync.Mutex
	Phase         Phase      `json:"phase"`
	TaskClass     TaskClass  `json:"taskClass"`
	Evidence      Evidence   `json:"evidence"`
	Review        *ReviewResult `json:"review,omitempty"`
	ReviewRetries int        `json:"reviewRetries"`
	MaxReviewRetries int     `json:"maxReviewRetries"`
	FilesChanged  bool       `json:"filesChanged"`
	LastReviewPrompt string  `json:"-"`
}

// NewEngineeringState classifies the prompt and starts in understand.
func NewEngineeringState(prompt string) *EngineeringState {
	return &EngineeringState{
		Phase:            PhaseUnderstand,
		TaskClass:        ClassifyTask(prompt),
		MaxReviewRetries: 2,
	}
}

// nonTrivialKeywords push a prompt into the gated engineering loop.
// Keep this list focused on sensitive / high-risk work — broad verbs like
// "fix"/"implement" alone must not force explore→verify→self-review theater.
var nonTrivialKeywords = []string{
	"auth", "authentication", "authorization", "oauth", "jwt",
	"migrat", "schema", "database", "sql", "transaction",
	"concurren", "goroutine", "race", "deadlock", "mutex",
	"microservice", "kubernetes", "deploy", "infra",
	"refactor", "architect", "graphql",
	"security", "encrypt", "password", "secret", "csrf", "xss", "ssrf",
	"add feature", "rewrite the", "from scratch",
}

// ClassifyTask returns trivial for short Q&A / local edits / demos, else nonTrivial.
func ClassifyTask(prompt string) TaskClass {
	p := strings.TrimSpace(prompt)
	if p == "" {
		return TaskTrivial
	}
	lower := strings.ToLower(p)
	words := strings.Fields(lower)
	// Explicit demo / scaffold / throwaway work always stays on the fast path,
	// even if the prompt mentions login/auth wording casually.
	if isDemoOrScaffoldPrompt(lower) {
		return TaskTrivial
	}
	if containsNonTrivialKeyword(lower) {
		return TaskNonTrivial
	}
	if len(words) <= 12 {
		// Short questions / lookups / one-liner tweaks stay ungated.
		return TaskTrivial
	}
	if len(words) > 40 {
		return TaskNonTrivial
	}
	// Substantial multi-step change requests.
	heavy := 0
	for _, verb := range []string{"implement ", "refactor ", "migrate ", "redesign ", "architect "} {
		if strings.Contains(lower, verb) {
			heavy++
		}
	}
	if heavy > 0 && len(words) > 18 {
		return TaskNonTrivial
	}
	return TaskTrivial
}

// isDemoOrScaffoldPrompt detects throwaway / test / simple UI scaffolds where
// explore→ask→plan theater is pure waste.
func isDemoOrScaffoldPrompt(lower string) bool {
	markers := []string{
		"test page", "demo page", "sample page", "throwaway",
		"just a test", "just a demo", "just a simple", "just a page",
		"not a full", "not production", "scaffold", "placeholder page",
		"simple login page", "simple html", "static page", "mockup",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// IsFastPath reports whether this run should use the slim toolset + short prompt.
func (e *EngineeringState) IsFastPath() bool {
	if e == nil {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.TaskClass == TaskTrivial
}

func containsNonTrivialKeyword(lower string) bool {
	for _, kw := range nonTrivialKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// Snapshot returns a copy safe for events/JSON.
func (e *EngineeringState) Snapshot() map[string]interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := map[string]interface{}{
		"phase":     string(e.Phase),
		"taskClass": string(e.TaskClass),
		"evidence": map[string]bool{
			"inspect":   e.Evidence.Inspect,
			"plan":      e.Evidence.Plan,
			"implement": e.Evidence.Implement,
			"verify":    e.Evidence.Verify,
			"review":    e.Evidence.Review,
		},
		"filesChanged":  e.FilesChanged,
		"reviewRetries": e.ReviewRetries,
	}
	if e.Review != nil {
		out["review"] = e.Review
	}
	return out
}

// PromoteToNonTrivial upgrades a trivial task when it starts mutating.
func (e *EngineeringState) PromoteToNonTrivial() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.TaskClass == TaskTrivial {
		e.TaskClass = TaskNonTrivial
		if e.Phase == PhaseUnderstand || e.Phase == PhaseDone {
			e.Phase = PhaseInspect
		}
	}
}

// SetPhase updates the current phase.
func (e *EngineeringState) SetPhase(p Phase) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Phase = p
}

// MarkInspect records inspect evidence.
func (e *EngineeringState) MarkInspect() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Evidence.Inspect = true
	if e.Phase == PhaseUnderstand || e.Phase == PhaseInspect {
		e.Phase = PhaseReason
	}
}

// MarkPlan records plan evidence.
func (e *EngineeringState) MarkPlan() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Evidence.Plan = true
	if e.Phase == PhaseUnderstand || e.Phase == PhaseInspect || e.Phase == PhaseReason {
		e.Phase = PhasePlan
	}
}

// MarkImplement records a mutating change.
func (e *EngineeringState) MarkImplement() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Evidence.Implement = true
	e.FilesChanged = true
	// Do not auto-promote trivial local edits into the full gated loop —
	// that forced inspect/verify/self-review theater on every small change.
	e.Phase = PhaseImplement
	// New changes invalidate prior verify/review for gated tasks.
	e.Evidence.Verify = false
	e.Evidence.Review = false
	e.Review = nil
}

// MarkVerify records successful verification.
func (e *EngineeringState) MarkVerify() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Evidence.Verify = true
	e.Phase = PhaseVerify
}

// MarkReview records a review result.
func (e *EngineeringState) MarkReview(r *ReviewResult) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Review = r
	e.Evidence.Review = r != nil && strings.EqualFold(r.Verdict, "pass")
	if e.Evidence.Review {
		e.Phase = PhaseReview
	} else {
		e.Phase = PhaseRefine
	}
}

// NeedsGates reports whether hybrid gates apply (agent mode + non-trivial).
func (e *EngineeringState) NeedsGates(mode string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if mode != "" && mode != "agent" && mode != "code" {
		return false
	}
	return e.TaskClass == TaskNonTrivial
}

// HasInspect returns inspect evidence.
func (e *EngineeringState) HasInspect() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Evidence.Inspect
}

// HasVerify returns verify evidence.
func (e *EngineeringState) HasVerify() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Evidence.Verify
}

// HasReviewPass returns whether review passed.
func (e *EngineeringState) HasReviewPass() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Evidence.Review
}

// HasFilesChanged returns whether files were mutated.
func (e *EngineeringState) HasFilesChanged() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.FilesChanged
}

// CanRetryReview reports whether another auto-review is allowed.
func (e *EngineeringState) CanRetryReview() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.ReviewRetries < e.MaxReviewRetries
}

// IncReviewRetry increments the review attempt counter.
func (e *EngineeringState) IncReviewRetry() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ReviewRetries++
}

// isMutatingToolCall reports whether this tool call would change the workspace.
func isMutatingToolCall(toolName string, args map[string]interface{}) bool {
	switch toolName {
	case "filesystem":
		op := argStr(args, "operation")
		return op == "write" || op == "edit"
	case "git":
		op := argStr(args, "operation")
		return op == "add" || op == "commit" || op == "branch"
	default:
		return false
	}
}

// isInspectToolCall reports whether the call gathers codebase evidence.
func isInspectToolCall(toolName string, args map[string]interface{}) bool {
	switch toolName {
	case "filesystem":
		op := argStr(args, "operation")
		return op == "read" || op == "list" || op == "exists"
	case "search", "versions":
		return true
	case "git":
		op := argStr(args, "operation")
		return op == "status" || op == "diff" || op == "log" || op == "show" || op == "branch"
	case "shell":
		cmd := strings.ToLower(argStr(args, "command"))
		return isInspectShellCommand(cmd)
	default:
		return false
	}
}

func isInspectShellCommand(cmd string) bool {
	prefixes := []string{
		"ls ", "dir ", "find ", "rg ", "grep ", "cat ", "head ", "tail ",
		"git status", "git diff", "git log", "git show", "git branch",
		"go list", "go env", "go version", "npm ls", "npm list",
		"tree ", "type ", "Get-ChildItem", "Get-Content",
	}
	cmd = strings.TrimSpace(cmd)
	for _, p := range prefixes {
		if strings.HasPrefix(cmd, p) || cmd == strings.TrimSpace(p) {
			return true
		}
	}
	return false
}

// isVerifyShellCommand reports whether a successful shell/test command counts as verification.
func isVerifyShellCommand(cmd string) bool {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	patterns := []string{
		"go test", "go build", "go vet", "gofmt", "golangci-lint",
		"npm test", "npm run test", "npm run lint", "npx tsc", "npx eslint",
		"yarn test", "pnpm test", "pytest", "cargo test", "cargo check",
		"make test", "make check", "dotnet test", "mvn test",
		"python -m compileall", "tsc --noEmit",
	}
	for _, p := range patterns {
		if strings.Contains(cmd, p) {
			return true
		}
	}
	return false
}

// toolResultSucceeded reports whether a tool result looks successful.
func toolResultSucceeded(result interface{}) bool {
	m, ok := result.(map[string]interface{})
	if !ok {
		return result != nil
	}
	if passed, ok := m["passed"].(bool); ok {
		return passed
	}
	if exit, ok := m["exitCode"].(float64); ok {
		return exit == 0
	}
	if exit, ok := m["exitCode"].(int); ok {
		return exit == 0
	}
	if errStr, ok := m["error"].(string); ok && errStr != "" {
		return false
	}
	return true
}

// checkWriteGate returns an error if a mutating call is blocked.
func checkWriteGate(eng *EngineeringState, mode string, toolName string, args map[string]interface{}) error {
	if eng == nil || !isMutatingToolCall(toolName, args) {
		return nil
	}
	if !eng.NeedsGates(mode) {
		return nil
	}
	if eng.HasInspect() {
		return nil
	}
	return fmt.Errorf("engineering gate: inspect before modifying — call filesystem(read/list), search, or git(status/diff) first so there is evidence you understood the codebase")
}

// checkCompleteGate returns an error if task_complete is blocked pending verify.
func checkCompleteGate(eng *EngineeringState, mode string) error {
	if eng == nil {
		return nil
	}
	// Only gated (non-trivial) agent tasks must verify before complete.
	// Local/trivial edits should not pay for a full test suite round-trip.
	if !eng.HasFilesChanged() || !eng.NeedsGates(mode) {
		return nil
	}
	if !eng.HasVerify() {
		return fmt.Errorf("engineering gate: verify before complete — run tests or a build/lint command (e.g. go test ./agent, npm test) after changing files; do not claim verification without tool evidence")
	}
	return nil
}

// needsAutoReview reports whether a self-review must run before complete.
// Reserved for gated non-trivial work — not every file tweak.
func needsAutoReview(eng *EngineeringState, mode string) bool {
	if eng == nil || !eng.NeedsGates(mode) {
		return false
	}
	if !eng.HasFilesChanged() {
		return false
	}
	return !eng.HasReviewPass()
}

// parseReviewResult extracts a ReviewResult from model JSON (or fenced JSON).
func parseReviewResult(content string) (*ReviewResult, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("empty review content")
	}
	// Strip markdown fences if present.
	if i := strings.Index(content, "{"); i >= 0 {
		if j := strings.LastIndex(content, "}"); j > i {
			content = content[i : j+1]
		}
	}
	var r ReviewResult
	if err := json.Unmarshal([]byte(content), &r); err != nil {
		return nil, err
	}
	r.Verdict = strings.ToLower(strings.TrimSpace(r.Verdict))
	if r.Verdict != "pass" && r.Verdict != "revise" {
		if len(r.SecurityIssues) > 0 || len(r.Races) > 0 || r.AISlopRisk || r.UnnecessaryComplexity || !r.SolvesProblem {
			r.Verdict = "revise"
		} else {
			r.Verdict = "pass"
		}
	}
	return &r, nil
}

// RoleToolAllowlist returns real tool names allowed for a role.
func RoleToolAllowlist(role string) []string {
	switch role {
	case "explorer", "researcher":
		return []string{"filesystem", "search", "git", "memory", "versions"}
	case "architect":
		return []string{"filesystem", "search", "memory", "versions"}
	case "coder", "refactorer":
		return []string{"filesystem", "search", "shell", "git", "test", "memory", "versions"}
	case "debugger":
		return []string{"filesystem", "search", "shell", "test", "memory", "versions"}
	case "tester":
		return []string{"filesystem", "shell", "test", "search", "memory"}
	case "reviewer":
		return []string{"filesystem", "search", "git", "memory"}
	case "documentation":
		return []string{"filesystem", "search", "memory"}
	default:
		return nil // all tools
	}
}

// ToolAllowedForRole reports whether toolName is permitted for role.
func ToolAllowedForRole(role, toolName string) bool {
	allow := RoleToolAllowlist(role)
	if allow == nil {
		return true
	}
	// Meta tools always allowed for subagents that receive them.
	if toolName == "update_plan" || toolName == "task_complete" {
		return true
	}
	for _, a := range allow {
		if a == toolName {
			return true
		}
	}
	return false
}
