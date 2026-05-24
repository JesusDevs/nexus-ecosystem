# Tasks: Rename engram-vec to gingx-mnemo

- [x] 1. Rename in server.go (Initialize, logs)
- [x] 2. Rename in main.go (usage, logs, version string, env vars)
- [x] 3. Change module path in go.mod: github.com/gingx-sdd/engram-vec -> gingx-mnemo
- [x] 4. Update imports in main.go and mcp/server.go
- [x] 5. Change DB path: ~/.engram/engram.db -> ~/.mnemo/mnemo.db
- [x] 6. Remove `engram_memory_id` column and `EngramMemoryID` field from schema + structs
- [x] 7. Rename env vars: ENGRAM_DIR -> MNEMO_DIR, ENGRAM_VEC_PROJECT -> MNEMO_PROJECT
- [x] 8. Update README.md — remove Engram extension framing, standalone identity
- [x] 9. Update CHANGELOG.md — remove Engram references
- [x] 10. Update install.sh
- [x] 11. Rename directory: engram-vec/ -> gingx-mnemo/
- [x] 12. Update references in gingx-sdd/install.sh
