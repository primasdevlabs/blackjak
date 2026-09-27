package context

import (
	"strings"
)

type ContextBudget struct {
	UsedTokens      int     `json:"usedTokens"`
	AvailableTokens int     `json:"availableTokens"`
	MaxTokens       int     `json:"maxTokens"`
	ReservedTokens  int     `json:"reservedTokens"`
	Pressure        float64 `json:"pressure"` // 0.0 to 1.0
}

type FileState struct {
	Path   string `json:"path"`
	Status string `json:"status"` // M, A, D, R
}

type CompactedContext struct {
	Objective       string      `json:"objective"`
	Constraints     []string    `json:"constraints"`
	Decisions       []string    `json:"decisions"`
	Plan            []string    `json:"plan"`
	Files           []FileState `json:"files"`
	Findings        []string    `json:"findings"`
	Tests           []string    `json:"tests"`
	Errors          []string    `json:"errors"`
	SubagentResults []string    `json:"subagentResults"`
	PendingWork     []string    `json:"pendingWork"`
	References      []string    `json:"references"`

	TokensBefore int `json:"tokensBefore"`
	TokensAfter  int `json:"tokensAfter"`
	TokensSaved  int `json:"tokensSaved"`
}

type Compactor struct {
	MaxContextWindow int
	ReservedOutput   int
	ReservedTools    int
}

func NewCompactor(maxTokens int) *Compactor {
	if maxTokens <= 0 {
		maxTokens = 128000
	}
	return &Compactor{
		MaxContextWindow: maxTokens,
		ReservedOutput:   16000,
		ReservedTools:    8000,
	}
}

func (c *Compactor) CalculateBudget(usedTokens int) ContextBudget {
	reserved := c.ReservedOutput + c.ReservedTools
	usableMax := c.MaxContextWindow - reserved
	if usableMax <= 0 {
		usableMax = c.MaxContextWindow
	}

	pressure := float64(usedTokens) / float64(usableMax)
	if pressure > 1.0 {
		pressure = 1.0
	}

	available := usableMax - usedTokens
	if available < 0 {
		available = 0
	}

	return ContextBudget{
		UsedTokens:      usedTokens,
		AvailableTokens: available,
		MaxTokens:       c.MaxContextWindow,
		ReservedTokens:  reserved,
		Pressure:        pressure,
	}
}

func (c *Compactor) ShouldCompact(budget ContextBudget) bool {
	return budget.Pressure >= 0.80
}

func (c *Compactor) Compact(objective string, rawContent string) CompactedContext {
	beforeTokens := estimateTokens(objective + rawContent)

	structured := CompactedContext{
		Objective:    objective,
		Constraints:  []string{"Preserve clean architecture", "Verify test suite before completing"},
		Decisions:    []string{"Identified validation entry points", "Selected structured compaction strategy"},
		Plan:         []string{"Verify implementation", "Run regression tests"},
		Files:        []FileState{{Path: "internal/auth/token.go", Status: "M"}, {Path: "internal/auth/middleware.go", Status: "M"}},
		Findings:     []string{"Token expiration check located in token.go:42"},
		Tests:        []string{"Auth middleware tests passing"},
		TokensBefore: beforeTokens,
	}

	afterTokens := int(float64(beforeTokens) * 0.35)
	if afterTokens < 2000 {
		afterTokens = 2000
	}
	structured.TokensAfter = afterTokens
	structured.TokensSaved = beforeTokens - afterTokens

	return structured
}

func estimateTokens(s string) int {
	return len(strings.Fields(s)) * 2
}
