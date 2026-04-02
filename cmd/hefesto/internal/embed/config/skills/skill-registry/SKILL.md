---
name: skill-registry
description: >
  Scan available skills and generate a compact registry for the orchestrator.
  Trigger: When user says "update skills", "skill registry", "update registry", or after installing/removing skills.
version: 1.0.0
---

## Purpose

Generate a **compact skill registry** that the orchestrator reads once per session. The registry contains trigger-based rules (5-15 lines per skill), NOT full skill content.

## When to Run

- After installing/removing skills
- When user explicitly requests it
- As part of `sdd-init`

## Process

### Step 1: Scan Skills

Scan these directories for `*/SKILL.md` files:

**User-level (global):**
- `~/.config/opencode/skills/`
- `~/.gemini/skills/`
- `~/.claude/skills/`

**Project-level (workspace):**
- `{project-root}/skills/`
- `{project-root}/.agent/skills/`

**Skip:**
- `sdd-*` directories (SDD workflow skills)
- `_shared/` directories (shared conventions)
- `skill-registry` (this skill)

### Step 2: Extract Compact Rules

For each skill, read frontmatter only (first 15 lines). Extract:
- `name` - skill identifier
- `description` - find trigger text after "Trigger:"
- Key rules from first section (if visible in frontmatter scan)

### Step 3: Generate Registry

Write to `.atl/skill-registry.md`:

```markdown
# Skill Registry

**Orchestrator use only.** Pre-resolve paths and pass to sub-agents.

## Skills

### {skill-name}
- **Trigger**: {when to load}
- **Path**: `{full-path}/SKILL.md`
- **Rules**:
  - {rule 1}
  - {rule 2}
  - {rule 3}

{repeat for each skill}
```

### Step 4: Persist

**Always:**
1. Create `.atl/` directory if needed
2. Write `.atl/skill-registry.md`

**If engram available:**
```
mem_save(
  title: "skill-registry",
  topic_key: "skill-registry/{project}",
  type: "config",
  project: "{project}",
  content: "{registry content}"
)
```

### Step 5: Summary

Return:
```markdown
## Skill Registry Updated

**Project**: {project}
**Skills found**: {count}
**Location**: .atl/skill-registry.md

| Skill | Trigger |
|-------|---------|
| ... | ... |
```

## Rules

- ALWAYS write `.atl/skill-registry.md`
- ALWAYS save to engram if available
- Keep rules compact (5-15 lines per skill)
- SKIP `sdd-*`, `_shared`, `skill-registry`
- Add `.atl/` to `.gitignore` if not present
