"""
Prompt Builder — shared engine for assembling sub-agent prompts.

Used by: CLI team spawn, MCP agent_dispatch, hook auto-dispatch.
"""

import subprocess
from pathlib import Path
from dataclasses import dataclass, field
from typing import Optional

from .skills.registry import SkillRegistry, AgentDef, TechSkillDef
from .profile import Profile, AgentProfileConfig


@dataclass
class AssembledPrompt:
    agent_name: str
    profile_name: str
    system_prompt: str
    task: str
    hdu_id: str | None = None
    tech_stacks_loaded: list[str] = field(default_factory=list)
    mnemo_context: str = ""
    codegraph_context: str = ""
    cross_project_context: str = ""
    interrogation_mode: bool = False


def assemble_agent_prompt(
    agent_name: str,
    task: str,
    profile: Profile,
    registry: SkillRegistry,
    hdu_id: str | None = None,
    tech_stack_override: list[str] | None = None,
    project: str | None = None,
    personas_dir: Path | None = None,
) -> AssembledPrompt:
    """
    Assemble the complete sub-agent prompt.

    Structure:
    [SYSTEM] You are {agent_name}: {persona_description}
    [PROFILE] Active profile: {profile}. Model: {model}
    [TECH STACK] Compacted rules from: {tech_skills}
    [MNEMO CONTEXT] Prior relevant decisions
    [INTERROGATION] Proactive questions (if enabled)
    [TASK] {task}
    [OUTPUT CONTRACT] status, summary, artifacts, next, risks
    """
    agent_def = registry.get_agent(agent_name)
    agent_cfg = profile.get_agent_config(agent_name)

    # Resolve tech stacks
    stack_names = tech_stack_override or agent_cfg.tech_stack
    tech_skills = registry.resolve_tech_stack(agent_name, stack_names)

    # Determine personas directory
    if personas_dir is None:
        personas_dir = Path(__file__).parent.parent.parent / "skills" / "team"

    lines: list[str] = []

    # ── SYSTEM / PERSONA ─────────────────────
    if agent_def:
        persona_text = _load_persona(agent_def, personas_dir)
        lines.append(persona_text)
    else:
        lines.append(f"# You are {agent_name}")
        lines.append("")

    # ── PROFILE ─────────────────────────────
    lines.append("## Profile")
    lines.append(f"Active profile: **{profile.name}** — {profile.description}")
    lines.append(f"Model: `{agent_cfg.model}` | Effort: `{agent_cfg.effort}`")
    lines.append("")

    # ── TECH STACK ──────────────────────────
    if tech_skills:
        lines.append("## Tech Stack")
        for ts in tech_skills:
            lines.append(f"### {ts.name}: {ts.description}")
            if ts.skill_path and ts.skill_path.exists():
                rules = _compact_skill(ts.skill_path.read_text(), max_rules=10)
                lines.append(rules)
            lines.append("")
        lines.append("")

    # ── CODEGRAPH CONTEXT ───────────────────
    codegraph_context = ""
    if agent_name in ("architect-agent", "dev-agent", "qa-agent"):
        codegraph_context = _search_codegraph(task)
        if codegraph_context:
            lines.append("## CodeGraph Context (codebase structure)")
            lines.append(codegraph_context)
            lines.append("")

    # ── MNEMO CONTEXT ───────────────────────
    mnemo_context = _search_mnemo(task, project)
    if mnemo_context:
        lines.append("## Mnemo Context (prior knowledge)")
        lines.append(mnemo_context)
        lines.append("")

    # ── CROSS-PROJECT KNOWLEDGE ─────────────
    cross_project = _search_cross_project(task, project)
    if cross_project:
        lines.append("## Cross-Project Knowledge (from other codebases)")
        lines.append(cross_project)
        lines.append("")

    # ── INTERROGATION ──────────────────────
    if profile.proactive_interrogation:
        interrogation = _load_interrogation(profile.interrogation_depth, project or "")
        lines.append("## Proactive Interrogation Mode")
        lines.append(interrogation)
        lines.append("")

    # ── TASK ───────────────────────────────
    lines.append("## Task")
    lines.append(task)
    lines.append("")

    # ── HDU CONTEXT ────────────────────────
    if hdu_id:
        hdu_context = _load_hdu_context(hdu_id)
        if hdu_context:
            lines.append("## HDU Context")
            lines.append(hdu_context)
            lines.append("")

    # ── OUTPUT CONTRACT ────────────────────
    lines.append("## Output Contract")
    lines.append("Structure your response with these sections:")
    lines.append("- **Status**: ok | blocked | needs_human")
    lines.append("- **Executive Summary**: 2-3 sentences")
    lines.append("- **Artifacts Produced**: file paths")
    lines.append("- **Next Recommended**: suggested action")
    lines.append("- **Risks Open**: unresolved concerns")

    system_prompt = "\n".join(lines)

    return AssembledPrompt(
        agent_name=agent_name,
        profile_name=profile.name,
        system_prompt=system_prompt,
        task=task,
        hdu_id=hdu_id,
        tech_stacks_loaded=stack_names,
        mnemo_context=mnemo_context,
        codegraph_context=codegraph_context,
        cross_project_context=cross_project,
        interrogation_mode=profile.proactive_interrogation,
    )


