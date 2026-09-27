package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"blackjak/llm"
	"blackjak/llm/providers/anthropic"
	"blackjak/llm/providers/compatible"
	"blackjak/llm/providers/gemini"
	"blackjak/llm/providers/openai"
)

// StorageMode defines how credentials are persisted.
type StorageMode string

const (
	StorageEnvironment StorageMode = "environment"
	StorageStored      StorageMode = "stored"
	StorageSession     StorageMode = "session"
)

// EffortLevel defines abstract reasoning depth.
type EffortLevel string

const (
	EffortLow       EffortLevel = "low"
	EffortMedium    EffortLevel = "medium"
	EffortHigh      EffortLevel = "high"
	EffortExtraHigh EffortLevel = "extra_high"
)

// AgentMode defines operating mode.
type AgentMode string

const (
	ModePlan AgentMode = "plan"
	ModeCode AgentMode = "code"
)

// ProviderCredentials holds provider connection settings.
type ProviderCredentials struct {
	ApiKey      string      `json:"apiKey,omitempty"`
	BaseURL     string      `json:"baseUrl,omitempty"`
	OrgID       string      `json:"orgId,omitempty"`
	ModelID     string      `json:"modelId,omitempty"`
	StorageMode StorageMode `json:"storageMode"`
}

// SettingsConfig holds all runtime and provider configuration.
type SettingsConfig struct {
	ActiveProvider      string                         `json:"activeProvider"`
	Providers           map[string]ProviderCredentials `json:"providers"`
	Thinking            llm.ModelConfig                `json:"thinking"`
	Coding              llm.ModelConfig                `json:"coding"`
	Fast                llm.ModelConfig                `json:"fast"`
	Review              llm.ModelConfig                `json:"review"`
	ThinkingModelID     string                         `json:"thinkingModelId"`
	CodingModelID       string                         `json:"codingModelId"`
	FastModelID         string                         `json:"fastModelId"`
	ReviewModelID       string                         `json:"reviewModelId"`
	UseSeparateModels   bool                           `json:"useSeparateModels"`
	Effort              EffortLevel                    `json:"effort"`
	Mode                AgentMode                      `json:"mode"`
	ParallelSubagents   bool                           `json:"parallelSubagents"`
	MaxSubagents        int                            `json:"maxSubagents"`
	PromptQueueBehavior string                         `json:"promptQueueBehavior"` // "sequential", "parallel", "ask"
	AutoOpenFile        bool                           `json:"autoOpenFile"`
	AutoOpenDiff        bool                           `json:"autoOpenDiff"`
	AskDestructiveOps   bool                           `json:"askDestructiveOps"`
	ModelRoutes         map[string]string              `json:"modelRoutes"` // task -> role
}

// SettingsManager thread-safely manages system settings, providers, and masked credentials.
type SettingsManager struct {
	mu             sync.RWMutex
	config         SettingsConfig
	activeProvider llm.Provider
	router         *llm.ModelRouter
	providers      map[string]llm.Provider
	filePath       string
}

