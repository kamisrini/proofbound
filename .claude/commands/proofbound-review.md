---
description: Adversarial pass against the Build Laws and the package SPEC, producing a durable non-author verdict
---

Run the adversarial review that CLAUDE.md requires before merging a meaningful chunk. The output is a
durable verdict committed to the repo — not a chat summary.

Tell me what is under review (package path, commit, or diff) if it is not already clear, and pin the
exact HEAD you are reviewing with `git rev-parse HEAD`.

## Law 9 is the gate on this command

**No actor grades its own work.** A verdict counts as *acceptance* only if the reviewer did not
author the change. So:

- If the change was authored by someone/something other than this session, you may adjudicate
  directly.
- If THIS session authored the change, you must get the independent lens from elsewhere: dispatch
  reviewer subagents (the Agent tool) that did not write the code, or hand off to a fresh session.
  A review you run over your own work is a **pre-check only** — write its verdict with
  `status: NON-ACCEPTING` and say plainly that a non-author verdict is still required. Do not dress a
  self-review up as acceptance.

Findings and remedies carry separate trust (`VD-verification-asymmetry-2dyjnd`): a verified finding
transfers ZERO trust to any fix it proposes. State findings as harms/broken invariants, not as
reproductions, and never as instructions to the author about what to do. The author fixes; a
non-author re-verifies.

## Pick the tier (VD-review-ceremony-tiers-v2b6kp)

- **Foundational** — anything owning identity, ordering, locking, crash semantics, or the ledger
  write path (`internal/core`, `internal/store`, and peers): one **instrumented parallel multi-lens**
  round — mutation-adequacy · spec-truth · concurrency · build-laws — dispatched as independent
  reviewers, run EARLY, with an adjudicator that discards any finding lacking a reproducible proof.
- **Connector / peripheral** — I/O-shaped, failures surface immediately (`internal/connector/*`,
  `internal/projections`, `internal/cli`, `cmd/*`): ONE reviewer, one adversarial pass, no mutation
  sweep. Escalate to the foundational tier ONLY when the change alters the ledger write path in
  `internal/store` or `internal/core`. Concurrency inside a connector is a question for the single
  reviewer, not a reason to add lenses.

## The budget (Law 11, `VD-review-ceremony-tiers-v2b6kp` amended 2026-11-19) — check it FIRST

- A spec gets at most **2** review rounds; code gets **1** review. Count the milestone's existing
  verdicts in `docs/verification/verdicts/` before dispatching. If the budget is spent, STOP and ask
  the owner; do not dispatch.
- **HIGH and MED findings** are fixed by the author. **LOW findings are recorded and left**: no fix,
  no re-check, unless the owner asks.
- A re-check, if any, covers only the MED/HIGH fixes, and its new LOWs are recorded too. It may run
  on a cheaper model.
- Tell the reviewer the budget in its prompt: findings must be HIGH/MED to change anything, so it
  should spend its effort finding those, not cataloguing corners.

## What the review covers

Adversarially, against BOTH:

- **The Build Laws** (CLAUDE.md) — walk each law that the change could violate: derived-state
  markers (Law 1), one-home/no-copy (Law 2), hooks self-test & fail-closed (Law 3), advisory
  expiry (Law 4), spec-before-code (Law 6), lessons-compile-to-enforcement (Law 7), and the rest.
- **The package SPEC** — every `INV-*` invariant: is it true, and does its named proving test
  actually discriminate the property, or does the fixture hold the key variable constant?

Mutation evidence is in scope at the FOUNDATIONAL tier only (or when a connector reviewer names one
specific test it doubts — then mutate that test's property, nothing more). Where it is in scope, run `scripts/mutants.sh <pkg>` and **state the calibration
before any count is believed** (gates.md row 41): a positive control that MUST die, a neutral control
that MUST survive, and a lethal injected-failing-test control that MUST be killed; `invalid`
(uncompilable) mutants are never counted as kills. Undeclared survivors fail; declare any survivor in
`docs/allowed-survivors.txt`.

## Commit the verdict (VD-verdicts-are-artifacts-rl0rab)

In the SAME session the verdict is received, commit it under
`docs/verification/verdicts/<round>.md`, verbatim except for mechanical cleanroom redaction of
forbidden strings (absolute paths, usernames) — each redaction marked in the provenance header with
its replacement token and count. The header names the run id, the driving script, and the HEAD under
review; if the tree moved between dispatch and receipt, say exactly what changed (a verdict is
evidence about one specific HEAD).

Match an existing schema-bearing verdict's frontmatter (see
`docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md`):

```
---
schema: vera.verdict.v1
verdict_id: <round>
status: ACCEPTABLE            # or the findings-bearing status; NON-ACCEPTING for a self pre-check
reviewed_commit: <full hex>
findings: []                  # or the list
artifact_path: docs/verification/verdicts/<round>.md
artifact_sha: <64 hex self-digest>
---
```

The self-digest is computed with its own line zeroed: write the file with
`artifact_sha: 0000…0000` (64 zeros), run `sha256sum` on it, and paste that digest back into the
line. Then validate every verdict artifact with `scripts/p6-artifact-integrity.sh --check` (it
re-zeroes the line, recomputes, and confirms the declared path and that `reviewed_commit` resolves).

End by printing: the tier, the reviewer identity (and whether it is a non-author, i.e. whether this
is acceptance or a pre-check), the verdict status, the committed verdict path, and — kept separate —
any findings for the author to act on.
