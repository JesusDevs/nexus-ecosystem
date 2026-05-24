# AGENTS.md — Gingx Ecosystem

Monorepo: gingx-mnemo (Go vector memory + MCP) + gingx-sdd (SDD harness) + multi-agent team.

## Your Role
AI coding agent. Follow SDD: **SPEC → PLAN → CODE → TEST → SECURITY → MEMORY**. Never code without an approved spec.

## Project Layout
```
gingx-ecosystem/
├── gingx-mnemo/       # Go — Vector memory (MCP server, 14 tools planned)
│   ├── vec/           # Vector store + embeddings (bge-m3, 1024-dim)
│   ├── mcp/           # MCP server (JSON-RPC 2.0 over stdio)
│   ├── swarm/         # (planned) Multi-agent orchestration
│   ├── main.go        # CLI (search, save, config, setup, mcp)
│   └── install.sh     # Zero-friction installer
├── gingx-sdd/         # SDD harness — markdown skills + bash + templates
│   ├── skills/team/   # 7 agent personas (supervisor, PO, UX, Architect, Dev, QA, DevOps)
│   ├── templates/     # .gingx/ templates
│   ├── install.sh     # Universal auto-installer
│   └── AGENTS.md      # SDD agent instructions
├── openspec/          # OpenSpec changes (active HDUs)
├── .gingx/            # Local config + profiles + installed skills
├── .claude/           # Claude Code hooks + settings
├── agent/             # (future) Banking agent runtime
└── frontend/          # (future) Banking UI
```

## Multi-Agent Team

7 specialized personas available as Skills AND as spawnable sub-agents via Agent Factory.

**Via Skills** (shape current agent behavior):
| Command | Role |
|---------|------|
| `/supervisor` | SDD Orchestrator — decompose, delegate, track |
| `/po-agent` | Product Owner — specs, Gherkin, scope |
| `/ux-agent` | UX Designer — usability, accessibility, design memory |
| `/architect-agent` | Solution Architect — design, trade-offs |
| `/dev-agent` | Developer — test-first implementation |
| `/qa-agent` | QA Engineer — adversarial testing, root cause |
| `/devops-agent` | DevOps — CI/CD, security, git push harness |

**Via Agent Factory** (spawn as parallel sub-agents):
```bash
gingx-sdd team spawn dev-agent -t "Implement auth" --profile fullstack-python-langgraph
gingx-sdd team spawn ux-agent -t "Review onboarding flow" --profile react-nextjs
gingx-sdd team list                    # List agents, tech stacks, profiles
gingx-sdd team profile set minimal     # Switch team profile
```

**Agent Dispatch via MCP:**
```json
{"tool": "agent_dispatch", "args": {"agent": "architect-agent", "task": "Design API", "profile": "fullstack-go"}}
```

**Profiles** (`.gingx/profiles/`):
| Profile | Stack | For |
|---------|-------|-----|
| `fullstack-python-langgraph` | Python, FastAPI, LangGraph | AI agent development |
| `fullstack-go` | Go, go-fiber, SQLite | High-perf backend + MCP |
| `react-nextjs` | React, Next.js, Tailwind | Frontend dashboards |
| `fullstack` | Python + React | Balanced web apps |
| `minimal` | Minimal | Quick fixes |
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
1. **Spec**: `gingx-sdd spec "Feature"` → creates `openspec/changes/<HDU>/`
2. **Orchestrate**: `gingx-sdd orchestrate <HDU>` → phase decomposition
3. **Code**: Implement tasks. Each task = one commit.
4. **Test**: QA agent verifies. BDD scenarios required.
5. **Save**: `gingx-sdd save --hdu-id <HDU>` → mnemo memory

## Configuration
Configuration lives in `~/.mnemo/mnemo.db` (table `vec_config`):
```bash
mnemo config              # Show all config
mnemo config set k v      # Update config
```
Env vars (OLLAMA_HOST, EMBEDDER_MOCK) act as overrides only.

## 30 Harnesses — Gingx ↔ Gentle AI

"Un harness transforma autonomía cruda en trabajo de ingeniería controlado." — Alan Buscaglia

Cada harness es un **contrato operacional**, no una sugerencia.
Configuración completa en `.gingx/config.yaml`.

### Bloque 1: Orquestación y Contexto (4)
| # | Harness | Gingx Implementación |
|---|---------|-------------------|
| 1 | **SDD Orchestrator** | `supervisor` + `orchestrate.py` — coordina, no ejecuta |
| 2 | **Delegation** | inline (≤3 files) / delegate / full SDD |
| 3 | **SDD Init** | `gingx-sdd spec` — detecta stack, crea artifacts |
| 4 | **Execution Mode** | `swarm.mode` (dag\|supervisor\|swarm\|hybrid) |

### Bloque 2: Fases y Artifactos (5)
| # | Harness | Gingx Implementación |
|---|---------|-------------------|
| 5 | **Phase DAG** | PHASES order — no se salta etapas |
| 6 | **Artifact Dependency** | spec→design→tasks→apply→verify |
| 7 | **Result Contract** | `OrchestrationStatus` envelope |
| 8 | **Artifact Grammar** | OpenSpec (proposal + specs + design + tasks) |
| 9 | **Artifact Store** | híbrido: `openspec/changes/` + mnemo |

### Bloque 3: Calidad y Continuidad (3)
| # | Harness | Gingx Implementación |
|---|---------|-------------------|
| 10 | **Strict TDD** | `dev-agent` — red, green, triangulate, refactor |
| 11 | **Verify** | `qa-agent` — "terminé ≠ verificado" |
| 12 | **Apply Continuity** | mnemo progress tracking entre sesiones |

### Bloque 4: Skills y Subagentes (4)
| # | Harness | Gingx Implementación |
|---|---------|-------------------|
| 13 | **Skill Registry** | `skills/team/` + `extras/skills/` |
| 14 | **Skill Digestion** | `build_prompt()` compacta reglas |
| 15 | **Skill Resolution** | `track_progress()` audita qué se aplicó |
| 16 | **Subagent Isolation** | cada agente recibe solo su contexto |

### Bloque 5: Entrega (3)
| # | Harness | Gingx Implementación |
|---|---------|-------------------|
| 17 | **Review Workload** | max 400 líneas/PR, 3 áreas |
| 18 | **Delivery Strategy** | stacked PRs / feature track / ask on risk |
| 19 | **Chain Strategy** | geometría de entrega (stacked\|feature_branch\|main) |

### Bloque 6: Extendidos (11)
| # | Harness | Gingx Implementación |
|---|---------|-------------------|
| 20 | **Engram Memory** | **Mnemo** — vector memory + semantic search |
| 21 | **Model Routing** | `/gentle models` — distintos modelos por fase |
| 22 | **Profile Isolation** | `.gingx/profiles/` — un perfil por developer |
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
