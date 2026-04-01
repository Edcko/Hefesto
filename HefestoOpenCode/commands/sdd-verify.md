---
agent: sdd-orchestrator
description: Validate implementation against specs — auto-export on PASS
---

## Context

- **Change Name**: {args} (if empty, use active change from recent apply-progress)
- **Working Directory**: {cwd}
- **Project**: {project}
- **Persistence**: Engram only

## Task

Verify implementation for change `{args}`:

1. **Load Artifacts** (CRITICAL: use `mem_get_observation` for full content):
   - `mem_search(query: "sdd/{args}/spec")` → get ID → `mem_get_observation(id)`
   - `mem_search(query: "sdd/{args}/tasks")` → get ID → `mem_get_observation(id)`
   - `mem_search(query: "sdd/{args}/apply-progress")` → get ID → `mem_get_observation(id)`

2. **Validate Each Scenario**:
   - Compare implementation against every Given/When/Then scenario
   - Run tests if test infrastructure exists
   - Check that all tasks are marked `[x]`

3. **Generate Report**:
   - PASS: All scenarios satisfied, all tasks complete
   - PARTIAL: Some scenarios met, issues documented
   - FAIL: Critical gaps found

4. **Persist Report**:
   ```
   mem_save(
     title: "sdd/{args}/verify-report",
     topic_key: "sdd/{args}/verify-report",
     type: "architecture",
     project: "{project}",
     content: "{verification report}"
   )
   ```

5. **Auto-Export on PASS**: If verification passes, export summary to user and mark change as complete.

## Instructions

- Load skill: `~/.hefesto/skills/sdd-verify/SKILL.md` for verification workflow
- Load: `~/.hefesto/skills/_shared/phase-common.md` for return envelope
- Return: status (PASS/PARTIAL/FAIL), scenario results, issues found
- `next_recommended`: Fix issues (if PARTIAL/FAIL) or start new change (if PASS)
