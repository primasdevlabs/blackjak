# Go (first-class backend)

When working in Go:

- Prefer idiomatic Go and the standard library over framework-heavy stacks
- Explicit error handling; propagate context; respect cancellation
- Structured concurrency: bounded goroutines, clear ownership, no leaks
- Interfaces at consumption boundaries, not for every type
- Tables tests, race detection when concurrency changes, gofmt/go vet as verification
- Package design by responsibility; avoid Java/C#/TS layering conventions
- Dependency management: current + stable + maintained + secure + compatible

Do not introduce frameworks when stdlib or a small dependency is sufficient.
