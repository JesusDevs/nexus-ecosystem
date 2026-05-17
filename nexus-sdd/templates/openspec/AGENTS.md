# OpenSpec Instructions

Slash commands for AI coding tools integrated with Nexus-SDD.

## Standard Workflow
- `/opsx:propose` — Create a new change (proposal + specs + design + tasks)
- `/opsx:apply` — Implement the tasks
- `/opsx:archive` — Archive a completed change

## Nexus-SDD Extended Workflow
- `/opsx:new` — Start a fresh change
- `/opsx:continue` — Resume a partially completed change
- `/opsx:ff` — Fast-forward through simple changes
- `/opsx:verify` — Run verification (tests + security scan)
- `/opsx:onboard` — Onboard a new developer

## Nexus-SDD Integration

1. **SPEC phase**: `/opsx:propose` → generates `openspec/changes/<HDU>/`
2. **BDD**: Every spec must include Gherkin scenarios (Given/When/Then)
3. **PLAN phase**: Plan agent reads proposal and generates `plan.md`
4. **CODE phase**: Dev agent implements `tasks.md` item by item
5. **TEST phase**: QA agent verifies all tests pass, including BDD
6. **SECURITY phase**: Scans for secrets before allowing archive

After archive, the change moves to `openspec/changes/archive/YYYY-MM-DD-<HDU>/`
and mnemo saves the memory:
```bash
mnemo save "Completed <HDU>" "<summary>" --type milestone
```
