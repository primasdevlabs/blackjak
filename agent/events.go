package agent

import (
	"sync"
	"sync/atomic"
	"time"
)

// EventType defines the type of event produced during agent execution.
type EventType string

const (
	// Run lifecycle events
	EventRunCreated   EventType = "run.created"
	EventRunStarted   EventType = "run.started"
	EventRunCompleted EventType = "run.completed"
	EventRunFailed    EventType = "run.failed"
	EventRunCancelled EventType = "run.cancelled"
	EventRunPaused    EventType = "run.paused"
	EventRunResumed   EventType = "run.resumed"

	// Subagent lifecycle & orchestration events
	EventAgentCreated   EventType = "agent.created"
	EventAgentStarted   EventType = "agent.started"
	EventAgentWaiting   EventType = "agent.waiting"
	EventAgentCompleted EventType = "agent.completed"
	EventAgentFailed    EventType = "agent.failed"
	EventAgentCancelled EventType = "agent.cancelled"
	EventAgentDelegated EventType = "agent.delegated"
	EventAgentHandoff   EventType = "agent.handoff"
	EventAgentActivity  EventType = "agent.activity"

	// Agent thinking & message events
	EventAgentMessage  EventType = "agent.message"
	EventAgentThinking EventType = "agent.thinking"
	EventAgentPlan     EventType = "agent.plan"

	// Tool execution events
	EventToolStarted   EventType = "tool.started"
	EventToolOutput    EventType = "tool.output"
	EventToolCompleted EventType = "tool.completed"
	EventToolFailed    EventType = "tool.failed"

	// File operation events
	EventFileRead     EventType = "file.read"
	EventFileCreated  EventType = "file.created"
	EventFileModified EventType = "file.modified"
	EventFileDeleted  EventType = "file.deleted"
	EventFileRenamed  EventType = "file.renamed"
	EventFileMoved    EventType = "file.moved"
	EventFileChangeReviewed EventType = "file.change.reviewed"

	// Workspace intelligence & lease events
	EventWorkspaceReference  EventType = "workspace.reference"
	EventWorkspaceAttachment EventType = "workspace.attachment"
	EventWorkspaceConflict   EventType = "workspace.conflict"
	EventWorkspaceLocked     EventType = "workspace.locked"
	EventWorkspaceUnlocked   EventType = "workspace.unlocked"
	EventWorkspaceOpenFile   EventType = "workspace.openFile"
	EventWorkspaceRevealFile EventType = "workspace.revealFile"
	EventWorkspaceOpenFolder EventType = "workspace.openFolder"
	EventDiffAvailable       EventType = "diff.available"

	// Command execution events
	EventCommandStarted   EventType = "command.started"
	EventCommandOutput    EventType = "command.output"
	EventCommandCompleted EventType = "command.completed"

	// Test execution events
	EventTestStarted EventType = "test.started"
	EventTestOutput  EventType = "test.output"
	EventTestPassed  EventType = "test.passed"
	EventTestFailed  EventType = "test.failed"

	// Approval events
	EventApprovalRequired EventType = "approval.required"
	EventApprovalGranted  EventType = "approval.granted"
	EventApprovalDenied    EventType = "approval.denied"

	// Context & Memory update events
	EventContextUpdated EventType = "context.updated"
	EventMemoryUpdated  EventType = "memory.updated"
)

// Event represents a structured, timestamped event in the agent lifecycle.
type Event struct {
	ID        string    `json:"id"`
	RunID     string    `json:"runId"`
	AgentID   string    `json:"agentId,omitempty"`
	Type      EventType `json:"type"`
	Sequence  uint64    `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// EventBroker manages thread-safe event subscriptions and fan-out broadcasting.
type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan Event]string // chan -> optional filter runID ("" for all)
	seq         atomic.Uint64
}

// NewEventBroker initializes a new EventBroker.
func NewEventBroker() *EventBroker {
	return &EventBroker{
		subscribers: make(map[chan Event]string),
	}
}

// Subscribe returns a channel that receives published events matching runID (or all if runID is empty),
// along with an unsubscribe cleanup function.
func (b *EventBroker) Subscribe(runID string) (<-chan Event, func()) {
	ch := make(chan Event, 256)
	b.mu.Lock()
	b.subscribers[ch] = runID
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}
		b.mu.Unlock()
	}

	return ch, unsubscribe
}

// Publish broadcasts an event to all matching subscribers non-blockingly.
func (b *EventBroker) Publish(event Event) {
	if event.Sequence == 0 {
		event.Sequence = b.seq.Add(1)
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch, runFilter := range b.subscribers {
		if runFilter == "" || runFilter == event.RunID {
			select {
			case ch <- event:
			default:
				// Avoid blocking publisher if subscriber buffer is full
			}
		}
	}
}
