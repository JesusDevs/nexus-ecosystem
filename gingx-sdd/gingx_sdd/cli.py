"""
Gingx-SDD CLI — Minimal harness.
Delegates to mnemo (Go) for all runtime logic.
Skills, specs, and personas are pure markdown.
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional

import typer
import yaml

app = typer.Typer(
    name="gingx-sdd",
    help="SDD Framework — Spec → Plan → Code → Test → Security → Memory",
    no_args_is_help=True,
)


# ── Init Command ──────────────────────────────────────────────────────

@app.command()
def init(
    dry_run: bool = typer.Option(False, "--dry-run", help="Preview without writing files"),
    force: bool = typer.Option(False, "--force", help="Overwrite existing .gingx/"),
    stack: Optional[str] = typer.Option(None, "--stack", help="Force a specific stack (python, go, react, node, langgraph, minimal)"),
):
    """Initialize a Gingx SDD project — scaffolds .gingx/, hooks, profiles, OpenSpec."""
    from gingx_sdd.init_project import init_project
    exit_code = init_project(dry_run=dry_run, force=force, stack=stack)
    raise typer.Exit(exit_code)


# ── Spec Command ─────────────────────────────────────────────────────

@app.command()
def spec(
    title: str = typer.Argument(..., help="Feature title"),
    hdu_id: Optional[str] = typer.Option(None, "--id", help="HDU ID (auto-generated)"),
):
    """Create an OpenSpec specification for a new feature."""
    hdu_id = hdu_id or f"HDU-{title.lower().replace(' ', '-')[:40]}"
    spec_path = Path.cwd() / "openspec" / "changes" / hdu_id

    spec_path.mkdir(parents=True, exist_ok=True)
    (spec_path / "specs").mkdir(exist_ok=True)

    (spec_path / "proposal.md").write_text(f"""\
# Proposal: {title}

## Why
[Describe why this change is needed]

## What Changes
- [Change 1]
- [Change 2]

## What Does NOT Change
-

## Impact
- HDU: {hdu_id}
- Complexity: [low/medium/high]
""")

    (spec_path / "specs" / f"{hdu_id}.md").write_text(f"""\
# Spec: {title}

## BDD Scenarios

### Scenario 1: Happy Path
```gherkin
Given [precondition]
When [action]
Then [expected result]
```

### Scenario 2: Error Case
```gherkin
Given [precondition]
When [invalid action]
Then [error response]
```
""")

    (spec_path / "design.md").write_text(f"""\
# Design: {title}

## Approach
[Describe the technical approach]

## Alternatives Considered
1. [Alternative A] — [Trade-off]
2. [Alternative B] — [Trade-off]

## Decision
[Final decision and rationale]
""")

    (spec_path / "tasks.md").write_text(f"""\
# Tasks: {title}

