package agent

import (
	"context"
	"testing"
	"time"
)

func TestRunManager_CreateAndCancelRun(t *testing.T) {
	broker := NewEventBroker()
	rm := NewRunManager(broker)

	run := rm.CreateRun("Fix bugs", "/tmp/test")
	if run.Status != RunPending {
		t.Errorf("Expected status pending, got %s", run.Status)
	}

	ok := rm.CancelRun(run.ID)
	if !ok {
		t.Errorf("Expected CancelRun to return true")
	}

	if run.Status != RunCancelled {
		t.Errorf("Expected status cancelled, got %s", run.Status)
	}
}

func TestRunManager_ApprovalFlow(t *testing.T) {
	broker := NewEventBroker()
	rm := NewRunManager(broker)
	run := rm.CreateRun("Dangerous command step", "/tmp/test")

	req := ApprovalRequest{
		ID:          "app_1",
		RunID:       run.ID,
		Operation:   "shell",
		Description: "Run shell command",
	}

	done := make(chan bool)
	go func() {
		granted, _, err := rm.RequestApproval(context.Background(), run, req)
		if err != nil || !granted {
			t.Errorf("RequestApproval failed or denied unexpectedly: %v", err)
		}
		done <- true
	}()

	time.Sleep(50 * time.Millisecond)
	if run.Status != RunWaiting {
		t.Errorf("Expected run status waiting, got %s", run.Status)
	}

	err := rm.SubmitApproval(run.ID, ApprovalResponse{
		RequestID: "app_1",
		Granted:   true,
	})
	if err != nil {
		t.Fatalf("SubmitApproval failed: %v", err)
	}

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for approval resolution")
	}
}
