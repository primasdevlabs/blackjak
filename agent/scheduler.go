package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScheduleNode represents a unit of work in the dependency graph.
type ScheduleNode struct {
	ID           string         `json:"id"`
	SubagentID   string         `json:"subagentId"`
	Role         string         `json:"role"`
	Task         string         `json:"task"`
	Dependencies []string       `json:"dependencies,omitempty"` // Node IDs that must complete first
	Scope        []string       `json:"scope,omitempty"`
	RetryPolicy  RetryPolicy    `json:"retryPolicy,omitempty"`
	Status       SubagentStatus `json:"status"`
}

// DependencyGraph manages the execution order of subagent tasks.
// Nodes are scheduled in waves: all nodes whose dependencies are satisfied
// run in parallel, and when they complete, the next wave of ready nodes starts.
type DependencyGraph struct {
	mu    sync.RWMutex
	Nodes map[string]*ScheduleNode `json:"nodes"`
}

// NewDependencyGraph creates an empty graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]*ScheduleNode),
	}
}

// AddNode registers a node in the graph.
func (g *DependencyGraph) AddNode(node *ScheduleNode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Nodes[node.ID] = node
}

// GetNode retrieves a node by ID.
func (g *DependencyGraph) GetNode(id string) (*ScheduleNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.Nodes[id]
	return n, ok
}

// SetNodeStatus updates the status of a node.
func (g *DependencyGraph) SetNodeStatus(id string, status SubagentStatus) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if n, ok := g.Nodes[id]; ok {
		n.Status = status
	}
}

// ReadyNodes returns all nodes whose dependencies are fully satisfied (completed)
// and that haven't started yet.
func (g *DependencyGraph) ReadyNodes() []*ScheduleNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ready := make([]*ScheduleNode, 0)
	for _, node := range g.Nodes {
		if node.Status != SubagentCreated && node.Status != SubagentQueued {
			continue
		}
		if g.allDependenciesMet(node) {
			ready = append(ready, node)
		}
	}
	return ready
}

// allDependenciesMet checks if all dependency nodes have completed. Must be called with lock held.
func (g *DependencyGraph) allDependenciesMet(node *ScheduleNode) bool {
	for _, depID := range node.Dependencies {
		dep, ok := g.Nodes[depID]
		if !ok {
			return false
		}
		if dep.Status != SubagentCompleted {
			return false
		}
	}
	return true
}

// IsComplete returns true when all nodes are in a terminal state.
func (g *DependencyGraph) IsComplete() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, node := range g.Nodes {
		switch node.Status {
		case SubagentCompleted, SubagentFailed, SubagentCancelled:
			continue
		default:
			return false
		}
	}
	return true
}

// HasFailed returns true if any node has failed.
func (g *DependencyGraph) HasFailed() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, node := range g.Nodes {
		if node.Status == SubagentFailed {
			return true
		}
	}
	return false
}

// PlanStage represents a single stage in a structured execution plan.
// The Planner generates these; the Scheduler converts them into a DependencyGraph.
type PlanStage struct {
	ID        string   `json:"id"`
	Role      string   `json:"role"`
	Task      string   `json:"task"`
	DependsOn []string `json:"dependsOn,omitempty"` // Stage IDs
	Scope     []string `json:"scope,omitempty"`
	Parallel  bool     `json:"parallel,omitempty"`
}

// BuildGraphFromStages converts a list of PlanStages into a DependencyGraph.
func BuildGraphFromStages(stages []PlanStage) *DependencyGraph {
	graph := NewDependencyGraph()
	for i := range stages {
		stage := stages[i]
		if stage.ID == "" {
			stage.ID = fmt.Sprintf("stage_%d_%d", i, time.Now().UnixNano())
		}
		graph.AddNode(&ScheduleNode{
			ID:           stage.ID,
			Role:         stage.Role,
			Task:         stage.Task,
			Dependencies: stage.DependsOn,
			Scope:        stage.Scope,
			Status:       SubagentCreated,
		})
	}
	return graph
}

// ConflictStrategy defines how the orchestrator handles file lease conflicts.
type ConflictStrategy string

const (
	ConflictWait  ConflictStrategy = "wait"  // Queue behind current holder
	ConflictForce ConflictStrategy = "force" // Orchestrator overrides lease
	ConflictAsk   ConflictStrategy = "ask"   // Prompt user for resolution
)

// RetryPolicy controls automatic retry behavior for failed subagents.
type RetryPolicy struct {
	MaxRetries int           `json:"maxRetries"`
	Backoff    time.Duration `json:"backoff"`
}

// RunSchedule executes the dependency graph, processing ready nodes in parallel waves.
// It blocks until all nodes complete or the context is cancelled.
func (o *Orchestrator) RunSchedule(ctx context.Context, graph *DependencyGraph, execFn func(ctx context.Context, node *ScheduleNode) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if graph.IsComplete() {
			break
		}

		if graph.HasFailed() {
			return fmt.Errorf("dependency graph execution aborted: one or more nodes failed")
		}

		readyNodes := graph.ReadyNodes()
		if len(readyNodes) == 0 {
			// No nodes ready — wait briefly for running nodes to complete
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Execute all ready nodes in parallel
		var wg sync.WaitGroup
		errCh := make(chan error, len(readyNodes))

		for _, node := range readyNodes {
			graph.SetNodeStatus(node.ID, SubagentRunning)
			wg.Add(1)

			go func(n *ScheduleNode) {
				defer wg.Done()
				var lastErr error
				maxAttempts := 1
				if n.RetryPolicy.MaxRetries > 0 {
					maxAttempts = 1 + n.RetryPolicy.MaxRetries
				}
				for attempt := 1; attempt <= maxAttempts; attempt++ {
					if err := execFn(ctx, n); err != nil {
						lastErr = err
						if attempt < maxAttempts {
							if n.RetryPolicy.Backoff > 0 {
								select {
								case <-ctx.Done():
									lastErr = ctx.Err()
									break
								case <-time.After(n.RetryPolicy.Backoff):
								}
							}
							continue
						}
					} else {
						lastErr = nil
						break
					}
				}

				if lastErr != nil {
					graph.SetNodeStatus(n.ID, SubagentFailed)
					errCh <- fmt.Errorf("node %s (%s) failed after %d attempt(s): %w", n.ID, n.Role, maxAttempts, lastErr)
				} else {
					graph.SetNodeStatus(n.ID, SubagentCompleted)
				}
			}(node)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				return err
			}
		}
	}

	return nil
}
