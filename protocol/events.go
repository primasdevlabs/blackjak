package protocol

// Event type names emitted on the wire as ServerMessage.Type. These mirror
// agent.EventType values; they are duplicated here so host adapters can
// depend on the protocol contract without importing agent internals.
const (
	EventRunCreated   = "run.created"
	EventRunStarted   = "run.started"
	EventRunCompleted = "run.completed"
	EventRunFailed    = "run.failed"
	EventRunCancelled = "run.cancelled"

	EventAgentCreated   = "agent.created"
	EventAgentStarted   = "agent.started"
	EventAgentWaiting   = "agent.waiting"
	EventAgentCompleted = "agent.completed"
	EventAgentFailed    = "agent.failed"
	EventAgentCancelled = "agent.cancelled"
	EventAgentDelegated = "agent.delegated"
	EventAgentHandoff   = "agent.handoff"
	EventAgentActivity  = "agent.activity"
	EventAgentMessage   = "agent.message"
	EventAgentThinking  = "agent.thinking"
	EventAgentPlan      = "agent.plan"

	EventToolStarted   = "tool.started"
	EventToolOutput    = "tool.output"
	EventToolCompleted = "tool.completed"
	EventToolFailed    = "tool.failed"

	EventFileRead     = "file.read"
	EventFileCreated  = "file.created"
	EventFileModified = "file.modified"
	EventFileDeleted  = "file.deleted"
	EventFileRenamed  = "file.renamed"
	EventFileMoved    = "file.moved"
	EventFileChanged  = "file.changed"

	EventDiffAvailable = "diff.available"

	EventCommandStarted   = "command.started"
	EventCommandOutput    = "command.output"
	EventCommandCompleted = "command.completed"

	EventTestStarted = "test.started"
	EventTestOutput  = "test.output"
	EventTestPassed  = "test.passed"
	EventTestFailed  = "test.failed"

	EventApprovalRequired = "approval.required"
	EventApprovalGranted  = "approval.granted"
	EventApprovalDenied   = "approval.denied"

	EventContextUpdated   = "context.updated"
	EventContextCompacted = "context.compacted"
	EventMemoryUpdated    = "memory.updated"
)
