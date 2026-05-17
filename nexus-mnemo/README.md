# Mnemo — Versionable Vector Memory for AI Agents

> *"Mnemo doesn't just remember. It versions, transfers, and shares knowledge across teams."*

Standalone vector memory system with semantic search, versioning, and portable knowledge packs. Part of the [Nexus-SDD](https://github.com/JesusDevs/nexus-sdd) ecosystem.

```
Keyword search:    "Google login" → matches "login", "Google"
                    Does NOT find "OAuth authentication"

Mnemo (vector):   "Google login" → semantic embedding
                    FINDS "OAuth authentication", "SSO provider"
                    Versioned: you know which release learned it
                    Portable: share memory packs across teams
```

## Quick Start

```bash
# 1. Install
git clone https://github.com/nexus-sdd/nexus-mnemo.git
cd nexus-mnemo && ./install.sh

# 2. Save memories
mnemo save "Race condition in Stripe payments" \
  "Two simultaneous webhooks caused double charge. Fix: idempotency keys." \
  --type bug --outcome resolved \
  --tags stripe,payments,webhook

# 3. Search by meaning
mnemo search "payment processing issues"

# 4. Create release (version knowledge)
mnemo release banking-app v1.0.0 --description "MVP with auth + payments"

# 5. Compare releases
mnemo diff banking-app v1.0.0 v1.1.0

# 6. Export portable pack
mnemo pack export banking-app --version v1.1.0 --output banking.mempack

# 7. Transfer to another project
mnemo transfer "implement payments" my-new-project

# 8. Detect conflicts
mnemo conflicts --project my-app
```

## 8 MCP Tools

| Tool | Description |
|---|---|
| `mem_search_semantic` | Search by meaning |
| `mem_similar` | Find similar memories |
| `mem_transfer` | Transfer lessons between projects |
| `mem_save` | Save memories (bugs, decisions, feedback) |
| `mem_release` | Create version snapshot |
| `mem_diff` | Compare memories between versions |
| `mem_list_releases` | List releases for a project |
| `mem_detect_conflicts_semantic` | Detect contradictions |

## Knowledge Marketplace

```bash
# Export your team's knowledge
mnemo pack export langgraph --version v2.1.0 --output langgraph-v2.1.0.mempack

# Share via Git
git push github.com/team/langgraph-memories

# Another team installs it
mnemo pack install github.com/team/langgraph-memories --version v2.1.0
# → 342 memories installed. The agent already knows about checkpoints, DynamoDB, TTL...
```

## Integration

### Claude Code
```bash
claude mcp add mnemo -- mnemo mcp
```

### OpenCode
```bash
opencode mcp add mnemo -- mnemo mcp
```

### Cursor / Windsurf / Any MCP
```json
{
  "mcpServers": {
    "mnemo": {
      "command": "mnemo",
      "args": ["mcp"]
    }
  }
}
```

### Nexus-SDD
The [Nexus-SDD](https://github.com/JesusDevs/nexus-sdd) installer includes Mnemo automatically.

## How It Works

```
Agent (Claude Code / OpenCode / Codex / Cursor)
    ↓ MCP stdio (JSON-RPC 2.0)
mnemo mcp
    ↓
~/.mnemo/mnemo.db
    ├── vec_memories        (float32 embeddings + versioning + media)
    ├── vec_projects        (project index)
    └── vec_releases        (version snapshots)
    ↓
Ollama (bge-large-en-v1.5, local, zero-cost)
    ↓
1024-dim embedding + cosine similarity
```

## Architecture

```
nexus-mnemo/
├── main.go              # CLI + commands (save, search, release, diff, pack...)
├── vec/
│   ├── store.go         # Vector store (SQLite + cosine + versioning)
│   └── embed.go         # Local embeddings via Ollama
├── mcp/
│   └── server.go        # MCP server (8 tools, JSON-RPC 2.0 over stdio)
└── install.sh           # Zero-friction installer
```

## Zero External APIs

- **Ollama** (local model server)
- **bge-large-en-v1.5** (open-source embedding model, 1024 dims)
- **SQLite** (embedded database, zero config)
- **Cosine similarity** in Go

Zero recurring costs. Zero data leaves your machine.

## License

MIT
