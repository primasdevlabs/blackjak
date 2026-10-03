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
	}
	for _, c := range cases {
		if got := ClassifyTask(c); got != TaskTrivial {
			t.Errorf("ClassifyTask(%q) = %s, want trivial", c, got)
		}
	}
}

func TestClassifyTask_NonTrivial(t *testing.T) {
	cases := []string{
		"Implement JWT authentication for the API",
		"Add a database migration for users table",
		"Fix the race condition in the orchestrator",
		"Refactor the agent loop to support hybrid engineering gates",
	}
	for _, c := range cases {
		if got := ClassifyTask(c); got != TaskNonTrivial {
			t.Errorf("ClassifyTask(%q) = %s, want nonTrivial", c, got)
		}
	}
}

func TestWriteGate_BlocksWithoutInspect(t *testing.T) {
	eng := NewEngineeringState("Implement a new filesystem helper")
	err := checkWriteGate(eng, "agent", "filesystem", map[string]interface{}{"operation": "write"})
	if err == nil {
		t.Fatal("expected write gate error")
	}
	if !strings.Contains(err.Error(), "inspect before modifying") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteGate_AllowsAfterInspect(t *testing.T) {
	eng := NewEngineeringState("Implement a new filesystem helper")
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

func TestCompleteGate_RequiresVerify(t *testing.T) {
	eng := NewEngineeringState("Fix the bug in shell.go")
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

func TestPromoteOnImplement(t *testing.T) {
	eng := NewEngineeringState("What is X?")
	if eng.TaskClass != TaskTrivial {
		t.Fatal("want trivial")
	}
	eng.MarkImplement()
	if eng.TaskClass != TaskNonTrivial {
		t.Fatalf("want nonTrivial after write, got %s", eng.TaskClass)
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
	eng := NewEngineeringState("Implement API rate limiting")
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
}
