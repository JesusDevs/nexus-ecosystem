"""
Auto-Delegation Engine — classifies tasks and dispatches the right agent.

Reads delegation strategy from config.yaml. In automatic mode, dispatches
without asking. In interactive mode, suggests and waits for confirmation.

Wireed into: CLI (gingx-sdd auto), SessionStart hook, PreToolUse hook.
"""

from __future__ import annotations

import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

import yaml

# ── Classification ────────────────────────────────────────────────

ARCHITECTURAL_KEYWORDS = [
    "architecture", "architectural", "refactor", "rewrite", "restructure",
    "migration", "database schema", "data model", "api design", "protocol",
    "breaking change", "deprecate", "new component", "new service",
    "microservice", "pipeline", "infrastructure", "auth", "authentication",
    "authorization", "security", "compliance", "performance critical",
]

AMBIGUOUS_KEYWORDS = [
    "maybe", "probably", "either", "could be", "not sure", "unclear",
    "explore", "investigate", "research", "spike", "prototype",
    "proof of concept", "feasibility", "options", "alternatives",
]

DELEGATION_KEYWORDS = [
    "multiple", "several", "various", "across", "integration",
    "end to end", "e2e", "full stack", "frontend and backend",
    "ui and api", "component and service",
]

RISK_KEYWORDS = [
    "critical", "production", "data loss", "downtime", "revenue",
    "customer impact", "sla", "pii", "gdpr", "hipaa",
]


def classify_task(
    task_description: str,
    changed_files: Optional[list[str]] = None,
    config: Optional[dict] = None,
) -> str:
    """Classify a task: inline, delegate, or full_sdd.

    Uses delegation strategy from config.yaml, falling back to heuristics.
    """
    changed_files = changed_files or []
    desc_lower = task_description.lower()

    # Rule 1: ambiguous or architectural → full_sdd
    if _matches(desc_lower, ARCHITECTURAL_KEYWORDS):
        return "full_sdd"
    if _matches(desc_lower, AMBIGUOUS_KEYWORDS):
        return "full_sdd"
    if _matches(desc_lower, RISK_KEYWORDS):
        return "full_sdd"

    # Rule 2: multiple areas or large scope → delegate
    if _matches(desc_lower, DELEGATION_KEYWORDS):
        return "delegate"
    if len(changed_files) > 5:
        return "delegate"
    if len(task_description.split()) > 50:
        return "delegate"

    # Rule 3: local, small, clear → inline
    if len(changed_files) <= 3:
        return "inline"

    return "delegate"


def _matches(text: str, keywords: list[str]) -> bool:
    return any(kw in text for kw in keywords)


# ── Agent Suggestion ──────────────────────────────────────────────

DEFAULT_PHASE_MAP = {
    "explore": "architect-agent",
    "propose": "po-agent",
    "spec": "po-agent",
    "design": "architect-agent",
    "tasks": "dev-agent",
    "apply": "dev-agent",
    "verify": "qa-agent",
    "security": "devops-agent",
    "archive": "supervisor",
}

DEFAULT_AGENT_ROLES = {
    "supervisor": "SDD Orchestrator — coordina fases, asigna agentes",
    "po-agent": "Product Owner — define alcance, prioridades, criterios de aceptacion",
    "ux-agent": "UX Designer — diseno de interaccion, accesibilidad, consistencia visual",
    "architect-agent": "Solution Architect — estructura de componentes, trade-offs, patterns",
    "dev-agent": "Developer — implementa siguiendo spec + design + TDD",
    "qa-agent": "QA Engineer — verifica BDD scenarios, coverage, regresiones",
    "devops-agent": "DevOps — CI/CD, secrets, infra, dependencias",
}


def suggest_agent(phase: str, config: Optional[dict] = None) -> str:
    """Suggest the right agent for a phase based on orchestrator config."""
    if config:
        phase_map = (
            config.get("harness", {})
            .get("orchestrator", {})
            .get("phase_agent_map", {})
        )
        if phase in phase_map:
            return phase_map[phase]
    return DEFAULT_PHASE_MAP.get(phase, "dev-agent")


def suggest_phase_for_classification(classification: str) -> str:
    """Map classification to the first phase that should run."""
    if classification == "full_sdd":
        return "explore"
    elif classification == "delegate":
        return "design"
    else:
        return "apply"


def agent_role(agent_name: str) -> str:
    return DEFAULT_AGENT_ROLES.get(agent_name, "Specialist")


# ── Config Loading ────────────────────────────────────────────────

def load_delegation_config(project_root: Path | None = None) -> dict:
    """Load delegation and orchestrator config from .gingx/config.yaml."""
    if project_root is None:
        project_root = _find_project_root()
    config_path = project_root / ".gingx" / "config.yaml"
    if not config_path.exists():
        return {}
    try:
        return yaml.safe_load(config_path.read_text()) or {}
    except Exception:
        return {}


def is_delegation_enabled(config: dict) -> bool:
    return config.get("harness", {}).get("delegation", {}).get("enabled", True)


# ── Dispatch ──────────────────────────────────────────────────────

@dataclass
class AutoResult:
    task: str
    classification: str
    suggested_agent: str
    suggested_phase: str
    agent_role: str
    prompt_preview: str
    dispatched: bool = False


def analyze(task_description: str, project_root: Path | None = None) -> AutoResult:
    """Analyze a task and return the classification + agent suggestion.

    Does NOT dispatch — just tells you what would happen.
    """
    project_root = project_root or _find_project_root()
    config = load_delegation_config(project_root)

    classification = classify_task(task_description, config=config)
    phase = suggest_phase_for_classification(classification)
    agent = suggest_agent(phase, config)

    preview_lines = [
        f"Task: {task_description[:120]}",
        f"Classification: {classification}",
        f"Phase: {phase} → Agent: {agent} ({agent_role(agent)})",
    ]

    if classification == "full_sdd":
        preview_lines.append("Workflow: explore → propose → spec → design → tasks → apply → verify → archive")
    elif classification == "delegate":
        preview_lines.append("Workflow: design → tasks → apply → verify")
    else:
        preview_lines.append("Workflow: apply directly (small, clear change)")

    return AutoResult(
        task=task_description,
        classification=classification,
        suggested_agent=agent,
        suggested_phase=phase,
        agent_role=agent_role(agent),
        prompt_preview="\n".join(preview_lines),
    )


def dispatch(
    task_description: str,
    mode: str = "interactive",
    profile_name: str = "developer",
    hdu_id: str | None = None,
    project_root: Path | None = None,
) -> AutoResult:
    """Analyze and dispatch a task to the right agent.

    In automatic mode, spawns the agent directly.
    In interactive mode, returns the suggestion for user approval.
    """
    project_root = project_root or _find_project_root()
    result = analyze(task_description, project_root)

    if mode == "automatic":
        # Actually spawn the agent via team spawn
        try:
            cmd = [
                "python3", "-m", "gingx_sdd", "team", "spawn",
                result.suggested_agent,
                "--task", task_description,
                "--profile", profile_name,
            ]
            if hdu_id:
                cmd.extend(["--hdu-id", hdu_id])

            subprocess.run(
                cmd,
                capture_output=True, text=True, timeout=60,
                cwd=str(project_root),
            )
            result.dispatched = True
        except Exception:
            pass

    return result


def _find_project_root() -> Path:
    current = Path.cwd()
    for parent in [current] + list(current.parents):
        if (parent / ".gingx").exists():
            return parent
    return current
