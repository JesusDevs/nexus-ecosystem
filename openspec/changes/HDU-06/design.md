# Design: God-Tier Multi-Agent Ecosystem

## 1. Multi-Agent Swarm Orchestrator (`swarm/`)

### Architecture
```
┌─────────────────────────────────────────────────┐
│  SWARM ORCHESTRATOR (task queue + heartbeat)     │
│                                                  │
│  Responsibilities:                               │
│  • Maintain shared task queue (SQLite-backed)    │
│  • Spawn phase agents with fresh context         │
│  • Monitor heartbeats (5s intervals)             │
│  • Re-assign dead agents' tasks                 │
│  • Resolve DAG dependencies before dispatch      │
│  • Report progress to user                      │
│  • Auto-scale: spawn more agents for parallel    │
│    phases (spec∥design, multi-file apply)       │
└──────────┬──────────────────────────────────────┘
           │
    ┌──────┴──────────────────────────────┐
    │         Shared Task Queue            │
    │  ┌────────┐ ┌────────┐ ┌────────┐   │
    │  │CLAIMED │ │PENDING │ │ DONE   │   │
    │  │(heart- │ │(open)  │ │(result)│   │
    │  │ beat)  │ │        │ │        │   │
    │  └────────┘ └────────┘ └────────┘   │
    └──────────────────────────────────────┘
           │
    ┌──────┴──────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼
 EXPLORE      PROPOSE     SPEC        DESIGN
 (agent)      (agent)    (agent)     (agent)
    │             │          │          │
    └─────────────┴──────────┴──────────┘
           │
    ┌──────┴──────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼
  TASKS        APPLY#1    APPLY#2    VERIFY
 (agent)      (agent)    (agent)    (agent)
    │             │          │          │
    └─────────────┴──────────┴──────────┘
           │
           ▼
       ARCHIVE
       (agent)
```

### Task Queue Schema
```sql
CREATE TABLE swarm_tasks (
    id TEXT PRIMARY KEY,
    phase TEXT NOT NULL,          -- explore|propose|spec|design|tasks|apply|verify|archive
    hdu_id TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending|claimed|done|failed
    agent_id TEXT,
    heartbeat_at TEXT,
    result_envelope TEXT,         -- JSON result contract
    dependencies TEXT,            -- JSON array of task IDs
    created_at TEXT DEFAULT (datetime('now')),
    claimed_at TEXT,
    completed_at TEXT
);
```

### Result Envelope Contract
```json
{
  "status": "success|partial|blocked|failed",
  "executive_summary": "1-3 sentences",
  "artifacts": ["artifact keys / file paths"],
  "memories_saved": ["memory ids"],
  "next_recommended": "phase name or none",
  "risks": ["risk descriptions"],
  "learnings": ["things the system should remember"]
}
```

## 2. Self-Evolving Skill Registry (`registry/`)

Skills start as static `.md` files but evolve:

```
Static Skill (.md)
    ↓
Agent applies skill successfully
    ↓
mem_save(outcome="resolved", type="skill_application")
    ↓
Registry detects 5+ successful applications
    ↓
Skill promoted to "proven" tier
    ↓
Agent modifies skill based on accumulated learnings
    ↓
mem_save(type="skill_evolution", content="improved version")
    ↓
Old skill versioned, new version active
```

### Skill Tiers
| Tier | Criteria | Behavior |
|------|----------|----------|
| `static` | Just installed | Follow exactly |
| `proven` | 5+ successes | Follow, minor adaptations allowed |
| `adaptive` | 20+ successes, 0 failures | Agent can adapt significantly |
| `deprecated` | 3+ failures | Flagged for review, not auto-loaded |

### Auto-Generation
```go
// registry/generator.go
func (r *Registry) GenerateCandidateSkills(project string) ([]SkillCandidate, error) {
    // Find memory clusters with high success rate
    // Extract common patterns from content
    // Generate skill template with BDD scenarios
    // Surface for human approval
}
```

## 3. Granular Progress + Session Replay (`progress/`)

### Progress Chain
Every agent action creates a progress memory:
```json
{
  "id": "prog-HDU-06-task-3-2026-05-15T10:30:00Z",
  "type": "progress",
  "hdu_id": "HDU-06",
  "phase": "apply",
  "task_id": "3",
  "action": "file_edit",
  "file": "swarm/orchestrator.go",
  "before_hash": "abc123",
  "after_hash": "def456",
  "diff_summary": "Added heartbeat goroutine",
  "test_result": "pending",
  "parent_progress": "prog-HDU-06-task-3-2026-05-15T10:29:00Z"
}
```

### Replay Protocol
```
Agent starts → detects no active session
    ↓
mem_search_semantic("HDU-06 progress in_progress")
    ↓
Finds progress chain: 7 actions, last was "file_edit swarm/orchestrator.go"
    ↓
Reconstructs state: which files were modified, tests ran, what's pending
    ↓
Resumes from last checkpoint
```

## 4. Evolutionary Judgment Day (`judgment/`)

### Beyond ATL's Judgment Day

ATL: 2 agents review code independently → compare → fix loop.

Gingx Evolutionary:
```
Phase 1: SPEC COMPLIANCE
  - Map every spec scenario to test execution
  - Report coverage gap if any scenario untested

Phase 2: ADVERSARIAL TESTING
  - Mutate inputs (fuzz testing)
  - Generate counterfactual scenarios
  - Inject simulated failures (network, disk, memory)

Phase 3: CAUSAL ANALYSIS
  - If bug found, trace to root decision in design
  - mem_search_semantic("similar bugs in other projects")
  - Report: "This bug pattern appeared in HDU-03, fixed with X"

Phase 4: AUTO-FIX + RE-VERIFY
  - Generate fix following the design's intent
  - Run full test suite (original + adversarial)
  - If passes: certify. If fails: escalate to human with full causal trace.
```