// NewSettingsManager initializes SettingsManager.
func NewSettingsManager() *SettingsManager {
	cfg := SettingsConfig{
		ActiveProvider: "OpenAI",
		Thinking: llm.ModelConfig{
			ProviderID: "Anthropic",
			ModelID:    "claude-opus-5",
			Role:       llm.ModelRoleThinking,
		},
		Coding: llm.ModelConfig{
			ProviderID: "OpenAI",
			ModelID:    "gpt-5.3-codex",
			Role:       llm.ModelRoleCoding,
		},
		Fast: llm.ModelConfig{
			ProviderID: "Google Gemini",
			ModelID:    "gemini-3.5-flash-lite",
			Role:       llm.ModelRoleFast,
		},
		Review: llm.ModelConfig{
			ProviderID: "Anthropic",
			ModelID:    "claude-sonnet-5",
			Role:       llm.ModelRoleReview,
		},
		ThinkingModelID:     "claude-opus-5",
		CodingModelID:       "gpt-5.3-codex",
		FastModelID:         "gemini-3.5-flash-lite",
		ReviewModelID:       "claude-sonnet-5",
		UseSeparateModels:   true,
		Effort:              EffortMedium,
		Mode:                ModeCode,
		ParallelSubagents:   true,
		MaxSubagents:        4,
		PromptQueueBehavior: "sequential",
		AutoOpenFile:        true,
		AutoOpenDiff:        true,
		AskDestructiveOps:   true,
		Providers: map[string]ProviderCredentials{
			"OpenAI":            {BaseURL: "https://api.openai.com/v1", StorageMode: StorageEnvironment},
			"Google Gemini":     {StorageMode: StorageEnvironment},
			"Anthropic":         {BaseURL: "https://api.anthropic.com/v1", StorageMode: StorageEnvironment},
			"OpenAI-compatible": {BaseURL: "http://localhost:11434/v1", ModelID: "llama3.2", StorageMode: StorageStored},
		},
		ModelRoutes: map[string]string{
			"planning":      "thinking",
			"exploration":   "fast",
			"coding":        "coding",
			"debugging":     "thinking",
			"testing":       "coding",
			"review":        "review",
			"summarization": "fast",
		},
	}

	sm := &SettingsManager{
		config:    cfg,
		router:    llm.NewModelRouter(llm.DefaultRouterConfig()),
		providers: make(map[string]llm.Provider),
	}
	sm.initActiveProvider()
	return sm
}

func (sm *SettingsManager) initActiveProvider() {
	pName := sm.config.ActiveProvider
	cred := sm.config.Providers[pName]

	apiKey := cred.ApiKey
	if cred.StorageMode == StorageEnvironment || apiKey == "" {
		switch pName {
		case "OpenAI":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "Google Gemini":
			apiKey = os.Getenv("GEMINI_API_KEY")
		case "Anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	switch pName {
	case "Google Gemini":
		sm.activeProvider = gemini.New(apiKey)
	case "Anthropic":
		sm.activeProvider = anthropic.New(apiKey, cred.BaseURL)
	case "OpenAI-compatible":
		sm.activeProvider = compatible.New(cred.BaseURL, apiKey, cred.ModelID)
	default:
		sm.activeProvider = openai.New(apiKey, cred.BaseURL, cred.OrgID)
	}

	if models, err := sm.activeProvider.ListModels(context.Background()); err == nil {
		sm.router.RegisterModels(models)
	}
}

// SetStoragePath points the manager at a JSON file used for persistence.
// If the file exists, it is loaded over the defaults so missing fields keep
// their default values.
func (sm *SettingsManager) SetStoragePath(path string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.filePath = path

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	loaded := sm.config
	if err := json.Unmarshal(data, &loaded); err != nil {
		return
	}
	sm.config = loaded
	sm.initActiveProvider()
}

// saveLocked persists the current config. Callers must hold sm.mu.
// Credentials marked as session-only are stripped before writing.
func (sm *SettingsManager) saveLocked() {
	if sm.filePath == "" {
		return
	}
	cfg := sm.config
	providers := make(map[string]ProviderCredentials, len(cfg.Providers))
	for name, cred := range cfg.Providers {
		if cred.StorageMode == StorageSession {
			cred.ApiKey = ""
		}
		providers[name] = cred
	}
	cfg.Providers = providers

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(sm.filePath), 0o755); err != nil {
		return
	}
	tmp := sm.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, sm.filePath)
}

// GetMaskedConfig returns a thread-safe copy of configuration with API keys masked.
func (sm *SettingsManager) GetMaskedConfig() SettingsConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	cp := sm.config
	cp.Providers = make(map[string]ProviderCredentials)
	for k, v := range sm.config.Providers {
		cp.Providers[k] = ProviderCredentials{
			ApiKey:      maskKey(v.ApiKey),
			BaseURL:     v.BaseURL,
			OrgID:       v.OrgID,
			ModelID:     v.ModelID,
			StorageMode: v.StorageMode,
		}
	}
	return cp
}

// GetConfig returns the unmasked current settings configuration.
func (sm *SettingsManager) GetConfig() SettingsConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.config
}

