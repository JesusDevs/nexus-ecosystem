# Nexus Ecosystem

**Harness-driven AI development.** Spec → Plan → Code → Test → Security → Memory. Every process is a harness contract. No code without an approved spec.

## Architecture

```
nexus-ecosystem/
├── nexus-mnemo/       # Go — Vector memory (MCP server, JSON-RPC 2.0)
├── nexus-sdd/         # Python — SDD harness (spec → archive)
├── agent/             # (future) Banking agent runtime
├── frontend/          # (future) Banking UI
├── openspec/          # Active HDUs (specs, designs, tasks)
└── .nexus/            # Local config, profiles, skill registry
```

## Components

### nexus-mnemo — Vector Memory System

SQLite + Ollama embeddings + cosine similarity. 8 MCP tools over stdio for universal agent compatibility (Claude Code, Codex, Cursor, Antigravity).

- **9 MCP tools**: search, save, similar, transfer, release, diff, pack, conflicts, stats
- **Zero external APIs** — Ollama local, SQLite, manual cosine similarity
- **Model-agnostic `.mempack`** — text always included, embeddings as optional cache

```bash
mnemo search "UX pattern" --limit 5
mnemo save "Decision" "What and why." --type decision --outcome resolved
mnemo release v0.3.0              # Snapshot with release notes
```

### nexus-sdd — Spec-Driven Development Harness

7 agent personas orchestrated through 8 phases. Each phase produces an artifact AND a memory.

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

| Agent | Trigger | Domain |
|-------|---------|--------|
| `/supervisor` | Orchestrate HDU | Decompose, delegate, track |
| `/po-agent` | Define features | Specs, Gherkin, scope |
| `/ux-agent` | Review UI/UX | Accessibility, usability, design memory |
| `/architect-agent` | Design systems | Trade-offs, API/DB design |
| `/dev-agent` | Implement | TDD, tests, code |
| `/qa-agent` | Verify | Adversarial testing, root cause |
| `/devops-agent` | Ship | CI/CD, security, releases, push |

## Versioning

Versionado mediante **mnemo releases**. Cada release es un snapshot de memoria vectorial + diff desde la release anterior.

```bash
nexus-sdd release v0.3.0          # Crea snapshot, genera changelog, guarda en mnemo
mnemo release v0.3.0              # Release manual con diff y changelog
mnemo releases                     # Historial de releases
```

El versionado **es necesario** porque:
- Cada release es un punto de restauracion en mnemo (backup harness #25)
- Los agentes buscan en releases anteriores para decisiones informadas
- El changelog se genera automaticamente desde mnemo diffs

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
# 1. Instalar mnemo
curl -fsSL https://raw.githubusercontent.com/JesusDevs/nexus-ecosystem/main/nexus-mnemo/install.sh | bash

# 2. Instalar nexus-sdd
curl -fsSL https://raw.githubusercontent.com/JesusDevs/nexus-ecosystem/main/nexus-sdd/install.sh | bash

# 3. Inicializar un proyecto
nexus-sdd spec "Mi feature" --project mi-proyecto

# 4. Orquestar
nexus-sdd orchestrate <HDU-ID>

# 5. Buscar en memoria antes de decidir
mnemo search "decision similar" --project mi-proyecto --limit 5

# 6. Guardar despues de cada avance significativo
nexus-sdd save --hdu-id <HDU-ID>
```

## Philosophy

> "Un harness transforma autonomia cruda en trabajo de ingenieria controlado." — Alan Buscaglia, Gentle AI

30 harnesses configurados en `.nexus/config.yaml`. Cada uno es un **contrato operacional**, no una sugerencia. Cubren orquestacion, fases, calidad, skills, entrega, y extensiones.

El proyecto se gestiona a si mismo — dogfooding nexus-sdd + mnemo + OpenSpec.

## Git Reminders
- `git add -A` is blocked by harness (#23: Permission Security)
- `git push --force` requires explicit confirmation
- Commits should reference HDU-ID when applicable
