package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"blackjak/agent"
	"blackjak/protocol"
	"blackjak/workspace"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, protocol.APIResponse{
		Success: false,
		Error:   message,
	})
}

// GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "ok",
		"ready":           true,
		"version":         "1.0.0",
		"protocolVersion": protocol.ProtocolVersion,
		"port":            s.Port(),
		"workspace":       s.workspace.RootPath,
		"timestamp":       time.Now(),
	})
}

// GET /api/initial-state
func (s *Server) handleInitialState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cfg := s.settingsManager.GetMaskedConfig()
	runs := s.runManager.ListRuns()
	dtos := make([]protocol.RunDTO, len(runs))
	for i, r := range runs {
		dtos[i] = toRun(r)
	}
	queue := s.queueManager.List()
	modelsMap := s.settingsManager.RefreshModels("")

	writeJSON(w, http.StatusOK, protocol.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"type": "agent.initialized",
			"data": map[string]interface{}{
				"workspace": map[string]interface{}{
					"root": s.workspace.RootPath,
				},
				"settings":  cfg,
				"providers": cfg.Providers,
				"models":    modelsMap,
				"runs":      dtos,
				"queue":     queue,
				"capabilities": map[string]bool{
					"tools":         true,
					"subagents":     true,
					"promptQueue":   true,
					"effortControl": true,
				},
				"host": s.HostInfo(),
			},
		},
	})
}

// GET & PATCH /api/settings
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.settingsManager.GetMaskedConfig()
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: cfg})
	case http.MethodPatch, http.MethodPost:
		var patch map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid settings payload")
			return
		}
		s.settingsManager.ApplyPatch(patch)
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: s.settingsManager.GetMaskedConfig()})
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
	writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: map[string]interface{}{
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
		writeJSON(w, http.StatusOK, protocol.APIResponse{
			Success: success,
			Data: map[string]interface{}{
				"message": msg,
				"models":  models,
			},
		})
		return
	}

	if len(parts) >= 2 && parts[1] == "refresh" && (r.Method == http.MethodPost || r.Method == http.MethodGet) {
		catalogMap := s.settingsManager.RefreshModels(providerID)
		models := catalogMap[providerID]
		writeJSON(w, http.StatusOK, protocol.APIResponse{
			Success: true,
			Data:    models,
		})
		return
	}

	if len(parts) >= 2 && parts[1] == "models" && r.Method == http.MethodGet {
		catalogMap := s.settingsManager.RefreshModels(providerID)
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: catalogMap[providerID]})
		return
	}

	writeError(w, http.StatusNotFound, "Subroute not found")
}

// GET & POST /api/queue
func (s *Server) handleQueue(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := s.queueManager.List()
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: items})
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
		writeJSON(w, http.StatusCreated, protocol.APIResponse{Success: true, Data: item})
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
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true})
		return
	}

	if len(parts) >= 2 && parts[1] == "run" && r.Method == http.MethodPost {
		item, ok := s.queueManager.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "Queue item not found")
			return
		}
		run := s.runManager.CreateRun(item.Prompt, s.workspace.RootPath)
		s.queueManager.SetStatus(id, agent.QueueStatusRunning)
		item.RunID = run.ID
		go func() {
			err := s.agent.ExecuteRun(run.Context(), run, s.runManager)
			if err != nil {
				s.queueManager.SetStatus(id, agent.QueueStatusFailed)
			} else {
				s.queueManager.SetStatus(id, agent.QueueStatusCompleted)
			}
		}()
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: map[string]interface{}{"runId": run.ID}})
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

	var req protocol.CreateRunPayload
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

	dto := toRun(run)
	writeJSON(w, http.StatusCreated, protocol.APIResponse{
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
	dtos := make([]protocol.RunDTO, len(runs))
	for i, r := range runs {
		dtos[i] = toRun(r)
	}

	writeJSON(w, http.StatusOK, protocol.APIResponse{
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
		switch r.Method {
		case http.MethodGet:
			// GET /api/runs/:id
			writeJSON(w, http.StatusOK, protocol.APIResponse{
				Success: true,
				Data:    toRun(run),
			})
		case http.MethodDelete:
			// DELETE /api/runs/:id — removes run + its event history
			s.runManager.DeleteRun(runID)
			writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true})
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
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
		writeJSON(w, http.StatusOK, protocol.APIResponse{
			Success: success,
			Data:    toRun(run),
		})

	case "pause":
		// POST /api/runs/:id/pause — graceful stop with context checkpointed.
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		success := s.runManager.PauseRun(runID)
		writeJSON(w, http.StatusOK, protocol.APIResponse{
			Success: success,
			Data:    toRun(run),
		})

	case "resume":
		// POST /api/runs/:id/resume — continue from the checkpointed context.
		// Body may carry {"prompt": "..."} to redirect the run on resume.
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var payload struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		resumedRun, err := s.runManager.ResumeRun(runID, payload.Prompt)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		go func() {
			_ = s.agent.ExecuteRun(resumedRun.Context(), resumedRun, s.runManager)
		}()
		writeJSON(w, http.StatusOK, protocol.APIResponse{
			Success: true,
			Data:    toRun(resumedRun),
		})

	case "approval":
		// POST /api/runs/:id/approval
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var payload protocol.ApprovalPayload
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
		writeJSON(w, http.StatusOK, protocol.APIResponse{
			Success: true,
			Data:    toRun(run),
		})

	case "changes":
		// POST /api/runs/:id/changes/:changeId — review a file change.
		// changeId "all" applies the action to every pending change.
		if r.Method != http.MethodPost || len(parts) < 3 || parts[2] == "" {
			writeError(w, http.StatusNotFound, "Subroute not found")
			return
		}
		var payload struct {
			Action string `json:"action"` // "accept" | "reject"
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		changeID := parts[2]
		if changeID == "all" {
			ids := pendingChangeIDs(run)
			if len(ids) == 0 {
				writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: toRun(run)})
				return
			}
			for _, id := range ids {
				if _, err := s.runManager.ReviewFileChange(runID, id, payload.Action); err != nil {
					writeError(w, http.StatusBadRequest, err.Error())
					return
				}
			}
			writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: toRun(run)})
			return
		}
		if _, err := s.runManager.ReviewFileChange(runID, changeID, payload.Action); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, protocol.APIResponse{Success: true, Data: toRun(run)})

	default:
		writeError(w, http.StatusNotFound, "Subroute not found")
	}
}

func pendingChangeIDs(run *agent.Run) []string {
	var ids []string
	for _, fc := range run.FileChanges {
		if fc.Status == "" || fc.Status == workspace.ReviewPending {
			ids = append(ids, fc.ID)
		}
	}
	return ids
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

func toRun(r *agent.Run) protocol.RunDTO {
	var subagentList []*agent.Subagent
	if r.Orchestration != nil {
		subagentList = r.Orchestration.ListSubagents()
	} else {
		subagentList = r.Subagents
	}

	return protocol.RunDTO{
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
