package gemini

import (
	"context"
	"fmt"

	"blackjak/llm"
)

type GeminiProvider struct {
	apiKey string
}

func New(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
	}
}

func (p *GeminiProvider) Name() string {
	return "Google Gemini"
}

func (p *GeminiProvider) Capabilities() llm.ProviderCapabilities {
	return llm.ProviderCapabilities{
		Streaming:     true,
		ToolCalling:   true,
		Vision:        true,
		Reasoning:     true,
		EffortControl: false,
	}
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{
		{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 2000000, DefaultMaxTokens: 8192},
		{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, ContextWindow: 1000000, DefaultMaxTokens: 8192},
	}, nil
}

func (p *GeminiProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Gemini API Key is missing")
	}
	return &llm.CompletionResponse{
		Content: "Response from Gemini Provider",
	}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, request llm.CompletionRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 10)
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{Type: llm.StreamEventContent, Content: "Streaming content from Gemini..."}
		ch <- llm.StreamEvent{Type: llm.StreamEventDone}
	}()
	return ch, nil
}
