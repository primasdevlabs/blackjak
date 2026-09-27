package workspace

import (
	"fmt"
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
func (w *Workspace) ResolvePath(relOrAbsPath string) (string, error) {
	var target string
	if filepath.IsAbs(relOrAbsPath) {
		target = relOrAbsPath
	} else {
		target = filepath.Join(w.RootPath, relOrAbsPath)
	}

	if !w.IsPathWithinWorkspace(target) {
		return "", fmt.Errorf("security violation: path '%s' is outside workspace boundary '%s'", relOrAbsPath, w.RootPath)
	}

	return filepath.Clean(target), nil
}
