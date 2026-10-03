package llm

import (
	"fmt"
	"testing"
)

func TestFormatHTTPStatusError_524(t *testing.T) {
	err := formatHTTPStatusError(524, []byte("error code: 524"))
	if err == nil || !isTransientProviderError(err) {
		t.Fatalf("524 should be transient, got %v", err)
	}
	if got := err.Error(); !contains(got, "timed out") {
		t.Fatalf("want timed out message, got %q", got)
	}
}

func TestIsTransientProviderError(t *testing.T) {
	if !IsTransientProviderError(fmt.Errorf("provider timed out (HTTP 524)")) {
		t.Fatal("expected transient")
	}
	if IsTransientProviderError(fmt.Errorf("provider error: invalid api key")) {
		t.Fatal("auth errors must not retry forever")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
