package llm

import (
	"context"
	"testing"
)

func TestModelRouter_Select(t *testing.T) {
	cfg := DefaultRouterConfig()
	router := NewModelRouter(cfg)

	router.RegisterModels([]Model{
		{ID: "claude-opus-5", Name: "Claude Opus 5"},
		{ID: "gpt-5.3-codex", Name: "GPT-5.3-Codex"},
		{ID: "gemini-3.5-flash-lite", Name: "Gemini 3.5 Flash-Lite"},
		{ID: "claude-sonnet-5", Name: "Claude Sonnet 5"},
	})

	ctx := context.Background()

	// Planning task -> Thinking model (claude-opus-5)
	mPlan, err := router.Select(ctx, TaskPlanning)
	if err != nil || mPlan.ID != "claude-opus-5" {
		t.Errorf("Expected claude-opus-5 for planning, got %s", mPlan.ID)
	}

	// Coding task -> Coding model (gpt-5.3-codex)
	mCode, err := router.Select(ctx, TaskCoding)
	if err != nil || mCode.ID != "gpt-5.3-codex" {
		t.Errorf("Expected gpt-5.3-codex for coding, got %s", mCode.ID)
	}

	// Exploration task -> Fast model (gemini-3.5-flash-lite)
	mExplore, err := router.Select(ctx, TaskExploration)
	if err != nil || mExplore.ID != "gemini-3.5-flash-lite" {
		t.Errorf("Expected gemini-3.5-flash-lite for exploration, got %s", mExplore.ID)
	}

	// Review task -> Review model (claude-sonnet-5)
	mReview, err := router.Select(ctx, TaskReview)
	if err != nil || mReview.ID != "claude-sonnet-5" {
		t.Errorf("Expected claude-sonnet-5 for review, got %s", mReview.ID)
	}
}
