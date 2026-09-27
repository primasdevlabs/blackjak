package api

import (
	"strings"
	"testing"
)

func TestSettingsManager_MaskedCredentials(t *testing.T) {
	sm := NewSettingsManager()

	cfg := sm.config
	cfg.Providers["OpenAI"] = ProviderCredentials{
		ApiKey:      "sk-1234567890abcdef1234567890abcdef",
		BaseURL:     "https://api.openai.com/v1",
		StorageMode: StorageStored,
	}
	sm.UpdateConfig(cfg)

	masked := sm.GetMaskedConfig()
	key := masked.Providers["OpenAI"].ApiKey
	if !strings.Contains(key, "•") {
		t.Fatalf("Expected masked API key, got %s", key)
	}

	sm.UpdateConfig(masked)
	sm.mu.RLock()
	actualKey := sm.config.Providers["OpenAI"].ApiKey
	sm.mu.RUnlock()

	if actualKey != "sk-1234567890abcdef1234567890abcdef" {
		t.Errorf("Expected original unmasked key to be preserved, got %s", actualKey)
	}
}

func TestSettingsManager_TestProvider(t *testing.T) {
	sm := NewSettingsManager()

	success, msg, models := sm.TestProvider("OpenAI")
	if !success {
		t.Fatalf("TestProvider failed: %s", msg)
	}

	if len(models) == 0 {
		t.Errorf("Expected mock OpenAI models, got 0")
	}
}
