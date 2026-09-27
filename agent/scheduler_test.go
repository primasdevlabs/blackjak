package agent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_WaveExecutionAndDependencies(t *testing.T) {
	broker := NewEventBroker()
	orch := NewOrchestrator("run_sched_1", broker, nil)

	stages := []PlanStage{
		{ID: "node1", Role: "explorer", Task: "Explore codebase"},
		{ID: "node2", Role: "architect", Task: "Plan architecture", DependsOn: []string{"node1"}},
		{ID: "node3", Role: "coder", Task: "Write code", DependsOn: []string{"node1"}},
		{ID: "node4", Role: "tester", Task: "Run tests", DependsOn: []string{"node2", "node3"}},
	}

	graph := BuildGraphFromStages(stages)
	var executionOrder []string

	ctx := context.Background()
	err := orch.RunSchedule(ctx, graph, func(ctx context.Context, node *ScheduleNode) error {
		time.Sleep(10 * time.Millisecond)
		executionOrder = append(executionOrder, node.ID)
		return nil
	})

	if err != nil {
		t.Fatalf("RunSchedule failed: %v", err)
	}

	if !graph.IsComplete() {
		t.Errorf("Expected graph to be complete")
	}

	if graph.HasFailed() {
		t.Errorf("Graph reported failures unexpectedly")
	}
}

func TestScheduler_RetryPolicy(t *testing.T) {
	broker := NewEventBroker()
	orch := NewOrchestrator("run_sched_2", broker, nil)

	graph := NewDependencyGraph()
	graph.AddNode(&ScheduleNode{
		ID:     "flaky_node",
		Role:   "coder",
		Task:   "Flaky task",
		Status: SubagentCreated,
		RetryPolicy: RetryPolicy{
			MaxRetries: 2,
			Backoff:    5 * time.Millisecond,
		},
	})

	var attempts atomic.Int32

	ctx := context.Background()
	err := orch.RunSchedule(ctx, graph, func(ctx context.Context, node *ScheduleNode) error {
		cnt := attempts.Add(1)
		if cnt < 3 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Expected RunSchedule to succeed after retries, got error: %v", err)
	}

	if attempts.Load() != 3 {
		t.Errorf("Expected 3 execution attempts (1 initial + 2 retries), got %d", attempts.Load())
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	broker := NewEventBroker()
	orch := NewOrchestrator("run_sched_3", broker, nil)

	graph := NewDependencyGraph()
	graph.AddNode(&ScheduleNode{
		ID:     "slow_node",
		Role:   "coder",
		Task:   "Slow task",
		Status: SubagentCreated,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := orch.RunSchedule(ctx, graph, func(ctx context.Context, node *ScheduleNode) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err == nil {
		t.Fatal("Expected error on context cancellation, got nil")
	}
}
