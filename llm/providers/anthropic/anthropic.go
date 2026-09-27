package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
		EffortControl: true,
	}
}

func (p *AnthropicProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	catalog := p.GetCatalog()

	if p.apiKey == "" {
		return catalog, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return catalog, nil
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return catalog, nil
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return catalog, nil
	}

	known := make(map[string]bool)
	for _, m := range catalog {
		known[m.ID] = true
	}

	merged := append([]llm.Model{}, catalog...)
	for _, fetched := range result.Data {
		if !known[fetched.ID] {
			name := fetched.DisplayName
			if name == "" {
				name = fetched.ID
			}
			merged = append(merged, llm.Model{
				ID:                fetched.ID,
				Name:              name,
				Provider:          "Anthropic",
				SupportsTools:     true,
				SupportsVision:    true,
				SupportsReasoning: true,
				SupportsEffort:    true,
				ContextWindow:     200000,
				DefaultMaxTokens:  8192,
				Category:          llm.CategoryCoding,
				Status:            llm.StatusCurrent,
				DefaultRoles:      []llm.ModelRole{llm.ModelRoleCoding},
			})
		}
	}
	return merged, nil
}

func (p *AnthropicProvider) GetCatalog() []llm.Model {
	return []llm.Model{
		// High-end
		{ID: "claude-fable-5.1", Name: "Claude Fable 5.1", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 500000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleReview}},
		{ID: "claude-fable-5", Name: "Claude Fable 5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 500000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-mythos-5.1", Name: "Claude Mythos 5.1", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 500000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-mythos-5", Name: "Claude Mythos 5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 500000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-opus-5", Name: "Claude Opus 5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 300000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleReview}},
		{ID: "claude-sonnet-5", Name: "Claude Sonnet 5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 300000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding, llm.ModelRoleReview}},

		// Claude 4.x
		{ID: "claude-opus-4.8", Name: "Claude Opus 4.8", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-opus-4.7", Name: "Claude Opus 4.7", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-opus-4.6", Name: "Claude Opus 4.6", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-opus-4.5", Name: "Claude Opus 4.5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-sonnet-4.6", Name: "Claude Sonnet 4.6", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "claude-sonnet-4.5", Name: "Claude Sonnet 4.5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "claude-sonnet-4", Name: "Claude Sonnet 4", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "claude-haiku-4.5", Name: "Claude Haiku 4.5", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},

		// Legacy / 3.x
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "Anthropic", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "claude-opus-4.1", Name: "Claude Opus 4.1", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 4096, Category: llm.CategoryReasoning, Status: llm.StatusDeprecated, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "claude-sonnet-3.7", Name: "Claude Sonnet 3.7", Provider: "Anthropic", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 4096, Category: llm.CategoryCoding, Status: llm.StatusDeprecated, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "claude-haiku-3.5", Name: "Claude Haiku 3.5", Provider: "Anthropic", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 4096, Category: llm.CategoryFast, Status: llm.StatusDeprecated, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
	}
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
