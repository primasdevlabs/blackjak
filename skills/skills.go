package skills

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Skill is a slash-invocable instruction pack.
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Body        string   `json:"body,omitempty"`
	Enabled     bool     `json:"enabled"`
	Path        string   `json:"path"`
	Tools       []string `json:"tools,omitempty"`
	Source      string   `json:"source,omitempty"` // builtin | workspace
}

// Registry loads builtin skills (embedded in the agent) plus workspace skills
// from .blackjak/skills/<name>/SKILL.md. Workspace skills override builtins by name.
type Registry struct {
	mu      sync.RWMutex
	root    string
	skills  map[string]*Skill
	enabled map[string]bool
}

func NewRegistry(workspaceRoot string) *Registry {
	r := &Registry{
		root:    filepath.Join(workspaceRoot, ".blackjak", "skills"),
		skills:  map[string]*Skill{},
		enabled: map[string]bool{},
	}
	_ = r.Reload()
	return r
}

func (r *Registry) Reload() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills = map[string]*Skill{}

	for _, s := range BuiltinSkills() {
		cp := s
		cp.Source = "builtin"
		if v, ok := r.enabled[cp.Name]; ok {
			cp.Enabled = v
		}
		r.skills[cp.Name] = &cp
	}

	entries, err := os.ReadDir(r.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillPath := filepath.Join(r.root, e.Name(), "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		body := string(data)
		desc := firstNonEmptyLine(body)
		name := e.Name()
		enabled := true
		if v, ok := r.enabled[name]; ok {
			enabled = v
		}
		r.skills[name] = &Skill{
			Name:        name,
			Description: desc,
			Body:        body,
			Enabled:     enabled,
			Path:        skillPath,
			Source:      "workspace",
		}
	}
	return nil
}

func (r *Registry) List() []Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		cp := *s
		cp.Body = "" // omit body in list
		out = append(out, cp)
	}
	return out
}

func (r *Registry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[name]
	if !ok {
		return nil, false
	}
	cp := *s
	return &cp, true
}

func (r *Registry) SetEnabled(name string, enabled bool) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled[name] = enabled
	if s, ok := r.skills[name]; ok {
		s.Enabled = enabled
		return true
	}
	return false
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		line = strings.TrimSpace(line)
		if line != "" {
			if len(line) > 120 {
				return line[:120]
			}
			return line
		}
	}
	return "Skill"
}
