# Design: `gingx-sdd init` -- Scaffolding Command

## Approach

### High-Level Architecture

```
gingx-sdd init [--dry-run] [--force] [--stack <name>]
  │
  ├─ 1. Validate: check .gingx/ doesn't exist (unless --force)
  │
  ├─ 2. Detect stack: call detect_project_type()
  │     └─ If --stack: override with user-specified stack
  │
  ├─ 3. Copy templates: shutil.copytree from templates/ to project root
  │     ├─ .gingx/config.yaml, suites.yaml, profiles/*.yaml
  │     ├─ .claude/hooks/*.sh + chmod +x
  │     ├─ .claude/settings.local.json
  │     ├─ openspec/AGENTS.md + create changes/
  │     └─ .mcp.json
  │
  ├─ 4. Create current_task.yaml with default "none" values
  │
  └─ 5. Print summary: table of created files + detected stack + next steps
```

### Decision 1: New Module vs. Inline in cli.py

```
Decision: New module gingx_sdd/init_project.py with a single public function, 
          registered as a typer command in cli.py
Alternatives:
  A. Inline the ~100 lines directly in cli.py — pro: no new file, con: cli.py already 1126 lines; adding scaffolding logic bloats it further
  B. New module init_project.py — pro: clean separation, testable in isolation, cli.py only gains ~15 lines for command registration
Chosen: B (new module)
Trade-off: Slightly more files. Mitigation: the module exports a single function `init_project()`, keeping the API surface minimal.
Reversal: If the module stays under 40 lines after implementation, inline it. Threshold: 40 lines.
```

### Decision 2: Template Copy Strategy

```
Decision: Use shutil.copytree for directory templates, shutil.copy2 for individual files, 
          with dirs_exist_ok=True (Python 3.8+)
Alternatives:
  A. Recursive file-by-file copy with manual mkdir — pro: more control, con: verbose, Python already handles dirs_exist_ok
  B. Shell out to cp -r — pro: one command, con: platform-dependent, no dry-run
  C. Python jinja2 template rendering — pro: variable substitution, con: new dependency, overkill for copying static templates
Chosen: A (shutil, pure stdlib)
Trade-off: No variable substitution in templates. Mitigation: config.yaml already generic; project-specific values (like project name) 
           can be set via `gingx-sdd config` later or edited manually.
Reversal: If users demand project-name-in-config, add a simple string.Template substitution step (stdlib, no new dep).
```

### Decision 3: Stack Detection Integration

```
Decision: Call detect_project_type() from the existing detector/scanner.py module, 
          map its recommended_profile to a user-friendly label, and display in summary.
          No changes to detector/scanner.py.
Alternatives:
  A. Re-implement detection logic in init_project.py — pro: self-contained, con: duplicates the already-working detector
  B. Extend detector to return richer metadata — pro: more info, con: touches detector, increases scope
Chosen: A (reuse detector as-is)
Trade-off: init_project.py depends on the detector module.
Reversal: If detector changes its return type, update the mapping in init_project.py (trivial).
```

### Decision 4: Force/Overwrite Behavior

```
Decision: If .gingx/ exists and --force is NOT set, exit with error code 1 and message. 
          If --force is set, delete .gingx/ (with shutil.rmtree) and proceed.
Alternatives:
  A. Merge with existing files — pro: preserves user changes, con: unpredictable, can break hooks
  B. Rename to .gingx.bak then proceed — pro: non-destructive, con: leaves stale artifacts
Chosen: B-but-delete (rmtree on --force)
Trade-off: Destructive. Mitigation: the error message when .gingx/ exists is clear: "Use --force to overwrite".
           Future: could add --backup flag that renames to .gingx.bak-<timestamp>.
Reversal: If users frequently want merge, switch to per-file overwrite with warnings for conflicts.
```

### Decision 5: Hook Executable Permissions

```
Decision: Use os.chmod(path, 0o755) on each hook file after copy. Platform-aware: 
          skip on Windows (os.chmod is a no-op there, but explicitly check os.name).
Alternatives:
  A. Rely on shutil.copy2 preserving source permissions — pro: simpler, con: templates may not have +x in git
  B. Shell out to chmod — pro: familiar, con: not portable
Chosen: A (os.chmod in Python)
Trade-off: None. os.chmod is stdlib and works everywhere.
```

### Decision 6: Template Path Resolution

