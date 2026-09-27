package agent

import (
	"context"
	"strings"
)

// Plan represents a sequence of execution steps.
type Plan struct {
	Steps       []string `json:"steps"`
	CurrentStep int      `json:"currentStep"`
}

// Planner designs execution steps based on goals.
type Planner struct{}

// NewPlanner initializes a Planner instance.
func NewPlanner() *Planner {
	return &Planner{}
}

// CreatePlan generates execution steps for a given goal.
func (p *Planner) CreatePlan(ctx context.Context, goal string) (*Plan, error) {
	steps := []string{
		"Inspect workspace files and search code references",
		"Analyze implementation requirements",
		"Implement file modifications",
		"Run test suites and verify execution",
	}

	goalLower := strings.ToLower(goal)
	if strings.Contains(goalLower, "test") {
		steps = []string{
			"Locate failing test suites",
			"Analyze test assertions and error trace",
			"Implement code fix in workspace",
			"Run test suite to verify passing status",
		}
	} else if strings.Contains(goalLower, "refactor") {
		steps = []string{
			"Search codebase for target interfaces",
			"Extract reusable helper functions",
			"Modify target module files",
			"Run tests to ensure zero regressions",
		}
	}

	return &Plan{
		Steps:       steps,
		CurrentStep: 0,
	}, nil
}
