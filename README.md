```
  ____  _            _      _       _    
 |  _ \| |          | |    | |     | |   
 | |_) | | __ _  ___| | __ | | __ _| | __
 |  _ <| |/ _` |/ __| |/ / | |/ _` | |/ /
 | |_) | | (_| | (__|   < _| | (_| |   < 
 |____/|_|\__,_|\___|_|\_(_)_|\__,_|_|\_\
```

# BlackJak - Multi-Agent IDE Coding System

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/Vite-5.0+-646CFF?style=for-the-badge&logo=vite&logoColor=white" alt="Vite" />
  <img src="https://img.shields.io/badge/VS_Code-Extension-007ACC?style=for-the-badge&logo=visualstudiocode&logoColor=white" alt="VS Code" />
  <img src="https://img.shields.io/badge/WebSocket-RFC_6455-000000?style=for-the-badge&logo=websocket&logoColor=white" alt="WebSocket" />
</p>

BlackJak is an extensible multi-agent IDE coding system. It consists of a Go API backend, a real-time event system, a React UI, and a VS Code extension host.

Instead of running as a single monolithic execution loop, BlackJak acts as an orchestrator that creates, configures, and coordinates specialized subagents running in parallel and sequential execution graphs.

---

## Architecture Overview

```
VS Code Host / Workspace
│
├── VS Code Extension (extension/vscode)
│   ├── Process Manager (Auto-spawns Go agent backend)
│   └── Webview Panel (Hosts React UI build)
│       │
│       │ WebSocket / HTTP (Protocol Version 1)
│       ▼
├── Go Agent Server (api/)
│   ├── REST & WebSocket Protocol Handlers
│   ├── Event Broker (agent/events.go)
│   └── Run Manager (agent/run.go)
│       │
│       ├── Orchestrator Agent (agent/orchestrator.go)
│       │   ├── Explorer Agent
│       │   ├── Architecture Agent
│       │   ├── Coder Agent
│       │   ├── Tester Agent
│       │   └── Reviewer Agent
│       │
│       ├── Tool Registry & Execution (tools/)
│       ├── Workspace Sandbox Security (workspace/sandbox.go)
│       └── File Lease Lock Manager (workspace/tracker.go)
```

---

## Key Features

### 1. Multi-Agent Orchestration
- **Orchestrator Agent**: Analyzes user goals, breaks down requirements into structured execution plans, and delegates subtasks to specialized subagents.
- **Specialized Roles**: Supports Explorer, Architect, Coder, Tester, Reviewer, Debugger, Refactorer, and custom subagent roles.
- **Parallel & Sequential Subagent Execution**: Uses Go structured concurrency (`sync.WaitGroup`, child contexts) for concurrent file exploration and architectural analysis.
- **Inter-Agent Handoffs**: Subagents report findings and pass structured data directly to successor subagents.

### 2. Workspace Intelligence & Reference Parsing
- **Prompt Reference Parser**: Automatically detects and resolves `@path/to/file.go:84` and `@path/to/folder` references against the workspace root.
- **Interactive File References**: UI renders file paths as clickable elements that instruct VS Code to open the exact file and jump to the target line.
- **Context-Aware Attachments**: Supports explicit file and folder attachments drag-and-dropped into the chat prompt.

### 3. Concurrency Protection & File Tracking
- **File Lease Locks**: Prevents multiple subagents from overwriting the same file concurrently. Requests exclusive locks and detects workspace edit conflicts.
- **Workspace Tracker**: Logs all file creations, modifications, deletions, renames, and moves with agent attribution and diff availability.

### 4. Ambient Activity Feedback System
- **Event-Driven Categories**: Converts factual agent events into ambient status categories (`orchestrating`, `exploring`, `thinking`, `debugging`, `testing`, `editing`, `finishing`).
- **Dynamic Rotation**: Periodically rotates ambient activity feedback using pseudo-random selection (3s to 8s intervals) without back-to-back message repetitions.

