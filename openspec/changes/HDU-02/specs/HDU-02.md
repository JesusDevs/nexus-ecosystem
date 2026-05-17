# Spec: Pack Import and Memory Marketplace

## BDD Scenarios

### Scenario 1: Import portable pack
```gherkin
Given team A exported banking-v1.mempack with 10 memories
And team B has mnemo installed with no previous memories
When team B runs mnemo pack import banking-v1.mempack
Then the 10 memories are imported into the local DB
And embeddings are preserved without regeneration
And mnemo stats shows 10 memories
```

### Scenario 2: ID conflict on import
```gherkin
Given the local DB already has a memory with id="mem-001"
And the pack to import contains a memory with the same id="mem-001"
When mnemo pack import pack.mempack --on-conflict keep-both is run
Then the conflicting memory is imported with a new UUID
And the original "mem-001" memory is not modified
And the result shows 1 conflict resolved with keep-both strategy
```

### Scenario 3: Install from GitHub
```gherkin
Given a repo github.com/team/banking-knowledge exists with banking-v1.mempack
When mnemo pack install github.com/team/banking-knowledge is run
Then the repo is shallow-cloned
And the .mempack found in root is imported
```

### Scenario 4: Schema validation
```gherkin
Given a corrupt.json file that is not a valid .mempack
When mnemo pack import corrupt.json is run
Then the command fails with "unsupported pack_version" or "invalid schema" error
And the local DB is not modified
```
