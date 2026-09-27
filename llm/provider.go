package llm

import (
	"context"
)

// StreamEventType defines the type of event in a streaming LLM response.
type StreamEventType string

const (
	StreamEventContent StreamEventType = "content"
	StreamEventToolCall StreamEventType = "tool_call"
	StreamEventError   StreamEventType = "error"
	StreamEventDone    StreamEventType = "done"
)

// StreamEvent represents a single event chunk in a streaming response.
type StreamEvent struct {
	Type     StreamEventType `json:"type"`
	Content  string          `json:"content,omitempty"`
	ToolCall *ToolCall       `json:"toolCall,omitempty"`
	Error    error           `json:"-"`
}

// Provider defines the common interface for all LLM backend integrations.
type Provider interface {
	Name() string
	ListModels(ctx context.Context) ([]Model, error)
	// Ping performs a real authenticated request to verify connectivity and credentials.
	Ping(ctx context.Context) error
	Chat(ctx context.Context, request CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, request CompletionRequest) (<-chan StreamEvent, error)
	Capabilities() ProviderCapabilities
}
