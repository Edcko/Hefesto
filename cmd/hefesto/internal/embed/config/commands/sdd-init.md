---
agent: sdd-orchestrator
description: Bootstrap SDD context in the project — detect stack, conventions, and persist to Engram
---

## Context

- **Working Directory**: {cwd}
- **Project**: {project}
- **Persistence**: Engram only

## Task

Initialize Spec-Driven Development for this project:

1. **Detect Stack**: Scan for package.json, go.mod, pyproject.toml, etc. Identify tech stack, test frameworks, linters.

2. **Build Skill Registry**: Scan `~/.hefesto/skills/` and `.agent/skills/` for relevant skills. Write `.atl/skill-registry.md` and save to Engram with topic_key `skill-registry`.

3. **Persist Context**: Save project context to Engram:
   ```
   mem_save(
     title: "sdd-init/{project}",
     topic_key: "sdd-init/{project}",
     type: "architecture",
     project: "{project}",
     content: "{detected context}"
   )
   ```

4. **Return Summary**: Stack detected, context saved, ready for `/sdd-new`.

## Instructions

- Load skill: `~/.hefesto/skills/_shared/persistence.md` for Engram protocol
- Load skill: `~/.hefesto/skills/_shared/phase-common.md` for return envelope format
- Return structured envelope: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks`
