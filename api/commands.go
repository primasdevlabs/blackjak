package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"blackjak/agent"
	cxt "blackjak/context"
	"blackjak/llm"
	"blackjak/protocol"
	"blackjak/workspace"
)

func execCommand(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// estimateTokens approximates the token footprint of a message list (~4 chars/token).
func estimateTokens(msgs []llm.Message) int {
	total := 0
	for _, m := range msgs {
		total += len(m.Content)
	}
	return total / 4
}

type CommandArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

type CommandHandler func(server *Server, args map[string]interface{}) (interface{}, error)

type AgentCommand struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   []CommandArgument `json:"arguments,omitempty"`
	Execute     CommandHandler    `json:"-"`
}

type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]AgentCommand
}

func NewCommandRegistry() *CommandRegistry {
	cr := &CommandRegistry{
		commands: make(map[string]AgentCommand),
	}
	cr.registerDefaults()
	return cr
}

func (cr *CommandRegistry) Register(cmd AgentCommand) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.commands[strings.ToLower(cmd.Name)] = cmd
}

func (cr *CommandRegistry) List() []AgentCommand {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	list := make([]AgentCommand, 0, len(cr.commands))
	for _, cmd := range cr.commands {
		list = append(list, cmd)
	}
	return list
}

func (cr *CommandRegistry) Get(name string) (AgentCommand, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	cmd, ok := cr.commands[strings.ToLower(name)]
	return cmd, ok
}

