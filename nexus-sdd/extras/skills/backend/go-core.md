---
name: go-core
description: Core Go patterns — interfaces, goroutines, channels, error handling, testing
category: backend
stack: [go, sqlite, stdlib]
triggers: [go, backend, api, cli, mcp]
---

# Go Core Patterns

## Rules
- Accept interfaces, return structs
- Errors are values — handle them, don't panic
- `defer` for cleanup, close in reverse order of acquisition
- Goroutines: always know how they stop. Use `context.Context` for cancellation.
- Channels: unbuffered for sync, buffered for async with known capacity
- `sync.WaitGroup` for waiting on goroutines, `sync.Mutex` for shared state
- Package names: short, lowercase, no underscores (except `_test`)

## Do's
- `go mod tidy` after every dependency change
- Table-driven tests with `testing` + `testify` for assertions
- Use `context.Context` as first parameter for all functions doing I/O
- WAL mode for SQLite: `?_journal_mode=WAL`
- `embed` for static files (templates, migrations, configs)
- `go fmt` on save (automatic in most editors)

## Don'ts
- No `init()` with side effects — except for registering drivers
- No `panic` in library code — return errors
- No global `var` for mutable state — pass it explicitly
- No `goto` — use labeled loops for complex breaks
- No `interface{}` (any) when generics or specific interfaces work

## Recommended Commands
```bash
go build ./...
go test ./...
go vet ./...
go fmt ./...
golangci-lint run
```
