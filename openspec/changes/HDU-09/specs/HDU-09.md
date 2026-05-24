# Spec: `gingx-sdd init` -- Scaffolding Command

## User Story
As a developer starting a new Gingx project,
I want to run `gingx-sdd init` in my project directory,
so that the full SDD harness (hooks, profiles, OpenSpec, mnemo config, settings) is set up in seconds with zero manual steps.

## BDD Scenarios

### Scenario 1: Init in an empty directory
```gherkin
Given an empty directory /tmp/my-new-project
And gingx-sdd is installed and on PATH
When the user runs `gingx-sdd init`
Then a .gingx/ directory is created with config.yaml, suites.yaml, and profiles/
And .claude/hooks/ is created with pre-tool-use.sh, stop.sh, and session-start.sh (all executable)
And .claude/settings.local.json is created with hook configuration
And openspec/ directory is created with changes/ and AGENTS.md
And .mcp.json is created with codegraph MCP server config
And the summary output lists each file/directory created with a status indicator
```

### Scenario 2: Init detects Python project and recommends profile
```gherkin
Given a directory containing pyproject.toml and a src/ folder
When the user runs `gingx-sdd init`
Then the stack detector identifies Python with setuptools
And the recommended profile is "fullstack"
And the init copies all profiles, with the recommended one highlighted in the summary
And the output includes "Detected: Python (fullstack profile recommended)"
```

### Scenario 3: Init detects Go project and recommends fullstack-go profile
```gherkin
Given a directory containing go.mod and go.sum
When the user runs `gingx-sdd init`
Then the stack detector identifies Go
And the recommended profile is "fullstack-go"
And the output includes "Detected: Go (fullstack-go profile recommended)"
```

### Scenario 4: Init detects Node/React project
```gherkin
Given a directory containing package.json with react dependency
And tsconfig.json is present
When the user runs `gingx-sdd init`
Then the stack detector identifies TypeScript and React
And the recommended profile is "react-nextjs"
And the output includes "Detected: TypeScript, React (react-nextjs profile recommended)"
```

### Scenario 5: Init detects no recognizable stack
```gherkin
Given an empty directory with no language-specific files
When the user runs `gingx-sdd init`
Then the stack detector returns type "cli" with no languages detected
And the recommended profile is "minimal"
And the output includes "No stack detected -- using minimal profile"
And all scaffolding files are still created successfully
```

### Scenario 6: Init with --stack override
```gherkin
Given any project directory
When the user runs `gingx-sdd init --stack go`
Then the stack detector is bypassed
And the init uses the "fullstack-go" profile regardless of actual files present
And the output includes "Stack override: go (fullstack-go profile)"
```

### Scenario 7: Init --dry-run shows what would happen
```gherkin
Given an empty directory
When the user runs `gingx-sdd init --dry-run`
Then no files or directories are created on disk
And the output lists every file and directory that would be created
And the output includes "(dry-run -- no files written)"
```

### Scenario 8: Init in a directory that already has .gingx/
```gherkin
Given a directory where .gingx/ already exists
When the user runs `gingx-sdd init`
Then the command warns: ".gingx/ already exists"
And the command exits with code 1 without modifying anything
And the error message includes "Use --force to overwrite"
```

### Scenario 9: Init --force overwrites existing .gingx/
```gherkin
Given a directory where .gingx/ already exists
When the user runs `gingx-sdd init --force`
Then the existing .gingx/ is backed up or overwritten
And all scaffolding proceeds normally
And the output includes "Overwriting existing .gingx/ (--force)"
```

### Scenario 10: Init creates current_task.yaml tracking file
```gherkin
Given any project directory
When the user runs `gingx-sdd init`
Then .gingx/current_task.yaml is created
And it contains `hdu_id: none`, `phase: none`, `agent: none`
And the SDD gate hook will block code writes until an HDU is created
```

### Scenario 11: Hooks are executable after init
```gherkin
Given a Unix-like system
When the user runs `gingx-sdd init`
Then .claude/hooks/pre-tool-use.sh has executable permissions
And .claude/hooks/stop.sh has executable permissions
And .claude/hooks/session-start.sh has executable permissions
```

### Scenario 12: Init respects GINGX_TEMPLATES environment variable
```gherkin
Given the environment variable GINGX_TEMPLATES=/custom/templates
And /custom/templates/ contains the same structure as gingx-sdd/templates/
When the user runs `gingx-sdd init`
Then templates are copied from /custom/templates/ instead of the built-in path
And the output includes "Using custom templates: /custom/templates/"
```

## Acceptance Criteria Checklist
- [ ] `gingx-sdd init` creates the full scaffolding in an empty directory
- [ ] Stack detection runs and recommends the correct profile
- [ ] All 7 profiles are copied to .gingx/profiles/
- [ ] Hooks are copied and made executable on Unix
- [ ] settings.local.json has correct hook configuration
- [ ] OpenSpec AGENTS.md and changes/ directory are created
- [ ] .mcp.json is copied
- [ ] --dry-run previews without writing files
- [ ] --force allows overwriting existing .gingx/
- [ ] Without --force, init refuses to overwrite
- [ ] --stack bypasses auto-detection with user-specified stack
- [ ] Summary output is clear and actionable
- [ ] current_task.yaml is created with "none" defaults
- [ ] Zero new Python dependencies

## Priority
| Item | Priority |
|------|----------|
| Scaffolding (directories + files) | Must have |
| Stack detection + profile recommendation | Must have |
| Hook installation with executable permissions | Must have |
| --dry-run flag | Must have |
| --force flag | Must have |
| Summary output | Must have |
| --stack override | Should have |
| GINGX_TEMPLATES env var | Could have |
| Backup of existing .gingx/ on --force | Could have |
