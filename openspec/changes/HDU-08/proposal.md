# Proposal: Portable Mnemo Memory via `.gingx/memory/`

## Why
Mnemo stores all memory in `~/.mnemo/mnemo.db` — a global SQLite database outside the project repo. This means:
- When someone clones the repo, the project's accumulated knowledge (decisions, bugs, architectural notes) does NOT travel with it
- `sync.remote` requires manually configuring a git remote URL, which is fragile and not project-portable
- Each new contributor starts with zero context from past decisions
- Currently 2 entries for `gingx-ecosystem`, 2 for old `nexus-ecosystem` — none are portable

The **user need**: "When I clone a Gingx project, I should also get its accumulated agent memory — with zero configuration."

## What Changes
- **New `.gingx/memory/` directory** inside the project repo, containing:
  - `entries.jsonl` — serialized memory entries (text-only, git-friendly, diffable)
  - `embeddings.json` — vector embeddings (optional; regenerable via `mnemo import --reindex`)
- **`mnemo save` dual-writes**: saves to both `~/.mnemo/mnemo.db` (local cache for fast search) AND `.gingx/memory/entries.jsonl` (repo for portability)
- **New `mnemo import` subcommand**: reads `.gingx/memory/entries.jsonl` (and optionally `embeddings.json`) and indexes into the local `mnemo.db`
- **`session-start.sh` auto-detects** `.gingx/memory/entries.jsonl` and calls `mnemo import` if the local DB lacks those entries
- **`.gitignore` recommendation**: `.gingx/memory/embeddings.json` can be gitignored since it is fully regenerable

## What Does NOT Change
- `~/.mnemo/mnemo.db` remains the primary local cache for fast semantic search
- Existing `mnemo pack export/import` commands (different use case: explicit backup/transfer)
- `sync.remote` mechanism (deprecated but not removed; `.gingx/memory/` is the preferred path)
- Mnemo MCP server interface
- No new dependencies, no servers, no S3, no external services

## Impact
- HDU: HDU-08
- Complexity: low
- Dependencies: none
- Estimated lines: ~50 Go (import subcommand + dual-write in save), ~4 bash (session-start auto-detect)
- User experience: **zero config** — `git clone` + `mnemo import` auto-detected on first session
