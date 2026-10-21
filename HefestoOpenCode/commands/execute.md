---
agent: sdd-orchestrator
description: Execute a routed change — follow the contract step by step until it closes
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Plan**: `.hefesto/{args}/plan.json`

## Task

Execute the change `{args}` by following the contract it was routed to.

### 1. Route, if the plan does not exist yet

```bash
hefesto route --type <type> --size <size> --risk <risk> --change {args} \
  --write-plan .hefesto/{args}/plan.json
```

Classify the request first: the dominant uncertainty (`known`,
`ambiguous_behavior`, `unknown_domain`, `unvalidated_idea`), the size (`small`
or `large`) and the high-risk item it touches (`none`, `data`, `security`,
`contract`, `concurrency`, `delivery`, `untestable`). The command decides the
profile; do not choose one yourself.

### 2. Loop

```bash
hefesto next --plan .hefesto/{args}/plan.json --json
```

Act on the phase it reports:

- **`produce`** — follow the brief, produce the artifact, record it in the plan.
- **`complete`** — the artifact exists but is missing content; fix the blockers.
- **`close`** — everything the contract requires is present; go to step 3.

Repeat until the phase is `close`.

### 3. Close

```bash
hefesto verify-plan .hefesto/{args}/plan.json
```

Do not report the change as done while this command reports errors.

## Instructions

- Load the skill: `~/.config/opencode/skills/profile-execute/SKILL.md`
- The plan is the audit trail: record every artifact before moving on
- Keep the change folder's `plan.json` current; an artifact that does not
  exist stays `present: false`
- If a brief contradicts the codebase, report it instead of working around it
- Return a structured envelope: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks`
