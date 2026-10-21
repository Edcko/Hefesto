---
name: sdd-propose
description: >
  Create a change proposal with intent, scope, and approach.
  Trigger: When the orchestrator launches you to create or update a proposal for a change.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

> **Reference only.** The commands no longer launch this as a sub-agent. In the current flow, `hefesto route` decides the profile, `hefesto next` says when this artifact is due, and the plan (`.hefesto/<change>/plan.json`) records the result. This document is the craft guide for producing this artifact well.

## Purpose

Create PROPOSALS. Take the exploration analysis (or direct user input) and produce a structured proposal document: intent, scope, approach, affected areas, risks, rollback plan, and success criteria.

## What You Receive

- Change name (e.g., "add-dark-mode")
- Exploration analysis (from the exploration phase) OR direct user description

## How to Create the Proposal

### Read Existing Specs

If the plan already records specs for the affected domains, read them to understand current behavior this change might affect.

### Write the Proposal

```markdown
# Proposal: {Change Title}

## Intent

{What problem are we solving? Why does this change need to happen?
Be specific about the user need or technical debt being addressed.}

## Scope

### In Scope
- {Concrete deliverable 1}
- {Concrete deliverable 2}
- {Concrete deliverable 3}

### Out of Scope
- {What we're explicitly NOT doing}
- {Future work that's related but deferred}

## Approach

{High-level technical approach. How will we solve this?
Reference the recommended approach from exploration if available.}

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `path/to/area` | New/Modified/Removed | {What changes} |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| {Risk description} | Low/Med/High | {How we mitigate} |

## Rollback Plan

{How to revert if something goes wrong. Be specific.}

## Dependencies

- {External dependency or prerequisite, if any}

## Success Criteria

- [ ] {How do we know this change succeeded?}
- [ ] {Measurable outcome}
```

Record the finished proposal in the plan of the change.

## Summary Format

```markdown
## Proposal Created

**Change**: {change-name}

### Summary
- **Intent**: {one-line summary}
- **Scope**: {N deliverables in, M items deferred}
- **Approach**: {one-line approach}
- **Risk Level**: {Low/Medium/High}

### Next Step
Ready for specs or design.
```

## Rules

- If a proposal already exists for the change, READ it first and UPDATE it
- Keep the proposal CONCISE — it's a thinking tool, not a novel
- Every proposal MUST have a rollback plan
- Every proposal MUST have success criteria
- Use concrete file paths in "Affected Areas" when possible
- Record the proposal in the plan — the spec phase builds on it
