# Design: Portable Mnemo Memory via `.gingx/memory/`

## Approach

### High-Level Architecture

```
Project Repo                         Local Machine
├── .gingx/                          ~/.mnemo/
│   └── memory/                      ├── mnemo.db (SQLite, fast search)
│       ├── entries.jsonl  ←─── dual-write on save ───→ vec_memories table
│       └── embeddings.json ←─── optional vectors
│
└── .claude/hooks/
    └── session-start.sh ─── detects entries.jsonl → mnemo import
```

### Decision 1: JSONL Format for entries.jsonl

```
Decision: entries.jsonl uses one JSON object per line (JSONL), text-only fields
Alternatives:
  A. Single JSON array — pro: single parse call, con: not append-friendly, git conflicts on whole array
  B. JSONL (one per line) — pro: append-only, git-diffable per entry, streamable read
  C. YAML — pro: human-friendly, con: harder to parse in Go, append semantics unclear
Chosen: B (JSONL)
Trade-off: Must parse line-by-line (trivial with bufio.Scanner). No random access without full scan.
Reversal: If we need random access to entries without loading the whole file, switch to indexed format.
```

### Decision 2: Separate Embedding File

```
Decision: Vectors stored in separate embeddings.json, not inline in entries.jsonl
Alternatives:
  A. Inline embeddings in entries.jsonl — pro: single file, con: huge binary arrays make git diffs unreadable, bloats repo
  B. Separate embeddings.json — pro: entries.jsonl stays small and text-only, embeddings can be gitignored
Chosen: B (separate)
Trade-off: Two files can desync. Mitigation: import always prefers entries.jsonl as source of truth; embeddings are regenerable.
Reversal: If two-file complexity causes bugs, fall back to inline (accepting repo bloat).
```

### Data Flow

#### Save (dual-write)

```
mnemo save <title> <content>
  │
  ├─ 1. Generate embedding via Ollama (existing flow)
  ├─ 2. INSERT OR REPLACE into ~/.mnemo/mnemo.db (existing flow, unchanged)
  │
  └─ 3. Find .gingx/ directory (walk up from cwd, max 5 levels)
        │
        ├─ NOT FOUND → silently skip (graceful degradation)
        │
        └─ FOUND .gingx/memory/
              ├─ 4a. Append one JSON line to entries.jsonl
              │      Fields: id, project, title, content, type, tags,
              │              outcome, media_type, version, created_at
              │              (NO embedding binary — text only)
              │
              └─ 4b. Update embeddings.json
                     Structure: { "id1": [0.1, 0.2, ...], "id2": [...] }
                     Upsert the entry's ID → embedding mapping
```

#### Import (on clone or explicit)

```
mnemo import [--path <file>] [--reindex] [--yes]
  │
  ├─ 1. Resolve entries path
  │     ├─ --path specified → use that file
  │     └─ no --path → walk up from cwd to find .gingx/memory/entries.jsonl
  │
  ├─ 2. Read entries.jsonl line by line (bufio.Scanner)
  │
  ├─ 3. For each entry:
  │     ├─ Check mnemo.db: does ID exist?
  │     │   ├─ EXISTS → skip, increment "skipped" counter
  │     │   └─ NEW →
  │     │       ├─ Load embedding from embeddings.json (if --reindex not set)
  │     │       │   ├─ FOUND → reuse it
  │     │       │   └─ NOT FOUND → generate via Ollama embedder
  │     │       │       └─ If Ollama unavailable → warn, skip this entry
  │     │       │
  │     │       └─ INSERT into vec_memories
  │     │           (id, project, title, content, type, embedding,
  │     │            embedding_model, embedding_dim, tags, outcome,
  │     │            media_type, version)
  │     │
  │     └─ Increment "imported" counter
  │
  ├─ 4. If --reindex:
  │     └─ After importing all entries, regenerate ALL embeddings
  │        (useful when model changed, e.g., bge-large → bge-m3)
  │
  └─ 5. Report: "imported N, skipped M" (JSON on stdout for programmatic use)
```

#### session-start.sh auto-detect

```
On every session start:
  │
  ├─ 1. Check: does .gingx/memory/entries.jsonl exist?
  │     └─ NO → skip (nothing to import)
  │
  ├─ 2. Check: is mnemo binary available?
  │     └─ NO → skip (can't import without mnemo)
  │
  └─ 3. Run: mnemo import --yes 2>/dev/null
        ├─ Returns summary: "N imported, M skipped"
        └─ Include in systemMessage if N > 0:
           "Mnemo: imported N memories from .gingx/memory/"
```

