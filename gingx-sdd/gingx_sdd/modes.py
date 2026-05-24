"""
Gingx Mode System — interactive, automatic, dry_run, off.

Writing the mode to .gingx/mode.yaml controls hook behavior.
"""

from __future__ import annotations

from enum import Enum
from pathlib import Path


class GingxMode(str, Enum):
    INTERACTIVE = "interactive"
    AUTOMATIC = "automatic"
    DRY_RUN = "dry_run"
    OFF = "off"


MODE_FILE = "mode.yaml"


def set_mode(mode: GingxMode, gingx_dir: Path) -> None:
    import yaml
    p = gingx_dir / MODE_FILE
    p.write_text(yaml.dump({"mode": mode.value}, default_flow_style=False))


def get_mode(gingx_dir: Path) -> GingxMode:
    import yaml
    p = gingx_dir / MODE_FILE
    if not p.exists():
        return GingxMode.INTERACTIVE
    try:
        data = yaml.safe_load(p.read_text()) or {}
        val = data.get("mode", "interactive")
        return GingxMode(val)
    except (ValueError, KeyError):
        return GingxMode.INTERACTIVE


def is_harness_active(gingx_dir: Path) -> bool:
    return get_mode(gingx_dir) != GingxMode.OFF
