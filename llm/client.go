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
	Model     string
	Messages  []Message
	Tools     []Tool
	MaxTokens int
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
	// ThoughtSignature carries Gemini's per-call signature that must be echoed
	// back verbatim in subsequent requests.
	ThoughtSignature string
}
