package api

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GET /api/diffs/<path> — returns a unified diff (git) or file contents for browser download.
func (s *Server) handleDiffDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/api/diffs/")
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || strings.Contains(rel, "..") {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	rel = filepath.ToSlash(rel)
	full := filepath.Join(s.workspace.RootPath, filepath.FromSlash(rel))
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(s.workspace.RootPath)) {
		writeError(w, http.StatusForbidden, "path outside workspace")
		return
	}

	cmd := exec.Command("git", "diff", "--", rel)
	cmd.Dir = s.workspace.RootPath
	out, err := cmd.CombinedOutput()
	body := string(out)
	if err != nil || strings.TrimSpace(body) == "" {
		data, readErr := os.ReadFile(full)
		if readErr != nil {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		body = string(data)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(rel)+".diff.txt\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}
