package openai

import (
	"context"
	"fmt"

	"blackjak/llm"
)

type OpenAIProvider struct {
	apiKey  string
	baseURL string
	orgID   string
}

func New(apiKey, baseURL, orgID string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		orgID:   orgID,
	}
}

func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

func (p *OpenAIProvider) Capabilities() llm.ProviderCapabilities {
	return llm.ProviderCapabilities{
		Streaming:     true,
		ToolCalling:   true,
		Vision:        true,
		Reasoning:     true,
		EffortControl: true,
	}
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 128000, DefaultMaxTokens: 4096},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, ContextWindow: 128000, DefaultMaxTokens: 4096},
		{ID: "o1-preview", Name: "o1-preview", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: true, ContextWindow: 128000, DefaultMaxTokens: 32768},
		{ID: "o3-mini", Name: "o3-mini", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: true, ContextWindow: 200000, DefaultMaxTokens: 100000},
	}, nil
}

func (p *OpenAIProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API Key is missing")
	}
	// Return structured response
	return &llm.CompletionResponse{
		Content: "Response from OpenAI Provider",
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, request llm.CompletionRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 10)
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{Type: llm.StreamEventContent, Content: "Streaming content from OpenAI..."}
		ch <- llm.StreamEvent{Type: llm.StreamEventDone}
	}()
	return ch, nil
}
