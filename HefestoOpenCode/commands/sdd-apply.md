---
agent: sdd-orchestrator
description: Implement tasks from the change — write code following specs and design
---

## Context

- **Change Name**: {args} (if empty, use active change from recent apply-progress)
- **Working Directory**: {cwd}
- **Project**: {project}
- **Persistence**: Engram only

## Task

Implement tasks for change `{args}`:

1. **Load Artifacts** (CRITICAL: use `mem_get_observation` for full content):
   - `mem_search(query: "sdd/{args}/spec")` → get ID → `mem_get_observation(id)`
   - `mem_search(query: "sdd/{args}/tasks")` → get ID → `mem_get_observation(id)`
   - `mem_search(query: "sdd/{args}/design")` → optional, if exists

2. **Load Skills**: 
   - `mem_search(query: "skill-registry")` → get registry
   - Load relevant coding skills for the stack

3. **Implement Tasks**:
   - Follow spec scenarios as acceptance criteria
   - Match existing code patterns
   - Detect TDD mode from project config or skills
   - Mark completed tasks with `[x]`

4. **Persist Progress**:
   ```
   mem_update(id: {tasks-id}, content: "{updated tasks with [x]}")
   mem_save(
     title: "sdd/{args}/apply-progress",
     topic_key: "sdd/{args}/apply-progress",
     type: "architecture",
     project: "{project}",
     content: "{progress report}"
   )
   ```

## Instructions

- Load skill: `~/.hefesto/skills/sdd-apply/SKILL.md` for implementation workflow
- Load: `~/.hefesto/skills/_shared/phase-common.md` for return envelope
- Return: completed tasks, files changed, deviations, remaining tasks
- `next_recommended`: `/sdd-verify {args}` when all tasks complete
