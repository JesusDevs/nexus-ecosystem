# Gingx-SDD Architecture

## Design Principles

1. **Markdown-first.** Skills, personas, specs — all portable markdown. No lock-in.
2. **Agent-Agnostic.** Works with Claude Code, OpenCode, Cursor, Codex, Antigravity.
3. **Spec-Driven.** No code without spec. No spec without BDD (Gherkin).
4. **Memory-First.** Every decision, bug, pattern persisted via mnemo vector DB.
5. **Minimal Python.** Only typer + rich. All heavy logic in Go (mnemo) or markdown.
6. **Zero-Friction.** `./install.sh` auto-configures everything.

---

## Project Structure

```
monorepo/                          # Gingx Ecosystem
├── gingx-mnemo/                   # Go — Vector memory + MCP server
│   ├── vec/                       #   store.go (SQLite), embed.go (bge-m3)
│   ├── mcp/                       #   server.go (JSON-RPC 2.0 over stdio)
│   ├── main.go                    #   CLI (search, save, config, setup, mcp)
│   └── install.sh
├── gingx-sdd/                     # SDD Harness (markdown + bash + minimal Python)
│   ├── gingx_sdd/                 #   ← paquete Python (solo typer + rich)
│   │   ├── cli.py                 #     spec, orchestrate, release, status
│   │   └── orchestrate.py         #     decompose + delegate + track
│   ├── skills/team/               #   7 personas markdown (core)
│   ├── extras/skills/             #   skills opcionales (--full install)
│   ├── templates/.gingx/          #   configs auto-copiadas al instalar
│   └── install.sh
├── openspec/                      # OpenSpec changes (HDUs activos)
├── .gingx/                        # Config local + profiles
│   ├── config.yaml                #   30-harness YAML
│   └── profiles/
└── AGENTS.md                      # Root-level agent instructions
```

**`gingx-sdd/` vs `gingx_sdd/`**: El primero es el proyecto (docs, templates, skills, install.sh). El segundo es el paquete Python mínimo (`cli.py` + `orchestrate.py`) que se instala con `pip install -e .`. Sin frameworks pesados — solo typer + rich.

---

## Layer Architecture

```
┌──────────────────────────────────────────────────────────┐
│  LAYER 1: Markdown Personas (skills/team/*.md)           │
│  Supervisor, PO, UX, Architect, Dev, QA, DevOps          │
│  Portable — same .md works across Claude Code, Codex,    │
│  OpenCode, Kiro, Antigravity                             │
├──────────────────────────────────────────────────────────┤
│  LAYER 2: CLI (gingx_sdd/)                                │
│  typer + rich — spec, orchestrate, save, status, release │
│  Delegates to mnemo (Go) via subprocess                  │
├──────────────────────────────────────────────────────────┤
│  LAYER 3: Mnemo Vector Memory (Go)                       │
│  SQLite + bge-m3 embeddings + cosine similarity           │
│  MCP server (JSON-RPC 2.0 over stdio, 8+ tools)          │
├──────────────────────────────────────────────────────────┤
│  LAYER 4: Harness Engineering                             │
│  Guides (skills, AGENTS.md, profiles)                     │
│  Sensors (gitleaks, build, tests, mnemo conflicts)        │
│  Steering (Claude Code hooks, hook-driven auto-correction)│
│  Orchstration (DAG + Supervisor + Swarm — hybrid mode)    │
├──────────────────────────────────────────────────────────┤
│  LAYER 5: OpenSpec (openspec/)                            │
│  proposal.md + specs/*.md + design.md + tasks.md          │
│  BDD scenarios (Gherkin) in every spec                    │
└──────────────────────────────────────────────────────────┘
```

---

## Multi-Agent Orchestration

### Phase → Agent Mapping

