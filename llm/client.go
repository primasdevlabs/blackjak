package llm

import (
	"context"
)

// Client provides an interface to query language models.
type Client interface {
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

// CompletionRequest holds inputs sent to the LLM.
type CompletionRequest struct {
	Messages []Message
	Tools    []Tool
}

// CompletionResponse holds output returned by the LLM.
type CompletionResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// ToolCall represents a requested function execution from the LLM.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}
