---
name: sdd-init
description: >
  Initialize the project context for the development flow. Detects stack, conventions, and architecture, and records them as project memory.
  Trigger: When user wants to initialize the project context, or says "sdd init", "iniciar sdd".
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

> **Reference only.** The commands no longer launch this as a sub-agent. In the current flow, `hefesto route` decides the profile, `hefesto next` says when this artifact is due, and the plan (`.hefesto/<change>/plan.json`) records the result. This document is the craft guide for producing this artifact well.

## Purpose

Initialize the Spec-Driven Development (SDD) context in a project: detect the project's stack and conventions, and produce a concise project-context summary that every later artifact builds on.

## What the Artifact Contains

### Detected Project Context

Read the project to understand:

- Tech stack (check `package.json`, `go.mod`, `pyproject.toml`, etc.)
- Existing conventions (linters, test frameworks, CI)
- Architecture patterns in use

Keep the recorded context CONCISE — no more than 10 lines covering stack, architecture, testing, and style.

### Summary Format

```markdown
## SDD Initialized

**Project**: {project name}
**Stack**: {detected stack}

### Context Detected
{summary of detected stack and conventions}

### Next Steps
Ready to explore or propose the first change.
```

Record the detected context in the plan so later artifacts build on it.

## Rules

- NEVER create placeholder spec files — specs are created during a change
- ALWAYS detect the real tech stack, don't guess
- Keep the recorded context CONCISE — no more than 10 lines
- Record the result in the plan of the change
