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
	Role    Role
	Content string
	Name    string
}
