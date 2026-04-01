---
name: sdd-verify
description: Validate that implementation matches specs and tasks.
trigger: When the orchestrator launches you to verify a completed or partially completed change
version: 1.0.0
---

## Purpose

QUALITY GATE. Prove — with execution evidence — that implementation is complete, correct, and compliant.

Static analysis alone is NOT enough. Execute tests.

## Dependencies

- **Required**: `sdd/{change-name}/spec`, `sdd/{change-name}/tasks`
- **Optional**: `sdd/{change-name}/apply-progress`, `sdd/{change-name}/design`

## Protocol

### Step 1: Load Dependencies

Retrieve via two-step (see `_shared/persistence.md`):

```
spec_id = mem_search(query: "sdd/{change-name}/spec", project: "{project}")
spec = mem_get_observation(id: spec_id)

tasks_id = mem_search(query: "sdd/{change-name}/tasks", project: "{project}")
tasks = mem_get_observation(id: tasks_id)

progress_id = mem_search(query: "sdd/{change-name}/apply-progress", project: "{project}")
if progress_id: progress = mem_get_observation(id: progress_id)
```

### Step 2: Check Completeness

```
Read tasks
├── Count total tasks
├── Count completed [x]
├── List incomplete [ ]
└── Flag: CRITICAL if core tasks incomplete
```

### Step 3: Run Tests

**Use the project's test command**. If unsure, ask orchestrator.

```
Execute: {test_command}
Capture:
├── Total run
├── Passed
├── Failed (name + error)
├── Skipped
└── Exit code

Flag: CRITICAL if exit code != 0
```

### Step 4: Compliance Matrix (Core Value)

For EACH spec requirement/scenario, verify implementation exists:

```
FOR EACH REQUIREMENT in spec:
  FOR EACH SCENARIO:
  ├── Find covering test (by name/description/path)
  ├── Look up test result from Step 3
  └── Assign status:
      ├── ✅ COMPLIANT  → test exists AND passed
      ├── ❌ FAILING    → test exists BUT failed
      ├── ❌ UNTESTED   → no test found
      └── ⚠️ PARTIAL   → test passes but incomplete coverage
```

**A scenario is only COMPLIANT when a test PASSED. Code existing is NOT sufficient.**

### Step 5: Check for Issues

Search codebase for:
- `TODO`, `FIXME`, `HACK` comments in changed files
- Incomplete implementations (stubs, empty functions)
- Deviations from design decisions

### Step 6: Persist Report

```javascript
mem_save(
  title: "sdd/{change-name}/verify-report",
  topic_key: "sdd/{change-name}/verify-report",
  type: "architecture",
  project: "{project}",
  content: "
## Verification Report

**Change**: {change-name}

### Completeness
| Metric | Value |
|--------|-------|
| Tasks total | {N} |
| Complete | {N} |
| Incomplete | {N} |

### Tests
**Result**: ✅ {N} passed / ❌ {N} failed / ⚠️ {N} skipped

{Failed tests if any}

### Compliance Matrix

| Requirement | Scenario | Test | Status |
|-------------|----------|------|--------|
| {REQ-01} | {scenario} | test_file > test_name | ✅ COMPLIANT |
| {REQ-01} | {scenario} | (none) | ❌ UNTESTED |

**Summary**: {N}/{total} scenarios compliant

### Issues

**CRITICAL** (must fix):
{List or None}

**WARNING** (should fix):
{List or None}

### Verdict
{PASS | PARTIAL | FAIL}

{One-line summary}
"
)
```

### Step 7: Return

Follow `_shared/phase-common.md` return envelope:

```
## Status: success | partial | failed

## Executive Summary
Verdict: {PASS/PARTIAL/FAIL}. {N}/{total} scenarios compliant. {Summary}

## Artifacts
- verify-report: sdd/{change-name}/verify-report

## Next Recommended
{none (if PASS) | sdd-apply (if incomplete) | fix issues}

## Risks
{None | critical issues to address}
```

## Rules

- ALWAYS execute tests — static analysis is not verification
- A scenario is COMPLIANT only when a test PASSED
- Compare against SPEC first (behavioral), DESIGN second (structural)
- Be objective — report what IS, not what should be
- CRITICAL = must fix, WARNING = should fix
- DO NOT fix issues — only report them
- On PASS: orchestrator may offer to export artifacts as files
