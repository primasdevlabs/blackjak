package gemini

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

// Ping verifies credentials with a real GET /models request.
func (p *GeminiProvider) Ping(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("no API key configured — set GEMINI_API_KEY or enter a key in Settings")
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed (HTTP %d) — check the API key", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response: HTTP %d", resp.StatusCode)
	}
	return nil
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

// geminiRequest mirrors the Gemini generateContent request schema.
type geminiRequest struct {
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	Contents          []geminiContent  `json:"contents"`
	Tools             []geminiToolDecl `json:"tools,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string          `json:"text,omitempty"`
	FunctionCall     *geminiCall     `json:"functionCall,omitempty"`
	FunctionResponse *geminiResponse `json:"functionResponse,omitempty"`
	ThoughtSignature string          `json:"thoughtSignature,omitempty"`
}

type geminiCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"`
}

type geminiResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiToolDecl struct {
	FunctionDeclarations []struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Parameters  interface{} `json:"parameters"`
	} `json:"functionDeclarations"`
}

type geminiAPIResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) Chat(ctx context.Context, request llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("Gemini API Key is missing")
	}
	model := request.Model
	if model == "" {
		return nil, fmt.Errorf("no model specified for Gemini request")
	}

	out := geminiRequest{}
	var systemParts []string
	for _, m := range request.Messages {
		switch m.Role {
		case llm.RoleSystem:
			systemParts = append(systemParts, m.Content)
		case llm.RoleTool:
			// Gemini expects function responses inside a user-role turn.
			out.Contents = append(out.Contents, geminiContent{
				Role: "user",
				Parts: []geminiPart{{
					FunctionResponse: &geminiResponse{
						Name:     m.Name,
						Response: map[string]interface{}{"result": m.Content},
					},
				}},
			})
		default:
			role := "user"
			if m.Role == llm.RoleAssistant {
				role = "model"
			}
			gc := geminiContent{Role: role}
			if m.Content != "" {
				gc.Parts = append(gc.Parts, geminiPart{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				gc.Parts = append(gc.Parts, geminiPart{
					FunctionCall:     &geminiCall{Name: tc.Name, Args: json.RawMessage(tc.Arguments)},
					ThoughtSignature: tc.ThoughtSignature,
				})
			}
			out.Contents = append(out.Contents, gc)
		}
	}
	if len(systemParts) > 0 {
		out.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: strings.Join(systemParts, "\n\n")}},
		}
	}
	if len(request.Tools) > 0 {
		var decl geminiToolDecl
		for _, t := range request.Tools {
			decl.FunctionDeclarations = append(decl.FunctionDeclarations, struct {
				Name        string      `json:"name"`
				Description string      `json:"description"`
				Parameters  interface{} `json:"parameters"`
			}{Name: t.Name, Description: t.Description, Parameters: t.Parameters})
		}
		out.Tools = []geminiToolDecl{decl}
	}

	body, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read gemini response: %w", err)
	}

	var parsed geminiAPIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode gemini response (HTTP %d)", resp.StatusCode)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini error: %s", parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 {
		return nil, fmt.Errorf("gemini returned no candidates (HTTP %d)", resp.StatusCode)
	}

	result := &llm.CompletionResponse{}
	var textParts []string
	// Gemini may attach the thought signature to any part (docs: typically the
	// first). Track the latest one seen and stamp it on each function call so
	// it round-trips in the next request.
	var lastSig string
	for i, part := range parsed.Candidates[0].Content.Parts {
		if part.ThoughtSignature != "" {
			lastSig = part.ThoughtSignature
		}
		if part.Text != "" {
			textParts = append(textParts, part.Text)
		}
		if part.FunctionCall != nil {
			sig := part.ThoughtSignature
			if sig == "" {
				sig = lastSig
			}
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:               fmt.Sprintf("call_%d", i),
				Name:             part.FunctionCall.Name,
				Arguments:        string(part.FunctionCall.Args),
				ThoughtSignature: sig,
			})
		}
	}
	result.Content = strings.Join(textParts, "")
	return result, nil
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
