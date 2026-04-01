---
name: skill-creator
description: >
  Creates new AI agent skills following the Agent Skills spec.
  Trigger: When user asks to create a new skill, add agent instructions, or document patterns for AI.
version: 1.0.0
---

## When to Create a Skill

**Create when:**
- Pattern is used repeatedly
- Project-specific conventions differ from generic practices
- Complex workflows need step-by-step guidance

**Don't create when:**
- Documentation already exists (reference instead)
- Pattern is trivial or one-off

## Skill Structure

```
skills/{skill-name}/
├── SKILL.md              # Required
├── assets/               # Optional - templates, schemas
│   └── template.md
└── references/           # Optional - links to local docs
    └── docs.md
```

## Creation Process

### Step 1: Gather Requirements

Ask the user:
1. What should the skill accomplish?
2. When should it trigger? (keywords, file patterns, contexts)
3. Any existing code patterns to follow?

### Step 2: Choose Name

| Type | Pattern | Example |
|------|---------|---------|
| Generic | `{technology}` | `pytest`, `typescript` |
| Project-specific | `{project}-{component}` | `hefesto-api` |
| Workflow | `{action}-{target}` | `skill-creator` |

### Step 3: Create Directory

```bash
mkdir -p skills/{skill-name}/assets
mkdir -p skills/{skill-name}/references  # if needed
```

### Step 4: Generate SKILL.md

Use the template from `assets/SKILL-TEMPLATE.md`. Fill in:
- YAML frontmatter (name, description with trigger, version)
- When to Use section
- Critical Patterns (most important rules FIRST)
- Code Examples (minimal, focused)
- Commands (copy-paste ready)

### Step 5: Register

Add to `AGENTS.md` skill table:
```markdown
| `{skill-name}` | {Description} | Trigger: {when} |
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Lowercase, hyphens |
| `description` | Yes | What + Trigger in one block |
| `version` | Yes | Semantic version |

## Content Guidelines

### DO
- Start with critical patterns
- Use tables for decisions
- Keep examples minimal
- Include Commands section

### DON'T
- Add Keywords section (frontmatter is searched)
- Duplicate existing docs
- Include lengthy explanations
- Use web URLs in references (local paths only)

## Checklist

- [ ] Skill doesn't already exist
- [ ] Pattern is reusable
- [ ] Name follows conventions
- [ ] Frontmatter includes trigger
- [ ] Critical patterns are clear
- [ ] Examples are minimal
- [ ] Added to AGENTS.md

## Resources

- **Template**: See [assets/SKILL-TEMPLATE.md](assets/SKILL-TEMPLATE.md)
