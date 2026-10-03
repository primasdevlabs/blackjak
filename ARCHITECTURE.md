# Blackjak Architecture

## Agent loop

Entry: `agent.ExecuteRun` → `runLoop` in [`agent/loop.go`](agent/loop.go).

Each iteration: cancel/pause checks → optional compaction → LLM completion with tools → dispatch tool calls → checkpoint.

Meta tools: `update_plan`, `delegate`, `task_complete`. Registry tools: filesystem, shell, search, git, test, versions, memory, optional browser_fetch.

## Engineering evidence

[`agent/engineering.go`](agent/engineering.go) classifies tasks (`trivial` / `nonTrivial`) and tracks phases + evidence.

| Gate | When | Requirement |
|---|---|---|
| Write gate | Agent + non-trivial + mutating filesystem/git | Inspect evidence first |
| Complete gate | Any file mutations in Agent mode | Verify evidence (test/build/lint) |
| Review gate | Non-trivial + mutations | Structured self-review `pass` (auto-run if missing) |

Events: `agent.phase` carries `{phase, taskClass, evidence, review}` for the UI.

## Policy (security below the prompt)

[`tools/policy.go`](tools/policy.go) + workspace path resolution:

- Modes: supervised / readonly / autonomous (Ask/Plan force readonly)
- Protected writes; **secret-pattern read deny** (`.env*`, keys, pem, …)
- Symlink-aware workspace boundary (`EvalSymlinks`)
- Destructive shell + git commit approval
- Browser: allowlist + private IP/metadata block + redirect re-check
- Subagent tool allowlists by role (real tool names)

## Context

- Base identity: [`agent/context.go`](agent/context.go)
- Always-on rules: embedded builtins in [`rules/defaults`](rules/defaults) via [`rules`](rules/rules.go); workspace `.blackjak/rules/*.md` overrides by name
- Skills: embedded builtins in [`skills/defaults`](skills/defaults); workspace `.blackjak/skills/<name>/SKILL.md` overrides by name
- Path index + repo snapshot: [`workspace/index.go`](workspace/index.go)
- Retrieval snippets: [`context/retrieval.go`](context/retrieval.go)
- Compaction at ≥80% usable budget; structured continuity summary

## Surfaces

```
IDE / Desktop / Browser UI
        │ WebSocket + HTTP
        ▼
   api.Server  →  agent.Agent  →  tools + LLM providers
```

Do not reintroduce prompt-only behavior for gates or security — enforce in Go.