```
Decision: Resolve templates relative to the gingx-sdd package installation path.
          Use importlib.resources or __file__ relative path to find gingx-sdd/templates/.
Alternatives:
  A. Hard-code path relative to cwd — pro: simple, con: fails when installed via pip
  B. Use pkg_resources — pro: standard, con: deprecated in favor of importlib.resources
  C. GINGX_TEMPLATES environment variable override — pro: flexible for development, con: extra surface
Chosen: B + C (importlib.resources as primary, GINGX_TEMPLATES env var as override)
Trade-off: importlib.resources requires Python 3.9+ for files() API, project already requires 3.11+.
Reversal: If importlib.resources proves fragile in editable installs, fall back to __file__ relative path.
```

### File Structure After Init

```
my-project/
├── .gingx/
│   ├── config.yaml              (copied from templates/.gingx/config.yaml)
│   ├── suites.yaml              (copied from templates/.gingx/suites.yaml)
│   ├── current_task.yaml        (generated: hdu_id: none, phase: none)
│   └── profiles/
│       ├── developer.profile.yaml
│       ├── fullstack.profile.yaml
│       ├── fullstack-go.profile.yaml
│       ├── fullstack-python-langgraph.profile.yaml
│       ├── minimal.profile.yaml
│       ├── react-nextjs.profile.yaml
│       └── team.profile.yaml
├── .claude/
│   ├── hooks/
│   │   ├── pre-tool-use.sh      (executable)
│   │   ├── stop.sh              (executable)
│   │   └── session-start.sh     (executable)
│   └── settings.local.json
├── openspec/
│   ├── AGENTS.md
│   └── changes/                 (empty directory)
├── .mcp.json
└── (existing project files untouched)
```

### Error Handling Matrix

| Case | Behavior |
|------|----------|
| `.gingx/` already exists, no `--force` | Print error, exit code 1 |
| `.gingx/` already exists, `--force` | rmtree `.gingx/`, `.claude/`, `openspec/` (only gingx-created dirs), proceed |
| `--dry-run` | Print all actions, write nothing, exit code 0 |
| Template directory not found | Print error with path tried, exit code 2 |
| `os.chmod` fails (permissions) | Warn but continue (non-fatal on restricted systems) |
| Detection fails (scanner raises) | Catch exception, use minimal profile, warn user |
| Project root is not writable | Fail early with clear permission error |
| `--stack` with unknown stack name | Print valid stack names, exit code 1 |

### Summary Output Format

```
  Gingx SDD initialized!

  Detected: Python, FastAPI (fullstack profile recommended)
  Stack override: none

  Created:
    .gingx/config.yaml
    .gingx/suites.yaml
    .gingx/current_task.yaml
    .gingx/profiles/ (7 profiles)
    .claude/hooks/pre-tool-use.sh
    .claude/hooks/stop.sh
    .claude/hooks/session-start.sh
    .claude/settings.local.json
    openspec/AGENTS.md
    openspec/changes/
    .mcp.json

  Next steps:
    gingx-sdd hdu create "Your first feature"
    gingx-sdd status
```

### What We Reuse

- **`gingx_sdd.detector.scanner.detect_project_type()`** -- existing, returns `ProjectType` with `recommended_profile`, `recommended_skills`, `languages`, `frameworks`
- **Templates at `gingx-sdd/templates/`** -- all 16 files already exist, no new templates needed
- **`shutil.copytree`** with `dirs_exist_ok=True` -- Python 3.8+ stdlib
- **`os.chmod`** -- stdlib, no-op safe on Windows
- **`typer`** -- already used by all CLI commands; `init` follows the same pattern as `spec`, `status`, etc.

### Changes to Existing Code

| File | Change | Lines |
|------|--------|-------|
| `gingx-sdd/gingx_sdd/init_project.py` | **New file**: `init_project()` with all scaffolding logic | ~85 |
| `gingx-sdd/gingx_sdd/cli.py` | Add `init` command registration (~15 lines) | +15 |
| No changes to templates, detector, or any other module | | 0 |

## Alternatives Considered

1. **Extend install.sh instead of Python CLI** -- More code in bash, platform-dependent. Rejected: the CLI goal is to replace ad-hoc bash with a typed, testable command.
2. **Cookiecutter-style project generation** -- Would require a templating engine and `.cookiecutter.json` config. Rejected: overkill. Users already have their project started; init adds Gingx harness on top.
3. **`gingx-sdd init --template <url>` for remote templates** -- Interesting but out of scope. Can be added later.
4. **Automatic `mnemo init` during scaffolding** -- Requires mnemo to be installed (not guaranteed). Rejected: init should not fail if mnemo is absent. User runs `mnemo init` separately.
5. **Prompt user for stack during init (interactive mode)** -- Adds complexity for marginal gain. Rejected: auto-detect with --stack override covers both use cases. Can add `--interactive` flag later.
