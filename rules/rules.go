package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Rule is an always-on instruction block.
type Rule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Source  string `json:"source"` // builtin | user | project
	Enabled bool   `json:"enabled"`
	Path    string `json:"path,omitempty"`
}

type rulePrefs struct {
	Enabled map[string]bool `json:"enabled"`
}

// Store manages builtin + user + project rules.
type Store struct {
	mu        sync.RWMutex
	workspace string
	userRules []Rule
	userPath  string
	prefsPath string
	enabled   map[string]bool // keyed by rule id
}

func NewStore(workspaceRoot string) *Store {
	s := &Store{
		workspace: workspaceRoot,
		userPath:  filepath.Join(workspaceRoot, ".blackjak", "user-rules.json"),
		prefsPath: filepath.Join(workspaceRoot, ".blackjak", "rule-prefs.json"),
		enabled:   map[string]bool{},
	}
	_ = s.loadUser()
	_ = s.loadPrefs()
	return s
}

func (s *Store) loadUser() error {
	data, err := os.ReadFile(s.userPath)
	if err != nil {
		return err
	}
	var rules []Rule
	if err := decodeJSON(data, &rules); err != nil {
		return err
	}
	s.userRules = rules
	return nil
}

func (s *Store) saveUser() error {
	_ = os.MkdirAll(filepath.Dir(s.userPath), 0o755)
	return writeJSON(s.userPath, s.userRules)
}

func (s *Store) loadPrefs() error {
	data, err := os.ReadFile(s.prefsPath)
	if err != nil {
		return err
	}
	var prefs rulePrefs
	if err := decodeJSON(data, &prefs); err != nil {
		return err
	}
	if prefs.Enabled == nil {
		prefs.Enabled = map[string]bool{}
	}
	s.enabled = prefs.Enabled
	return nil
}

func (s *Store) savePrefs() error {
	_ = os.MkdirAll(filepath.Dir(s.prefsPath), 0o755)
	return writeJSON(s.prefsPath, rulePrefs{Enabled: s.enabled})
}

func (s *Store) isEnabled(id string, fallback bool) bool {
	if v, ok := s.enabled[id]; ok {
		return v
	}
	return fallback
}

func (s *Store) projectRulesLocked() []Rule {
	dir := filepath.Join(s.workspace, ".blackjak", "rules")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []Rule
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		id := "project:" + name
		out = append(out, Rule{
			ID:      id,
			Name:    name,
			Body:    string(data),
			Source:  "project",
			Enabled: s.isEnabled(id, true),
			Path:    path,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Store) ProjectRules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projectRulesLocked()
}

// List returns builtin rules, then project overrides by name, then user rules.
// Project rules with the same name as a builtin replace that builtin.
func (s *Store) List() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	byName := map[string]Rule{}
	order := []string{}

	for _, r := range BuiltinRules() {
		r.Enabled = s.isEnabled(r.ID, true)
		byName[r.Name] = r
		order = append(order, r.Name)
	}
	for _, r := range s.projectRulesLocked() {
		if _, exists := byName[r.Name]; !exists {
			order = append(order, r.Name)
		}
		byName[r.Name] = r
	}

	out := make([]Rule, 0, len(order)+len(s.userRules))
	for _, name := range order {
		out = append(out, byName[name])
	}
	for _, r := range s.userRules {
		fallback := true
		if !r.Enabled {
			fallback = false
		}
		r.Enabled = s.isEnabled(r.ID, fallback)
		out = append(out, r)
	}
	return out
}

// Get returns a single rule by id.
func (s *Store) Get(id string) (Rule, bool) {
	for _, r := range s.List() {
		if r.ID == id {
			return r, true
		}
	}
	return Rule{}, false
}

func (s *Store) ruleExistsLocked(id string) bool {
	for _, r := range BuiltinRules() {
		if r.ID == id {
			return true
		}
	}
	for _, r := range s.projectRulesLocked() {
		if r.ID == id {
			return true
		}
	}
	for _, r := range s.userRules {
		if r.ID == id {
			return true
		}
	}
	return false
}

