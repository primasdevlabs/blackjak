package agent

import (
	"fmt"
	"time"

	"blackjak/llm"
	"blackjak/memory"
	"blackjak/tools"
	"blackjak/workspace"
)

// Agent drives execution of coding tasks with multi-agent orchestration.
type Agent struct {
	broker     *EventBroker
	workspace  *workspace.Workspace
	planner    *Planner
	context    *ContextAssembler
	llm        llm.Client
	resolver   ClientResolver
	policyFn   PolicyResolver
	persistent *memory.PersistentMemory
}

// New creates a fully configured Agent. llmClient is used as a fallback when
// no ClientResolver is configured (e.g. CLI mode or tests).
func New(broker *EventBroker, ws *workspace.Workspace, llmClient llm.Client) *Agent {
	return &Agent{
		broker:    broker,
		workspace: ws,
		planner:   NewPlanner(),
		context:   NewContextAssembler(),
		llm:       llmClient,
	}
}

// SetClientResolver installs a per-role LLM client resolver.
func (a *Agent) SetClientResolver(r ClientResolver) {
	a.resolver = r
}

// SetPolicyResolver installs a resolver for the runtime guardrail policy.
// Called per run so settings changes take effect without restarting.
func (a *Agent) SetPolicyResolver(r PolicyResolver) {
	a.policyFn = r
}

// policy returns the active guardrails, defaulting to supervised.
func (a *Agent) policy() tools.Policy {
	if a.policyFn != nil {
		return a.policyFn()
	}
	return tools.DefaultPolicy()
}

// SetPersistentMemory installs the durable memory store.
func (a *Agent) SetPersistentMemory(m *memory.PersistentMemory) {
	a.persistent = m
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
