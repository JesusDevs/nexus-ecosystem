"""
Orchestration module — SDD phase decomposition + agent delegation.

Connects nexus-sdd to mnemo for cross-session memory so sub-agents
always resume from the last checkpoint.
"""

from __future__ import annotations

import subprocess
from pathlib import Path
from dataclasses import dataclass
from typing import Optional

from rich.console import Console
from rich.table import Table
from rich.panel import Panel

console = Console()

# ── Phase → Agent mapping ──────────────────────────────────────────────

PHASE_AGENT_MAP = {
    "explore":  "architect-agent",
    "propose":  "po-agent",
    "spec":     "po-agent",
    "design":   "architect-agent",
    "tasks":    "dev-agent",
    "apply":    "dev-agent",
    "verify":   "qa-agent",
    "security": "devops-agent",
    "archive":  "supervisor",
}

AGENT_ROLES = {
    "supervisor":       "SDD Orchestrator — delegates and tracks",
    "po-agent":         "Product Owner — features, acceptance criteria, scope",
    "ux-agent":         "UX Designer — usability, accessibility, consistency",
    "architect-agent":  "Solution Architect — design, trade-offs, boundaries",
    "dev-agent":        "Developer — implementation, tests, bug fixes",
    "qa-agent":         "QA Engineer — adversarial testing, root cause analysis",
    "devops-agent":     "DevOps — CI/CD, security scans, dependencies",
}

PHASES = ["explore", "propose", "spec", "design", "tasks", "apply", "verify", "archive"]


@dataclass
class OrchestrationStatus:
    hdu_id: str
    current_phase: str
    agent: str
    progress: str  # pending | in_progress | done | blocked
    artifacts: list[str]
    memories: list[str]


def _run_mnemo(*args: str) -> str:
    """Run a mnemo CLI command and return stdout."""
    try:
        result = subprocess.run(
            ["mnemo", *args],
            capture_output=True, text=True, timeout=30,
        )
        return result.stdout.strip()
    except FileNotFoundError:
        return "[mnemo CLI not found]"
    except Exception as e:
        return f"[mnemo error: {e}]"


def get_swarm_mode() -> str:
    """Read swarm.mode from mnemo config. Returns 'hybrid' if not set."""
    output = _run_mnemo("config")
    if output and "[mnemo" not in output:
        for line in output.split("\n"):
            if "swarm.mode" in line:
                parts = line.split("=", 1)
                return parts[1].strip() if len(parts) == 2 else "hybrid"
    return "hybrid"


SWARM_MODE_DESCRIPTIONS = {
    "dag":        "Dependency-only — max parallelism, no delegation",
    "supervisor": "Centralized delegation to specialist agents",
    "swarm":      "Distributed claim-based execution",
    "hybrid":     "DAG resolves + Supervisor assigns + Swarm executes",
}


def decompose_hdu(hdu_id: str) -> list[dict]:
    """Read an HDU from openspec/changes/ and decompose into phases with tasks."""
    spec_path = Path.cwd() / "openspec" / "changes" / hdu_id
    if not spec_path.exists():
        console.print(f"[red]HDU not found: {hdu_id}[/]")
        return []

    phases = []

    # Read available files to determine phase completion
    has_proposal = (spec_path / "proposal.md").exists()
    has_specs = (spec_path / "specs").exists()
    has_design = (spec_path / "design.md").exists()
    has_tasks = (spec_path / "tasks.md").exists()

    # Determine current phase
    if not has_proposal:
        current = "propose"
    elif not has_specs:
        current = "spec"
    elif not has_design:
        current = "design"
    elif not has_tasks:
        current = "tasks"
    else:
        # Check task completion
        tasks_content = (spec_path / "tasks.md").read_text()
        completed = tasks_content.count("[x]")
        total = max(tasks_content.count("[ ]") + completed, 1)
        if completed == 0:
            current = "apply"
        elif completed < total:
            current = "apply"
        else:
            current = "verify"

    # Build phase list from current forward
    start_idx = PHASES.index(current) if current in PHASES else 0
    for phase in PHASES[start_idx:]:
        agent = PHASE_AGENT_MAP.get(phase, "dev-agent")
        phases.append({
            "phase": phase,
            "agent": agent,
            "status": "pending",
        })

    return phases


