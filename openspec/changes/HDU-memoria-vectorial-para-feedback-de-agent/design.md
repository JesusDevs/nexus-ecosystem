# Design: Vector Memory for Agent Feedback

## Approach
SQLite inside `~/.mnemo/mnemo.db` (standalone). New tables:
- `vec_memories`: float32 embeddings + metadata + media_type + version
- `vec_releases`: version snapshots per project

MCP via JSON-RPC 2.0 over stdio. 8 tools total. Embeddings via local Ollama (bge-large-en-v1.5, 1024-dim). O(n) manual cosine similarity (sufficient for <100K memories).

The harness (nexus-sdd) calls `mnemo` via subprocess for save/release. Will use MCP directly in the future.

## Alternatives Considered
1. pgvector / Pinecone — Overkill. Adds external dependency. SQLite + cosine is sufficient for the use case (team memories, not millions).
2. ChromaDB — Similar problem. Extra Python dependency. We want pure Go and zero external dependencies.
3. ANN indexes (FAISS) — Unnecessary complexity for <100K vectors. Direct cosine is O(n) but for 10K-50K memories it's <1ms.

## Decision
SQLite + manual cosine similarity. Zero external dependencies. Standalone DB. Local embeddings with Ollama. MCP so any agent (Claude Code, Codex, Cursor) can use it.
