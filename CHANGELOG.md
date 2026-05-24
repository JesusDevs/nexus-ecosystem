# Changelog — Gingx Ecosystem

Generated from SDD harness (`gingx-sdd changelog`). Each entry comes from completed HDUs in `openspec/changes/`.

---

## v0.3.1 — Knowledge Graph + Explorer Agent (2026-05-24)

### New Features

**Explorer Agent** (9th agent persona)
- First responder for codebase questions. Maps domains, discovers patterns.
- Maintains knowledge graph: domain-map.yaml, component-index.yaml, decisions-log.yaml
- Auto-triggered on SessionStart, before architect-agent

**Knowledge Graph Layer**
- `domain-map.yaml`: 7 domains with key symbols, dependencies, patterns, cross-domain edges
- `component-index.yaml`: 9 components with callers, callees, tests, domain assignment
- `decisions-log.yaml`: architecture decision records with rationale and trade-offs
- Auto-generated on first session start, incrementally updated on every stop
- `gingx-sdd knowledge vault` — Obsidian vault with [[wikilinks]] and native graph view
- `gingx-sdd knowledge status|search|explore|save-decision` — knowledge CLI ops

**Graphify Integration**
- Full codebase graph: 2,467 nodes, 4,547 edges, 157 communities
- AST-only incremental update in stop hook (code changes auto-refresh, no LLM cost)

**Obsidian Vault Auto-Refresh**
- Stop hook regenerates vault on every session end
- Home.md (MOC) + domains/ + components/ + .obsidian/graph.json

### Improvements

- README complete rewrite with practical usage: init projects, add skills/hooks/agents
- Phase agent map: explore routes to explorer-agent (was architect-agent)
- Stop hook graphify call fixed (broken → Python API incremental)
- Auto-update matrix: knowledge graph + vault + graphify + mnemo all in hooks

---

## v0.3.0 — Autonomous Goal System (2026-05-24)

### New Features

**Goal System for Autonomous Agents**
- `gingx-sdd goal` command group: `create`, `list`, `status`, `loop`, `complete`
- `goal-agent` persona: autonomous executor that works without human interaction
- `goal-autonomous` profile: reduced interrogation, no human-in-the-loop, designed for overnight/weekend work
- `goal` tech stack skill: GoalGraph pattern with plan-act-observe-reflect cycle
- `GoalStore` / `GoalState`: YAML persistence in `.gingx/goals/<id>.yaml` with mnemo sync
- `ai-goal` suite: pre-packaged langgraph-python + goal + fastapi + bdd-behave

**GoalGraph Pattern** (LangGraph)
- StateGraph with GoalState (objective, key_results, progress, history, iteration)
- 4-node cycle: planner → executor → observer → reflector
- Completion criteria: all KRs at 100% OR max_iterations reached
- Checkpointed execution for resilience between sessions

### Improvements

- `gingx-sdd init --stack langgraph` now includes goal-agent.md persona and `.gingx/goals/` directory
- `fullstack-python-langgraph` profile extended with goal-agent
- `langgraph-python` skill updated with GoalGraph pattern reference
- team spawn help now lists goal-agent

---

## v0.2.0 — Ecosystem Foundation (2026-05-23)

### New Features

**Portable Mnemo Memory (HDU-08)**
- `.gingx/memory/entries.jsonl` travels with the repo — clone a project, get its agent memory
- `embeddings.json` is gitignored and regenerable via `mnemo import --reindex`
- Dual-write on every `mnemo save` — DB + JSONL simultaneously
- Auto-import on session start via hooks — zero-config knowledge transfer between teams
- `mnemo import` command reads entries.jsonl line-by-line, checks for duplicates, inserts into local DB

**`gingx-sdd init` Scaffolding Command (HDU-09)**
- Single command bootstraps a complete Gingx project: `gingx-sdd init`
- Creates 23 files across `.gingx/`, `.claude/`, `openspec/`, `.mcp.json`
- `--dry-run` flag to preview without writing files
- `--force` flag to overwrite existing `.gingx/`
- `--stack` flag to force a specific tech stack (python, go, react, node, langgraph, minimal)
- 7 SDD agent personas scaffolded into `.claude/agents/`
- 3 enforcement hooks with executable permissions (`chmod 755`)
- Stack auto-detection via existing `detector/scanner.py`

### Improvements

- **Agent personas included in init**: architect, dev, devops, po, qa, supervisor, ux (7 total)
- **Profiles scaffolded**: developer, fullstack, fullstack-go, fullstack-python-langgraph, minimal, react-nextjs, team
- **Hooks auto-installed**: PreToolUse (blocks Write without spec), Stop (auto-saves to mnemo), SessionStart (loads context + imports memory)
- **Go Gingx skill** (`/go-gingx`): Claude Code skill with ecosystem Go conventions, CLI patterns, embedded templates, porting guide
- **Deprecation hints** on `mnemo sync push/pull` — git-based sync deprecated in favor of `.gingx/memory/` portable approach
- **Ecosystem-level CHANGELOG.md** with auto-generation via `gingx-sdd changelog`

### Architecture

| Component | Language | Role |
|-----------|----------|------|
| gingx-mnemo | Go | Vector memory, MCP server, embeddings |
| gingx-sdd | Python | SDD harness, init, orchestration |
| agent/ | (future) | Banking agent runtime |
| frontend/ | (future) | Banking UI |

**Language rationale**: Go handles performance-critical vector operations. Python handles workflow orchestration. Each tool in its strength. CLI migrates to Go progressively.

---

## v0.1.0 — Initial Release (2026-05-02)

- Rename: `nexus-ecosystem` → `gingx-ecosystem`
- Mnemo: semantic search via Ollama BGE-M3 embeddings, cosine similarity in SQLite
- SDD: 8-phase pipeline (init → explore → propose → spec → design → tasks → apply → verify → archive)
- 7 agent personas: supervisor, po, ux, architect, dev, qa, devops
- 6 agent profiles: developer, fullstack, fullstack-go, react-nextjs, minimal, fullstack-python-langgraph
- 3 SDD enforcement hooks: PreToolUse gate, Stop auto-save, SessionStart context
- Agent Factory System: spawnable sub-agents with profiles and tech stacks
- Design memory system: color palettes, font pairings, UX heuristics in vector DB
- 30 harness contracts in `.gingx/config.yaml`
