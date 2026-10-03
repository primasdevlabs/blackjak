package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSkillsAPI_ListAndEnable(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".blackjak", "skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Demo\nA demo skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewServer(ServerConfig{Host: "127.0.0.1", Port: 0, Workspace: dir}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/skills", nil)
	rec := httptest.NewRecorder()
	s.handleSkills(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d", rec.Code)
	}
	var wrap struct {
		Success bool `json:"success"`
		Data    []struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&wrap); err != nil {
		t.Fatal(err)
	}
	if !wrap.Success {
		t.Fatal("list failed")
	}
	found := false
	for _, sk := range wrap.Data {
		if sk.Name == "demo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("demo skill missing: %+v", wrap.Data)
	}

	body, _ := json.Marshal(map[string]any{"enabled": false})
	req = httptest.NewRequest(http.MethodPost, "/api/skills/demo/enable", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	s.handleSkillsSubroutes(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable status %d body=%s", rec.Code, rec.Body.String())
	}
	sk, ok := s.skillsRegistry.Get("demo")
	if !ok || sk.Enabled {
		t.Fatalf("expected demo disabled, got %+v ok=%v", sk, ok)
	}
}
