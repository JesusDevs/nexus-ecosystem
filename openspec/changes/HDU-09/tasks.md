# Tasks: `gingx-sdd init` -- Scaffolding Command

## Phase 1 -- Core Scaffolding Module (init_project.py)

- [ ] 1. Create `gingx-sdd/gingx_sdd/init_project.py` with function signature:
  ```python
  def init_project(dry_run: bool = False, force: bool = False, stack: Optional[str] = None) -> int:
  ```
- [ ] 2. Implement template path resolution: use `os.environ.get("GINGX_TEMPLATES")` as override, else resolve relative to `Path(__file__).parent.parent / "templates"` (the package-relative path to `gingx-sdd/templates/`). Return error if templates dir not found.
- [ ] 3. Implement `.gingx/` pre-check: if `.gingx/` exists and `force=False`, print error "`.gingx/ already exists. Use --force to overwrite.`" and return 1. If `force=True`, run `shutil.rmtree(".gingx/")` and also remove `.claude/`, `openspec/`, `.mcp.json` (only files that init would create).
- [ ] 4. Implement `.gingx/` scaffold: `shutil.copytree` from `templates/.gingx/` to `./gingx/` with `dirs_exist_ok=True`. This copies `config.yaml`, `suites.yaml`, and the `profiles/` directory with all 7 `.profile.yaml` files.
- [ ] 5. Create `.gingx/current_task.yaml` with content:
  ```yaml
  hdu_id: none
  phase: none
  agent: none
  ```
  (Not from template -- generated programmatically.)
- [ ] 6. Implement `.claude/` scaffold: create `.claude/hooks/` dir, copy the 3 hook `.sh` files from `templates/.claude/hooks/`, and set `os.chmod(path, 0o755)` on each. Copy `templates/.claude/settings.local.json` to `.claude/settings.local.json`.
- [ ] 7. Implement OpenSpec scaffold: copy `templates/openspec/AGENTS.md` to `openspec/AGENTS.md`, create empty `openspec/changes/` directory.
- [ ] 8. Implement `.mcp.json` copy: copy `templates/.mcp.json` to project root.
- [ ] 9. Implement `--dry-run` guard: wrap all filesystem operations (copy, mkdir, chmod, rmtree) in an `if not dry_run:` block. In dry-run mode, build a list of `(action, path)` tuples and print them instead.

## Phase 2 -- Stack Detection Integration

- [ ] 10. Import and call `gingx_sdd.detector.scanner.detect_project_type()`. Handle the case where the scanner raises an exception (catch, warn, fall back to `type="unknown"` with `minimal` profile).
- [ ] 11. Map the `--stack` override: if `--stack go` is passed, bypass detection and force `recommended_profile="fullstack-go"`. Valid stack names: `python`, `go`, `react`, `node`, `langgraph`, `minimal`. Print error for unknown stack names.
- [ ] 12. Store detected metadata for the summary: `languages`, `frameworks`, `recommended_profile`, `type`.

## Phase 3 -- CLI Registration (cli.py)

- [ ] 13. Add `init` command to `cli.py` `app` using typer decorator pattern (matching existing commands like `spec`, `status`):
  ```python
  @app.command()
  def init(
      dry_run: bool = typer.Option(False, "--dry-run", help="Preview without writing files"),
      force: bool = typer.Option(False, "--force", help="Overwrite existing .gingx/"),
      stack: Optional[str] = typer.Option(None, "--stack", help="Force a specific stack (python, go, react, node, langgraph, minimal)"),
  ):
      """Initialize a Gingx SDD project -- scaffolds .gingx/, hooks, profiles, OpenSpec."""
      from gingx_sdd.init_project import init_project
      exit_code = init_project(dry_run=dry_run, force=force, stack=stack)
      raise typer.Exit(exit_code)
  ```
- [ ] 14. Register the import at top of `init_project` call (the `from gingx_sdd.init_project import init_project` is a deferred import to avoid circular deps -- matches existing patterns in cli.py like `from gingx_sdd.orchestrate import ...`).

## Phase 4 -- Summary Output

- [ ] 15. Implement summary table: after all operations complete (or in dry-run preview), print a formatted summary showing:
  - Detected stack info (languages, frameworks, recommended profile)
  - List of created files/directories with checkmarks
  - Next steps (commands to run)
  - Dry-run indicator if applicable
- [ ] 16. Edge case handling in summary: if 0 templates were copied (templates dir empty/broken), show warning. If stack detection failed, show fallback notice.

## Phase 5 -- Integration Verification

- [ ] 17. Manual test: `gingx-sdd init` in empty dir -- verify `.gingx/`, `.claude/`, `openspec/`, `.mcp.json` all created.
- [ ] 18. Manual test: `gingx-sdd init` in dir with `pyproject.toml` -- verify Python/fullstack detected.
- [ ] 19. Manual test: `gingx-sdd init` in dir with `go.mod` -- verify Go/fullstack-go detected.
- [ ] 20. Manual test: `gingx-sdd init --dry-run` -- verify no files written, preview shown.
- [ ] 21. Manual test: `gingx-sdd init` twice -- verify second run errors with "already exists" message.
- [ ] 22. Manual test: `gingx-sdd init --force` in existing gingx project -- verify overwrite works.
- [ ] 23. Manual test: `gingx-sdd init --stack go` in Python project -- verify stack override takes effect.
- [ ] 24. Manual test: Hook scripts are executable (`ls -la .claude/hooks/`).
- [ ] 25. Manual test: `gingx-sdd status` works after init (should show no active HDUs, harness active).
