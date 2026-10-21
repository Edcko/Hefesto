---
agent: sdd-orchestrator
description: Implement and verify a routed change by following its contract
---

## Context

- **Change Name**: {args}
- **Working Directory**: {cwd}
- **Project**: {project}
- **Plan**: `.hefesto/{args}/plan.json`

## Task

Drive `{args}` to its close.

Run the loop until the phase is `close`:

```bash
hefesto next --plan .hefesto/{args}/plan.json --json
```

- **`produce`** — follow the brief. For implementation artifacts, detect TDD
  mode and follow RED → GREEN → REFACTOR; record the test as the artifact
  demands. For the verification artifact, run the real commands and record
  the observed results.
- **`complete`** — fix the blockers it reports, then record.
- **`close`** — run `hefesto verify-plan .hefesto/{args}/plan.json`.

Do not report the change as done while `hefesto verify-plan` reports errors.

## Instructions

- Load skill: `~/.config/opencode/skills/profile-execute/SKILL.md`
- Read the phase skills as reference when producing an artifact well:
  `sdd-apply` for the TDD cycle, `sdd-verify` for the compliance matrix
- The plan is the audit trail; an artifact nobody recorded does not exist
- `next_recommended`: `/sdd-verify {args}`
