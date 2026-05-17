# Proposal: Nexus-SDD Save Harness Integration

## Why
The agent completes an HDU but the knowledge is lost if not explicitly saved. `nexus-sdd save --hdu-id <ID>` must read the HDU's OpenSpec files, automatically build a vector memory, detect outcome and bugs, and save it via mnemo. This closes the SDD -> memory loop without manual intervention.

## What Changes
- `nexus-sdd save --hdu-id <ID>` reads spec + design + tasks from the HDU
- Auto-detection of outcome: counts [x]/[ ] in tasks.md
- Bug detection from Ralph Loop in task descriptions
- Post-test hook: if tests pass, suggests saving automatically
- `nexus-sdd release <version>` wrapper for `mnemo release`
- Tags auto-detected from tech stack mentioned in design.md

## What Does NOT Change
- mnemo CLI (nexus-sdd calls it via subprocess)
- Memory format (VectorMemory with version, media_type, tags)
- OpenSpec structure (proposal.md, design.md, tasks.md)

## Impact
- HDU: HDU-03
- Complexity: low
- Files: nexus_sdd/cli.py
- New commands: 2 (save, release)
- Dependency: mnemo CLI installed and in PATH
