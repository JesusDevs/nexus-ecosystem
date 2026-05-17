# Spec: Knowledge Pipeline — Auto-ingest, Classification, and Federation

## BDD Scenarios

### Scenario 1: Auto-detect new file and ingest
```gherkin
Given the watcher is running: mnemo ingest --watch ./fuentes/ --project banking-app
And a new file ./fuentes/docs/architecture.md is created
When the watcher detects the file (within 500ms debounce window)
Then the file is read and classified
And type is auto-detected as "spec" based on path and content
And the memory is saved to vec_memories with project="banking-app"
And the file_hash is stored to detect future modifications
```

### Scenario 2: Modified file triggers re-ingest
```gherkin
Given architecture.md was previously ingested with hash="abc123"
When the file is modified and hash changes
Then the watcher detects the modification
And the previous memory is flagged as superseded
And a new memory is created with the updated content
And both versions are linked via a metadata field
```

### Scenario 3: Classification fallback to LLM
```gherkin
Given a file content.md with ambiguous content (no clear keywords)
When the keyword classifier returns confidence < 0.7
Then the LLM classifier is invoked with the file content
And the LLM returns {type: "architecture", outcome: "active", tags: ["distributed", "event-driven"]}
And the memory is saved with those classifications
```

### Scenario 4: Merge projects into federated knowledge
```gherkin
Given project "core-banking" has 150 memories
And project "payment-service" has 80 memories
And there are 5 near-duplicate memories across both (cosine > 0.95)
When mnemo merge --projects core-banking,payment-service --into global is run
Then 225 memories are copied to project "global" (230 - 5 deduped)
And each memory has source_project tag preserved
And the merge result shows: total=230, deduped=5, new=225
```

### Scenario 5: Agent ingests file via MCP
```gherkin
Given the mnemo MCP server is running
And a file ./docs/postmortems/outage-2026-05.md exists
When the agent calls mem_ingest_file with path="./docs/postmortems/outage-2026-05.md" and project="banking-app"
Then the file is read, classified as "postmortem", embedded, and saved
And the response includes the memory id and detected type
```

### Scenario 6: Cloned repo inherits team knowledge
```gherkin
Given a new developer clones the banking-app repo
And the repo includes banking-knowledge.mempack
When the developer runs mnemo pack import banking-knowledge.mempack
Then all 500+ team memories are available locally
And mnemo search "auth patterns" returns relevant results immediately
```

### Scenario 7: Non-md files are skipped
```gherkin
Given the watcher is monitoring ./fuentes/
When a file image.png is added
Then the file is skipped (not an ingestable type)
And a debug log is emitted: "skipping non-ingestable file: image.png"
```
