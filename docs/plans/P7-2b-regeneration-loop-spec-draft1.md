# P7 #2b — regeneration-as-repair loop (BR O-2) — design spec (DRAFT 1)

**Status:** **DRAFT 1 — FOR NON-AUTHOR REVIEW. Not accepted; authorizes no code.** This is the
follow-on design spec the detector-power spec (§7) and `DEF-P7-REGENLOOP-001` defer to. Law 6
(spec-before-code): no `kernel/*`, `scripts/*`, `gates/*` work is authorized by this document, and
no new provider, event kind, wire identity, or writable truth. Implementation requires this spec
accepted by a non-author, a NEW owner-approved change intent, and a non-author package verdict.
**Date:** 2026-11-09
**Authorities:**
- `vision-2028.md` §item **2b** (regeneration-as-repair: "discard the implementation and re-derive it
  against the lock plus the new counterexample, instead of patching") and its line ~156 floor ("the
  whole ladder can be climbed while detector power goes to zero").
- Requirement `BR-regeneration-trust-5f8a03` obligation **O-2** (declared `active`, unimplemented).
- The accepted **O-1** detector-power trust gate, `docs/plans/P7-2b-detector-power-spec-draft1.md`
  (ACCEPTED round-3; implementation non-author accepted
  `docs/verification/verdicts/p7-2b-detector-power-impl-review-round1-20261109.md`, ACCEPTABLE-WITH-NOTES).
- The accepted-for-planning behavior-lock contract `P7-behavior-locks-regeneration-plan-draft1.md`
  (subject `kernel/internal/twin.ReplayIsolated`; lock members, `UNVERIFIED`/`UNDECLARED` outcomes,
  exit criteria).

## 1. Purpose — what the loop adds over O-1

O-1 **measures** whether a behavior lock can reject a wrong implementation (class-U unproven-latitude
count ≤ per-tier budget). It authorizes nothing by itself (vision line ~156: detector power is
*necessary, not sufficient*). O-2 is the **act** O-1 gates: when an invariant fails, the loop
**discards the implementation and re-derives it** against the lock plus the counterexample, rather
than patching — the measured-expensive act (vision 2b: patching introduced 1–3 new defects per cycle
for four cycles). The loop's whole trust rests on one claim that must be made defensible, not
assumed: *a lock whose detector power clears the bound will reject a wrongly-regenerated candidate.*
This spec defines the loop and the safeguards that make that claim honest.

## 2. The loop (mechanism)

A single repair cycle, given a lock `L@r` (subject, revision `r`) and a failing-invariant
**counterexample** `X` (a reproducible input + the invariant it violates — e.g. a replay case that
yields `changed` where the lock requires `preserved`, or a property violation):

1. **Admit the counterexample.** `X` is content-addressed and folded into a NEW lock revision
   `L@r+1` (behavior-lock plan: a later revision is a new artifact; it cannot rewrite `L@r`). `X`
   becomes a pinned, tested line of the new lock.
2. **Re-measure detector power of `L@r+1`.** Run O-1 (`scripts/detector-power.sh`) against the new
   lock's trusted surface. Detector power does **not** inherit across revisions: `L@r`'s budget-clear
   says nothing about `L@r+1`. Regeneration is authorized **only if** `L@r+1` clears the per-tier
   `U_BUDGET` (top tier `U=0`). If it does not, the loop **halts** and the remedy is authored
   lock-strengthening (O-1 §3) — never regeneration against an under-powered lock.
3. **Re-derive, in isolation.** Discard the failing implementation and regenerate it from the
   authorized intent + `L@r+1`, under the **no-overfitting isolation** of §3. Output: a candidate
   implementation, declared as a candidate (behavior-lock plan: synthetic candidates never become
   observations).
4. **Verify against the lock.** Evaluate the candidate through the lock: the full invariant set, the
   `vera.replay.v1` proof bindings, and the behavior-lock exit criteria. Any `UNVERIFIED`/
   `UNDECLARED`/mismatch fails closed.
5. **Accept iff** (a) `L@r+1` cleared the budget in step 2, (b) the candidate passed step 4, and (c)
   the isolation obligation (§3) is attested. Acceptance is of the **candidate against `L@r+1`**, and
   is itself subject to Law 9 (a non-author verdict), never self-accepted by the loop.

The loop never "accepts a candidate that satisfies a lock whose detector power is below the
authorized bound" (O-2 statement) because step 2 gates step 5 on the **new** lock's re-measured
power.

## 3. The two circularity dependencies the loop MUST discharge (detector-power spec §7)

These are the spec's hard core. A loop that skips them launders O-1's measurement into a guarantee it
does not carry.

- **D-a — operator representativeness.** Detector power is the lock's kill rate against **mutation
  operators**. The loop uses it as a proxy for the lock's kill rate against **regenerator** failure
  modes. The proxy holds only if the operator set plausibly covers the regenerator's failure classes.
  The spec requires, per loop deployment, an explicit **representativeness argument** — either
  evidence that the operator basis spans the regenerator's observed failure classes, or a declared
  **bounded gap** registered as a Law-4 advisory (owner + ISO expiry) with a standing calibration
  obligation (vision's self-calibrating ladder: nightly promote/demote). An un-argued proxy fails
  closed: no representativeness record ⇒ no regeneration authority.
- **D-b — no overfitting against the measured tests.** The regenerator must not be trained or
  selected to satisfy the lock's specific test bodies (which would make a pass vacuous and the trust
  circular). The spec requires an **isolation obligation**: the regeneration provenance attests that
  the candidate was derived from the lock's *contract* (intent, invariants, counterexample) and not
  from the measured test implementations, and that the detector-power sweep of step 2 was run on the
  lock surface the candidate did **not** get to overfit. The residual risk after D-b — operator-
  unreachable blind spots that produce no survivor — is exactly O-1's "necessary not sufficient"
  caveat and is bounded, not eliminated; it is carried forward as a declared limitation, never
  silently closed.

## 4. Invariants (proposed contract — proving-test shapes TBD at implementation)

| # | Invariant | Proving test (shape) |
|---|---|---|
| RL-1 | A counterexample is admitted only as a new lock revision `L@r+1` (content-addressed, ancestry to `L@r`); it never rewrites `L@r` or its prior proofs/verdicts | `TestCounterexampleIsNewLockRevision` |
| RL-2 | Detector power is **re-measured on `L@r+1`** and regeneration is authorized iff `L@r+1` clears `U_BUDGET(tier)`; an under-powered new lock halts the loop (no inheritance from `L@r`) | `TestRegenerationGatedOnNewLockPower` |
| RL-3 | A candidate is accepted iff it passes the full lock evaluation (invariants + `vera.replay.v1` bindings + behavior-lock exit criteria) with no `UNVERIFIED`/`UNDECLARED`/mismatch | `TestCandidateAcceptedOnlyOnFullLockPass` |
| RL-4 | **D-a:** absent a representativeness record (argument or unexpired Law-4 advisory), the loop grants no regeneration authority (fails closed) | `TestNoRegenWithoutOperatorRepresentativeness` |
| RL-5 | **D-b:** acceptance requires an isolation attestation that the candidate was derived from the contract, not the measured test bodies; a missing/false attestation fails closed | `TestNoAcceptWithoutIsolationAttestation` |
| RL-6 | No satisfy-and-reopen: a candidate that passes the lock's tests while the lock is below budget is **never** accepted (RL-2 ∧ RL-3 enforced jointly, not severally) | `TestNoSatisfyAndReopen` |
| RL-7 | Every acceptance is a Law-9 artifact: a non-author verdict of the candidate@`L@r+1`; the loop never self-accepts | `TestAcceptanceRequiresNonAuthorVerdict` |
| RL-8 | The residual operator-blind-spot limitation is **declared** (carried as a standing advisory), never implied closed by a budget-clear | `TestResidualLimitationDeclared` |

## 5. Decisions required from the owner (§6-style, OPEN)
1. **Regenerator boundary.** What performs re-derivation (a model-driven generator, a bounded
   transform, a human-in-the-loop), and what its provenance/isolation attestation (D-b) must contain.
2. **Representativeness bar (D-a).** The initial standard: a required argument, a calibration
   obligation, or a bounded advisory with expiry — and who owns the nightly calibration.
3. **Scope of the first loop slice.** The narrowest honest subject (candidate: the `ReplayIsolated`
   behavior lock on the repo-native ledger feed, per the behavior-lock plan) and whether draft-1
   implements the full loop or only the **gate** (steps 2 + 5 — authorize/deny regeneration) with
   re-derivation left manual.
4. **Halting policy.** On an under-powered `L@r+1` (step 2 fail), confirm the remedy is authored
   lock-strengthening and the loop records the halt as evidence (no silent retry).

## 6. Non-goals (this piece)
- **No regenerator implementation and no new generation capability** — this spec defines the trust
  gate around re-derivation, not the generator.
- **No new event kind, provider, wire identity, or writable truth**; the lock, replay proof, and
  intent binding are the accepted behavior-lock artifacts.
- **No change to O-1** (detector-power gate) semantics, the zero-undeclared-survivors bar, or the
  mutation harness.
- **No automatic lock strengthening** — raising an under-powered lock remains authored work (O-1 §7).
- **No claim that a budget-clear is sufficient** — D-a/D-b and the declared residual limitation (RL-8)
  stand precisely because it is not.

## 7. Sequence & gates
1. **This draft → non-author review** (the deltas to scrutinize: the §3 D-a/D-b discharge, RL-2's
   no-inheritance re-measurement, and RL-6's joint enforcement — the satisfy-and-reopen guard).
2. Owner settles §5 (regenerator boundary, representativeness bar, first-slice scope, halting).
3. An owner-approved **new change intent** for the governed loop work (separate from
   `CI-p7-detector-power-1a2b3c`, whose scope explicitly excludes O-2).
4. Implement spec-first (the loop gate + its fail-closed self-test proving each RL control
   load-bearing), reusing O-1's `scripts/detector-power.sh` for step 2.
5. Package acceptance: calibrated `make mutants` of new code + a non-author verdict (Law 9).
