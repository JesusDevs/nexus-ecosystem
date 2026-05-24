"""
gingx-sdd init — Bootstrap a Gingx SDD project from templates.

Scaffolds .gingx/, .claude/hooks/, openspec/, and .mcp.json
into the current directory. Detects tech stack automatically.
"""
import os
import shutil
import sys
from pathlib import Path
from typing import Optional


def _resolve_templates_dir() -> Optional[Path]:
    """Resolve the templates directory path."""
    env_override = os.environ.get("GINGX_TEMPLATES")
    if env_override:
        p = Path(env_override)
        if p.is_dir():
            return p
        return None

    # Package-relative: init_project.py → gingx_sdd/ → gingx-sdd/
    pkg_dir = Path(__file__).resolve().parent.parent
    templates = pkg_dir / "templates"
    if templates.is_dir():
        return templates
    return None


def _detect_stack(stack_override: Optional[str] = None) -> dict:
    """Detect the project tech stack. Returns metadata dict."""
    valid_stacks = {
        "python":    {"profile": "fullstack", "type": "backend"},
        "go":        {"profile": "fullstack-go", "type": "backend"},
        "react":     {"profile": "react-nextjs", "type": "web"},
        "node":      {"profile": "fullstack", "type": "backend"},
        "langgraph": {"profile": "fullstack-python-langgraph", "type": "ai"},
        "minimal":   {"profile": "minimal", "type": "cli"},
    }

    if stack_override:
        if stack_override not in valid_stacks:
            print(f"Error: unknown stack '{stack_override}'. Valid: {', '.join(valid_stacks)}")
            sys.exit(1)
        s = valid_stacks[stack_override]
        return {
            "languages": [stack_override],
            "frameworks": [],
            "recommended_profile": s["profile"],
            "type": s["type"],
            "override": True,
        }

    try:
        from gingx_sdd.detector.scanner import detect_project_type
        project = detect_project_type()
        return {
            "languages": project.languages or ["unknown"],
            "frameworks": project.frameworks or [],
            "recommended_profile": project.recommended_profile,
            "type": project.type,
            "override": False,
        }
    except Exception:
        return {
            "languages": ["unknown"],
            "frameworks": [],
            "recommended_profile": "minimal",
            "type": "unknown",
            "override": False,
            "fallback": True,
        }


_VALID_STACKS_STR = "python, go, react, node, langgraph, minimal"
_GINGX_ITEMS = [".gingx", ".claude", "openspec", ".mcp.json"]
_DRY_RUN_FILES = [
    ".claude/agents/architect-agent.md",
    ".claude/agents/dev-agent.md",
    ".claude/agents/devops-agent.md",
    ".claude/agents/goal-agent.md",
    ".claude/agents/po-agent.md",
    ".claude/agents/qa-agent.md",
    ".claude/agents/supervisor.md",
    ".claude/agents/ux-agent.md",
    ".claude/hooks/pre-tool-use.sh",
    ".claude/hooks/stop.sh",
    ".claude/hooks/session-start.sh",
    ".claude/settings.local.json",
    ".gingx/config.yaml",
    ".gingx/current_task.yaml",
    ".gingx/goals/",
    ".gingx/profiles/",
    ".gingx/suites.yaml",
    ".mcp.json",
    "openspec/AGENTS.md",
    "openspec/changes/",
]


