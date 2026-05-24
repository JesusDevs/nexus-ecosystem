# Gingx Ecosystem v0.3.0

**SDD harness + vector memory + autonomous agents + knowledge graph.** Convierte cualquier proyecto en un equipo multi-agente con memoria persistente, spec-driven development, y ejecución autónoma de objetivos.

## En 30 segundos

```bash
# 1. Inicializar cualquier proyecto con el ecosistema
cd mi-proyecto
gingx-sdd init

# 2. El ecosistema ahora tiene 9 agentes, 3 hooks, 19 skills, 8 perfiles
#     - memoria vectorial (mnemo) para búsqueda semántica entre sesiones
#     - grafo de conocimiento (graphify) del codebase completo
#     - spec-gate que bloquea código sin spec aprobada
#     - goal-agent que ejecuta objetivos autónomamente (ej. overnight)

# 3. Crear un HDU y delegar
gingx-sdd hdu create "Mi feature" --question "¿Cuál es el stack?"
gingx-sdd auto "implementar endpoint de health check"

# 4. Lanzar un objetivo autónomo
gingx-sdd goal create aprender-codebase \
  --objective "Mapear y documentar toda la arquitectura" \
  --key-results "KR1: domain-map completo, KR2: component-index, KR3: decisions log"

# 5. Todo se guarda en .gingx/memory/ — viaja con el repo
git add .gingx/memory/entries.jsonl && git commit -m "memoria del equipo"
```

## Qué incluye

| Capa | Componente | Lenguaje | Qué hace |
|------|-----------|----------|----------|
| **Memoria** | [gingx-mnemo](gingx-mnemo/) | Go | Vector memory con SQLite + Ollama embeddings. 12 MCP tools. Portable via `.gingx/memory/entries.jsonl` |
| **SDD Harness** | [gingx-sdd](gingx-sdd/) | Python | 8 fases (explore→archive), 9 agentes, 30 harness contracts, goal system |
| **Conocimiento** | Knowledge Graph | YAML + Obsidian + graphify | domain-map, component-index, decisions-log. Auto-generado y versionado |
| **Agentes** | [.claude/agents/](.claude/agents/) | Markdown | 9 personas con triggers, reglas, y herramientas |
| **Hooks** | [.claude/hooks/](.claude/hooks/) | Bash | SessionStart (carga contexto), PreToolUse (spec gate), Stop (persiste progreso) |
| **Skills** | [skills/](gingx-sdd/skills/) + [extras/](gingx-sdd/extras/skills/) | Markdown | 9 team skills + 19 tech stack skills en 6 categorías |
| **Perfiles** | [.gingx/profiles/](.gingx/profiles/) | YAML | 8 composiciones de equipo pre-armadas |

## Los 9 agentes

| # | Agente | Disparador | Qué hace |
|---|--------|-----------|----------|
| 1 | `supervisor` | Orquestar HDUs | Descompone, delega, trackea progreso entre fases |
| 2 | `explorer-agent` | SessionStart, preguntas "dónde/cómo" | Mapea el codebase, mantiene el knowledge graph |
| 3 | `po-agent` | Definir features | Specs, Gherkin, scope negotiation |
| 4 | `architect-agent` | Diseñar sistemas | Trade-offs, DB schema, APIs, dependencias |
| 5 | `dev-agent` | Implementar | TDD, test-first, convenciones de código |
| 6 | `qa-agent` | Verificar | Adversarial testing, BDD, root cause |
| 7 | `ux-agent` | Revisar UI/UX | Accesibilidad WCAG, usabilidad, diseño |
| 8 | `devops-agent` | Shipping | CI/CD, security scans, releases |
| 9 | `goal-agent` | Objetivos autónomos | Ejecuta goals sin interacción humana (overnight/weekend) |

## Las 3 fases de operación

```
SessionStart                  Sesión activa                   Stop
───────────                   ─────────────                   ────
carga contexto        →   9 agentes trabajan     →    persiste progreso
knowledge graph             spec gate activo           knowledge vault refresh
mnemo sync pull             memoria vectorial          mnemo sync push
HDUs activos                goal loops                 HDU timestamps
blockers                    auto-delegation            graphify update
```

## Cómo usar en otro proyecto

