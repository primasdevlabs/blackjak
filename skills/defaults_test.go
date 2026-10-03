package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinSkills_Present(t *testing.T) {
	got := BuiltinSkills()
	names := map[string]bool{}
	for _, s := range got {
		names[s.Name] = true
	}
	for _, want := range []string{"review", "verify", "explain"} {
		if !names[want] {
			t.Errorf("missing builtin skill %q (have %#v)", want, names)
		}
	}
}

func TestRegistry_BuiltinsWithoutWorkspace(t *testing.T) {
	r := NewRegistry(t.TempDir())
	sk, ok := r.Get("review")
	if !ok || sk == nil {
		t.Fatal("expected builtin review skill")
	}
	if sk.Source != "builtin" {
		t.Fatalf("source = %q", sk.Source)
	}
}

func TestRegistry_WorkspaceOverridesBuiltin(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".blackjak", "skills", "review")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Workspace review\noverride"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewRegistry(dir)
	sk, ok := r.Get("review")
	if !ok {
		t.Fatal("missing review")
	}
	if sk.Source != "workspace" {
		t.Fatalf("want workspace, got %s", sk.Source)
	}
	if sk.Body != "# Workspace review\noverride" {
		t.Fatalf("body = %q", sk.Body)
	}
}
