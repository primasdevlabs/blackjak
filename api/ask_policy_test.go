package api

import (
	"encoding/json"
	"testing"
)

func TestNormalizeAskPolicy(t *testing.T) {
	if NormalizeAskPolicy(AskPolicyAcceptAll) != AskPolicyAcceptAll {
		t.Fatal("accept_all should stay accept_all")
	}
	if NormalizeAskPolicy(AskPolicyStandard) != AskPolicyStandard {
		t.Fatal("standard should stay standard")
	}
	if NormalizeAskPolicy(AskPolicy("bogus")) != AskPolicyStandard {
		t.Fatal("unknown should map to standard")
	}
	if NormalizeAskPolicy("") != AskPolicyStandard {
		t.Fatal("empty should map to standard")
	}
}

func TestAskPolicy_DefaultAndPatch(t *testing.T) {
	sm := NewSettingsManager()
	if sm.GetConfig().AskPolicy != AskPolicyStandard {
		t.Fatalf("default askPolicy = %q, want standard", sm.GetConfig().AskPolicy)
	}

	sm.ApplyPatch(map[string]json.RawMessage{
		"askPolicy": json.RawMessage(`"accept_all"`),
	})
	if sm.GetConfig().AskPolicy != AskPolicyAcceptAll {
		t.Fatalf("after patch askPolicy = %q, want accept_all", sm.GetConfig().AskPolicy)
	}

	// Unrelated patch must not clobber askPolicy.
	sm.ApplyPatch(map[string]json.RawMessage{
		"effort": json.RawMessage(`"high"`),
	})
	if sm.GetConfig().AskPolicy != AskPolicyAcceptAll {
		t.Fatalf("askPolicy clobbered by unrelated patch: %q", sm.GetConfig().AskPolicy)
	}

	sm.ApplyPatch(map[string]json.RawMessage{
		"askPolicy": json.RawMessage(`"nope"`),
	})
	if sm.GetConfig().AskPolicy != AskPolicyStandard {
		t.Fatalf("invalid askPolicy should normalize to standard, got %q", sm.GetConfig().AskPolicy)
	}
}

func TestAskPolicy_MaskedConfigIncludesField(t *testing.T) {
	sm := NewSettingsManager()
	sm.ApplyPatch(map[string]json.RawMessage{
		"askPolicy": json.RawMessage(`"accept_all"`),
	})
	masked := sm.GetMaskedConfig()
	if masked.AskPolicy != AskPolicyAcceptAll {
		t.Fatalf("masked config askPolicy = %q", masked.AskPolicy)
	}
}
