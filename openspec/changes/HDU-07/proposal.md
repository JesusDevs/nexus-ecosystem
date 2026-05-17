# Proposal: Semantic Release System + Harness Engineering

## Why
The Nexus ecosystem needs deterministic, reproducible releases. Every version must produce versioned artifacts (binaries, mempacks, changelogs) and a mnemo snapshot capturing what was learned. The harness engineering framework from Birgitta Böckeler (Thoughtworks) validates our approach: feedforward guides (skills, personas), feedback sensors (QA, security), and steering loops (hooks). This HDU formalizes both into the SDD workflow.

## What Changes
- **Release system**: `nexus-sdd release <version>` command with semver validation, git tags, mnemo snapshots, artifact generation, and pre-release dry-run checks
- **Swarm mode switching**: `mnemo config set swarm.mode dag|supervisor|swarm|hybrid` — mode changes apply without restart
- **Harness engineering formalization**: config.yaml extended with sensor/guide configuration, hook validation
- **CHANGELOG auto-generation**: integrate existing changelog-generator skill into release flow
- **Pre-release checks**: security scan, test coverage, dependency audit, drift detection

## What Does NOT Change
- Mnemo DB schema (vec_config already extensible)
- OpenSpec format (HDU structure unchanged)
- Existing skills/personas (backward compatible)

## Impact
- HDU: HDU-07
- Complexity: medium
- Dependencies: HDU-06 (swarm orchestration concepts)
