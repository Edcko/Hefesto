---
name: sdd-init
description: Bootstrap SDD context in any project. Detects stack, conventions, and persistence backend.
trigger: When user wants to initialize SDD in a project, or says "sdd init", "iniciar sdd"
version: 1.0.0
---

## Purpose

Initialize Spec-Driven Development context in a project. Detect stack, conventions, and verify Engram availability.

## Protocol

Follow `skills/_shared/phase-common.md` for skill loading and return envelope format.

---

## What to Do

### Step 1: Detect Project Context

Read the project to understand:

- **Tech stack** — check package.json, go.mod, pyproject.toml, Cargo.toml, etc.
- **Framework** — Angular, React, Next.js, Django, FastAPI, etc.
- **Build tool** — npm/pnpm/yarn, go build, cargo, etc.
- **Test runner** — Jest, Vitest, pytest, go test, etc.
- **Linting/formatting** — ESLint, Prettier, Black, gofmt, etc.
- **Architecture patterns** — Clean, Hexagonal, Screaming, etc.

### Step 2: Check Engram Availability

Attempt a test call to `mem_search`. If successful, Engram is active. If it fails, announce graceful degradation.

### Step 3: Build Skill Registry

1. **Scan user skills**: glob `*/SKILL.md` in `~/.config/opencode/skills/`, `~/.claude/skills/`, `~/.gemini/skills/`, `~/.cursor/skills/`, project's `skills/` directory
2. **Scan project conventions**: check for `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `GEMINI.md` in project root
3. **Write `.atl/skill-registry.md`** (create `.atl/` if needed) with format:

```markdown
# Skill Registry — {project}

## Stack
- Language: {detected}
- Framework: {detected}
- Build: {detected}
- Test: {detected}

## Coding Skills
| Skill | Trigger | Path |
|-------|---------|------|
| {name} | {trigger} | {path} |

## Project Conventions
| File | Purpose |
|------|---------|
| {file} | {purpose} |
```

4. **If Engram available**, ALSO save: `mem_save(title: "skill-registry/{project}", topic_key: "skill-registry/{project}", type: "config", project: "{project}", content: "{registry}")`

### Step 4: Persist Project Context

Save to Engram with topic_key for upsert:

```
mem_save(
  title: "sdd-init/{project}",
  topic_key: "sdd-init/{project}",
  type: "architecture",
  project: "{project}",
  content: |
    # Project Context — {project}
    
    ## Stack
    - Language: {detected}
    - Framework: {detected}
    - Build: {detected}
    - Test: {detected}
    
    ## Architecture
    {detected patterns}
    
    ## Conventions
    {detected conventions}
    
    ## Skill Registry
    Path: .atl/skill-registry.md
    Engram: skill-registry/{project}
)
```

---

## Return Envelope

```markdown
## Status: {success|partial|failed}

## Executive Summary
SDD initialized for {project}. Stack: {detected}. Persistence: {engram available| degraded}.

## Artifacts
- project-context: sdd-init/{project}
- skill-registry: skill-registry/{project}

## Next Recommended
sdd-plan — Ready to plan a change.

## Risks
{None | "Engram unavailable — artifacts will not persist across sessions"}
```

---

## Rules

- NEVER create `openspec/` directories — Hefesto uses Engram only
- ALWAYS detect the real tech stack — don't guess
- If Engram is unavailable, announce it but continue (graceful degradation)
- Keep project context CONCISE — no more than 20 lines
