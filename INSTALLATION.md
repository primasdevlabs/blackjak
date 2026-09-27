# Installation & Execution Guide

BlackJak runs in three modes:

| Mode | Who it's for | Backend lifecycle |
|---|---|---|
| **IDE extension** | End users | Extension spawns and manages the agent binary automatically |
| **Standalone server + browser UI** | Development, headless use | You run `bin/agent` yourself |
| **Terminal CLI** | Quick prompts, scripting | Single-process interactive session |

---

## Prerequisites

| Tool | Version | Required for |
|---|---|---|
| Go | 1.22+ | Building the agent backend |
| Node.js | 20+ | Building the webview UI and extension |
| npm | 10+ | Dependency installs |
| VS Code-family IDE | engine 1.85+ | Extension mode (VS Code, Cursor, Windsurf, VSCodium, Antigravity, Theia) |

No VS Code-family IDE is needed for standalone or CLI mode.

---

## 1. Build from source

```bash
git clone https://github.com/primasdevlabs/blackjak.git
cd blackjak

# Install JS dependencies (extension + webview)
cd extension && npm install
cd webview/react && npm install
cd ../..

# Build everything: Go binary → webview bundle → extension JS
make build
```

`make build` produces:

```
bin/agent.exe                         # backend (Windows)
extension/bin/agent-win32-x64.exe     # copy bundled into the VSIX
extension/out/                        # compiled extension host code
extension/webview/react/dist/         # production React bundle
```

### Makefile reference

| Target | Action |
|---|---|
| `make build` | Build backend, webview UI, and extension |
| `make vsix` | Build + package `extension/agent-vscode-extension-1.0.0.vsix` |
| `make test` | `go test ./...` + TypeScript typecheck |
| `make run` | Standalone backend on `127.0.0.1:47811` |
| `make ui` | Vite dev server for the webview (`localhost:5173`) |
| `make clean` | Remove all build artifacts |

---

## 2. Install the extension (IDE mode)

### 2a. Package the VSIX

```bash
cd extension
npx @vscode/vsce package --allow-missing-repository
# → agent-vscode-extension-1.0.0.vsix (~6.5 MB)
```

The `.vscodeignore` already excludes sources, `node_modules`, maps, and dev artifacts — the package ships only `out/`, `webview/react/dist`, `resources/`, `ws`, and the agent binary.

### 2b. Install per IDE

```bash
# VS Code
code --install-extension agent-vscode-extension-1.0.0.vsix

# Cursor
cursor --install-extension agent-vscode-extension-1.0.0.vsix

# Windsurf
windsurf --install-extension agent-vscode-extension-1.0.0.vsix

# Antigravity IDE — IMPORTANT: pass --extensions-dir, its CLI defaults to ~/.windsurf
"%LOCALAPPDATA%\Programs\Antigravity IDE\bin\antigravity-ide.cmd" ^
  --extensions-dir "%USERPROFILE%\.antigravity-ide\extensions" ^
  --install-extension agent-vscode-extension-1.0.0.vsix --force
```

Or install from inside the IDE: Extensions view (`Ctrl+Shift+X`) → `⋯` menu → **Install from VSIX…** → pick the file.

### 2c. Activate

1. **Reload the window** after install (`Ctrl+Shift+P` → *Developer: Reload Window*).
2. Click the **BlackJak icon** (robot) in the activity bar.
3. The sidebar opens; the extension auto-starts the bundled agent on a free port and shows `Agent ● Ready` in the status bar.

Commands also available via `Ctrl+Shift+P`: `Agent: Open`, `Agent: New Task`, `Agent: Cancel Task`, `Agent: Restart Agent`, `Agent: Show Agent Logs`.

---

## 3. Execution modes in detail

### IDE mode (recommended)

On activation (`onStartupFinished`), the extension:

1. Resolves the agent binary:
   `extension/bin/agent-<platform>-<arch>[.exe]` → workspace `bin/agent` → `cmd/agent/main.go` via `go run` (dev fallback).
