package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltinRules_Present(t *testing.T) {
	got := BuiltinRules()
	if len(got) < 5 {
		t.Fatalf("expected at least 5 builtin rules, got %d", len(got))
	}
	names := map[string]bool{}
	for _, r := range got {
		names[r.Name] = true
		if r.Source != "builtin" {
			t.Errorf("%s source = %q", r.Name, r.Source)
		}
		if strings.TrimSpace(r.Body) == "" {
			t.Errorf("%s empty body", r.Name)
		}
	}
	for _, want := range []string{"engineering", "anti-slop", "go", "security", "ui"} {
		if !names[want] {
			t.Errorf("missing builtin rule %q", want)
		}
	}
}

func TestStore_AssemblesBuiltinsWithoutWorkspaceRules(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	text := s.AssembleEnabled()
	if !strings.Contains(text, "Engineering judgment") {
		t.Fatalf("expected builtin engineering rule in assemble, got:\n%s", text)
	}
	if !strings.Contains(text, "Anti-AI-slop") {
		t.Fatal("expected anti-slop")
	}
}

func TestStore_ProjectOverridesBuiltin(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".blackjak", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(rulesDir, "engineering.md")
	if err := os.WriteFile(path, []byte("# Custom engineering\nworkspace override"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewStore(dir)
	var eng *Rule
	for _, r := range s.List() {
		if r.Name == "engineering" {
			cp := r
			eng = &cp
			break
		}
	}
	if eng == nil {
		t.Fatal("engineering missing")
	}
	if eng.Source != "project" {
		t.Fatalf("want project override, got %s", eng.Source)
	}
	if !strings.Contains(eng.Body, "workspace override") {
		t.Fatalf("body not overridden: %s", eng.Body)
	}
}

func TestStore_UpsertEditEnable(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	saved, err := s.Upsert(Rule{Name: "api", Body: "use REST", Source: "project"})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Source != "project" || saved.ID != "project:api" {
		t.Fatalf("unexpected saved rule: %+v", saved)
	}

	if !s.SetEnabled(saved.ID, false) {
		t.Fatal("SetEnabled failed")
	}
	got, ok := s.Get(saved.ID)
	if !ok || got.Enabled {
		t.Fatalf("expected disabled, got %+v", got)
	}

	edited, err := s.Upsert(Rule{ID: "builtin:ui", Name: "ui", Body: "custom ui rule", Source: "project"})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Source != "project" || !strings.Contains(edited.Body, "custom ui") {
		t.Fatalf("builtin edit should become project override: %+v", edited)
	}

	if !s.Delete(saved.ID) {
		t.Fatal("delete project rule failed")
	}
	if _, ok := s.Get(saved.ID); ok {
		t.Fatal("deleted rule still listed")
	}
}