// ApplyPatch merges only the fields present in the given JSON patch into the
// config, so partial updates cannot clobber unrelated settings with zero values.
func (sm *SettingsManager) ApplyPatch(patch map[string]json.RawMessage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	str := func(key string, dst *string) {
		if raw, ok := patch[key]; ok {
			var v string
			if json.Unmarshal(raw, &v) == nil && v != "" {
				*dst = v
			}
		}
	}
	boolean := func(key string, dst *bool) {
		if raw, ok := patch[key]; ok {
			var v bool
			if json.Unmarshal(raw, &v) == nil {
				*dst = v
			}
		}
	}
	num := func(key string, dst *int) {
		if raw, ok := patch[key]; ok {
			var v int
			if json.Unmarshal(raw, &v) == nil && v > 0 {
				*dst = v
			}
		}
	}
	model := func(key string, idKey string, dst *llm.ModelConfig, dstID *string) {
		if raw, ok := patch[key]; ok {
			var v llm.ModelConfig
			if json.Unmarshal(raw, &v) == nil && v.ModelID != "" {
				*dst = v
				*dstID = v.ModelID
			}
		}
		if raw, ok := patch[idKey]; ok {
			var v string
			if json.Unmarshal(raw, &v) == nil && v != "" {
				*dstID = v
				dst.ModelID = v
			}
		}
	}

	str("activeProvider", &sm.config.ActiveProvider)

	if raw, ok := patch["providers"]; ok {
		var v map[string]ProviderCredentials
		if json.Unmarshal(raw, &v) == nil {
			for name, newCred := range v {
				if old, exists := sm.config.Providers[name]; exists && strings.Contains(newCred.ApiKey, "••••") {
					newCred.ApiKey = old.ApiKey
				}
				sm.config.Providers[name] = newCred
			}
		}
	}

	model("thinking", "thinkingModelId", &sm.config.Thinking, &sm.config.ThinkingModelID)
	model("coding", "codingModelId", &sm.config.Coding, &sm.config.CodingModelID)
	model("fast", "fastModelId", &sm.config.Fast, &sm.config.FastModelID)
	model("review", "reviewModelId", &sm.config.Review, &sm.config.ReviewModelID)

	boolean("useSeparateModels", &sm.config.UseSeparateModels)
	str("effort", (*string)(&sm.config.Effort))
	str("mode", (*string)(&sm.config.Mode))
	boolean("parallelSubagents", &sm.config.ParallelSubagents)
	num("maxSubagents", &sm.config.MaxSubagents)
	str("promptQueueBehavior", &sm.config.PromptQueueBehavior)
	boolean("autoOpenFile", &sm.config.AutoOpenFile)
	boolean("autoOpenDiff", &sm.config.AutoOpenDiff)
	boolean("askDestructiveOps", &sm.config.AskDestructiveOps)

	if raw, ok := patch["modelRoutes"]; ok {
		var v map[string]string
		if json.Unmarshal(raw, &v) == nil && v != nil {
			sm.config.ModelRoutes = v
		}
	}

	sm.saveLocked()
	sm.initActiveProvider()
}

