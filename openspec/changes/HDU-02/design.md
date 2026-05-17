# Design: Pack Import and Memory Marketplace

## Approach
Import `.mempack` JSON in 3 phases:
1. **Validation**: check compatible `pack_version`, JSON schema, embeddings present
2. **Conflict detection**: compare each entry `id` against `vec_memories.id`. If exists -> conflict.
3. **Merge**: based on `--on-conflict` flag (skip, overwrite, keep-both). Default: `keep-both` (generates new UUID).

`mnemo pack install <url>` does a shallow `git clone` of the repo, finds `.mempack` in root, runs import.

Marketplace: `index.json` file in a shared repo. Lists packs with metadata (project, version, description, tags, download_count). `mnemo pack search` uses embeddings for semantic matching against pack descriptions.

## Alternatives Considered
1. **Centralized registry (npm/PyPI-style)** — Requires server, auth, maintenance. Overkill for the use case (teams share via git).
2. **IPFS for distribution** — Decentralized but adds complex dependency. Git is already sufficient for team sharing.
3. **Import without merge strategies** — Simpler code but forces users to resolve conflicts manually.

## Decision
Git as transport + portable JSON + merge strategies in CLI. The marketplace is a static index in git, not a service. Zero external dependencies.
