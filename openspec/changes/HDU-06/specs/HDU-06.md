# Spec: God-Tier Multi-Agent Ecosystem

## BDD Scenarios

### Scenario 1: Swarm auto-recovers from agent death
```gherkin
Given the swarm orchestrator is running HDU-06 with 4 agents active
And agent "apply-2" has claimed task "task-5" with heartbeat every 5s
When agent "apply-2" crashes and heartbeats stop for 15s
Then the orchestrator detects the dead agent
And task "task-5" is re-queued as "pending"
And another available agent claims it within 10s
And the user sees: "Agent apply-2 lost. Task-5 reassigned to apply-3."
```

### Scenario 2: Skill auto-evolves from repeated success
```gherkin
Given the static skill "go-testing" has been applied 5 times
And all 5 applications have outcome="resolved" in mnemo
When the skill registry runs its nightly evolution cycle
Then "go-testing" is promoted to tier "proven"
And an evolution proposal is generated with 3 improvements learned from the 5 sessions
And the old version is archived to vec_releases
```

### Scenario 3: Adversarial testing catches edge case
```gherkin
Given the verify phase is running on HDU-06's auth module
And all spec scenarios pass with normal inputs
When the adversarial tester generates 200 mutated inputs
Then one mutation triggers a nil pointer: empty email field bypasses validation
And causal analysis traces it to design decision "email validation is optional"
And the bug is auto-fixed with input sanitization
And full test suite re-runs and passes
```

### Scenario 4: Session replay resumes interrupted work
```gherkin
Given the agent was working on HDU-06 task-3 and saved 7 progress memories
And the session was killed mid-edit on file "swarm/orchestrator.go"
When a new agent starts and searches for progress
Then it finds the 7-action progress chain
And reconstructs that "swarm/orchestrator.go" was being edited
And reads the before_hash and diff_summary from the last checkpoint
And resumes work showing: "Recovered 7 actions. Last was editing swarm/orchestrator.go. Continuing..."
```

### Scenario 5: Persona follows user across tools
```gherkin
Given the user has worked 20 sessions in Claude Code generating 200+ persona signals
And the persona pack was exported via mnemo pack export persona --output my-persona.mempack
When the user opens Antigravity and imports the persona pack
And the agent runs mem_persona_search("code style preferences")
Then it retrieves: tabs over spaces, short variable names, functional composition
And the agent adapts its code generation accordingly
And the user sees code in their preferred style without configuring anything
```

### Scenario 6: Predictive pre-load saves 15 minutes of context gathering
```gherkin
Given the user is on branch "feature/add-2fa"
And the last 3 edits were in "auth/" directory
And mnemo has 12 memories about auth from 4 previous projects
When the user opens their agent and runs gingx-sdd orchestrate
Then the predictor auto-loads 8 relevant memories before the user types anything
And the agent greets: "Working on 2FA. I found 8 relevant memories including 2 bugs from HDU-03."
And the context is pre-warmed, saving 15 minutes of manual search
```

### Scenario 7: Autonomous improvement proposes HDU from failure patterns
```gherkin
Given the system has recorded 8 auth-related bugs across 3 HDUs in 2 weeks
And all 8 bugs cluster with cosine similarity > 0.85
When the evolve analyzer runs its weekly cycle
Then it detects the pattern: "Auth bugs cluster around token expiry edge cases"
And generates improvement proposal HDU: "Strengthen token expiry testing in verify phase"
And the HDU is saved to openspec/changes/HDU-auto-strengthen-auth-testing/
And the user is notified: "I found a pattern. Proposed HDU-auto-strengthen-auth-testing for your review."
```

### Scenario 8: Self-healing detects and merges duplicate memories
```gherkin
Given mnemo has 3 memories with cosine similarity > 0.99
And they have slightly different titles but identical content
When mnemo heal --full is run
Then the 3 duplicates are detected
And the 2 older ones are merged into the newest
And a release note is created: "Auto-merged 2 duplicate memories on token expiry"
And vec_memories now has 1 consolidated entry instead of 3
```

### Scenario 9: Contradictory knowledge flagged for human review
```gherkin
Given memory A says "Use JWT with 15min expiry" (outcome=resolved)
And memory B says "Use JWT with 24h expiry" (outcome=resolved)
And they have cosine similarity > 0.9 on the same project
When mnemo heal runs contradiction detection
Then the contradiction is flagged
And the user sees: "Contradiction: short vs long JWT expiry. Which is correct?"
And the user selects A as authoritative
And B is updated with outcome="superseded" and linked to A
```
