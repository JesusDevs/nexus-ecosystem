# Nexus-SDD Agent Instructions

## Your Role
AI coding agent within the Nexus-SDD framework. Follow SDD methodology:

```
SPEC → PLAN → CODE → TEST → SECURITY
```

## Multi-Agent Team

Nexus-SDD provides 7 specialized sub-agent personas in `skills/team/`:

| Persona | Role | When to Use |
|---------|------|-------------|
| `supervisor` | SDD Orchestrator | Decompose HDUs, delegate phases, track progress |
| `po-agent` | Product Owner | Specs, acceptance criteria, scope, Gherkin |
| `ux-agent` | UX Designer | Usability, accessibility, interaction design |
| `architect-agent` | Solution Architect | System design, trade-offs, API/DB design |
| `dev-agent` | Developer | Implementation, tests, bug fixes |
| `qa-agent` | QA Engineer | Adversarial testing, root cause, BDD validation |
| `devops-agent` | DevOps | CI/CD, security scans, dependencies, deploy |

Invoke any persona via Claude Code slash command: `/architect-agent design the auth module`

## Orchestrator Workflow

```bash
# Check progress of an HDU
nexus-sdd orchestrate HDU-06 --status

# See phase decomposition
nexus-sdd orchestrate HDU-06

# Execute a specific phase
nexus-sdd orchestrate HDU-06 --phase apply

# Invoke a specific agent
nexus-sdd orchestrate HDU-06 --agent qa
```

## Memory (mnemo)

Every agent persona reads from and writes to mnemo. Before any significant action:

```bash
mnemo search "<topic>" --project $(basename $(pwd)) --limit 5
mnemo transfer "<topic>" $(basename $(pwd))
```

After decisions, bugs, or progress:

```bash
mnemo save "Title summarizing insight" \
  "What happened, why, what we did, what we learned." \
  --type bugfix|decision|pattern|progress \
  --outcome resolved|applied|noted|in_progress \
  --tags relevant,tags,here
```

## Project Structure

```
nexus-sdd/
├── nexus_sdd/
│   ├── harness/          # LangGraph supervisor + agents
│   │   ├── supervisor.py # Director de Orquesta
│   │   └── agents/       # spec, plan, code, test, security
│   ├── detector/         # Project stack scanner
│   ├── skills/           # Skill registry + generator
│   ├── security/         # Security middleware
│   ├── orchestrate.py    # Team orchestration module (NEW)
│   └── cli.py            # CLI (Typer)
├── skills/               # Technology-specific SKILL.md catalog
│   ├── team/             # Multi-agent personas (NEW)
│   │   ├── supervisor.md
│   │   ├── po-agent.md
│   │   ├── ux-agent.md
│   │   ├── architect-agent.md
│   │   ├── dev-agent.md
│   │   ├── qa-agent.md
│   │   └── devops-agent.md
│   ├── web/              # React, Vue, Next.js, Svelte
│   ├── mobile/           # Kotlin KMP, Flutter, SwiftUI
│   ├── backend/          # FastAPI, Django, Go-Fiber
│   └── testing/          # BDD, Playwright, Vitest
├── templates/            # .nexus/ templates for new projects
├── install.sh            # Universal installation script
└── docs/                 # Architecture + quickstart
```

## Core Rules

1. **SPEC first.** Every change starts with OpenSpec. Use `/opsx:propose`.
2. **Mnemo memory.** After every significant decision or bug: save to mnemo.
3. **Security.** No hardcoded secrets. `nexus-sdd security` scans on every commit.
4. **Skills are protocol.** Every skill file is a behavior contract for agents.
5. **Test-first.** BDD scenarios before code. QA agent verifies before archive.

## When Adding a New Team Agent

1. Create `skills/team/<role>-agent.md` with frontmatter + role definition
2. Register in `nexus_sdd/orchestrate.py` → `PHASE_AGENT_MAP` and `AGENT_ROLES`
3. Add to `cli.py` agent_map
4. Test: `nexus-sdd orchestrate <HDU> --agent <role>`

## Technology Stack (for this project)

- **Python 3.11+** — LangGraph harness, CLI (Typer), detector
- **Go** — nexus-mnemo (vector memory, MCP server)
- **Markdown** — Skill definitions, OpenSpec specs, agent personas
- **SQLite** — mnemo.db for vector memory + config
- **Ollama** — Local embeddings (bge-m3, 1024-dim)
