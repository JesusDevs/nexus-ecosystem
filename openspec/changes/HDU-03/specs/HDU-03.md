# Spec: Nexus-SDD Save Harness Integration

## BDD Scenarios

### Scenario 1: Save completed HDU
```gherkin
Given HDU "HDU-fix-login" has proposal.md, design.md, and tasks.md with all tasks [x]
And mnemo CLI is installed and in PATH
When nexus-sdd save --hdu-id HDU-fix-login is run
Then the 3 HDU files are read
And outcome is auto-detected as "resolved"
And mnemo save is called with --type decision --outcome resolved --hdu-id HDU-fix-login
```

### Scenario 2: Auto-detect partial outcome
```gherkin
Given HDU "HDU-partial" has 8 tasks total, 5 marked [x] and 3 marked [ ]
When nexus-sdd save --hdu-id HDU-partial is run
Then outcome is auto-detected as "partial"
And memory is saved with outcome="partial"
```

### Scenario 3: Release wrapper
```gherkin
Given the "nexus-ecosystem" project has saved memories
When nexus-sdd release v0.3.0 is run
Then mnemo release nexus-ecosystem v0.3.0 is called
And a snapshot is created in vec_releases
```

### Scenario 4: Mnemo not installed -> MCP fallback
```gherkin
Given mnemo CLI is not in PATH
But the mnemo MCP server is running on the system
When nexus-sdd save --hdu-id HDU-test is run
Then mnemo save via subprocess is attempted (fails)
And _save_via_mcp() is used as fallback
And the memory is correctly saved via JSON-RPC
```