### Certification Levels
| Level | Criteria | Impact |
|-------|----------|--------|
| `certified` | All specs pass + adversarial suite passes | Auto-merge allowed |
| `provisional` | All specs pass, adversarial has warnings | Manual review recommended |
| `rejected` | Spec failure or critical adversarial finding | Blocked, causal report attached |

## 5. Adaptive Agent Persona (`persona/`)

### Persona Memory
```
User actions → mem_save(type="persona_signal", content="preferred tabs over spaces")
Agent observes → mem_search_semantic("user preferences")
Agent adapts → proposes solutions aligned with user's historical choices
```

### Cross-Tool Identity
Persona is stored in mnemo's portable `.mempack`. Export once, import everywhere:
```
Claude Code ──┐
Antigravity ──┼── mem_persona_search("user prefers X") → same answer
Codex ────────┘
```

### Persona Dimensions
| Dimension | Signals |
|-----------|---------|
| Code style | tabs/spaces, naming conventions, comment density |
| Architecture | monolith/microservice, FP/OOP, test-first/test-after |
| Risk tolerance | aggressive refactors / conservative changes |
| Communication | verbose/terse, technical/non-technical |
| Trust level | auto-approve / always review |

## 6. Self-Healing Infrastructure (`heal/`)

### Health Checks (run daily or on-demand)
```bash
mnemo heal --full
# → Memory consistency: 342/342 ok
# → Stale memories (>90 days, unreferenced): 12 flagged
# → Contradictory memories (same topic, opposite outcome): 2 flagged
# → Unused skills: 3 flagged for deprecation
# → DB integrity: ok, indexes optimized
# → Model performance: embedding latency p50=3ms, p99=15ms
```

### Auto-Healing Actions
| Issue | Auto-Action |
|-------|------------|
| Duplicate memories (cosine > 0.99) | Merge, keep newest |
| Contradictory outcomes | Flag for human, link both |
| Stale skill (no use in 90 days) | Deprecate, archive |
| DB fragmentation | Auto-VACUUM |
| Embedding model outdated | Suggest upgrade, regenerate cache |

## 7. Predictive Knowledge Pre-Loading (`predictor/`)

### Signals for Prediction
```go
type PredictionSignals struct {
    GitBranch     string   // "feature/add-2fa"
    RecentEdits   []string // files changed in last hour
    OpenIssues    []string // from GitHub/Linear
    CurrentHDUs   []string // active HDUs in openspec/changes/
    UserActivity  string   // "editing auth module"
}
```

### Pre-Load Pipeline
```
Signals collected
    ↓
mem_search_semantic(gitBranch + recentEdits)
    ↓
Results: "Here's how auth was done in 3 other projects"
       + "2 bugs related to 2FA in HDU-03"
       + "Design decision: use TOTP, not SMS"
    ↓
Agent starts with context already loaded
    ↓
"Welcome back. You're working on 2FA. I found 5 relevant memories
 from 3 projects. Ready to continue where HDU-06-task-3 left off."
```

## 8. Autonomous Improvement Loops (`evolve/`)

### Self-Improvement Cycle
```
System collects failure patterns across sessions
    ↓
Clusters similar failures (cosine > 0.85)
    ↓
If cluster size > 5: generate improvement proposal
    ↓
Proposal auto-saved as HDU in openspec/changes/
    ↓
Human reviews: "System detected 8 auth-related bugs in 2 weeks.
                 Proposed: strengthen auth testing in verify phase"
    ↓
Human approves → skill updated → system improved itself
```

### Meta-Cognition
The system maintains `meta_memories` — memories about its own performance:
- "HDU-06 took 3x longer than estimated because..."
- "The spec phase was redundant here — proposal was detailed enough"
- "Adversarial testing caught 2 bugs that normal testing missed"

These inform future task estimation and phase optimization.

## Architecture Overview

```
gingx-mnemo/
├── vec/store.go          # Vector DB (existing, extended with new tables)
├── vec/embed.go          # Embeddings (existing)
├── mcp/server.go         # MCP server (existing, 14 tools)
├── main.go               # CLI (existing, extended)
│
├── swarm/                # NEW: Multi-agent orchestration
│   ├── orchestrator.go   # Task queue, heartbeat, dispatch
│   ├── task.go           # Task type + result envelope
│   └── agent.go          # Agent lifecycle management
│
├── registry/             # NEW: Self-evolving skills
│   ├── registry.go       # Skill database + tier management
│   ├── generator.go      # Auto-skill generation from memories
│   └── loader.go         # Skill injection for sub-agents
│
├── progress/             # NEW: Granular progress tracking
│   ├── tracker.go        # Progress chain recording
│   └── replay.go         # Session reconstruction
│
├── judgment/             # NEW: Evolutionary verification
│   ├── verifier.go       # Multi-phase verification
│   ├── adversarial.go    # Mutation + fuzz testing
│   └── causal.go         # Root cause analysis
│
├── persona/              # NEW: Adaptive agent persona
│   ├── persona.go        # Persona memory + dimensions
│   └── inference.go      # Style/preference prediction
│
├── heal/                 # NEW: Self-healing
│   ├── checker.go        # Health check suite
│   └── autoheal.go       # Automated fixes
│
├── predictor/            # NEW: Predictive pre-loading
│   └── predictor.go      # Signal collection + pre-load
│
└── evolve/               # NEW: Autonomous improvement
    ├── analyzer.go        # Pattern detection
    └── proposer.go        # Auto-HDU generation
```
