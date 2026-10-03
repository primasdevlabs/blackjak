package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blackjak/rules"
	"blackjak/skills"
)

// GET /api/skills
func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if s.skillsRegistry == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": []any{}})
		return
	}
	_ = s.skillsRegistry.Reload()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.skillsRegistry.List()})
}

// /api/skills/:name ...
func (s *Server) handleSkillsSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/skills/")
	parts := strings.Split(path, "/")
	name := parts[0]
	if name == "" {
		writeError(w, http.StatusBadRequest, "skill name required")
		return
	}
	if len(parts) >= 2 && parts[1] == "enable" && r.Method == http.MethodPost {
		var body struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		ok := s.skillsRegistry != nil && s.skillsRegistry.SetEnabled(name, body.Enabled)
		if !ok {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
		return
	}
	if r.Method == http.MethodGet {
		sk, ok := s.skillsRegistry.Get(name)
		if !ok {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": sk})
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// GET/POST /api/rules
func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	if s.rulesStore == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": []any{}})
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.rulesStore.List()})
	case http.MethodPost:
		var rule rules.Rule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			writeError(w, http.StatusBadRequest, "invalid rule")
			return
		}
		if strings.TrimSpace(rule.Source) == "" {
			rule.Source = "project"
		}
		saved, err := s.rulesStore.Upsert(rule)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": saved})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// POST /api/rules/enable  { "id": "project:anti-slop", "enabled": false }
// Avoids putting ":" in the URL path (breaks some mux/path stacks).
func (s *Server) handleRulesEnable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	if s.rulesStore == nil || !s.rulesStore.SetEnabled(body.ID, body.Enabled) {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// PUT /api/rules/update  { rule fields including id }
func (s *Server) handleRulesUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var rule rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule")
		return
	}
	if strings.TrimSpace(rule.ID) == "" && strings.TrimSpace(rule.Name) == "" {
		writeError(w, http.StatusBadRequest, "id or name required")
		return
	}
	if rule.Name == "" && strings.Contains(rule.ID, ":") {
		rule.Name = strings.SplitN(rule.ID, ":", 2)[1]
	}
	if rule.Source == "" {
		if strings.HasPrefix(rule.ID, "user:") {
			rule.Source = "user"
		} else {
			rule.Source = "project"
		}
	}
	if s.rulesStore == nil {
		writeError(w, http.StatusServiceUnavailable, "rules unavailable")
		return
	}
	saved, err := s.rulesStore.Upsert(rule)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": saved})
}

// POST /api/rules/delete  { "id": "project:anti-slop" }
func (s *Server) handleRulesDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.ID) == "" {
			writeError(w, http.StatusBadRequest, "id required")
			return
		}
	} else {
		body.ID = r.URL.Query().Get("id")
	}
	if strings.TrimSpace(body.ID) == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	if s.rulesStore == nil || !s.rulesStore.Delete(body.ID) {
		writeError(w, http.StatusNotFound, "rule not found or cannot delete builtin")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// parseRuleRef turns path leftovers into a rule id + optional action.
// Accepts: "project/anti-slop/enable", "project:anti-slop/enable", "user/foo".
func parseRuleRef(path string) (id string, action string) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "", ""
	}
	parts := strings.Split(path, "/")
	if len(parts) >= 2 && (parts[0] == "project" || parts[0] == "user" || parts[0] == "builtin") {
		id = parts[0] + ":" + parts[1]
		if len(parts) >= 3 {
			action = parts[2]
		}
		return id, action
	}
	// Legacy colon form in the first segment: "project:anti-slop" or ".../enable"
	if len(parts) >= 2 && parts[len(parts)-1] == "enable" {
		id = strings.Join(parts[:len(parts)-1], "/")
		return id, "enable"
	}
	return strings.Join(parts, "/"), ""
}