### 5. VS Code Integration
- **Auto Backend Lifecycle**: Extension host checks health (`/health`) and auto-spawns the Go agent process on an available port.
- **Native Editor & Diff Viewers**: Triggers native VS Code side-by-side diffs (`vscode.diff`) and opens newly created files automatically.

---

## Wire Protocol & API Endpoints

The Go server exposes a transport-neutral HTTP and WebSocket interface:

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Returns server status, version, protocol version, and root workspace path. |
| `POST` | `/api/runs` | Initializes and launches a new agent run task. |
| `GET` | `/api/runs` | Lists all agent runs in reverse chronological order. |
| `GET` | `/api/runs/:id` | Retrieves detailed state, events, subagents, and file changes for a run. |
| `POST` | `/api/runs/:id/cancel` | Cancels an active run and propagates cancellation to all subagents. |
| `POST` | `/api/runs/:id/approval` | Submits a human approval decision (`granted` or `denied`). |
| `GET` | `/api/events` | Server-Sent Events (SSE) fallback event stream. |
| `WS` | `/ws` | Real-time WebSocket connection endpoint (RFC 6455). |

### WebSocket Client Commands
```json
{
  "type": "run.start",
  "data": {
    "prompt": "Fix authentication bug in @src/middleware/auth.ts:84",
    "workspace": "C:/wamp64/www/blackjak",
    "attachments": []
  }
}
```

```json
{
  "type": "approval.respond",
  "data": {
    "runId": "run_12345",
    "requestId": "app_67890",
    "granted": true
  }
}
```

---

## Project Structure

```
.
├── cmd/
│   └── agent/
│       └── main.go              # CLI & Server launcher
├── agent/
│   ├── agent.go                 # Main execution loop
│   ├── orchestrator.go          # Multi-agent orchestrator
│   ├── subagent.go              # Subagent model & message types
│   ├── run.go                   # Run model & approval flow
│   ├── events.go                # Multi-subscriber EventBroker
│   ├── state.go                 # Agent state
│   ├── planner.go               # Plan generation
│   ├── context.go               # Context assembler
│   └── memory.go                # Agent memory
├── api/
│   ├── server.go                # HTTP & WebSocket server
│   ├── handlers.go              # REST & SSE handlers
│   ├── websocket.go             # RFC 6455 WebSocket upgrader
│   └── protocol.go              # Versioned protocol DTOs
├── workspace/
│   ├── workspace.go             # Workspace boundary validator
│   ├── sandbox.go               # Security sandbox
│   ├── reference.go             # Prompt @-reference parser
│   └── tracker.go               # File change & lease lock manager
├── tools/                       # File, shell, git, search, test tools
├── ui/
│   ├── src/
│   │   ├── activity/            # Activity feedback manager
│   │   ├── api/                 # REST & WebSocket client
│   │   ├── components/          # React components (Heroicons)
│   │   ├── hooks/               # Custom React hooks
│   │   ├── state/               # State store
│   │   └── types/               # TypeScript interfaces
│   ├── package.json
│   └── vite.config.ts
├── extension/
│   └── vscode/                  # VS Code extension host
├── Makefile                     # Build & developer tasks
└── .gitignore
```

---

## Developer Commands & Makefile Targets

### Build Everything
```bash
make build
```
Executes:
1. `go build -o bin/agent.exe ./cmd/agent` (Go agent server binary)
2. `cd ui && npm run build` (React UI Vite production build)
3. `cd extension/vscode && npm run build` (VS Code extension build)

### Run Unit Tests
```bash
make test
```
Runs all Go backend unit tests across `agent/`, `api/`, and `workspace/`, followed by React UI TypeScript typechecks.

### Run Agent Server
```bash
make run
```
Launches the Go agent server on port 8080 listening on `127.0.0.1`.

### Run React UI Dev Server
```bash
make ui
```
Starts the Vite UI development server at `http://localhost:5173`.

---

## License

MIT License.
