package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blackjak/agent"
	"blackjak/protocol"
)

func TestServer_HealthEndpoint(t *testing.T) {
	server := NewServer(ServerConfig{Host: "127.0.0.1", Port: 0, Workspace: "/tmp/test"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", rec.Code)
	}

	var resp protocol.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "ok" || resp.ProtocolVersion != protocol.ProtocolVersion {
		t.Errorf("Unexpected health payload: %+v", resp)
	}
}

func TestServer_CreateAndGetRun(t *testing.T) {
	server := NewServer(ServerConfig{Host: "127.0.0.1", Port: 0, Workspace: "/tmp/test"}, nil)

	payload := protocol.CreateRunPayload{Prompt: "Implement feature X"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/runs", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	server.handleCreateRun(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected HTTP 201 Created, got %d", rec.Code)
	}

	var apiResp protocol.APIResponse
	_ = json.NewDecoder(rec.Body).Decode(&apiResp)

	if !apiResp.Success {
		t.Fatalf("API response reported failure: %s", apiResp.Error)
	}

	runData, _ := json.Marshal(apiResp.Data)
	var runDTO protocol.RunDTO
	_ = json.Unmarshal(runData, &runDTO)

	if runDTO.Prompt != "Implement feature X" {
		t.Errorf("Unexpected run prompt: %s", runDTO.Prompt)
	}

	// Test GET /api/runs/:id
	reqGet := httptest.NewRequest(http.MethodGet, "/api/runs/"+runDTO.ID, nil)
	recGet := httptest.NewRecorder()

	server.handleRunSubroutes(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for GET run, got %d", recGet.Code)
	}
}

func TestServer_CancelRun(t *testing.T) {
	server := NewServer(ServerConfig{Host: "127.0.0.1", Port: 0, Workspace: "/tmp/test"}, nil)
	run := server.runManager.CreateRun("Test cancel", "/tmp/test")

	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+run.ID+"/cancel", nil)
	rec := httptest.NewRecorder()

	server.handleRunSubroutes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", rec.Code)
	}

	if run.Status != agent.RunCancelled {
		t.Errorf("Expected run status cancelled, got %s", run.Status)
	}
}
