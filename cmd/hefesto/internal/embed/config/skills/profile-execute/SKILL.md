---
name: profile-execute
description: >
  Drive a routed change through its contract: ask `hefesto next` what the run
  owes, produce that artifact, record it in the plan, and close with
  `hefesto verify-plan`.
  Trigger: When a change has been routed and the work must follow the profile it was given.
license: MIT
metadata:
  author: Edcko
  version: "1.0"
---

## Purpose

The router decided the methodology. The runner says what to do next. This skill
is the loop between them: **the run advances on a recorded artifact, never on
a promise.**

## The loop

1. **Route the change and write the plan.**

   ```bash
   hefesto route --type <type> --size <size> --risk <risk> --change <change> \
     --write-plan .hefesto/<change>/plan.json
   ```

2. **Ask what is next.**

   ```bash
   hefesto next --plan .hefesto/<change>/plan.json --json
   ```

3. **Act on the phase it reports.**

   | Phase | What it means | What to do |
   |---|---|---|
   | `produce` | The artifact does not exist yet | Follow the brief, produce the artifact, record it |
   | `complete` | The artifact exists but is missing content | Fix the blockers it lists, then record it |
   | `close` | Everything the contract requires is present | Run `hefesto verify-plan` to close the run |

4. **Record the artifact in the plan.** The brief lists exactly what to record:
   mark the artifact present, fill its fields, and add its scenarios, tasks or
   verifications. The plan is the audit trail — an artifact nobody recorded
   does not exist.

5. **Repeat from step 2** until the phase is `close`.

6. **Close the run.**

   ```bash
   hefesto verify-plan .hefesto/<change>/plan.json
   ```

   The command exits non-zero while work is outstanding.

## Rules

- **Never advance on a promise.** The plan must record the artifact before
  `hefesto next` moves past it.
- **Never mark an artifact present before it exists.** The validator reports
  an artifact that is present but carries none of its required content.
- **Never skip an artifact, and never produce one the decision did not
  promise.** The validator rejects both.
- **If the brief contradicts the code, report it.** Do not work around the
  contract: a brief that does not fit the codebase is a finding, not an
  obstacle.
- **The run is finished when `hefesto verify-plan` exits 0**, not when the
  last task looks done.
- **The contract checks completeness, not quality.** A specification that
  exists and carries its scenarios can still be wrong. That still needs a
  reader or a review.
