# Gingx Ecosystem v0.3.1

**Harness-driven AI development.** Spec → Plan → Code → Test → Security → Memory. Every phase is a contract. No code without an approved spec. Persistent vector memory that travels with the repo.

## In 30 Seconds

```bash
# 1. Install everything (one command)
curl -fsSL https://raw.githubusercontent.com/JesusDevs/gingx-ecosystem/main/gingx-sdd/install.sh | bash

# 2. Bootstrap any project with the full ecosystem
cd my-project
gingx-sdd init

# 3. Your project now has: 9 agents, 3 hooks, 19 skills, 8 profiles
#    - vector memory (mnemo) for semantic search across sessions
#    - knowledge graph (graphify) of the entire codebase
#    - spec gate that blocks code writes without an approved spec
#    - goal agent that runs objectives autonomously (e.g. overnight)

# 4. Create a spec and delegate to the right agent
gingx-sdd hdu create "My feature" --question "What's the right stack?"
gingx-sdd auto "implement health check endpoint"

# 5. Launch an autonomous goal
gingx-sdd goal create map-architecture \
  --objective "Map and document the entire architecture" \
  --key-results "KR1: complete domain map, KR2: component index, KR3: decisions log"

# 6. Everything auto-saves. Memory travels with the repo.
git add .gingx/memory/entries.jsonl && git commit -m "team memory"
```

## Why Python for the SDD Harness?

**Pragmatic, not ideological.** Each layer uses the right tool:

| Layer | Language | Why |
|-------|----------|-----|
| **Vector memory** (mnemo) | **Go** | Performance-critical: embeddings, cosine similarity, SQLite. Compiles to a single static binary. Zero dependencies. |
| **SDD harness** (gingx-sdd) | **Python** | Workflow orchestration: CLI (Typer), YAML config, LangGraph goal graphs, multi-agent dispatch, template rendering. Python is the best "glue" for stitching tools together. |
| **Knowledge graph** (graphify) | **Python** | tree-sitter AST parsing, community detection (Leiden), LLM extraction. Rich ecosystem. |
| **Agent personas** | **Markdown** | Portable across AI coding tools (Claude Code, Codex, Cursor, Kiro, Antigravity). |

The CLI will progressively migrate to Go for zero-dependency distribution. Python remains the orchestration layer until then — the harness logic is complex enough that rapid iteration in Python delivers more value than a Go port right now.

## What's Included

| Layer | Component | Language | What It Does |
|------|-----------|----------|--------------|
| **Memory** | [gingx-mnemo](gingx-mnemo/) | Go | Vector memory with SQLite + Ollama embeddings. 12 MCP tools. Portable via `.gingx/memory/entries.jsonl` |
| **SDD Harness** | [gingx-sdd](gingx-sdd/) | Python | 8-phase pipeline, 9 agents, 30 harness contracts, goal system, auto-delegation |
| **Knowledge** | Knowledge Graph | YAML + Obsidian + graphify | domain-map, component-index, decisions-log. Auto-generated, versioned, visualized |
| **Agents** | [.claude/agents/](.claude/agents/) | Markdown | 9 personas with triggers, rules, and tool access |
| **Hooks** | [.claude/hooks/](.claude/hooks/) | Bash | SessionStart (load context), PreToolUse (spec gate), Stop (persist progress) |
| **Skills** | [skills/](gingx-sdd/skills/) + [extras/](gingx-sdd/extras/skills/) | Markdown | 9 team skills + 19 tech stack skills across 6 categories |
| **Profiles** | [.gingx/profiles/](.gingx/profiles/) | YAML | 8 pre-built team compositions |

## The 9 Agents

| # | Agent | Trigger | What It Does |
|---|-------|---------|--------------|
| 1 | `supervisor` | Orchestrate HDUs | Decompose, delegate, track progress across phases |
| 2 | `explorer-agent` | SessionStart, "where/how" questions | Map the codebase, maintain the knowledge graph |
| 3 | `po-agent` | Define features | Specs, Gherkin scenarios, scope negotiation |
| 4 | `architect-agent` | Design systems | Trade-offs, DB schemas, API contracts, dependencies |
| 5 | `dev-agent` | Implement | TDD, test-first, code conventions |
| 6 | `qa-agent` | Verify | Adversarial testing, BDD validation, root cause |
| 7 | `ux-agent` | Review UI/UX | Accessibility (WCAG), usability, design patterns |
| 8 | `devops-agent` | Ship | CI/CD, security scans, dependency audits |
| 9 | `goal-agent` | Autonomous objectives | Executes goals without human interaction (overnight/weekend) |

