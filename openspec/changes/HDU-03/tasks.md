# Tasks: Nexus-SDD Save Harness Integration

- [x] 1. `nexus-sdd save --hdu-id <ID>` — reads spec + design + tasks and builds memory
- [x] 2. Detect Ralph Loop bugs -> include them in memory automatically
- [x] 3. Post-test hook in harness: when tests pass, suggest saving
- [x] 4. Integrate with `mnemo` CLI for saving
- [x] 5. Auto-detect outcome: if tests pass -> "resolved"
- [x] 6. `nexus-sdd release <version>` — wrapper for `mnemo release`
