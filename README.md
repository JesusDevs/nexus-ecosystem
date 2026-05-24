# Gingx Ecosystem

**Harness-driven AI development.** Spec → Plan → Code → Test → Security → Memory. Every process is a harness contract. No code without an approved spec.

## Architecture

```
gingx-ecosystem/
├── gingx-mnemo/       # Go — Vector memory (MCP server, JSON-RPC 2.0)
├── gingx-sdd/         # Python — SDD harness (spec → archive)
├── agent/             # (future) Banking agent runtime
├── frontend/          # (future) Banking UI
├── openspec/          # Active HDUs (specs, designs, tasks)
└── .gingx/            # Local config, profiles, skill registry
```

> **Language choice is pragmatic, not ideological.** Go handles performance-critical vector operations (mnemo). Python handles workflow orchestration (SDD harness). Each tool in its strength. Eventually the full CLI migrates to Go for zero-dependency distribution.

## Components

### gingx-mnemo — Vector Memory System

SQLite + Ollama BGE-M3 embeddings + cosine similarity. 12 MCP tools over stdio for universal agent compatibility (Claude Code, Codex, Cursor, Antigravity, Kiro).

- **Portable memory** — `.gingx/memory/entries.jsonl` travels with the repo. Clone a project, get its agent memory. `embeddings.json` is gitignored and regenerable.
- **Auto-import on session start** — hooks call `mnemo import` automatically. Zero-config knowledge transfer between teams.
- **Dual-write on save** — every `mnemo save` writes to both the local DB and `.gingx/memory/entries.jsonl`.
- **Zero external APIs** — Ollama local, SQLite, manual cosine similarity
- **Model-agnostic `.mempack`** — text always included, embeddings as optional cache

```bash
mnemo search "UX pattern" --limit 5
mnemo save "Decision" "What and why." --type decision --outcome resolved
mnemo release v0.3.0              # Snapshot with release notes
mnemo import                       # Import from .gingx/memory/entries.jsonl
```

### gingx-sdd — Spec-Driven Development Harness

8-phase SDD pipeline orchestrated through 7 agent personas. Each phase produces an artifact AND a memory.

```bash
gingx-sdd init                        # Bootstrap a new project (23 files scaffolded)
gingx-sdd status                      # Check active HDU and harness state
gingx-sdd hdu create "Feature"        # Start a new HDU
gingx-sdd orchestrate <HDU-ID>        # Delegate to multi-agent team
```

### Scaffolding (`gingx-sdd init`)

Bootstraps a complete Gingx project from templates:

| What | Where |
|------|-------|
| 7 SDD agent personas | `.claude/agents/` |
| 3 enforcement hooks (executable) | `.claude/hooks/` |
| Claude Code settings | `.claude/settings.local.json` |
| 6 agent profiles + 1 team profile | `.gingx/profiles/` |
| Project config + suites | `.gingx/config.yaml`, `.gingx/suites.yaml` |
| Task tracking | `.gingx/current_task.yaml` |
| OpenSpec structure | `openspec/AGENTS.md` + `changes/` |
| MCP servers config | `.mcp.json` |

### HDU Phases

| Phase | Agent | Artifact |
|-------|-------|----------|
| Explore | Architect | findings |
| Propose | PO | proposal.md |
| Spec | PO + QA | specs/ + BDD scenarios |
| Design | Architect | design.md + trade-offs |
| Tasks | Dev | tasks.md |
| Apply | Dev + UX + DevOps | implementation |
| Verify | QA + DevOps | test evidence + security scan |
| Archive | Supervisor | mnemo release snapshot |

### Multi-Agent Team

Scaffolded automatically by `gingx-sdd init`. 7 personas, each a Claude Code agent:

