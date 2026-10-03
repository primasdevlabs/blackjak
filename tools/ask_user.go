package tools

import (
	"context"
	"fmt"
	"strings"
)

// AskUserFunc prompts the human with a multiple-choice (or free-text) question
// and returns the selected answer. Implementations typically block on the UI.
type AskUserFunc func(ctx context.Context, question string, options []string, allowOther bool) (string, error)

// AskUserTool lets the model pause and ask the user a walkthrough-style question.
type AskUserTool struct {
	Ask AskUserFunc
}

func (t *AskUserTool) Name() string { return "ask_user" }

func (t *AskUserTool) Description() string {
	return "Ask the human a clarifying walkthrough question with clickable options (and optional free-text Other). Use when a choice materially changes the plan."
}

func (t *AskUserTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"question": map[string]interface{}{
				"type":        "string",
				"description": "The question to present to the user",
			},
			"options": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "2–6 short answer choices (e.g. 'a) Hybrid (recommended): …')",
			},
			"allowOther": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, user can type a custom answer (default true)",
			},
		},
		"required": []string{"question", "options"},
	}
}

func (t *AskUserTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	if t.Ask == nil {
		return nil, fmt.Errorf("ask_user is unavailable in this host")
	}
	question := strings.TrimSpace(argString(args, "question"))
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	var options []string
	if raw, ok := args["options"].([]interface{}); ok {
		for _, o := range raw {
			if s, ok := o.(string); ok && strings.TrimSpace(s) != "" {
				options = append(options, strings.TrimSpace(s))
			}
		}
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("at least one option is required")
	}
	if len(options) > 8 {
		options = options[:8]
	}
	allowOther := true
	if v, ok := args["allowOther"].(bool); ok {
		allowOther = v
	}
	answer, err := t.Ask(ctx, question, options, allowOther)
	if err != nil {
		return nil, err
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return nil, fmt.Errorf("no answer provided")
	}
	return map[string]interface{}{
		"question": question,
		"answer":   answer,
	}, nil
}