// SetEnabled toggles a rule without deleting its body.
func (s *Store) SetEnabled(id string, enabled bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" || !s.ruleExistsLocked(id) {
		return false
	}
	s.enabled[id] = enabled
	for i, r := range s.userRules {
		if r.ID == id {
			s.userRules[i].Enabled = enabled
			_ = s.saveUser()
			break
		}
	}
	_ = s.savePrefs()
	return true
}

func (s *Store) upsertUserLocked(rule Rule) Rule {
	rule.Source = "user"
	rule.Name = strings.TrimSpace(rule.Name)
	if rule.Name == "" {
		rule.Name = "untitled"
	}
	if rule.ID == "" || !strings.HasPrefix(rule.ID, "user:") {
		rule.ID = "user:" + slug(rule.Name)
	}
	if v, ok := s.enabled[rule.ID]; ok {
		rule.Enabled = v
	} else {
		rule.Enabled = true
		s.enabled[rule.ID] = true
	}
	found := false
	for i, r := range s.userRules {
		if r.ID == rule.ID {
			s.userRules[i] = rule
			found = true
			break
		}
	}
	if !found {
		s.userRules = append(s.userRules, rule)
	}
	_ = s.saveUser()
	_ = s.savePrefs()
	return rule
}

func (s *Store) UpsertUser(rule Rule) Rule {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.upsertUserLocked(rule)
}

func (s *Store) upsertProjectLocked(name, body string) (Rule, error) {
	name = slug(name)
	if name == "" {
		return Rule{}, fmt.Errorf("rule name required")
	}
	dir := filepath.Join(s.workspace, ".blackjak", "rules")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Rule{}, err
	}
	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return Rule{}, err
	}
	id := "project:" + name
	if _, ok := s.enabled[id]; !ok {
		s.enabled[id] = true
		_ = s.savePrefs()
	}
	// Prefer project id over builtin when enabling by name.
	return Rule{
		ID:      id,
		Name:    name,
		Body:    body,
		Source:  "project",
		Enabled: s.isEnabled(id, true),
		Path:    path,
	}, nil
}

// UpsertProject writes/updates `.blackjak/rules/<name>.md`.
func (s *Store) UpsertProject(name, body string) (Rule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.upsertProjectLocked(name, body)
}

// Upsert creates or updates a rule. source "project" (or builtin/project ids)
// writes a markdown file; otherwise stores a user rule.
func (s *Store) Upsert(rule Rule) (Rule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	source := strings.ToLower(strings.TrimSpace(rule.Source))
	if source == "project" || strings.HasPrefix(rule.ID, "project:") || strings.HasPrefix(rule.ID, "builtin:") {
		name := rule.Name
		if name == "" && strings.Contains(rule.ID, ":") {
			name = strings.SplitN(rule.ID, ":", 2)[1]
		}
		return s.upsertProjectLocked(name, rule.Body)
	}
	return s.upsertUserLocked(rule), nil
}

func (s *Store) deleteUserLocked(id string) bool {
	for i, r := range s.userRules {
		if r.ID == id {
			s.userRules = append(s.userRules[:i], s.userRules[i+1:]...)
			delete(s.enabled, id)
			_ = s.saveUser()
			_ = s.savePrefs()
			return true
		}
	}
	return false
}

func (s *Store) DeleteUser(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deleteUserLocked(id)
}

// Delete removes a user rule or a project .md file. Builtins cannot be deleted.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.HasPrefix(id, "builtin:") {
		return false
	}
	if strings.HasPrefix(id, "project:") {
		name := strings.TrimPrefix(id, "project:")
		path := filepath.Join(s.workspace, ".blackjak", "rules", name+".md")
		if err := os.Remove(path); err != nil {
			return false
		}
		delete(s.enabled, id)
		_ = s.savePrefs()
		return true
	}
	return s.deleteUserLocked(id)
}

// AssembleEnabled returns concatenated enabled rule bodies for system prompt.
func (s *Store) AssembleEnabled() string {
	var parts []string
	for _, r := range s.List() {
		if !r.Enabled || strings.TrimSpace(r.Body) == "" {
			continue
		}
		parts = append(parts, "## Rule: "+r.Name+"\n"+strings.TrimSpace(r.Body))
	}
	return strings.Join(parts, "\n\n")
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, s)
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}
