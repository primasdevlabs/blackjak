package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"blackjak/activity"
	"blackjak/agent"
	bjctx "blackjak/context"
	"blackjak/llm"
	"blackjak/memory"
	"blackjak/protocol"
	"blackjak/rules"
	"blackjak/skills"
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

	skillsRegistry *skills.Registry
	rulesStore     *rules.Store
	activityLog    *activity.Ring
	fileIndex      *workspace.FileIndex
	usageStats     *UsageStats
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
	// Guardrails resolve per-run so Settings edits apply immediately.
	ag.SetPolicyResolver(sm.Policy)
	skillsReg := skills.NewRegistry(cfg.Workspace)
	rulesStore := rules.NewStore(cfg.Workspace)
	ag.SetRulesProvider(func() string { return rulesStore.AssembleEnabled() })
	ag.SetModeProvider(func() string {
		return string(NormalizeMode(sm.GetConfig().Mode))
	})
	ag.SetAskPolicyProvider(func() string {
		return string(NormalizeAskPolicy(sm.GetConfig().AskPolicy))
	})
	ag.SetSkillProvider(func(run *agent.Run) string {
		// Skill name may be passed as "/skill-name ..." in the prompt.
		prompt := strings.TrimSpace(run.Prompt)
		if !strings.HasPrefix(prompt, "/") {
			return ""
		}
		name := strings.TrimPrefix(strings.Fields(prompt)[0], "/")
		if sk, ok := skillsReg.Get(name); ok && sk.Enabled {
			return sk.Body
		}
		return ""
	})
	fileIndex := workspace.NewFileIndex(cfg.Workspace)
	ag.SetBrowserToolProvider(func() (bool, []string) {
		cfg := sm.GetConfig()
		enabled := cfg.BetaFlags != nil && cfg.BetaFlags["browserTool"]
		return enabled, cfg.NetworkAllowlist
	})
	ag.SetRetriever(func(query string, limit int) []string {
		retriever := bjctx.NewRetriever(cfg.Workspace, fileIndex)
		return retriever.Retrieve(query, limit)
	})
	ag.SetRepoSnapshotProvider(func() string {
		snap := fileIndex.Snapshot()
		if snap.FileCount == 0 && len(snap.Languages) == 0 {
			return ""
		}
		var b strings.Builder
		b.WriteString("Repository snapshot:\n")
		if len(snap.Languages) > 0 {
			fmt.Fprintf(&b, "- Languages: %s\n", strings.Join(snap.Languages, ", "))
		}
		fmt.Fprintf(&b, "- Indexed files: %d\n", snap.FileCount)
		if len(snap.BuildFiles) > 0 {
			fmt.Fprintf(&b, "- Build files: %s\n", strings.Join(snap.BuildFiles, ", "))
		}
		if len(snap.Entrypoints) > 0 {
			fmt.Fprintf(&b, "- Entrypoints: %s\n", strings.Join(snap.Entrypoints, ", "))
		}
		return b.String()
	})
	// Rehydrate checkpointed runs so paused/stopped work survives restarts.
	rm.RestoreFromWorkspace(cfg.Workspace)
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
		skillsRegistry:  skillsReg,
		rulesStore:      rulesStore,
		activityLog:     activity.NewRing(2000),
		fileIndex:       fileIndex,
		usageStats:      &UsageStats{},
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
	mux.HandleFunc("/api/workspace/files", s.handleWorkspaceFiles)
	mux.HandleFunc("/api/workspace/symbols", s.handleWorkspaceSymbols)
	mux.HandleFunc("/api/diffs/", s.handleDiffDownload)
	mux.HandleFunc("/api/skills", s.handleSkills)
	mux.HandleFunc("/api/skills/", s.handleSkillsSubroutes)
	mux.HandleFunc("/api/rules", s.handleRules)
	mux.HandleFunc("/api/rules/enable", s.handleRulesEnable)
	mux.HandleFunc("/api/rules/update", s.handleRulesUpdate)
	mux.HandleFunc("/api/rules/delete", s.handleRulesDelete)
	mux.HandleFunc("/api/rules/", s.handleRulesSubroutes)
	mux.HandleFunc("/api/activity", s.handleActivity)
	mux.HandleFunc("/api/index", s.handleIndex)
	mux.HandleFunc("/api/index/rebuild", s.handleIndexRebuild)
	mux.HandleFunc("/api/usage", s.handleUsage)
	mux.HandleFunc("/api/uploads", s.handleUploads)
	mux.HandleFunc("/api/raw", s.handleRawFile)

	// Mirror agent events into the activity ring + usage counters.
	if s.broker != nil && s.activityLog != nil {
		ch, _ := s.broker.Subscribe("")
		go func() {
			for evt := range ch {
				msg, data := auditActivityPayload(evt.Type, evt.Data)
				s.activityLog.Add(evt.RunID, string(evt.Type), msg, data)
				if evt.Type == agent.EventCacheStats && s.usageStats != nil {
					if m, ok := evt.Data.(map[string]interface{}); ok {
						num := func(k string) int64 {
							switch v := m[k].(type) {
							case float64:
								return int64(v)
							case int64:
								return v
							case int:
								return int64(v)
							}
							return 0
						}
						s.usageStats.Record(num("inputTokens"), num("outputTokens"), num("cacheReadTokens"), num("cacheWriteTokens"))
					}
				}
			}
		}()
	}

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
