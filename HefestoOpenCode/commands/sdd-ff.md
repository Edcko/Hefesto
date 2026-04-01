---
agent: sdd-orchestrator
description: Fast-forward — runs plan → spec → tasks sequentially (skip to implementation)
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Persistence**: Engram only

## Task

Execute **plan → spec → tasks** in sequence for change `{args}`:

### Phase 1: Plan
1. Explore codebase and create change proposal
2. Persist: `mem_save(topic_key: "sdd/{args}/plan", ...)`

### Phase 2: Spec
1. Read plan: `mem_search(query: "sdd/{args}/plan")` → `mem_get_observation(id)`
2. Write detailed requirements and scenarios (Given/When/Then)
3. Persist: `mem_save(topic_key: "sdd/{args}/spec", ...)`

### Phase 3: Tasks
1. Read spec: `mem_search(query: "sdd/{args}/spec")` → `mem_get_observation(id)`
2. Break into implementation checklist (hierarchical: 1.1, 1.2)
3. Persist: `mem_save(topic_key: "sdd/{args}/tasks", ...)`

## Instructions

- Load skills in order:
  1. `~/.hefesto/skills/sdd-plan/SKILL.md`
  2. `~/.hefesto/skills/sdd-spec/SKILL.md`
  3. `~/.hefesto/skills/sdd-tasks/SKILL.md`
- Load: `~/.hefesto/skills/_shared/phase-common.md` for return envelope
- If any phase fails, STOP and report the issue
- Return structured envelope with final status after all phases complete
- `next_recommended`: `/sdd-apply {args}` to start implementation
