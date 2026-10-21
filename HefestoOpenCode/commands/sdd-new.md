---
agent: sdd-orchestrator
description: Start a change — classify the request, route it, write its plan
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Plan**: `.hefesto/{args}/plan.json`

## Task

Start the change `{args}` under the methodology it is routed to.

1. **Classify** — report the three facts, one line each:
   - **Type** — `known`, `ambiguous_behavior`, `unknown_domain`, `unvalidated_idea`
   - **Size** — `small` or `large`
   - **Risk** — `none`, `data`, `security`, `contract`, `concurrency`, `delivery`, `untestable`

2. **Route and write the plan** — the command decides the profile, never you:

   ```bash
   hefesto route --type <type> --size <size> --risk <risk> --change {args} \
     --write-plan .hefesto/{args}/plan.json
   ```

3. **Report** — the profile, the verification tier, the artifacts the profile
   owes and the rationale.

## Instructions

- Load skill: `~/.config/opencode/skills/profile-route/SKILL.md`
- Never pick the profile yourself; classify honestly and let the command decide
- `next_recommended`: `/sdd-ff {args}` (planning artifacts) or `/sdd-apply {args}` (straight to work)