## The 3-Phase Operation Cycle

```
SessionStart                  Active Session                   Stop
───────────                   ──────────────                   ────
Load context          →   9 agents work            →   Persist progress
Load knowledge graph       Spec gate active             Refresh Obsidian vault
Mnemo sync pull            Vector memory queries        Graphify AST update (code)
Active HDUs + blockers     Goal loops                   Mnemo sync push
Import portable memory     Auto-delegation              Knowledge export
```

## Installation

### Automatic (recommended)

```bash
# One command. Installs Python, Go, Node, Ollama, mnemo, gingx-sdd, graphify, codegraph.
curl -fsSL https://raw.githubusercontent.com/JesusDevs/gingx-ecosystem/main/gingx-sdd/install.sh | bash
```

What it does:
1. Detects OS (macOS, Linux, Windows)
2. Installs prerequisites: Python 3.11+, Go, Node 20+
3. Installs Ollama + BGE-M3 embedding model (local, zero API cost)
4. Compiles `mnemo` from source → `/usr/local/bin/mnemo`
5. Installs `gingx-sdd` Python package (`pip install -e`)
6. Creates `.gingx/` directory structure (config, profiles, tracking)
7. Configures Claude Code: copies hooks, agents, settings, MCP servers
8. Registers graphify skill (`graphify claude install`)
9. Detects project stack and installs matching skills
10. Initializes OpenSpec structure

### Manual

```bash
git clone https://github.com/JesusDevs/gingx-ecosystem.git
cd gingx-ecosystem

# Install mnemo (Go)
cd gingx-mnemo && go build -o mnemo . && cp mnemo /usr/local/bin/

# Install gingx-sdd (Python)
cd ../gingx-sdd && pip install -e .

# Install graphify
uv tool install graphifyy && graphify claude install

# Bootstrap a project
cd ~/my-project && gingx-sdd init
```

## Using the Ecosystem in Another Project

```bash
cd ~/my-new-project
gingx-sdd init                          # scaffolds 23+ files

# Pick a stack (affects which skills are loaded)
gingx-sdd init --stack langgraph        # AI + LangGraph + FastAPI
gingx-sdd init --stack go               # Go + Fiber
gingx-sdd init --stack react            # React + Next.js
gingx-sdd init --stack minimal          # Bare essentials only

# Preview without writing
gingx-sdd init --dry-run

# Overwrite existing config
gingx-sdd init --force
```

### After Init, Your Project Has:

```
my-project/
├── .claude/
│   ├── agents/           # 8 agent personas
│   ├── hooks/            # 3 hooks: SessionStart, PreToolUse, Stop
│   └── settings.local.json
├── .gingx/
│   ├── config.yaml       # 30 harness contracts
│   ├── profiles/         # 7 team profiles
│   ├── knowledge/        # domain-map, component-index (auto-generated)
│   ├── memory/           # entries.jsonl — portable team memory
│   ├── goals/            # autonomous goal definitions
│   └── current_task.yaml # active HDU tracking
├── openspec/
│   ├── AGENTS.md
│   └── changes/
└── .mcp.json             # mnemo + codegraph MCP servers
```

## Auto-Update Matrix

Everything stays current without manual intervention:

| What | Hook | Trigger |
|------|------|---------|
| Knowledge graph (domain-map) | SessionStart | First run (auto-discover) |
| Knowledge timestamps | Stop | Every session end |
| Obsidian vault refresh | Stop | Every session end |
| Graphify AST update (code only) | Stop | Every session end |
| Mnemo knowledge export | Stop | Every session end |
| Mnemo sync push | Stop | Every session end |
| Mnemo import (portable memory) | SessionStart | Every session start |
| Mnemo sync pull | SessionStart | Every session start |

## Adding More Skills

Skills are Markdown files with YAML frontmatter. Auto-discovered by `gingx-sdd team list`.

### Add a Tech Stack Skill

