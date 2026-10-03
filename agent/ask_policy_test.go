package agent

import (
	"strings"
	"testing"

	"blackjak/tools"
)

func TestBuildSystemPrompt_IncludesAskPolicy(t *testing.T) {
	a := New(nil, nil, nil)
	a.SetAskPolicyProvider(func() string { return "accept_all" })
	a.SetModeProvider(func() string { return "agent" })

	run := &Run{ID: "r1", Workspace: t.TempDir(), Prompt: "test"}
	prompt := a.buildSystemPrompt(run, nil, tools.DefaultPolicy())

	if !strings.Contains(prompt, "Ask Policy") {
		t.Fatal("expected Ask Policy section")
	}
	if !strings.Contains(prompt, "ACCEPT ALL") {
		t.Fatalf("expected accept_all policy text, got:\n%s", prompt)
	}
	if !strings.Contains(strings.ToLower(prompt), "red team") {
		t.Fatal("expected security research framing")
	}
}

func TestBuildSystemPrompt_StandardAskPolicy(t *testing.T) {
	a := New(nil, nil, nil)
	// No provider → standard default
	run := &Run{ID: "r1", Workspace: t.TempDir(), Prompt: "test"}
	prompt := a.buildSystemPrompt(run, nil, tools.DefaultPolicy())
	if !strings.Contains(prompt, "STANDARD") {
		t.Fatalf("expected STANDARD ask policy, got:\n%s", prompt)
	}
}
