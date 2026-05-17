# Tasks: God-Tier Multi-Agent Ecosystem

## Phase 1 — Swarm Orchestrator
- [ ] 1. `swarm/task.go` — Task type, result envelope, status machine
- [ ] 2. `swarm/orchestrator.go` — SQLite-backed task queue with claim/heartbeat
- [ ] 3. `swarm/agent.go` — Agent lifecycle: spawn, heartbeat, death detection, reassign
- [ ] 4. `swarm/dag.go` — DAG dependency resolver for phase ordering
- [ ] 5. `nexus-sdd orchestrate <hdu-id>` — CLI command to launch swarm

## Phase 2 — Self-Evolving Skill Registry
- [ ] 6. `registry/registry.go` — Skill database with tier management
- [ ] 7. `registry/generator.go` — Auto-skill generation from successful memory clusters
- [ ] 8. `registry/loader.go` — Compact rule injection for sub-agents
- [ ] 9. Skill evolution cycle: track applications, promote tiers, flag deprecated

## Phase 3 — Granular Progress + Session Replay
- [ ] 10. `progress/tracker.go` — Progress chain: every agent action recorded as memory
- [ ] 11. `progress/replay.go` — Session reconstruction from progress memories
- [ ] 12. MCP tool `mem_save_progress` — Agent saves granular progress
- [ ] 13. MCP tool `mem_replay_session` — Agent replays session state

## Phase 4 — Evolutionary Judgment Day
- [ ] 14. `judgment/verifier.go` — Multi-phase verification (spec compliance matrix)
- [ ] 15. `judgment/adversarial.go` — Mutation testing, fuzzing, counterfactual scenarios
- [ ] 16. `judgment/causal.go` — Root cause analysis linking bugs to design decisions
- [ ] 17. Auto-fix loop: generate fix → retest → certify or escalate

## Phase 5 — Adaptive Agent Persona
- [ ] 18. `persona/persona.go` — Persona dimensions: style, architecture, risk, communication
- [ ] 19. `persona/inference.go` — Predict user preferences from historical signals
- [ ] 20. MCP tool `mem_persona_search` — Query user preferences by dimension
- [ ] 21. Cross-tool identity via `.mempack` export/import

## Phase 6 — Self-Healing Infrastructure
- [ ] 22. `heal/checker.go` — Health checks: duplicates, contradictions, stale memories, DB integrity
- [ ] 23. `heal/autoheal.go` — Auto-merge duplicates, deprecate stale, VACUUM, suggest model upgrades
- [ ] 24. CLI command `mnemo heal --full` with human-readable report

## Phase 7 — Predictive Knowledge Pre-Loading
- [ ] 25. `predictor/predictor.go` — Signal collection from git, fsnotify, activity
- [ ] 26. MCP tool `mem_predict_context` — Pre-load relevant memories before agent starts

## Phase 8 — Autonomous Improvement Loops
- [ ] 27. `evolve/analyzer.go` — Detect failure patterns across HDUs, cluster by similarity
- [ ] 28. `evolve/proposer.go` — Auto-generate improvement HDU proposals
- [ ] 29. Meta-memory: track system's own performance for self-optimization

## Phase 9 — Integration & Polish
- [ ] 30. Update MCP server with 6 new tools (total: 8 → 14)
- [ ] 31. Update CLI with new commands: orchestrate, persona, heal, predict, evolve, replay
- [ ] 32. Update `AGENTS.md` with full ecosystem instructions
- [ ] 33. End-to-end test: full SDD cycle with swarm + judgment day + persona + replay
