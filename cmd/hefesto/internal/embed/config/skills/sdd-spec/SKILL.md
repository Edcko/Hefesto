---
name: sdd-spec
description: Write specifications with requirements and scenarios for a change.
trigger: When the orchestrator launches you to write specs for a change
version: 1.0.0
---

## Purpose

You are a sub-agent responsible for SPECIFICATIONS. Take the plan and produce delta specs — structured requirements and scenarios describing what's being ADDED, MODIFIED, or REMOVED.

## Protocol

Follow `skills/_shared/phase-common.md` for skill loading and return envelope.
Follow `skills/_shared/persistence.md` for retrieval and persistence.

---

## What You Receive

From the orchestrator:
- Change name
- Project name

---

## What to Do

### Step 1: Load Dependencies (REQUIRED)

**CRITICAL: `mem_search` returns truncated previews. You MUST call `mem_get_observation` for full content.**

```
# Get IDs
plan_id = mem_search(query: "sdd/{change-name}/plan", project: "{project}")

# Get full content
plan = mem_get_observation(id: {plan_id})
```

If no plan found, FAIL — specs require a plan.

### Step 2: Identify Affected Domains

From the plan's "Affected Areas", determine which domains are touched. Group changes by domain (e.g., `auth/`, `payments/`, `ui/`).

### Step 3: Write Delta Specs

Use this format for each domain:

```markdown
# Spec: {Change Name} — {Domain}

## ADDED Requirements

### Requirement: {Name}

The system {MUST|SHALL|SHOULD|MAY} {behavior}.

#### Scenario: {Happy path}
- GIVEN {precondition}
- WHEN {action}
- THEN {expected outcome}

#### Scenario: {Edge case}
- GIVEN {precondition}
- WHEN {action}
- THEN {expected outcome}

## MODIFIED Requirements

### Requirement: {Existing Name}

{New description}

Previously: {what it was before}

#### Scenario: {Updated scenario}
- GIVEN {precondition}
- WHEN {action}
- THEN {outcome}

## REMOVED Requirements

### Requirement: {Name}

Reason: {why removed}
```

#### For NEW Domains (No Existing Spec)

Write a FULL spec instead of delta:

```markdown
# {Domain} Specification

## Purpose

{High-level description}

## Requirements

### Requirement: {Name}

The system {MUST|SHALL|SHOULD|MAY} {behavior}.

#### Scenario: {Name}
- GIVEN {precondition}
- WHEN {action}
- THEN {outcome}
```

### Step 4: Persist Artifact

**MANDATORY — do NOT skip.**

Concatenate all domain specs into one artifact:

```
mem_save(
  title: "sdd/{change-name}/spec",
  topic_key: "sdd/{change-name}/spec",
  type: "architecture",
  project: "{project}",
  content: "{all domain specs concatenated}"
)
```

---

## Return Envelope

```markdown
## Status: {success|partial|failed}

## Executive Summary
Specs created for {change-name}. {N domains}, {M requirements}, {K scenarios}.

## Artifacts
- spec: sdd/{change-name}/spec

## Next Recommended
sdd-tasks — Break down into implementation tasks.

## Risks
{Coverage gaps, ambiguous requirements — or "None"}
```

---

## Rules

- ALWAYS use Given/When/Then format for scenarios
- Use RFC 2119 keywords (MUST, SHALL, SHOULD, MAY) — LLMs already know these
- Every requirement MUST have at least ONE scenario
- Include happy path AND edge case scenarios
- Keep scenarios TESTABLE — someone should write automated tests from them
- DO NOT include implementation details — specs describe WHAT, not HOW
- If the plan is missing or incomplete, FAIL with clear message
