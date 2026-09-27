package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	catalog := p.GetCatalog()

	if p.apiKey == "" {
		return catalog, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return catalog, nil
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	if p.orgID != "" {
		req.Header.Set("OpenAI-Organization", p.orgID)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return catalog, nil
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return catalog, nil
	}

	// Merge fetched models with catalog
	known := make(map[string]bool)
	for _, m := range catalog {
		known[m.ID] = true
	}

	merged := append([]llm.Model{}, catalog...)
	for _, fetched := range result.Data {
		if !known[fetched.ID] {
			merged = append(merged, llm.Model{
				ID:                fetched.ID,
				Name:              fetched.ID,
				Provider:          "OpenAI",
				SupportsTools:     true,
				SupportsVision:    true,
				SupportsReasoning: false,
				SupportsEffort:    false,
				ContextWindow:     128000,
				DefaultMaxTokens:  4096,
				Category:          llm.CategoryCoding,
				Status:            llm.StatusCurrent,
				DefaultRoles:      []llm.ModelRole{llm.ModelRoleCoding},
			})
		}
	}
	return merged, nil
}

func (p *OpenAIProvider) GetCatalog() []llm.Model {
	return []llm.Model{
		// Flagship
		{ID: "gpt-6-astra", Name: "GPT-6 Astra", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 500000, DefaultMaxTokens: 32768, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleReview}},
		{ID: "gpt-6-sol", Name: "GPT-6 Sol", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 500000, DefaultMaxTokens: 32768, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding, llm.ModelRoleThinking}},
		{ID: "gpt-6-luna", Name: "GPT-6 Luna", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 200000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},

		// Professional / Coding
		{ID: "gpt-5.6-sol", Name: "GPT-5.6 Sol", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5.6-terra", Name: "GPT-5.6 Terra", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5.6-luna", Name: "GPT-5.6 Luna", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 256000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "gpt-5.5-pro", Name: "GPT-5.5 Pro", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleReview}},
		{ID: "gpt-5.5", Name: "GPT-5.5", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5.4-pro", Name: "GPT-5.4 Pro", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "gpt-5.4", Name: "GPT-5.4", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 256000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5.4-mini", Name: "GPT-5.4 Mini", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 128000, DefaultMaxTokens: 4096, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast, llm.ModelRoleCoding}},
		{ID: "gpt-5.4-nano", Name: "GPT-5.4 Nano", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 128000, DefaultMaxTokens: 4096, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "gpt-5.3-codex", Name: "GPT-5.3-Codex", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5.2-pro", Name: "GPT-5.2 Pro", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "gpt-5.2", Name: "GPT-5.2", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 256000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5.1", Name: "GPT-5.1", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 256000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5-pro", Name: "GPT-5 Pro", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 256000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "gpt-5", Name: "GPT-5", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 256000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gpt-5-mini", Name: "GPT-5 Mini", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 128000, DefaultMaxTokens: 4096, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "gpt-5-nano", Name: "GPT-5 Nano", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 128000, DefaultMaxTokens: 4096, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 128000, DefaultMaxTokens: 4096, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding, llm.ModelRoleReview}},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 128000, DefaultMaxTokens: 4096, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},

		// Reasoning
		{ID: "o3", Name: "o3", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 100000, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "o3-pro", Name: "o3-pro", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 200000, DefaultMaxTokens: 100000, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "o1", Name: "o1", Provider: "OpenAI", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 128000, DefaultMaxTokens: 32768, Category: llm.CategoryReasoning, Status: llm.StatusDeprecated, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "o1-mini", Name: "o1-mini", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 128000, DefaultMaxTokens: 32768, Category: llm.CategoryFast, Status: llm.StatusDeprecated, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},

		// Open-weight
		{ID: "gpt-oss-120b", Name: "GPT OSS 120B", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 128000, DefaultMaxTokens: 8192, Category: llm.CategoryOpenWeight, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding, llm.ModelRoleThinking}},
		{ID: "gpt-oss-20b", Name: "GPT OSS 20B", Provider: "OpenAI", SupportsTools: true, SupportsVision: false, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 64000, DefaultMaxTokens: 4096, Category: llm.CategoryOpenWeight, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast, llm.ModelRoleCoding}},

		// Non-coding (Embedding, Audio, Image)
		{ID: "text-embedding-3-large", Name: "Text Embedding 3 Large", Provider: "OpenAI", Category: llm.CategoryEmbedding, Status: llm.StatusCurrent},
		{ID: "text-embedding-3-small", Name: "Text Embedding 3 Small", Provider: "OpenAI", Category: llm.CategoryEmbedding, Status: llm.StatusCurrent},
		{ID: "gpt-realtime-2.1", Name: "GPT-Realtime 2.1", Provider: "OpenAI", Category: llm.CategoryRealtime, Status: llm.StatusCurrent},
		{ID: "gpt-audio-1.5", Name: "GPT Audio 1.5", Provider: "OpenAI", Category: llm.CategoryAudio, Status: llm.StatusCurrent},
		{ID: "gpt-image-2.5-sunburst", Name: "GPT-Image 2.5 Sunburst", Provider: "OpenAI", Category: llm.CategoryImage, Status: llm.StatusCurrent},
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API Key is missing")
	}
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
