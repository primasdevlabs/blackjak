package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"blackjak/agent"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error:   message,
	})
}

// GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:          "ok",
		Version:         "1.0.0",
		ProtocolVersion: ProtocolVersion,
		Workspace:       s.workspace.RootPath,
		Timestamp:       time.Now(),
	})
}

// GET & PATCH /api/settings
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.settingsManager.GetMaskedConfig()
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: cfg})
	case http.MethodPatch, http.MethodPost:
		var newCfg SettingsConfig
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid settings payload")
			return
		}
		s.settingsManager.UpdateConfig(newCfg)
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: s.settingsManager.GetMaskedConfig()})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// GET /api/providers
func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	cfg := s.settingsManager.GetMaskedConfig()
	providersList := make([]string, 0, len(cfg.Providers))
	for name := range cfg.Providers {
		providersList = append(providersList, name)
	}
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: map[string]interface{}{
		"active":    cfg.ActiveProvider,
		"providers": providersList,
	}})
}

// Routes matching /api/providers/:id/test or /models
func (s *Server) handleProviderSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/providers/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		s.handleProviders(w, r)
		return
	}

	providerID := parts[0]
	if len(parts) >= 2 && parts[1] == "test" && r.Method == http.MethodPost {
		success, msg, models := s.settingsManager.TestProvider(providerID)
		writeJSON(w, http.StatusOK, APIResponse{
			Success: success,
			Data: map[string]interface{}{
				"message": msg,
				"models":  models,
			},
		})
		return
	}

	if len(parts) >= 2 && parts[1] == "models" && r.Method == http.MethodGet {
		_, _, models := s.settingsManager.TestProvider(providerID)
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: models})
		return
	}

	writeError(w, http.StatusNotFound, "Subroute not found")
}

// GET & POST /api/queue
func (s *Server) handleQueue(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := s.queueManager.List()
		writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: items})
	case http.MethodPost:
		var req struct {
			Prompt       string   `json:"prompt"`
			Mode         string   `json:"mode"`
			Dependencies []string `json:"dependencies"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid queue payload")
			return
		}
		item := s.queueManager.Add(req.Prompt, req.Mode, req.Dependencies)
		writeJSON(w, http.StatusCreated, APIResponse{Success: true, Data: item})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Routes matching /api/queue/:id (DELETE, PATCH)
func (s *Server) handleQueueSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/queue/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		s.handleQueue(w, r)
		return
	}

	id := parts[0]
	if r.Method == http.MethodDelete {
		s.queueManager.Remove(id)
		writeJSON(w, http.StatusOK, APIResponse{Success: true})
		return
	}

	if len(parts) >= 2 && parts[1] == "run" && r.Method == http.MethodPost {
		s.queueManager.SetStatus(id, agent.QueueStatusRunning)
		writeJSON(w, http.StatusOK, APIResponse{Success: true})
		return
	}

	writeError(w, http.StatusNotFound, "Subroute not found")
}

// POST /api/runs
func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateRunPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "Prompt is required")
		return
	}

	wsPath := req.Workspace
	if wsPath == "" {
		wsPath = s.workspace.RootPath
	}

	run := s.runManager.CreateRun(req.Prompt, wsPath, req.Attachments...)

	go func() {
		_ = s.agent.ExecuteRun(run.Context(), run, s.runManager)
	}()

	dto := toRunDTO(run)
	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    dto,
	})
}

// GET /api/runs
func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	runs := s.runManager.ListRuns()
	dtos := make([]RunDTO, len(runs))
	for i, r := range runs {
		dtos[i] = toRunDTO(r)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    dtos,
	})
}

// Routes matching /api/runs/:id and sub-actions (/cancel, /approval)
func (s *Server) handleRunSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		s.handleListRuns(w, r)
		return
	}

	runID := parts[0]
	run, ok := s.runManager.GetRun(runID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Run '%s' not found", runID))
		return
	}

	if len(parts) == 1 {
		// GET /api/runs/:id
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, APIResponse{
				Success: true,
				Data:    toRunDTO(run),
			})
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	subaction := parts[1]
	switch subaction {
	case "cancel":
		// POST /api/runs/:id/cancel
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		success := s.runManager.CancelRun(runID)
		writeJSON(w, http.StatusOK, APIResponse{
			Success: success,
			Data:    toRunDTO(run),
		})

	case "approval":
		// POST /api/runs/:id/approval
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var payload ApprovalPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		err := s.runManager.SubmitApproval(runID, agent.ApprovalResponse{
			RequestID: payload.RequestID,
			Granted:   payload.Granted,
			Reason:    payload.Reason,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{
			Success: true,
			Data:    toRunDTO(run),
		})

	default:
		writeError(w, http.StatusNotFound, "Subroute not found")
	}
}

// GET /api/events (SSE Stream)
func (s *Server) handleEventsSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	eventCh, unsubscribe := s.broker.Subscribe("")
	defer unsubscribe()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-eventCh:
			if !ok {
				return
			}
			data, err := json.Marshal(evt)
			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}

func toRunDTO(r *agent.Run) RunDTO {
	var subagentList []*agent.Subagent
	if r.Orchestration != nil {
		subagentList = r.Orchestration.ListSubagents()
	} else {
		subagentList = r.Subagents
	}

	return RunDTO{
		ID:          r.ID,
		Prompt:      r.Prompt,
		Workspace:   r.Workspace,
		Status:      r.Status,
		Error:       r.Error,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Plan:        r.Plan,
		Events:      r.GetEvents(),
		Subagents:   subagentList,
		FileChanges: r.FileChanges,
		Attachments: r.Attachments,
		References:  r.References,
		PendingReq:  r.PendingReq,
	}
}
