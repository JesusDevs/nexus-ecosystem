# Tasks: Knowledge Pipeline — Auto-ingest, Classification, and Federation

- [ ] 1. `mnemo ingest --file <path> --project <name>` — one-shot file ingestion via CLI
- [ ] 2. `ingest/classifier.go` — keyword-based classifier (Tier 1)
- [ ] 3. `ingest/classifier.go` — LLM fallback classifier via Ollama (Tier 2)
- [ ] 4. `ingest/watcher.go` — fsnotify watcher with debounce and file_hash tracking
- [ ] 5. `mnemo ingest --watch <dir> --project <name>` — watch mode CLI
- [ ] 6. `vec/store.go` — `MergeProjects()` method with semantic dedup (>0.95 threshold)
- [ ] 7. `mnemo merge --projects a,b,c --into global` — CLI command
- [ ] 8. MCP tool `mem_ingest_file` — agent-driven file ingestion
- [ ] 9. MCP tool `mem_merge_projects` — agent-driven project federation
- [ ] 10. Relationship detection — link new memories to similar existing ones via semantic search
- [ ] 11. Updated `vec_memories` schema: add `file_hash` and `superseded_by` columns
- [ ] 12. Updated `.mempack` format: model-agnostic (text always, embeddings optional cache)
- [ ] 13. Test with real multi-project directory: ingest 100+ files, merge, search
