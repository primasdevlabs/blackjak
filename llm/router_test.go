package llm

import (
	"context"
	"testing"
)

func TestModelRouter_Select(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewModelRouter(cfg)

	router.RegisterModels([]Model{
		{ID: "gpt-4o", Name: "GPT-4o"},
		{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet"},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini"},
	})

	ctx := context.Background()

	// Planning task -> Thinking model (gpt-4o)
	mPlan, err := router.Select(ctx, TaskPlanning)
	if err != nil || mPlan.ID != "gpt-4o" {
		t.Errorf("Expected gpt-4o for planning, got %s", mPlan.ID)
	}

	// Coding task -> Coding model (claude-3-5-sonnet)
	mCode, err := router.Select(ctx, TaskCoding)
	if err != nil || mCode.ID != "claude-3-5-sonnet" {
		t.Errorf("Expected claude-3-5-sonnet for coding, got %s", mCode.ID)
	}

	// Exploration task -> Fast model (gpt-4o-mini)
	mExplore, err := router.Select(ctx, TaskExploration)
	if err != nil || mExplore.ID != "gpt-4o-mini" {
		t.Errorf("Expected gpt-4o-mini for exploration, got %s", mExplore.ID)
	}
}
