# Spec: Rename engram-vec to gingx-mnemo

## BDD Scenarios

### Scenario 1: mnemo binary functional
```gherkin
Given the project was renamed from engram-vec to gingx-mnemo
And the binary was compiled
When mnemo version is run
Then the output shows "gingx-mnemo v0.2.0"
And does not mention "engram-vec"
```

### Scenario 2: MCP server identifies correctly
```gherkin
Given the MCP server is running with mnemo mcp
When an MCP client calls initialize
Then the server responds with name="gingx-mnemo" and version="0.2.0"
And does not mention "engram-vec" in the identity
```

### Scenario 3: Standalone DB
```gherkin
Given gingx-mnemo is installed fresh
When mnemo setup is run
Then the DB is created at ~/.mnemo/mnemo.db
And no references to Engram exist in the schema
And the `engram_memory_id` column does not exist
```

### Scenario 4: MCP registration in Claude Code
```gherkin
Given gingx-mnemo is installed via install.sh
When the script runs claude mcp add mnemo -- mnemo mcp
Then Claude Code registers the server as "mnemo"
And all 8 tools are available to the agent
```
