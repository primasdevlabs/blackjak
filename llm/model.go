package llm

// Model represents a language model specification and its capabilities.
type Model struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	SupportsTools     bool   `json:"supportsTools"`
	SupportsVision    bool   `json:"supportsVision"`
	SupportsReasoning bool   `json:"supportsReasoning"`
	ContextWindow     int    `json:"contextWindow"`
	DefaultMaxTokens  int    `json:"defaultMaxTokens"`
}

// ProviderCapabilities defines features supported by an LLM provider.
type ProviderCapabilities struct {
	Streaming     bool `json:"streaming"`
	ToolCalling   bool `json:"toolCalling"`
	Vision        bool `json:"vision"`
	Reasoning     bool `json:"reasoning"`
	EffortControl bool `json:"effortControl"`
}
