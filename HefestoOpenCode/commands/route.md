---
agent: sdd-orchestrator
description: Route a change to a methodology profile — the profile is computed, not chosen
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}

## Task

Route `{args}` to the methodology profile that fits it, then follow that profile.

### 1. Classify

Report the three facts explicitly, in one line each:

- **Type** — the dominant uncertainty: `known`, `ambiguous_behavior`, `unknown_domain`, `unvalidated_idea`.
- **Size** — `small` or `large`.
- **Risk** — the high-risk item the change touches: `none`, `data`, `security`, `contract`, `concurrency`, `delivery`, `untestable`.

Do not guess the profile. The command decides it.

### 2. Route

```bash
hefesto route --type <type> --size <size> --risk <risk> --change {args} --json
```

### 3. Record

Write the payload's `plan` to `plan.json` and its `scaffold` to `plan.md`, both
inside the change folder for `{args}`.

### 4. Follow the profile

Work through the artifacts the payload lists, in order. Load the `profile-route`
skill for the full rules of each profile.

### 5. Close

Before reporting the work as done:

```bash
hefesto verify-plan <change folder>/plan.json
```

If the command reports errors, fix the plan or the work and run it again. Do not
report success while it still fails.

## Instructions

- Load the skill: `~/.config/opencode/skills/profile-route/SKILL.md`
- The decision and its rationale belong in `plan.md`; the plan is the audit trail
- Keep `plan.json` current: an artifact that does not exist stays `present: false`
- Return a structured envelope: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks`
