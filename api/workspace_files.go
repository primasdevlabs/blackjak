package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type workspaceFileItem struct {
	Type string `json:"type"` // file | folder
	Path string `json:"path"`
	Name string `json:"name"`
}

// GET /api/workspace/files?q=&limit=
func (s *Server) handleWorkspaceFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	limit := 40
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	root := s.workspace.RootPath
	skipDirs := map[string]bool{
		".git": true, "node_modules": true, ".blackjak": true,
		"dist": true, "out": true, "bin": true, ".cursor": true,
	}

	var results []workspaceFileItem
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || len(results) >= limit {
			if len(results) >= limit {
				return filepath.SkipAll
			}
			return nil
		}
		name := d.Name()
		if d.IsDir() && skipDirs[name] {
			return filepath.SkipDir
		}
		if name == "." || strings.HasPrefix(name, ".") && name != ".env.example" {
			if d.IsDir() && skipDirs[name] {
				return filepath.SkipDir
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if q != "" && !strings.Contains(strings.ToLower(rel), q) && !strings.Contains(strings.ToLower(name), q) {
			return nil
		}
		itemType := "file"
		if d.IsDir() {
			itemType = "folder"
		}
		results = append(results, workspaceFileItem{
			Type: itemType,
			Path: rel,
			Name: name,
		})
		return nil
	})

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": results})
}
