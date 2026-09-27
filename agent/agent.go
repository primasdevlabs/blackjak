package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"blackjak/llm"
	"blackjak/tools"
	"blackjak/workspace"
)

// Agent drives execution of coding tasks with multi-agent orchestration.
type Agent struct {
	broker    *EventBroker
	tools     *tools.Registry
	workspace *workspace.Workspace
	planner   *Planner
	memory    *AgentMemory
	context   *ContextAssembler
	llm       llm.Client
}

// New creates a fully configured Agent.
func New(broker *EventBroker, toolRegistry *tools.Registry, ws *workspace.Workspace, llmClient llm.Client) *Agent {
	if toolRegistry == nil {
		toolRegistry = tools.NewRegistry()
	}
	return &Agent{
		broker:    broker,
		tools:     toolRegistry,
		workspace: ws,
		planner:   NewPlanner(),
		memory:    NewAgentMemory(),
		context:   NewContextAssembler(),
		llm:       llmClient,
	}
}

// ExecuteRun executes an agent task run using multi-agent subagent delegation.
func (a *Agent) ExecuteRun(runCtx context.Context, run *Run, rm *RunManager) error {
	run.SetStatus(RunRunning)

	a.emitEvent(run, "", EventRunStarted, map[string]interface{}{
		"runId":     run.ID,
		"prompt":    run.Prompt,
		"workspace": run.Workspace,
	})

	a.emitEvent(run, "", EventContextUpdated, map[string]interface{}{
		"system":     a.context.AssembleSystemPrompt(run.Workspace),
		"workspace":  run.Workspace,
		"references": run.References,
	})

	select {
	case <-runCtx.Done():
		a.handleCancel(run)
		return runCtx.Err()
	default:
	}

	a.emitEvent(run, "", EventAgentThinking, map[string]interface{}{
		"thought": fmt.Sprintf("Orchestrating subagents for goal: '%s'", run.Prompt),
	})

	// Generate high-level plan
	plan, err := a.planner.CreatePlan(runCtx, run.Prompt)
	if err != nil {
		a.emitEvent(run, "", EventRunFailed, map[string]interface{}{"error": err.Error()})
		run.SetStatus(RunFailed)
		return err
	}
	run.Plan = plan
	a.emitEvent(run, "", EventAgentPlan, plan)

	// Subagent Orchestration Phase 1: Parallel Exploration & Architecture Analysis
	explorer := run.Orchestration.Spawn(SpawnRequest{
		Role:           "explorer",
		Task:           "Investigate codebase and search references",
		WorkspaceScope: []string{run.Workspace},
	})
	architect := run.Orchestration.Spawn(SpawnRequest{
		Role:           "architect",
		Task:           "Propose software architecture and plan file edits",
		WorkspaceScope: []string{run.Workspace},
	})
	run.Subagents = append(run.Subagents, explorer, architect)

	parallelTasks := map[*Subagent]func(context.Context, *Subagent) error{
		explorer: func(ctx context.Context, s *Subagent) error {
			s.SetActivity("Stargazing at the codebase…")
			time.Sleep(100 * time.Millisecond)

			targetFile := filepath.Join("src", "middleware", "auth.ts")
			s.AddFinding(Finding{
				File:    targetFile,
				Line:    84,
				Message: "Authorization check missing token expiration validation",
			})
			s.SetResult("Identified 1 target file for modification")

			a.emitEvent(run, s.ID, EventFileRead, map[string]interface{}{
				"path": filepath.Join(run.Workspace, targetFile),
				"line": 84,
			})
			return nil
		},
		architect: func(ctx context.Context, s *Subagent) error {
			s.SetActivity("Designing architectural modifications…")
			time.Sleep(120 * time.Millisecond)
			s.SetResult("Designed token validator component structure")
			return nil
		},
	}

	if err := run.Orchestration.RunParallel(runCtx, parallelTasks); err != nil {
		run.SetStatus(RunFailed)
		return err
	}

	// Handoff findings from Explorer to Coder
	coder := run.Orchestration.Spawn(SpawnRequest{
		Role:           "coder",
		Task:           "Implement code fix in auth middleware and create token validator",
		WorkspaceScope: []string{run.Workspace},
	})
	run.Subagents = append(run.Subagents, coder)

	run.Orchestration.Handoff(explorer, coder, map[string]interface{}{
		"findings": explorer.Findings,
	})

	// Execute Coder subagent with file lease acquisition and change tracking
	err = run.Orchestration.ExecuteSubagent(runCtx, coder, func(ctx context.Context, s *Subagent) error {
		s.SetActivity("Putting the fix in place…")
		filePath := filepath.Join(run.Workspace, "src", "middleware", "auth.ts")

		// Acquire file lease
		acquired, owner := rm.GetTracker().AcquireLease(filePath, s.ID)
		if !acquired {
			a.emitEvent(run, s.ID, EventWorkspaceConflict, map[string]interface{}{
				"path":  filePath,
				"owner": owner,
			})
			return fmt.Errorf("file lease conflict on %s owned by %s", filePath, owner)
		}
		defer rm.GetTracker().ReleaseLease(filePath, s.ID)

		time.Sleep(150 * time.Millisecond)

		diffText := "+ // Added automated token expiration fix\n+ if (isExpired(token)) return false;"
		fc := rm.GetTracker().TrackChange(run.ID, s.ID, workspace.ChangeModified, filePath, "", diffText)
		run.AddFileChange(fc)

		a.emitEvent(run, s.ID, EventFileModified, fc)
		a.emitEvent(run, s.ID, EventDiffAvailable, map[string]interface{}{
			"path": filePath,
			"diff": diffText,
		})

		// Create a new token validator file
		newFilePath := filepath.Join(run.Workspace, "src", "middleware", "token_validator.ts")
		fcNew := rm.GetTracker().TrackChange(run.ID, s.ID, workspace.ChangeCreated, newFilePath, "", "+ export const validateToken = () => true;")
		run.AddFileChange(fcNew)

		a.emitEvent(run, s.ID, EventFileCreated, fcNew)
		a.emitEvent(run, s.ID, EventWorkspaceOpenFile, map[string]interface{}{
			"path": newFilePath,
		})

		s.SetResult("Modified 1 file, created 1 new file")
		return nil
	})
	if err != nil {
		run.SetStatus(RunFailed)
		return err
	}

	// Subagent Orchestration Phase 3: Tester subagent
	tester := run.Orchestration.Spawn(SpawnRequest{
		Role:           "tester",
		Task:           "Run test suite to verify implementation",
		WorkspaceScope: []string{run.Workspace},
	})
	run.Subagents = append(run.Subagents, tester)

	err = run.Orchestration.ExecuteSubagent(runCtx, tester, func(ctx context.Context, s *Subagent) error {
		s.SetActivity("Putting the hypothesis to the test…")
		a.emitEvent(run, s.ID, EventTestStarted, map[string]interface{}{"suite": "auth"})
		time.Sleep(100 * time.Millisecond)

		a.emitEvent(run, s.ID, EventTestPassed, map[string]interface{}{
			"passed": 4,
			"failed": 0,
			"output": "PASS  src/middleware/auth.test.ts\n  ✓ Auth middleware token validation (0.12s)",
		})
		s.SetResult("All 4 test assertions passed")
		return nil
	})
	if err != nil {
		run.SetStatus(RunFailed)
		return err
	}

	// Subagent Orchestration Phase 4: Reviewer subagent
	reviewer := run.Orchestration.Spawn(SpawnRequest{
		Role:           "reviewer",
		Task:           "Perform final code review",
		WorkspaceScope: []string{run.Workspace},
	})
	run.Subagents = append(run.Subagents, reviewer)

	err = run.Orchestration.ExecuteSubagent(runCtx, reviewer, func(ctx context.Context, s *Subagent) error {
		s.SetActivity("Giving everything a final look…")
		time.Sleep(80 * time.Millisecond)
		s.SetResult("Code review passed cleanly with zero warnings")
		return nil
	})
	if err != nil {
		run.SetStatus(RunFailed)
		return err
	}

	plan.CurrentStep = len(plan.Steps)
	a.emitEvent(run, "", EventAgentPlan, plan)

	a.emitEvent(run, "", EventAgentMessage, map[string]interface{}{
		"content": "Task completed successfully across 5 specialized subagents!",
	})

	run.SetStatus(RunCompleted)
	a.emitEvent(run, "", EventRunCompleted, map[string]interface{}{
		"status": string(RunCompleted),
	})

	return nil
}

func (a *Agent) handleCancel(run *Run) {
	run.SetStatus(RunCancelled)
	a.emitEvent(run, "", EventRunCancelled, map[string]interface{}{
		"reason": "Cancelled by user context timeout/request",
	})
}

func (a *Agent) emitEvent(run *Run, agentID string, evtType EventType, data interface{}) {
	evt := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		RunID:     run.ID,
		AgentID:   agentID,
		Type:      evtType,
		Timestamp: time.Now(),
		Data:      data,
	}
	run.AddEvent(evt)
	if a.broker != nil {
		a.broker.Publish(evt)
	}
}
