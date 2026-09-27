package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatCompletionConfig configures an OpenAI-compatible /chat/completions call.
type ChatCompletionConfig struct {
	BaseURL  string // e.g. https://api.openai.com/v1
	APIKey   string // optional for local servers (Ollama, LM Studio)
	OrgID    string // optional, OpenAI only
	Model    string
	Timeout  time.Duration
	Endpoint string // optional override; defaults to {BaseURL}/chat/completions
}

// openAIChatRequest mirrors the OpenAI chat completions request schema.
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIMessage     `json:"messages"`
	Tools       []openAITool        `json:"tools,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature *float64            `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Parameters  interface{} `json:"parameters"`
	} `json:"function"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// CompleteChatCompletion executes a real OpenAI-compatible chat completion
// request including tool definitions and tool_call parsing. Shared by the
// OpenAI provider and the OpenAI-compatible provider (Ollama, LM Studio,
// OpenRouter, vLLM, ...).
func CompleteChatCompletion(ctx context.Context, cfg ChatCompletionConfig, req *CompletionRequest) (*CompletionResponse, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	model := cfg.Model
	if model == "" {
		model = req.Model
	}
	if model == "" {
		return nil, fmt.Errorf("no model specified for chat completion")
	}

	out := openAIChatRequest{
		Model:     model,
		Messages:  make([]openAIMessage, 0, len(req.Messages)),
		MaxTokens: req.MaxTokens,
	}

	for _, m := range req.Messages {
		om := openAIMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			Name:       m.Name,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			var c openAIToolCall
			c.ID = tc.ID
			c.Type = "function"
			c.Function.Name = tc.Name
			c.Function.Arguments = tc.Arguments
			om.ToolCalls = append(om.ToolCalls, c)
		}
		out.Messages = append(out.Messages, om)
	}

	for _, t := range req.Tools {
		var ot openAITool
		ot.Type = "function"
		ot.Function.Name = t.Name
		ot.Function.Description = t.Description
		ot.Function.Parameters = t.Parameters
		out.Tools = append(out.Tools, ot)
	}

	body, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = strings.TrimSuffix(cfg.BaseURL, "/") + "/chat/completions"
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}
	if cfg.OrgID != "" {
		httpReq.Header.Set("OpenAI-Organization", cfg.OrgID)
	}

	resp, err := (&http.Client{Timeout: cfg.Timeout}).Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chat completion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read chat response: %w", err)
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode chat response (HTTP %d): %s", resp.StatusCode, truncate(string(respBody), 400))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("provider error: %s", parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat completion HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 400))
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("chat completion returned no choices")
	}

	result := &CompletionResponse{Content: parsed.Choices[0].Message.Content}
	for _, tc := range parsed.Choices[0].Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}
	return result, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
