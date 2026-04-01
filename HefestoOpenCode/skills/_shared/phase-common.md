# SDD Phase — Common Protocol

Boilerplate identical across all Hefesto SDD phases (init, plan, spec, tasks, apply, verify).

---

## Step 0: Skill Loading

Every phase MUST:

1. **Load own SKILL.md** — Your phase-specific instructions
2. **Check for passed standards** — Orchestrator may include `SKILL: Load "{path}"`
3. **If no standards passed** — Load relevant skills based on code context:
   - React code → `react-19`
   - TypeScript → `typescript`
   - Tailwind → `tailwind-4`
   - Tests → `pytest` or `playwright`

If skill registry unavailable, proceed without (not an error).

---

## Step 1: Load Dependencies

If your phase has dependencies (check SKILL.md), retrieve them via two-step:

```
mem_search(query: "sdd/{change-name}/{artifact}", project: "{project}") → ID
mem_get_observation(id: {ID}) → full content
```

---

## Step 2: Execute Phase

Follow your SKILL.md instructions. Save discoveries to Engram as you go:

```
mem_save(title: "{descriptive title}", type: "{decision|discovery|bugfix}",
         project: "{project}", content: "**What**: ...\n**Why**: ...\n**Where**: ...")
```

---

## Step 3: Persist Artifact

Before returning, save your phase artifact:

```
mem_save(
  title: "{topic_key}",
  topic_key: "{topic_key}",
  type: "architecture",
  project: "{project}",
  content: "{your full artifact}"
)
```

---

## Return Envelope (MANDATORY)

Every phase MUST return this exact structure:

```markdown
## Status: {success|partial|failed}

## Executive Summary
{2-3 sentences for the orchestrator}

## Artifacts
- {artifact-name}: {topic_key}

## Next Recommended
{next phase in DAG, or "none"}

## Risks
{concerns, blockers, decisions needed — or "None"}
```

---

## Sub-Agent Rules

1. **Fresh context** — No memory of previous phases
2. **Orchestrator passes references** — Topic keys, NOT content
3. **You retrieve via two-step** — Search then get_observation
4. **Save before returning** — Artifact AND any discoveries
5. **No further delegation** — Anti-recursion rule
6. **Report skill gaps** — If registry was incomplete, note in Risks

---

## Upsert Pattern

Using `topic_key` ensures updates don't create duplicates:

- First save: creates artifact
- Subsequent saves: updates existing (same topic_key)

Use `mem_suggest_topic_key` before first save if unsure about key format.
