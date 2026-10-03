package agent

import (
	"os"
	"path/filepath"
	"testing"

	"blackjak/llm"
)

func TestEnsureToolCallResults_FillsGaps(t *testing.T) {
	msgs := []llm.Message{
		{Role: llm.RoleUser, Content: "hi"},
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "c1", Name: "filesystem"},
				{ID: "c2", Name: "test"},
			},
		},
		{Role: llm.RoleTool, ToolCallID: "c1", Name: "filesystem", Content: `{"ok":true}`},
	}
	out := ensureToolCallResults(msgs)
	if len(out) != 4 {
		t.Fatalf("want 4 messages, got %d", len(out))
	}
	last := out[len(out)-1]
	if last.Role != llm.RoleTool || last.ToolCallID != "c2" {
		t.Fatalf("expected synthetic tool result for c2, got %+v", last)
	}
}

func TestEnsureToolCallResults_NoOpWhenComplete(t *testing.T) {
	msgs := []llm.Message{
		{
			Role:      llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{{ID: "c1", Name: "shell"}},
		},
		{Role: llm.RoleTool, ToolCallID: "c1", Name: "shell", Content: "ok"},
	}
	out := ensureToolCallResults(msgs)
	if len(out) != 2 {
		t.Fatalf("want unchanged length 2, got %d", len(out))
	}
}

func TestResumeRun_FromDiskAfterMissingMemory(t *testing.T) {
	dir := t.TempDir()
	rm := NewRunManager(nil)
	run := rm.CreateRun("fix the bug", dir)
	run.SetMessages([]llm.Message{
		{Role: llm.RoleSystem, Content: "sys"},
		{Role: llm.RoleUser, Content: "fix the bug"},
		{
			Role:      llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{{ID: "t1", Name: "test"}},
		},
	})
	run.Engineering = NewEngineeringState("fix the bug")
	run.Engineering.MarkInspect()
	run.SetStatus(RunFailed)
	rm.SaveCheckpoint(run, run.GetMessages())

	// Drop from memory (simulates ClearFinished / restart gap).
	rm.mu.Lock()
	delete(rm.runs, run.ID)
	rm.mu.Unlock()

	resumed, err := rm.ResumeRun(run.ID, "continue from the failure", dir)
	if err != nil {
		t.Fatalf("ResumeRun: %v", err)
	}
	if resumed.Status != RunPending {
		t.Fatalf("status = %s", resumed.Status)
	}
	msgs := resumed.GetMessages()
	if len(msgs) < 4 {
		t.Fatalf("expected sanitized+steering messages, got %d", len(msgs))
	}
	// Incomplete tool batch must be closed before the steering prompt.
	foundSynthetic := false
	for _, m := range msgs {
		if m.Role == llm.RoleTool && m.ToolCallID == "t1" {
			foundSynthetic = true
		}
	}
	if !foundSynthetic {
		t.Fatal("expected synthetic tool result for interrupted call")
	}
	if resumed.Engineering == nil || !resumed.Engineering.HasInspect() {
		t.Fatal("expected engineering evidence restored")
	}
}

func TestSaveCheckpoint_WritesTerminalStatus(t *testing.T) {
	dir := t.TempDir()
	rm := NewRunManager(nil)
	run := rm.CreateRun("task", dir)
	run.SetMessages([]llm.Message{{Role: llm.RoleUser, Content: "x"}})
	run.SetStatus(RunPaused)
	rm.SaveCheckpoint(run, run.GetMessages())

	cp, err := rm.LoadCheckpoint(run.ID, dir)
	if err != nil {
		t.Fatal(err)
	}
	if cp.Status != RunPaused {
		t.Fatalf("checkpoint status = %s, want paused", cp.Status)
	}
	if _, err := os.Stat(filepath.Join(dir, ".blackjak", "runs", run.ID+".json")); err != nil {
		t.Fatal(err)
	}
}

func TestClearFinished_KeepsFailedAndPaused(t *testing.T) {
	rm := NewRunManager(nil)
	a := rm.CreateRun("a", t.TempDir())
	b := rm.CreateRun("b", t.TempDir())
	c := rm.CreateRun("c", t.TempDir())
	a.SetStatus(RunCompleted)
	b.SetStatus(RunFailed)
	c.SetStatus(RunPaused)
	n := rm.ClearFinished()
	if n != 1 {
		t.Fatalf("removed %d, want 1", n)
	}
	if _, ok := rm.GetRun(b.ID); !ok {
		t.Fatal("failed run should be kept for resume")
	}
	if _, ok := rm.GetRun(c.ID); !ok {
		t.Fatal("paused run should be kept for resume")
	}
}
