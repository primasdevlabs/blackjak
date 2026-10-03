package agent

import (
	"strings"
	"testing"
)

func TestClassifyTask_TrivialQuestions(t *testing.T) {
	cases := []string{
		"What does ExecuteRun do?",
		"Where is the Plan type defined?",
		"Explain the event broker",
		"how does compaction work?",
		"Fix the typo in README",
		"Move the Files bar above the composer",
		"create a simple login page\nits just a test page, not a full implementation",
		"Make a demo page with a login form",
	}
	for _, c := range cases {
		if got := ClassifyTask(c); got != TaskTrivial {
			t.Errorf("ClassifyTask(%q) = %s, want trivial", c, got)
		}
	}
}

func TestDemoPromptForcesFastPath(t *testing.T) {
	eng := NewEngineeringState("create a simple login page — just a test page, not a full implementation")
	if !eng.IsFastPath() {
		t.Fatal("demo login page must use fast path")
	}
	if eng.NeedsGates("agent") {
		t.Fatal("demo must not enable engineering gates")
	}
}

func TestClassifyTask_NonTrivial(t *testing.T) {
	cases := []string{
		"Implement JWT authentication for the API",
		"Add a database migration for users table",
		"Fix the race condition in the orchestrator",
		"Refactor the agent loop to support hybrid engineering gates across resume",
	}
	for _, c := range cases {
		if got := ClassifyTask(c); got != TaskNonTrivial {
			t.Errorf("ClassifyTask(%q) = %s, want nonTrivial", c, got)
		}
	}
}

func TestWriteGate_BlocksWithoutInspect(t *testing.T) {
	eng := NewEngineeringState("Implement JWT authentication for the API")
	err := checkWriteGate(eng, "agent", "filesystem", map[string]interface{}{"operation": "write"})
	if err == nil {
		t.Fatal("expected write gate error")
	}
	if !strings.Contains(err.Error(), "inspect before modifying") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteGate_AllowsAfterInspect(t *testing.T) {
	eng := NewEngineeringState("Implement JWT authentication for the API")
	eng.MarkInspect()
	if err := checkWriteGate(eng, "agent", "filesystem", map[string]interface{}{"operation": "write"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestWriteGate_TrivialSkips(t *testing.T) {
	eng := NewEngineeringState("What is Plan?")
	if eng.TaskClass != TaskTrivial {
		t.Fatalf("expected trivial, got %s", eng.TaskClass)
	}
	if err := checkWriteGate(eng, "agent", "filesystem", map[string]interface{}{"operation": "write"}); err != nil {
		t.Fatalf("trivial should skip write gate: %v", err)
	}
}

func TestCompleteGate_RequiresVerifyOnlyWhenGated(t *testing.T) {
	trivial := NewEngineeringState("Fix typo in shell.go")
	trivial.MarkImplement()
	if err := checkCompleteGate(trivial, "agent"); err != nil {
		t.Fatalf("trivial edits should not require verify: %v", err)
	}

	eng := NewEngineeringState("Implement JWT authentication for the API")
	eng.MarkInspect()
	eng.MarkImplement()
	err := checkCompleteGate(eng, "agent")
	if err == nil {
		t.Fatal("expected verify gate")
	}
	eng.MarkVerify()
	if err := checkCompleteGate(eng, "agent"); err != nil {
		t.Fatalf("unexpected after verify: %v", err)
	}
}

func TestImplementDoesNotAutoPromote(t *testing.T) {
	eng := NewEngineeringState("What is X?")
	if eng.TaskClass != TaskTrivial {
		t.Fatal("want trivial")
	}
	eng.MarkImplement()
	if eng.TaskClass != TaskTrivial {
		t.Fatalf("local edits must stay trivial, got %s", eng.TaskClass)
	}
	if !eng.HasFilesChanged() {
		t.Fatal("want filesChanged")
	}
}

func TestReviewReviseCycle(t *testing.T) {
	eng := NewEngineeringState("Implement secure session auth")
	eng.MarkInspect()
	eng.MarkImplement()
	eng.MarkVerify()
	eng.MarkReview(&ReviewResult{Verdict: "revise", Findings: []string{"hardcoded secret"}})
	if eng.HasReviewPass() {
		t.Fatal("revise should not pass review")
	}
	if snap := eng.Snapshot(); snap["phase"] != string(PhaseRefine) {
		t.Fatalf("want refine, got %v", snap["phase"])
	}
	eng.MarkReview(&ReviewResult{Verdict: "pass", SolvesProblem: true})
	if !eng.HasReviewPass() {
		t.Fatal("want pass")
	}
}

func TestParseReviewResult(t *testing.T) {
	raw := "```json\n{\"solvesProblem\":true,\"unnecessaryComplexity\":false,\"aiSlopRisk\":false,\"verdict\":\"pass\"}\n```"
	r, err := parseReviewResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Verdict != "pass" {
		t.Fatalf("got %s", r.Verdict)
	}
}

func TestVerifyShellPatterns(t *testing.T) {
	if !isVerifyShellCommand("go test ./agent -count=1") {
		t.Fatal("go test should verify")
	}
	if isVerifyShellCommand("echo hello") {
		t.Fatal("echo should not verify")
	}
}

func TestToolAllowedForRole(t *testing.T) {
	if !ToolAllowedForRole("explorer", "filesystem") {
		t.Fatal("explorer should allow filesystem")
	}
	if ToolAllowedForRole("explorer", "shell") {
		t.Fatal("explorer should not allow shell")
	}
	if !ToolAllowedForRole("coder", "shell") {
		t.Fatal("coder should allow shell")
	}
}

func TestNeedsAutoReview(t *testing.T) {
	eng := NewEngineeringState("Implement JWT authentication for the API")
	eng.MarkInspect()
	eng.MarkImplement()
	eng.MarkVerify()
	if !needsAutoReview(eng, "agent") {
		t.Fatal("want auto review")
	}
	eng.MarkReview(&ReviewResult{Verdict: "pass"})
	if needsAutoReview(eng, "agent") {
		t.Fatal("pass should clear auto review need")
	}

	local := NewEngineeringState("Tweak FilesReviewBar padding")
	local.MarkImplement()
	if needsAutoReview(local, "agent") {
		t.Fatal("local edits should skip auto review")
	}
}

func TestToolOutcomeStripsPreviousContent(t *testing.T) {
	out := toolOutcome(map[string]interface{}{
		"path":            "a.go",
		"replaced":        1,
		"previousContent": strings.Repeat("x", 5000),
	}, nil)
	if strings.Contains(out, "previousContent") {
		t.Fatal("previousContent must not reach the model")
	}
	if !strings.Contains(out, "previousOmitted") {
		t.Fatal("expected previousOmitted marker")
	}
}
