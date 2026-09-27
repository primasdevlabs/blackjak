```
  ____  _            _      _       _    
 |  _ \| |          | |    | |     | |   
 | |_) | | __ _  ___| | __ | | __ _| | __
 |  _ <| |/ _` |/ __| |/ / | |/ _` | |/ /
 | |_) | | (_| | (__|   < _| | (_| |   < 
 |____/|_|\__,_|\___|_|\_(_)_|\__,_|_|\_\
```

# BlackJak - Autonomous Multi-Agent IDE Coding System

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/TypeScript-5.7+-3178C6?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/Vite-6.0+-646CFF?style=for-the-badge&logo=vite&logoColor=white" alt="Vite" />
  <img src="https://img.shields.io/badge/Tailwind_CSS-v4-38B2AC?style=for-the-badge&logo=tailwindcss&logoColor=white" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/VS_Code-Extension-007ACC?style=for-the-badge&logo=visualstudiocode&logoColor=white" alt="VS Code" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="MIT License" />
</p>

BlackJak is an open-source, autonomous multi-agent coding system designed natively for developer environments. It features a high-performance Go backend, a real-time RFC 6455 WebSocket protocol engine, a React 19 sidebar cockpit UI, and a VS Code extension host.

Rather than running as a single prompt loop, BlackJak acts as an agent control plane—orchestrating specialized subagents (Explorer, Architect, Coder, Tester, Reviewer) across structured execution graphs.

---

## 🚀 Zero-Configuration Startup

BlackJak requires **zero manual server configuration** for end users:

1. Install the VS Code extension.
2. Open any workspace in VS Code.
3. The extension activates automatically, spawns the Go agent backend binary on a free port, verifies health, and connects the WebSocket control plane.

You immediately see:
`Agent ● Ready`

---

## ⚙️ Architecture

```
VS Code-family Host (VS Code · Cursor · Windsurf · VSCodium · Theia)
│
├── Extension Host Adapter (extension/)
│   ├── Host Layer (src/host/) — IDEHost interface, VSCodeHost adapter,
│   │   capability detection & host identification (compatibility.ts)
│   ├── Process Manager (Auto-spawns Go agent backend)
│   ├── File Change Tracker & Dirty Document Protection
│   ├── Tab Manager (Intelligent tab policy & editor placement)
│   ├── Webview Sidebar Panel (Hosts the React build)
│   └── Webview Bridge (host-neutral commands → IDEHost calls)
│       │
│       │ Agent Protocol (protocol/) over
│       │ WebSocket / HTTP (127.0.0.1:<allocated-port>)
│       ▼
├── Go Agent Server (api/)
│   ├── REST & RFC 6455 WebSocket Protocol Handlers
│   ├── Capability negotiation (host.hello → HostInfo/HostCapabilities)
│   ├── Command Registry (/api/commands & /api/commands/execute)
│   ├── Settings & Provider Manager (OpenAI, Gemini, Anthropic, Custom)
│   ├── Event Broker (agent/events.go)
│   └── Run Manager (agent/run.go)
│       │
│       ├── Orchestrator Agent & Subagents (agent/orchestrator.go)
│       ├── Context Compactor & Budget Engine (context/compaction.go)
│       ├── Workspace Watcher & Normalized Event Engine (workspace/watcher.go)
│       └── File Lease Lock Manager (workspace/tracker.go)
│
└── React Webview UI (extension/webview/react/)
    └── Host-neutral: speaks AgentHost — never IDE APIs directly
```

The agent runtime is IDE-agnostic: it knows workspaces, files, editors,
terminals and git — never Cursor internals or VS Code internals. The
extension's `IDEHost` layer translates those concepts into the concrete
host's public extension API, and capability negotiation (`host.hello`) lets
the agent degrade gracefully on hosts with partial API coverage.

---

## ✨ Key Capabilities

### 1. Multi-Agent Subagent Orchestration
- **Orchestrator**: Breaks complex user goals into high-level plans and spawns specialized subagents (`Explorer`, `Architect`, `Coder`, `Tester`, `Reviewer`).
- **Parallel & Sequential Execution**: Spawns parallel exploration tasks using Go structured concurrency.
- **Inter-Agent Handoffs**: Subagents pass findings and structured artifacts directly to successor subagents.

