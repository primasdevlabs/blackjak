package api

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type workspaceSymbolItem struct {
	Type string `json:"type"` // symbol
	Path string `json:"path"`
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"`
	Line int    `json:"line,omitempty"`
}

var symbolPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_][\w]*)`),
	regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_][\w]*)\s*=\s*(?:async\s*)?\(`),
	regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:default\s+)?class\s+([A-Za-z_][\w]*)`),
	regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:type|interface|enum)\s+([A-Za-z_][\w]*)`),
	regexp.MustCompile(`(?m)^\s*func\s+(?:\([^)]*\)\s*)?([A-Za-z_][\w]*)\s*\(`),
	regexp.MustCompile(`(?m)^\s*type\s+([A-Za-z_][\w]*)\s+struct\b`),
	regexp.MustCompile(`(?m)^\s*(?:pub\s+)?(?:async\s+)?fn\s+([A-Za-z_][\w]*)`),
	regexp.MustCompile(`(?m)^\s*(?:def|class)\s+([A-Za-z_][\w]*)`),
}

// GET /api/workspace/symbols?q=&limit=
func (s *Server) handleWorkspaceSymbols(w http.ResponseWriter, r *http.Request) {
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
		"dist": true, "out": true, "bin": true, "vendor": true,
	}
	extOK := map[string]bool{
		".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
		".py": true, ".rs": true, ".java": true, ".kt": true,
	}

	var results []workspaceSymbolItem
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || len(results) >= limit {
			if len(results) >= limit {
				return filepath.SkipAll
			}
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !extOK[ext] {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 256*1024)
		lineNo := 0
		for sc.Scan() {
			lineNo++
			line := sc.Text()
			for _, re := range symbolPatterns {
				m := re.FindStringSubmatch(line)
				if len(m) < 2 {
					continue
				}
				name := m[1]
				if q != "" && !strings.Contains(strings.ToLower(name), q) && !strings.Contains(strings.ToLower(rel), q) {
					continue
				}
				kind := "symbol"
				low := strings.ToLower(line)
				switch {
				case strings.Contains(low, "class "):
					kind = "class"
				case strings.Contains(low, "interface "), strings.Contains(low, "type "):
					kind = "type"
				case strings.Contains(low, "func "), strings.Contains(low, "function "), strings.Contains(low, " fn "), strings.Contains(low, "def "):
					kind = "function"
				}
				results = append(results, workspaceSymbolItem{
					Type: "symbol",
					Path: rel + ":" + strconv.Itoa(lineNo),
					Name: name,
					Kind: kind,
					Line: lineNo,
				})
				if len(results) >= limit {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": results})
}
