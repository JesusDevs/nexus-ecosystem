"""
HDU Tracking Engine — blocking questions, progress, YAML persistence.

BlockingQuestion: a question that must be answered before code can be written.
HDUEntry: full HDU state. HDUStore: reads/writes .gingx/hdus/<id>.yaml
"""

from __future__ import annotations

import subprocess
import sys
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Optional

import yaml

PHASE_WEIGHTS = {
    "init": 5, "explore": 15, "propose": 10, "spec": 15,
    "design": 15, "tasks": 10, "apply": 20, "verify": 10, "archive": 0,
}


@dataclass
class BlockingQuestion:
    """A question that blocks progress until answered."""

    question: str
    context: str = ""
    created_at: str = ""
    answered_at: Optional[str] = None
    answer: Optional[str] = None
    blocked_by: str = ""  # agent name that raised the question

    def is_resolved(self) -> bool:
        return self.answer is not None


@dataclass
class HDUEntry:
    """Full tracked state for one HDU."""

    id: str
    title: str
    phase: str = "init"
    status: str = "active"  # active | blocked | completed | archived
    progress: float = 0.0
    blockers: list[BlockingQuestion] = field(default_factory=list)
    artifacts_produced: list[str] = field(default_factory=list)
    executive_summary: str = ""
    next_recommended: str = ""
    risks_open: list[str] = field(default_factory=list)
    started_at: str = ""
    updated_at: str = ""
    completed_phases: list[str] = field(default_factory=list)

    def __post_init__(self):
        if not self.started_at:
            self.started_at = datetime.now().strftime("%Y-%m-%d %H:%M")
        if not self.updated_at:
            self.updated_at = self.started_at

    def is_blocked(self) -> bool:
        return any(not b.is_resolved() for b in self.blockers)

    def open_blockers(self) -> list[BlockingQuestion]:
        return [b for b in self.blockers if not b.is_resolved()]

    def calc_progress(self) -> float:
        if not self.completed_phases:
            return 0.0
        completed_weight = sum(
            PHASE_WEIGHTS.get(p, 5) for p in self.completed_phases
        )
        total_weight = sum(PHASE_WEIGHTS.values())
        return round(completed_weight / total_weight * 100, 1)

    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "title": self.title,
            "phase": self.phase,
            "status": self.status,
            "progress": self.progress,
            "blockers": [
                {
                    "question": b.question,
                    "context": b.context,
                    "created_at": b.created_at,
                    "answered_at": b.answered_at,
                    "answer": b.answer,
                    "blocked_by": b.blocked_by,
                }
                for b in self.blockers
            ],
            "artifacts_produced": self.artifacts_produced,
            "executive_summary": self.executive_summary,
            "next_recommended": self.next_recommended,
            "risks_open": self.risks_open,
            "started_at": self.started_at,
            "updated_at": self.updated_at,
            "completed_phases": self.completed_phases,
        }

    @classmethod
    def from_dict(cls, d: dict) -> HDUEntry:
        blockers = []
        for b in d.get("blockers", []):
            blockers.append(BlockingQuestion(
                question=b["question"],
                context=b.get("context", ""),
                created_at=b.get("created_at", ""),
                answered_at=b.get("answered_at"),
                answer=b.get("answer"),
                blocked_by=b.get("blocked_by", ""),
            ))
        return cls(
            id=d["id"],
            title=d.get("title", d["id"]),
            phase=d.get("phase", "init"),
            status=d.get("status", "active"),
            progress=d.get("progress", 0.0),
            blockers=blockers,
            artifacts_produced=d.get("artifacts_produced", []),
            executive_summary=d.get("executive_summary", ""),
            next_recommended=d.get("next_recommended", ""),
            risks_open=d.get("risks_open", []),
            started_at=d.get("started_at", ""),
            updated_at=d.get("updated_at", ""),
            completed_phases=d.get("completed_phases", []),
        )


