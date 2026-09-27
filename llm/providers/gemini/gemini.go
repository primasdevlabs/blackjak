package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
		EffortControl: true,
	}
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	catalog := p.GetCatalog()
	if p.apiKey == "" {
		return catalog, nil
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return catalog, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return catalog, nil
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return catalog, nil
	}

	known := make(map[string]bool)
	for _, m := range catalog {
		known[m.ID] = true
	}

	merged := append([]llm.Model{}, catalog...)
	for _, fetched := range result.Models {
		cleanID := fetched.Name
		if len(cleanID) > 7 && cleanID[:7] == "models/" {
			cleanID = cleanID[7:]
		}
		if !known[cleanID] {
			merged = append(merged, llm.Model{
				ID:                cleanID,
				Name:              fetched.DisplayName,
				Provider:          "Google Gemini",
				SupportsTools:     true,
				SupportsVision:    true,
				SupportsReasoning: true,
				SupportsEffort:    true,
				ContextWindow:     1000000,
				DefaultMaxTokens:  8192,
				Category:          llm.CategoryCoding,
				Status:            llm.StatusCurrent,
				DefaultRoles:      []llm.ModelRole{llm.ModelRoleCoding},
			})
		}
	}
	return merged, nil
}

func (p *GeminiProvider) GetCatalog() []llm.Model {
	return []llm.Model{
		// Gemini 3.x
		{ID: "gemini-3.8-flash", Name: "Gemini 3.8 Flash", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 2000000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding, llm.ModelRoleThinking, llm.ModelRoleReview}},
		{ID: "gemini-3.7-flash", Name: "Gemini 3.7 Flash", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 2000000, DefaultMaxTokens: 16384, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gemini-3.6-flash", Name: "Gemini 3.6 Flash", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gemini-3.5-flash", Name: "Gemini 3.5 Flash", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gemini-3.5-flash-lite", Name: "Gemini 3.5 Flash-Lite", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "gemini-3.1-flash-lite", Name: "Gemini 3.1 Flash-Lite", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},
		{ID: "gemini-3.1-pro-preview", Name: "Gemini 3.1 Pro Preview", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 2000000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleReview}},
		{ID: "gemini-3-flash-preview", Name: "Gemini 3 Flash Preview", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},

		// Gemini 2.5
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, SupportsEffort: true, ContextWindow: 2000000, DefaultMaxTokens: 8192, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryCoding, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleCoding}},
		{ID: "gemini-2.5-flash-lite", Name: "Gemini 2.5 Flash-Lite", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: false, SupportsEffort: false, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryFast, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleFast}},

		// Agent / Research
		{ID: "gemini-computer-use", Name: "Gemini Computer Use", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 1000000, DefaultMaxTokens: 8192, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "gemini-deep-research", Name: "Gemini Deep Research", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 2000000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking}},
		{ID: "antigravity-agent", Name: "Antigravity Agent", Provider: "Google Gemini", SupportsTools: true, SupportsVision: true, SupportsReasoning: true, ContextWindow: 2000000, DefaultMaxTokens: 16384, Category: llm.CategoryReasoning, Status: llm.StatusCurrent, DefaultRoles: []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleCoding}},

		// Non-coding (Image, Audio, Embedding)
		{ID: "nano-banana-2", Name: "Nano Banana 2", Provider: "Google Gemini", Category: llm.CategoryImage, Status: llm.StatusCurrent},
		{ID: "gemini-3.8-live", Name: "Gemini 3.8 Live", Provider: "Google Gemini", Category: llm.CategoryAudio, Status: llm.StatusCurrent},
		{ID: "gemini-embedding-2", Name: "Gemini Embedding 2", Provider: "Google Gemini", Category: llm.CategoryEmbedding, Status: llm.StatusCurrent},
	}
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