### File Format Details

#### entries.jsonl (example)
```jsonl
{"id":"vec-gingx-ecosystem-bug-login","project":"gingx-ecosystem","title":"Bug: login timeout","content":"Fixed by increasing ttl","type":"bug","tags":["auth"],"outcome":"solucionado","media_type":"text","version":"","created_at":"2026-05-23 10:15:00"}
{"id":"vec-gingx-ecosystem-arch-oauth","project":"gingx-ecosystem","title":"Arch: OAuth2 flow","content":"Decided on PKCE","type":"decision","tags":["auth","oauth"],"outcome":"applied","media_type":"text","version":"v0.1.0","created_at":"2026-05-23 11:00:00"}
```

#### embeddings.json (example)
```json
{
  "vec-gingx-ecosystem-bug-login": [0.0123, -0.0456, 0.0789, ...],
  "vec-gingx-ecosystem-arch-oauth": [0.0234, 0.0567, -0.0890, ...]
}
```

### Finding the `.gingx/` Directory

```go
// Walk up from cwd, max 5 levels (same as git's discovery)
func findGingxDir() (string, error) {
    cwd, _ := os.Getwd()
    dir := cwd
    for i := 0; i < 5; i++ {
        gingxDir := filepath.Join(dir, ".gingx")
        if info, err := os.Stat(gingxDir); err == nil && info.IsDir() {
            return gingxDir, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break // reached root
        }
        dir = parent
    }
    return "", fmt.Errorf(".gingx/ not found")
}
```

### Changes to Existing Code

| File | Change | Lines |
|------|--------|-------|
| `gingx-mnemo/main.go` | Add `import` subcommand (`runImport()`) | ~30 |
| `gingx-mnemo/main.go` | Modify `runSave()`: append dual-write | ~15 |
| `gingx-mnemo/vec/store.go` | Add `ExportEntryJSON()` for single-entry serialization | ~10 |
| `.claude/hooks/session-start.sh` | Add auto-detect + `mnemo import` block | ~6 |
| Template `session-start.sh` | Mirror the same block | ~6 |

### What We Reuse

- **`MemoryPackEntry` struct** — already defines all portable fields (ID, Project, Title, Content, Type, Embedding, Tags, Outcome, MediaType, Version)
- **`ImportPack()` method** — already handles INSERT OR REPLACE for bulk import; import subcommand can build on this
- **`Embedder.Embed()` interface** — for regenerating missing embeddings
- **`bytesToFloats` / `floatsToBytes`** — existing encoding utilities

### Edge Cases Handled

| Case | Behavior |
|------|----------|
| `.gingx/` directory does not exist | Save silently skips dual-write. Import reports "nothing to import". |
| `entries.jsonl` exists but empty | Import does nothing, reports "0 imported, 0 skipped". |
| `embeddings.json` missing | Import generates embeddings. `--reindex` rewrites it. |
| Ollama not available during import | Report warning, skip entries without embeddings. User can run `mnemo import --reindex` after starting Ollama. |
| Concurrent writes to `entries.jsonl` | File append is atomic on POSIX for writes < PIPE_BUF. Our writes are single-line JSON, well under 4KB. No explicit locking. |
| Entry with same ID already in DB | `INSERT OR REPLACE` handles idempotency. Counted as "skipped". |
| Very large `entries.jsonl` (1000+ entries) | Streamed line-by-line, memory O(1) per entry. |

### .gitignore Recommendation

```gitignore
# .gingx/memory/embeddings.json is fully regenerable
# Uncomment to keep repo size small:
# .gingx/memory/embeddings.json
```

entries.jsonl is NEVER gitignored — it is the source of truth for portable memory.

## Alternatives Considered

1. **Git-based sync only (current sync.remote)** — Requires manual config per machine. No clone-time discovery. Rejected: not portable.
2. **YAML instead of JSONL** — More human-readable but harder to parse in Go without a YAML library (new dependency). Rejected: adds dependency for marginal benefit.
3. **SQLite backup (.sql dump) instead of JSONL** — Binary, not git-diffable, ties format to SQLite schema. Rejected: JSONL is future-proof.
4. **Store only in .gingx/memory/, drop mnemo.db** — Would lose local caching and fast search. Rejected: DB serves different purpose (local index).
5. **Git LFS for embeddings** — Adds external dependency. Rejected: embeddings are small (1024 x 4 bytes = 4KB per entry) and regenerable.
