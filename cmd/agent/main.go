package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"blackjak/agent"
	"blackjak/api"
	"blackjak/workspace"
)

func main() {
	serverMode := flag.Bool("server", false, "Start the HTTP/WebSocket API server")
	host := flag.String("host", "127.0.0.1", "Host address to listen on")
	port := flag.Int("port", 0, "Port to listen on (0 for random available port)")
	wsDir := flag.String("workspace", "", "Path to working workspace directory")
	flag.Parse()

	if *wsDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		*wsDir = cwd
	}

	absWs, err := filepath.Abs(*wsDir)
	if err == nil {
		*wsDir = absWs
	}

	if *serverMode {
		runServer(*host, *port, *wsDir)
	} else {
		runCLI(*wsDir)
	}
}

func runServer(host string, port int, wsDir string) {
	cfg := api.ServerConfig{
		Host:      host,
		Port:      port,
		Workspace: wsDir,
	}

	server := api.NewServer(cfg, nil)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err != fmt.Errorf("http: Server closed") {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// Give server time to bind port
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Agent API Server running at http://%s:%d (Workspace: %s)\n", host, server.Port(), wsDir)
	fmt.Printf("Health check: http://%s:%d/health\n", host, server.Port())

	<-sigCh
	fmt.Println("\nShutting down Agent Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
	fmt.Println("Server stopped cleanly.")
}

func runCLI(wsDir string) {
	fmt.Println("==========================================")
	fmt.Println("         Go Coding Agent CLI              ")
	fmt.Printf(" Workspace: %s\n", wsDir)
	fmt.Println("==========================================")
	fmt.Println("Type a prompt and press Enter (or type 'exit' to quit):")

	broker := agent.NewEventBroker()
	rm := agent.NewRunManager(broker)
	ws := workspace.New(wsDir)
	ag := agent.New(broker, nil, ws, nil)

	// Subscribe to events for stdout rendering in CLI
	eventCh, _ := broker.Subscribe("")
	go func() {
		for evt := range eventCh {
			switch evt.Type {
			case agent.EventAgentPlan:
				if plan, ok := evt.Data.(*agent.Plan); ok {
					fmt.Println("\n--- Plan ---")
					for i, s := range plan.Steps {
						status := "○"
						if i < plan.CurrentStep {
							status = "✓"
						} else if i == plan.CurrentStep {
							status = "●"
						}
						fmt.Printf(" %s %s\n", status, s)
					}
					fmt.Println("------------")
				}
			case agent.EventAgentMessage:
				if data, ok := evt.Data.(map[string]interface{}); ok {
					fmt.Printf("\n[Agent] %v\n", data["content"])
				}
			case agent.EventToolStarted:
				if data, ok := evt.Data.(map[string]interface{}); ok {
					fmt.Printf("  -> Tool [%v]: %v\n", data["tool"], data["step"])
				}
			case agent.EventApprovalRequired:
				if req, ok := evt.Data.(agent.ApprovalRequest); ok {
					fmt.Printf("\n[APPROVAL REQUIRED] %s\nAuto-approving in CLI mode...\n", req.Description)
					_ = rm.SubmitApproval(req.RunID, agent.ApprovalResponse{
						RequestID: req.ID,
						Granted:   true,
					})
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		run := rm.CreateRun(line, wsDir)
		if err := ag.ExecuteRun(run.Context(), run, rm); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
