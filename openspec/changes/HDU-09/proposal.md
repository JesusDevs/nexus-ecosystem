# Proposal: `gingx-sdd init` -- Scaffolding Command

## Why
The Gingx ecosystem currently requires a 500-line `install.sh` bash script to scaffold a new project. This script:
- Copies templates manually with `cp` commands
- Hard-codes directory creation
- Depends on bash being available (not cross-platform)
- Cannot be invoked from the CLI as `gingx-sdd init`
- Has no dry-run mode, no force flag, no stack detection feedback

The **user need**: "cd into my new project folder, run `gingx-sdd init`, and everything is set up -- hooks, profiles, OpenSpec, mnemo, settings, and stack detection."

`gingx-sdd` already has:
- A `typer`-based CLI with commands like `spec`, `orchestrate`, `hdu`, `team`, `mode`, `knowledge`
- Rich templates at `gingx-sdd/templates/` (hooks, profiles, config, settings, OpenSpec, MCP)
- A stack detector at `gingx_sdd/detector/scanner.py` that identifies languages, frameworks, testing, databases
- Profile system with 7 profiles matching detected stacks

There is NO `init` command. The scaffolding gap forces users to either run a bash script or copy files manually.

## What Changes
- **New `init` subcommand** on the `gingx-sdd` CLI, accessible as `gingx-sdd init`
- **Project scaffolding logic**: create `.gingx/` structure, install hooks, copy settings, initialize OpenSpec
- **Stack detection integration**: run `detector.scanner.detect_project_type()`, select recommended profile, report findings
- **Flags**: `--dry-run` (preview only), `--force` (overwrite existing), `--stack` (manual override)
- **Summary output**: a clear table showing what was created and detected

## What Does NOT Change
- `install.sh` remains as-is (serves a different purpose: system-level install of prerequisites)
- Templates at `gingx-sdd/templates/` remain unchanged (init reads from them)
- No new dependencies (uses `shutil`, `pathlib`, and existing `typer`/`yaml` imports)
- No changes to `detector/scanner.py` or any other existing module
- Mnemo, Ollama, or other tool installation stays in install.sh scope -- init scaffolds project structure only

## Impact
- HDU: HDU-09
- Complexity: low
- Dependencies: none (reuses detector, profiles, templates -- all already present)
- Estimated lines: ~100 Python (one new `init_project.py` module + ~20 lines in `cli.py` to register command)
- Files touched: `gingx_sdd/cli.py` (add command), `gingx_sdd/init_project.py` (new), `gingx-sdd/pyproject.toml` (no change needed)
- User experience: **single command** -- `gingx-sdd init` in any directory
