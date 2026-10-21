# Runner

The router decides **which methodology fits a request**. The runner turns that
decision into a **sequence**: what to produce next, what to read first, and
what "done" means for the artifact in hand.

```go
step, _ := runner.Next(plan)
```

## What it does, and what it cannot do

A runner does not write the artifacts. A model or a person does. What the
runner guarantees is the **order** and the **completion criteria**, so
"the work is done" stops being a claim and becomes a check.

That distinction matters. The runner cannot tell whether a specification is
*good* — only whether it *exists and carries what its contract requires*. The
quality of the content still needs a reader, a reviewer or a native review.

## The three functions

| Function | Answers |
|---|---|
| `Next(plan)` | What is the next action, and what blocks it |
| `Check(plan, artifact)` | What is missing for one artifact to be complete |
| `Brief(plan, artifact)` | What to read, what to write, and what "done" means |

`Next` walks the artifacts the decision promised, in the order their
dependencies allow, and stops at the first one that is not complete. It
returns one of three phases:

- **`produce`** — the artifact does not exist yet.
- **`complete`** — the artifact exists but is missing content. The blockers say
  what.
- **`close`** — everything the contract requires is present. Run
  `hefesto verify-plan` to close the run.

## The loop

```
hefesto route --type known --size large --risk none --change add-widget --json
  → plan.json

loop:
  hefesto next --plan plan.json --json
    ├─ phase "complete" → fix the blockers it reports
    └─ phase "produce"  → follow the brief, then record the artifact in plan.json

hefesto verify-plan plan.json
  → exit 0 when the contract is satisfied
```

The loop never advances on a promise. It advances on a recorded artifact: the
plan says the artifact exists, it carries the fields its contract requires, and
its scenarios, tasks or verifications are there.

## Completion rules

The rules live in the artifact catalog, next to the artifact they belong to, so
the runner and the validator cannot disagree about what "complete" means:

| Field | Meaning |
|---|---|
| `RequiredFields` | The content keys the artifact must carry |
| `RequiresScenarios` | The artifact must state scenarios |
| `RequiresTasks` | The artifact must list tasks |
| `AllTasksDone` | Every task of the plan must be finished |
| `RequiresVerifications` | The artifact must record verifications |

## Tests

```bash
cd cmd/hefesto
go test ./runner/... -count=1 -race
```

The suite pins the phase the runner returns, the order it follows, the blockers
it reports, the brief it writes and the failures it refuses to accept: an
unknown profile, a contract that promises an artifact the catalog does not
know, and a dependency the decision does not promise.