# ── Internals ─────────────────────────────────────────────────

def _load_persona(agent_def: AgentDef, personas_dir: Path) -> str:
    """Load agent persona markdown, stripping the YAML frontmatter."""
    if agent_def.persona_path and agent_def.persona_path.exists():
        text = agent_def.persona_path.read_text()
    else:
        candidate = personas_dir / f"{agent_def.name}.md"
        if candidate.exists():
            text = candidate.read_text()
        else:
            return f"# Agent: {agent_def.name}\n\n{agent_def.description}\n"

    # Strip YAML frontmatter (between first --- and second ---)
    parts = text.split("---", 2)
    if len(parts) >= 3:
        return parts[2].strip()
    elif len(parts) == 2:
        return parts[1].strip()
    return text.strip()


def _compact_skill(skill_text: str, max_rules: int = 10) -> str:
    """Compact a tech stack skill into digestible rules (max N rules)."""
    # Extract bullet points, numbered rules, do's/don'ts
    lines = skill_text.split("\n")
    rules: list[str] = []
    in_rules = False

    for line in lines:
        stripped = line.strip()
        if not stripped:
            continue
        # Skip frontmatter
        if stripped.startswith("---"):
            in_rules = not in_rules
            continue
        if not in_rules:
            if stripped.startswith("##") or stripped.startswith("# "):
                in_rules = True
            continue
        # Collect rules
        if stripped.startswith("- ") or stripped.startswith("* "):
            rules.append(stripped)
        elif stripped.startswith("##"):
            break  # stop at next section
        elif rules and len(stripped) > 20:
            rules.append(stripped)

        if len(rules) >= max_rules:
            break

    return "\n".join(rules[:max_rules]) if rules else "No compacted rules available."


def _search_mnemo(task: str, project: str | None = None) -> str:
    """Search mnemo for context relevant to this task."""
    try:
        cmd = ["mnemo", "search", task, "--limit", "5", "--min-sim", "0.2"]
        if project:
            cmd.extend(["--project", project])
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip()
    except Exception:
        pass
    return ""


def _search_codegraph(task: str) -> str:
    """Search codegraph for relevant code structure context."""
    try:
        cmd = ["codegraph", "context", task, "--format", "markdown",
               "--max-nodes", "20"]
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip()[:3000]
    except Exception:
        pass
    return ""


def _search_cross_project(task: str, project: str | None = None) -> str:
    """Search mnemo for code structures from other projects."""
    try:
        cmd = ["mnemo", "search", f"code structure {task}", "--limit", "3", "--type", "code-structure"]
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip()
    except Exception:
        pass
    return ""


def _load_interrogation(depth: str, project: str) -> str:
    """Load the interrogation mode template. Splits by '---\\ndepth:' markers."""
    template_path = Path(__file__).parent.parent.parent / "templates" / "interrogation-mode.md"
    if not template_path.exists():
        return _default_interrogation(depth)

    text = template_path.read_text()

    # Split template into sections by "---\ndepth:" markers
    sections: dict[str, str] = {}
    current_depth: str | None = None
    current_lines: list[str] = []

    for line in text.split("\n"):
        # Check for section delimiter: "---\ndepth: <name>"
        stripped = line.strip()
        if stripped == "---":
            # Save previous section
            if current_depth:
                sections[current_depth] = "\n".join(current_lines).strip()
                current_lines = []
            current_depth = None
            continue

        if stripped.startswith("depth:"):
            current_depth = stripped[len("depth:"):].strip()
            continue

        if current_depth:
            current_lines.append(line)

    # Save last section
    if current_depth and current_lines:
        sections[current_depth] = "\n".join(current_lines).strip()

    # Get the right depth, fallback to deep
    content = sections.get(depth) or sections.get("deep", "")
    if not content:
        return _default_interrogation(depth)

    # Replace variables
    content = content.replace("{project}", project)
    return content


def _default_interrogation(depth: str) -> str:
    """Fallback interrogation template."""
    if depth == "basic":
        return (
            "Before acting, ask:\n"
            "1. What is the exact scope boundary?\n"
            "2. What are the priority constraints?\n"
            "3. What does success look like?"
        )
    return (
        "MANDATORY before any action:\n"
        "1. SEARCH mnemo for related context\n"
        "2. ASK clarifying questions if ANY of these are unclear:\n"
        "   - Scope: MVP vs full vision?\n"
        "   - Constraints: latency, scale, security?\n"
        "   - Dependencies: what systems does this touch?\n"
        "   - Patterns: has something similar been done?\n"
        "   - Success: how do we measure completion?\n"
        "3. PROPOSE options before executing.\n"
        "NEVER assume context. If you don't know, ASK."
    )


def _load_hdu_context(hdu_id: str) -> str:
    """Load HDU context from openspec/changes/<HDU>/."""
    openspec_dir = Path.cwd() / "openspec" / "changes" / hdu_id
    if not openspec_dir.exists():
        return ""

    parts: list[str] = []
    for fname in ["proposal.md", "design.md", "tasks.md"]:
        fpath = openspec_dir / fname
        if fpath.exists():
            content = fpath.read_text()
            # Take first 500 chars
            parts.append(f"### {fname}\n{content[:500].strip()}\n")

    return "\n".join(parts) if parts else ""
