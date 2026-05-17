# Design: Knowledge Pipeline — Auto-ingest, Classification, and Federation

## Approach

### 1. File Watcher (`ingest/watcher.go`)
```go
// Watch starts fsnotify on a directory tree, debounces changes,
// and calls the classifier + store pipeline for each new/modified .md file.
func Watch(dir string, project string, store *Store, embedder Embedder) error
```
- Uses `fsnotify` for cross-platform file watching
- Debounce: 500ms window to batch rapid changes
- Ignores `.git/`, `node_modules/`, `*.tmp` patterns
- Tracks `file_hash` in `vec_memories` to skip unchanged files

### 2. Auto-Classifier (`ingest/classifier.go`)
Two-tier classification (fast path → LLM fallback):

**Tier 1 — Keyword heuristics (offline, instant):**
```
file path contains "postmortem"      → type: postmortem
file path contains "spec/" or *.spec.md → type: spec
file path contains "adr/" or "decision" → type: decision
content matches "fix|bug|crash|error"   → type: bug
content matches "## Why|## Decision"    → type: decision
content has [x] and [ ] checkboxes      → type: task
confidence < 0.7                        → escalate to Tier 2
```

**Tier 2 — LLM classification (Ollama, local):**
```
Prompt: "Classify this file. Return JSON: {type, outcome, tags, summary, relates_to_project}
File: <content>"
Model: llama3.2 or mistral (lightweight, local)
Outcome: only prompt if confidence from Tier 1 is low
```

### 3. Federation (`vec/store.go` — new method)
```go
func (s *Store) MergeProjects(projects []string, targetProject string) (*MergeResult, error)
```
- Copies all memories from source projects to target
- Adds `source_project` tag to preserve origin
- Deduplicates by semantic similarity (>0.95 threshold)
- In the merge result: deduped count, conflict count, new count

### 4. MCP Tools
- `mem_ingest_file {path, project}` — reads file, classifies, embeds, saves
- `mem_merge_projects {projects: ["a","b"], into: "global"}` — federates

### 5. CLI Interface
```bash
# Watch mode (long-running)
mnemo ingest --watch ./docs/ --project banking-app

# One-shot ingest
mnemo ingest --file ./postmortems/incident-42.md --project banking-app

# Merge
mnemo merge --projects core-banking,payment-service,api-gateway --into global
```

## Alternatives Considered
1. **Inotify/kqueue direct** — Platform-specific, more code. fsnotify is the Go standard, cross-platform, battle-tested.
2. **Only LLM classification** — Too expensive and slow for every file. Two-tier gives speed + accuracy.
3. **Separate DB per project** — Would need connection management, can't merge. Single DB with `project` field is simpler and federation is just a query.
4. **Chroma/LanceDB** — Format lock-in. `.mempack` with text-first model-agnostic approach means any model can regenerate embeddings.

## Decision
Two-tier classification (keyword + LLM fallback). Single DB, project field for isolation. `mnemo merge` for federation. Text-first `.mempack` format: content always included, embeddings are optional cache. Model-agnostic by design — import regenerates embeddings if model differs.