```bash
# Opción A — Instalación completa (recomendada para equipos)
cd gingx-ecosystem
./gingx-sdd/install.sh                    # Go, Python, Ollama, mnemo, graphify

# Opción B — Solo el CLI (desarrollo individual)
pip install -e gingx-sdd/
```

```bash
# Inicializar un proyecto nuevo
cd ~/mi-nuevo-proyecto
gingx-sdd init                             # scaffolds 23+ archivos

# Elegir stack (afecta qué skills se cargan)
gingx-sdd init --stack langgraph           # AI + LangGraph + FastAPI
gingx-sdd init --stack go                  # Go + fiber
gingx-sdd init --stack react               # React + Next.js
gingx-sdd init --stack minimal             # Solo lo esencial

# Ver qué se crea sin escribir
gingx-sdd init --dry-run

# Sobrescribir config existente
gingx-sdd init --force
```

### Después del init, tu proyecto tiene:

```
mi-proyecto/
├── .claude/
│   ├── agents/           # 8 agentes (todos menos goal-agent por defecto)
│   ├── hooks/            # 3 hooks: SessionStart, PreToolUse, Stop
│   └── settings.local.json
├── .gingx/
│   ├── config.yaml       # 30 harness contracts
│   ├── profiles/         # 7 perfiles de equipo
│   ├── knowledge/        # domain-map, component-index (auto-generados)
│   ├── memory/           # entries.jsonl — memoria portable del equipo
│   ├── goals/            # objetivos autónomos (goal-agent)
│   └── current_task.yaml # tracking de HDU activo
├── openspec/
│   ├── AGENTS.md
│   └── changes/
└── .mcp.json             # mnemo + codegraph MCP servers
```

## Cómo agregar más skills

Las skills son archivos Markdown con frontmatter. Se auto-cargan por categoría.

### Agregar un tech stack skill

```bash
# Crear el archivo en la categoría correcta
cat > gingx-sdd/extras/skills/backend/rust.md << 'EOF'
---
name: rust
description: Rust language conventions — ownership, borrowing, async, tokio
category: backend
model: sonnet
effort: high
---

# Rust Conventions

## Rules
- Clippy strict. No warnings allowed in CI.
- `cargo test` before every commit.
- Use `anyhow` for application errors, `thiserror` for libraries.
- Async: tokio + axum for web, tracing for observability.

## Patterns
- Repository pattern with sqlx
- Actor model with actix for concurrency
- Builder pattern for complex configs
EOF
```

**Categorías disponibles:** `ai/`, `backend/`, `mobile/`, `web/`, `testing/`, `infra/`

### Agregar un agent persona skill

```bash
cat > gingx-sdd/skills/team/security-agent.md << 'EOF'
---
name: security-agent
description: Security auditor — OWASP, secret scanning, dependency audit
model: sonnet
effort: high
tools: Bash, Read, Grep, Glob
trigger: /security-audit
---

# Security Agent

## Your Job
1. Scan for OWASP Top 10 vulnerabilities
2. Audit dependencies with known CVEs
3. Check for hardcoded secrets
4. Review auth flows for common weaknesses

## Protocol
- `gitleaks detect --no-git` first
- `safety check` for Python, `cargo audit` for Rust, `npm audit` for JS
- Report: severity, file, line, remediation
EOF
```

Luego se registra en un perfil o se invoca directamente:
```bash
gingx-sdd team spawn security-agent -t "audit auth module" --profile developer
```

### Cómo funcionan las skills

- **auto-discovery**: `gingx-sdd team list` escanea `skills/team/` y `extras/skills/`
- **frontmatter YAML**: name, description, model, effort
- **carga por perfil**: cada perfil YAML asigna skills a agentes
- **digestión**: el sistema compacta skills a ≤10 reglas por agente antes de enviar

## Cómo agregar más hooks

Los hooks son scripts bash en `.claude/hooks/`. Claude Code los ejecuta automáticamente.

### Tipos de hooks disponibles

