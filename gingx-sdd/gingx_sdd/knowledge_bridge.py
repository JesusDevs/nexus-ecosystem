"""
Knowledge Bridge — exporta codegraph a mnemo para cross-project memory.

Permite que proyectos compartan estructura de codigo:
- Export: codegraph symbols → mnemo type=code-structure
- Import: buscar en mnemo estructuras de otros proyectos
"""

from __future__ import annotations

import subprocess
from datetime import datetime
from pathlib import Path
from typing import Optional


def export_codegraph_to_mnemo(
    project_path: str | Path,
    project_name: str | None = None,
    hdu_id: str | None = None,
    max_symbols: int = 50,
) -> dict:
    """Export codegraph symbols to mnemo as code-structure memories.

    Returns a dict with status and count of exported symbols.
    """
    project_path = Path(project_path)
    project_name = project_name or project_path.name

    # Check codegraph is available
    if not _check_cmd("codegraph"):
        return {"status": "error", "message": "codegraph not installed", "count": 0}

    if not (project_path / ".codegraph").exists():
        return {"status": "error", "message": "codegraph not initialized in project", "count": 0}

    # Get codegraph status for stats
    try:
        result = subprocess.run(
            ["codegraph", "status", str(project_path)],
            capture_output=True, text=True, timeout=30,
        )
        stats_output = result.stdout[:500] if result.returncode == 0 else ""
    except Exception:
        stats_output = ""

    # Query top symbols from codegraph
    exported = 0
    errors = []

    for kind in ["class", "function", "method"]:
        try:
            result = subprocess.run(
                ["codegraph", "query", "", "--kind", kind, "--limit", str(max_symbols // 3), "--json"],
                capture_output=True, text=True, timeout=30,
                cwd=str(project_path),
            )
            if result.returncode != 0 or not result.stdout.strip():
                continue

            import json
            items = json.loads(result.stdout)
            for item in items[:max_symbols // 3]:
                name = item.get("name", "unknown")
                file_path = item.get("file", "")
                kind_str = item.get("kind", kind)

                try:
                    subprocess.run(
                        ["mnemo", "save",
                         f"{kind_str}: {name}",
                         f"{kind_str} '{name}' defined in {file_path}. Project: {project_name}",
                         "--type", "code-structure",
                         "--project", project_name,
                         "--tags", f"codegraph,{kind_str},export"],
                        capture_output=True, text=True, timeout=15,
                    )
                    exported += 1
                except Exception as e:
                    errors.append(str(e))
        except Exception as e:
            errors.append(f"{kind}: {e}")

    # Save an aggregate summary
    try:
        summary = f"CodeGraph export from {project_name}. Symbols: {exported}. Stats: {stats_output}"
        subprocess.run(
            ["mnemo", "save",
             f"CodeGraph export: {project_name}",
             summary,
             "--type", "code-structure",
             "--project", project_name,
             "--tags", "codegraph,export,summary"],
            capture_output=True, text=True, timeout=15,
        )
    except Exception:
        pass

    return {
        "status": "ok" if exported > 0 else "empty",
        "message": f"Exported {exported} symbols to mnemo" + (f". Errors: {len(errors)}" if errors else ""),
        "count": exported,
    }


def import_codegraph_from_mnemo(project_name: str, query: str = "", limit: int = 10) -> list[str]:
    """Search mnemo for code structures from other projects.

    Returns a list of relevant code structure descriptions.
    """
    if not _check_cmd("mnemo"):
        return []

    try:
        search_query = query or "code structure"
        cmd = ["mnemo", "search", search_query, "--limit", str(limit), "--type", "code-structure"]
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip().split("\n")
    except Exception:
        pass
    return []


def get_knowledge_status(project_path: str | Path) -> dict:
    """Get combined status of codegraph + mnemo cross-project knowledge."""
    project_path = Path(project_path)
    status = {
        "codegraph_installed": _check_cmd("codegraph"),
        "codegraph_initialized": (project_path / ".codegraph").exists(),
        "mnemo_installed": _check_cmd("mnemo"),
    }

    if status["codegraph_initialized"]:
        try:
            result = subprocess.run(
                ["codegraph", "status", str(project_path)],
                capture_output=True, text=True, timeout=30,
            )
            status["codegraph_stats"] = result.stdout[:300] if result.returncode == 0 else "unknown"
        except Exception:
            status["codegraph_stats"] = "error"

    return status


def _check_cmd(cmd: str) -> bool:
    try:
        subprocess.run(["which", cmd], capture_output=True, text=True)
        return True
    except Exception:
        return False
