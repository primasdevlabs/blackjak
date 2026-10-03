package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Workspace tracks active workspace path and provides boundary security validation.
type Workspace struct {
	RootPath string `json:"rootPath"`
}

// New initializes a new Workspace instance with cleaned absolute root path.
func New(rootPath string) *Workspace {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		abs = rootPath
	}
	// Resolve the root itself through symlinks when possible.
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return &Workspace{
		RootPath: filepath.Clean(abs),
	}
}

// IsPathWithinWorkspace verifies that a given file path stays inside the workspace root.
func (w *Workspace) IsPathWithinWorkspace(targetPath string) bool {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	cleanTarget := filepath.Clean(absTarget)
	cleanRoot := filepath.Clean(w.RootPath)

	rel, err := filepath.Rel(cleanRoot, cleanTarget)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// ResolvePath converts a relative path to absolute within workspace or errors if outside.
// Symlink targets are evaluated so a link cannot escape the workspace root.
func (w *Workspace) ResolvePath(relOrAbsPath string) (string, error) {
	var target string
	if filepath.IsAbs(relOrAbsPath) {
		target = relOrAbsPath
	} else {
		target = filepath.Join(w.RootPath, relOrAbsPath)
	}
	target = filepath.Clean(target)

	resolved, err := resolveAgainstSymlinks(target)
	if err != nil {
		return "", err
	}

	if !w.IsPathWithinWorkspace(resolved) {
		return "", fmt.Errorf("security violation: path '%s' is outside workspace boundary '%s'", relOrAbsPath, w.RootPath)
	}

	return resolved, nil
}

// resolveAgainstSymlinks evaluates existing path components through symlinks.
// For paths that do not exist yet (e.g. new file writes), parent directories
// are resolved and the final segment is appended.
func resolveAgainstSymlinks(target string) (string, error) {
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		return filepath.Clean(resolved), nil
	} else if !os.IsNotExist(err) {
		// EvalSymlinks can fail for other reasons (permission); fall back to
		// cleaned path but still validate later via IsPathWithinWorkspace.
		return filepath.Clean(target), nil
	}

	// Walk up until a prefix exists, resolve it, then rejoin the remainder.
	cur := target
	var suffix []string
	for {
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		suffix = append([]string{filepath.Base(cur)}, suffix...)
		if resolved, err := filepath.EvalSymlinks(parent); err == nil {
			return filepath.Clean(filepath.Join(append([]string{resolved}, suffix...)...)), nil
		} else if !os.IsNotExist(err) {
			break
		}
		cur = parent
	}
	return filepath.Clean(target), nil
}