class HDUStore:
    """Persists HDU entries to .gingx/hdus/<id>.yaml."""

    def __init__(self, gingx_dir: Path):
        self.store_dir = gingx_dir / "hdus"
        self.store_dir.mkdir(parents=True, exist_ok=True)
        self._project = gingx_dir.resolve().parent.name
        self._mnemo_available = self._check_mnemo()

    def _path(self, hdu_id: str) -> Path:
        return self.store_dir / f"{hdu_id}.yaml"

    def _check_mnemo(self) -> bool:
        """Check if mnemo CLI is available."""
        try:
            subprocess.run(
                ["mnemo", "version"], capture_output=True, timeout=2
            )
            return True
        except (FileNotFoundError, subprocess.TimeoutExpired):
            return False

    def _sync_to_mnemo(self, entry: HDUEntry) -> None:
        """Dual-write HDU to mnemo vector memory. Fail-open: if mnemo
        isn't available, silently skip with a stderr warning."""
        if not self._mnemo_available:
            return
        try:
            subprocess.run([
                "mnemo", "hdu", "save", entry.id,
                "--title", entry.title,
                "--phase", entry.phase,
                "--status", entry.status,
                "--project", self._project,
                "--content", entry.executive_summary or entry.title,
            ], capture_output=True, timeout=5)
        except (FileNotFoundError, subprocess.TimeoutExpired):
            pass  # fail-open
        except Exception:
            print(
                f"[hdu_tracker] warning: mnemo sync failed for {entry.id}",
                file=sys.stderr,
            )

    def exists(self, hdu_id: str) -> bool:
        return self._path(hdu_id).exists()

    def get(self, hdu_id: str) -> Optional[HDUEntry]:
        p = self._path(hdu_id)
        if not p.exists():
            return None
        data = yaml.safe_load(p.read_text()) or {}
        return HDUEntry.from_dict(data)

    def save(self, entry: HDUEntry) -> None:
        entry.updated_at = datetime.now().strftime("%Y-%m-%d %H:%M")
        entry.progress = entry.calc_progress()
        p = self._path(entry.id)
        p.write_text(yaml.dump(entry.to_dict(), default_flow_style=False, allow_unicode=True, sort_keys=False))
        self._sync_to_mnemo(entry)

    def load_all(self) -> list[HDUEntry]:
        if not self.store_dir.exists():
            return []
        entries = []
        for f in sorted(self.store_dir.glob("*.yaml")):
            data = yaml.safe_load(f.read_text())
            if data:
                entries.append(HDUEntry.from_dict(data))
        return entries

    def create(self, title: str, hdu_id: str, question: Optional[str] = None) -> HDUEntry:
        entry = HDUEntry(
            id=hdu_id,
            title=title,
            status="blocked" if question else "active",
        )
        if question:
            entry.blockers.append(BlockingQuestion(
                question=question,
                created_at=datetime.now().strftime("%Y-%m-%d %H:%M"),
                blocked_by="supervisor",
            ))
        self.save(entry)
        return entry

    def add_blocker(self, hdu_id: str, question: str, context: str = "", blocked_by: str = "") -> Optional[HDUEntry]:
        entry = self.get(hdu_id)
        if not entry:
            return None
        entry.blockers.append(BlockingQuestion(
            question=question,
            context=context,
            created_at=datetime.now().strftime("%Y-%m-%d %H:%M"),
            blocked_by=blocked_by,
        ))
        entry.status = "blocked"
        self.save(entry)
        return entry

    def answer_blocker(self, hdu_id: str, answer: str) -> Optional[HDUEntry]:
        entry = self.get(hdu_id)
        if not entry:
            return None
        for b in entry.blockers:
            if not b.is_resolved():
                b.answer = answer
                b.answered_at = datetime.now().strftime("%Y-%m-%d %H:%M")
                break
        if not entry.is_blocked():
            entry.status = "active"
        self.save(entry)
        return entry

    def update_phase(self, hdu_id: str, phase: str) -> Optional[HDUEntry]:
        entry = self.get(hdu_id)
        if not entry:
            return None
        entry.phase = phase
        if phase not in entry.completed_phases and phase != "init":
            all_phases = list(PHASE_WEIGHTS.keys())
            current_idx = all_phases.index(phase) if phase in all_phases else 0
            for p in all_phases[:current_idx]:
                if p not in entry.completed_phases:
                    entry.completed_phases.append(p)
        self.save(entry)
        return entry

    def mark_completed(self, hdu_id: str) -> Optional[HDUEntry]:
        entry = self.get(hdu_id)
        if not entry:
            return None
        entry.status = "completed"
        entry.completed_phases = list(PHASE_WEIGHTS.keys())
        self.save(entry)
        return entry

    def delete(self, hdu_id: str) -> bool:
        p = self._path(hdu_id)
        if p.exists():
            p.unlink()
            return True
        return False
