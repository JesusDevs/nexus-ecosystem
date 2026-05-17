# Proposal: Knowledge Pipeline — Auto-ingest, Classification, and Federation

## Why
Teams have knowledge scattered across docs, postmortems, ADRs, and decisions in their repos. Currently each file must be manually saved via `mnemo save` or `mem_save`. This is friction — teams won't do it consistently. We need auto-detection of new/modified markdown files, automatic classification by type and relevance, and the ability to merge knowledge from multiple projects into a single federated base. One dev should be able to clone a repo and inherit the entire team's knowledge instantly.

## What Changes
- `mnemo ingest --watch <dir> --project <name>` — fsnotify watcher that auto-ingests new/modified files
- Auto-classifier: determines type (spec, decision, bug, postmortem), outcome, and tags from file content
- `mnemo merge --projects a,b,c --into global` — federate multiple projects into one knowledge base
- `mem_ingest_file` MCP tool — agent-driven file ingestion
- `mem_merge_projects` MCP tool — agent-driven project federation
- Relationship detection: when ingesting, check semantic similarity against existing memories and link related ones
- Classification via local LLM (Ollama) with fallback to keyword heuristics

## What Does NOT Change
- `vec_memories` schema (project field already exists)
- `.mempack` format (model-agnostic: text always included, embeddings optional cache)
- MCP protocol (JSON-RPC 2.0 over stdio)
- Zero external APIs (fsnotify is the only new Go dependency)

## Impact
- HDU: HDU-05
- Complexity: medium
- Files: vec/store.go, mcp/server.go, main.go, new: ingest/watcher.go, ingest/classifier.go
- New CLI commands: 2 (ingest, merge)
- New MCP tools: 2 (mem_ingest_file, mem_merge_projects)
- New dependency: `github.com/fsnotify/fsnotify`
