package llm

type ModelRole string

const (
	ModelRoleThinking ModelRole = "thinking"
	ModelRoleCoding   ModelRole = "coding"
	ModelRoleFast     ModelRole = "fast"
	ModelRoleReview   ModelRole = "review"
)

type ModelCategory string

const (
	CategoryCoding     ModelCategory = "coding"
	CategoryReasoning  ModelCategory = "reasoning"
	CategoryFast       ModelCategory = "fast"
	CategoryRealtime   ModelCategory = "realtime"
	CategoryAudio      ModelCategory = "audio"
	CategoryImage      ModelCategory = "image"
	CategoryEmbedding  ModelCategory = "embedding"
	CategoryOpenWeight ModelCategory = "openweight"
)

type ModelStatus string

const (
	StatusCurrent    ModelStatus = "current"
	StatusDeprecated ModelStatus = "deprecated"
	StatusRetired    ModelStatus = "retired"
)

type EffortLevel string

const (
	EffortLow       EffortLevel = "low"
	EffortMedium    EffortLevel = "medium"
	EffortHigh      EffortLevel = "high"
	EffortExtraHigh EffortLevel = "extra_high"
)

// Model represents a language model specification and its capabilities.
type Model struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Provider          string        `json:"provider"`
	SupportsTools     bool          `json:"supportsTools"`
	SupportsVision    bool          `json:"supportsVision"`
	SupportsReasoning bool          `json:"supportsReasoning"`
	SupportsEffort    bool          `json:"supportsEffort"`
	ContextWindow     int           `json:"contextWindow"`
	DefaultMaxTokens  int           `json:"defaultMaxTokens"`
	Category          ModelCategory `json:"category"`
	Status            ModelStatus   `json:"status"`
	DefaultRoles      []ModelRole   `json:"defaultRoles"`
}

// ProviderCapabilities defines features supported by an LLM provider.
type ProviderCapabilities struct {
	Streaming     bool `json:"streaming"`
	ToolCalling   bool `json:"toolCalling"`
	Vision        bool `json:"vision"`
	Reasoning     bool `json:"reasoning"`
	EffortControl bool `json:"effortControl"`
}

// ModelConfig represents provider + model selection for a specific agent role.
type ModelConfig struct {
	ProviderID string    `json:"providerId"`
	ModelID    string    `json:"modelId"`
	Role       ModelRole `json:"role"`
}

// AgentModelSettings defines settings for thinking, coding, fast, and review models.
type AgentModelSettings struct {
	Thinking ModelConfig `json:"thinking"`
	Coding   ModelConfig `json:"coding"`
	Fast     ModelConfig `json:"fast"`
	Review   ModelConfig `json:"review"`

	Effort EffortLevel `json:"effort"`
}

