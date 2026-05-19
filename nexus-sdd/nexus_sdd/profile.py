"""
Profile Loader — loads, validates, and lists team profiles from .nexus/profiles/
"""

import re
from pathlib import Path
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class AgentProfileConfig:
    model: str = "sonnet"
    effort: str = "high"
    tech_stack: list[str] = field(default_factory=list)


@dataclass
class Profile:
    name: str
    description: str = ""
    type: str = "team"
    default_model: str = "sonnet"
    default_stack: list[str] = field(default_factory=list)
    proactive_interrogation: bool = True
    interrogation_depth: str = "deep"  # basic | deep | exhaustive
    agents: dict[str, AgentProfileConfig] = field(default_factory=dict)
    patterns: dict[str, list[str]] = field(default_factory=dict)
    testing: dict = field(default_factory=dict)

    def get_agent_config(self, agent_name: str) -> AgentProfileConfig:
        """Get config for a specific agent, with profile defaults."""
        if agent_name in self.agents:
            cfg = self.agents[agent_name]
            if not cfg.tech_stack:
                cfg.tech_stack = list(self.default_stack)
            if cfg.model == "sonnet":
                cfg.model = self.default_model
            return cfg
        return AgentProfileConfig(
            model=self.default_model,
            tech_stack=list(self.default_stack),
        )


def load_profile(name: str, profiles_dir: Path | None = None) -> Profile | None:
    """Load a profile by name from .nexus/profiles/<name>.profile.yaml or .yaml"""
    if profiles_dir is None:
        profiles_dir = _default_profiles_dir()

    candidates = [
        profiles_dir / f"{name}.profile.yaml",
        profiles_dir / f"{name}.profile.yml",
        profiles_dir / f"{name}.yaml",
        profiles_dir / f"{name}.yml",
    ]

    profile_path = None
    for c in candidates:
        if c.exists():
            profile_path = c
            break

    if not profile_path:
        return None

    try:
        raw = profile_path.read_text()
        return _parse_profile(raw, name)
    except Exception:
        return None


def list_profiles(profiles_dir: Path | None = None) -> list[str]:
    """List available profile names."""
    if profiles_dir is None:
        profiles_dir = _default_profiles_dir()

    if not profiles_dir.exists():
        return []

    names = set()
    for f in profiles_dir.glob("*.yaml"):
        names.add(_profile_name_from_path(f))
    for f in profiles_dir.glob("*.yml"):
        names.add(_profile_name_from_path(f))
    return sorted(names)


def set_active_profile(name: str, nexus_dir: Path | None = None) -> bool:
    """Write active profile name to .nexus/current_profile.yaml."""
    if nexus_dir is None:
        nexus_dir = Path.cwd() / ".nexus"

    nexus_dir.mkdir(parents=True, exist_ok=True)
    profile_file = nexus_dir / "current_profile.yaml"
    profile_file.write_text(f"active_profile: {name}\n")
    return True


def get_active_profile(nexus_dir: Path | None = None) -> str:
    """Read active profile from .nexus/current_profile.yaml."""
    if nexus_dir is None:
        nexus_dir = Path.cwd() / ".nexus"
    pf = nexus_dir / "current_profile.yaml"
    if pf.exists():
        text = pf.read_text()
        m = re.search(r'active_profile:\s*(\S+)', text)
        if m:
            return m.group(1)
    return "developer"


# ── Internals ─────────────────────────────────────────────────

def _default_profiles_dir() -> Path:
    return Path.cwd() / ".nexus" / "profiles"


def _profile_name_from_path(path: Path) -> str:
    stem = path.stem
    if stem.endswith(".profile"):
        stem = stem[:-8]
    return stem


def _parse_profile(raw: str, name: str) -> Profile:
    """Parse profile YAML using PyYAML for robust parsing."""
    import yaml

    data = yaml.safe_load(raw)
    if not isinstance(data, dict):
        return Profile(name=name)

    profile = Profile(
        name=data.get("name", name),
        description=data.get("description", ""),
        type=data.get("type", "team"),
        default_model=data.get("default_model", "sonnet"),
        default_stack=data.get("default_stack", []),
        proactive_interrogation=data.get("proactive_interrogation", True),
        interrogation_depth=data.get("interrogation_depth", "deep"),
        patterns=data.get("patterns", {}),
        testing=data.get("testing", {}),
    )

    agents_raw = data.get("agents", {})
    if isinstance(agents_raw, dict):
        for ag_name, ag_data in agents_raw.items():
            if isinstance(ag_data, dict):
                profile.agents[ag_name] = AgentProfileConfig(
                    model=ag_data.get("model", "sonnet"),
                    effort=ag_data.get("effort", "high"),
                    tech_stack=ag_data.get("tech_stack", []),
                )
            else:
                profile.agents[ag_name] = AgentProfileConfig()

    return profile
