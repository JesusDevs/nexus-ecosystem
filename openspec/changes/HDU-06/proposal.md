# Proposal: God-Tier Multi-Agent Ecosystem

## Why

gentle-ai/ATL solved orchestration but left memory as an afterthought (FTS5 keyword search). nexus-mnemo has semantic memory, versioning, and transfer — but no orchestration layer. The market has NO system that combines real vector memory with autonomous multi-agent orchestration, self-evolving skills, and predictive knowledge.

This HDU is the convergence: a self-sustaining ecosystem where agents not only follow SDD — they improve the system itself, learn from every session, share knowledge across projects, and pre-load context before you even ask.

## What Changes

### 1. Multi-Agent Swarm Orchestrator (`nexus-sdd orchestrate`)
Beyond delegate-only. Agents self-organize via a shared task queue with claim/heartbeat. The orchestrator spawns phase agents, monitors progress, and re-assigns failed tasks automatically. New agents can join mid-execution.

### 2. Self-Evolving Skill Registry
Skills are not static markdown files. Every successful code pattern, bug fix, and decision becomes a skill candidate. The registry auto-generates skills from mnemo memories with high outcome="resolved" and high similarity clusters. Skills decay if unused, strengthen if reused.

### 3. Granular Progress with Session Replay
Every agent action is recorded as a memory with `type="progress"`. If a session dies, the next agent replays the progress chain and resumes exactly where it left off — including partial file edits, test results, and reasoning context. Nothing is lost.

### 4. Evolutionary Judgment Day
Beyond 2-review blind comparison. The verifier generates adversarial test cases, mutates inputs, fuzzes edge cases, and runs counterfactual scenarios. Code that survives is certified. Code that fails triggers auto-fix loops with causal analysis.

### 5. Adaptive Agent Persona
Agents learn your coding style, preferences, and patterns across sessions. Persona is stored as vector memories — the agent retrieves "how would the user solve this?" before proposing solutions. Cross-tool identity: same persona whether you're in Claude Code, Antigravity, or Codex.

### 6. Self-Healing Infrastructure
The system monitors its own health. Detects stale memories, contradictory decisions, unused skills. Auto-prunes, auto-consolidates, auto-upgrades. Canary deployments for new system versions with automatic rollback if quality degrades.

### 7. Predictive Knowledge Pre-Loading
Before you type a command, the system predicts what you're about to work on based on git state, recent activity, and open issues. Pre-loads relevant memories, past bugs, design decisions, and similar implementations from other projects.

### 8. Autonomous Improvement Loops
The system detects patterns in its own failures. If the same type of bug recurs across projects, it proposes a system-wide fix. If a skill consistently fails, it self-deprecates. The ecosystem improves itself.

## What Does NOT Change
- Mnemo vector memory (underlying storage)
- `.mempack` format (still the universal exchange)
- MCP protocol (JSON-RPC 2.0, agents communicate via MCP)
- Zero external APIs (all local: Ollama, SQLite, fsnotify)
- OpenSpec file structure (still the canonical spec format)

## Impact
- HDU: HDU-06
- Complexity: epic (largest HDU, 20+ files, 4 new packages)
- New Go packages: `swarm/`, `registry/`, `progress/`, `judgment/`, `persona/`, `predictor/`
- New CLI commands: 6 (orchestrate, persona, predict, evolve, heal, replay)
- New MCP tools: 6 (mem_save_progress, mem_replay_session, mem_evolve_skill, mem_predict_context, mem_persona_search, mem_health_check)
- Total MCP tools: 8 → 14
