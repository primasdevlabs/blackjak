package compatible

import "testing"

func TestNormalizeCompatibleBaseURL(t *testing.T) {
	cases := map[string]string{
		"https://api.notrack.ai/v1":                       "https://api.notrack.ai/v1",
		"https://api.notrack.ai/v1/":                      "https://api.notrack.ai/v1",
		"https://api.notrack.ai/v1/chat/completions":      "https://api.notrack.ai/v1",
		"https://api.notrack.ai/v1/chat/completions/":     "https://api.notrack.ai/v1",
		"http://localhost:11434/v1/models":                "http://localhost:11434/v1",
		"https://openrouter.ai/api/v1/completions":        "https://openrouter.ai/api/v1",
	}
	for in, want := range cases {
		if got := normalizeCompatibleBaseURL(in); got != want {
			t.Fatalf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNew_StripsChatCompletionsPath(t *testing.T) {
	p := New("https://api.notrack.ai/v1/chat/completions", "sk-test", "notrack-uncensored")
	if p.baseURL != "https://api.notrack.ai/v1" {
		t.Fatalf("baseURL = %q", p.baseURL)
	}
	if p.modelID != "notrack-uncensored" {
		t.Fatalf("modelID = %q", p.modelID)
	}
}
