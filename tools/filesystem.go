package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"blackjak/workspace"
)

const maxReadBytes = 256 * 1024

// FilesystemTool enables reading, writing, editing and listing files within
// the workspace boundary.
type FilesystemTool struct {
	ws *workspace.Workspace
}

func (f *FilesystemTool) Name() string { return "filesystem" }

func (f *FilesystemTool) Description() string {
	return "Read, write, edit, and list files. Paths are resolved relative to the workspace root and cannot escape it."
}

func (f *FilesystemTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"read", "write", "edit", "list", "exists"},
				"description": "read: file contents; write: create/overwrite; edit: replace exact text; list: directory tree; exists: check presence",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Workspace-relative path",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Full file contents for write",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "Exact text to replace for edit",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "Replacement text for edit",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "Line offset for read (0-based)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Max lines for read/list (default 500)",
			},
		},
		"required": []string{"operation", "path"},
	}
}

func (f *FilesystemTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	op := argString(args, "operation")
	path := argString(args, "path")
	if path == "" && op != "list" {
		return nil, fmt.Errorf("path is required")
	}

	abs, err := f.ws.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	switch op {
	case "read":
		return f.read(abs, argInt(args, "offset", 0), argInt(args, "limit", 500))
	case "write":
		return f.write(abs, argString(args, "content"))
	case "edit":
		return f.edit(abs, argString(args, "old_string"), argString(args, "new_string"))
	case "list":
		return f.list(abs, argInt(args, "limit", 500))
	case "exists":
		_, err := os.Stat(abs)
		return map[string]interface{}{"exists": err == nil}, nil
	default:
		return nil, fmt.Errorf("unknown filesystem operation: %s", op)
	}
}

func (f *FilesystemTool) read(abs string, offset, limit int) (interface{}, error) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	if len(data) > maxReadBytes {
		data = data[:maxReadBytes]
	}
	lines := strings.Split(string(data), "\n")
	total := len(lines)
	if offset > 0 {
		if offset >= total {
			return map[string]interface{}{"content": "", "totalLines": total, "truncated": true}, nil
		}
		lines = lines[offset:]
	}
	truncated := false
	if limit > 0 && len(lines) > limit {
		lines = lines[:limit]
		truncated = true
	}
	return map[string]interface{}{
		"content":    strings.Join(lines, "\n"),
		"totalLines": total,
		"truncated":  truncated || len(data) == maxReadBytes,
	}, nil
}

func (f *FilesystemTool) write(abs, content string) (interface{}, error) {
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(f.ws.RootPath, abs)
	return map[string]interface{}{"path": filepath.ToSlash(rel), "bytes": len(content)}, nil
}

func (f *FilesystemTool) edit(abs, oldStr, newStr string) (interface{}, error) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	text := string(data)
	count := strings.Count(text, oldStr)
	if count == 0 {
		return nil, fmt.Errorf("old_string not found in file")
	}
	if count > 1 {
		return nil, fmt.Errorf("old_string matches %d locations; provide more context for a unique match", count)
	}
	if err := os.WriteFile(abs, []byte(strings.Replace(text, oldStr, newStr, 1)), 0o644); err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(f.ws.RootPath, abs)
	return map[string]interface{}{"path": filepath.ToSlash(rel), "replaced": 1}, nil
}

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "dist": true,
	"out": true, "bin": true, ".blackjak": true, "__pycache__": true,
}

func (f *FilesystemTool) list(root string, limit int) (interface{}, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		rel, _ := filepath.Rel(f.ws.RootPath, root)
		return map[string]interface{}{"entries": []string{filepath.ToSlash(rel)}}, nil
	}
	var entries []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if len(entries) >= limit {
			return filepath.SkipAll
		}
		if info.IsDir() && path != root && skipDirs[info.Name()] {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(f.ws.RootPath, path)
		if info.IsDir() {
			entries = append(entries, filepath.ToSlash(rel)+"/")
		} else {
			entries = append(entries, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	return map[string]interface{}{"entries": entries, "truncated": len(entries) >= limit}, nil
}