- [ ] 1. [Task 1]
- [ ] 2. [Task 2]
- [ ] 3. Write tests (BDD + unit)
- [ ] 4. Security scan
""")

    typer.echo(f"Spec created: {spec_path}")
    typer.echo(f"Next: gingx-sdd orchestrate {hdu_id}")


# ── Orchestrate Command ──────────────────────────────────────────────

@app.command()
def orchestrate(
    hdu_id: str = typer.Argument(..., help="HDU to orchestrate"),
    phase: Optional[str] = typer.Option(None, "--phase", help="Execute a specific phase"),
    agent: Optional[str] = typer.Option(None, "--agent", help="Invoke a specific agent persona"),
    status: bool = typer.Option(False, "--status", help="Show orchestration progress"),
):
    """Orchestrate an HDU delegating phases to specialized agents (PO, UX, Architect, Dev, QA, DevOps).

    Examples:
        gingx-sdd orchestrate HDU-06 --status
        gingx-sdd orchestrate HDU-06 --phase apply
        gingx-sdd orchestrate HDU-06 --agent qa
    """
    from gingx_sdd.orchestrate import decompose_hdu, suggest_agent, show_status, run_phase

    if status:
        show_status(hdu_id)
    elif phase:
        run_phase(hdu_id, phase)
    elif agent:
        agent_map = {
            "supervisor": "supervisor", "po": "po-agent", "ux": "ux-agent",
            "architect": "architect-agent", "arch": "architect-agent",
            "dev": "dev-agent", "qa": "qa-agent", "devops": "devops-agent",
        }
        full_agent = agent_map.get(agent.lower(), agent)
        role = {
            "supervisor": "SDD Orchestrator", "po-agent": "Product Owner",
            "ux-agent": "UX Designer", "architect-agent": "Solution Architect",
            "dev-agent": "Developer", "qa-agent": "QA Engineer",
            "devops-agent": "DevOps",
        }.get(full_agent, "Specialist")

        typer.echo(f"Agent: {full_agent} ({role})")
        typer.echo(f"HDU: {hdu_id}")

        skill_path = Path(__file__).parent.parent / "skills" / "team" / f"{full_agent}.md"
        if skill_path.exists():
            typer.echo(f"Skill: {skill_path}")
            typer.echo(f"Invoke: /{full_agent} <your task>")
        else:
            typer.echo(f"Skill not found: {full_agent}.md")
    else:
        phases = decompose_hdu(hdu_id)
        if not phases:
            typer.echo(f"HDU {hdu_id} not found or all phases complete.")
            return

        typer.echo(f"\nPhase decomposition for {hdu_id}:")
        for p in phases:
            typer.echo(f"  {p['phase']:12s} → {p['agent']:20s} [{p['status']}]")

        typer.echo(f"\nNext: gingx-sdd orchestrate {hdu_id} --phase {phases[0]['phase']}")


# ── Save Command ─────────────────────────────────────────────────────

@app.command()
def save(
    hdu_id: str = typer.Option(..., "--hdu-id", help="HDU to save as memory"),
    outcome: Optional[str] = typer.Option(None, "--outcome", help="resolved/partial/pending"),
    tags: Optional[str] = typer.Option(None, "--tags", help="Extra tags"),
):
    """Save HDU knowledge to mnemo vector memory."""
    spec_path = Path.cwd() / "openspec" / "changes" / hdu_id
    if not spec_path.exists():
        typer.echo(f"HDU not found: {hdu_id}", err=True)
        raise typer.Exit(1)

    proposal = (spec_path / "proposal.md").read_text() if (spec_path / "proposal.md").exists() else ""
    design = (spec_path / "design.md").read_text() if (spec_path / "design.md").exists() else ""

    title = hdu_id
    for line in proposal.split("\n"):
        if line.startswith("# Proposal:"):
            title = line.replace("# Proposal:", "").strip()
            break

    project = Path.cwd().name
    content = f"Proposal: {proposal[:500]}\n\nDesign: {design[:500]}"

    cmd = ["mnemo", "save", title, content, "--type", "decision",
           "--outcome", outcome or "noted", "--project", project]
    if tags:
        cmd.extend(["--tags", tags])

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        typer.echo(result.stdout.strip() if result.returncode == 0 else result.stderr.strip())
    except FileNotFoundError:
        typer.echo("mnemo CLI not found. Install: brew install gingx-sdd/tap/mnemo", err=True)


# ── Status Command ───────────────────────────────────────────────────

@app.command()
def status():
    """Show project status — active HDUs, harness health, and sensor status."""
    openspec_dir = Path.cwd() / "openspec" / "changes"
    if not openspec_dir.exists():
        typer.echo("No changes in progress.")
        return

    hdus = []
    for hdu_dir in sorted(openspec_dir.iterdir()):
        if hdu_dir.is_dir() and not hdu_dir.name.startswith("archive"):
            tasks_path = hdu_dir / "tasks.md"
            completed = 0
            total = 0
            if tasks_path.exists():
                content = tasks_path.read_text()
                completed = content.count("[x]")
                total = max(content.count("[ ]") + completed, 1)

            phase = "spec" if not (hdu_dir / "design.md").exists() else \
                    "design" if not (hdu_dir / "tasks.md").exists() else \
                    "apply" if completed == 0 else \
                    "verify" if completed < total else "archive"

            hdus.append((hdu_dir.name, phase, completed, total))

    if not hdus:
        typer.echo("No active HDUs.")
    else:
        typer.echo(f"\n{'HDU':<25} {'Phase':<12} {'Progress':<10}")
        typer.echo("-" * 47)
        for name, phase, done, tot in hdus:
            pct = f"{int(done/tot*100)}%" if tot > 0 else "0%"
            typer.echo(f"{name:<25} {phase:<12} {pct:<10}")

    # ── Harness Health ──────────────────────────────────────────
    typer.echo(f"\n{'Sensor':<30} {'Status':<12} {'Type':<15}")
    typer.echo("-" * 57)

    # Feedforward guides
    skills_dir = Path.cwd() / "gingx-sdd" / "skills" / "team"
    skills = list(skills_dir.glob("*.md")) if skills_dir.exists() else []
    typer.echo(f"{'Feedforward guides (skills)':<30} {f'{len(skills)} loaded':<12} {'guide':<15}")
    if skills:
        for s in sorted(skills):
            typer.echo(f"  └─ /{s.stem:<25}")

    # Agents
    agents_dir = Path.cwd() / ".claude" / "skills"
    agents = list(agents_dir.glob("*.md")) if agents_dir.exists() else []
    typer.echo(f"{'Agent personas':<30} {f'{len(agents)} registered':<12} {'guide':<15}")

    # Computational sensors
    gitleaks_ok = bool(list(Path.cwd().glob("**/gitleaks*"))) or _check_cmd_available("gitleaks")
    typer.echo(f"{'Security scan (gitleaks)':<30} {f'{"active" if gitleaks_ok else "not found"}':<12} {'computational':<15}")

    # Build check
    go_build_ok = _check_cmd_available("go")
    typer.echo(f"{'Build check (go)':<30} {f'{"available" if go_build_ok else "N/A"}':<12} {'computational':<15}")

    # Dependency audit
    pip_audit_ok = _check_cmd_available("pip-audit")
    go_vuln_ok = _check_cmd_available("govulncheck")
    dep_audit_status = "active" if pip_audit_ok or go_vuln_ok else "not found"
    typer.echo(f"{'Dependency audit':<30} {dep_audit_status:<12} {'computational':<15}")

    # Inferential sensors
    try:
        subprocess.run(["mnemo", "conflicts", "--help"], capture_output=True, timeout=5)
        mnemo_conf_ok = True
    except Exception:
        mnemo_conf_ok = False
    typer.echo(f"{'Mnemo conflict detection':<30} {f'{"available" if mnemo_conf_ok else "no mnemo"}':<12} {'inferential':<15}")

    # Steering
    hooks_config = Path.cwd() / ".claude" / "settings.local.json"
    hooks_active = hooks_config.exists()
    typer.echo(f"{'Steering (hooks)':<30} {f'{"active" if hooks_active else "not configured"}':<12} {'steering':<15}")

    # Swarm mode
    try:
        mode_out = subprocess.run(["mnemo", "config"], capture_output=True, text=True, timeout=10)
        for line in mode_out.stdout.split("\n"):
            if "swarm.mode" in line:
                mode = line.split("=", 1)[1].strip() if "=" in line else "hybrid"
                typer.echo(f"{'Swarm mode':<30} {mode:<12} {'orchestration':<15}")
                break
    except Exception:
        pass


def _check_cmd_available(cmd: str) -> bool:
    """Check if a command is available on PATH."""
    try:
        result = subprocess.run(["which", cmd], capture_output=True, text=True)
        return result.returncode == 0
    except Exception:
        return False


# ── Release Command ──────────────────────────────────────────────────

SEMVER_RE = re.compile(
    r"^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)"
    r"(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
)

PROJECTS = {
    "gingx-mnemo": {
        "type": "go",
        "path": "gingx-mnemo",
        "build_cmd": ["go", "build", "./..."],
        "test_cmd": ["go", "test", "./..."],
    },
    "gingx-sdd": {
        "type": "python",
        "path": "gingx-sdd",
        "build_cmd": [sys.executable, "-m", "py_compile", "gingx_sdd/cli.py"],
        "test_cmd": None,
    },
}


def _validate_semver(version: str) -> bool:
    """Validate semantic version format v<MAJOR>.<MINOR>.<PATCH>[-prerelease]."""
    return bool(SEMVER_RE.match(version))


def _run(cmd: list[str], cwd: str | None = None, timeout: int = 120) -> tuple[int, str, str]:
    """Run a command and return (returncode, stdout, stderr)."""
    try:
        result = subprocess.run(
            cmd, capture_output=True, text=True, timeout=timeout, cwd=cwd,
        )
        return result.returncode, result.stdout.strip(), result.stderr.strip()
    except FileNotFoundError:
        return -1, "", f"command not found: {cmd[0]}"
    except subprocess.TimeoutExpired:
        return -1, "", f"timed out after {timeout}s"


def _git_clean(project_path: str) -> tuple[bool, str]:
    """Check if the git working tree is clean."""
    code, out, err = _run(["git", "status", "--porcelain"], cwd=project_path)
    if code != 0:
        return False, f"git status failed: {err}"
    if out:
        return False, f"uncommitted changes found:\n{out}"
    return True, "working tree clean"


def _security_scan(project_path: str) -> tuple[bool, str]:
    """Run gitleaks detect if available, otherwise skip."""
    code, out, err = _run(["gitleaks", "detect", "--source", project_path, "--no-git", "--verbose"], timeout=60)
    if code == -1 and "command not found" in err:
        return True, "gitleaks not installed — skipping"
    if code != 0:
        return False, f"security scan found issues:\n{err or out}"
    return True, "security scan clean"


def _mnemo_conflicts(project: str) -> tuple[bool, str]:
    """Check for semantic conflicts via mnemo."""
    try:
        result = subprocess.run(
            ["mnemo", "conflicts", "--project", project],
            capture_output=True, text=True, timeout=30,
        )
        if result.returncode != 0:
            return True, "no conflicts detected"
        if result.stdout.strip():
            return False, f"mnemo conflicts found:\n{result.stdout.strip()[:500]}"
        return True, "no conflicts detected"
    except FileNotFoundError:
        return True, "mnemo CLI not found — skipping conflict check"
    except Exception as e:
        return True, f"mnemo conflict check skipped: {e}"


def _pre_release_checks(project_name: str, project_config: dict) -> tuple[bool, list[str]]:
    """Run all pre-release checks. Returns (passed, [results])."""
    project_path = str(Path.cwd() / project_config["path"])
    results = []
    all_ok = True

    # 1. Git cleanliness
    ok, msg = _git_clean(project_path)
    results.append(f"{'✓' if ok else '✗'} Git clean: {msg}")
    all_ok = all_ok and ok

    # 2. Security scan
    ok, msg = _security_scan(project_path)
    results.append(f"{'✓' if ok else '✗'} Security: {msg}")
    all_ok = all_ok and ok

    # 3. Build check
    if project_config.get("build_cmd"):
        code, out, err = _run(project_config["build_cmd"], cwd=project_path)
        ok = code == 0
        results.append(f"{'✓' if ok else '✗'} Build: {out or err}")
        all_ok = all_ok and ok

    # 4. Test check
    if project_config.get("test_cmd"):
        code, out, err = _run(project_config["test_cmd"], cwd=project_path)
        ok = code == 0
        results.append(f"{'✓' if ok else '✗'} Tests: {out or err}")
        all_ok = all_ok and ok

    # 5. Mnemo conflict detection
    ok, msg = _mnemo_conflicts(project_name)
    results.append(f"{'✓' if ok else '✗'} Mnemo conflicts: {msg}")
    all_ok = all_ok and ok

    return all_ok, results


def _generate_changelog(project_name: str, version: str, project_path: str) -> str:
    """Generate changelog entry from git commits since last tag."""
    # Get last tag
    code, last_tag, _ = _run(
        ["git", "describe", "--tags", "--abbrev=0"], cwd=project_path,
    )
    last_tag = last_tag if code == 0 else ""

    if last_tag:
        code, log, _ = _run(
            ["git", "log", f"{last_tag}..HEAD", "--pretty=format:%s", "--no-merges"],
            cwd=project_path,
        )
    else:
        code, log, _ = _run(
            ["git", "log", "--pretty=format:%s", "--no-merges"],
            cwd=project_path,
        )

    if code != 0 or not log:
        return "No changes to report."

    # Categorize commits
    features, fixes, improvements, breaking = [], [], [], []

    for line in log.split("\n"):
        line = line.strip()
        if not line:
            continue
        lower = line.lower()
        if lower.startswith(("feat", "add", "new", "create")):
            features.append(f"- {line}")
        elif lower.startswith(("fix", "bug", "resolve", "patch", "correct")):
            fixes.append(f"- {line}")
        elif lower.startswith(("break", "major", "!")):
            breaking.append(f"- {line}")
        else:
            improvements.append(f"- {line}")

    sections = []
    if breaking:
        sections.append(f"## Breaking Changes\n\n" + "\n".join(breaking))
    if features:
        sections.append(f"## Features\n\n" + "\n".join(features))
    if improvements:
        sections.append(f"## Improvements\n\n" + "\n".join(improvements))
    if fixes:
        sections.append(f"## Fixes\n\n" + "\n".join(fixes))

    if not sections:
        return "No significant changes to report."

    header = f"## {version} ({datetime.now().strftime('%Y-%m-%d')})\n"
    commit_range = f"{last_tag}...{version}" if last_tag else f"initial...{version}"
    header += f"_{project_name} commit range: {commit_range}_\n"

    return header + "\n" + "\n\n".join(sections)


def _do_release(project_name: str, project_config: dict, version: str, dry_run: bool) -> tuple[bool, str]:
    """Execute the release for a single project. Returns (success, message)."""
    project_path = str(Path.cwd() / project_config["path"])
    messages = []

    # Pre-release checks
    checks_ok, check_results = _pre_release_checks(project_name, project_config)
    messages.extend(check_results)

    if dry_run:
        messages.append("\n--dry-run: no tag or snapshot created.")
        return checks_ok, "\n".join(messages)

    if not checks_ok:
        return False, "\n".join(messages) + "\n\nRelease blocked — fix issues above and retry."

    # Create git tag
    tag_msg = f"Release {version}"
    code, out, err = _run(
        ["git", "tag", "-a", version, "-m", tag_msg],
        cwd=project_path,
    )
    if code != 0:
        return False, "\n".join(messages) + f"\n\nGit tag failed: {err}"

    messages.append(f"✓ Git tag: {version}")

    # Mnemo release snapshot
    try:
        result = subprocess.run(
            ["mnemo", "release", project_name, version],
            capture_output=True, text=True, timeout=30,
        )
        if result.returncode == 0:
            messages.append(f"✓ Mnemo snapshot: {result.stdout.strip()}")
        else:
            messages.append(f"✗ Mnemo snapshot failed: {result.stderr.strip()}")
    except FileNotFoundError:
        messages.append("! Mnemo CLI not found — skipping snapshot")
    except Exception as e:
        messages.append(f"! Mnemo snapshot skipped: {e}")

    # Generate changelog entry
    changelog_entry = _generate_changelog(project_name, version, project_path)
    changelog_path = Path.cwd() / "CHANGELOG.md"

    existing = changelog_path.read_text() if changelog_path.exists() else ""
    new_content = changelog_entry + "\n\n" + existing if existing else changelog_entry
    changelog_path.write_text(new_content)
    messages.append(f"✓ Changelog updated: CHANGELOG.md")

    # Record version in mnemo config
    try:
        subprocess.run(
            ["mnemo", "config", "set", "mnemo.version", version],
            capture_output=True, text=True, timeout=10,
        )
        messages.append(f"✓ Version recorded in mnemo config")
    except Exception:
        pass

    return True, "\n".join(messages)


@app.command()
def release(
    version: str = typer.Argument(..., help="Semantic version (e.g., v0.1.0)"),
    project: Optional[str] = typer.Option(None, "--project", help="Target project (gingx-mnemo or gingx-sdd)"),
    all_projects: bool = typer.Option(False, "--all", help="Release all projects in monorepo"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Run pre-release checks without creating tags"),
):
    """Create a semantic release with git tag, mnemo snapshot, and changelog.

    Examples:
        gingx-sdd release v0.1.0 --project gingx-mnemo
        gingx-sdd release v0.1.0 --all
        gingx-sdd release v0.1.0 --all --dry-run
    """
    if not _validate_semver(version):
        typer.echo(f"Invalid semver: {version}. Use format v<MAJOR>.<MINOR>.<PATCH>", err=True)
        raise typer.Exit(1)

    targets = []

    if all_projects:
        targets = [(name, cfg) for name, cfg in PROJECTS.items()
                    if (Path.cwd() / cfg["path"]).exists()]
        if not targets:
            typer.echo("No projects found in monorepo.", err=True)
            raise typer.Exit(1)
    elif project:
        if project not in PROJECTS:
            typer.echo(f"Unknown project: {project}. Known: {', '.join(PROJECTS)}", err=True)
            raise typer.Exit(1)
        proj_path = Path.cwd() / PROJECTS[project]["path"]
        if not proj_path.exists():
            typer.echo(f"Project path not found: {proj_path}", err=True)
            raise typer.Exit(1)
        targets = [(project, PROJECTS[project])]
    else:
        typer.echo("Specify --project <name> or --all", err=True)
        raise typer.Exit(1)

    typer.echo(f"\n{'DRY RUN' if dry_run else 'RELEASE'} {version}")
    typer.echo(f"Projects: {', '.join(name for name, _ in targets)}")
    typer.echo("=" * 50)

    all_ok = True
    for proj_name, proj_cfg in targets:
        typer.echo(f"\n── {proj_name} ──")
        ok, msg = _do_release(proj_name, proj_cfg, version, dry_run)
        typer.echo(msg)
        all_ok = all_ok and ok

    if not all_ok and not dry_run:
        typer.echo("\nSome releases failed. Check output above.", err=True)
        raise typer.Exit(1)

    if dry_run:
        typer.echo("\nDry run complete. Re-run without --dry-run to execute.")
    else:
        typer.echo(f"\nRelease {version} complete.")


# ── HDU Command Group ──────────────────────────────────────────────────

hdu_app = typer.Typer(
    help="HDU Tracking — create, status, summary, block, unblock",
    no_args_is_help=True,
)
app.add_typer(hdu_app, name="hdu")


@hdu_app.command("create")
def hdu_create(
    title: str = typer.Argument(..., help="Feature title"),
    question: Optional[str] = typer.Option(None, "--question", "-q", help="Blocking question that must be answered"),
    context: Optional[str] = typer.Option(None, "--context", "-c", help="Extra context for the question"),
    hdu_id: Optional[str] = typer.Option(None, "--id", help="HDU ID (auto-generated)"),
):
    """Create a new HDU with an optional blocking question."""
    project_root = _find_project_root()
    hdu_id = hdu_id or _generate_hdu_id(title)

    # Create openspec directory
    spec_path = project_root / "openspec" / "changes" / hdu_id
    spec_path.mkdir(parents=True, exist_ok=True)
    (spec_path / "specs").mkdir(exist_ok=True)

    # Create spec artifacts
    (spec_path / "proposal.md").write_text(f"# Proposal: {title}\n\n## Why\n\n\n## What Changes\n- \n\n## Impact\n- HDU: {hdu_id}\n")
    (spec_path / "specs" / f"{hdu_id}.md").write_text(f"# Spec: {title}\n\n## BDD Scenarios\n\n### Scenario 1: Happy Path\n```gherkin\nGiven \nWhen \nThen \n```\n")
    (spec_path / "design.md").write_text(f"# Design: {title}\n\n## Approach\n\n\n## Decision\n")
    (spec_path / "tasks.md").write_text(f"# Tasks: {title}\n\n- [ ] 1. \n- [ ] 2. Write tests\n")

    # Create HDU tracking entry
    from .hdu_tracker import HDUStore
    store = HDUStore(project_root / ".gingx")
    entry = store.create(title=title, hdu_id=hdu_id, question=question)

    # Set as active
    (project_root / ".gingx" / "current_task.yaml").write_text(
        f"hdu_id: {hdu_id}\nphase: init\nstarted: {entry.started_at}\ndescription: {title}\n"
    )

    typer.echo(f"HDU created: {hdu_id}")
    if question:
        typer.echo(f"  Blocker: {question}")
        typer.echo(f"  Resolve: gingx-sdd hdu unblock {hdu_id} --answer \"<your answer>\"")
    typer.echo(f"  Status:  gingx-sdd hdu status {hdu_id}")


@hdu_app.command("status")
def hdu_status(
    hdu_id: Optional[str] = typer.Argument(None, help="HDU ID (omit for active HDU)"),
):
    """Show detailed status of an HDU including blockers and progress."""
    project_root = _find_project_root()
    from .hdu_tracker import HDUStore
    store = HDUStore(project_root / ".gingx")

    if hdu_id:
        entry = store.get(hdu_id)
        if not entry:
            typer.echo(f"HDU '{hdu_id}' not found.", err=True)
            raise typer.Exit(1)
    else:
        active_path = project_root / ".gingx" / "current_task.yaml"
        if active_path.exists():
            data = yaml.safe_load(active_path.read_text())
            if isinstance(data, dict):
                hdu_id = data.get("hdu_id")
            else:
                for line in active_path.read_text().split("\n"):
                    if line.startswith("hdu_id:"):
                        hdu_id = line.split(":", 1)[1].strip()
                        break
        if not hdu_id:
            typer.echo("No active HDU. Use 'gingx-sdd hdu summary' to list all.", err=True)
            raise typer.Exit(1)
        entry = store.get(hdu_id)
        if not entry:
            typer.echo(f"Active HDU '{hdu_id}' not found in store.", err=True)
            raise typer.Exit(1)

    typer.echo(f"\n{'─' * 50}")
    typer.echo(f"  HDU: {entry.id}")
    typer.echo(f"  Title: {entry.title}")
    typer.echo(f"  Status: {entry.status.upper()}")
    typer.echo(f"  Phase: {entry.phase}")
    typer.echo(f"  Progress: {entry.progress}%")
    typer.echo(f"  Started: {entry.started_at}")
    typer.echo(f"  Updated: {entry.updated_at}")

    if entry.blockers:
        typer.echo(f"\n  Blockers ({len(entry.open_blockers())} open):")
        for i, b in enumerate(entry.blockers):
            state = "ANSWERED" if b.is_resolved() else "OPEN"
            icon = "✓" if b.is_resolved() else "▸"
            typer.echo(f"    {icon} [{state}] {b.question}")
            if b.answer:
                typer.echo(f"       Answer: {b.answer}")
            elif b.context:
                typer.echo(f"       Context: {b.context}")

    if entry.artifacts_produced:
        typer.echo(f"\n  Artifacts: {', '.join(entry.artifacts_produced)}")

    if entry.risks_open:
        typer.echo(f"\n  Risks: {', '.join(entry.risks_open)}")

    if entry.executive_summary:
        typer.echo(f"\n  Summary: {entry.executive_summary}")

    if entry.next_recommended:
        typer.echo(f"\n  Next: {entry.next_recommended}")

    typer.echo(f"{'─' * 50}\n")


@hdu_app.command("summary")
def hdu_summary():
    """Show a summary table of all HDUs."""
    project_root = _find_project_root()
    from .hdu_tracker import HDUStore
    store = HDUStore(project_root / ".gingx")
    entries = store.load_all()

    if not entries:
        typer.echo("No HDUs found.")
        return

    typer.echo(f"\n{'ID':<25} {'Title':<25} {'Phase':<10} {'Progress':<10} {'Blockers':<10} {'Updated':<12}")
    typer.echo("-" * 92)
    for e in entries:
        open_count = len(e.open_blockers())
        blocker_str = f"{open_count} open" if open_count else "clear"
        pct = f"{e.progress}%"
        updated = e.updated_at[:10] if e.updated_at else "-"
        typer.echo(f"{e.id:<25} {e.title[:24]:<25} {e.phase:<10} {pct:<10} {blocker_str:<10} {updated:<12}")
    typer.echo()


@hdu_app.command("block")
def hdu_block(
    hdu_id: str = typer.Argument(..., help="HDU ID to block"),
    question: str = typer.Option(..., "--question", "-q", help="Blocking question"),
    context: Optional[str] = typer.Option(None, "--context", "-c", help="Extra context"),
    agent: str = typer.Option("supervisor", "--agent", "-a", help="Agent raising the question"),
):
    """Add a blocking question to an HDU."""
    project_root = _find_project_root()
    from .hdu_tracker import HDUStore
    store = HDUStore(project_root / ".gingx")
    entry = store.add_blocker(hdu_id, question=question, context=context or "", blocked_by=agent)
    if not entry:
        typer.echo(f"HDU '{hdu_id}' not found.", err=True)
        raise typer.Exit(1)
    typer.echo(f"Blocker added to {hdu_id}: {question}")
    typer.echo(f"Resolve: gingx-sdd hdu unblock {hdu_id} --answer \"<answer>\"")


@hdu_app.command("unblock")
def hdu_unblock(
    hdu_id: str = typer.Argument(..., help="HDU ID to unblock"),
    answer: str = typer.Option(..., "--answer", "-a", help="Answer to the blocking question"),
):
    """Answer the oldest open blocking question and unblock the HDU."""
    project_root = _find_project_root()
    from .hdu_tracker import HDUStore
    store = HDUStore(project_root / ".gingx")
    entry = store.answer_blocker(hdu_id, answer=answer)
    if not entry:
        typer.echo(f"HDU '{hdu_id}' not found.", err=True)
        raise typer.Exit(1)
    if entry.is_blocked():
        typer.echo(f"Blocker answered for {hdu_id}. {len(entry.open_blockers())} remaining.")
    else:
        typer.echo(f"All blockers resolved for {hdu_id}. Status: active.")


def _generate_hdu_id(title: str) -> str:
    slug = title.lower().replace(" ", "-")[:40].strip("-")
    return f"HDU-{slug}"


# ── Mode Command Group ─────────────────────────────────────────────────

mode_app = typer.Typer(
    help="Mode control — set execution mode or disable harness",
    no_args_is_help=True,
)
app.add_typer(mode_app, name="mode")


@mode_app.command("set")
def mode_set(
    mode: str = typer.Argument(..., help="interactive | automatic | dry_run | off"),
):
    """Set the Gingx execution mode.

    interactive — explore, propose, pause for human decisions
    automatic  — advance phase by phase with fewer interruptions
    dry_run    — pre-release checks only, no real execution
    off        — disable harness completely (no spec gate, no hooks)
    """
    project_root = _find_project_root()
    from .modes import GingxMode, set_mode
    try:
        gingx_mode = GingxMode(mode.lower())
    except ValueError:
        valid = ", ".join(m.value for m in GingxMode)
        typer.echo(f"Invalid mode: {mode}. Valid: {valid}", err=True)
        raise typer.Exit(1)
    set_mode(gingx_mode, project_root / ".gingx")
    typer.echo(f"Mode set to: {gingx_mode.value}")


@mode_app.command("status")
def mode_status():
    """Show current mode and harness health."""
    project_root = _find_project_root()
    from .modes import get_mode, is_harness_active
    current = get_mode(project_root / ".gingx")
    active = is_harness_active(project_root / ".gingx")

    typer.echo(f"Mode: {current.value}")
    typer.echo(f"Harness active: {'yes' if active else 'no (hooks disabled, no spec gate)'}")

    hooks_config = project_root / ".claude" / "settings.local.json"
    typer.echo(f"Hooks configured: {'yes' if hooks_config.exists() else 'no'}")


# ── Knowledge Command Group ───────────────────────────────────────────

knowledge_app = typer.Typer(
    help="Cross-project knowledge — export/import codegraph to mnemo",
    no_args_is_help=True,
)
app.add_typer(knowledge_app, name="knowledge")


@knowledge_app.command("export")
def knowledge_export(
    hdu_id: Optional[str] = typer.Option(None, "--hdu-id", help="HDU context for the export"),
    max_symbols: int = typer.Option(50, "--max", "-m", help="Max symbols to export"),
):
    """Export codegraph symbols to mnemo for cross-project memory."""
    project_root = _find_project_root()
    from .knowledge_bridge import export_codegraph_to_mnemo

    result = export_codegraph_to_mnemo(
        project_path=str(project_root),
        project_name=project_root.name,
        hdu_id=hdu_id,
        max_symbols=max_symbols,
    )

    typer.echo(f"Status: {result['status']}")
    typer.echo(f"Message: {result['message']}")
    if result['count'] > 0:
        typer.echo(f"Symbols exported: {result['count']}")


@knowledge_app.command("import")
def knowledge_import(
    project: str = typer.Argument(..., help="Project name or keywords to search"),
    limit: int = typer.Option(10, "--limit", "-l", help="Max results"),
):
    """Search mnemo for code structures from other projects."""
    from .knowledge_bridge import import_codegraph_from_mnemo

    results = import_codegraph_from_mnemo(project_name=project, query=project, limit=limit)
    if results:
        for r in results:
            typer.echo(r)
    else:
        typer.echo("No cross-project code structures found.")


@knowledge_app.command("status")
def knowledge_status():
    """Show codegraph + mnemo knowledge bridge status."""
    project_root = _find_project_root()
    from .knowledge_bridge import get_knowledge_status

    status = get_knowledge_status(str(project_root))
    typer.echo(f"CodeGraph installed: {'yes' if status['codegraph_installed'] else 'no'}")
    typer.echo(f"CodeGraph initialized: {'yes' if status['codegraph_initialized'] else 'no'}")
    if status.get("codegraph_stats"):
        typer.echo(f"\nCodeGraph stats:\n{status['codegraph_stats']}")
    typer.echo(f"Mnemo installed: {'yes' if status['mnemo_installed'] else 'no'}")


# ── Auto Command ─────────────────────────────────────────────────────

@app.command()
def auto(
    task: str = typer.Argument(..., help="Task description to classify and dispatch"),
    profile_name: str = typer.Option("developer", "--profile", "-p", help="Profile name"),
    hdu_id: Optional[str] = typer.Option(None, "--hdu-id", help="HDU context"),
    dispatch_flag: bool = typer.Option(False, "--dispatch", "-d", help="Actually spawn the agent (automatic mode)"),
):
    """Analyze a task and suggest/dispatch the right agent.

    Classifies the task (inline/delegate/full_sdd) and maps it to
    the correct agent persona from orchestrator config.

    Examples:
        gingx-sdd auto "Agregar endpoint de health check"
        gingx-sdd auto "Redisenar la capa de autenticacion" --dispatch
        gingx-sdd auto "Fix typo in README" --profile senior-dev
    """
    from .auto_delegate import analyze, dispatch, is_delegation_enabled, load_delegation_config
    from .modes import get_mode, GingxMode

    project_root = _find_project_root()
    config = load_delegation_config(project_root)

    if not is_delegation_enabled(config):
        typer.echo("Delegation is disabled in config.yaml. Enable it first.", err=True)
        raise typer.Exit(1)

    mode = get_mode(project_root / ".gingx")
    should_dispatch = dispatch_flag or mode == GingxMode.AUTOMATIC

    if should_dispatch:
        result = dispatch(
            task_description=task,
            mode="automatic" if mode == GingxMode.AUTOMATIC else "interactive",
            profile_name=profile_name,
            hdu_id=hdu_id,
            project_root=project_root,
        )
        typer.echo(result.prompt_preview)
        if result.dispatched:
            typer.echo(f"\nDispatched to {result.suggested_agent} ({result.classification})")
        else:
            typer.echo(f"\nSuggestion only — use --dispatch to spawn {result.suggested_agent}")
    else:
        result = analyze(task, project_root)
        typer.echo(result.prompt_preview)
        typer.echo(f"\nRun with --dispatch to spawn {result.suggested_agent}")
        typer.echo(f"Or: gingx-sdd team spawn {result.suggested_agent} -t \"{task[:80]}...\"")


# ── Team Command Group ────────────────────────────────────────────────

team_app = typer.Typer(
    help="Agent Factory — spawn, list, and manage sub-agents",
    no_args_is_help=True,
)
app.add_typer(team_app, name="team")


@team_app.command("spawn")
def team_spawn(
    agent: str = typer.Argument(..., help="Agent: supervisor, po-agent, ux-agent, architect-agent, dev-agent, qa-agent, devops-agent"),
    task: str = typer.Option(..., "--task", "-t", help="Task description"),
    profile_name: str = typer.Option("developer", "--profile", "-p", help="Profile name"),
    hdu_id: Optional[str] = typer.Option(None, "--hdu-id", help="HDU context"),
    mode: str = typer.Option("prompt", "--mode", "-m", help="prompt | mcp"),
    tech_stack_str: Optional[str] = typer.Option(None, "--stack", "-s", help="Comma-separated tech stack override"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show loaded skills and context"),
):
    """
    Spawn a sub-agent with persona + tech skills + mnemo context + interrogation.

    Generates the complete prompt for Claude Code's Agent tool.
    """
    from .skills.registry import SkillRegistry
    from .profile import load_profile, list_profiles
    from .prompt_builder import assemble_agent_prompt

    project_root = _find_project_root()
    profiles_dir = project_root / ".gingx" / "profiles"
    personas_dir = project_root / "gingx-sdd" / "skills" / "team"
    extras_dir = project_root / "gingx-sdd" / "extras" / "skills"

    # Load registry
    registry = SkillRegistry()
    registry.scan_personas(personas_dir)
    registry.scan_tech_stacks(extras_dir)

    # Load profile
    profile = load_profile(profile_name, profiles_dir)
    if not profile:
        available = ", ".join(list_profiles(profiles_dir))
        typer.echo(f"Profile '{profile_name}' not found. Available: {available}", err=True)
        raise typer.Exit(1)

    # Resolve tech stack override
    tech_override = None
    if tech_stack_str:
        tech_override = [s.strip() for s in tech_stack_str.split(",") if s.strip()]

    # Assemble prompt
    result = assemble_agent_prompt(
        agent_name=agent,
        task=task,
        profile=profile,
        registry=registry,
        hdu_id=hdu_id,
        tech_stack_override=tech_override,
        project=str(project_root.name),
        personas_dir=personas_dir,
    )

    typer.echo(result.system_prompt)

    if verbose:
        typer.echo("\n" + "=" * 60, err=True)
        typer.echo(f"Agent: {result.agent_name}", err=True)
        typer.echo(f"Profile: {result.profile_name}", err=True)
        typer.echo(f"Tech stacks: {result.tech_stacks_loaded}", err=True)
        typer.echo(f"Interrogation: {result.interrogation_mode}", err=True)
        typer.echo(f"Mnemo context: {'yes' if result.mnemo_context else 'no'}", err=True)


@team_app.command("list")
def team_list(
    profile_name: Optional[str] = typer.Option(None, "--profile", "-p", help="Show agents for specific profile"),
):
    """List all registered agents and their tech stacks."""
    from .skills.registry import SkillRegistry
    from .profile import load_profile, list_profiles

    project_root = _find_project_root()
    personas_dir = project_root / "gingx-sdd" / "skills" / "team"
    extras_dir = project_root / "gingx-sdd" / "extras" / "skills"

    registry = SkillRegistry()
    registry.scan_personas(personas_dir)
    registry.scan_tech_stacks(extras_dir)

    if profile_name:
        profiles_dir = project_root / ".gingx" / "profiles"
        profile = load_profile(profile_name, profiles_dir)
        if not profile:
            typer.echo(f"Profile '{profile_name}' not found.", err=True)
            raise typer.Exit(1)

        typer.echo(f"\nProfile: {profile.name} — {profile.description}")
        typer.echo(f"Default model: {profile.default_model}")
        typer.echo(f"Default stack: {profile.default_stack}")
        typer.echo(f"Interrogation: {profile.interrogation_depth}\n")
        for ag_name, ag_cfg in sorted(profile.agents.items()):
            agent_def = registry.get_agent(ag_name)
            desc = agent_def.description[:80] if agent_def else "(unknown)"
            stack = ag_cfg.tech_stack or profile.default_stack
            typer.echo(f"  {ag_name}: {ag_cfg.model}, {stack}")
            typer.echo(f"    {desc}")
    else:
        typer.echo(f"\n{len(registry.agents)} agents registered:\n")
        for name, agent_def in sorted(registry.agents.items()):
            typer.echo(f"  {name} ({agent_def.model}, {agent_def.effort})")
            typer.echo(f"    {agent_def.description[:100]}")

        typer.echo(f"\n{len(registry.tech_stacks)} tech stacks by category:")
        by_cat = registry.list_by_category()
        for cat in sorted(by_cat):
            typer.echo(f"  {cat}: {', '.join(sorted(by_cat[cat]))}")

        profiles_dir = project_root / ".gingx" / "profiles"
        profiles = list_profiles(profiles_dir)
        typer.echo(f"\n{len(profiles)} profiles: {', '.join(profiles)}")


@team_app.command("profile")
def team_profile(
    action: str = typer.Argument("show", help="show | set <name> | list"),
    name: Optional[str] = typer.Argument(None, help="Profile name for 'set' action"),
):
    """Show, set, or list active team profiles."""
    from .profile import load_profile, list_profiles, set_active_profile, get_active_profile

    project_root = _find_project_root()
    profiles_dir = project_root / ".gingx" / "profiles"

    if action == "list":
        profiles = list_profiles(profiles_dir)
        active = get_active_profile(project_root / ".gingx")
        typer.echo(f"Active: {active}")
        typer.echo(f"Available: {', '.join(profiles)}")

    elif action == "set":
        if not name:
            typer.echo("Usage: gingx-sdd team profile set <name>", err=True)
            raise typer.Exit(1)
        set_active_profile(name, project_root / ".gingx")
        typer.echo(f"Active profile set to: {name}")

    elif action == "show":
        active = get_active_profile(project_root / ".gingx")
        profile = load_profile(active, profiles_dir)
        if profile:
            typer.echo(f"Active profile: {active}")
            typer.echo(f"Description: {profile.description}")
            typer.echo(f"Model: {profile.default_model}")
            typer.echo(f"Stack: {profile.default_stack}")
            typer.echo(f"Interrogation: {profile.proactive_interrogation} ({profile.interrogation_depth})")
            typer.echo(f"\nAgents:")
            for ag_name, ag_cfg in sorted(profile.agents.items()):
                tech = ag_cfg.tech_stack or profile.default_stack
                typer.echo(f"  {ag_name}: {ag_cfg.model}, stack={tech}")
        else:
            typer.echo(f"Profile '{active}' not loaded.", err=True)


def _find_project_root() -> Path:
    """Find the project root by looking for .gingx/ directory."""
    current = Path.cwd()
    for parent in [current] + list(current.parents):
        if (parent / ".gingx").exists():
            return parent
    return current


# ── Entry Point ──────────────────────────────────────────────────────

def main():
    app()


if __name__ == "__main__":
    main()