func (cr *CommandRegistry) registerDefaults() {
	cr.Register(AgentCommand{
		Name:        "compact",
		Description: "Compact context immediately",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil {
				return nil, fmt.Errorf("no run to compact")
			}
			msgs := run.GetMessages()
			before := estimateTokens(msgs)

			if run.Status == agent.RunRunning || run.Status == agent.RunWaiting || run.Status == agent.RunPending {
				// Live run — the loop performs the actual compaction and
				// emits context.updated with real before/after counts.
				if !run.RequestCompaction() {
					return nil, fmt.Errorf("compaction already pending")
				}
				return map[string]interface{}{
					"status":       "requested",
					"tokensBefore": before,
					"message":      "Compaction will apply at the next step.",
				}, nil
			}

			// Finished run — report what compacting its transcript would look like.
			return map[string]interface{}{
				"status":       "idle",
				"tokensBefore": before,
				"tokensAfter":  int(float64(before) * 0.35),
				"message":      "Run is finished; context is no longer live.",
			}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "context",
		Description: "Show context composition and budget",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			compactor := cxt.NewCompactor(128000)
			run := s.runManager.LatestRun()

			var system, task, conversation, toolResults, agentSummaries int
			if run != nil {
				for _, m := range run.GetMessages() {
					switch m.Role {
					case "system":
						system += len(m.Content) / 4
					case "user":
						if m.Content == run.Prompt {
							task += len(m.Content) / 4
						} else {
							conversation += len(m.Content) / 4
						}
					case "assistant":
						conversation += len(m.Content) / 4
					case "tool":
						toolResults += len(m.Content) / 4
					}
				}
				for _, sub := range run.Subagents {
					agentSummaries += len(sub.Result) / 4
				}
			}

			total := system + task + conversation + toolResults + agentSummaries
			budget := compactor.CalculateBudget(total)
			return map[string]interface{}{
				"system":          system,
				"task":            task,
				"conversation":    conversation,
				"toolResults":     toolResults,
				"agentSummaries":  agentSummaries,
				"total":           total,
				"limit":           compactor.MaxContextWindow,
				"pressure":        int(budget.Pressure * 100),
				"usedTokens":      budget.UsedTokens,
				"availableTokens": budget.AvailableTokens,
				"reservedTokens":  budget.ReservedTokens,
			}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "clear",
		Description: "Start fresh context",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			removed := s.runManager.ClearFinished()
			return map[string]interface{}{
				"cleared":      true,
				"runsCleared":  removed,
				"message":      "Finished run history cleared.",
			}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "summarize",
		Description: "Create durable task summary",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil {
				return nil, fmt.Errorf("no run to summarize")
			}

			completed := []string{}
			remaining := []string{}
			if run.Plan != nil {
				for i, step := range run.Plan.Steps {
					if i < run.Plan.CurrentStep {
						completed = append(completed, step)
					} else {
						remaining = append(remaining, step)
					}
				}
			}

			files := []cxt.FileState{}
			for _, fc := range run.FileChanges {
				status := "M"
				switch fc.Type {
				case workspace.ChangeCreated:
					status = "A"
				case workspace.ChangeDeleted:
					status = "D"
				}
				files = append(files, cxt.FileState{Path: fc.Path, Status: status})
			}

			return map[string]interface{}{
				"objective": run.Prompt,
				"workspace": run.Workspace,
				"status":    string(run.Status),
				"completed": completed,
				"remaining": remaining,
				"files":     files,
				"subagents": len(run.Subagents),
				"steps":     len(run.GetEvents()),
			}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "plan",
		Description: "Show the current task plan",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil || run.Plan == nil {
				return map[string]interface{}{"steps": []string{}, "currentStep": 0}, nil
			}
			return run.Plan, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "model",
		Description: "Change active LLM model",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			model, _ := args["model"].(string)
			if model != "" {
				raw, _ := json.Marshal(model)
				s.settingsManager.ApplyPatch(map[string]json.RawMessage{"codingModelId": raw})
			}
			return map[string]interface{}{"codingModelId": s.settingsManager.GetConfig().CodingModelID}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "effort",
		Description: "Change reasoning effort level",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			effort, _ := args["effort"].(string)
			if effort != "" {
				raw, _ := json.Marshal(effort)
				s.settingsManager.ApplyPatch(map[string]json.RawMessage{"effort": raw})
			}
			return map[string]interface{}{"effort": s.settingsManager.GetConfig().Effort}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "agents",
		Description: "Show active subagents",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil {
				return map[string]interface{}{"subagents": []interface{}{}}, nil
			}
			return map[string]interface{}{"subagents": run.Subagents}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "queue",
		Description: "Show prompt queue",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			return s.queueManager.List(), nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "changes",
		Description: "Show modified files diffs",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil {
				return map[string]interface{}{"changes": []interface{}{}}, nil
			}
			return map[string]interface{}{"changes": run.FileChanges}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "files",
		Description: "View run file attachments and references",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil {
				return map[string]interface{}{"attachments": []interface{}{}, "references": []interface{}{}}, nil
			}
			return map[string]interface{}{
				"attachments": run.Attachments,
				"references":  run.References,
			}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "undo",
		Description: "Undo last agent action",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil || len(run.FileChanges) == 0 {
				return nil, fmt.Errorf("nothing to undo")
			}
			last := run.FileChanges[len(run.FileChanges)-1]
			if last.Type == workspace.ChangeCreated {
				if err := os.Remove(last.Path); err != nil {
					return nil, fmt.Errorf("undo failed: %w", err)
				}
				return map[string]interface{}{"status": "reverted", "path": last.Path}, nil
			}
			return nil, fmt.Errorf("cannot revert a %s change — restore from git or backup", last.Type)
		},
	})

	cr.Register(AgentCommand{
		Name:        "stop",
		Description: "Stop agent execution",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			if run == nil {
				return nil, fmt.Errorf("no run to stop")
			}
			if s.runManager.CancelRun(run.ID) {
				return map[string]interface{}{"status": "cancelled", "runId": run.ID}, nil
			}
			return map[string]interface{}{"status": string(run.Status), "runId": run.ID}, nil
		},
	})

	cr.Register(AgentCommand{
		Name:        "pr",
		Description: "Summarize current git changes for a pull request",
		Execute: func(s *Server, args map[string]interface{}) (interface{}, error) {
			run := s.runManager.LatestRun()
			statusOut, _ := execCommand(s.workspace.RootPath, "git", "status", "--short")
			diffOut, _ := execCommand(s.workspace.RootPath, "git", "diff", "--stat")
			logOut, _ := execCommand(s.workspace.RootPath, "git", "log", "-5", "--oneline")
			var changed []string
			if run != nil {
				for _, fc := range run.FileChanges {
					changed = append(changed, string(fc.Type)+" "+fc.Path)
				}
			}
			title := "Update project"
			if run != nil && strings.TrimSpace(run.Prompt) != "" {
				title = truncateStr(strings.TrimSpace(run.Prompt), 72)
			}
			body := fmt.Sprintf("## Summary\n%s\n\n## Agent file changes\n%s\n\n## Git status\n```\n%s\n```\n\n## Diffstat\n```\n%s\n```\n\n## Recent commits\n```\n%s\n```\n",
				title,
				strings.Join(changed, "\n"),
				strings.TrimSpace(statusOut),
				strings.TrimSpace(diffOut),
				strings.TrimSpace(logOut),
			)
			return map[string]interface{}{
				"title": title,
				"body":  body,
			}, nil
		},
	})
}

// GET /api/commands
func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	cmds := s.commandRegistry.List()
	writeJSON(w, http.StatusOK, protocol.APIResponse{
		Success: true,
		Data:    cmds,
	})
}

// POST /api/commands/execute
func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var payload struct {
		Command string                 `json:"command"`
		Args    map[string]interface{} `json:"args,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid command payload")
		return
	}

	cmd, ok := s.commandRegistry.Get(payload.Command)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Command '/%s' not found", payload.Command))
		return
	}

	result, err := cmd.Execute(s, payload.Args)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, protocol.APIResponse{
		Success: true,
		Data:    result,
	})
}
