# Tasks: Portable Mnemo Memory

## Phase 1 -- Entries Serialization (Go: store.go)
- [ ] 1. Add `SerializeEntryJSON(mem *VectorMemory) ([]byte, error)` to `vec/store.go` -- serialize entry fields (excluding embedding binary) as a single JSON line. Reuse MemoryPackEntry struct fields.
- [ ] 2. Add `findGingxDir() (string, error)` helper to `gingx-mnemo/main.go` -- walk up from cwd max 5 levels looking for `.gingx/` directory
- [ ] 3. Add `appendToMemoryFiles(gingxDir string, mem *VectorMemory) error` to `gingx-mnemo/main.go` -- append JSONL line to `entries.jsonl`, update `embeddings.json` with ID->embedding mapping

## Phase 2 -- Dual-Write on Save (Go: main.go)
- [ ] 4. Modify `runSave()` in `main.go`: after `store.Save(mem)`, call `appendToMemoryFiles()` if `.gingx/` found. Silently skip if not found.
- [ ] 5. Same dual-write logic in `mcp/server.go` `handleSave()`: after `s.store.Save(mem)`, append to `.gingx/memory/` files

## Phase 3 -- Import Subcommand (Go: main.go + store.go)
- [ ] 6. Add `import` case to `main()` switch in `main.go`
- [ ] 7. Implement `runImport()` in `main.go`:
  - Resolve entries path (--path flag or walk-up find `.gingx/memory/entries.jsonl`)
  - Read entries.jsonl line by line with `bufio.Scanner`
  - For each entry, check if ID exists in DB (new method `store.Exists(id) bool`)
  - If new: load embedding from embeddings.json, or generate via embedder if missing
  - Insert into DB via existing `store.Save()` or new bulk method
  - Report: "imported N, skipped M"
- [ ] 8. Support `--reindex` flag: ignore embeddings.json, regenerate all embeddings, rewrite embeddings.json
- [ ] 9. Support `--path <file>` flag: import from arbitrary path instead of auto-discovery
- [ ] 10. Support `--yes` flag: skip confirmation prompt (for scripted use in session-start.sh)

## Phase 4 -- session-start.sh Auto-Detect (Bash)
- [ ] 11. Add mnemo import block to `.claude/hooks/session-start.sh`:
  ```bash
  if command -v mnemo &>/dev/null && [[ -f ".gingx/memory/entries.jsonl" ]]; then
      IMPORT_RESULT=$(mnemo import --yes 2>/dev/null || echo "")
      if [[ -n "$IMPORT_RESULT" ]]; then
          MNEMO_IMPORT_MSG=" Mnemo memory imported from .gingx/memory/ ($IMPORT_RESULT)."
      fi
  fi
  ```
  Include `$MNEMO_IMPORT_MSG` in the systemMessage JSON output.
- [ ] 12. Apply same change to template `gingx-sdd/templates/.claude/hooks/session-start.sh`

## Phase 5 -- Sync Deprecation Notice (Go)
- [ ] 13. Add deprecation hint to `mnemo sync push` output: append "Hint: use .gingx/memory/ for portable memory -- commit entries.jsonl to your repo."
- [ ] 14. Add same hint to `mnemo sync pull` output

## Phase 6 -- Integration Verification
- [ ] 15. Manual test: `mnemo save "test" "content"` -- verify entries.jsonl created with 1 line, embeddings.json has embedding
- [ ] 16. Manual test: `mnemo import` in a fresh clone -- verify entries appear in local DB
- [ ] 17. Manual test: `mnemo import` twice -- verify idempotent (no duplicates)
- [ ] 18. Manual test: run session-start.sh with entries.jsonl present, verify systemMessage includes import notification
- [ ] 19. Manual test: `mnemo import --reindex` -- verify embeddings regenerated, embeddings.json rewritten
