package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"blackjak/workspace"
)

// SpawnRequest contains parameters for creating a specialized subagent.
type SpawnRequest struct {
	Role           string   `json:"role"`
	Task           string   `json:"task"`
	WorkspaceScope []string `json:"workspaceScope,omitempty"`
}

// Orchestrator manages subagent spawning, parallel execution, and handoffs.
type Orchestrator struct {
	mu          sync.RWMutex
	parentRunID string
	broker      *EventBroker
	tracker     *workspace.FileTrackerManager
	subagents   map[string]*Subagent
}

// NewOrchestrator initializes an Orchestrator bound to a parent run.
func NewOrchestrator(parentRunID string, broker *EventBroker, tracker *workspace.FileTrackerManager) *Orchestrator {
	if tracker == nil {
		tracker = workspace.NewFileTrackerManager()
	}
	return &Orchestrator{
		parentRunID: parentRunID,
		broker:      broker,
		tracker:     tracker,
		subagents:   make(map[string]*Subagent),
	}
}

// Spawn creates and registers a new subagent under the parent run.
func (o *Orchestrator) Spawn(req SpawnRequest) *Subagent {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := fmt.Sprintf("sub_%s_%d", req.Role, time.Now().UnixNano())
	sub := NewSubagent(id, o.parentRunID, req.Role, req.Task, req.WorkspaceScope)
	o.subagents[id] = sub

	o.emitEvent(EventAgentCreated, sub, map[string]interface{}{
		"id":        sub.ID,
		"role":      sub.Role,
		"task":      sub.Task,
		"scope":     sub.WorkspaceScope,
		"parentRun": o.parentRunID,
	})

	return sub
}

// GetSubagent retrieves a subagent by ID.
func (o *Orchestrator) GetSubagent(id string) (*Subagent, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	sub, ok := o.subagents[id]
	return sub, ok
}

// ListSubagents returns all subagents managed by this orchestrator.
func (o *Orchestrator) ListSubagents() []*Subagent {
	o.mu.RLock()
	defer o.mu.RUnlock()
	list := make([]*Subagent, 0, len(o.subagents))
	for _, sub := range o.subagents {
		list = append(list, sub)
	}
	return list
}

// ExecuteSubagent runs a subagent execution function with state tracking and cancellation handling.
func (o *Orchestrator) ExecuteSubagent(ctx context.Context, sub *Subagent, execFn func(ctx context.Context, s *Subagent) error) error {
	sub.SetStatus(SubagentRunning)
	o.emitEvent(EventAgentStarted, sub, map[string]interface{}{
		"id":   sub.ID,
		"role": sub.Role,
		"task": sub.Task,
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- execFn(ctx, sub)
	}()

	select {
	case <-ctx.Done():
		sub.SetStatus(SubagentCancelled)
		o.emitEvent(EventAgentCancelled, sub, map[string]interface{}{
			"id":     sub.ID,
			"reason": "Cancelled by context",
		})
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			sub.SetError(err.Error())
			o.emitEvent(EventAgentFailed, sub, map[string]interface{}{
				"id":    sub.ID,
				"error": err.Error(),
			})
			return err
		}

		sub.SetStatus(SubagentCompleted)
		o.emitEvent(EventAgentCompleted, sub, map[string]interface{}{
			"id":       sub.ID,
			"result":   sub.Result,
			"findings": sub.Findings,
		})
		return nil
	}
}

// RunParallel executes multiple subagents concurrently using structured concurrency.
func (o *Orchestrator) RunParallel(ctx context.Context, tasks map[*Subagent]func(ctx context.Context, s *Subagent) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))

	for sub, fn := range tasks {
		wg.Add(1)
		go func(s *Subagent, exec func(context.Context, *Subagent) error) {
			defer wg.Done()
			if err := o.ExecuteSubagent(ctx, s, exec); err != nil {
				errCh <- fmt.Errorf("subagent %s (%s) failed: %w", s.ID, s.Role, err)
			}
		}(sub, fn)
	}

	wg.Wait()
	close(errCh)

	// Collect any subagent error
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// Handoff passes findings or results from one subagent to another.
func (o *Orchestrator) Handoff(fromSub, toSub *Subagent, data any) {
	msg := AgentMessage{
		FromAgentID: fromSub.ID,
		ToAgentID:   toSub.ID,
		Type:        MsgTypeHandoff,
		Data:        data,
		Timestamp:   time.Now(),
	}

	o.emitEvent(EventAgentHandoff, fromSub, map[string]interface{}{
		"from": fromSub.ID,
		"to":   toSub.ID,
		"data": data,
		"msg":  msg,
	})
}

func (o *Orchestrator) emitEvent(evtType EventType, sub *Subagent, data interface{}) {
	evt := Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		RunID:     o.parentRunID,
		AgentID:   sub.ID,
		Type:      evtType,
		Timestamp: time.Now(),
		Data:      data,
	}
	if o.broker != nil {
		o.broker.Publish(evt)
	}
}
