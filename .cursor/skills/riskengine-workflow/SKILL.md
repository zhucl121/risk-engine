---
name: riskengine-workflow
description: Spec-driven development workflow for riskengine. Fuses OpenSpec (specs as source of truth, agree before you code) with Superpowers (Architect/Implementor/Reviewer sequential roles). Read this skill when asked to implement a feature, add a rule/model/fetcher, plan an architecture change, run /opsx: commands, or generate proposal/design/tasks documents.
---

# RiskEngine — Spec-Driven Workflow

## Design Principle: Why Two Mechanisms?

| Mechanism | Purpose | When active |
|-----------|---------|-------------|
| **Rule** `dev-workflow.mdc` | Hard gates — STOP and enforce | Always (every session) |
| **This Skill** | Complete executable guide — HOW to do it | On demand (read when needed) |

The Rule says **what not to do**. This Skill says **exactly how to do it right**.

---

## The 3 Roles (Superpowers-style)

Each `/opsx:` command activates a different role with strict output boundaries.

### 🏗️ Architect (`/opsx:review`)
**Input**: `proposal.md` + `specs/requirements.md`  
**Output**: REVIEW section appended to `design.md` — structured, concise, no code  
**Focus**: interface stability, performance budget, failure modes, alternatives  
**Checklist**:
- Does any interface in `engine/rule/feature/model/list` change? → flag BREAKING
- Does this add hot-path allocations? → require benchmark targets in design.md
- Is there a simpler alternative that fits the existing abstractions?
- Are all open `?` questions resolved before signing off?

### ⚙️ Implementor (`/opsx:apply`)
**Input**: first unchecked `[ ]` task in `tasks.md`  
**Output**: Go code for ONE task + mark `[x]` + one commit  
**Focus**: correctness, spec compliance, `make test && make lint` passing  
**Constraints**: see `riskengine-domain.mdc` (always active).

### 🔍 Reviewer (after all tasks `[x]`)
**Input**: all changed files + `specs/requirements.md`  
**Output**: `REVIEW.md` in the change folder OR inline comments  
**Focus**: spec compliance, no regressions, benchmark numbers recorded  

---

## Commands — What Each Does

### `/opsx:new <name>`
Create the change scaffold. Stop here; do not fill files yet.

```
openspec/changes/<name>/
├── proposal.md        ← WHY (fill first, must be agreed before design)
├── specs/
│   └── requirements.md  ← WHAT (functional + non-functional)
├── design.md          ← HOW (technical approach + ADR + REVIEW section)
└── tasks.md           ← atomic ordered checklist
```

### `/opsx:ff` — Fast-Forward
Generate all four planning documents in one pass. Use when the feature is clear enough to plan without iteration.

Generation order (strict):
1. `proposal.md` — motivation, goals, non-goals, success criteria, risk
2. `specs/requirements.md` — Given/When/Then scenarios + NFR table
3. `design.md` — approach, component impact table, interface delta, alternatives
4. `tasks.md` — ordered atomic tasks (one file change per task)

**Task quality rules**:
- Each task changes ≤ 1 file OR creates 1 new file
- Each task is independently verifiable with `make test`
- Tasks are ordered: no task depends on a later task

### `/opsx:review`
Switch to Architect role. Append this structure to `design.md`:

```markdown
## REVIEW (Architect)
Date: YYYY-MM-DD

### Open Questions
- [ ] Q1: ...

### Risks
- Risk: ... | Mitigation: ...

### Interface Impact
- engine.Engine: no change / BREAKING: ...

### Performance Budget
- Expected P99 impact: +Xms (acceptable / needs benchmark)

### Decision
[ ] GO — proceed to /opsx:apply
[ ] NO-GO — resolve questions above first
```

### `/opsx:apply`
Switch to Implementor role. Strict loop:

```
WHILE unchecked tasks remain in tasks.md:
  1. Read next [ ] task
  2. Implement (one file only)
  3. make test && make lint
  4. PASS → mark [x] → commit with "feat(scope): TASK-N description"
  5. FAIL → fix, return to step 3
  6. If scope exceeds the task → add task to tasks.md bottom, resume original task
```

### `/opsx:archive`
```bash
mv openspec/changes/<name> openspec/changes/archive/$(date +%Y-%m-%d)-<name>/
# Then: merge updated requirements into openspec/specs/<component>.md
# Then: add entry to CHANGELOG.md under [Unreleased]
```

---

## Scenario Decision Map

