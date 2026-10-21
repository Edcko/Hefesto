---
agent: sdd-orchestrator
description: Bootstrap project context — detect stack, build the skill registry, save it as memory
---

## Context

- **Working Directory**: {cwd}
- **Project**: {project}

## Task

Bootstrap the project context. This is memory for the agent, not flow state:
the flow lives in the plan of each change.

1. **Detect Stack**: Scan for package.json, go.mod, pyproject.toml, etc.
   Identify the tech stack, test frameworks and linters.

2. **Build Skill Registry**: Scan the installed skills and write
   `.atl/skill-registry.md`; save a copy to Engram (`skill-registry`).

3. **Persist Context**: Save the detected context to Engram as project memory
   (`sdd-init/{project}`).

4. **Return Summary**: Stack detected, registry written, ready for `/sdd-new`.

## Instructions

- Load skill: `~/.config/opencode/skills/profile-route/SKILL.md` to see where
  this fits in the flow
- The plan of a change is `.hefesto/<change>/plan.json` — this command does
  not create one
- Return structured envelope: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks`
