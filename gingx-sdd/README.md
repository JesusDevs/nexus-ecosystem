# Gingx-SDD — Spec-Driven Development Harness

> *Zero-friction SDD framework. Markdown skills + bash + minimal Python. Agent-agnostic.*

Self-installing harness that gives AI agents a structured workflow: SPEC → PLAN → CODE → TEST → SECURITY → MEMORY.

```bash
gingx-sdd spec "My first feature"
gingx-sdd orchestrate HDU-01 --phase apply
gingx-sdd release v0.1.0 --all
```

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## The Problem

AI agents do "vibe coding" — generating code fast but accumulating technical debt faster.

Gingx-SDD is the **control layer**:
- Spec before code (OpenSpec + Gherkin BDD)
- Strict phases (SDD workflow)
- Cross-session memory (Mnemo vector DB)
- Multi-agent delegation (7 specialized personas)
- Harness engineering (guides + sensors + steering)

---

## Quick Start

```bash
git clone https://github.com/JesusDevs/gingx-sdd.git && cd gingx-sdd
./install.sh
gingx-sdd spec "My feature"
gingx-sdd orchestrate HDU-01 --status
```

---

## Architecture

```
gingx-sdd/                  # SDD Harness (markdown + bash + minimal Python)
├── skills/team/            # 7 multi-agent personas (markdown)
│   ├── supervisor.md       # Orchestrator — phase decomposition + delegation
│   ├── po-agent.md         # Product Owner — specs, BDD, acceptance criteria
│   ├── architect-agent.md  # Solution Architect — design, trade-offs
│   ├── dev-agent.md        # Developer — implementation, tests
│   ├── qa-agent.md         # QA — adversarial testing, root cause
│   ├── ux-agent.md         # UX — accessibility, usability
│   └── devops-agent.md     # DevOps — CI/CD, security, dependencies
├── gingx_sdd/              # Minimal Python CLI (typer + rich only)
│   ├── cli.py              # spec, orchestrate, save, status, release
│   └── orchestrate.py      # Phase → agent delegation
├── templates/              # Auto-installed config templates
│   ├── .gingx/config.yaml  # Project harness config (sensors/guides/steering)
│   └── openspec/           # OpenSpec template for new HDUs
├── extras/skills/          # Optional: web, mobile, backend, AI, testing skills
└── install.sh              # Zero-friction auto-installer
```

### External Services
- **Mnemo** (Go) — Vector memory, MCP server, SQLite, semantic search
- **Claude Code hooks** — Steering loop (PreToolUse, PostToolUse, Stop)

---

## SDD Workflow

```
gingx-sdd spec "Feature X"
  └─→ openspec/changes/HDU-NNN/
       ├── proposal.md    (why + what changes)
       ├── specs/HDU.md   (Gherkin BDD scenarios)
       ├── design.md      (approach + trade-offs)
       └── tasks.md       (implementation checklist)

gingx-sdd orchestrate HDU-NNN --phase apply
  └─→ Decomposes phases → delegates to agent personas
       explore → architect  |  propose → po
       spec    → po         |  design  → architect
       tasks   → dev        |  apply   → dev
       verify  → qa         |  archive → supervisor

gingx-sdd save --hdu-id HDU-NNN
  └─→ mnemo save → vector memory (survives sessions)

gingx-sdd release v0.1.0 --all
  └─→ Pre-release checks → git tag → mnemo snapshot → CHANGELOG
```

---

## Commands

| Command | Description |
|---------|-------------|
| `gingx-sdd spec "Title"` | Create OpenSpec specification |
| `gingx-sdd orchestrate <HDU> --status` | Show phase progress + swarm mode |
| `gingx-sdd orchestrate <HDU> --phase <phase>` | Execute phase with right agent |
| `gingx-sdd orchestrate <HDU> --agent <name>` | Invoke specific persona |
| `gingx-sdd save --hdu-id <HDU>` | Save HDU to mnemo memory |
| `gingx-sdd status` | HDUs + harness health + sensor status |
| `gingx-sdd release v0.1.0 --all` | Semantic release with checks |

---

## Multi-Agent Team

7 personas available as slash commands (`/supervisor`, `/po-agent`, etc.):

| Persona | Role | Phase |
|---------|------|-------|
| supervisor | Orchestrator | All phases |
| po-agent | Product Owner | Spec, BDD, scope |
| architect-agent | Solution Architect | Design, trade-offs |
| dev-agent | Developer | Implementation |
| qa-agent | QA Engineer | Adversarial testing |
| ux-agent | UX Designer | Accessibility, usability |
| devops-agent | DevOps | CI/CD, security |

Personas are portable markdown — work with Claude Code, OpenCode, Kiro, Codex, Antigravity.

---

## Harness Engineering

Following the Thoughtworks harness model:

| Component | Implementation |
|-----------|---------------|
| **Feedforward guides** | Skills (.md), AGENTS.md, .gingx/profiles/ |
| **Feedback sensors** | Security scan, build check, test coverage, mnemo conflicts |
| **Steering loop** | Claude Code hooks, mnemo cross-session memory |

```bash
gingx-sdd status  # Shows sensor/guide health
```

---

## Configuration

```yaml
# .gingx/config.yaml
gingx_version: "0.2.0"
swarm_mode: hybrid  # dag | supervisor | swarm | hybrid

harness:
  feedforward:
    skills_dir: skills/
    agent_profiles_dir: .gingx/profiles/
  feedback:
    sensors:
      security_scan: { enabled: true, tool: gitleaks }
      mnemo_conflicts: { enabled: true, min_similarity: 0.85 }
      dependency_audit: { enabled: true, on_release: true }
  steering:
    hooks_enabled: true
    memory_learning: true
```

Configuration lives in `~/.mnemo/mnemo.db` (via `mnemo config`). Env vars are overrides only.

---

## Optional: Full Skill Catalog

```bash
# Core install (multi-agent team only)
./install.sh

# Full install (all technology skills)
./install.sh --full   # adds extras/skills/{web,mobile,backend,ai,testing}
```

---

## License

MIT — Open Source. Community-owned standard.
