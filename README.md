```
  ____  _            _      _       _    
 |  _ \| |          | |    | |     | |   
 | |_) | | __ _  ___| | __ | | __ _| | __
 |  _ <| |/ _` |/ __| |/ / | |/ _` | |/ /
 | |_) | | (_| | (__|   < _| | (_| |   < 
 |____/|_|\__,_|\___|_|\_(_)_|\__,_|_|\_\
```

# Blackjak — Engineering Agent

Blackjak is an open-source AI **engineering agent** for developer environments. Its default identity is that of a senior software / systems engineer — not a chatbot, autocomplete engine, or prompt-to-code generator.

## The Blackjak Principle

The agent's default internal question is not *"What code should I generate?"*

It is:

> **What is the correct engineering solution to this problem, and what evidence do I need to trust it?**

That principle shapes the system prompt, tool policy, engineering loop, verification gates, context system, and UI.

## Engineering loop

```
Understand → Inspect → Reason → Plan → Implement → Verify → Review → Refine
```

For **non-trivial Agent-mode** tasks, Blackjak uses **hybrid gates**:

- Writes are blocked until there is inspect evidence (read / search / git inspect).
- `task_complete` after file changes requires verify evidence (tests / build / lint).
- Non-trivial mutating runs run a structured self-review; `revise` forces refine.
- **Trivial** tasks (short Q&A) skip gates unless they start mutating files.

Ask / Plan modes remain read-only via tool policy.

## What you get

| Surface | Role |
|---|---|
| Go agent backend | Tool loop, policy, engineering state, subagents, compaction |
| VS Code-family extension | Spawns the binary, hosts the React cockpit |
| React UI | Chat, plan, phase/evidence strip, activity, settings |
| Tauri desktop | Native shell over the same UI + agent binary |
| Rules & skills | Always-on engineering / anti-slop / Go / security / UI packs |

Go is a first-class backend language. Security is enforced below the prompt (workspace boundaries, secret-read blocks, destructive approval, browser SSRF controls). Context is progressive — not a full-repo dump every turn.

## Quick start

```bash
git clone https://github.com/primasdevlabs/blackjak.git
cd blackjak
cd extension && npm install && cd webview/react && npm install && cd ../../..
make build
make test
```

- **IDE:** package the VSIX (`make vsix`) and install — the extension spawns the agent.
- **Standalone:** `make run` then `make ui` (browser shell).
- **Desktop:** Rust + Tauri CLI required; `make desktop-dev` (builds backend/UI first, free-port spawn, workspace picker).

Full install guide: [INSTALLATION.md](INSTALLATION.md) · Architecture: [ARCHITECTURE.md](ARCHITECTURE.md)

## Stack

Go 1.22+ · TypeScript 5.7+ · React 19 · Vite · Tauri 2 · VS Code engine 1.85+

## License

MIT — see [LICENSE](LICENSE). Security reports: [SECURITY.md](SECURITY.md).
