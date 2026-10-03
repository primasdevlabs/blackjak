package tools

import "testing"

func TestCheckRead_BlocksSecrets(t *testing.T) {
	p := DefaultPolicy()
	for _, path := range []string{".env", ".env.local", "certs/server.pem", "id_rsa", "keys/app.key"} {
		if err := p.CheckRead(path); err == nil {
			t.Errorf("expected read deny for %q", path)
		}
	}
}

func TestCheckRead_AllowsSource(t *testing.T) {
	p := DefaultPolicy()
	for _, path := range []string{"agent/loop.go", "README.md", "go.mod", ".blackjak/rules/go.md"} {
		if err := p.CheckRead(path); err != nil {
			t.Errorf("unexpected deny for %q: %v", path, err)
		}
	}
}

func TestCheckWrite_AllowsRulesAndSkills(t *testing.T) {
	p := DefaultPolicy()
	for _, path := range []string{".blackjak/rules/ui.md", ".blackjak/skills/review/SKILL.md"} {
		if err := p.CheckWrite(path); err != nil {
			t.Errorf("expected allow write for %q: %v", path, err)
		}
	}
	for _, path := range []string{".blackjak/settings.json", ".blackjak/runs/r1.json", ".env"} {
		if err := p.CheckWrite(path); err == nil {
			t.Errorf("expected deny write for %q", path)
		}
	}
}

func TestRejectPrivateHost(t *testing.T) {
	cases := []string{"127.0.0.1", "10.0.0.1", "192.168.1.1", "169.254.169.254", "localhost", "::1"}
	for _, h := range cases {
		if err := rejectPrivateHost(h); err == nil {
			t.Errorf("expected block for %q", h)
		}
	}
	if err := rejectPrivateHost("example.com"); err != nil {
		t.Errorf("example.com should be allowed: %v", err)
	}
}

func TestBrowserRedirectAllowlist(t *testing.T) {
	bt := &BrowserTool{Enabled: true, Allowlist: []string{"example.com"}}
	if !bt.hostAllowed("example.com") {
		t.Fatal("allow")
	}
	if bt.hostAllowed("evil.com") {
		t.Fatal("deny")
	}
}
