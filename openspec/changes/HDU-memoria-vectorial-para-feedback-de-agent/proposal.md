# Proposal: Vector Memory for Agent Feedback

## Why
AI agents start from zero in every session. Keyword search (FTS5) only matches exact words. We need the agent to autonomously SAVE (via MCP) what it learns (bugs, decisions, feedback) so another agent in another project can retrieve it by meaning, not by keywords.

## What Changes
- New MCP tool `mem_save` so agents can save without using shell
- `media_type` field (text, image, pdf, audio, video) for multimodal support
- `version` field for versioning memories by release
- `vec_releases` table for knowledge snapshots
- CLI commands: `mnemo release`, `mnemo diff`, `mnemo releases`, `mnemo pack export`
- New MCP tools: `mem_release`, `mem_diff`, `mem_list_releases`
- `nexus-sdd save --hdu-id` for harness integration
- `nexus-sdd release <version>` wrapper

## What Does NOT Change
- Existing databases are untouched (new tables, not altered)
- Ollama + bge-large-en-v1.5 as embedding engine
- Manual cosine similarity (no external indexes)
- Zero external APIs

## Impact
- HDU: HDU-memoria-vectorial-para-feedback-de-agent
- Complexity: medium
- Files: vec/store.go, mcp/server.go, main.go, nexus_sdd/cli.py
- MCP tools: 4 -> 8
