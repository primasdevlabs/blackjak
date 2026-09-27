package tools

import (
	"context"
)

// SearchTool provides code search capability.
type SearchTool struct{}

// NewSearchTool initializes a new SearchTool.
func NewSearchTool() *SearchTool {
	return &SearchTool{}
}

func (s *SearchTool) Name() string { return "search" }

func (s *SearchTool) Description() string {
	return "Search code and text across workspace files."
}

func (s *SearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return nil, nil
}