func (s *Server) handleRulesSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/rules/")
	// Dedicated action routes are registered separately; keep this for
	// slash-safe ids like /api/rules/project/anti-slop/enable
	if path == "enable" || path == "update" || path == "delete" {
		switch path {
		case "enable":
			s.handleRulesEnable(w, r)
		case "update":
			s.handleRulesUpdate(w, r)
		case "delete":
			s.handleRulesDelete(w, r)
		}
		return
	}

	id, action := parseRuleRef(path)
	if id == "" {
		writeError(w, http.StatusBadRequest, "rule id required")
		return
	}

	if action == "enable" && r.Method == http.MethodPost {
		var body struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if s.rulesStore == nil || !s.rulesStore.SetEnabled(id, body.Enabled) {
			writeError(w, http.StatusNotFound, "rule not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
		return
	}

	switch r.Method {
	case http.MethodGet:
		rule, ok := s.rulesStore.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "rule not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": rule})
	case http.MethodPut, http.MethodPatch:
		var rule rules.Rule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			writeError(w, http.StatusBadRequest, "invalid rule")
			return
		}
		rule.ID = id
		if rule.Name == "" && strings.Contains(id, ":") {
			rule.Name = strings.SplitN(id, ":", 2)[1]
		}
		if rule.Source == "" {
			if strings.HasPrefix(id, "user:") {
				rule.Source = "user"
			} else {
				rule.Source = "project"
			}
		}
		saved, err := s.rulesStore.Upsert(rule)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": saved})
	case http.MethodDelete:
		if s.rulesStore == nil || !s.rulesStore.Delete(id) {
			writeError(w, http.StatusNotFound, "rule not found or cannot delete builtin")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// GET /api/activity?runId=&limit=
func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	runID := r.URL.Query().Get("runId")
	if s.activityLog == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": []any{}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.activityLog.Tail(limit, runID)})
}

// GET /api/index
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if s.fileIndex == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]any{"status": "unavailable"}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.fileIndex.Stats()})
}

// POST /api/index/rebuild
func (s *Server) handleIndexRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if s.fileIndex == nil {
		writeError(w, http.StatusServiceUnavailable, "index unavailable")
		return
	}
	go s.fileIndex.Rebuild()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": s.fileIndex.Stats()})
}

// GET /api/usage — local token/cache aggregates
func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	stats := s.usageStats
	if stats == nil {
		stats = &UsageStats{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": stats})
}

// GET /api/raw?path= — serve a workspace-relative file (image previews, etc.)
func (s *Server) handleRawFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	rel := strings.TrimSpace(r.URL.Query().Get("path"))
	if rel == "" || strings.Contains(rel, "..") {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	abs := filepath.Join(s.workspace.RootPath, filepath.FromSlash(rel))
	absClean, err := filepath.Abs(abs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}
	root, _ := filepath.Abs(s.workspace.RootPath)
	if !strings.HasPrefix(absClean, root) {
		writeError(w, http.StatusForbidden, "path outside workspace")
		return
	}
	http.ServeFile(w, r, absClean)
}

// POST /api/uploads — browser file attachment upload
func (s *Server) handleUploads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()
	dir := filepath.Join(s.workspace.RootPath, ".blackjak", "uploads")
	_ = os.MkdirAll(dir, 0o755)
	name := filepath.Base(hdr.Filename)
	dest := filepath.Join(dir, strconv.FormatInt(time.Now().UnixNano(), 10)+"_"+name)
	out, err := os.Create(dest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rel, _ := filepath.Rel(s.workspace.RootPath, dest)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"path": filepath.ToSlash(rel),
			"name": name,
		},
	})
}

// UsageStats tracks local token/cache usage (no billing).
type UsageStats struct {
	InputTokens      int64   `json:"inputTokens"`
	OutputTokens     int64   `json:"outputTokens"`
	CacheReadTokens  int64   `json:"cacheReadTokens"`
	CacheWriteTokens int64   `json:"cacheWriteTokens"`
	CacheHitRate     float64 `json:"cacheHitRate"`
	Runs             int64   `json:"runs"`
}

func (u *UsageStats) Record(input, output, cacheRead, cacheWrite int64) {
	if u == nil {
		return
	}
	u.InputTokens += input
	u.OutputTokens += output
	u.CacheReadTokens += cacheRead
	u.CacheWriteTokens += cacheWrite
	u.Runs++
	total := u.CacheReadTokens + u.CacheWriteTokens
	if total > 0 {
		u.CacheHitRate = float64(u.CacheReadTokens) / float64(total)
	}
}

// Ensure types are referenced when packages used from NewServer.
var (
	_ = skills.Skill{}
	_ = rules.Rule{}
)
