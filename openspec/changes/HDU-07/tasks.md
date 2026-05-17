# Tasks: Semantic Release System + Harness Engineering

## Phase 1 — Swarm Mode Configuration
- [x] 1. Add `swarm.mode` default to `vec/store.go` initDefaults() (value: "hybrid")
- [x] 2. Update `nexus-sdd orchestrate` to read swarm.mode from mnemo config
- [x] 3. Display current mode in `nexus-sdd orchestrate --status`

## Phase 2 — Release Command
- [x] 4. Add `release` command to `nexus_sdd/cli.py` with semver validation
- [x] 5. Implement `--dry-run` pre-release checks (git clean, build, tests, security)
- [x] 6. Implement git tag creation with version
- [x] 7. Implement mnemo release snapshot delegation
- [x] 8. Integrate changelog-generator skill for auto-changelog

## Phase 3 — Harness Engineering Config
- [x] 9. Update `.nexus/config.yaml` template with harness section
- [x] 10. Document harness model in AGENTS.md: feedforward guides + feedback sensors + steering
- [x] 11. Add harness health check: `nexus-sdd status` shows sensor/guide status

## Phase 4 — First Release
- [ ] 12. Execute `nexus-sdd release v0.1.0 --all` for nexus-mnemo
- [ ] 13. Execute `nexus-sdd release v0.1.0 --all` for nexus-sdd
- [ ] 14. Generate unified CHANGELOG.md for the monorepo
