# Proposal: Rename engram-vec to gingx-mnemo

## Why
The project grew beyond "Engram with vectors". It's now an independent memory system with MCP, versioning, exportable packs, and marketplace. The name `gingx-mnemo` reflects: belonging to the Gingx-SDD ecosystem + Mnemosyne (Greek goddess of memory). The `mnemo` binary is short, memorable, and distinct. Complete independence from Engram — own DB, own identity, own ecosystem.

## What Changes
- Module path: `github.com/gingx-sdd/engram-vec` -> `github.com/gingx-sdd/gingx-mnemo`
- Binary name: `engram-vec` -> `mnemo`
- Directory: `engram-vec/` -> `gingx-mnemo/`
- MCP server identity: `"engram-vec"` -> `"gingx-mnemo"`
- MCP registration: `claude mcp add engram-vec` -> `claude mcp add mnemo`
- DB path: `~/.engram/engram.db` -> `~/.mnemo/mnemo.db`
- Remove `engram_memory_id` column (legacy link)
- Remove `EngramMemoryID` from all structs
- Env vars: `ENGRAM_DIR` -> `MNEMO_DIR`, `ENGRAM_VEC_PROJECT` -> `MNEMO_PROJECT`
- README, install.sh, CHANGELOG updated
- References in gingx-sdd/install.sh

## What Does NOT Change
- Store API (same methods, updated signatures)
- MCP protocol (same tools, same handlers)
- Vec tables (same schemas minus the removed column)

## Impact
- HDU: HDU-04
- Complexity: low
- Files: go.mod, main.go, mcp/server.go, vec/store.go, README.md, install.sh, CHANGELOG.md, gingx-sdd/install.sh
- Breaking change: DB path changes (old DB needs manual migration)
