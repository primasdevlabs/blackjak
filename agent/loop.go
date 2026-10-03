package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	bjctx "blackjak/context"
	"blackjak/llm"
	"blackjak/memory"
	"blackjak/tools"
	"blackjak/workspace"
)

// ClientResolver resolves an LLM client for a semantic role
// ("coding", "thinking", "fast", "review"). Implemented by the API layer's
// settings manager so model routing changes take effect per run.
type ClientResolver func(role string) llm.Client

// PolicyResolver resolves the active guardrail policy. Implemented by the
// API layer's settings manager so policy changes take effect per run.
type PolicyResolver func() tools.Policy

const (
	maxSubagentDepth = 2
	defaultMaxTokens = 8192
)

// completionResult carries the run's final summary and follow-up suggestions.
type completionResult struct {
	Summary         string   `json:"summary"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// approvalGate adapts RunManager approval to tools.ApprovalFunc.
func approvalGate(rm *RunManager, run *Run) tools.ApprovalFunc {
	return func(ctx context.Context, operation, description string) (bool, error) {
		if rm == nil {
			return false, fmt.Errorf("no run manager for approval")
		}
		req := ApprovalRequest{
			ID:          fmt.Sprintf("appr_%d", time.Now().UnixNano()),
			RunID:       run.ID,
			Operation:   operation,
			Description: description,
		}
		granted, reason, err := rm.RequestApproval(ctx, run, req)
		if err != nil {
			return false, err
		}
		if !granted {
			return false, fmt.Errorf("denied: %s", reason)
		}
		return true, nil
	}
}

// askUserGate adapts RunManager approval into a walkthrough question prompt.
// The selected answer is returned via ApprovalResponse.Reason.
func askUserGate(rm *RunManager, run *Run) tools.AskUserFunc {
	return func(ctx context.Context, question string, options []string, allowOther bool) (string, error) {
		if rm == nil {
			return "", fmt.Errorf("no run manager for ask_user")
		}
		req := ApprovalRequest{
			ID:          fmt.Sprintf("ask_%d", time.Now().UnixNano()),
			RunID:       run.ID,
			Operation:   "ask_user",
			Description: question,
			Details: map[string]any{
				"kind":       "walkthrough",
				"options":    options,
				"allowOther": allowOther,
			},
		}
		granted, reason, err := rm.RequestApproval(ctx, run, req)
		if err != nil {
			return "", err
		}
		if !granted {
			return "", fmt.Errorf("user dismissed question: %s", reason)
		}
		if strings.TrimSpace(reason) == "" {
			return "", fmt.Errorf("empty answer")
		}
		return reason, nil
	}
}

// toolOutcome is the serialized result fed back to the model.
// Large revert snapshots (previousContent) stay on the FileChange for the UI
// but are stripped here — they otherwise bloat every subsequent LLM turn.
func toolOutcome(result interface{}, err error) string {
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	result = stripModelHeavyFields(result)
	data, merr := json.Marshal(result)
	if merr != nil {
		return fmt.Sprintf("%v", result)
	}
	out := string(data)
	if len(out) > 12288 {
		out = out[:12288] + "… [truncated]"
	}
	return out
}

func stripModelHeavyFields(result interface{}) interface{} {
	m, ok := result.(map[string]interface{})
	if !ok {
		return result
	}
	if _, has := m["previousContent"]; !has {
		return result
	}
	cp := make(map[string]interface{}, len(m))
	for k, v := range m {
		if k == "previousContent" {
			continue
		}
		cp[k] = v
	}
	if prev, ok := m["previousContent"].(string); ok && prev != "" {
		cp["previousBytes"] = len(prev)
		cp["previousOmitted"] = true
	}
	return cp
}

// ExecuteRun runs the real LLM tool-calling loop for a task.
func (a *Agent) ExecuteRun(runCtx context.Context, run *Run, rm *RunManager) error {
	run.SetStatus(RunRunning)

	a.emitEvent(run, "", EventRunStarted, map[string]interface{}{
		"runId":     run.ID,
		"prompt":    run.Prompt,
		"workspace": run.Workspace,
	})

	client := a.resolveClient("coding")
	if client == nil {
		err := fmt.Errorf("no LLM provider configured — set an active provider and API key in settings")
		a.emitEvent(run, "", EventRunFailed, map[string]interface{}{"error": err.Error()})
		run.SetStatus(RunFailed)
		return err
	}

	policy := a.policy()
	working := memory.NewWorkingMemory()
	if run.Engineering == nil {
		run.Engineering = NewEngineeringState(run.Prompt)
	}
	fastPath := run.Engineering.IsFastPath()
	registry := a.buildRegistry(run, rm, working, policy)
	if fastPath {
		// Trivial / demo tasks: filesystem + finish only. No ask_user, shell,
		// search, or delegate — those caused the "Waiting for input" + dir theater.
		registry.FilterKeep([]string{"filesystem"})
	}
	toolSchemas := append(registry.Schemas(), taskCompleteSchema())
	if !fastPath {
		toolSchemas = append(toolSchemas, planSchema(), delegateSchema())
	}
	a.emitPhase(run)

	systemPrompt := a.buildSystemPrompt(run, working, policy)
	messages := run.GetMessages()
	resumed := len(messages) > 0
	if !resumed {
		userContent := run.Prompt
		if fastPath {
			userContent = fastPathUserDirective() + "\n\n" + run.Prompt
		} else if a.retrieveFn != nil {
			if hits := a.retrieveFn(run.Prompt, 6); len(hits) > 0 {
				userContent = "Retrieved workspace context:\n- " + strings.Join(hits, "\n- ") + "\n\n" + run.Prompt
			}
		}
		messages = []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: userContent},
		}
		run.SetMessages(messages)
	} else {
		// Resumed run: refresh the system prompt so current guardrails and
		// memory apply, keeping the rest of the conversation intact.
		if messages[0].Role == llm.RoleSystem {
			messages[0] = llm.Message{Role: llm.RoleSystem, Content: systemPrompt}
		}
		run.SetMessages(messages)
		a.emitEvent(run, "", EventRunResumed, map[string]interface{}{
			"runId":         run.ID,
			"messageCount":  len(messages),
		})
	}

	a.emitEvent(run, "", EventContextUpdated, map[string]interface{}{
		"workspace":  run.Workspace,
		"references": run.References,
		"toolCount":  len(toolSchemas),
		"fastPath":   fastPath,
	})

	maxSteps := policy.StepsLimit()
	if fastPath && maxSteps > 8 {
		maxSteps = 8
	}
	completed, err := a.runLoop(runCtx, run, rm, client, registry, toolSchemas, messages, 0, maxSteps)
	if err != nil {
		if errors.Is(err, errRunPaused) || run.IsPauseRequested() {
			run.SetStatus(RunPaused)
			a.emitEvent(run, "", EventRunPaused, map[string]interface{}{
				"runId":   run.ID,
				"message": "Run paused — context checkpointed",
			})
			// Persist terminal status (defer may have written "running").
			checkpointRun(rm, run, 0, run.GetMessages())
			return nil
		}
		if runCtx.Err() != nil {
			a.handleCancel(run)
			checkpointRun(rm, run, 0, run.GetMessages())
			return runCtx.Err()
		}
		run.Error = err.Error()
		a.emitEvent(run, "", EventRunFailed, map[string]interface{}{"error": err.Error()})
		run.SetStatus(RunFailed)
		checkpointRun(rm, run, 0, run.GetMessages())
		return err
	}

	run.SetStatus(RunCompleted)
	a.emitEvent(run, "", EventRunCompleted, map[string]interface{}{
		"status":          string(RunCompleted),
		"result":          completed.Summary,
		"recommendations": completed.Recommendations,
		"durationMs":      time.Since(run.CreatedAt).Milliseconds(),
		"filesChanged":    len(run.FileChanges),
		"subagents":       len(run.Subagents),
		"steps":           countToolEvents(run),
	})
	checkpointRun(rm, run, 0, run.GetMessages())
	return nil
}

// countToolEvents returns how many tool executions happened during the run.
func countToolEvents(run *Run) int {
	n := 0
	for _, evt := range run.GetEvents() {
		if evt.Type == EventToolCompleted || evt.Type == EventToolFailed {
			n++
		}
	}
	return n
}

// checkpointRun persists the main loop's message state. Subagent loops
// (depth > 0) share the run's message slot but never overwrite its checkpoint.
func checkpointRun(rm *RunManager, run *Run, depth int, messages []llm.Message) {
	if depth == 0 && rm != nil {
		rm.SaveCheckpoint(run, messages)
	}
}

// checkpointRunMaybe debounces disk writes during hot tool loops.
func checkpointRunMaybe(rm *RunManager, run *Run, depth int, messages []llm.Message, last *time.Time, force bool) {
	if depth != 0 || rm == nil {
		return
	}
	if !force && last != nil && !last.IsZero() && time.Since(*last) < 2*time.Second {
		return
	}
	rm.SaveCheckpoint(run, messages)
	if last != nil {
		*last = time.Now()
	}
}

// runLoop executes LLM turns until the model stops calling tools or calls
// task_complete. Returns the final completion result. maxSteps is a
// guardrail — when it's hit the loop terminates rather than running forever.
// errRunPaused signals that the loop stopped because a pause was requested —
// distinct from ctx.Err() cancellation so the run ends 'paused', not 'cancelled'.
var errRunPaused = fmt.Errorf("run paused")

func (a *Agent) runLoop(ctx context.Context, run *Run, rm *RunManager,
	client llm.Client, registry *tools.Registry, toolSchemas []llm.Tool,
	messages []llm.Message, depth int, maxSteps int) (completionResult, error) {

	window := resolveContextWindow(client)
	maxOut := defaultMaxTokens
	if window > 0 && maxOut > window/4 {
		maxOut = window / 4
		if maxOut < 1024 {
			maxOut = 1024
		}
	}
	if run.Engineering != nil && run.Engineering.IsFastPath() && maxOut > 4096 {
		// Demo/scaffold tasks need a short HTML/file write, not an 8k essay.
		maxOut = 4096
	}
	var lastCheckpoint time.Time

	// Persist the conversation on every exit path — completed, failed,
	// paused, or cancelled — so resume always has the latest context.
	// Only the main loop (depth 0) owns the run checkpoint; subagent loops
	// share the run's message slot but must not overwrite its checkpoint.
	defer func() {
		if depth == 0 && rm != nil {
			rm.SaveCheckpoint(run, messages)
		}
	}()

	for i := 0; i < maxSteps; i++ {
		if ctx.Err() != nil {
			if run.IsPauseRequested() {
				return completionResult{}, errRunPaused
			}
			return completionResult{}, ctx.Err()
		}
		if run.IsPauseRequested() {
			return completionResult{}, errRunPaused
		}

		// Manual /compact or auto-compact at ≥80% usable pressure for this model.
		compactor := bjctx.NewCompactor(window)
		budget := compactor.CalculateBudget(estimateMessageTokens(messages))
		shouldAuto := depth == 0 && compactor.ShouldCompact(budget)
		if run.TakeCompactionRequest() || shouldAuto {
			before := estimateMessageTokens(messages)
			messages = a.compactMessages(ctx, run, client, messages)
			after := estimateMessageTokens(messages)
			run.SetMessages(messages)
			checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, true)
			a.emitEvent(run, "", EventContextCompacted, map[string]interface{}{
				"compacted":    true,
				"auto":         shouldAuto,
				"tokensBefore": before,
				"tokensAfter":  after,
				"tokensSaved":  before - after,
				"pressure":     budget.Pressure,
				"window":       window,
			})
			a.emitEvent(run, "", EventContextUpdated, map[string]interface{}{
				"compacted":    true,
				"tokensBefore": before,
				"tokensAfter":  after,
				"tokensSaved":  before - after,
			})
		}

		a.emitEvent(run, "", EventAgentThinking, map[string]interface{}{
			"step": i + 1,
		})

		resp, err := client.Complete(ctx, &llm.CompletionRequest{
			Messages:          messages,
			Tools:             toolSchemas,
			MaxTokens:         maxOut,
			EnablePromptCache: true,
		})
		if err != nil {
			return completionResult{}, fmt.Errorf("llm completion failed: %w", err)
		}
		if resp.Cache.InputTokens > 0 || resp.Cache.CacheReadTokens > 0 || resp.Cache.CacheWriteTokens > 0 {
			a.emitEvent(run, "", EventCacheStats, map[string]interface{}{
				"inputTokens":      resp.Cache.InputTokens,
				"outputTokens":     resp.Cache.OutputTokens,
				"cacheReadTokens":  resp.Cache.CacheReadTokens,
				"cacheWriteTokens": resp.Cache.CacheWriteTokens,
			})
		}

		if resp.Content != "" {
			a.emitEvent(run, "", EventAgentMessage, map[string]interface{}{
				"content": resp.Content,
			})
			messages = append(messages, llm.Message{Role: llm.RoleAssistant, Content: resp.Content})
			run.SetMessages(messages)
			checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, false)
		}

		if len(resp.ToolCalls) == 0 {
			// Model produced a final text response with no tool calls.
			if resp.Content != "" {
				return completionResult{Summary: resp.Content}, nil
			}
			return completionResult{Summary: "Task finished"}, nil
		}

		// Append assistant message carrying the tool calls.
		messages = append(messages, llm.Message{
			Role:      llm.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})
		run.SetMessages(messages)
		checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, false)

		parsed := make([]toolCallArgs, len(resp.ToolCalls))
		for i, call := range resp.ToolCalls {
			parsed[i] = toolCallArgs{call: call, args: parseToolArgs(call)}
		}

		runParallel := len(parsed) > 1 && allInspectToolCalls(parsed)
		if runParallel {
			results := a.executeToolCallsParallel(ctx, run, rm, registry, parsed, depth, toolSchemas, i+1)
			for idx, tr := range results {
				call := parsed[idx].call
				messages = append(messages, llm.Message{
					Role:       llm.RoleTool,
					Name:       call.Name,
					ToolCallID: call.ID,
					Content:    toolOutcome(tr.result, tr.err),
				})
				if tr.done != nil {
					run.SetMessages(messages)
					checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, true)
					return *tr.done, nil
				}
			}
			run.SetMessages(messages)
			checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, false)
			continue
		}

		for _, item := range parsed {
			if ctx.Err() != nil || run.IsPauseRequested() {
				// Close out any unanswered tool_calls so resume stays valid.
				messages = ensureToolCallResults(messages)
				run.SetMessages(messages)
				checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, true)
				if run.IsPauseRequested() {
					return completionResult{}, errRunPaused
				}
				return completionResult{}, ctx.Err()
			}

			call, args := item.call, item.args
			a.emitEvent(run, "", EventToolStarted, map[string]interface{}{
				"tool": call.Name,
				"args": args,
				"step": i + 1,
			})

			result, done, err := a.executeToolCall(ctx, run, rm, registry, call, args, depth, toolSchemas)

			if err != nil {
				a.emitEvent(run, "", EventToolFailed, map[string]interface{}{
					"tool":  call.Name,
					"error": err.Error(),
				})
			} else {
				a.emitEvent(run, "", EventToolCompleted, map[string]interface{}{
					"tool":   call.Name,
					"result": stripModelHeavyFields(result),
				})
			}

			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				Name:       call.Name,
				ToolCallID: call.ID,
				Content:    toolOutcome(result, err),
			})
			run.SetMessages(messages)
			checkpointRunMaybe(rm, run, depth, messages, &lastCheckpoint, false)

			if done != nil {
				return *done, nil
			}
		}
	}

	return completionResult{Summary: fmt.Sprintf("Reached the configured step limit (%d)", maxSteps)}, nil
}

// agentMode returns the current mode string (ask|plan|agent).
func (a *Agent) agentMode() string {
	if a.modeFn != nil {
		return a.modeFn()
	}
	return "agent"
}

// emitPhase publishes the current engineering phase snapshot.
func (a *Agent) emitPhase(run *Run) {
	if run == nil || run.Engineering == nil {
		return
	}
	a.emitEvent(run, "", EventAgentPhase, run.Engineering.Snapshot())
}

// executeToolCall dispatches one tool call. Returns (result, done, err).
// done is non-nil when the run should finish.
func (a *Agent) executeToolCall(ctx context.Context, run *Run, rm *RunManager,
	registry *tools.Registry, call llm.ToolCall, args map[string]interface{},
	depth int, toolSchemas []llm.Tool) (interface{}, *completionResult, error) {

	mode := a.agentMode()
	eng := run.Engineering
	if eng == nil {
		eng = NewEngineeringState(run.Prompt)
		run.Engineering = eng
	}

	// Hybrid write gate: non-trivial Agent tasks must inspect before mutating.
	if err := checkWriteGate(eng, mode, call.Name, args); err != nil {
		return nil, nil, err
	}
	if eng.IsFastPath() && (call.Name == "ask_user" || call.Name == "delegate" || call.Name == "shell" || call.Name == "search") {
		return nil, nil, fmt.Errorf("fast path: %s is disabled — use filesystem(write/edit) then task_complete", call.Name)
	}

	switch call.Name {
	case "task_complete":
		if err := checkCompleteGate(eng, mode); err != nil {
			return nil, nil, err
		}
		if depth == 0 && needsAutoReview(eng, mode) {
			review, revErr := a.runSelfReview(ctx, run)
			if revErr != nil {
				return nil, nil, fmt.Errorf("engineering gate: self-review failed: %w", revErr)
			}
			eng.MarkReview(review)
			a.emitPhase(run)
			if review == nil || !strings.EqualFold(review.Verdict, "pass") {
				eng.IncReviewRetry()
				findings, _ := json.Marshal(review)
				return map[string]interface{}{
					"blocked": true,
					"reason":  "self-review requested revisions",
					"review":  review,
				}, nil, fmt.Errorf("engineering gate: self-review verdict=revise — address findings then verify again before task_complete: %s", string(findings))
			}
		}
		var recs []string
		if raw, ok := args["recommendations"].([]interface{}); ok {
			for _, r := range raw {
				if s, ok := r.(string); ok {
					recs = append(recs, s)
				}
			}
		}
		eng.SetPhase(PhaseDone)
		a.emitPhase(run)
		done := &completionResult{Summary: argStr(args, "summary"), Recommendations: recs}
		return map[string]interface{}{"summary": done.Summary, "recommendations": recs}, done, nil

	case "update_plan":
		steps := toStringSlice(args["steps"])
		current := 0
		if v, ok := args["current_step"].(float64); ok {
			current = int(v)
		}
		run.mu.Lock()
		if run.Plan == nil {
			run.Plan = &Plan{}
		}
		run.Plan.Steps = steps
		run.Plan.CurrentStep = current
		run.mu.Unlock()
		eng.MarkPlan()
		a.emitPhase(run)
		a.emitEvent(run, "", EventAgentPlan, run.Plan)
		return map[string]interface{}{"steps": len(steps), "currentStep": current}, nil, nil

	case "delegate":
		res, err := a.executeDelegate(ctx, run, rm, args, depth)
		return res, nil, err

	default:
		tool, err := registry.Get(call.Name)
		if err != nil {
			return nil, nil, err
		}
		result, err := tool.Execute(ctx, args)
		a.trackToolSideEffects(run, rm, call.Name, args, result, err)
		a.recordEngineeringEvidence(run, call.Name, args, result, err)
		return result, nil, err
	}
}

// recordEngineeringEvidence updates phase/evidence from successful tool results.
func (a *Agent) recordEngineeringEvidence(run *Run, toolName string, args map[string]interface{}, result interface{}, execErr error) {
	eng := run.Engineering
	if eng == nil || execErr != nil {
		return
	}
	changed := false
	if isInspectToolCall(toolName, args) {
		eng.MarkInspect()
		changed = true
	}
	if isMutatingToolCall(toolName, args) {
		eng.MarkImplement()
		changed = true
	}
	if toolName == "test" {
		if toolResultSucceeded(result) {
			eng.MarkVerify()
			changed = true
			a.emitEvent(run, "", EventTestPassed, map[string]interface{}{"tool": "test", "result": result})
		} else {
			a.emitEvent(run, "", EventTestFailed, map[string]interface{}{"tool": "test", "result": result})
		}
	}
	if toolName == "shell" && toolResultSucceeded(result) {
		cmd := argStr(args, "command")
		if isVerifyShellCommand(cmd) {
			eng.MarkVerify()
			changed = true
		}
	}
	if changed {
		a.emitPhase(run)
	}
}

// runSelfReview performs a structured critical review of tracked changes.
func (a *Agent) runSelfReview(ctx context.Context, run *Run) (*ReviewResult, error) {
	eng := run.Engineering
	if eng != nil && !eng.CanRetryReview() {
		return nil, fmt.Errorf("self-review retry limit reached")
	}

	client := a.resolveClient("review")
	if client == nil {
		client = a.resolveClient("thinking")
	}
	if client == nil {
		// No review model: synthesize a pass with explicit note (evidence still required for verify).
		return &ReviewResult{
			SolvesProblem: true,
			Verdict:       "pass",
			Findings:      []string{"review model unavailable — skipped structured review"},
		}, nil
	}

	var changes strings.Builder
	for _, fc := range run.FileChanges {
		fmt.Fprintf(&changes, "- %s (%s)\n", fc.Path, fc.Type)
	}
	prompt := `You are performing a critical engineering self-review of Blackjak's own changes.
Respond with ONLY a JSON object (no markdown) matching:
{"solvesProblem":bool,"unnecessaryComplexity":bool,"securityIssues":[],"races":[],"leaks":[],"contractBreaks":[],"aiSlopRisk":bool,"findings":[],"verdict":"pass"|"revise"}
Ask: Does this solve the problem? Unnecessary complexity? Security/race/leak/contract issues? Does it look like AI slop rather than engineered code?
Be strict. Verdict "revise" if any serious issue.`

	resp, err := client.Complete(ctx, &llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: prompt},
			{Role: llm.RoleUser, Content: fmt.Sprintf("User goal:\n%s\n\nChanged files:\n%s", run.Prompt, changes.String())},
		},
		MaxTokens: 1024,
	})
	if err != nil {
		return nil, err
	}
	return parseReviewResult(resp.Content)
}

// toStringSlice converts a JSON array arg to []string.
func toStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// executeDelegate spawns a real subagent with its own LLM loop.
func (a *Agent) executeDelegate(ctx context.Context, run *Run, rm *RunManager,
	args map[string]interface{}, depth int) (interface{}, error) {

	if depth >= maxSubagentDepth {
		return nil, fmt.Errorf("max subagent depth reached")
	}
	if err := a.policy().CheckDelegate(len(run.Subagents)); err != nil {
		return nil, err
	}

	role := argStr(args, "role")
	task := argStr(args, "task")
	if task == "" {
		return nil, fmt.Errorf("delegate requires a task")
	}
	if role == "" {
		role = "worker"
	}

	client := a.resolveClient(role)
	if client == nil {
		return nil, fmt.Errorf("no LLM client available for role %s", role)
	}

	sub := run.Orchestration.Spawn(SpawnRequest{
		Role:           role,
		Task:           task,
		WorkspaceScope: []string{run.Workspace},
	})
	run.Subagents = append(run.Subagents, sub)

	a.emitEvent(run, sub.ID, EventAgentDelegated, map[string]interface{}{
		"role": role,
		"task": task,
	})

	var subResult string
	err := run.Orchestration.ExecuteSubagent(ctx, sub, func(subCtx context.Context, s *Subagent) error {
		s.SetActivity(fmt.Sprintf("Working on: %s", task))

		subWorking := memory.NewWorkingMemory()
		subRegistry := a.buildRegistry(run, rm, subWorking, a.policy())
		if allow := RoleToolAllowlist(role); len(allow) > 0 {
			subRegistry.FilterKeep(allow)
		}
		subTools := append(subRegistry.Schemas(), planSchema(), taskCompleteSchema())

		subPrompt := fmt.Sprintf("%s\n\nYou are a %s subagent. Complete this task and report results:\n\n%s",
			a.buildSystemPrompt(run, subWorking, a.policy()), role, task)

		subMessages := []llm.Message{
			{Role: llm.RoleSystem, Content: subPrompt},
			{Role: llm.RoleUser, Content: task},
		}

		result, err := a.runLoop(subCtx, run, rm, client, subRegistry, subTools, subMessages, depth+1, a.policy().StepsLimit())
		if err != nil {
			return err
		}
		subResult = result.Summary
		s.SetResult(subResult)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("subagent %s failed: %w", sub.ID, err)
	}

	return map[string]interface{}{
		"subagentId": sub.ID,
		"role":       role,
		"result":     subResult,
	}, nil
}

// trackToolSideEffects converts tool results into file/change events.
func (a *Agent) trackToolSideEffects(run *Run, rm *RunManager, toolName string,
	args map[string]interface{}, result interface{}, execErr error) {

	if execErr != nil {
		return
	}

	switch toolName {
	case "filesystem":
		op := argStr(args, "operation")
		m, ok := result.(map[string]interface{})
		if !ok {
			return
		}
		path, _ := m["path"].(string)
		if path == "" {
			return
		}
		abs := filepath.Join(run.Workspace, path)
		existed, _ := m["existed"].(bool)
		prevContent, _ := m["previousContent"].(string)
		snapshotLost, _ := m["snapshotLost"].(bool)
		switch op {
		case "write":
			changeType := workspace.ChangeCreated
			if existed {
				changeType = workspace.ChangeModified
			}
			fc := rm.GetTracker().TrackChange(run.ID, "", changeType, abs, "", "")
			fc.PreviousContent = prevContent
			fc.CanRevert = !snapshotLost
			run.AddFileChange(fc)
			if changeType == workspace.ChangeCreated {
				a.emitEvent(run, "", EventFileCreated, fc)
				a.emitEvent(run, "", EventWorkspaceOpenFile, map[string]interface{}{"path": abs})
			} else {
				a.emitEvent(run, "", EventFileModified, fc)
			}
		case "edit":
			fc := rm.GetTracker().TrackChange(run.ID, "", workspace.ChangeModified, abs, "", "")
			fc.PreviousContent = prevContent
			fc.CanRevert = !snapshotLost
			run.AddFileChange(fc)
			a.emitEvent(run, "", EventFileModified, fc)
		case "read":
			a.emitEvent(run, "", EventFileRead, map[string]interface{}{"path": abs})
		}
	case "shell":
		if m, ok := result.(map[string]interface{}); ok {
			a.emitEvent(run, "", EventCommandCompleted, map[string]interface{}{
				"exitCode": m["exitCode"],
				"stdout":   m["stdout"],
			})
		}
	case "memory":
		a.emitEvent(run, "", EventMemoryUpdated, map[string]interface{}{
			"operation": argStr(args, "operation"),
			"key":       argStr(args, "key"),
		})
	}
}

// buildRegistry creates a per-run tool registry with approval gating and
// guardrail enforcement baked into the tools.
func (a *Agent) buildRegistry(run *Run, rm *RunManager, working *memory.WorkingMemory, policy tools.Policy) *tools.Registry {
	var approve tools.ApprovalFunc
	if rm != nil {
		approve = approvalGate(rm, run)
	}
	var store tools.Store
	if a.persistent != nil {
		store = a.persistent
	}
	var ask tools.AskUserFunc
	if rm != nil {
		ask = askUserGate(rm, run)
	}
	reg := tools.DefaultRegistryWithAsk(workspace.New(run.Workspace), approve, ask, store, policy)
	if a.browserFn != nil {
		enabled, allow := a.browserFn()
		if enabled {
			reg.Register(&tools.BrowserTool{Enabled: true, Allowlist: allow})
		}
	}
	return reg
}

// resolveClient returns the client for a role, falling back to coding.
func (a *Agent) resolveClient(role string) llm.Client {
	if a.resolver == nil {
		return a.llm
	}
	client := a.resolver(role)
	if client == nil && role != "coding" {
		client = a.resolver("coding")
	}
	return client
}

func fastPathUserDirective() string {
	return `FAST PATH — execute immediately:
- Do NOT ask clarifying questions.
- Do NOT explore the repo with shell/search/list.
- Create or edit the requested file(s) with filesystem(write/edit) in your first tool call(s).
- Keep the result minimal (test/demo quality is fine when the user said so).
- Call task_complete as soon as the file(s) exist.`
}

func fastPathSystemPrompt(workspacePath string) string {
	return `You are Blackjak in FAST PATH mode for a simple/demo task.

Tools available: filesystem (read/write/edit/list), task_complete.
Rules:
1. Write the requested file(s) immediately — no planning theater, no questions.
2. Prefer a single new file under the workspace root unless a path was specified.
3. Keep output minimal and self-contained.
4. Finish with task_complete.

Workspace: ` + workspacePath
}

// buildSystemPrompt assembles the full system prompt including persistent
// memory and the active guardrail contract, so the model knows its limits
// before it tries a call that will be rejected.
func (a *Agent) buildSystemPrompt(run *Run, working *memory.WorkingMemory, policy tools.Policy) string {
	if run != nil && run.Engineering != nil && run.Engineering.IsFastPath() {
		return fastPathSystemPrompt(run.Workspace)
	}

	mode := "agent"
	if a.modeFn != nil {
		mode = a.modeFn()
	}
	askPolicy := "standard"
	if a.askPolicyFn != nil {
		askPolicy = a.askPolicyFn()
	}
	rulesText := ""
	if a.rulesFn != nil {
		rulesText = a.rulesFn()
	}
	skillText := ""
	if a.skillFn != nil {
		skillText = a.skillFn(run)
	}
	var refs []string
	for _, r := range run.References {
		refs = append(refs, r.Path)
	}
	var recentDiffs strings.Builder
	for i, fc := range run.FileChanges {
		if i >= 12 {
			break
		}
		fmt.Fprintf(&recentDiffs, "- %s (%s)\n", fc.Path, fc.Type)
	}
	var workingBits []string
	if working != nil {
		if v, _ := working.Get(context.Background(), "retrieved"); v != nil {
			if s, ok := v.(string); ok && s != "" {
				workingBits = append(workingBits, s)
			}
		}
	}
	memText := ""
	if a.snapshotFn != nil {
		if snap := a.snapshotFn(); snap != "" {
			memText = snap
		}
	}
	built := bjctx.NewBuilder().Build(bjctx.BuildInput{
		SystemPrompt: a.context.AssembleSystemPrompt(run.Workspace),
		UserRules:    rulesText,
		SkillBody:    skillText,
		ModePolicy:   bjctx.ModePolicyText(mode),
		References:   refs,
		RecentDiffs:  recentDiffs.String(),
		Retrieved:    workingBits,
		Memory:       memText,
		UserPrompt:   "",
	})

	var b strings.Builder
	b.WriteString(built.System)
	if built.User != "" {
		// Builder puts refs/retrieved/diffs into User; fold into system for the agent loop.
		b.WriteString("\n\n")
		b.WriteString(built.User)
	}

	b.WriteString("\n\n# Ask Policy\n")
	b.WriteString(bjctx.AskPolicyText(askPolicy))
	b.WriteString("\n")

	b.WriteString("\n\nGuardrails currently in effect:\n")
	switch policy.Mode {
	case tools.PolicyReadOnly:
		b.WriteString("- READ-ONLY mode: file writes/edits and non-readonly shell commands will be rejected. Inspect and plan only — do not attempt modifications.\n")
	case tools.PolicyAutonomous:
		b.WriteString("- Autonomous mode: no approval prompts, but deny-listed commands and protected paths are still hard-blocked.\n")
	default:
		b.WriteString("- Supervised mode: destructive commands and git commits require user approval.\n")
	}
	if !policy.ShellAllowed {
		b.WriteString("- Shell commands are disabled entirely.\n")
	}
	if !policy.SubagentsAllowed {
		b.WriteString("- Subagent delegation is disabled.\n")
	}

	if a.persistent != nil {
		entries := a.persistent.Entries()
		if len(entries) > 0 {
			b.WriteString("\n\nPersistent memory from previous sessions:\n")
			for _, e := range entries {
				val, _ := json.Marshal(e.Value)
				fmt.Fprintf(&b, "- %s: %s\n", e.Key, val)
			}
		}
	}

	if len(run.References) > 0 {
		b.WriteString("\nReferenced files:\n")
		for _, r := range run.References {
			fmt.Fprintf(&b, "- %s\n", r.Path)
		}
	}
	if len(run.Attachments) > 0 {
		b.WriteString("\nAttached files:\n")
		for _, at := range run.Attachments {
			fmt.Fprintf(&b, "- %s\n", at.Path)
		}
	}
	return b.String()
}

// estimateMessageTokens approximates token count (~4 chars/token).
func estimateMessageTokens(messages []llm.Message) int {
	total := 0
	for _, m := range messages {
		total += len(m.Content) + len(m.ToolCalls)*64 // rough overhead per call
	}
	return total / 4
}

// compactMessages summarizes the middle of the conversation with the model and
// replaces it with a single context note. Keeps the system prompt, the original
// user prompt, and the tail of recent messages. The cut point is moved forward
// past any leading tool messages so a tool result is never orphaned from the
// assistant call that produced it.
func (a *Agent) compactMessages(ctx context.Context, run *Run, client llm.Client, messages []llm.Message) []llm.Message {
	const keepTail = 8
	if len(messages) <= keepTail+2 {
		return messages
	}

	cut := len(messages) - keepTail
	for cut < len(messages) && messages[cut].Role == llm.RoleTool {
		cut++
	}
	if cut <= 2 {
		return messages
	}

	omitted := messages[2:cut]

	var transcript strings.Builder
	for _, m := range omitted {
		content := m.Content
		if len(content) > 800 {
			content = content[:800] + "…"
		}
		fmt.Fprintf(&transcript, "[%s] %s\n", m.Role, content)
	}

	// Prefer the fast model for compaction — coding model RTT is wasted here.
	summarizer := a.resolveClient("fast")
	if summarizer == nil {
		summarizer = client
	}
	summary := ""
	if summarizer != nil {
		resp, err := summarizer.Complete(ctx, &llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: llm.RoleSystem, Content: "Summarize agent history for continuity as structured bullets: Objective, Decisions, Files touched, Commands/tests and outcomes, Pending work, Risks. Be terse. Do not invent results."},
				{Role: llm.RoleUser, Content: transcript.String()},
			},
			MaxTokens: 768,
		})
		if err == nil {
			summary = resp.Content
		}
	}
	if summary == "" {
		summary = fmt.Sprintf("%d earlier messages were removed to reduce context size.", len(omitted))
	}

	compacted := make([]llm.Message, 0, 2+1+keepTail)
	compacted = append(compacted, messages[0], messages[1])
	compacted = append(compacted, llm.Message{
		Role:    llm.RoleUser,
		Content: "[Earlier context compacted]\n" + summary,
	})
	compacted = append(compacted, messages[cut:]...)
	return compacted
}

type toolCallArgs struct {
	call llm.ToolCall
	args map[string]interface{}
}

type toolCallResult struct {
	result interface{}
	done   *completionResult
	err    error
}

func parseToolArgs(call llm.ToolCall) map[string]interface{} {
	var args map[string]interface{}
	if call.Arguments != "" {
		if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
			args = map[string]interface{}{"_raw": call.Arguments}
		}
	}
	if args == nil {
		args = map[string]interface{}{}
	}
	return args
}

func allInspectToolCalls(items []toolCallArgs) bool {
	for _, item := range items {
		if item.call.Name == "task_complete" || item.call.Name == "update_plan" || item.call.Name == "delegate" {
			return false
		}
		if !isInspectToolCall(item.call.Name, item.args) {
			return false
		}
	}
	return true
}

func (a *Agent) executeToolCallsParallel(ctx context.Context, run *Run, rm *RunManager,
	registry *tools.Registry, items []toolCallArgs, depth int, toolSchemas []llm.Tool, step int) []toolCallResult {
	out := make([]toolCallResult, len(items))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i, item := range items {
		wg.Add(1)
		go func(idx int, item toolCallArgs) {
			defer wg.Done()
			mu.Lock()
			a.emitEvent(run, "", EventToolStarted, map[string]interface{}{
				"tool":     item.call.Name,
				"args":     item.args,
				"step":     step,
				"parallel": true,
			})
			mu.Unlock()

			result, done, err := a.executeToolCall(ctx, run, rm, registry, item.call, item.args, depth, toolSchemas)

			mu.Lock()
			if err != nil {
				a.emitEvent(run, "", EventToolFailed, map[string]interface{}{
					"tool":  item.call.Name,
					"error": err.Error(),
				})
			} else {
				a.emitEvent(run, "", EventToolCompleted, map[string]interface{}{
					"tool":   item.call.Name,
					"result": stripModelHeavyFields(result),
				})
			}
			mu.Unlock()
			out[idx] = toolCallResult{result: result, done: done, err: err}
		}(i, item)
	}
	wg.Wait()
	return out
}

// resolveContextWindow returns the model's context window when known.
func resolveContextWindow(client llm.Client) int {
	const fallback = 128000
	pc, ok := client.(*llm.ProviderClient)
	if !ok || pc == nil || pc.Provider == nil {
		return fallback
	}
	models, err := pc.Provider.ListModels(context.Background())
	if err != nil {
		return fallback
	}
	for _, m := range models {
		if m.ID == pc.Model && m.ContextWindow > 0 {
			return m.ContextWindow
		}
	}
	// OpenAI-compatible providers often list a single dynamic model.
	if len(models) == 1 && models[0].ContextWindow > 0 {
		return models[0].ContextWindow
	}
	return fallback
}

// argStr extracts a string arg.
func argStr(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

// delegateSchema returns the delegate tool definition for the main loop.
func delegateSchema() llm.Tool {
	return llm.Tool{
		Name:        "delegate",
		Description: "Delegate a subtask to a specialized subagent. The subagent gets its own tool access and returns a result.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"role": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"explorer", "coder", "tester", "reviewer"},
					"description": "Subagent specialty",
				},
				"task": map[string]interface{}{
					"type":        "string",
					"description": "Clear description of what the subagent should accomplish",
				},
			},
			"required": []string{"task"},
		},
	}
}

// planSchema returns the update_plan tool definition — lets the model publish
// a visible task list the UI renders as a checklist.
func planSchema() llm.Tool {
	return llm.Tool{
		Name:        "update_plan",
		Description: "Publish or update the task plan shown to the user. Call early with the steps you intend to take, and update current_step as you progress.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"steps": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Ordered task steps",
				},
				"current_step": map[string]interface{}{
					"type":        "integer",
					"description": "Index of the step currently in progress",
				},
			},
			"required": []string{"steps"},
		},
	}
}

// taskCompleteSchema returns the task_complete tool definition.
func taskCompleteSchema() llm.Tool {
	return llm.Tool{
		Name:        "task_complete",
		Description: "Signal that the task is fully complete. Provide a concise summary of what was accomplished and optional follow-up recommendations.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "What was accomplished",
				},
				"recommendations": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Suggested follow-up actions, if any",
				},
			},
			"required": []string{"summary"},
		},
	}
}
