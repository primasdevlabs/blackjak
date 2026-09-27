package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestSettingsManager_TestProvider_NoKey(t *testing.T) {
	sm := NewSettingsManager()

	// No key configured — must fail, not return the static catalog.
	success, msg, _ := sm.TestProvider("OpenAI")
	if success {
		t.Fatal("TestProvider reported success without an API key")
	}
	if !strings.Contains(msg, "API key") {
		t.Errorf("Expected missing-key message, got %q", msg)
	}
}

func TestSettingsManager_TestProvider_LiveSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{{"id": "gpt-4o"}},
		})
	}))
	defer server.Close()

	sm := NewSettingsManager()
	cfg := sm.config
	cfg.Providers["OpenAI"] = ProviderCredentials{
		ApiKey:      "sk-test",
		BaseURL:     server.URL,
		StorageMode: StorageStored,
	}
	sm.UpdateConfig(cfg)

	success, msg, models := sm.TestProvider("OpenAI")
	if !success {
		t.Fatalf("TestProvider failed against live endpoint: %s", msg)
	}
	if len(models) == 0 {
		t.Error("Expected models from live endpoint")
	}
}

func TestSettingsManager_TestProvider_BadKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	sm := NewSettingsManager()
	cfg := sm.config
	cfg.Providers["OpenAI"] = ProviderCredentials{
		ApiKey:      "sk-wrong",
		BaseURL:     server.URL,
		StorageMode: StorageStored,
	}
	sm.UpdateConfig(cfg)

	success, msg, _ := sm.TestProvider("OpenAI")
	if success {
		t.Fatal("TestProvider succeeded with a rejected key")
	}
	if !strings.Contains(msg, "authentication failed") {
		t.Errorf("Expected auth-failure message, got %q", msg)
	}
}

func TestSettingsManager_TestProvider_Unreachable(t *testing.T) {
	sm := NewSettingsManager()
	cfg := sm.config
	cfg.Providers["OpenAI"] = ProviderCredentials{
		ApiKey:      "sk-test",
		BaseURL:     "http://127.0.0.1:1", // nothing listens here
		StorageMode: StorageStored,
	}
	sm.UpdateConfig(cfg)

	success, msg, _ := sm.TestProvider("OpenAI")
	if success {
		t.Fatal("TestProvider succeeded against unreachable endpoint")
	}
	if !strings.Contains(msg, "connection failed") {
		t.Errorf("Expected connection-failure message, got %q", msg)
	}
}
