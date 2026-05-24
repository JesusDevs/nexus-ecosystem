# Spec: Semantic Release System + Harness Engineering

## BDD Scenarios

### Scenario 1: First release of a project
```gherkin
Given a project with mnemo memories and committed code
When the user runs `gingx-sdd release v0.1.0 --project gingx-mnemo`
Then a git tag "v0.1.0" is created
And a mnemo release snapshot is saved
And a CHANGELOG.md entry is generated
And the version is recorded in mnemo config
```

### Scenario 2: Pre-release check catches issues
```gherkin
Given a project with uncommitted changes
When the user runs `gingx-sdd release v0.1.0 --dry-run`
Then the system reports "uncommitted changes found"
And the release is blocked
And no tag or snapshot is created
```

### Scenario 3: Swarm mode change without restart
```gherkin
Given mnemo is running in "supervisor" mode
When the user runs `mnemo config set swarm.mode swarm`
And invokes `gingx-sdd orchestrate HDU-06 --phase apply`
Then tasks are claimed in distributed swarm mode
And the mode change took effect immediately
```

### Scenario 4: Release with failing security scan
```gherkin
Given a project with a hardcoded API key
When the user runs `gingx-sdd release v0.1.0`
Then the security scan detects the key
And the release is blocked
And a report is saved to mnemo
```

### Scenario 5: Changelog auto-generation
```gherkin
Given commits since the last release tag
When `gingx-sdd release` generates the changelog
Then commits are categorized into features, fixes, improvements
And the changelog is appended to CHANGELOG.md
And the entry includes the version, date, and commit range
```

### Scenario 6: Cross-project release
```gherkin
Given a monorepo with gingx-mnemo (Go) and gingx-sdd (markdown)
When the user runs `gingx-sdd release v0.2.0 --all`
Then both projects get tagged
And each project's memories are snapshotted independently
And a unified CHANGELOG.md is generated
```
