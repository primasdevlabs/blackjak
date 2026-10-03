package context

import (
	"regexp"
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

	TokensBefore int    `json:"tokensBefore"`
	TokensAfter  int    `json:"tokensAfter"`
	TokensSaved  int    `json:"tokensSaved"`
	Summary      string `json:"summary"`
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
	// Scale reserves to the real window so small (e.g. 32k) models still have
	// usable prompt budget and auto-compact triggers before the API rejects.
	reservedOut := maxTokens / 6
	if reservedOut < 2048 {
		reservedOut = 2048
	}
	if reservedOut > 16000 {
		reservedOut = 16000
	}
	reservedTools := maxTokens / 10
	if reservedTools < 1024 {
		reservedTools = 1024
	}
	if reservedTools > 8000 {
		reservedTools = 8000
	}
	return &Compactor{
		MaxContextWindow: maxTokens,
		ReservedOutput:   reservedOut,
		ReservedTools:    reservedTools,
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

// Compact builds a structured summary from objective + raw conversation text.
// Prefer LLM summarization in the agent loop; this is the deterministic fallback.
func (c *Compactor) Compact(objective string, rawContent string) CompactedContext {
	beforeTokens := estimateTokens(objective + rawContent)

	files := extractFileStates(rawContent)
	findings := extractLines(rawContent, []string{"found", "identified", "note:"}, 8)
	errors := extractLines(rawContent, []string{"error", "failed", "panic"}, 6)
	tests := extractLines(rawContent, []string{"test", "PASS", "FAIL"}, 6)
	decisions := extractLines(rawContent, []string{"decided", "will use", "chose"}, 6)
	pending := extractLines(rawContent, []string{"TODO", "next", "remaining"}, 6)

	summary := strings.TrimSpace(objective)
	if summary == "" {
		summary = "Continue the current engineering task."
	}
	if len(findings) > 0 {
		summary += "\nKey findings: " + strings.Join(clamp(findings, 3), "; ")
	}
	if len(files) > 0 {
		var paths []string
		for _, f := range clampFiles(files, 8) {
			paths = append(paths, f.Path)
		}
		summary += "\nTouched files: " + strings.Join(paths, ", ")
	}

	afterTokens := estimateTokens(summary)
	if afterTokens > beforeTokens {
		afterTokens = int(float64(beforeTokens) * 0.4)
	}

	return CompactedContext{
		Objective:    objective,
		Constraints:  []string{"Preserve working behavior", "Follow existing project conventions"},
		Decisions:    clamp(decisions, 8),
		Plan:         []string{"Resume from compacted state", "Finish remaining work", "Verify"},
		Files:        clampFiles(files, 12),
		Findings:     clamp(findings, 10),
		Tests:        clamp(tests, 8),
		Errors:       clamp(errors, 8),
		PendingWork:  clamp(pending, 8),
		TokensBefore: beforeTokens,
		TokensAfter:  afterTokens,
		TokensSaved:  beforeTokens - afterTokens,
		Summary:      summary,
	}
}

func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	// Rough chars/4 heuristic used elsewhere in the codebase.
	n := len(s) / 4
	if n < 1 {
		return 1
	}
	return n
}

var pathRe = regexp.MustCompile(`(?i)(?:[\w./\\-]+\.(?:go|ts|tsx|js|jsx|py|md|json|css|html|yml|yaml|toml|rs|java|kt))`)

func extractFileStates(raw string) []FileState {
	seen := map[string]bool{}
	var out []FileState
	for _, m := range pathRe.FindAllString(raw, 40) {
		p := filepathToSlash(m)
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, FileState{Path: p, Status: "M"})
	}
	return out
}

func filepathToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func extractLines(raw string, keywords []string, limit int) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		l := strings.TrimSpace(line)
		if l == "" || len(l) > 240 {
			continue
		}
		low := strings.ToLower(l)
		for _, kw := range keywords {
			if strings.Contains(low, strings.ToLower(kw)) {
				out = append(out, l)
				break
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func clamp(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func clampFiles(in []FileState, n int) []FileState {
	if len(in) <= n {
		return in
	}
	return in[:n]
}
