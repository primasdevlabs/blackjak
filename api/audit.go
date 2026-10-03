package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"blackjak/agent"
)

// auditActivityPayload builds a concise, redacted activity message + data blob.
func auditActivityPayload(typ agent.EventType, data any) (string, any) {
	m, _ := data.(map[string]interface{})
	switch typ {
	case agent.EventToolStarted:
		tool, _ := m["tool"].(string)
		args := redactMap(asMap(m["args"]))
		msg := fmt.Sprintf("tool %s started", tool)
		return msg, map[string]interface{}{"tool": tool, "args": args, "outcome": "started"}
	case agent.EventToolCompleted:
		tool, _ := m["tool"].(string)
		return fmt.Sprintf("tool %s completed", tool), map[string]interface{}{"tool": tool, "outcome": "ok"}
	case agent.EventToolFailed:
		tool, _ := m["tool"].(string)
		errStr, _ := m["error"].(string)
		return fmt.Sprintf("tool %s failed: %s", tool, truncate(errStr, 160)), map[string]interface{}{
			"tool": tool, "outcome": "error", "error": truncate(errStr, 240),
		}
	case agent.EventAgentPhase:
		phase, _ := m["phase"].(string)
		taskClass, _ := m["taskClass"].(string)
		return fmt.Sprintf("phase %s (%s)", phase, taskClass), m
	default:
		raw, _ := json.Marshal(data)
		return truncate(string(raw), 240), nil
	}
}

func asMap(v any) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func redactMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "secret") || strings.Contains(lk, "password") ||
			strings.Contains(lk, "token") || strings.Contains(lk, "apikey") ||
			strings.Contains(lk, "api_key") || strings.Contains(lk, "authorization") ||
			lk == "content" || lk == "old_string" || lk == "new_string" {
			out[k] = "[redacted]"
			continue
		}
		if s, ok := v.(string); ok && len(s) > 200 {
			out[k] = truncate(s, 200)
			continue
		}
		out[k] = v
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
