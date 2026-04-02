---
name: sdd-plan
description: Investigate and create a change proposal. Merges exploration and proposal into one phase.
trigger: When the orchestrator launches you to plan a new change
version: 1.0.0
---

## Purpose

You are a sub-agent responsible for PLANNING. This phase merges exploration and proposal into one coherent step:

1. **Explore** — Investigate the codebase, understand constraints, compare approaches
2. **Propose** — Convert findings into a structured proposal with intent, scope, approach, risks

## Protocol

Follow `skills/_shared/phase-common.md` for skill loading and return envelope.
Follow `skills/_shared/persistence.md` for retrieval and persistence.

---

## What You Receive

From the orchestrator:
- Change name (e.g., "add-dark-mode")
- Topic or feature description
- Project name

---

## What to Do

### Step 1: Load Dependencies

Load project context (optional but recommended):

```
mem_search(query: "sdd-init/{project}", project: "{project}") → ID
mem_get_observation(id: {ID}) → full context
```

Load skill registry for relevant coding standards:

```
mem_search(query: "skill-registry/{project}", project: "{project}") → ID
mem_get_observation(id: {ID}) → full registry
```

### Step 2: Explore (READ-ONLY Investigation)

Understand the request and investigate:

- Is this a new feature? Bug fix? Refactor?
- What files/modules are affected?
- What patterns are already in use?
- What constraints exist?

```
INVESTIGATE:
├── Read entry points and key files
├── Search for related functionality
├── Check existing tests
├── Identify dependencies
└── Note risks and blockers
```

### Step 3: Analyze Approaches

If multiple solutions exist, compare:

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| Option A | ... | ... | Low/Med/High |
| Option B | ... | ... | Low/Med/High |

### Step 4: Write Proposal

```markdown
# Plan: {Change Title}

## Intent

{What problem are we solving? Why does this change need to happen?}

## Exploration Findings

### Current State
{How the system works today}

### Affected Areas
- `path/to/file.ext` — {why affected}
- `path/to/other.ext` — {why affected}

### Approaches Considered
{Summary of Step 3 analysis}

## Proposal

### Scope

**In Scope:**
- {Deliverable 1}
- {Deliverable 2}

**Out of Scope:**
- {Explicitly NOT doing}
- {Future work deferred}

### Approach
{Recommended approach and rationale}

### Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| {risk} | Low/Med/High | {how to mitigate} |

### Rollback Plan
{How to revert if something goes wrong}

### Success Criteria
- [ ] {Measurable outcome 1}
- [ ] {Measurable outcome 2}

## Recommendation

{GO / NO-GO with justification}
```

### Step 5: Persist Artifact

**MANDATORY — do NOT skip.**

```
mem_save(
  title: "sdd/{change-name}/plan",
  topic_key: "sdd/{change-name}/plan",
  type: "architecture",
  project: "{project}",
  content: "{your full plan markdown}"
)
```

---

## Return Envelope

```markdown
## Status: {success|partial|failed}

## Executive Summary
Plan created for {change-name}. {GO/NO-GO recommendation}. {N deliverables in scope}.

## Artifacts
- plan: sdd/{change-name}/plan

## Next Recommended
sdd-spec — Write specifications from this plan.

## Risks
{Concerns, blockers, decisions needed — or "None"}
```

---

## Rules

- DO NOT modify any code — this is READ-ONLY investigation
- ALWAYS read real code, never guess about the codebase
- Every plan MUST have a rollback plan and success criteria
- Keep CONCISE — the orchestrator needs a summary, not a novel
- If you can't find enough information, say so clearly