```bash
cat > gingx-sdd/extras/skills/backend/rust.md << 'EOF'
---
name: rust
description: Rust conventions — ownership, borrowing, async, tokio
category: backend
model: sonnet
effort: high
---

# Rust Conventions

## Rules
- Clippy strict. No warnings in CI.
- `cargo test` before every commit.
- Use `anyhow` for application errors, `thiserror` for libraries.

## Patterns
- Repository pattern with sqlx
- Actor model with actix for concurrency
EOF
```

**Categories:** `ai/`, `backend/`, `mobile/`, `web/`, `testing/`, `infra/`

### Add an Agent Persona Skill

```bash
cat > gingx-sdd/skills/team/security-agent.md << 'EOF'
---
name: security-agent
description: Security auditor — OWASP, secrets, dependency audit
model: sonnet
effort: high
tools: Bash, Read, Grep, Glob
trigger: /security-audit
---

# Security Agent

## Your Job
1. Scan for OWASP Top 10 vulnerabilities
2. Audit dependencies for known CVEs
3. Check for hardcoded secrets

## Protocol
- `gitleaks detect --no-git` first
- Report: severity, file, line, remediation
EOF
```

Then register it in a profile:
```bash
gingx-sdd team spawn security-agent -t "audit auth module" --profile developer
```

## Adding More Hooks

Hooks are bash scripts in `.claude/hooks/`. Claude Code runs them automatically.

| Hook | When | Use For |
|------|------|---------|
| `SessionStart` | Session begins | Load context, knowledge graph, blockers |
| `PreToolUse` | Before each tool call | Block tools without approved spec |
| `PostToolUse` | After each tool call | Logging, metrics, auto-save |
| `PostToolBatch` | After tool batch | Post-change validation |
| `Stop` | Session ends | Persist progress, sync mnemo, update graph |
| `PreCompact` | Before context compression | Save decisions before losing context |
| `Notification` | System notifications | Blockers, completed goals |

### Create a New Hook

```bash
cat > .claude/hooks/PostToolUse.sh << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

TOOL_NAME="${CLAUDE_TOOL_NAME:-unknown}"
DURATION_MS="${CLAUDE_TOOL_DURATION_MS:-0}"
PROJECT=$(basename "$(pwd)")

# Log tool usage for cost/performance metrics
if command -v mnemo &>/dev/null; then
    mnemo save "Tool: $TOOL_NAME" \
        "Duration: ${DURATION_MS}ms" \
        --type metric --tags tool-usage,performance \
        2>/dev/null || true
fi

echo '{"continue": true}'
EOF

chmod +x .claude/hooks/PostToolUse.sh
```

### Hook Environment Variables

- `$CLAUDE_TOOL_NAME` — tool name
- `$CLAUDE_TOOL_INPUT` — tool input (JSON)
- `$CLAUDE_PROJECT` — project name
- `$CLAUDE_SESSION_ID` — unique session ID

## Adding More Agents

### 1. Create the Persona File

```bash
cat > .claude/agents/data-engineer-agent.md << 'EOF'
---
name: data-engineer-agent
description: Data pipeline engineer — ETL, schemas, data quality
tools: Bash, Read, Write, Grep, Glob
when_to_use: |
  Use when designing data pipelines, schemas, or ETL flows.
  Triggers on: "pipeline", "ETL", "schema", "data quality".
---

# Data Engineer Agent

## Your Job
1. Design and review database schemas
2. Build ETL pipeline definitions
3. Data quality validation rules
EOF
```

### 2. Add to a Profile

```yaml
# In .gingx/profiles/developer.profile.yaml
agents:
  data-engineer-agent:
    model: sonnet
    tech_stack: [python-core, postgres]
```

### 3. Use It

```bash
gingx-sdd team spawn data-engineer-agent -t "design ETL for user events"
```

## Knowledge Graph

A living knowledge graph of the codebase, maintained by the explorer agent:

```bash
gingx-sdd knowledge status                  # graph + codegraph + mnemo status
gingx-sdd knowledge explore GoalState       # callers, callees, impact
gingx-sdd knowledge search "auth pattern"   # cross-search all layers
gingx-sdd knowledge save-decision "Use Redis" --rationale "..." --trade-off "..."
gingx-sdd knowledge vault                   # generate Obsidian vault with [[wikilinks]]
/graphify . --update                        # incremental graph update (code-only = free)
```