2. Spawns it with `--server --port=0 --workspace=<workspace root>` — port 0 means the OS assigns a free port, so it never collides.
3. Parses the JSON readiness line from stdout (`{"status":"ok","ready":true,...,"port":N}`), then polls `/health`.
4. Connects the control-plane WebSocket and sends `host.hello` with host capabilities.
5. Injects `__AGENT_PORT__` into the webview; the webview also receives a live `agent.endpoint` push so it reconnects if the backend restarts on a new port.

You do **not** need `make run` in this mode — the extension owns the process. It dies with the IDE window.

### Standalone server + browser UI

```bash
make run        # builds then runs: bin\agent.exe --server --port 47811 --workspace .
make ui         # separate terminal: Vite dev server
```

Open `http://localhost:5173`. The web UI resolves the backend via `?port=` → injected port → `localStorage` → 47811, so it works against `make run` unchanged.

Direct binary invocation:

```bash
bin\agent.exe --server --port 47811 --workspace C:\path\to\project
# or any port; 0 = auto-assign
bin\agent.exe --server --port 0 --workspace .
```

### Terminal CLI (no server, no UI)

```bash
bin\agent.exe --cli --workspace .
```

Interactive prompt loop; approvals are auto-granted. Type `exit` to quit. LLM provider is resolved from env vars: `ANTHROPIC_API_KEY` → `OPENAI_API_KEY` → `GEMINI_API_KEY` (first match wins).

---

## 4. Configure a model provider

Runs need an LLM provider. Two ways:

**Settings UI** (per-workspace, recommended): Sidebar → gear icon → *Providers* → pick provider → enter API key / base URL → **Save** → **Test** to verify real connectivity. Keys marked "stored" are persisted in `<workspace>/.blackjak/settings.json` (gitignored).

**Environment variables** (must be set before the backend starts):

```bash
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=...
# OpenAI-compatible endpoints (OpenRouter, Ollama, vLLM, local):
# set base URL + key in Settings → Providers → OpenAI-compatible
```

OpenRouter example: provider *OpenAI-compatible*, base URL `https://openrouter.ai/api/v1`, then pick a model in *Settings → Models*.

> **Separate thinking/coding models**: when the checkbox is off, `activeProvider` + the coding model are used for everything — per-role provider dropdowns only apply when it's on.

---

## 5. Where state lives

| Path | Contents |
|---|---|
| `<workspace>/.blackjak/settings.json` | Provider keys, model routing, agent prefs (gitignored) |
| `<workspace>/.blackjak/memory.json` | Persistent agent memory |
| Run history | In-memory on the backend — survives UI reloads, **not** backend restarts |

---

## 6. Troubleshooting

| Symptom | Cause / fix |
|---|---|
| Icon missing after install | Wrong extensions dir (Antigravity: see 2b) or window not reloaded. Check `Ctrl+Shift+X` for "BlackJak AI Coding Agent". |
| "No data provider registered" | VSIX built before `"type": "webview"` was added — repackage and reinstall. |
| Connection flapping / dead after a while | Stale build. Rebuild (`make build`) so the bundled binary is current, repackage, reinstall, reload window. |
| `bind: Only one usage of each socket address` (standalone) | Old backend still holds 47811 — `netstat -ano \| findstr 47811`, `taskkill /PID <pid> /F`. |
| Settings save fails in browser | Backend running is an old binary without `PATCH` CORS — restart it. |
| Runs fail with "no API key" | Key not configured for this workspace's `.blackjak/settings.json`, or env var was set after backend start. |
| Extension-side errors | `Ctrl+Shift+P` → *Developer: Show Logs…* → *Extension Host*, plus the **AI Agent** output channel — process spawn, health check, and connection state changes are logged there. |

### Manual uninstall

```bash
# VS Code
code --uninstall-extension primasdevlabs.agent-vscode-extension

# Or just delete the folder:
#   ~/.vscode/extensions/primasdevlabs.agent-vscode-extension-1.0.0
#   ~/.antigravity-ide/extensions/primasdevlabs.agent-vscode-extension-1.0.0   (Antigravity)
```