| Hook | Cuándo se ejecuta | Para qué sirve |
|------|------------------|----------------|
| `SessionStart` | Al iniciar sesión | Cargar contexto, knowledge graph, blockers |
| `PreToolUse` | Antes de cada tool call | Bloquear herramientas sin spec aprobada |
| `PostToolUse` | Después de cada tool call | Logging, métricas, auto-save |
| `PostToolBatch` | Después de lote de tools | Validación post-cambios |
| `Stop` | Al terminar sesión | Persistir progreso, sync mnemo, graphify |
| `PreCompact` | Antes de comprimir contexto | Guardar decisiones antes de perder contexto |
| `Notification` | En notificaciones del sistema | Alertas de blockers o goals completados |

### Crear un hook nuevo

```bash
cat > .claude/hooks/PostToolUse.sh << 'EOF'
#!/usr/bin/env bash
# Track tool usage for cost/performance metrics
set -euo pipefail

TOOL_NAME="${CLAUDE_TOOL_NAME:-unknown}"
DURATION_MS="${CLAUDE_TOOL_DURATION_MS:-0}"
PROJECT=$(basename "$(pwd)")

# Log to mnemo for analytics
if command -v mnemo &>/dev/null; then
    mnemo save "Tool: $TOOL_NAME" \
        "Duration: ${DURATION_MS}ms, Project: $PROJECT" \
        --type metric --outcome in_progress \
        --tags tool-usage,performance \
        2>/dev/null || true
fi

echo '{"continue": true}'
exit 0
EOF

chmod +x .claude/hooks/PostToolUse.sh
```

### Variables disponibles en hooks

- `$CLAUDE_TOOL_NAME` — nombre de la herramienta
- `$CLAUDE_TOOL_INPUT` — input de la herramienta (JSON)
- `$CLAUDE_PROJECT` — nombre del proyecto
- `$CLAUDE_SESSION_ID` — ID único de sesión

## Cómo agregar más agentes

### 1. Crear el persona file

```bash
cat > .claude/agents/data-engineer-agent.md << 'EOF'
---
name: data-engineer-agent
description: Data pipeline engineer — ETL, schemas, data quality
tools: Bash, Read, Write, Grep, Glob
when_to_use: |
  Use when designing data pipelines, schemas, or ETL flows.
  Triggers on keywords: "pipeline", "ETL", "schema", "data quality".
---

# Data Engineer Agent

## Your Job
1. Design and review database schemas
2. Build ETL pipeline definitions
3. Data quality validation rules
4. Query optimization

## Before Acting
- Check mnemo for prior schema decisions
- Review existing pipelines in the codebase
- Validate against data contracts
EOF
```

### 2. Agregarlo a un perfil

```yaml
# En .gingx/profiles/developer.profile.yaml
agents:
  data-engineer-agent:
    model: sonnet
    tech_stack: [python-core, postgres]
```

### 3. Agregarlo a la fase SDD (opcional)

```yaml
# En .gingx/config.yaml, harness.orchestrator.phase_agent_map:
data_pipeline: data-engineer-agent
```

### 4. Usarlo

```bash
gingx-sdd team spawn data-engineer-agent -t "design ETL for user events" --profile developer
```

## Knowledge Graph

El ecosistema mantiene un grafo de conocimiento vivo del codebase:

```bash
# Ver estado del knowledge graph
gingx-sdd knowledge status

# Explorar un símbolo (callers, callees, impacto)
gingx-sdd knowledge explore GoalState

# Buscar en mnemo + codegraph + knowledge files
gingx-sdd knowledge search "auth pattern"

# Guardar una decisión de arquitectura
gingx-sdd knowledge save-decision "Usar Redis para caching" \
  --rationale "Menos de 1ms de latencia, equipo ya lo conoce" \
  --trade-off "Costo de infraestructura adicional"

# Generar vault de Obsidian (graph view nativo)
gingx-sdd knowledge vault

# Grafo completo con graphify (2467 nodos en este repo)
/graphify . --update   # incremental, solo archivos cambiados
```

### Capas del knowledge graph

| Archivo | Qué contiene | Auto-generado |
|---------|-------------|---------------|
| `domain-map.yaml` | Dominios, símbolos clave, dependencias, patrones | SessionStart hook |
| `component-index.yaml` | Componentes con callers/callees, tipo, rol | Explorer agent |
| `decisions-log.yaml` | ADR: decisiones, racional, trade-offs | Manual via `knowledge save-decision` |
| `vault/` | Obsidian vault con [[wikilinks]] y graph view | Stop hook (auto-refresh) |
| `graphify-out/` | Grafo completo: graph.json, GRAPH_REPORT.md, graph.html | `/graphify .` |