```
explore  → architect-agent   (investigate codebase, find prior art)
propose  → po-agent          (define scope, write proposal)
spec     → po-agent          (BDD scenarios, acceptance criteria)
design   → architect-agent   (system design, trade-off analysis)
tasks    → dev-agent         (break into implementation tasks)
apply    → dev-agent         (implement, test-first)
verify   → qa-agent          (adversarial testing, BDD validation)
security → devops-agent      (secret scan, dependency audit)
archive  → supervisor        (save to mnemo, move to archive)
```

### Swarm Modes

| Mode | Behavior |
|------|----------|
| `dag` | Dependency-only — max parallelism, no delegation |
| `supervisor` | Centralized delegation to specialist agents |
| `swarm` | Distributed claim-based from shared queue |
| `hybrid` | DAG resolves + Supervisor assigns + Swarm executes (default) |

Mode is read from `vec_config` at the start of each phase — no restart needed.

---

## Harness Engineering Model

Following Birgitta Böckeler's Thoughtworks harness pattern:

### Feedforward Guides (before work)
- `AGENTS.md` — project-level instructions loaded at session start
- `skills/team/*.md` — 7 agent personas with role-specific rules
- `.gingx/profiles/` — per-agent configuration overrides
- `.gingx/config.yaml` — project harness configuration

### Feedback Sensors (during/after work)

| Sensor | Type | Tool | Trigger |
|--------|------|------|---------|
| Security scan | Computational | gitleaks | commit, release |
| Build check | Computational | go build / python compile | release |
| Test coverage | Computational | go test / pytest | release |
| Dependency audit | Computational | pip-audit, go-vuln | release |
| Mnemo conflicts | Inferential | mnemo conflicts | release |
| QA adversarial | Inferential | qa-agent persona | verify phase |

### Steering Loop
```
Sensor detects issue
    ↓
Hook evaluates (PreToolUse, PostToolUse, Stop)
    ↓
Behavior adjusted (block, warn, auto-fix)
    ↓
Pattern saved to mnemo (cross-session learning)
```

---

## Mnemo Memory Flow

```
Agent completes significant work
    ↓
gingx-sdd save --hdu-id HDU-NNN
    ↓
mnemo save (embeds via Ollama bge-m3)
    ↓
SQLite vec_memories table (float32 vectors)
    ↓
Next session:
    mnemo search "<query>" --project <project>
    ↓
Top-K cosine similarity results injected into agent context
```

Decision → save rationales. Bugs → save root causes. Patterns → save with outcome. Over 3 sessions, the agent develops institutional knowledge.

---

## Cross-Language Bridge

```
Python (gingx-sdd)          Go (gingx-mnemo)
─────────────────          ─────────────────
orchestrate.py              main.go
    │                           │
    ├─ subprocess.run ──────→ mnemo search
    ├─ subprocess.run ──────→ mnemo save
    ├─ subprocess.run ──────→ mnemo release
    └─ subprocess.run ──────→ mnemo config

All communication via CLI subprocess.
No shared library, no RPC, no network dependency.
Mnemo is a single static binary.
```

---

## Release System

```
gingx-sdd release v0.1.0 --project gingx-mnemo
    ├─ 1. Semver validation (v<MAJOR>.<MINOR>.<PATCH>)
    ├─ 2. Pre-release checks
    │     ├─ Git cleanliness
    │     ├─ Security scan (gitleaks)
    │     ├─ Build (go build / python compile)
    │     ├─ Tests (go test / pytest)
    │     └─ Mnemo conflicts
    ├─ 3. Git tag (annotated)
    ├─ 4. Mnemo release snapshot
    └─ 5. CHANGELOG.md update
```

---

## Skill Protocol

Skills follow a standard markdown frontmatter format:

```yaml
---
name: agent-id
description: What this agent does
when_to_use: Trigger conditions
model: claude-sonnet-4-6
effort: medium
---
# Agent Title

## Role
## Rules
## Checklist
## Mnemo Integration
```

All 7 team personas follow this format. Usable as Claude Code slash commands, OpenCode skills, or Codex prompts.
