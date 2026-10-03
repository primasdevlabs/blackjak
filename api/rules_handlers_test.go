package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"blackjak/rules"
)

func newRulesTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	s := NewServer(ServerConfig{Host: "127.0.0.1", Port: 0, Workspace: dir}, nil)
	if s.rulesStore == nil {
		t.Fatal("expected rules store")
	}
	return s
}

func decodeAPIData(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()
	var wrap struct {
		Success bool            `json:"success"`
		Error   string          `json:"error"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&wrap); err != nil {
		t.Fatalf("decode wrap: %v body=%s", err, rec.Body.String())
	}
	if !wrap.Success {
		t.Fatalf("api failure: %s", wrap.Error)
	}
	if dest != nil && len(wrap.Data) > 0 {
		if err := json.Unmarshal(wrap.Data, dest); err != nil {
			t.Fatalf("decode data: %v", err)
		}
	}
}

func TestParseRuleRef(t *testing.T) {
	cases := []struct {
		in     string
		id     string
		action string
	}{
		{"project/anti-slop/enable", "project:anti-slop", "enable"},
		{"builtin/ui", "builtin:ui", ""},
		{"user/foo", "user:foo", ""},
		{"project:anti-slop/enable", "project:anti-slop", "enable"},
		{"project:anti-slop", "project:anti-slop", ""},
		{"", "", ""},
	}
	for _, c := range cases {
		id, action := parseRuleRef(c.in)
		if id != c.id || action != c.action {
			t.Fatalf("parseRuleRef(%q) = (%q,%q), want (%q,%q)", c.in, id, action, c.id, c.action)
		}
	}
}

func TestRulesAPI_CreateEnableUpdateDelete(t *testing.T) {
	s := newRulesTestServer(t)

	// Create project rule
	createBody, _ := json.Marshal(map[string]any{
		"name": "anti-slop", "body": "avoid slop", "source": "project", "enabled": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/rules", bytes.NewReader(createBody))
	rec := httptest.NewRecorder()
	s.handleRules(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create status %d body=%s", rec.Code, rec.Body.String())
	}
	var created rules.Rule
	decodeAPIData(t, rec, &created)
	if created.ID != "project:anti-slop" || created.Source != "project" {
		t.Fatalf("unexpected created: %+v", created)
	}
	md := filepath.Join(s.workspace.RootPath, ".blackjak", "rules", "anti-slop.md")
	if _, err := os.Stat(md); err != nil {
		t.Fatalf("expected project md at %s: %v", md, err)
	}

	// Enable toggle via body endpoint (no colon in URL)
	enBody, _ := json.Marshal(map[string]any{"id": created.ID, "enabled": false})
	req = httptest.NewRequest(http.MethodPost, "/api/rules/enable", bytes.NewReader(enBody))
	rec = httptest.NewRecorder()
	s.handleRulesEnable(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable status %d body=%s", rec.Code, rec.Body.String())
	}
	got, ok := s.rulesStore.Get(created.ID)
	if !ok || got.Enabled {
		t.Fatalf("expected disabled, got %+v ok=%v", got, ok)
	}

	// Update body
	upBody, _ := json.Marshal(map[string]any{
		"id": created.ID, "name": "anti-slop", "body": "updated body", "source": "project",
	})
	req = httptest.NewRequest(http.MethodPut, "/api/rules/update", bytes.NewReader(upBody))
	rec = httptest.NewRecorder()
	s.handleRulesUpdate(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status %d body=%s", rec.Code, rec.Body.String())
	}
	got, _ = s.rulesStore.Get(created.ID)
	if got.Body != "updated body" {
		t.Fatalf("body not updated: %q", got.Body)
	}

	// Slash-safe enable path
	enBody, _ = json.Marshal(map[string]any{"enabled": true})
	req = httptest.NewRequest(http.MethodPost, "/api/rules/project/anti-slop/enable", bytes.NewReader(enBody))
	rec = httptest.NewRecorder()
	s.handleRulesSubroutes(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("slash enable status %d body=%s", rec.Code, rec.Body.String())
	}
	got, _ = s.rulesStore.Get(created.ID)
	if !got.Enabled {
		t.Fatal("expected re-enabled via slash path")
	}

	// Delete via body endpoint
	delBody, _ := json.Marshal(map[string]any{"id": created.ID})
	req = httptest.NewRequest(http.MethodPost, "/api/rules/delete", bytes.NewReader(delBody))
	rec = httptest.NewRecorder()
	s.handleRulesDelete(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status %d body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := s.rulesStore.Get(created.ID); ok {
		t.Fatal("rule still present after delete")
	}
}

func TestRulesAPI_EnableMissingID(t *testing.T) {
	s := newRulesTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/rules/enable", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	s.handleRulesEnable(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestRulesAPI_ColonPathEnable_DoesNot405WhenRouted(t *testing.T) {
	// Even the legacy colon form should work when the subroute handler is invoked directly.
	s := newRulesTestServer(t)
	_, _ = s.rulesStore.UpsertProject("anti-slop", "body")

	enBody, _ := json.Marshal(map[string]any{"enabled": false})
	req := httptest.NewRequest(http.MethodPost, "/api/rules/project:anti-slop/enable", bytes.NewReader(enBody))
	rec := httptest.NewRecorder()
	s.handleRulesSubroutes(rec, req)
	if rec.Code == http.StatusMethodNotAllowed {
		t.Fatalf("colon path enable returned 405: %s", rec.Body.String())
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("colon path enable status %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRulesAPI_ListIncludesBuiltins(t *testing.T) {
	s := newRulesTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/rules", nil)
	rec := httptest.NewRecorder()
	s.handleRules(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d", rec.Code)
	}
	var list []rules.Rule
	decodeAPIData(t, rec, &list)
	if len(list) < 5 {
		t.Fatalf("expected builtins, got %d", len(list))
	}
}