### Knowledge Layers

| File | Contents | Auto-Generated |
|------|----------|----------------|
| `domain-map.yaml` | Domains, key symbols, dependencies, patterns | SessionStart hook |
| `component-index.yaml` | Components with callers/callees, type, role | Explorer agent |
| `decisions-log.yaml` | Architecture decisions with rationale | Manual (`knowledge save-decision`) |
| `vault/` | Obsidian vault with [[wikilinks]] + graph view | Stop hook (auto-refresh) |
| `graphify-out/` | Full graph: graph.json, GRAPH_REPORT.md, graph.html | `/graphify .` |

## Goal System v0.3.0

Autonomous objectives executed by the goal-agent without human interaction:

```bash
gingx-sdd goal create document-auth \
  --objective "Document the entire auth system" \
  --key-results "KR1: OAuth2 flow diagram, KR2: SessionStore docs, KR3: test cases" \
  --max-iterations 30

gingx-sdd goal status document-auth
gingx-sdd goal list

# Launch autonomous loop
gingx-sdd team spawn goal-agent \
  -t "Execute next step for goal document-auth" \
  --profile goal-autonomous

gingx-sdd goal complete document-auth
gingx-sdd goal complete refactor-db --blocked --reason "Waiting for prod migration"
```

**GoalGraph pattern** (LangGraph): `plan → act → observe → reflect` with checkpoints across sessions.

## Team Profiles

8 pre-built profiles for different project types:

| Profile | Agents | Stack | Best For |
|---------|--------|-------|----------|
| `developer` | All 9 agents | python-core, fastapi | General development |
| `fullstack` | 7 agents | react, fastapi, postgres | Full-stack web apps |
| `fullstack-go` | 7 agents | go-fiber, react | Go backend + frontend |
| `fullstack-python-langgraph` | 9 agents | langgraph-python, fastapi | AI + agent apps |
| `goal-autonomous` | goal-agent | goal, langgraph | Overnight/weekend autonomous |
| `react-nextjs` | 7 agents | nextjs, react, tailwind | Frontend-heavy |
| `minimal` | 4 agents | python-core | Quick prototypes |
| `team` | 9 agents | variable | Team conventions |

```bash
gingx-sdd team profile set fullstack-python-langgraph
gingx-sdd team profile show
gingx-sdd team profile list
```

## Execution Modes

```bash
gingx-sdd mode set interactive   # Ask for confirmation on key decisions
gingx-sdd mode set automatic     # Advance phase by phase, fewer interruptions
gingx-sdd mode set dry_run       # Pre-release checks only, no execution
gingx-sdd mode set off           # Harness disabled, free mode
gingx-sdd mode status
```

## Essential Commands

```bash
# Ecosystem info
gingx-sdd status                     # Active HDU, mode, blockers
gingx-sdd knowledge status           # Knowledge graph + codegraph + mnemo
gingx-sdd team list                  # Available agents and skills
gingx-sdd team profile show          # Active profile
gingx-sdd changelog                  # Generate CHANGELOG from HDUs

# Memory
mnemo search "similar pattern" --project $(basename $(pwd)) --limit 5
mnemo save "Decision" "What and why" --type decision --outcome resolved
mnemo stats                          # Store statistics
mnemo import                         # Import from .gingx/memory/
```

## Philosophy

> "A harness transforms raw autonomy into controlled engineering work." — Alan Buscaglia, Gentle AI

30 harness contracts in `.gingx/config.yaml`. Each is an **operational contract**, not a suggestion.

**Principles:**
- **No spec, no code.** The spec gate blocks Write/Edit without an approved spec.
- **Memory travels with the repo.** `.gingx/memory/entries.jsonl` is committed. Clone = team memory.
- **Each agent gets only its context.** Subagent isolation: the dev agent doesn't see the whole repo.
- **Autonomy with traceability.** Every goal leaves history in `.gingx/goals/<id>.yaml`.

---

**[CHANGELOG.md](CHANGELOG.md)** · **[gingx-mnemo](gingx-mnemo/)** · **[gingx-sdd](gingx-sdd/)**
