---
agent: sdd-orchestrator
description: Fast-forward a routed change through its planning artifacts
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Plan**: `.hefesto/{args}/plan.json`

## Task

Produce the planning artifacts of `{args}` and stop before implementation.

Run the loop until every planning artifact of the profile is complete:

```bash
hefesto next --plan .hefesto/{args}/plan.json --json
```

- **`produce`** — follow the brief, produce the artifact, record it in the
  plan exactly as the brief lists it.
- **`complete`** — the artifact exists but is missing content; fix the blockers.
- **`close`** or an implementation artifact (a test, a task execution) — stop;
  hand off to `/sdd-apply {args}`.

The loop advances on a recorded artifact, never on a promise.

## Instructions

- Load skill: `~/.config/opencode/skills/profile-execute/SKILL.md`
- Save discoveries and decisions to Engram as memory; artifacts go in the plan
- `next_recommended`: `/sdd-apply {args}`
