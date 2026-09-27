package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"path/filepath"

	"blackjak/agent"
	"blackjak/llm"
	"blackjak/memory"
	"blackjak/protocol"
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
	commandRegistry *CommandRegistry
	agent           *agent.Agent
	httpServer      *http.Server
	listener        net.Listener
	mu              sync.Mutex

	hostInfo    *protocol.HostInfo
	hostInfoMtx sync.RWMutex
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
	sm.SetStoragePath(filepath.Join(cfg.Workspace, ".blackjak", "settings.json"))
	qm := agent.NewQueueManager(broker)
	cr := NewCommandRegistry()

	ag := agent.New(broker, ws, llmClient)
	ag.SetPersistentMemory(memory.NewPersistentMemory(
		filepath.Join(cfg.Workspace, ".blackjak", "memory.json")))
	ag.SetClientResolver(func(role string) llm.Client {
		cfg := sm.GetConfig()
		var mc llm.ModelConfig
		switch role {
		case "thinking", "explorer", "reviewer":
			mc = cfg.Thinking
		case "fast", "tester":
			mc = cfg.Fast
		case "review":
			mc = cfg.Review
		default:
			mc = cfg.Coding
		}
		if !cfg.UseSeparateModels || mc.ModelID == "" {
			mc = llm.ModelConfig{
				ProviderID: cfg.ActiveProvider,
				ModelID:    cfg.CodingModelID,
			}
		}
		prov := sm.createProviderInstance(mc.ProviderID)
		if prov == nil {
			prov = sm.activeProvider
		}
		if prov == nil {
			return nil
		}
		return llm.NewProviderClient(prov, mc.ModelID)
	})

	return &Server{
		host:            cfg.Host,
		port:            cfg.Port,
		workspace:       ws,
		broker:          broker,
		runManager:      rm,
		settingsManager: sm,
		queueManager:    qm,
		commandRegistry: cr,
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

// SetHostInfo records the host adapter identity and capabilities negotiated
// via the host.hello protocol message.
func (s *Server) SetHostInfo(info protocol.HostInfo) {
	s.hostInfoMtx.Lock()
	defer s.hostInfoMtx.Unlock()
	s.hostInfo = &info
}

// HostInfo returns the negotiated host adapter info, or nil when the agent
// is running without an IDE host (e.g. CLI mode).
func (s *Server) HostInfo() *protocol.HostInfo {
	s.hostInfoMtx.RLock()
	defer s.hostInfoMtx.RUnlock()
	return s.hostInfo
}

// Start launches the HTTP and WebSocket server listening on the configured host:port.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/initial-state", s.handleInitialState)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/events", s.handleEventsSSE)
	mux.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleListRuns(w, r)
			return
		}
		s.handleCreateRun(w, r)
	})
	mux.HandleFunc("/api/runs/", s.handleRunSubroutes)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/providers", s.handleProviders)
	mux.HandleFunc("/api/providers/", s.handleProviderSubroutes)
	mux.HandleFunc("/api/queue", s.handleQueue)
	mux.HandleFunc("/api/queue/", s.handleQueueSubroutes)
	mux.HandleFunc("/api/commands", s.handleCommands)
	mux.HandleFunc("/api/commands/execute", s.handleExecuteCommand)

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
