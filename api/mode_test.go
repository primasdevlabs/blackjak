package api

import (
	"encoding/json"
	"testing"
)

func TestNormalizeMode(t *testing.T) {
	if NormalizeMode(ModeCode) != ModeAgent {
		t.Fatal("code should map to agent")
	}
	if NormalizeMode(ModeAsk) != ModeAsk {
		t.Fatal("ask should stay ask")
	}
	if NormalizeMode(ModePlan) != ModePlan {
		t.Fatal("plan should stay plan")
	}
}

func TestPolicyAskIsReadOnly(t *testing.T) {
	sm := NewSettingsManager()
	sm.ApplyPatch(map[string]json.RawMessage{"mode": json.RawMessage(`"ask"`)})
	p := sm.Policy()
	if string(p.Mode) != "readonly" {
		t.Fatalf("ask mode should force readonly, got %s", p.Mode)
	}
}
