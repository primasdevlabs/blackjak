package context

import (
	"fmt"
	"strings"
)

// BuildInput is everything available when assembling a run prompt.
type BuildInput struct {
	SystemPrompt string
	UserRules    string
	ProjectRules string
	SkillBody    string
	ModePolicy   string
	Memory       string
	OpenFiles    []string
	References   []string
	Retrieved    []string
	RecentDiffs  string
	UserPrompt   string
}

// BuiltContext is the assembled message payload for the LLM.
type BuiltContext struct {
	System string
	User   string
	Parts  map[string]int // approximate char counts per section
}

// Builder constructs prompts and context payloads for LLM requests.
type Builder struct{}

// NewBuilder creates a new context Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build assembles system + user messages with rules, skills, memory, refs.
func (b *Builder) Build(in BuildInput) BuiltContext {
	var sysParts []string
	if strings.TrimSpace(in.SystemPrompt) != "" {
		sysParts = append(sysParts, strings.TrimSpace(in.SystemPrompt))
	}
	if strings.TrimSpace(in.UserRules) != "" {
		sysParts = append(sysParts, "# User Rules\n"+strings.TrimSpace(in.UserRules))
	}
	if strings.TrimSpace(in.ProjectRules) != "" {
		sysParts = append(sysParts, "# Project Rules\n"+strings.TrimSpace(in.ProjectRules))
	}
	if strings.TrimSpace(in.ModePolicy) != "" {
		sysParts = append(sysParts, "# Mode Policy\n"+strings.TrimSpace(in.ModePolicy))
	}
	if strings.TrimSpace(in.SkillBody) != "" {
		sysParts = append(sysParts, "# Active Skill\n"+strings.TrimSpace(in.SkillBody))
	}
	if strings.TrimSpace(in.Memory) != "" {
		sysParts = append(sysParts, "# Persistent Memory\n"+strings.TrimSpace(in.Memory))
	}

	var userParts []string
	if len(in.References) > 0 {
		userParts = append(userParts, "Referenced paths:\n- "+strings.Join(in.References, "\n- "))
	}
	if len(in.OpenFiles) > 0 {
		userParts = append(userParts, "Open editors:\n- "+strings.Join(in.OpenFiles, "\n- "))
	}
	if len(in.Retrieved) > 0 {
		userParts = append(userParts, "Retrieved context:\n- "+strings.Join(in.Retrieved, "\n- "))
	}
	if strings.TrimSpace(in.RecentDiffs) != "" {
		userParts = append(userParts, "Recent diffs:\n"+strings.TrimSpace(in.RecentDiffs))
	}
	userParts = append(userParts, in.UserPrompt)

	system := strings.Join(sysParts, "\n\n")
	user := strings.Join(userParts, "\n\n")
	parts := map[string]int{
		"system":     len(system),
		"userRules":  len(in.UserRules),
		"projectRules": len(in.ProjectRules),
		"skill":      len(in.SkillBody),
		"memory":     len(in.Memory),
		"references": len(strings.Join(in.References, "")),
		"retrieved":  len(strings.Join(in.Retrieved, "")),
		"prompt":     len(in.UserPrompt),
		"total":      len(system) + len(user),
	}
	return BuiltContext{System: system, User: user, Parts: parts}
}

// ModePolicyText returns mode-specific instructions.
func ModePolicyText(mode string) string {
	switch mode {
	case "ask":
		return "ASK mode: answer questions using read-only tools. Do not modify files or run mutating shell commands."
	case "plan":
		return "PLAN mode: explore and produce a clear plan. Prefer read-only tools; do not implement changes until the user switches to Agent mode."
	default:
		return "AGENT mode: implement the task end-to-end using tools. Make minimal correct changes and verify."
	}
}

// AskPolicyText returns LLM-facing instructions for request refusal behavior.
// Tool guardrails are enforced separately and are not relaxed by this text.
func AskPolicyText(policy string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "accept_all":
		return strings.TrimSpace(`Request policy: ACCEPT ALL (security research).
The operator enabled Accept All for authorized offensive security work — red teaming, penetration testing, vulnerability research, reverse engineering, exploit analysis, and related lab/engagement tasks on systems they own or are contracted to assess.

Instructions:
- Do not refuse dual-use cybersecurity topics. Provide concrete technical assistance.
- Do not moralize, lecture, or append unnecessary capability disclaimers.
- Treat the workspace and stated engagement targets as authorized unless the user clearly asks to attack unrelated third parties with no authorization.
- Prefer actionable detail (commands, payloads, analysis steps, mitigation gaps) over high-level refusals.
- Tool guardrails (deny-listed commands, protected paths, approval prompts) still apply — this policy only changes conversational refusal behavior.
- Still refuse requests involving child sexual exploitation.`)
	default:
		return strings.TrimSpace(`Request policy: STANDARD.
Refuse clearly harmful or unauthorized requests. Prefer safer alternatives when asked to do something reckless.
Dual-use security topics are fine when framed as hardening, education, or authorized assessment of this workspace.`)
	}
}

// FormatCharBudget returns a short status line.
func FormatCharBudget(parts map[string]int) string {
	return fmt.Sprintf("context ~%d chars", parts["total"]/4*4)
}
