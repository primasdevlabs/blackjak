package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"blackjak/agent"
	"blackjak/llm"
	"blackjak/tools"
	"blackjak/workspace"
)

// ServerConfig holds configuration options for the Go Agent API server.
type ServerConfig struct {
	Host      string
	Port      int
	Workspace string
}

// Server provides HTTP & WebSocket API access to the agent.
type Server struct {
	host            string
	port            int
	workspace       *workspace.Workspace
	broker          *agent.EventBroker
	runManager      *agent.RunManager
	settingsManager *SettingsManager
	queueManager    *agent.QueueManager
	agent           *agent.Agent
	httpServer      *http.Server
	listener        net.Listener
	mu              sync.Mutex
}

// NewServer initializes an API Server.
func NewServer(cfg ServerConfig, llmClient llm.Client) *Server {
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}

	ws := workspace.New(cfg.Workspace)
	broker := agent.NewEventBroker()
	rm := agent.NewRunManager(broker)
	sm := NewSettingsManager()
	qm := agent.NewQueueManager(broker)

	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(tools.NewFilesystemTool())
	toolRegistry.Register(tools.NewShellTool())
	toolRegistry.Register(tools.NewGitTool())
	toolRegistry.Register(tools.NewSearchTool())
	toolRegistry.Register(tools.NewTestTool())

	ag := agent.New(broker, toolRegistry, ws, llmClient)

	return &Server{
		host:            cfg.Host,
		port:            cfg.Port,
		workspace:       ws,
		broker:          broker,
		runManager:      rm,
		settingsManager: sm,
		queueManager:    qm,
		agent:           ag,
	}
}

// Port returns the assigned port of the running server.
func (s *Server) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr().(*net.TCPAddr).Port
	}
	return s.port
}

// Start launches the HTTP and WebSocket server listening on the configured host:port.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/events", s.handleEventsSSE)
	mux.HandleFunc("/api/runs", s.handleCreateRun)
	mux.HandleFunc("/api/runs/", s.handleRunSubroutes)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/providers", s.handleProviders)
	mux.HandleFunc("/api/providers/", s.handleProviderSubroutes)
	mux.HandleFunc("/api/queue", s.handleQueue)
	mux.HandleFunc("/api/queue/", s.handleQueueSubroutes)

	handler := s.withCORS(mux)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.listener = ln
	s.port = ln.Addr().(*net.TCPAddr).Port
	s.httpServer = &http.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	s.mu.Unlock()

	return s.httpServer.Serve(ln)
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