// UpdateConfig updates settings and re-initializes the active provider.
func (sm *SettingsManager) UpdateConfig(updates SettingsConfig) {
	sm.mu.Lock()

	// Retain unmasked API keys if UI sent masked value
	for name, newCred := range updates.Providers {
		oldCred, exists := sm.config.Providers[name]
		if exists && strings.Contains(newCred.ApiKey, "••••") {
			newCred.ApiKey = oldCred.ApiKey
		}
		sm.config.Providers[name] = newCred
	}

	if updates.ActiveProvider != "" {
		sm.config.ActiveProvider = updates.ActiveProvider
	}

	if updates.Thinking.ModelID != "" {
		sm.config.Thinking = updates.Thinking
		sm.config.ThinkingModelID = updates.Thinking.ModelID
	} else if updates.ThinkingModelID != "" {
		sm.config.ThinkingModelID = updates.ThinkingModelID
		sm.config.Thinking.ModelID = updates.ThinkingModelID
	}

	if updates.Coding.ModelID != "" {
		sm.config.Coding = updates.Coding
		sm.config.CodingModelID = updates.Coding.ModelID
	} else if updates.CodingModelID != "" {
		sm.config.CodingModelID = updates.CodingModelID
		sm.config.Coding.ModelID = updates.CodingModelID
	}

	if updates.Fast.ModelID != "" {
		sm.config.Fast = updates.Fast
		sm.config.FastModelID = updates.Fast.ModelID
	} else if updates.FastModelID != "" {
		sm.config.FastModelID = updates.FastModelID
		sm.config.Fast.ModelID = updates.FastModelID
	}

	if updates.Review.ModelID != "" {
		sm.config.Review = updates.Review
		sm.config.ReviewModelID = updates.Review.ModelID
	} else if updates.ReviewModelID != "" {
		sm.config.ReviewModelID = updates.ReviewModelID
		sm.config.Review.ModelID = updates.ReviewModelID
	}

	sm.config.UseSeparateModels = updates.UseSeparateModels
	sm.config.Effort = updates.Effort
	sm.config.Mode = updates.Mode
	sm.config.ParallelSubagents = updates.ParallelSubagents
	sm.config.MaxSubagents = updates.MaxSubagents
	sm.config.PromptQueueBehavior = updates.PromptQueueBehavior
	sm.config.AutoOpenFile = updates.AutoOpenFile
	sm.config.AutoOpenDiff = updates.AutoOpenDiff
	sm.config.AskDestructiveOps = updates.AskDestructiveOps
	if updates.ModelRoutes != nil {
		sm.config.ModelRoutes = updates.ModelRoutes
	}

	sm.saveLocked()
	sm.initActiveProvider()
	sm.mu.Unlock()
}

// TestProvider performs a real connectivity + credential check via Ping.
func (sm *SettingsManager) TestProvider(providerID string) (bool, string, []llm.Model) {
	prov := sm.createProviderInstance(providerID)
	if prov == nil {
		return false, fmt.Sprintf("Provider %s not configured", providerID), nil
	}

	if err := prov.Ping(context.Background()); err != nil {
		return false, err.Error(), nil
	}

	// Ping passed — fetch the model list (falls back to the static catalog if
	// the provider can't enumerate).
	models, _ := prov.ListModels(context.Background())
	return true, fmt.Sprintf("Connected to %s (%d models available)", providerID, len(models)), models
}

// RefreshModels fetches the latest available models for a given provider or all providers.
func (sm *SettingsManager) RefreshModels(providerID string) map[string][]llm.Model {
	result := make(map[string][]llm.Model)
	providers := []string{"OpenAI", "Google Gemini", "Anthropic", "OpenAI-compatible"}

	if providerID != "" {
		providers = []string{providerID}
	}

	for _, p := range providers {
		prov := sm.createProviderInstance(p)
		if prov != nil {
			if models, err := prov.ListModels(context.Background()); err == nil {
				result[p] = models
				sm.router.RegisterModels(models)
			}
		}
	}
	return result
}

func (sm *SettingsManager) createProviderInstance(providerID string) llm.Provider {
	sm.mu.RLock()
	cred, exists := sm.config.Providers[providerID]
	sm.mu.RUnlock()

	if !exists {
		return nil
	}

	apiKey := cred.ApiKey
	if apiKey == "" || cred.StorageMode == StorageEnvironment {
		switch providerID {
		case "OpenAI":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "Google Gemini":
			apiKey = os.Getenv("GEMINI_API_KEY")
		case "Anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	switch providerID {
	case "Google Gemini":
		return gemini.New(apiKey)
	case "Anthropic":
		return anthropic.New(apiKey, cred.BaseURL)
	case "OpenAI-compatible":
		return compatible.New(cred.BaseURL, apiKey, cred.ModelID)
	default:
		return openai.New(apiKey, cred.BaseURL, cred.OrgID)
	}
}

func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "••••••••"
	}
	return key[:3] + "••••••••••••" + key[len(key)-4:]
}
