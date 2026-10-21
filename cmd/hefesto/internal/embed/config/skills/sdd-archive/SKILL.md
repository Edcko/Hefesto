---
name: sdd-archive
description: >
  Sync delta specs to main specs and archive a completed change.
  Trigger: When the orchestrator launches you to archive a change after implementation and verification.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

> **Reference only.** The commands no longer launch this as a sub-agent. In the current flow, `hefesto route` decides the profile, `hefesto next` says when this artifact is due, and the plan (`.hefesto/<change>/plan.json`) records the result. This document is the craft guide for producing this artifact well.

## Purpose

ARCHIVING closes a completed change: apply the delta specs to the main specs (the source of truth), confirm the verification gate passed, and record the closure. This completes the SDD cycle.

## What You Receive

- Change name
- The change's artifacts (proposal, specs, design, tasks, verification report) as recorded in the plan

## How to Archive

### Check the Gate

NEVER archive a change whose verification report has CRITICAL issues. Re-read the verification report first.

### Apply Delta Specs to Main Specs

For each delta spec of the change:

#### If a Main Spec Exists for the Domain

Read the existing main spec and apply the delta:

```
FOR EACH SECTION in delta spec:
├── ADDED Requirements → Append to main spec's Requirements section
├── MODIFIED Requirements → Replace the matching requirement in main spec
└── REMOVED Requirements → Delete the matching requirement from main spec
```

**Merge carefully:**

- Match requirements by name (e.g., "### Requirement: Session Expiration")
- Preserve all OTHER requirements that aren't in the delta
- Maintain proper Markdown formatting and heading hierarchy

#### If No Main Spec Exists

The delta spec IS a full spec (not a delta). Use it directly as the domain's main spec.

### Record Closure in the Plan

Confirm and record:

- [ ] Main specs updated correctly
- [ ] Closure date stamped (ISO format, YYYY-MM-DD)
- [ ] All artifacts accounted for (proposal, specs, design, tasks, verification report)
- [ ] The change no longer counts as active

## Closure Summary Format

```markdown
## Change Archived

**Change**: {change-name}
**Closed on**: {YYYY-MM-DD}

### Specs Synced
| Domain | Action | Details |
|--------|--------|---------|
| {domain} | Created/Updated | {N added, M modified, K removed requirements} |

### Source of Truth Updated
The following specs now reflect the new behavior:
- {domain} specification

### SDD Cycle Complete
The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
```

## Rules

- NEVER archive a change that has CRITICAL issues in its verification report
- ALWAYS apply delta specs BEFORE recording closure
- When merging into existing specs, PRESERVE requirements not mentioned in the delta
- Use ISO date format (YYYY-MM-DD) to stamp the closure
- If the merge would be destructive (removing large sections), WARN and ask for confirmation
- The archive is an AUDIT TRAIL — never delete or modify archived changes
- Record the closure in the plan
