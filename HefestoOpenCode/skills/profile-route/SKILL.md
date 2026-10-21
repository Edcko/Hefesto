---
name: profile-route
description: >
  Choose the development methodology profile for a request with `hefesto route`, and close the run with `hefesto verify-plan`.
  Trigger: When starting work that is more than a one-line change, or when a change needs a recorded plan.
license: MIT
metadata:
  author: Edcko
  version: "1.0"
---

## Purpose

The methodology is **computed, not chosen by taste**. `hefesto route` takes three
observed facts about the work and returns the profile that fits it, the artifacts
that profile owes, the verification tier it demands and the rationale behind the
decision. `hefesto verify-plan` checks at the end that the run actually produced
what the decision promised.

Use this skill to route a change before working on it, and to close it honestly.

## Classify the request

Three facts decide the profile. Report them explicitly before routing.

| Fact | Values | How to read it |
|---|---|---|
| `--type` | `known`, `ambiguous_behavior`, `unknown_domain`, `unvalidated_idea` | The dominant uncertainty. `known`: the expected behavior is understood. `ambiguous_behavior`: the behavior is not pinned down. `unknown_domain`: the affected area is not understood. `unvalidated_idea`: the outcome itself is untested. |
| `--size` | `small`, `large` | `small`: the outcome, the files and the risk are known and one diff finishes it. `large`: the work needs a tracked artifact to stay resumable. |
| `--risk` | `none`, `data`, `security`, `contract`, `concurrency`, `delivery`, `untestable` | The high-risk item the change touches. Naming one forces an independent verifier and the probe that evidences it. |

## Route

```bash
hefesto route --type <type> --size <size> --risk <risk> --change <name> --json
```

The payload carries the decision, the contract of the chosen profile, the
markdown scaffold and the plan the run starts from.

## Follow the profile

Each profile owes a fixed set of artifacts. The `Artifacts` list in the payload
is the complete list — there is nothing else to produce, and nothing you may
add without extending the contract.

| Profile | What it means | What it owes |
|---|---|---|
| `inline-light` | Small and understood: one focused change in the parent session. | A decision record and a focused test. |
| `explore-first` | The area is not understood: map it before deciding. | An exploration map, an options table and the open decisions. Then `spec-first`. |
| `prototype-first` | The outcome is a hypothesis: test it before writing production code. | A prototype and the feedback it produced. Then `spec-first`. |
| `spec-first` | The behavior is ambiguous or the work is large: write it down first. | A proposal, a specification with testable scenarios, the tasks and the verification report. |

## Record the run

1. Write the `plan` from the payload to `plan.json` in the change folder.
2. Write the `scaffold` to `plan.md` next to it.
3. As the work progresses, mark each artifact `present` once it exists, state the
   scenarios in the `spec` artifact, and record the verifications that prove
   them in the last artifact of the profile.
4. Keep the plan honest: an artifact that does not exist stays `present: false`.

## Close the run

```bash
hefesto verify-plan <change folder>/plan.json
```

The command reports every rule the plan breaks and exits non-zero while work
is outstanding. The rules are:

- the profile has a contract, and the decision was recorded;
- every artifact name resolves, and every promised artifact exists;
- the artifact order holds: nothing exists before what it depends on;
- every scenario in the specification has a verification;
- the verification tier the decision required is satisfied;
- the plan holds no artifact the decision did not ask for.

## Rules

- **Never pick the profile yourself.** The command decides; your job is to
  classify the request honestly and then follow what comes back.
- **Never declare the work finished while `hefesto verify-plan` reports
  errors.** Fix the plan or the work, then re-run it.
- **Never add artifacts the decision did not promise.** The validator rejects
  them on purpose: an extra artifact is a contract that drifted.
- **The rationale is part of the record.** If the profile looks wrong, the
  classification is wrong — reclassify and route again instead of working
  around the decision.
