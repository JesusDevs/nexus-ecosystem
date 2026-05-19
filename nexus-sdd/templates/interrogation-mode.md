# Proactive Interrogation Mode

---
depth: basic
---

## Basic Interrogation

Before acting, ask these 3 clarifying questions:

1. **Scope boundary**: What exactly is in scope? What's explicitly OUT of scope?
2. **Priority constraints**: Time, quality, tools — which is the dominant constraint?
3. **Success criteria**: How will we know this is done? What's the acceptance threshold?

---
depth: deep
---

## Deep Interrogation — Tony Stark Mode

MANDATORY before any action. Do not skip. Do not assume.

### Step 1: Search Memory
Search mnemo for context: `mnemo search "<task>" --project {project} --limit 5`

### Step 2: Ask Clarifying Questions
If ANY of these are unclear, STOP and ask before proceeding:

1. **Scope**: What's the minimum viable version (MVP) vs the full vision? Where's the boundary?
2. **Constraints**: Non-functional requirements? Latency target? Scale expectation? Security level?
3. **Dependencies**: What systems, services, or components does this change touch?
4. **Patterns**: Has something similar been done before in this codebase? What did we learn?
5. **Success**: How do we measure completion? What's the acceptance threshold? What tests?

### Step 3: Propose Options
Before executing, lay out what's possible:

> "Here's what we could achieve."
>
> **MVP** (minimum to deliver value): <scope>
> **Full scope** (complete vision): <scope>
> **Recommended** (given constraints): <recommendation>
>
> To achieve the full vision, we'd need: <prerequisites>
>
> Which path should we take?

### Core Rule
NEVER assume context. If you don't know, ASK.
If unsure, STATE your assumption explicitly and ask for confirmation.
Your job is to ILLUMINATE options, not gamble on understanding.

---
depth: exhaustive
---

## Exhaustive Interrogation — Full Spec Gate

Apply ALL 5 gates from `spec-gate.md` before any action.

### Gate 1: Problem Definition
- [ ] Is there a clearly defined USER problem?
- [ ] Do we know WHO has this problem?
- [ ] Is success MEASURABLE?
- [ ] Verified no existing solution in codebase?
- [ ] Searched mnemo for related decisions?

### Gate 2: Context Gathering
1. What classes/components already exist that relate to this?
2. What patterns were used in similar features?
3. What trade-offs apply in this specific case?
4. What dependencies (internal/external) are affected?
5. What's the minimum viable scope vs full scope?

### Gate 3: Spec Artifacts
Before code, verify these exist:
```
openspec/changes/<HDU>/
├── proposal.md
├── specs/
├── design.md
└── tasks.md
```

### Gate 4: Smart Questions by Change Type
| Type | Question |
|------|----------|
| New feature | What patterns exist? What did prior HDUs teach us? |
| UI change | What components are reused? What visual patterns apply? |
| API/DB change | What services depend on this? Migration plan? |
| Bug fix | What was the original decision? Why was it done that way? |
| Refactor | Which HDU introduced this code? What trade-offs were documented? |

### Gate 5: Delivery Strategy
- [ ] Single PR (small scope, <= 400 lines, <= 3 areas)
- [ ] Stacked PRs (large feature, chain into linked PRs)
- [ ] Feature track (empty branch as target)

### Core Rule
NEVER assume context. If you don't know, ASK.
If a gate fails: DOCUMENT the blockage. DO NOT proceed to code.
