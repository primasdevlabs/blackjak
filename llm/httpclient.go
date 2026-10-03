package llm

import (
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	proxyMu  sync.RWMutex
	proxyURL string
)

// SetHTTPProxy sets the global HTTP(S) proxy used by LLM clients.
// Empty string clears the proxy (direct).
func SetHTTPProxy(raw string) {
	proxyMu.Lock()
	defer proxyMu.Unlock()
	proxyURL = raw
}

// HTTPProxy returns the configured proxy URL (may be empty).
func HTTPProxy() string {
	proxyMu.RLock()
	defer proxyMu.RUnlock()
	if proxyURL != "" {
		return proxyURL
	}
	if v := os.Getenv("BLACKJAK_HTTP_PROXY"); v != "" {
		return v
	}
	if v := os.Getenv("HTTPS_PROXY"); v != "" {
		return v
	}
	if v := os.Getenv("HTTP_PROXY"); v != "" {
		return v
	}
	return ""
}

// NewHTTPClient returns an HTTP client with optional proxy and timeout.
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if raw := HTTPProxy(); raw != "" {
		if u, err := url.Parse(raw); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}
