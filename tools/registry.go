package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"blackjak/llm"
	"blackjak/workspace"
)

// ApprovalFunc gates destructive operations. It returns true if the operation
// may proceed. Implementations may block waiting on a human decision.
type ApprovalFunc func(ctx context.Context, operation, description string) (bool, error)

// Tool defines the interface for executable agent capabilities.
type Tool interface {
	Name() string
	Description() string
	// Schema returns a JSON Schema object describing the arguments.
	Schema() interface{}
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// Registry manages the set of available tools.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool instance.
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, error) {
	t, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return t, nil
}

// List returns all registered tools sorted by name.
func (r *Registry) List() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// Schemas returns llm.Tool definitions for every registered tool.
func (r *Registry) Schemas() []llm.Tool {
	defs := make([]llm.Tool, 0, len(r.tools))
	for _, t := range r.List() {
		defs = append(defs, llm.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema(),
		})
	}
	return defs
}

// FilterKeep retains only tools whose names are in allow. Empty allow keeps all.
// Meta tools (not in the registry) are unaffected.
func (r *Registry) FilterKeep(allow []string) {
	if len(allow) == 0 {
		return
	}
	keep := make(map[string]struct{}, len(allow))
	for _, n := range allow {
		keep[n] = struct{}{}
	}
	for name := range r.tools {
		if _, ok := keep[name]; !ok {
			delete(r.tools, name)
		}
	}
}

// DefaultRegistry builds the standard tool set for a workspace. The approval
// callback is invoked before destructive shell commands and file deletions.
// The policy enforces guardrails inside the tools themselves — hard blocks
// (deny-listed commands, protected paths, read-only mode) cannot be approved
// away.
func DefaultRegistry(ws *workspace.Workspace, approve ApprovalFunc, mem Store, policy Policy) *Registry {
	return DefaultRegistryWithAsk(ws, approve, nil, mem, policy)
}

// DefaultRegistryWithAsk is DefaultRegistry plus an optional ask_user tool.
func DefaultRegistryWithAsk(ws *workspace.Workspace, approve ApprovalFunc, ask AskUserFunc, mem Store, policy Policy) *Registry {
	r := NewRegistry()
	fs := &FilesystemTool{ws: ws, policy: policy}
	r.Register(fs)
	r.Register(&ShellTool{ws: ws, approve: approve, policy: policy})
	r.Register(&SearchTool{ws: ws})
	r.Register(&GitTool{ws: ws, approve: approve, policy: policy})
	r.Register(&TestTool{ws: ws, approve: approve, policy: policy})
	r.Register(&VersionsTool{ws: ws})
	if mem != nil {
		r.Register(&MemoryTool{store: mem})
	}
	if ask != nil {
		r.Register(&AskUserTool{Ask: ask})
	}
	return r
}

// Store is the memory surface used by the memory tool.
type Store interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context) ([]string, error)
}

// arg helpers ----------------------------------------------------------------

func argString(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func argInt(args map[string]interface{}, key string, def int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return int(n)
		}
	}
	return def
}

func argBool(args map[string]interface{}, key string) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return false
}
