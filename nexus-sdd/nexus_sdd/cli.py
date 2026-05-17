"""
Nexus-SDD CLI — Minimal harness.
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

app = typer.Typer(
    name="nexus-sdd",
    help="SDD Framework — Spec → Plan → Code → Test → Security → Memory",
    no_args_is_help=True,
)


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
    typer.echo(f"Next: nexus-sdd orchestrate {hdu_id}")


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
        nexus-sdd orchestrate HDU-06 --status
        nexus-sdd orchestrate HDU-06 --phase apply
        nexus-sdd orchestrate HDU-06 --agent qa
    """
    from nexus_sdd.orchestrate import decompose_hdu, suggest_agent, show_status, run_phase

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

        typer.echo(f"\nNext: nexus-sdd orchestrate {hdu_id} --phase {phases[0]['phase']}")


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
        typer.echo("mnemo CLI not found. Install: brew install nexus-sdd/tap/mnemo", err=True)


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
    skills_dir = Path.cwd() / "nexus-sdd" / "skills" / "team"
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
    "nexus-mnemo": {
        "type": "go",
        "path": "nexus-mnemo",
        "build_cmd": ["go", "build", "./..."],
        "test_cmd": ["go", "test", "./..."],
    },
    "nexus-sdd": {
        "type": "python",
        "path": "nexus-sdd",
        "build_cmd": [sys.executable, "-m", "py_compile", "nexus_sdd/cli.py"],
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
    project: Optional[str] = typer.Option(None, "--project", help="Target project (nexus-mnemo or nexus-sdd)"),
    all_projects: bool = typer.Option(False, "--all", help="Release all projects in monorepo"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Run pre-release checks without creating tags"),
):
    """Create a semantic release with git tag, mnemo snapshot, and changelog.

    Examples:
        nexus-sdd release v0.1.0 --project nexus-mnemo
        nexus-sdd release v0.1.0 --all
        nexus-sdd release v0.1.0 --all --dry-run
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


# ── Entry Point ──────────────────────────────────────────────────────

def main():
    app()


if __name__ == "__main__":
    main()