| What you want | Command sequence | Architect required? |
|--------------|-----------------|-------------------|
| Add a fraud rule | `new` → `ff` → `apply` | No (unless new interface) |
| Add a feature fetcher | `new` → `ff` → `apply` | No |
| Add an ML model | `new` → `ff` → `review` → `apply` | Yes (ONNX contract) |
| Change engine interface | `new` → `ff` → `review` → `apply` | **Mandatory** |
| Fix a bug (root cause) | `new` → `ff` → `apply` | No |
| Fix a bug (trivial) | Direct `apply` with inline task | No |
| Performance optimization | `new` → `ff` (include benchmark targets) → `review` → `apply` | Yes |
| Add API endpoint | `new` → `ff` → `apply` | No |

---

## Spec Templates (copy-paste ready)

### proposal.md
```markdown
# Proposal: <Feature Name>

## Motivation
<!-- What problem does this solve? Why now? -->

## Goals
- [ ] Goal 1

## Non-Goals
<!-- Explicitly excluded scope -->

## Success Criteria
- Performance: P99 < Xms for <component>
- Correctness: <specific behaviour>
- Observability: metric `riskengine_<subsystem>_<event>_total` added

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|-----------|
| | | |
```

### specs/requirements.md
```markdown
# Requirements: <Feature Name>

## Functional Requirements

### FR-1: <Name>
**Given** <precondition>
**When** <trigger>
**Then** <expected result>

## Non-Functional Requirements
| Attribute | Target |
|-----------|--------|
| Latency (P99) | < Xms |
| Throughput | no regression |
| Hot-reload | < 30s propagation |

## Acceptance Criteria
- [ ] All FRs covered by table-driven tests
- [ ] BenchmarkXxx meets NFR targets
- [ ] `make lint` zero warnings
- [ ] CHANGELOG updated
```

### design.md
```markdown
# Design: <Feature Name>

## Approach
<!-- One paragraph: what we're building and why this approach -->

## Component Changes
| Component | Change | Reason |
|-----------|--------|--------|
| `internal/rule/` | Add XxxRule | FR-1 |

## Interface Delta
- `engine.Engine`: no change
- `rule.Rule`: no change

## Alternatives Considered
### Option A (chosen): ...
### Option B (rejected): ... Reason: ...

## ADR
### ADR-001: <Decision>
Context / Decision / Consequences

## Benchmark Targets
| Path | Current P99 | Target P99 |
|------|------------|-----------|
| | | |

## REVIEW (Architect)
<!-- Filled during /opsx:review -->
```

### tasks.md
```markdown
# Tasks: <Feature Name>
<!-- Implementor: one task at a time. Mark [x] only after make test passes. -->
<!-- Commit format: feat(scope): TASK-N description -->

## Phase 1: Foundations
- [ ] TASK-1: Add feature keys to `internal/feature/keys.go`
- [ ] TASK-2: Add config field + default in `internal/config/config.go`

## Phase 2: Core
- [ ] TASK-3: Implement `XxxRule` in `internal/rule/rules/xxx.go`
- [ ] TASK-4: Register in `internal/rule/registry.go`
- [ ] TASK-5: Add YAML entry in `configs/rules/<group>.yaml`

## Phase 3: Tests
- [ ] TASK-6: Table-driven unit tests in `internal/rule/rules/xxx_test.go`
- [ ] TASK-7: `BenchmarkEvaluateXxx` — target < 500ns/op

## Phase 4: Wrap-up
- [ ] TASK-8: Update `CHANGELOG.md`
- [ ] TASK-9: Run `/opsx:archive`
```

---

## Implementation Patterns (Implementor quick reference)

### New Rule
```
TASK A: internal/rule/rules/<name>.go — stateless struct, feature.Keys constants, early-return on miss
TASK B: internal/rule/registry.go — register
TASK C: configs/rules/<group>.yaml — YAML entry
TASK D: _test.go — table-driven + BenchmarkEvaluate (target < 500ns/op)
```

### New Feature Fetcher
```
TASK A: internal/feature/fetchers/<name>.go — Timeout() from config, returns nil on ctx.Err()
TASK B: internal/feature/registry.go — register
TASK C: feature/keys.go — new keys if needed
TASK D: _test.go — mock client, test timeout path, BenchmarkFetch
```

### New ONNX Model
```
TASK A: configs/models/registry.yaml — entry with inputFeatures, outputKey
TASK B: internal/model/testdata/ — tiny test model for integration test
TASK C: configs/config.example.yaml — champion-challenger example
NOTE: .onnx file goes to object storage, NOT git
```

### Interface Change (post Architect review only)
```
TASK A: internal/<pkg>/<pkg>.go — update interface
TASK B: all implementations — update every impl
TASK C: make mock — regenerate mocks
TASK D: .cursor/skills/go-riskengine/reference.md — update interface contracts
TASK E: CHANGELOG — note BREAKING: if method removed or renamed
```

---

> Openspec folder layout → `openspec/README.md`  
> Domain coding constraints → `.cursor/rules/riskengine-domain.mdc`
