package llm

import "testing"

func TestSetHTTPProxy(t *testing.T) {
	SetHTTPProxy("http://127.0.0.1:7890")
	if HTTPProxy() != "http://127.0.0.1:7890" {
		t.Fatalf("expected proxy to stick, got %q", HTTPProxy())
	}
	SetHTTPProxy("")
	// After clear, may fall back to env — just ensure no panic and client builds.
	_ = NewHTTPClient(1)
}
