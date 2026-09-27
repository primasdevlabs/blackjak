package tools

import (
	"context"
	"fmt"
)

// MemoryTool exposes the agent's persistent memory store to the LLM.
type MemoryTool struct {
	store Store
}

func (m *MemoryTool) Name() string { return "memory" }

func (m *MemoryTool) Description() string {
	return "Read and write persistent memory that survives across sessions. Use it to remember project facts, user preferences, and decisions."
}

func (m *MemoryTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"get", "set", "delete", "list"},
				"description": "Memory operation",
			},
			"key": map[string]interface{}{
				"type":        "string",
				"description": "Memory key (required for get/set/delete)",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Value to store (required for set)",
			},
		},
		"required": []string{"operation"},
	}
}

func (m *MemoryTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	op := argString(args, "operation")
	key := argString(args, "key")

	switch op {
	case "get":
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}
		v, err := m.store.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"key": key, "value": v, "found": v != nil}, nil
	case "set":
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}
		value := argString(args, "value")
		if err := m.store.Set(ctx, key, value); err != nil {
			return nil, err
		}
		return map[string]interface{}{"key": key, "stored": true}, nil
	case "delete":
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}
		if err := m.store.Delete(ctx, key); err != nil {
			return nil, err
		}
		return map[string]interface{}{"key": key, "deleted": true}, nil
	case "list":
		keys, err := m.store.Keys(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"keys": keys}, nil
	default:
		return nil, fmt.Errorf("unknown memory operation: %s", op)
	}
}
