package tools

import (
	"os"
	"path/filepath"
	"testing"

	"blackjak/workspace"
)

func TestFilesystemEdit_UniqueAndReplaceAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(path, []byte("foo\nbar\nfoo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ft := &FilesystemTool{ws: workspace.New(dir), policy: DefaultPolicy()}

	_, err := ft.edit(path, "foo", "baz", false)
	if err == nil {
		t.Fatal("expected uniqueness error")
	}

	res, err := ft.edit(path, "foo", "baz", true)
	if err != nil {
		t.Fatal(err)
	}
	m := res.(map[string]interface{})
	if m["replaced"].(int) != 2 {
		t.Fatalf("replaced=%v", m["replaced"])
	}
	data, _ := os.ReadFile(path)
	if string(data) != "baz\nbar\nbaz\n" {
		t.Fatalf("content=%q", data)
	}
}

func TestFilesystemEdit_ContextualExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	src := "package main\n\nfunc Hello() string { return \"hi\" }\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	ft := &FilesystemTool{ws: workspace.New(dir), policy: DefaultPolicy()}
	_, err := ft.edit(path, `func Hello() string { return "hi" }`, `func Hello() string { return "hey" }`, false)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !contains(string(data), `return "hey"`) {
		t.Fatalf("edit not applied: %s", data)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
