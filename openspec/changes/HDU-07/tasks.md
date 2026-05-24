# Tasks: Semantic Release System + Harness Engineering

## Phase 1 — Swarm Mode Configuration
- [x] 1. Add `swarm.mode` default to `vec/store.go` initDefaults() (value: "hybrid")
- [x] 2. Update `gingx-sdd orchestrate` to read swarm.mode from mnemo config
- [x] 3. Display current mode in `gingx-sdd orchestrate --status`

## Phase 2 — Release Command
- [x] 4. Add `release` command to `gingx_sdd/cli.py` with semver validation
- [x] 5. Implement `--dry-run` pre-release checks (git clean, build, tests, security)
- [x] 6. Implement git tag creation with version
- [x] 7. Implement mnemo release snapshot delegation
- [x] 8. Integrate changelog-generator skill for auto-changelog

## Phase 3 — Harness Engineering Config
- [x] 9. Update `.gingx/config.yaml` template with harness section
- [x] 10. Document harness model in AGENTS.md: feedforward guides + feedback sensors + steering
- [x] 11. Add harness health check: `gingx-sdd status` shows sensor/guide status

## Phase 4 — First Release
- [ ] 12. Execute `gingx-sdd release v0.1.0 --all` for gingx-mnemo
- [ ] 13. Execute `gingx-sdd release v0.1.0 --all` for gingx-sdd
- [ ] 14. Generate unified CHANGELOG.md for the monorepo
