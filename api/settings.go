package api

import (
	"context"
	"fmt"
	"os"
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
	EffortMinimal EffortLevel = "minimal"
	EffortLow     EffortLevel = "low"
	EffortMedium  EffortLevel = "medium"
	EffortHigh    EffortLevel = "high"
	EffortMaximum EffortLevel = "maximum"
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
	ActiveProvider     string                         `json:"activeProvider"`
	Providers          map[string]ProviderCredentials `json:"providers"`
	ThinkingModelID    string                         `json:"thinkingModelId"`
	CodingModelID      string                         `json:"codingModelId"`
	FastModelID        string                         `json:"fastModelId"`
	UseSeparateModels  bool                           `json:"useSeparateModels"`
	Effort             EffortLevel                    `json:"effort"`
	Mode               AgentMode                      `json:"mode"`
	ParallelSubagents  bool                           `json:"parallelSubagents"`
	MaxSubagents       int                            `json:"maxSubagents"`
	PromptQueueBehavior string                        `json:"promptQueueBehavior"` // "sequential", "parallel", "ask"
	AutoOpenFile       bool                           `json:"autoOpenFile"`
	AutoOpenDiff       bool                           `json:"autoOpenDiff"`
	AskDestructiveOps  bool                           `json:"askDestructiveOps"`
	ModelRoutes        map[string]string              `json:"modelRoutes"` // task -> role
}

// SettingsManager thread-safely manages system settings, providers, and masked credentials.
type SettingsManager struct {
	mu            sync.RWMutex
	config        SettingsConfig
	activeProvider llm.Provider
	router        *llm.ModelRouter
}

// NewSettingsManager initializes SettingsManager.
func NewSettingsManager() *SettingsManager {
	cfg := SettingsConfig{
		ActiveProvider:    "OpenAI",
		ThinkingModelID:   "gpt-4o",
		CodingModelID:     "claude-3-5-sonnet-20241022",
		FastModelID:       "gpt-4o-mini",
		UseSeparateModels: true,
		Effort:            EffortMedium,
		Mode:              ModeCode,
		ParallelSubagents: true,
		MaxSubagents:      4,
		PromptQueueBehavior: "sequential",
		AutoOpenFile:      true,
		AutoOpenDiff:      true,
		AskDestructiveOps: true,
		Providers: map[string]ProviderCredentials{
			"OpenAI":            {BaseURL: "https://api.openai.com/v1", StorageMode: StorageEnvironment},
			"Google Gemini":     {StorageMode: StorageEnvironment},
			"Anthropic":         {BaseURL: "https://api.anthropic.com/v1", StorageMode: StorageEnvironment},
			"OpenAI-compatible": {BaseURL: "http://localhost:11434/v1", ModelID: "llama3.2", StorageMode: StorageStored},
		},
		ModelRoutes: map[string]string{
			"planning":    "thinking",
			"exploration": "fast",
			"coding":      "coding",
			"debugging":   "thinking",
			"testing":     "coding",
			"review":      "thinking",
			"summarization": "fast",
		},
	}

	sm := &SettingsManager{
		config: cfg,
		router: llm.NewModelRouter(llm.DefaultRouterConfig()),
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
	sm.config.ThinkingModelID = updates.ThinkingModelID
	sm.config.CodingModelID = updates.CodingModelID
	sm.config.FastModelID = updates.FastModelID
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

	sm.initActiveProvider()
	sm.mu.Unlock()
}

// TestProvider Connection helper
func (sm *SettingsManager) TestProvider(providerID string) (bool, string, []llm.Model) {
	sm.mu.RLock()
	cred, exists := sm.config.Providers[providerID]
	sm.mu.RUnlock()

	if !exists {
		return false, fmt.Sprintf("Provider %s not configured", providerID), nil
	}

	var prov llm.Provider
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
		prov = gemini.New(apiKey)
	case "Anthropic":
		prov = anthropic.New(apiKey, cred.BaseURL)
	case "OpenAI-compatible":
		prov = compatible.New(cred.BaseURL, apiKey, cred.ModelID)
	default:
		prov = openai.New(apiKey, cred.BaseURL, cred.OrgID)
	}

	models, err := prov.ListModels(context.Background())
	if err != nil {
		return false, err.Error(), nil
	}

	return true, fmt.Sprintf("Successfully connected to %s (%d models available)", providerID, len(models)), models
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