def suggest_agent(phase: str) -> tuple[str, str]:
    """Return (agent_name, role_description) for a phase."""
    agent = PHASE_AGENT_MAP.get(phase, "dev-agent")
    role = AGENT_ROLES.get(agent, "Generalist")
    return agent, role


def build_prompt(persona: str, task: str, hdu_id: str) -> str:
    """Assemble persona + task + mnemo context into a prompt."""
    # Get relevant mnemo memories
    memories = _run_mnemo("search", task, "--project", Path.cwd().name, "--limit", "3")

    return f"""## Persona: {persona}
## Task: {task}
## HDU: {hdu_id}

### Relevant Memories (mnemo)
{memories if memories else "No prior memories found."}

### Instructions
Apply the {persona} persona to this task. Save your findings to mnemo when done.
"""


def track_progress(hdu_id: str, phase: str, status: str, summary: str = "") -> str:
    """Save orchestration progress to mnemo."""
    project = Path.cwd().name
    return _run_mnemo(
        "save",
        f"{hdu_id} → {phase}: {status}",
        f"Phase: {phase}. Status: {status}. {summary}",
        "--type", "progress",
        "--outcome", status,
        "--project", project,
        "--tags", f"{hdu_id},{phase},orchestration",
    )


def show_status(hdu_id: str) -> OrchestrationStatus:
    """Show orchestration progress for an HDU by reading mnemo."""
    project = Path.cwd().name
    output = _run_mnemo("search", f"progress {hdu_id}", "--project", project, "--limit", "10")

    mode = get_swarm_mode()
    mode_desc = SWARM_MODE_DESCRIPTIONS.get(mode, "")

    phases = decompose_hdu(hdu_id)

    table = Table(title=f"Orchestration Status — {hdu_id}")
    table.add_column("Phase", style="cyan")
    table.add_column("Agent", style="blue")
    table.add_column("Status", style="yellow")

    for p in phases:
        table.add_row(p["phase"], p["agent"], p["status"])

    console.print(table)
    console.print(f"[dim]Swarm mode: {mode} — {mode_desc}[/]")

    if output and "[mnemo" not in output:
        console.print(Panel(output[:1000], title="Recent Progress (mnemo)"))

    return OrchestrationStatus(
        hdu_id=hdu_id,
        current_phase=phases[0]["phase"] if phases else "unknown",
        agent=phases[0]["agent"] if phases else "unknown",
        progress="in_progress" if phases else "done",
        artifacts=[],
        memories=[],
    )


def run_phase(hdu_id: str, phase: str) -> None:
    """Execute a specific phase with the right agent persona."""
    agent, role = suggest_agent(phase)
    mode = get_swarm_mode()
    mode_desc = SWARM_MODE_DESCRIPTIONS.get(mode, "")

    console.print(Panel.fit(
        f"HDU: {hdu_id}\n"
        f"Phase: {phase}\n"
        f"Agent: {agent}\n"
        f"Role: {role}\n"
        f"Swarm Mode: {mode} ({mode_desc})",
        title="Orchestrator — Phase Execution",
    ))

    # Show relevant skill path
    skill_path = Path(__file__).parent.parent / "skills" / "team" / f"{agent}.md"
    if skill_path.exists():
        console.print(f"[dim]Skill: {skill_path}[/]")

    # Track start in mnemo
    result = track_progress(hdu_id, phase, "in_progress", f"Delegated to {agent}")
    console.print(f"[green]{result}[/]")

    # Print invocation instructions for the user
    console.print(f"\n[bold]To invoke this agent:[/]")
    console.print(f"  [cyan]/{agent} <task-description>[/]")
    console.print(f"\n[bold]Or with nexus-sdd:[/]")
    console.print(f"  [cyan]nexus-sdd orchestrate {hdu_id} --agent {agent}[/]")
