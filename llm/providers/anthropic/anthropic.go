package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// Ping verifies credentials with a real GET /models request.
func (p *AnthropicProvider) Ping(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("no API key configured — set ANTHROPIC_API_KEY or enter a key in Settings")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed (HTTP %d) — check the API key", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response: HTTP %d", resp.StatusCode)
	}
	return nil
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

// anthropicRequest mirrors the Anthropic /v1/messages request schema.
type anthropicRequest struct {
	Model     string             `json:"model"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
	MaxTokens int                `json:"max_tokens"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type      string          `json:"type"` // text | tool_use | tool_result
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

type anthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Anthropic API Key is missing")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("no model specified for Anthropic request")
	}

	out := anthropicRequest{
		Model:     request.Model,
		MaxTokens: request.MaxTokens,
	}
	if out.MaxTokens <= 0 {
		out.MaxTokens = 8192
	}

	var systemParts []string
	for _, m := range request.Messages {
		switch m.Role {
		case llm.RoleSystem:
			systemParts = append(systemParts, m.Content)
		case llm.RoleTool:
			// Anthropic expects tool results inside a user-role message.
			out.Messages = append(out.Messages, anthropicMessage{
				Role: "user",
				Content: []anthropicContent{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}},
			})
		default:
			am := anthropicMessage{Role: string(m.Role)}
			if m.Content != "" {
				am.Content = append(am.Content, anthropicContent{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				am.Content = append(am.Content, anthropicContent{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: json.RawMessage(tc.Arguments),
				})
			}
			out.Messages = append(out.Messages, am)
		}
	}
	out.System = strings.Join(systemParts, "\n\n")

	for _, t := range request.Tools {
		out.Tools = append(out.Tools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	body, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimSuffix(p.baseURL, "/")+"/messages", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := (&http.Client{Timeout: 180 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read anthropic response: %w", err)
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode anthropic response (HTTP %d): %s", resp.StatusCode, truncate(string(respBody), 400))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s", parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 400))
	}

	result := &llm.CompletionResponse{}
	var textParts []string
	for _, block := range parsed.Content {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(block.Input),
			})
		}
	}
	result.Content = strings.Join(textParts, "")
	return result, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
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