| Agent | Trigger | Domain |
|-------|---------|--------|
| `/supervisor` | Orchestrate HDU | Decompose, delegate, track progress |
| `/po-agent` | Define features | Specs, Gherkin, scope negotiation |
| `/ux-agent` | Review UI/UX | Accessibility, usability, design memory |
| `/architect-agent` | Design systems | Trade-offs, API/DB design, dependencies |
| `/dev-agent` | Implement | TDD, test-first, code conventions |
| `/qa-agent` | Verify | Adversarial testing, BDD, root cause |
| `/devops-agent` | Ship | CI/CD, security scans, releases |

## Versioning

Versionado mediante **mnemo releases** + **SDD harness changelog generation**. Cada release es un snapshot de memoria vectorial generado desde los HDUs completados.

```bash
gingx-sdd changelog                # Generate CHANGELOG from openspec/changes/ (auto)
gingx-sdd release v0.x.0           # Create snapshot, generate changelog, save to mnemo
mnemo release v0.x.0               # Manual release with diff
mnemo releases                     # Release history
```

El versionado **es necesario** porque:
- Cada release es un punto de restauracion en mnemo (backup harness #25)
- Los agentes buscan en releases anteriores para decisiones informadas
- El changelog se genera automaticamente desde los HDUs en `openspec/changes/`

**Estrategia**: Semantic versioning. `0.x.y` mientras el proyecto esta en fase experimental. `MAJOR.MINOR.PATCH` cuando los 4 componentes esten estables.

## Design Memory System

El UX agent no empieza desde cero. El harness carga una base de conocimiento de diseno desde mnemo.

Que se guarda en memoria de diseno:
- **Paletas de color** — combinaciones pre-aprobadas con contraste WCAG AA/AAA
- **Pares tipograficos** — font pairings testeados para web/mobile
- **Patrones de componentes** — estados (loading, empty, error, success, disabled)
- **Heuristicas UX** — Nielsen/Norman + experiencia de apps similares
- **Templates de referencia** — ejemplos de dashboards, landing pages, forms, etc.

```bash
# El UX agent hace esto automaticamente antes de cada review:
mnemo search "design system color palette" --limit 5
mnemo search "UX pattern dashboard" --limit 3
mnemo transfer "design-system-base" $(basename $(pwd))
```

## SDD Gatekeeping

El proyecto no avanza sin respuestas satisfactorias. Antes de cualquier feature:

1. **Spec Gate** — Que problema resuelve? Para quien? Como se mide el exito?
2. **Design Gate** — Que patrones aplican? Que trade-offs hay?
3. **Task Gate** — Cada tarea tiene criterio de aceptacion?

Si la IA no puede responder estas preguntas con contexto, **no se escribe codigo**. Esto fuerza a que cada decision tenga trazabilidad en mnemo.

## Quick Start

```bash
# 1. Install the ecosystem
cd gingx-ecosystem
./gingx-sdd/install.sh          # Python, Go, Ollama, mnemo, gingx-sdd

# 2. Initialize a new project
cd my-project
gingx-sdd init                   # Scaffolds everything: agents, hooks, profiles, OpenSpec

# 3. Create your first HDU
gingx-sdd hdu create "My first feature"

# 4. Orchestrate the multi-agent team
gingx-sdd orchestrate <HDU-ID>

# 5. Search memory before any decision
mnemo search "similar decisions" --limit 5

# 6. Memory auto-saves via hooks. Portable via .gingx/memory/
```

## Philosophy

> "Un harness transforma autonomia cruda en trabajo de ingenieria controlado." — Alan Buscaglia, Gentle AI

30 harnesses configurados en `.gingx/config.yaml`. Cada uno es un **contrato operacional**, no una sugerencia. Cubren orquestacion, fases, calidad, skills, entrega, y extensiones.

El proyecto se gestiona a si mismo — dogfooding gingx-sdd + mnemo + OpenSpec.

## Git Reminders
- `git add -A` is blocked by harness (#23: Permission Security)
- `git push --force` requires explicit confirmation
- Commits should reference HDU-ID when applicable
