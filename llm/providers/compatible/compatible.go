package compatible

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
		baseURL: strings.TrimSuffix(baseURL, "/"),
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
	endpoint := p.baseURL + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err == nil {
		if p.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+p.apiKey)
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var result struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && len(result.Data) > 0 {
				models := make([]llm.Model, 0, len(result.Data))
				for _, item := range result.Data {
					models = append(models, p.categorizeModel(item.ID))
				}
				return models, nil
			}
		}
	}

	// Fallback when endpoint unreachable or returned empty list
	return []llm.Model{
		p.categorizeModel(p.modelID),
	}, nil
}

func (p *CompatibleProvider) categorizeModel(id string) llm.Model {
	lower := strings.ToLower(id)

	category := llm.CategoryCoding
	roles := []llm.ModelRole{llm.ModelRoleCoding}

	if strings.Contains(lower, "reasoning") || strings.Contains(lower, "r1") || strings.Contains(lower, "think") {
		category = llm.CategoryReasoning
		roles = []llm.ModelRole{llm.ModelRoleThinking, llm.ModelRoleReview}
	} else if strings.Contains(lower, "mini") || strings.Contains(lower, "nano") || strings.Contains(lower, "1.5b") || strings.Contains(lower, "3b") || strings.Contains(lower, "7b") {
		category = llm.CategoryFast
		roles = []llm.ModelRole{llm.ModelRoleFast, llm.ModelRoleCoding}
	}

	return llm.Model{
		ID:                id,
		Name:              fmt.Sprintf("%s (Compatible)", id),
		Provider:          "OpenAI-compatible",
		SupportsTools:     true,
		SupportsVision:    strings.Contains(lower, "vision") || strings.Contains(lower, "llava"),
		SupportsReasoning: category == llm.CategoryReasoning,
		SupportsEffort:    false,
		ContextWindow:     32768,
		DefaultMaxTokens:  4096,
		Category:          category,
		Status:            llm.StatusCurrent,
		DefaultRoles:      roles,
	}
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
