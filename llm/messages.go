package llm

// Role represents the author of a message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a prompt/response message item.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`

	// ToolCalls is set on assistant messages that request tool executions.
	ToolCalls []ToolCall `json:"toolCalls,omitempty"`
	// ToolCallID links a tool-result message back to the assistant's call.
	ToolCallID string `json:"toolCallId,omitempty"`
}
