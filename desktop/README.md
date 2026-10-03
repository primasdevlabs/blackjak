# Blackjak Desktop (Tauri)

Native shell that loads the shared React UI and spawns the Go agent binary.

## Layout

```
desktop/
  src-tauri/     Tauri 2 app (Rust)
  package.json   Thin npm wrapper
```

Production loads `../extension/webview/react/dist`. Dev uses the Vite server (`localhost:5173`).

## Behavior

- Allocates a **free local port** for the agent (not hard-coded only).
- Resolves `bin/agent` from resource dir, exe-relative paths, or repo `bin/`.
- **Workspace picker** on first launch; path persisted under the app config dir.
- Injects `window.__AGENT_PORT__` so the UI connects to the spawned backend.
- Settings / Activity secondary windows use `#/settings` and `#/activity`.
- Desktop notifications on agent start and (via UI host) run complete/fail.

## Prerequisites

- Rust toolchain + Tauri CLI (`cargo install tauri-cli` or use `@tauri-apps/cli`)
- `make build-backend` and `make build-ui` (the Makefile targets do this)

## Commands

```bash
# From repo root
make desktop-dev     # build backend+UI, then tauri dev
make desktop-build   # production bundle
```

## Change workspace later

The UI can invoke the Tauri command `pick_workspace`, which saves the folder and restarts the agent against it.
