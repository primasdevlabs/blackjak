package agent

import (
	"fmt"

	"blackjak/llm"
)

// ensureToolCallResults appends synthetic tool results for any assistant
// tool_calls that lack a matching tool message. Prevents resume from sending
// an incomplete batch to providers that require one result per call.
func ensureToolCallResults(messages []llm.Message) []llm.Message {
	if len(messages) == 0 {
		return messages
	}
	answered := map[string]bool{}
	for _, m := range messages {
		if m.Role == llm.RoleTool && m.ToolCallID != "" {
			answered[m.ToolCallID] = true
		}
	}
	var out []llm.Message
	out = append(out, messages...)
	for _, m := range messages {
		if m.Role != llm.RoleAssistant || len(m.ToolCalls) == 0 {
			continue
		}
		for _, call := range m.ToolCalls {
			if call.ID == "" || answered[call.ID] {
				continue
			}
			out = append(out, llm.Message{
				Role:       llm.RoleTool,
				Name:       call.Name,
				ToolCallID: call.ID,
				Content:    fmt.Sprintf(`{"error":"interrupted before tool completed","tool":%q,"resumable":true}`, call.Name),
			})
			answered[call.ID] = true
		}
	}
	return out
}
