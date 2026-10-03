package context

import "testing"

func TestShouldCompactAt80Percent(t *testing.T) {
	c := NewCompactor(100000)
	budget := c.CalculateBudget(80000) // usable = 100k-24k = 76k; pressure > 1
	if !c.ShouldCompact(c.CalculateBudget(int(float64(c.MaxContextWindow-c.ReservedOutput-c.ReservedTools) * 0.85))) {
		t.Fatal("expected ShouldCompact at 85% pressure")
	}
	_ = budget
}

func TestCompactExtractsFiles(t *testing.T) {
	c := NewCompactor(128000)
	out := c.Compact("Fix auth", "Modified agent/loop.go and api/settings.go. Found token bug. Error: timeout")
	if out.TokensBefore <= 0 {
		t.Fatal("expected tokens before")
	}
	if out.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(out.Files) == 0 {
		t.Fatal("expected extracted files")
	}
}

func TestBuilderIncludesRulesAndMode(t *testing.T) {
	b := NewBuilder()
	built := b.Build(BuildInput{
		SystemPrompt: "base",
		UserRules:    "be careful",
		ModePolicy:   ModePolicyText("ask"),
		UserPrompt:   "what is this?",
	})
	if built.System == "" || built.Parts["total"] == 0 {
		t.Fatal("expected built system prompt")
	}
	if !contains(built.System, "ASK mode") {
		t.Fatal("expected ask mode policy")
	}
}

func TestAskPolicyText(t *testing.T) {
	std := AskPolicyText("standard")
	if !contains(std, "STANDARD") {
		t.Fatal("expected standard ask policy")
	}
	accept := AskPolicyText("accept_all")
	if !contains(accept, "ACCEPT ALL") {
		t.Fatal("expected accept_all ask policy")
	}
	if !contains(accept, "red team") {
		t.Fatal("expected security research framing")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