## Goal System v0.3.0

Objetivos autónomos que ejecuta el goal-agent sin interacción humana:

```bash
# Crear un goal
gingx-sdd goal create documentar-auth \
  --objective "Documentar todo el sistema de autenticación" \
  --key-results "KR1: diagrama de flujo OAuth2, KR2: documentar SessionStore, KR3: test cases documentados" \
  --max-iterations 30

# Ver progreso
gingx-sdd goal status documentar-auth
gingx-sdd goal list

# Lanzar loop autónomo (usar con /loop en Claude Code)
gingx-sdd team spawn goal-agent \
  -t "Execute next step for goal documentar-auth" \
  --profile goal-autonomous

# Completar o bloquear
gingx-sdd goal complete documentar-auth
gingx-sdd goal complete refactor-db --blocked --reason "Esperando migración de prod"
```

**GoalGraph pattern** (LangGraph): `plan → act → observe → reflect` con checkpoints entre sesiones.

## Perfiles de equipo

8 perfiles pre-armados para diferentes tipos de proyecto:

| Perfil | Agentes | Stack | Ideal para |
|--------|---------|-------|------------|
| `developer` | Los 9 agentes | python-core, fastapi | Desarrollo general |
| `fullstack` | 7 agentes | react, fastapi, postgres | Apps web full-stack |
| `fullstack-go` | 7 agentes | go-fiber, react | Backend Go + frontend |
| `fullstack-python-langgraph` | 9 agentes | langgraph-python, fastapi, ai | Apps AI + agentes |
| `goal-autonomous` | goal-agent | goal, langgraph | Overnight/weekend autónomo |
| `react-nextjs` | 7 agentes | nextjs, react, tailwind | Frontend pesado |
| `minimal` | 4 agentes (dev, qa, architect, supervisor) | python-core | Prototipos rápidos |
| `team` | 9 agentes | variable | Convenciones de equipo |

```bash
# Cambiar perfil activo
gingx-sdd team profile set fullstack-python-langgraph
gingx-sdd team profile show
gingx-sdd team profile list
```

## Modos de ejecución

```bash
# Interactivo — pide confirmación en decisiones importantes
gingx-sdd mode set interactive

# Automático — avanza fase por fase con menos interrupciones
gingx-sdd mode set automatic

# Dry run — solo pre-release checks, sin ejecución real
gingx-sdd mode set dry_run

# Off — harness desactivado, modo libre
gingx-sdd mode set off

gingx-sdd mode status
```

## Comandos esenciales

```bash
# Info del ecosistema
gingx-sdd status                     # HDU activo, modo, blockers
gingx-sdd knowledge status           # Knowledge graph + codegraph + mnemo
gingx-sdd team list                  # Agentes y skills disponibles
gingx-sdd team profile show          # Perfil activo
gingx-sdd changelog                  # Generar CHANGELOG desde HDUs

# Memoria
mnemo search "patrón similar" --project $(basename $(pwd)) --limit 5
mnemo save "Decisión" "Qué y por qué" --type decision --outcome resolved
mnemo stats                          # Estadísticas del store
mnemo import                         # Importar desde .gingx/memory/
```

## Filosofía

> "Un harness transforma autonomía cruda en trabajo de ingeniería controlado." — Alan Buscaglia, Gentle AI

30 harness contracts en `.gingx/config.yaml`. Cada uno es un **contrato operacional**, no una sugerencia.

**Principios:**
- **No hay código sin spec.** El spec gate bloquea Write/Edit si no hay spec aprobada.
- **La memoria viaja con el repo.** `.gingx/memory/entries.jsonl` se commitea. Clone = memoria del equipo.
- **Cada agente recibe solo su contexto.** Subagent isolation: el dev no ve todo el repo, solo lo que necesita.
- **Autonomía con trazabilidad.** Cada goal deja historia en `.gingx/goals/<id>.yaml`.

---

**[CHANGELOG.md](CHANGELOG.md)** · **[gingx-mnemo](gingx-mnemo/)** · **[gingx-sdd](gingx-sdd/)**
