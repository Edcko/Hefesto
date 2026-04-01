---
agent: sdd-orchestrator
description: Start a NEW change — triggers plan phase (exploration + proposal)
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Persistence**: Engram only

## Task

Start the **plan phase** for change `{args}`. This phase MERGES exploration and proposal:

1. **Explore**: Investigate the codebase to understand current state, affected modules, and constraints.

2. **Propose**: Create a change proposal with:
   - Intent (what and why)
   - Scope (files/modules affected)
   - Approach (high-level implementation strategy)
   - Risks and rollback plan

3. **Persist**: Save plan to Engram:
   ```
   mem_save(
     title: "sdd/{args}/plan",
     topic_key: "sdd/{args}/plan",
     type: "architecture",
     project: "{project}",
     content: "{plan content}"
   )
   ```

4. **Return Summary**: Plan created, ready for `/sdd-ff` (skip to tasks) or `/sdd-spec` (detailed specs).

## Instructions

- First: `mem_search(query: "sdd-init/{project}", project: "{project}")` → get project context
- Load skill: `~/.hefesto/skills/sdd-plan/SKILL.md` for plan phase instructions
- Load skill: `~/.hefesto/skills/_shared/phase-common.md` for return envelope
- Return structured envelope: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks`
