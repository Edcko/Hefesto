---
agent: sdd-orchestrator
description: Validate a change against its contract and report the result
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Plan**: `.hefesto/{args}/plan.json`

## Task

Validate `{args}` against the contract of its profile:

```bash
hefesto verify-plan .hefesto/{args}/plan.json
```

The command reports every rule the plan breaks and exits non-zero while work
is outstanding. Turn its report into the verdict:

- **exit 0** — PASS: every promised artifact is present, complete, ordered,
  and every scenario has a verification.
- **exit non-zero** — FAIL: report the broken rules and hand the fixes back to
  `/sdd-apply {args}`.

The contract checks completeness, not quality: a passing plan can still hold
a bad specification. Say so in the report when the content warrants it.

## Instructions

- Load skill: `~/.config/opencode/skills/profile-execute/SKILL.md`
- Save lessons and decisions to Engram as memory; the verdict belongs to the
  report, not to a topic key
- Return structured envelope: `status` (PASS/FAIL), `executive_summary`, the
  broken rules (if any), `next_recommended`, `risks`
