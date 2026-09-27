package agent

import (
	"context"
	"testing"
	"time"
)

func TestOrchestrator_SpawnAndExecuteSubagent(t *testing.T) {
	broker := NewEventBroker()
	orch := NewOrchestrator("run_test_1", broker, nil)

	sub := orch.Spawn(SpawnRequest{
		Role:           "explorer",
		Task:           "Inspect workspace files",
		WorkspaceScope: []string{"src"},
	})

	if sub.Role != "explorer" || sub.Status != SubagentCreated {
		t.Fatalf("Unexpected subagent state: %+v", sub)
	}

	ctx := context.Background()
	err := orch.ExecuteSubagent(ctx, sub, func(ctx context.Context, s *Subagent) error {
		s.SetResult("Found 10 files")
		return nil
	})

	if err != nil {
		t.Fatalf("ExecuteSubagent failed: %v", err)
	}

	if sub.GetStatus() != SubagentCompleted {
		t.Errorf("Expected status completed, got %s", sub.GetStatus())
	}
	if sub.Result != "Found 10 files" {
		t.Errorf("Unexpected subagent result: %s", sub.Result)
	}
}

func TestOrchestrator_RunParallel(t *testing.T) {
	broker := NewEventBroker()
	orch := NewOrchestrator("run_test_2", broker, nil)

	sub1 := orch.Spawn(SpawnRequest{Role: "explorer", Task: "Task 1"})
	sub2 := orch.Spawn(SpawnRequest{Role: "architect", Task: "Task 2"})

	ctx := context.Background()
	tasks := map[*Subagent]func(context.Context, *Subagent) error{
		sub1: func(ctx context.Context, s *Subagent) error {
			time.Sleep(20 * time.Millisecond)
			s.SetResult("Explorer result")
			return nil
		},
		sub2: func(ctx context.Context, s *Subagent) error {
			time.Sleep(30 * time.Millisecond)
			s.SetResult("Architect result")
			return nil
		},
	}

	err := orch.RunParallel(ctx, tasks)
	if err != nil {
		t.Fatalf("RunParallel failed: %v", err)
	}

	if sub1.GetStatus() != SubagentCompleted || sub2.GetStatus() != SubagentCompleted {
		t.Errorf("Expected both subagents completed, got sub1=%s sub2=%s", sub1.GetStatus(), sub2.GetStatus())
	}
}

func TestOrchestrator_SubagentCancellation(t *testing.T) {
	broker := NewEventBroker()
	orch := NewOrchestrator("run_test_3", broker, nil)

	sub := orch.Spawn(SpawnRequest{Role: "coder", Task: "Long running task"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := orch.ExecuteSubagent(ctx, sub, func(ctx context.Context, s *Subagent) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}

	if sub.GetStatus() != SubagentCancelled {
		t.Errorf("Expected status cancelled, got %s", sub.GetStatus())
	}
}
