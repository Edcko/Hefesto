---
name: sdd-apply
description: Implement tasks from the change, writing actual code following specs.
trigger: When the orchestrator launches you to implement one or more tasks from a change
version: 1.0.0
---

## Purpose

IMPLEMENT assigned tasks. Write actual code following spec + design strictly.

## Dependencies

- **Required**: `sdd/{change-name}/tasks`, `sdd/{change-name}/spec`
- **Optional**: `sdd/{change-name}/design`

## Protocol

### Step 1: Load Dependencies

Retrieve via two-step (see `_shared/persistence.md`):

```
tasks_id = mem_search(query: "sdd/{change-name}/tasks", project: "{project}")
tasks = mem_get_observation(id: tasks_id)

spec_id = mem_search(query: "sdd/{change-name}/spec", project: "{project}")
spec = mem_get_observation(id: spec_id)

design_id = mem_search(query: "sdd/{change-name}/design", project: "{project}")
if design_id: design = mem_get_observation(id: design_id)
```

### Step 2: Read Context

Before writing code:
1. Read spec — understand WHAT to build
2. Read design — understand HOW to structure
3. Read existing code — match project patterns

### Step 3: Implement Tasks

For each assigned task:

```
FOR EACH TASK:
├── Read task description
├── Read relevant spec scenarios (acceptance criteria)
├── Read design decisions (constraints)
├── Read existing code patterns
├── Write the code
├── Mark task complete [x]
└── Note deviations/issues
```

#### TDD Support

If project uses TDD (detect from existing test patterns):

```
FOR EACH TASK:
├── RED: Write failing test first
├── GREEN: Write minimum code to pass
├── REFACTOR: Clean up while keeping tests green
└── Run relevant tests only (not full suite)
```

**Test command**: Use whatever the project already has. If unsure, ask orchestrator.

### Step 4: Mark Complete

Update tasks artifact:

```javascript
mem_update(id: tasks_id, content: "{updated tasks with [x] marks}")
```

### Step 5: Persist Progress

```javascript
mem_save(
  title: "sdd/{change-name}/apply-progress",
  topic_key: "sdd/{change-name}/apply-progress",
  type: "architecture",
  project: "{project}",
  content: "
## Implementation Progress

**Change**: {change-name}
**Mode**: {TDD | Standard}

### Completed
- [x] {task 1.1}
- [x] {task 1.2}

### Files Changed
| File | Action | What |
|------|--------|------|
| path/to/file | Created | {brief} |

### Deviations
{None | list deviations and why}

### Remaining
- [ ] {next task}

### Status
{N}/{total} tasks complete.
"
)
```

### Step 6: Return

Follow `_shared/phase-common.md` return envelope:

```
## Status: success | partial | blocked

## Executive Summary
{N}/{total} tasks implemented. {Ready for next batch | Blocked by X}

## Artifacts
- apply-progress: sdd/{change-name}/apply-progress

## Next Recommended
{sdd-apply (more tasks) | sdd-verify (all done)}

## Risks
{None | blockers | design issues found}
```

## Rules

- ALWAYS read spec before implementing — specs are acceptance criteria
- ALWAYS follow design — don't freelance a different approach
- ALWAYS match existing code patterns
- Mark tasks complete AS you go
- If design is wrong, REPORT IT — don't silently deviate
- If blocked, STOP and report — don't guess
- NEVER implement unassigned tasks
