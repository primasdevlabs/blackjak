package anthropic

import (
	"context"
	"fmt"

	"blackjak/llm"
)

type AnthropicProvider struct {
	apiKey  string
	baseURL string
}

func New(apiKey, baseURL string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (p *AnthropicProvider) Name() string {
	return "Anthropic"
}

func (p *AnthropicProvider) Capabilities() llm.ProviderCapabilities {
	return llm.ProviderCapabilities{
		Streaming:     true,
		ToolCalling:   true,
		Vision:        true,
		Reasoning:     true,
		EffortControl: false,
	}
}

func (p *AnthropicProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 200000, DefaultMaxTokens: 8192},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "Anthropic", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, ContextWindow: 200000, DefaultMaxTokens: 8192},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 200000, DefaultMaxTokens: 4096},
	}, nil
}

func (p *AnthropicProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Anthropic API Key is missing")
	}
	return &llm.CompletionResponse{
		Content: "Response from Anthropic Provider",
	}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, request llm.CompletionRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 10)
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{Type: llm.StreamEventContent, Content: "Streaming content from Anthropic..."}
		ch <- llm.StreamEvent{Type: llm.StreamEventDone}
	}()
	return ch, nil
}
