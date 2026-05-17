# Design: Semantic Release System

## Approach

### Release Command Flow
```
nexus-sdd release v0.1.0 --project nexus-mnemo
          │
          ├─ 1. Pre-release checks (--dry-run)
          │     ├─ Security scan: gitleaks or mnemo security
          │     ├─ Build check: go build ./... or python -m build
          │     ├─ Test check: test suite pass?
          │     ├─ mnemo conflicts: any contradictions?
          │     └─ Git cleanliness: no uncommitted changes
          │
          ├─ 2. Git tag
          │     ├─ git tag -a v0.1.0 -m "Release v0.1.0"
          │     └─ Tag format: v<MAJOR>.<MINOR>.<PATCH>
          │
          ├─ 3. Mnemo snapshot
          │     └─ mnemo release <project> <version>
          │
          ├─ 4. Artifacts
          │     ├─ Binary (Go): go build -o dist/mnemo-v0.1.0-darwin-arm64
          │     ├─ Mempack: mnemo pack export <project> --version v0.1.0
          │     └─ Skills: tar -czf skills-v0.1.0.tar.gz skills/
          │
          └─ 5. Changelog
                └─ changelog-generator skill → CHANGELOG.md
```

### Semver Validation
- Format: `v<MAJOR>.<MINOR>.<PATCH>[-prerelease]`
- MAJOR: breaking changes
- MINOR: new features, backward compatible
- PATCH: bug fixes
- First release: `v0.1.0`

### Swarm Mode Configuration
```bash
mnemo config set swarm.mode hybrid    # Default: DAG + Supervisor + Swarm
mnemo config set swarm.mode dag       # Dependency-only, max parallelism
mnemo config set swarm.mode supervisor # Centralized delegation
mnemo config set swarm.mode swarm     # Distributed claim-based
```
Mode is read from DB at the start of each phase — no restart needed.

### Harness Engineering Integration
Following the Thoughtworks harness model:
- **Feedforward guides**: Skills (.md) + AGENTS.md + .nexus/profiles/
- **Computational sensors**: Security scan, build check, test coverage, dependency audit
- **Inferential sensors**: QA agent adversarial testing, mnemo conflict detection
- **Steering**: Hooks auto-update behavior based on detected patterns
- **Pre-release gate**: All sensors must pass before release

## Alternatives Considered
1. **GitHub Releases API** — Tight coupling to GitHub. Rejected: must work offline, multi-platform.
2. **goreleaser** — Excellent for Go, but doesn't cover Python or mempacks. Can integrate later.

## Decision
Custom release command in `nexus-sdd` (Python CLI) that delegates to git, mnemo, and shell for artifact building. Lightweight, portable, no new dependencies beyond `typer`.
