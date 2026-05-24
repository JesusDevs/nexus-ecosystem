# Spec: Portable Mnemo Memory

## User Story
As a developer cloning a Gingx project,
I want the project's agent memory to come with the repo,
so that I have immediate context from past decisions, bugs, and architectural notes without any setup.

## BDD Scenarios

### Scenario 1: Save dual-writes to project repo
```gherkin
Given a Gingx project with .gingx/ directory
When the user runs `mnemo save "Bug: login timeout" "Fixed by increasing ttl" --type bug --outcome solucionado`
Then the memory is saved to ~/.mnemo/mnemo.db (local cache)
And the memory is appended to .gingx/memory/entries.jsonl as a new JSON line
And the embedding is appended to .gingx/memory/embeddings.json
And .gingx/memory/entries.jsonl contains the id, project, title, content, type, tags, outcome, media_type, version, and created_at fields
```

### Scenario 2: First-time clone auto-imports memory
```gherkin
Given a fresh git clone of a Gingx project
And the repo contains .gingx/memory/entries.jsonl with 4 entries
And mnemo is installed but ~/.mnemo/mnemo.db has no entries for this project
When session-start.sh runs on first session
Then it detects .gingx/memory/entries.jsonl exists
And it runs `mnemo import --yes` silently
And the 4 entries are indexed into the local mnemo.db
And the systemMessage confirms: "Mnemo memory imported from .gingx/memory/"
```

### Scenario 3: Import skips already-imported entries
```gherkin
Given the local mnemo.db already has entries matching .gingx/memory/entries.jsonl (by ID)
When session-start.sh runs
Then `mnemo import` detects entries already exist
And no duplicate entries are created
And the import reports "0 new, 4 already present"
```

### Scenario 4: Import with reindex regenerates embeddings from scratch
```gherkin
Given a project with .gingx/memory/entries.jsonl but no embeddings.json
And Ollama is running with bge-m3 available
When the user runs `mnemo import --reindex`
Then all entries are read from entries.jsonl
And new embeddings are generated via Ollama
And the result is written to both mnemo.db and .gingx/memory/embeddings.json
```

### Scenario 5: Save without .gingx/ directory does not crash
```gherkin
Given a project directory without .gingx/ directory
When the user runs `mnemo save "test" "content"`
Then the memory is saved to ~/.mnemo/mnemo.db normally
And no error is shown about missing .gingx/memory/
And the save completes successfully
```

### Scenario 6: entries.jsonl is git-friendly and diffable
```gherkin
Given .gingx/memory/entries.jsonl with 5 entries
When `git diff` is run on the file
Then the diff shows one entry (JSON object) per line
And changes to individual entries are human-readable
And the file can be merged without special tooling
```

### Scenario 7: Import subcommand from arbitrary path
```gherkin
Given a memory pack at /tmp/backup/entries.jsonl
When the user runs `mnemo import --path /tmp/backup/entries.jsonl`
Then the entries are indexed into the local mnemo.db
And embeddings are generated if not present
And the import reports the count of new vs. skipped entries
```

### Scenario 8: Sync remote is deprecated but still functional
```gherkin
Given a user has sync.remote configured
When they run `mnemo sync push`
Then a deprecation notice is shown: "Hint: use .gingx/memory/ for portable memory instead"
And the push still proceeds as before
```

## Acceptance Criteria Checklist
- [ ] `mnemo save` writes to both mnemo.db AND .gingx/memory/entries.jsonl
- [ ] `mnemo import` reads entries.jsonl and indexes into local DB
- [ ] `mnemo import --reindex` regenerates embeddings from scratch
- [ ] `mnemo import --path <file>` imports from arbitrary path
- [ ] session-start.sh auto-detects .gingx/memory/entries.jsonl and imports
- [ ] Idempotent: importing twice does not duplicate entries
- [ ] Graceful degradation: works without .gingx/ directory
- [ ] entries.jsonl is one JSON object per line (JSONL format)
- [ ] No new dependencies introduced

## Priority
| Item | Priority |
|------|----------|
| Dual-write on save | Must have |
| import subcommand | Must have |
| session-start auto-detect | Must have |
| Idempotent import | Must have |
| --reindex flag | Should have |
| --path flag | Should have |
| Sync deprecation notice | Could have |
| embeddings.json auto-generation | Could have |
