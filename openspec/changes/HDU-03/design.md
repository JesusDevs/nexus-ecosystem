# Design: Gingx-SDD Save Harness Integration

## Approach
`gingx-sdd save --hdu-id <ID>`:
1. Reads `openspec/changes/<ID>/proposal.md` -> title and description
2. Reads `openspec/changes/<ID>/design.md` -> extracts tech keywords as tags
3. Reads `openspec/changes/<ID>/tasks.md` -> counts `[x]` vs `[ ]`, detects bugs
4. Builds combined content: `# <title>\n\n## Proposal\n<proposal>\n\n## Design\n<design>`
5. Detects outcome: all [x] -> "resolved", some [x] -> "partial", none -> "pending"
6. Calls `mnemo save` via subprocess with: `--type decision --project <project> --outcome <outcome> --tags <tags> --hdu-id <ID>`

Fallback: if mnemo CLI is not available, tries `_save_via_mcp()` which sends JSON-RPC directly to the MCP server.

`gingx-sdd release <version>`:
- Wrapper for `mnemo release <project> <version> --description "Release from gingx-sdd"`
- Project is detected from `pyproject.toml` or root directory name

## Alternatives Considered
1. **Always call MCP directly** — Cleaner but requires MCP server to be running. Subprocess is more robust for CI/CD.
2. **Auto-save without command** — User loses control. Better to have explicit command with optional `--auto` flag for CI.
3. **Integrate into Ralph Loop** — The loop already knows if tests passed. But post-test hook is cleaner than mixing concerns.

## Decision
Subprocess as primary path, MCP as fallback. Auto-detection of outcome, bugs, and tags to reduce friction. The user only needs to know the HDU ID.
