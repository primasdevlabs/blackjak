package llm

import (
	"context"
	"sync"
)

// TaskType defines specific task categories for model routing.
type TaskType string

const (
	TaskPlanning    TaskType = "planning"
	TaskExploration TaskType = "exploration"
	TaskCoding      TaskType = "coding"
	TaskDebugging   TaskType = "debugging"
	TaskTesting     TaskType = "testing"
	TaskReview      TaskType = "review"
	TaskSummary     TaskType = "summarization"
)

// ModelRole defines model assignment roles.
type ModelRole string

const (
	RoleThinking ModelRole = "thinking"
	RoleCoding   ModelRole = "coding"
	RoleFast     ModelRole = "fast"
)

// RouterConfig maps task types to model roles.
type RouterConfig struct {
	UseSeparateModels bool                 `json:"useSeparateModels"`
	ThinkingModelID   string               `json:"thinkingModelId"`
	CodingModelID     string               `json:"codingModelId"`
	FastModelID       string               `json:"fastModelId"`
	TaskRoutes        map[TaskType]ModelRole `json:"taskRoutes"`
}

// DefaultRouterConfig creates reasonable default task-to-model role routes.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		UseSeparateModels: true,
		ThinkingModelID:   "gpt-4o",
		CodingModelID:     "claude-3-5-sonnet",
		FastModelID:       "gpt-4o-mini",
		TaskRoutes: map[TaskType]ModelRole{
			TaskPlanning:    RoleThinking,
			TaskExploration: RoleFast,
			TaskCoding:      RoleCoding,
			TaskDebugging:   RoleThinking,
			TaskTesting:     RoleCoding,
			TaskReview:      RoleThinking,
			TaskSummary:     RoleFast,
		},
	}
}

// ModelRouter selects models dynamically based on task type and configuration.
type ModelRouter struct {
	mu     sync.RWMutex
	config RouterConfig
	models map[string]Model
}

// NewModelRouter initializes a new ModelRouter.
func NewModelRouter(cfg RouterConfig) *ModelRouter {
	return &ModelRouter{
		config: cfg,
		models: make(map[string]Model),
	}
}

// UpdateConfig updates routing assignments.
func (r *ModelRouter) UpdateConfig(cfg RouterConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = cfg
}

// RegisterModels registers available models for lookup.
func (r *ModelRouter) RegisterModels(models []Model) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range models {
		r.models[m.ID] = m
	}
}

// Select returns the target model for a given task type.
func (r *ModelRouter) Select(ctx context.Context, task TaskType) (Model, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.config.UseSeparateModels {
		m, ok := r.models[r.config.CodingModelID]
		if ok {
			return m, nil
		}
		return Model{ID: r.config.CodingModelID, Name: r.config.CodingModelID}, nil
	}

	targetRole, ok := r.config.TaskRoutes[task]
	if !ok {
		targetRole = RoleCoding
	}

	var modelID string
	switch targetRole {
	case RoleThinking:
		modelID = r.config.ThinkingModelID
	case RoleFast:
		modelID = r.config.FastModelID
	default:
		modelID = r.config.CodingModelID
	}

	m, ok := r.models[modelID]
	if !ok {
		return Model{ID: modelID, Name: modelID}, nil
	}

	return m, nil
}
