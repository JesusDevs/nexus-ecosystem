"""
Goal Tracking Engine — objective, key results, autonomous loops.

GoalState: full goal state with progress tracking.
GoalStore: reads/writes .gingx/goals/<id>.yaml
Integrates with mnemo for vector persistence.
"""

from __future__ import annotations

import subprocess
import sys
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Optional

import yaml


@dataclass
class KeyResult:
    """A measurable key result within a goal."""

    description: str
    progress: float = 0.0  # 0.0 .. 1.0

    def is_complete(self) -> bool:
        return self.progress >= 1.0


@dataclass
class GoalState:
    """Full tracked state for one goal."""

    goal_id: str
    objective: str
    key_results: list[KeyResult] = field(default_factory=list)
    status: str = "active"  # active | blocked | completed | archived
    iteration: int = 0
    max_iterations: int = 50
    history: list[str] = field(default_factory=list)
    current_step: str = ""
    blocked_reason: str = ""
    created_at: str = ""
    updated_at: str = ""

    def __post_init__(self):
        now = datetime.now().strftime("%Y-%m-%d %H:%M")
        if not self.created_at:
            self.created_at = now
        if not self.updated_at:
            self.updated_at = now

    def overall_progress(self) -> float:
        if not self.key_results:
            return 0.0
        return round(sum(kr.progress for kr in self.key_results) / len(self.key_results), 2)

    def is_complete(self) -> bool:
        return all(kr.is_complete() for kr in self.key_results)

    def is_blocked(self) -> bool:
        return self.status == "blocked"

    def should_continue(self) -> bool:
        return (
            self.status == "active"
            and not self.is_complete()
            and self.iteration < self.max_iterations
        )

    def to_dict(self) -> dict:
        return {
            "goal_id": self.goal_id,
            "objective": self.objective,
            "key_results": [
                {"description": kr.description, "progress": kr.progress}
                for kr in self.key_results
            ],
            "status": self.status,
            "iteration": self.iteration,
            "max_iterations": self.max_iterations,
            "history": self.history,
            "current_step": self.current_step,
            "blocked_reason": self.blocked_reason,
            "created_at": self.created_at,
            "updated_at": self.updated_at,
        }

    @classmethod
    def from_dict(cls, d: dict) -> GoalState:
        krs = [
            KeyResult(description=kr["description"], progress=kr.get("progress", 0.0))
            for kr in d.get("key_results", [])
        ]
        return cls(
            goal_id=d["goal_id"],
            objective=d.get("objective", d["goal_id"]),
            key_results=krs,
            status=d.get("status", "active"),
            iteration=d.get("iteration", 0),
            max_iterations=d.get("max_iterations", 50),
            history=d.get("history", []),
            current_step=d.get("current_step", ""),
            blocked_reason=d.get("blocked_reason", ""),
            created_at=d.get("created_at", ""),
            updated_at=d.get("updated_at", ""),
        )


class GoalStore:
    """Persists goal entries to .gingx/goals/<id>.yaml."""

    def __init__(self, gingx_dir: Path):
        self.store_dir = gingx_dir / "goals"
        self.store_dir.mkdir(parents=True, exist_ok=True)
        self._project = gingx_dir.resolve().parent.name
        self._mnemo_available = self._check_mnemo()

    def _path(self, goal_id: str) -> Path:
        return self.store_dir / f"{goal_id}.yaml"

    def _check_mnemo(self) -> bool:
        try:
            subprocess.run(
                ["mnemo", "version"], capture_output=True, timeout=2
            )
            return True
        except (FileNotFoundError, subprocess.TimeoutExpired):
            return False

    def _sync_to_mnemo(self, goal: GoalState) -> None:
        if not self._mnemo_available:
            return
        try:
            progress_str = ", ".join(
                f"{kr.description}: {kr.progress:.0%}" for kr in goal.key_results
            )
            subprocess.run([
                "mnemo", "save",
                f"Goal: {goal.objective} — Iter {goal.iteration}",
                f"Status: {goal.status}. Progress: {progress_str}. "
                f"Step: {goal.current_step or 'N/A'}.",
                "--type", "progress",
                "--outcome", "in_progress" if goal.should_continue() else goal.status,
                "--project", self._project,
                "--tags", f"goal,autonomous,{goal.goal_id}",
            ], capture_output=True, timeout=5)
        except (FileNotFoundError, subprocess.TimeoutExpired):
            pass
        except Exception:
            print(
                f"[goal_tracker] warning: mnemo sync failed for {goal.goal_id}",
                file=sys.stderr,
            )

    def exists(self, goal_id: str) -> bool:
        return self._path(goal_id).exists()

    def get(self, goal_id: str) -> Optional[GoalState]:
        p = self._path(goal_id)
        if not p.exists():
            return None
        data = yaml.safe_load(p.read_text()) or {}
        return GoalState.from_dict(data)

    def save(self, goal: GoalState) -> None:
        goal.updated_at = datetime.now().strftime("%Y-%m-%d %H:%M")
        p = self._path(goal.goal_id)
        p.write_text(yaml.dump(
            goal.to_dict(), default_flow_style=False, allow_unicode=True, sort_keys=False
        ))
        self._sync_to_mnemo(goal)

    def load_all(self) -> list[GoalState]:
        if not self.store_dir.exists():
            return []
        goals = []
        for f in sorted(self.store_dir.glob("*.yaml")):
            data = yaml.safe_load(f.read_text())
            if data and data.get("goal_id"):
                goals.append(GoalState.from_dict(data))
        return goals

    def create(
        self,
        goal_id: str,
        objective: str,
        key_results: list[str],
        max_iterations: int = 50,
    ) -> GoalState:
        krs = [KeyResult(description=kr.strip(), progress=0.0) for kr in key_results]
        goal = GoalState(
            goal_id=goal_id,
            objective=objective,
            key_results=krs,
            max_iterations=max_iterations,
            history=[f"Goal created: {objective}"],
        )
        self.save(goal)
        return goal

    def update_progress(self, goal_id: str, kr_index: int, progress: float) -> Optional[GoalState]:
        goal = self.get(goal_id)
        if not goal or kr_index >= len(goal.key_results):
            return None
        goal.key_results[kr_index].progress = min(1.0, max(0.0, progress))
        self.save(goal)
        return goal

    def add_history(self, goal_id: str, entry: str) -> Optional[GoalState]:
        goal = self.get(goal_id)
        if not goal:
            return None
        goal.history.append(entry)
        goal.iteration += 1
        self.save(goal)
        return goal

    def mark_completed(self, goal_id: str) -> Optional[GoalState]:
        goal = self.get(goal_id)
        if not goal:
            return None
        goal.status = "completed"
        for kr in goal.key_results:
            kr.progress = 1.0
        goal.history.append(f"Goal completed at iteration {goal.iteration}")
        self.save(goal)
        return goal

    def mark_blocked(self, goal_id: str, reason: str) -> Optional[GoalState]:
        goal = self.get(goal_id)
        if not goal:
            return None
        goal.status = "blocked"
        goal.blocked_reason = reason
        goal.history.append(f"BLOCKED: {reason}")
        self.save(goal)
        return goal

    def delete(self, goal_id: str) -> bool:
        p = self._path(goal_id)
        if p.exists():
            p.unlink()
            return True
        return False
