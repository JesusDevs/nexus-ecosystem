# Spec: Vector Memory for Agent Feedback

## BDD Scenarios

### Scenario 1: Agent saves memory autonomously via MCP
```gherkin
Given the mnemo MCP server is running
And the agent just resolved a bug
When the agent calls mem_save with title="Fix NPE in login", content="Null pointer was due to uninitialized session", type="bugfix", outcome="resolved"
Then the memory is saved in vec_memories with a generated 1024-dim embedding
And the media_type field is "text"
And the response includes the created memory id
```

### Scenario 2: Agent searches by meaning, not keywords
```gherkin
Given there are memories about "HTTP connection timeout" and "cooking recipes"
When the agent calls mem_search_semantic with query="network issues"
Then the HTTP timeout memory appears first
And the cooking recipe memory does not appear in results
```

### Scenario 3: Memory versioning
```gherkin
Given the "banking-app" project has 5 saved memories
When the agent runs mnemo release banking-app v1.0.0 --description "First release"
Then a snapshot is created in vec_releases with memory_count=5
And mem_list_releases shows v1.0.0 with the correct description
```

### Scenario 4: Multimodal memory (image)
```gherkin
Given the agent has a screenshot of a UI error
When it calls mem_save with title="Visual error in dashboard", media_type="image", media_file=<blob>
Then the memory is saved with media_type="image"
And the media_file is stored as BLOB in SQLite
```
