# AGENTS.md — Nexus Ecosystem

Monorepo: nexus-mnemo (Go vector memory + MCP) + nexus-sdd (SDD harness) + multi-agent team.

## Your Role
AI coding agent. Follow SDD: **SPEC → PLAN → CODE → TEST → SECURITY → MEMORY**. Never code without an approved spec.

## Project Layout
```
nexus-ecosystem/
├── nexus-mnemo/       # Go — Vector memory (MCP server, 14 tools planned)
│   ├── vec/           # Vector store + embeddings (bge-m3, 1024-dim)
│   ├── mcp/           # MCP server (JSON-RPC 2.0 over stdio)
│   ├── swarm/         # (planned) Multi-agent orchestration
│   ├── main.go        # CLI (search, save, config, setup, mcp)
│   └── install.sh     # Zero-friction installer
├── nexus-sdd/         # SDD harness — markdown skills + bash + templates
│   ├── skills/team/   # 7 agent personas (supervisor, PO, UX, Architect, Dev, QA, DevOps)
│   ├── templates/     # .nexus/ templates
│   ├── install.sh     # Universal auto-installer
│   └── AGENTS.md      # SDD agent instructions
├── openspec/          # OpenSpec changes (active HDUs)
├── .nexus/            # Local config + profiles + installed skills
├── .claude/           # Claude Code hooks + settings
├── agent/             # (future) Banking agent runtime
└── frontend/          # (future) Banking UI
```

## Multi-Agent Team
7 specialized personas available as slash commands:

| Command | Role | Use for |
|---------|------|---------|
| `/supervisor` | SDD Orchestrator | Decompose HDUs, delegate phases |
| `/po-agent` | Product Owner | Specs, acceptance criteria, Gherkin |
| `/ux-agent` | UX Designer | Usability, accessibility, design |
| `/architect-agent` | Solution Architect | System design, trade-offs, API/DB |
| `/dev-agent` | Developer | Implementation, tests, bug fixes |
| `/qa-agent` | QA Engineer | Adversarial testing, root cause |
| `/devops-agent` | DevOps | CI/CD, security scans, deps |

## Before Any Decision
```bash
mnemo search "<query>" --project $(basename $(pwd)) --limit 5
mnemo transfer "<context>" $(basename $(pwd))
```

## After Significant Work
```bash
mnemo save "Title" "What happened, why, what we did." \
  --type bugfix|decision|pattern|progress \
  --outcome resolved|applied|noted|in_progress
```

## SDD Workflow
1. **Spec**: `nexus-sdd spec "Feature"` → creates `openspec/changes/<HDU>/`
2. **Orchestrate**: `nexus-sdd orchestrate <HDU>` → phase decomposition
3. **Code**: Implement tasks. Each task = one commit.
4. **Test**: QA agent verifies. BDD scenarios required.
5. **Save**: `nexus-sdd save --hdu-id <HDU>` → mnemo memory

## Configuration
Configuration lives in `~/.mnemo/mnemo.db` (table `vec_config`):
```bash
mnemo config              # Show all config
mnemo config set k v      # Update config
```
Env vars (OLLAMA_HOST, EMBEDDER_MOCK) act as overrides only.

## 30 Harnesses — Nexus ↔ Gentle AI

"Un harness transforma autonomía cruda en trabajo de ingeniería controlado." — Alan Buscaglia

Cada harness es un **contrato operacional**, no una sugerencia.
Configuración completa en `.nexus/config.yaml`.

### Bloque 1: Orquestación y Contexto (4)
| # | Harness | Nexus Implementación |
|---|---------|-------------------|
| 1 | **SDD Orchestrator** | `supervisor` + `orchestrate.py` — coordina, no ejecuta |
| 2 | **Delegation** | inline (≤3 files) / delegate / full SDD |
| 3 | **SDD Init** | `nexus-sdd spec` — detecta stack, crea artifacts |
| 4 | **Execution Mode** | `swarm.mode` (dag\|supervisor\|swarm\|hybrid) |

### Bloque 2: Fases y Artifactos (5)
| # | Harness | Nexus Implementación |
|---|---------|-------------------|
| 5 | **Phase DAG** | PHASES order — no se salta etapas |
| 6 | **Artifact Dependency** | spec→design→tasks→apply→verify |
| 7 | **Result Contract** | `OrchestrationStatus` envelope |
| 8 | **Artifact Grammar** | OpenSpec (proposal + specs + design + tasks) |
| 9 | **Artifact Store** | híbrido: `openspec/changes/` + mnemo |

### Bloque 3: Calidad y Continuidad (3)
| # | Harness | Nexus Implementación |
|---|---------|-------------------|
| 10 | **Strict TDD** | `dev-agent` — red, green, triangulate, refactor |
| 11 | **Verify** | `qa-agent` — "terminé ≠ verificado" |
| 12 | **Apply Continuity** | mnemo progress tracking entre sesiones |

### Bloque 4: Skills y Subagentes (4)
| # | Harness | Nexus Implementación |
|---|---------|-------------------|
| 13 | **Skill Registry** | `skills/team/` + `extras/skills/` |
| 14 | **Skill Digestion** | `build_prompt()` compacta reglas |
| 15 | **Skill Resolution** | `track_progress()` audita qué se aplicó |
| 16 | **Subagent Isolation** | cada agente recibe solo su contexto |

### Bloque 5: Entrega (3)
| # | Harness | Nexus Implementación |
|---|---------|-------------------|
| 17 | **Review Workload** | max 400 líneas/PR, 3 áreas |
| 18 | **Delivery Strategy** | stacked PRs / feature track / ask on risk |
| 19 | **Chain Strategy** | geometría de entrega (stacked\|feature_branch\|main) |

### Bloque 6: Extendidos (11)
| # | Harness | Nexus Implementación |
|---|---------|-------------------|
| 20 | **Engram Memory** | **Mnemo** — vector memory + semantic search |
| 21 | **Model Routing** | `/gentle models` — distintos modelos por fase |
| 22 | **Profile Isolation** | `.nexus/profiles/` — un perfil por developer |
| 23 | **Permission Security** | bloquea comandos destructivos sin confirmación |
| 24 | **MCP Injection** | mnemo MCP server (9 tools activos) |
| 25 | **Backup** | `mnemo release` snapshots |
| 26 | **Rollback** | `git revert` auto en test/security failure |
| 27 | **Component Dependency** | DAG engine (HDU-06) |
| 28 | **Command Wrapper** | Claude Code hooks (5 eventos) |
| 29 | **Per-Agent Adapter** | markdown portable (Claude, OpenCode, Codex, Kiro) |
| 30 | **Session Summary** | Stop hook → mnemo save

## Security
- Never save secrets, tokens, or keys to memory
- Security scan before every release
- If you detect sensitive info, warn the user immediately
