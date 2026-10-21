---
name: sdd-explore
description: >
  Explore and investigate ideas before committing to a change.
  Trigger: When the orchestrator launches you to think through a feature, investigate the codebase, or clarify requirements.
license: MIT
metadata:
  author: gentleman-programming
  version: "2.0"
---

> **Reference only.** The commands no longer launch this as a sub-agent. In the current flow, `hefesto route` decides the profile, `hefesto next` says when this artifact is due, and the plan (`.hefesto/<change>/plan.json`) records the result. This document is the craft guide for producing this artifact well.

## Purpose

EXPLORATION. Investigate the codebase, think through problems, compare approaches, and return a structured analysis. Exploration only researches and reports back — it never modifies code.

## What You Receive

- A topic or feature to explore
- The change name, when the exploration is tied to a named change

## How to Explore

### Understand the Request

Parse what needs exploring:

- Is this a new feature? A bug fix? A refactor?
- What domain does it touch?

### Investigate the Codebase

Read relevant code to understand:

- Current architecture and patterns
- Files and modules that would be affected
- Existing behavior that relates to the request
- Potential constraints and risks

```
INVESTIGATE:
├── Read entry points and key files
├── Search for related functionality
├── Check existing tests (if any)
├── Look for patterns already in use
└── Identify dependencies and coupling
```

### Analyze Options

If there are multiple approaches, compare them:

| Approach | Pros | Cons | Complexity |
|----------|------|------|------------|
| Option A | ... | ... | Low/Med/High |
| Option B | ... | ... | Low/Med/High |

## Artifact Format: Exploration Analysis

```markdown
## Exploration: {topic}

### Current State
{How the system works today relevant to this topic}

### Affected Areas
- `path/to/file.ext` — {why it's affected}
- `path/to/other.ext` — {why it's affected}

### Approaches
1. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

2. **{Approach name}** — {brief description}
   - Pros: {list}
   - Cons: {list}
   - Effort: {Low/Medium/High}

### Recommendation
{Your recommended approach and why}

### Risks
- {Risk 1}
- {Risk 2}

### Ready for Proposal
{Yes/No — and what remains to be clarified}
```

When the exploration is tied to a named change, record the analysis in the plan so the proposal can build on it.

## Rules

- DO NOT modify any existing code or files — exploration is read-only
- ALWAYS read real code, never guess about the codebase
- Keep your analysis CONCISE — the reader needs a summary, not a novel
- If you can't find enough information, say so clearly
- If the request is too vague to explore, say what clarification is needed
- When tied to a change, record the analysis in the plan — the proposal depends on it
