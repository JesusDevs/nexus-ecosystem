# Proposal: Pack Import and Memory Marketplace

## Why
`mnemo pack export` already generates portable `.mempack` JSON. But there's no way to IMPORT that pack on another machine or team. Without import, the "memory marketplace" is one-way export only. Teams need to share knowledge via git and others need to import with conflict detection and merge strategies.

## What Changes
- `mnemo pack import <file>` — imports `.mempack` JSON into local DB
- Schema validation of the pack (pack_version, compatibility)
- `mnemo pack install <url>` — installs from a GitHub repo
- Conflict detection on import (duplicate IDs)
- Merge strategies: skip, overwrite, keep-both
- `mem_pack_import` MCP tool so agents can import without shell
- `mnemo pack search <query>` — search relevant packs by semantics
- Available packs index (marketplace)

## What Does NOT Change
- `.mempack` format (already defined in `ExportPack`)
- Embeddings (always included in pack, never regenerated)
- `vec_memories` table (import uses the same schema)

## Impact
- HDU: HDU-02
- Complexity: medium
- Files: vec/store.go, mcp/server.go, main.go
- New CLI commands: 3 (pack import, pack install, pack search)
- New MCP tools: 1 (mem_pack_import)
