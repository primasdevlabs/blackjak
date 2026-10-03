package tools

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BrowserTool fetches allowlisted HTTP(S) URLs when the beta browser tool is enabled.
type BrowserTool struct {
	Allowlist []string
	Enabled   bool
}

func (t *BrowserTool) Name() string { return "browser_fetch" }

func (t *BrowserTool) Description() string {
	return "Fetch a public HTTP(S) URL from an allowlisted host (beta). Returns truncated text content. Private/link-local/metadata addresses are blocked."
}

func (t *BrowserTool) Schema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "Absolute http(s) URL to fetch",
			},
		},
		"required": []string{"url"},
	}
}

func (t *BrowserTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	if !t.Enabled {
		return nil, fmt.Errorf("browser tool disabled — enable Beta → Browser tool in settings")
	}
	raw, _ := args["url"].(string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("url required")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("only http(s) URLs are allowed")
	}
	host := strings.ToLower(u.Hostname())
	if !t.hostAllowed(host) {
		return nil, fmt.Errorf("host %q is not in the network allowlist", host)
	}
	if err := rejectPrivateHost(host); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "BlackJak-Agent/1.0")

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			h := strings.ToLower(req.URL.Hostname())
			if !t.hostAllowed(h) {
				return fmt.Errorf("redirect host %q is not in the network allowlist", h)
			}
			if err := rejectPrivateHost(h); err != nil {
				return err
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":   resp.StatusCode,
		"url":      u.String(),
		"content":  string(body),
		"truncated": len(body) >= 64*1024,
	}, nil
}

func (t *BrowserTool) hostAllowed(host string) bool {
	if len(t.Allowlist) == 0 {
		return false
	}
	for _, entry := range t.Allowlist {
		e := strings.ToLower(strings.TrimSpace(entry))
		if e == "" {
			continue
		}
		if host == e || strings.HasSuffix(host, "."+e) {
			return true
		}
	}
	return false
}

// rejectPrivateHost blocks literal IPs and well-known metadata/local hostnames.
func rejectPrivateHost(host string) error {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "localhost" || h == "metadata.google.internal" || strings.HasSuffix(h, ".local") || strings.HasSuffix(h, ".internal") {
		return fmt.Errorf("host %q is blocked (private/local)", host)
	}
	ip := net.ParseIP(h)
	if ip == nil {
		// Hostname — also block obvious metadata DNS names.
		if h == "169.254.169.254" {
			return fmt.Errorf("host %q is blocked (link-local metadata)", host)
		}
		return nil
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return fmt.Errorf("host %q is blocked (private/link-local IP)", host)
	}
	return nil
}