def init_project(
    dry_run: bool = False,
    force: bool = False,
    stack: Optional[str] = None,
) -> int:
    """Bootstrap a Gingx SDD project in the current directory.

    Returns 0 on success, 1 on error.
    """
    templates = _resolve_templates_dir()
    if templates is None:
        print("Error: templates directory not found. Set GINGX_TEMPLATES env var or run from the gingx-sdd package.")
        return 1

    cwd = Path.cwd()
    actions: list[tuple[str, str]] = []  # (action, path) for dry-run

    # ── Pre-check: .gingx/ exists ────────────────────────────────────
    gingx_dir = cwd / ".gingx"
    if gingx_dir.exists() and not force:
        print("Error: .gingx/ already exists. Use --force to overwrite.")
        return 1

    if force and not dry_run:
        for item in _GINGX_ITEMS:
            target = cwd / item
            if target.exists():
                if target.is_dir():
                    shutil.rmtree(target)
                else:
                    target.unlink()
        actions.append(("rmtree", str(cwd / ".gingx")))
        actions.append(("rmtree", str(cwd / ".claude")))
        actions.append(("rmtree", str(cwd / "openspec")))
    elif force:
        actions.append(("rmtree", str(cwd / ".gingx")))
        actions.append(("rmtree", str(cwd / ".claude")))
        actions.append(("rmtree", str(cwd / "openspec")))

    # ── Detect stack ─────────────────────────────────────────────────
    meta = _detect_stack(stack)

    # ══════════════════════════════════════════════════════════════════
    if dry_run:
        _print_dry_run(templates, actions, meta)
        return 0

    # ══════════════════════════════════════════════════════════════════
    # Real file operations below
    # ══════════════════════════════════════════════════════════════════

    created: list[str] = []

    # ── .gingx/ from template ────────────────────────────────────────
    gingx_tpl = templates / ".gingx"
    if gingx_tpl.is_dir():
        shutil.copytree(gingx_tpl, str(gingx_dir), dirs_exist_ok=True)
        created.append(".gingx/")

    # ── current_task.yaml ────────────────────────────────────────────
    tracking_path = gingx_dir / "current_task.yaml"
    tracking_path.write_text(
        "# Gingx SDD — Active Task Tracking\n"
        "hdu_id: none\n"
        "phase: none\n"
        "agent: none\n"
    )
    created.append(".gingx/current_task.yaml")

    # ── .claude/hooks/ ───────────────────────────────────────────────
    hooks_dst = cwd / ".claude" / "hooks"
    hooks_dst.mkdir(parents=True, exist_ok=True)
    hooks_tpl = templates / ".claude" / "hooks"
    if hooks_tpl.is_dir():
        for hook_file in hooks_tpl.iterdir():
            if hook_file.suffix == ".sh":
                dst = hooks_dst / hook_file.name
                shutil.copy2(hook_file, dst)
                os.chmod(dst, 0o755)
                created.append(str(dst.relative_to(cwd)))
        created.append(".claude/hooks/")

    # ── settings.local.json ──────────────────────────────────────────
    settings_tpl = templates / ".claude" / "settings.local.json"
    if settings_tpl.is_file():
        settings_dst = cwd / ".claude" / "settings.local.json"
        settings_dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(settings_tpl, settings_dst)
        created.append(".claude/settings.local.json")

    # ── openspec/ ────────────────────────────────────────────────────
    openspec_tpl = templates / "openspec" / "AGENTS.md"
    if openspec_tpl.is_file():
        openspec_dir = cwd / "openspec"
        openspec_dir.mkdir(parents=True, exist_ok=True)
        shutil.copy2(openspec_tpl, openspec_dir / "AGENTS.md")
        (openspec_dir / "changes").mkdir(exist_ok=True)
        created.append("openspec/")

    # ── .gingx/goals/ ───────────────────────────────────────────────
    (gingx_dir / "goals").mkdir(parents=True, exist_ok=True)
    created.append(".gingx/goals/")

    # ── .claude/agents/ (8 SDD personas) ─────────────────────────────
    agents_tpl = templates / ".claude" / "agents"
    if agents_tpl.is_dir():
        agents_dst = cwd / ".claude" / "agents"
        agents_dst.mkdir(parents=True, exist_ok=True)
        for agent_file in agents_tpl.iterdir():
            if agent_file.suffix == ".md":
                shutil.copy2(agent_file, agents_dst / agent_file.name)
        created.append(".claude/agents/")

    # ── .mcp.json ────────────────────────────────────────────────────
    mcp_tpl = templates / ".mcp.json"
    if mcp_tpl.is_file():
        shutil.copy2(mcp_tpl, cwd / ".mcp.json")
        created.append(".mcp.json")

    # ── Summary ──────────────────────────────────────────────────────
    _print_summary(meta, created)
    return 0


def _print_dry_run(templates: Path, actions: list[tuple[str, str]], meta: dict) -> None:
    """Print a preview of what init would do."""
    print()
    print("╔══════════════════════════════════════════════╗")
    print("║  🏭  gingx-sdd init — DRY RUN               ║")
    print("╚══════════════════════════════════════════════╝")
    print()
    print(f"  Templates: {templates}")
    print(f"  Target:    {Path.cwd()}")
    print()
    print(f"  Detected stack: {', '.join(meta['languages'])} ({meta['type']})")
    print(f"  Profile:        {meta['recommended_profile']}")
    if meta.get("override"):
        print(f"  (stack forced via --stack)")
    print()
    print("  Would create:")
    for item in _DRY_RUN_FILES:
        print(f"    + {item}")
    print()


def _print_summary(meta: dict, created: list[str]) -> None:
    """Print the completion summary."""
    print()
    print("╔══════════════════════════════════════════════╗")
    print("║   🏭  GINGX-SDD — Project Initialized       ║")
    print("╚══════════════════════════════════════════════╝")
    print()
    print(f"  Stack:   {', '.join(meta['languages'])} ({meta['type']})")
    print(f"  Profile: {meta['recommended_profile']}")
    if meta.get("fallback"):
        print(f"  (detector failed, using minimal profile)")
    print()
    print("  Created:")
    for item in sorted(created):
        print(f"    ✓ {item}")
    print()
    print("  Next:")
    print("    gingx-sdd hdu create 'My first feature' --question 'What problem?'")
    print("    gingx-sdd status")
    print("    gingx-sdd mode status")
    print()
