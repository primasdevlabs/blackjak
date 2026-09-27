package compatible

import (
	"context"
	"fmt"

	"blackjak/llm"
)

type CompatibleProvider struct {
	baseURL string
	apiKey  string
	modelID string
}

func New(baseURL, apiKey, modelID string) *CompatibleProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	if modelID == "" {
		modelID = "llama3.2"
	}
	return &CompatibleProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		modelID: modelID,
	}
}

func (p *CompatibleProvider) Name() string {
	return "OpenAI-compatible"
}

func (p *CompatibleProvider) Capabilities() llm.ProviderCapabilities {
	return llm.ProviderCapabilities{
		Streaming:     true,
		ToolCalling:   true,
		Vision:        false,
		Reasoning:     false,
		EffortControl: false,
	}
}

func (p *CompatibleProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{
		{ID: p.modelID, Name: fmt.Sprintf("%s (Compatible)", p.modelID), Provider: "OpenAI-compatible", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, ContextWindow: 32768, DefaultMaxTokens: 4096},
	}, nil
}

func (p *CompatibleProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		Content: fmt.Sprintf("Response from OpenAI-compatible provider at %s using model %s", p.baseURL, p.modelID),
	}, nil
}

func (p *CompatibleProvider) Stream(ctx context.Context, request llm.CompletionRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 10)
	go func() {
		defer close(ch)
		ch <- llm.StreamEvent{Type: llm.StreamEventContent, Content: "Streaming content from OpenAI-compatible provider..."}
		ch <- llm.StreamEvent{Type: llm.StreamEventDone}
	}()
	return ch, nil
}
