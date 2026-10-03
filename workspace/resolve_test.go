package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolvePath_BlocksDotDot(t *testing.T) {
	root := t.TempDir()
	ws := New(root)
	if _, err := ws.ResolvePath("../outside.txt"); err == nil {
		t.Fatal("expected outside path rejection")
	}
}

func TestResolvePath_SymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Creating symlinks often requires elevation on Windows CI.
		t.Skip("symlink escape test skipped on windows")
	}
	root := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("leak"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	ws := New(root)
	if _, err := ws.ResolvePath("escape/secret.txt"); err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}

func TestResolvePath_InsideOK(t *testing.T) {
	root := t.TempDir()
	f := filepath.Join(root, "ok.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ws := New(root)
	got, err := ws.ResolvePath("ok.txt")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "ok.txt" {
		t.Fatalf("got %s", got)
	}
}
