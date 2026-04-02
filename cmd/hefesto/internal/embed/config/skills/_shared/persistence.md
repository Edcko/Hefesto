# Persistence — Hefesto SDD

Engram is the ONLY persistence mode. No mode branching in skills.

---

## Topic Keys (Deterministic)

| Artifact | Topic Key |
|----------|-----------|
| Project context | `sdd-init/{project}` |
| Plan (explore+propose) | `sdd/{change-name}/plan` |
| Spec | `sdd/{change-name}/spec` |
| Design (optional) | `sdd/{change-name}/design` |
| Tasks | `sdd/{change-name}/tasks` |
| Apply progress | `sdd/{change-name}/apply-progress` |
| Verify report | `sdd/{change-name}/verify-report` |
| DAG state | `sdd/{change-name}/state` |
| Skill registry | `skill-registry/{project}` |

---

## Retrieval Protocol (Two-Step)

Search results are truncated (300 chars). Full content requires two calls:

```
1. mem_search(query: "{topic_key}", project: "{project}") → get ID
2. mem_get_observation(id: {id}) → full content (REQUIRED for SDD dependencies)
```

### Retrieving Multiple Artifacts

Group searches first, then retrievals:

```
# Step A: Get all IDs
proposal_id = mem_search(query: "sdd/{change}/plan", project: "{project}")
spec_id = mem_search(query: "sdd/{change}/spec", project: "{project}")

# Step B: Get full content
proposal = mem_get_observation(id: proposal_id)
spec = mem_get_observation(id: spec_id)
```

---

## Writing Artifacts

```
mem_save(
  title: "{topic_key}",
  topic_key: "{topic_key}",
  type: "architecture",
  project: "{project}",
  content: "{full markdown}"
)
```

`topic_key` enables upserts — saving again updates, not duplicates.

### Updating Existing

When you have the observation ID:

```
mem_update(id: {id}, content: "{updated content}")
```

---

## DAG State (Orchestrator Only)

Persist after each phase transition:

```
mem_save(
  title: "sdd/{change-name}/state",
  topic_key: "sdd/{change-name}/state",
  type: "architecture",
  project: "{project}",
  content: |
    change: {change-name}
    phase: {last-phase}
    artifacts:
      plan: true
      spec: true
      design: false
      tasks: false
    tasks_progress:
      completed: []
      pending: []
    last_updated: {ISO date}
)
```

Recovery: `mem_search("sdd/{change-name}/state")` → `mem_get_observation(id)` → parse.

---

## Auto-Degradation

If Engram MCP is unavailable:

1. **Orchestrator announces**: "Engram unavailable — artifacts will not persist across sessions"
2. **Mode switches to**: inline-only (artifacts returned in conversation)
3. **No persistence**: State cannot survive compaction
4. **Warn user**: Recommend fixing Engram MCP before continuing

This is NOT a user-facing mode choice — it's automatic graceful degradation.

---

## Export (Post-Hoc)

After verify PASS, orchestrator can offer to export all artifacts as files:

- **Purpose**: Git tracking, code review, human readability
- **Format**: Markdown files in `openspec/changes/{change-name}/`
- **Trigger**: User request or explicit orchestrator decision
- **Replaces**: Old "openspec mode" and "archive phase"

Export is optional — artifacts live in Engram regardless.

---

## Skill Registry

Registry lives at `.atl/skill-registry.md` (project root). If Engram available, also saved there as `skill-registry/{project}`.

**Read priority**: Engram first (fast), file fallback.

---

## Sub-Agent Persistence Rules

| Context | Who reads | Who writes |
|---------|-----------|------------|
| Non-SDD task | Orchestrator (passes summary in prompt) | Sub-agent via `mem_save` |
| SDD phase | Sub-agent (two-step retrieval) | Sub-agent via `mem_save` |

Sub-agents ALWAYS write — they have complete detail. Orchestrator NEVER inlines SDD artifacts.

### Orchestrator Prompt Template

```
PERSISTENCE (MANDATORY):
Read artifacts: mem_search("{topic_key}", project="{project}") → mem_get_observation(id)
Save your artifact: mem_save(title="{topic_key}", topic_key="{topic_key}", type="architecture", project="{project}", content="...")
```
