# Contributing to BlackJak

Thank you for your interest in contributing to **BlackJak**! We welcome contributions from developers of all skill levels. As an open-source project, BlackJak thrives on community involvement, feedback, and pull requests.

This document outlines the engineering standards, rules, development workflow, and guidelines for contributing to BlackJak.

---

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please report unacceptable behavior to the project maintainers.

---

## Toolchain & Prerequisites

Before setting up BlackJak locally, ensure your environment satisfies the following requirements:

- **Go**: Version 1.22 or higher
- **Node.js**: Version 20.0 or higher
- **npm**: Version 10.0 or higher
- **Git**: Latest version
- **VS Code**: Version 1.85+ (for testing the extension)

---

## Project Architecture

BlackJak uses a modular architecture:

```
BlackJak Repository
│
├── cmd/agent/          → Go CLI & backend entry point
├── agent/              → Multi-agent orchestrator, run lifecycle, & compaction
├── api/                → REST, SSE, RFC 6455 WebSocket handlers & slash commands
├── workspace/          → File watcher, file leases, & change tracker
├── tools/              → Agent tool execution (Filesystem, Shell, Git, Search, Test)
├── llm/                → Provider abstraction (OpenAI, Gemini, Anthropic, Custom)
│
├── protocol/           → Stable wire contract (messages, events, host info)
│
└── extension/          → VS Code-family extension host adapter (Code, Cursor,
    │                       Windsurf, VSCodium, Theia via the public API)
    ├── src/host/       → IDEHost interface, VSCodeHost, capability detection
    └── webview/react/  → React 19 + Vite 6 + Tailwind CSS v4 Sidebar UI
```

---

## Local Development Setup

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/primasdevlabs/blackjak.git
   cd blackjak
   ```

2. **Install Extension & Webview Dependencies**:
   ```bash
   cd extension && npm install
   cd webview/react && npm install
   cd ../../..
   ```

3. **Build All Artifacts**:
   ```bash
   make build
   ```

4. **Run Backend Server (Standalone Dev Mode)**:
   ```bash
   make run
   ```

5. **Run React UI Dev Server**:
   ```bash
   make ui
   ```
   Open [http://localhost:5173](http://localhost:5173) in your browser.

6. **Launch VS Code Extension in Extension Host**:
   - Open `extension/` in VS Code.
   - Press `F5` to open a new Extension Development Host window with BlackJak preloaded.

---

## Engineering Standards & Guidelines

When submitting code to BlackJak, you MUST adhere to the following core engineering rules:

### 1. Code Style & Styling Rules
- **UI Aesthetics**: Monochrome-first, technical, dense, IDE-native styling (`#050505` bg, `#111111` surface, `#202020` borders, `#f5f5f5` text). No bright SaaS gradients, glassmorphism, or marketing illustrations.
- **Font Rule**: `font-mono` is **strictly forbidden** across UI components and CSS styling. Use standard system/UI typography (`Inter`, `-apple-system`, `sans-serif`).
- **Sidebar-First Design**: The extension sidebar is optimized for narrow viewports (280px–420px). Do not force permanent 3-column layouts into sidebar viewports.

### 2. Modularity & Boundaries
- Maintain clear package boundaries. Public APIs, interfaces, and DTOs should be used rather than leaking internal implementation details.
- Prefer small, composable units with single responsibilities.

### 3. Security
- Validate external input at HTTP/WebSocket boundaries.
- Never commit secrets, API keys, or tokens. Use environment variables or session storage mode.
- Parameterize database/file operations; sanitize raw command inputs.

### 4. Quality & Empirical Verification
- **Never claim success without running verification**: Always run `make test` (`go test ./...`, `npm run typecheck`, and `npm run build`) before submitting a pull request.
- **No superficial symptom patches**: Identify underlying root causes rather than wrapping errors in silent `try/catch` or swallowing exceptions.
- **Explicit Error Context**: Log errors with correlation context (module, operation, task ID).

---

## Git & Pull Request Workflow

1. **Create a Feature Branch**:
   ```bash
   git checkout -b feat/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Commit Message Conventions**:
   Follow [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat(agent): add new subagent strategy`
   - `fix(ui): repair slash command popover positioning`
   - `docs(readme): update installation instructions`
   - `refactor(workspace): optimize file watcher debouncing`

3. **Verify Local Build**:
   ```bash
   make test
   make build
   ```

4. **Submit a Pull Request**:
   - Push your branch to your fork or origin:
     ```bash
     git push origin feat/your-feature-name
     ```
   - Open a Pull Request against the `main` branch.
   - Fill out the PR template with clear description, objective, and testing evidence.

---

## Licensing

By contributing to BlackJak, you agree that your contributions will be licensed under the project's [MIT License](README.md#license).
