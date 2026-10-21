# Router

The router package decides **which development methodology profile fits a request**, so an
orchestrator can select a discipline on its own instead of asking the human to
name a method every time.

It is a dependency-free package, so it can be consumed by the Hefesto CLI, by
OpenCode, by Pi, or by any other orchestrator without pulling an installer or
an agent runtime along.

```go
import "github.com/Edcko/Hefesto/cmd/hefesto/router"
```

## Why this exists

Two development practices are usually presented as alternatives: a *spec-driven*
flow, where requirements and scenarios are written down before code, and an
*organic* flow, where work stays small and proportional until the request earns
tracking. Choosing between them by hand is friction, and choosing between them
by vibes is worse.

The router makes the choice explicit, small and testable. It does not replace
either practice: the profiles describe **how the work is carried out**, and the
organizer's own rules (authorization, risk classification, evidence, delivery)
stay in force on top of whichever profile is selected.

## The decision table

The router is a pure function over three observed facts.

| Input | Values | Meaning |
|---|---|---|
| `Type` | `known`, `ambiguous_behavior`, `unknown_domain`, `unvalidated_idea` | The dominant uncertainty of the work |
| `Size` | `small`, `large` | The size classification of the work |
| `Risk` | `none`, `data`, `security`, `contract`, `concurrency`, `delivery`, `untestable` | The high-risk item the change touches |

| Type | Size | Risk | Profile | Verification | Follow-up |
|---|---|---|---|---|---|
| `unvalidated_idea` | any | any | `prototype-first` | writer-self, or independent when a risk item is named | `spec-first` |
| `unknown_domain` | any | any | `explore-first` | writer-self, or independent when a risk item is named | `spec-first` |
| `ambiguous_behavior` | `large` | any | `spec-first` | writer-self, or independent when a risk item is named | — |
| `ambiguous_behavior` | `small` | any | `inline-light` | writer-self, or independent when a risk item is named | — |
| `known` | `large` | any | `spec-first` | writer-self, or independent when a risk item is named | — |
| `known` | `small` | any | `inline-light` | writer-self, or independent when a risk item is named | — |

Naming a risk item always moves verification to `independent-verifier`, and adds
the probe that produces evidence about that risk. The `data`, `contract`,
`delivery` and `untestable` items also require a `rollback-plan` artifact.

### Profiles

- **`inline-light`** — the work stays in the parent session: one focused change,
  one test, a short report.
- **`explore-first`** — map the affected area, compare approaches, and list the
  open decisions before choosing one.
- **`prototype-first`** — build a throwaway artifact to test the hypothesis,
  record the feedback, then decide.
- **`spec-first`** — write the requirements and the testable scenarios down
  before implementation starts.

A design document is not part of the table: it is on demand, when a
specification contains an architectural decision worth separating.

## Usage

```go
decision, err := router.Decide(router.Request{
	Type: router.WorkAmbiguous,
	Size: router.SizeLarge,
	Risk: router.RiskContract,
})
if err != nil {
	return err
}

fmt.Println(decision.Profile)      // spec-first
fmt.Println(decision.FollowUp)     // (empty)
fmt.Println(decision.Verification) // independent-verifier
fmt.Println(decision.Artifacts)    // [decision-record proposal spec tasks verify-report compat-matrix rollback-plan]
fmt.Println(decision.Rationale)    // why this profile was chosen
```

`Route` assembles everything a run needs in one step: the decision, the
contract of the chosen profile, the markdown document the run records itself in
and the plan the run starts from.

```go
payload, err := router.Route(request, "add-widget")
if err != nil {
	return err
}

report := router.Validate(payload.Plan) // what is still outstanding
```

`Decide` fails closed: an unknown work type, size or risk item returns an
error instead of a guess, because a silent wrong guess routes work into the
wrong discipline.

## The contract layer

A decision is only half of the contract. The other half is what a run owes and
what a run must show.

**Artifacts.** Every artifact a profile can owe is cataloged: its name, its kind
(`decision`, `document`, `test`, `probe`, `report`), its purpose and the
content its contract requires. A profile promises artifacts by name, and a name
that is not in the catalog does not exist. That is what stops a profile from
promising an artifact no phase produces.

Each artifact also declares how it is completed: the fields it must carry,
whether it must state scenarios or tasks or verifications, and whether every
task of the plan must be finished. The runner and the validator read the same
declaration, so they cannot disagree about what "complete" means.

**Profile contracts.** Each profile states its purpose, the entry criteria that
justify it and the exit criteria that close it.

**Plans.** A plan is what a run produces: the decision that opened it, the
artifacts it created, the content each artifact carries, the scenarios the
specification states and the verifications that prove them.

**Validation.** `Validate` checks a plan against the contract its decision
promised. Every rule is evaluated, and the report says which ones failed:

| Rule | What it catches |
|---|---|
| `profile-known` | A plan that names a profile with no contract |
| `decision-recorded` | A run that never wrote down its decision |
| `artifact-known` | An artifact name that does not exist in the catalog |
| `required-artifacts-present` | A promised artifact that was never produced |
| `artifact-complete` | An artifact that exists but carries none of the content its contract requires |
| `dependency-order` | An artifact that exists before what it depends on |
| `scenario-coverage` | A scenario with no verification, or a verification of a scenario that does not exist |
| `verification-gate` | A run that did not reach the verification tier the decision required |
| `no-orphan-artifacts` | An artifact the decision did not ask for |

The validator fails closed: an unknown input is a finding, never a guess.

**Ordering.** `OrderArtifacts` returns the artifacts of a decision in the order
they must be produced: an artifact never comes before the artifacts it depends
on. It fails closed as well — an unknown artifact, a dependency the decision
does not promise and a dependency cycle are all errors, because each of them
means the contract and the plan disagree.

## Using it from a shell

The Hefesto CLI exposes the router and the validator:

```bash
hefesto route --type known --size large --risk contract --change add-widget --json
hefesto verify-plan path/to/plan.json
```

`route` prints the decision, the contract, the scaffold and the plan template.
`verify-plan` validates a plan the run has been updating, reports every broken
rule and exits non-zero while work is outstanding.

## Tests

```bash
cd cmd/hefesto
go test ./router/... -count=1 -race
```

The test suite pins the decision table case by case, checks that every profile
is reachable, that every decision carries artifacts, that every cataloged
artifact is promised by some decision, that discovery profiles carry a follow-up
and terminal ones do not, that the decision is deterministic, that invalid input
is rejected instead of guessed, and that each validation rule catches the
problem it exists for.

## Status

The decision layer, the contract layer and the plan validator are implemented
and tested. The profiles themselves are contracts, not yet orchestration: no
runtime drives them yet, and the artifacts remain free-form documents rather
than machine-produced ones.
