"""
Agent Registry — discovers, validates, and indexes agent personas + tech stack skills.

Referenced by install.sh: from nexus_sdd.skills.registry import SkillRegistry
"""

import re
from pathlib import Path
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class AgentDef:
    name: str
    description: str
    when_to_use: str
    model: str = "sonnet"
    effort: str = "high"
    triggers: list[str] = field(default_factory=list)
    persona_path: Path | None = None
    tech_stacks: list[str] = field(default_factory=list)
    profile: str = "developer"


@dataclass
class TechSkillDef:
    name: str
    description: str
    category: str = "general"
    stack: list[str] = field(default_factory=list)
    triggers: list[str] = field(default_factory=list)
    skill_path: Path | None = None


class SkillRegistry:
    """Discovers, validates, and indexes agent personas + tech stack skills."""

    def __init__(self):
        self.agents: dict[str, AgentDef] = {}
        self.tech_stacks: dict[str, TechSkillDef] = {}
        self._personas_dir: Path | None = None
        self._tech_dir: Path | None = None

    def scan_personas(self, team_dir: Path) -> dict[str, AgentDef]:
        """Scan skills/team/*.md, extract YAML frontmatter, validate fields."""
        self._personas_dir = Path(team_dir)
        self.agents.clear()

        if not self._personas_dir.exists():
            return self.agents

        for md_file in sorted(self._personas_dir.glob("*.md")):
            agent = self._parse_agent_frontmatter(md_file)
            if agent:
                self.agents[agent.name] = agent

        return self.agents

    def scan_tech_stacks(self, extras_dir: Path) -> dict[str, TechSkillDef]:
        """Scan extras/skills/**/*.md, extract YAML frontmatter."""
        self._tech_dir = Path(extras_dir)
        self.tech_stacks.clear()

        if not self._tech_dir.exists():
            return self.tech_stacks

        for md_file in sorted(self._tech_dir.rglob("*.md")):
            skill = self._parse_tech_frontmatter(md_file)
            if skill:
                self.tech_stacks[skill.name] = skill

        return self.tech_stacks

    def get_agent(self, name: str) -> AgentDef | None:
        """Return full agent definition by name."""
        return self.agents.get(name)

    def resolve_tech_stack(self, agent_name: str, profile_stack: list[str] | None = None) -> list[TechSkillDef]:
        """Resolve tech stack: profile override > default."""
        resolved: list[TechSkillDef] = []
        if profile_stack:
            for ts_name in profile_stack:
                if ts_name in self.tech_stacks:
                    resolved.append(self.tech_stacks[ts_name])
        return resolved

    def install_for_project(self, skill_names: list[str], target_dir: Path) -> list[str]:
        """Install specific skills to target directory. Returns installed names."""
        target = Path(target_dir)
        target.mkdir(parents=True, exist_ok=True)
        installed = []
        for name in skill_names:
            if name in self.tech_stacks and self.tech_stacks[name].skill_path:
                dest = target / self.tech_stacks[name].skill_path.name
                dest.write_text(self.tech_stacks[name].skill_path.read_text())
                installed.append(name)
        return installed

    def generate_catalog(self, output_path: Path) -> None:
        """Generate .nexus/skill-registry.md catalog."""
        lines = [
            "# Skill Registry",
            "",
            f"Generated: {len(self.agents)} agents, {len(self.tech_stacks)} tech stacks",
            "",
            "## Agents",
            "",
        ]
        for name, agent in sorted(self.agents.items()):
            triggers = ", ".join(agent.triggers) if agent.triggers else "none"
            lines.append(f"- **{name}** ({agent.model}, {agent.effort}) — {agent.description.split(chr(10))[0]}")
            lines.append(f"  Triggers: {triggers}")

        lines.extend(["", "## Tech Stacks", ""])
        by_category: dict[str, list[TechSkillDef]] = {}
        for ts in self.tech_stacks.values():
            by_category.setdefault(ts.category, []).append(ts)

        for cat in sorted(by_category):
            lines.append(f"### {cat}")
            for ts in sorted(by_category[cat], key=lambda t: t.name):
                stacks = ", ".join(ts.stack) if ts.stack else "general"
                lines.append(f"- **{ts.name}** ({stacks}) — {ts.description}")

        Path(output_path).parent.mkdir(parents=True, exist_ok=True)
        Path(output_path).write_text("\n".join(lines) + "\n")

    def list_by_category(self) -> dict[str, list[str]]:
        """Group skills by category."""
        result: dict[str, list[str]] = {}
        for ts in self.tech_stacks.values():
            result.setdefault(ts.category, []).append(ts.name)
        return result

    # ── YAML frontmatter parser (regex, no PyYAML) ──────────────

    _FM_RE = re.compile(r'^---\s*\n(.*?)\n---', re.DOTALL)

    def _parse_agent_frontmatter(self, filepath: Path) -> AgentDef | None:
        try:
            text = filepath.read_text()
        except Exception:
            return None

        m = self._FM_RE.match(text)
        if not m:
            return None

        fm = self._parse_yaml_simple(m.group(1))
        return AgentDef(
            name=fm.get("name", filepath.stem),
            description=fm.get("description", "").replace("\n", " ").strip(),
            when_to_use=fm.get("when_to_use", "").replace("\n", " ").strip(),
            model=fm.get("model", "sonnet"),
            effort=fm.get("effort", "high"),
            triggers=self._parse_yaml_list(fm.get("triggers", "")),
            persona_path=filepath,
        )

    def _parse_tech_frontmatter(self, filepath: Path) -> TechSkillDef | None:
        try:
            text = filepath.read_text()
        except Exception:
            return None

        m = self._FM_RE.match(text)
        if not m:
            return None

        fm = self._parse_yaml_simple(m.group(1))
        return TechSkillDef(
            name=fm.get("name", filepath.stem),
            description=fm.get("description", "").replace("\n", " ").strip(),
            category=fm.get("category", "general"),
            stack=self._parse_yaml_list(fm.get("stack", "")),
            triggers=self._parse_yaml_list(fm.get("triggers", "")),
            skill_path=filepath,
        )

    def _parse_yaml_simple(self, text: str) -> dict[str, str]:
        """Parse flat YAML key: value pairs (sufficient for frontmatter)."""
        result: dict[str, str] = {}
        current_key: str | None = None
        current_val: list[str] = []

        for line in text.split("\n"):
            if line.startswith(" ") or line.startswith("\t"):
                # continuation of previous value
                if current_key:
                    current_val.append(line.strip())
                continue

            # flush previous key
            if current_key:
                result[current_key] = "\n".join(current_val)
                current_val = []

            if ":" in line:
                key, _, val = line.partition(":")
                current_key = key.strip()
                current_val = [val.strip()] if val.strip() else []
            else:
                current_key = None
                current_val = []

        if current_key:
            result[current_key] = "\n".join(current_val)

        return result

    def _parse_yaml_list(self, text: str) -> list[str]:
        """Parse YAML list: '[a, b, c]' or comma-separated."""
        text = text.strip()
        if text.startswith("[") and text.endswith("]"):
            inner = text[1:-1]
            return [item.strip().strip("'\"") for item in inner.split(",") if item.strip()]
        if text:
            return [text]
        return []
