package agent

import (
	"fmt"
	"time"

	"blackjak/llm"
	"blackjak/memory"
	"blackjak/tools"
	"blackjak/workspace"
)

// RulesProvider supplies always-on rule text for the system prompt.
type RulesProvider func() string

// SkillProvider supplies an active skill body for the current run (optional).
type SkillProvider func(run *Run) string

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
	rulesFn    RulesProvider
	skillFn    SkillProvider
	modeFn      func() string
	askPolicyFn func() string
	browserFn   func() (enabled bool, allowlist []string)
	retrieveFn  func(query string, limit int) []string
	snapshotFn  func() string
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

// SetRulesProvider installs always-on rules injection.
func (a *Agent) SetRulesProvider(fn RulesProvider) {
	a.rulesFn = fn
}

// SetSkillProvider installs per-run skill injection.
func (a *Agent) SetSkillProvider(fn SkillProvider) {
	a.skillFn = fn
}

// SetModeProvider installs agent mode (ask|plan|agent) for prompt policy.
func (a *Agent) SetModeProvider(fn func() string) {
	a.modeFn = fn
}

// SetAskPolicyProvider installs ask-policy resolution (standard|accept_all).
func (a *Agent) SetAskPolicyProvider(fn func() string) {
	a.askPolicyFn = fn
}

// SetBrowserToolProvider installs beta browser_fetch configuration.
func (a *Agent) SetBrowserToolProvider(fn func() (enabled bool, allowlist []string)) {
	a.browserFn = fn
}

// SetRetriever installs workspace retrieval for prompt enrichment.
func (a *Agent) SetRetriever(fn func(query string, limit int) []string) {
	a.retrieveFn = fn
}

// SetRepoSnapshotProvider installs a lightweight repo snapshot injector.
func (a *Agent) SetRepoSnapshotProvider(fn func() string) {
	a.snapshotFn = fn
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
