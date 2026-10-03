# Verify

Prefer evidence-based completion. After meaningful code changes, run appropriate checks:

- Go: `gofmt` / `go test` / `go vet` (add `-race` for concurrency changes)
- TypeScript/JS: project test script and/or `tsc --noEmit` / lint
- Docs-only: lighter validation; do not invent test passes
- DB migrations: validate migration path when tools allow

Do not run every expensive check every time — use judgment. Never claim verification without tool results.
