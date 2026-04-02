---
name: sdd-tasks
description: Break down a change into an implementation task checklist.
trigger: When the orchestrator launches you to create or update the task breakdown for a change
version: 1.0.0
---

## Purpose

Create the TASK BREAKDOWN from spec + design. Produce concrete, ordered implementation steps.

## Dependencies

- **Required**: `sdd/{change-name}/spec`
- **Optional**: `sdd/{change-name}/design`

## Protocol

### Step 1: Load Dependencies

Retrieve via two-step (see `_shared/persistence.md`):

```
spec_id = mem_search(query: "sdd/{change-name}/spec", project: "{project}")
spec = mem_get_observation(id: spec_id)

design_id = mem_search(query: "sdd/{change-name}/design", project: "{project}")  # optional
if design_id: design = mem_get_observation(id: design_id)
```

### Step 2: Analyze

From spec/design, identify:
- All files to create/modify/delete
- Dependency order (what must come first)
- Testing requirements per component

### Step 3: Write Task Breakdown

```markdown
# Tasks: {Change Title}

## Phase 1: Foundation
- [ ] 1.1 {Specific action — file + what change} [S|M|L]
- [ ] 1.2 {Specific action} [S]

## Phase 2: Core Implementation
- [ ] 2.1 {Specific action} [M]
- [ ] 2.2 {Specific action} [L]

## Phase 3: Integration
- [ ] 3.1 {Connect components, wiring} [M]

## Phase 4: Testing
- [ ] 4.1 {Test for scenario X} [S]
- [ ] 4.2 {Test for scenario Y} [M]
```

### Task Rules

| Rule | Good | Bad |
|------|------|-----|
| Specific | "Create `auth/middleware.go` with JWT validation" | "Add auth" |
| Actionable | "Add `ValidateToken()` to `AuthService`" | "Handle tokens" |
| Verifiable | "Test: `POST /login` returns 401 without token" | "Make it work" |
| Small | One file or logical unit | "Implement the feature" |

### Phase Guidelines

```
Phase 1: Foundation    → Types, interfaces, config, DB changes
Phase 2: Core          → Main logic, business rules
Phase 3: Integration   → Connect components, routes, UI wiring
Phase 4: Testing       → Unit, integration, e2e tests
Phase 5: Cleanup       → Docs, remove dead code (if needed)
```

### Step 4: Persist

```javascript
mem_save(
  title: "sdd/{change-name}/tasks",
  topic_key: "sdd/{change-name}/tasks",
  type: "architecture",
  project: "{project}",
  content: "{full task breakdown}"
)
```

### Step 5: Return

Follow `_shared/phase-common.md` return envelope:

```
## Status: success

## Executive Summary
{N} tasks across {M} phases for {change-name}. Ready for apply.

## Artifacts
- tasks: sdd/{change-name}/tasks

## Next Recommended
sdd-apply

## Risks
{None | concerns}
```

## Rules

- Reference concrete file paths in every task
- Tasks ordered by dependency (Phase 1 shouldn't depend on Phase 2)
- Each task completable in ONE session
- Hierarchical numbering: 1.1, 1.2, 2.1, 2.2
- If TDD: integrate test-first tasks (RED → GREEN → REFACTOR)
- NO vague tasks like "implement feature"
