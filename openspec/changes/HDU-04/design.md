# Design: Rename engram-vec to nexus-mnemo

## Approach
Surgical rename in 9 steps:
1. `server.go`: `Initialize()` name string, log messages
2. `main.go`: usage text, `version` command output, log prefixes, env vars
3. `go.mod`: module path
4. `vec/store.go`: DB path `~/.engram/engram.db` -> `~/.mnemo/mnemo.db`, remove `engram_memory_id` column and `EngramMemoryID` field
5. Imports in `main.go` and `mcp/server.go` (depend on #3)
6. `README.md`: title, examples, remove Engram extension framing
7. `CHANGELOG.md`: remove Engram references
8. `install.sh`: directory, binary, MCP registration
9. `nexus-sdd/install.sh`: references to binary and directory

Order matters: code first (1-5), then docs (6-7), then installers (8-9).

## Alternatives Considered
1. **Keep `engram-vec`** — Confusing: not part of Engram, it's standalone. Name doesn't reflect MCP or versioning.
2. **`mnemo` standalone without namespace** — Loses connection to Nexus-SDD ecosystem. The `nexus-` prefix indicates belonging.
3. **Keep old DB path for migration** — Would maintain backward compat but keep Engram reference forever. Better to break cleanly now.

## Decision
Full independence. `nexus-mnemo` as project name, `mnemo` as binary, `~/.mnemo/mnemo.db` as DB. No legacy columns, no env vars referencing Engram. Clean break.