### 2. Native Context Management & Automatic Compaction
- **Usable Budget Monitoring**: Calculates usable working memory budget (`UsableMax = MaxTokens - ReservedOutput(16k) - ReservedTools(8k)`).
- **Automatic Threshold Compaction**: Automatically compacts context at ≥80% usable pressure.
- **Structured Representation**: Preserves critical task objectives, implementation decisions, modified file states, test results, and subagent findings while discarding old tool outputs and stale search results.
- **Inline Badge & Modal**: Displays subtle inline activity badges (`✦ Context compacted · 42.1k → 11.7k`) with details modal popovers.

### 3. Extensible `/` Slash Command System
Typing `/` in the composer opens an inline autocomplete menu supporting keyboard navigation:

| Command | Description | Action |
|---|---|---|
| `/compact` | Compact context | Immediately compacts current context & displays token savings. |
| `/context` | Context status | Opens context composition breakdown modal. |
| `/clear` | New context | Prompts for confirmation and resets conversation context. |
| `/summarize` | Summarize task | Generates a durable task summary card. |
| `/plan` | Create plan | Switches operating mode to Plan. |
| `/model` | Change model | Opens model configuration control plane. |
| `/effort` | Change effort | Opens reasoning effort selector. |
| `/agents` | Show subagents | Opens active subagents overlay modal. |
| `/queue` | Show queue | Displays prompt queue items. |
| `/changes` | Show changes | Opens file changes panel. |
| `/files` | Attach files | Opens file attachment picker. |
| `/undo` | Undo last step | Reverts last agent step. |
| `/stop` | Stop agent | Cancels active run execution graph. |

### 4. Live File Change Tracking & Intelligent Tab Policy
- **Normalized Workspace Watcher**: Watches filesystem modifications and normalizes raw OS events (debouncing multiple writes into single `MODIFIED` events).
- **Unsaved Document Safety**: Detects dirty documents (`document.isDirty == true`) and prompts before destructive modifications.
- **Intelligent Tab Policy**:
  - Automatically opens 1–5 touched files beside current editor with `preserveFocus: true`.
  - For 6+ touched files, opens primary files and shows a notification for remaining files.
  - Never steals active editor focus or creates duplicate tabs for already open files.

---

## 🛠️ Installation & Developer Setup

### Prerequisites
- **Go**: 1.22+
- **Node.js**: 20.0+
- **npm**: 10.0+
- **VS Code**: 1.85+

### Build from Source

1. Clone repository:
   ```bash
   git clone https://github.com/primasdevlabs/blackjak.git
   cd blackjak
   ```

2. Build all components:
   ```bash
   make build
   ```

3. Run unit tests & typechecks:
   ```bash
   make test
   ```

4. Run local backend server:
   ```bash
   make run
   ```

5. Run UI Vite dev server:
   ```bash
   make ui
   ```

---

## 🌐 Wire Protocol & API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check probe returning readiness status & workspace path. |
| `GET` | `/api/initial-state` | Handshake endpoint returning initial configuration, runs, queue & models. |
| `GET` / `PATCH` | `/api/settings` | Settings control plane API. |
| `GET` | `/api/providers` | Lists active and available LLM providers. |
| `GET` / `POST` | `/api/commands` | Slash command registry endpoint. |
| `POST` | `/api/commands/execute` | Executes registered slash commands. |
| `GET` / `POST` | `/api/queue` | Manage prompt queue items. |
| `POST` | `/api/runs` | Initializes and launches a new agent run task. |
| `GET` | `/api/events` | Server-Sent Events (SSE) event stream. |
| `WS` | `/ws` | Real-time WebSocket connection endpoint (RFC 6455). |

---

## 🤝 Contributing

We welcome open-source contributions! Please read our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before submitting pull requests.

For security vulnerability reports, please review our [Security Policy](SECURITY.md).

---

## 📄 License

BlackJak is released under the [MIT License](LICENSE).
