# P7 #2b — regeneration-as-repair loop (BR O-2) — design spec (DRAFT 2)

**Status:** **DRAFT 2 — FOR NON-AUTHOR RE-REVIEW. Not accepted; authorizes no code.** This is the
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
  `docs/verification/verdicts/p7-2b-detector-power-impl-review-round1-20261109.md`, status ACCEPTABLE
  with finding F1 since fixed and re-confirmed).
- The accepted-for-planning behavior-lock contract `P7-behavior-locks-regeneration-plan-draft1.md`
  (subject `kernel/internal/twin.ReplayIsolated`; lock members, `UNVERIFIED`/`UNDECLARED` outcomes,
  exit criteria, lock identity = SHA-256 of canonical bytes).

## 0. Revision history

**Draft 2 — addresses the round-1 non-author review
(`verdicts/p7-2b-regeneration-loop-spec-review-round1-20261109.md`, NEEDS_WORK, 8 findings):**
- **F1 (HIGH) — RL-6 didn't block satisfy-and-reopen on a non-zero budget.** Resolved by **scoping the
  first loop slice to `U=0` tiers only** (§3 gate; new non-goal §6), where RL-2 ∧ RL-3 close
  satisfy-and-reopen airtight; RL-6 renamed and re-scoped accordingly. For any FUTURE non-zero-budget
  extension, new **RL-12** requires that a counterexample falsifying one class-U "safe-to-vary"
  justification **revokes every sibling budgeted-U justification** on that lock until each is re-argued.
- **F2 (MED) — step 2 misdescribed the mechanism and hid the per-cycle cost.** Step 2 reworded: the
  loop **produces fresh execution-bound detector-power evidence for `L@r+1`** (a new reproducible-or-
  attested O-1 record per the O-1 spec §4/DP-8) and then `scripts/detector-power.sh` *verifies* it
  (it verifies committed evidence; it does not run a sweep). The per-cycle cost is named (§2.1, §5):
  for slow embedded-postgres subjects (the named `internal/twin` slice) each cycle needs DP-8
  **attested** mode — an out-of-band sweep + a named non-author independent re-runner (O-1 CL-1) + an
  unexpired Law-4 advisory.
- **F3 (MED) — no counterexample-lethality invariant.** New **RL-9**: `X`'s admitted pinned line MUST
  be demonstrated to FAIL on the pre-repair implementation (reproduce the violation) before admission
  — the loop's analog of O-1's DP-7 lethal control; a non-lethal/strawman `X` fails closed.
- **F4 (MED) — D-b "false attestation fails closed" was undefined / collapsible to self-attestation.**
  D-b strengthened (RL-5): the isolation attestation MUST be made or independently checked by a
  **non-author of the regeneration** (analog of O-1 CL-1), not a self-sign-off. The "lock IS the
  tests" seam (O-1 N2) is acknowledged: in a single-context setting D-b cannot be fully mechanically
  enforced, so the residual is carried as a **declared bounded Law-4 limitation** (RL-8), not claimed
  airtight.
- **F5 (MED) — no lock-level Law-9 separation.** New **RL-10**: the actor that authors `L@r+1` (folds
  `X`, strengthens the lock) MUST be distinct from the regenerator, and both distinct from the final
  non-author verdict (RL-7). No grading a candidate against its own author's lock.
- **F6 (LOW/MED) — Law 2: no home for the D-a / D-b records.** §3.1 assigns a single durable home:
  `docs/verification/regeneration/<lock-revision-digest>/{representativeness.md,isolation-attestation.md}`,
  linked from the lock record; never a second copy.
- **F7 (LOW) — steps 1/2/4 not bound to one lock digest.** New **RL-11**: the `L@r+1` canonical
  SHA-256 folded (step 1), measured (step 2), and evaluated against (step 4) MUST be byte-identical; a
  revision swap fails closed.
- **F8 (LOW) — D-a evidence arm unreachable at genesis.** Stated openly (§3 D-a): at genesis no
  regenerator-failure corpus exists, so the first deployment **necessarily** uses the bounded-advisory
  path; the evidence arm becomes available only as the loop accumulates observed failures.

**Draft 1:** the repair cycle, RL-1..RL-8, and the two §7 circularity dependencies as first-class
fail-closed invariants. (Confirmed sound by round 1: RL-2 no-inheritance, both dependencies carried,
design-only, behavior-lock alignment, RL-7 Law 9 on the final verdict.)

## 1. Purpose — what the loop adds over O-1

O-1 **measures** whether a behavior lock can reject a wrong implementation (class-U unproven-latitude
count ≤ per-tier budget). It authorizes nothing by itself (vision line ~156: detector power is
*necessary, not sufficient*). O-2 is the **act** O-1 gates: when an invariant fails, the loop
**discards the implementation and re-derives it** against the lock plus the counterexample, rather
than patching — the measured-expensive act (vision 2b: patching introduced 1–3 new defects per cycle
for four cycles). The loop's whole trust rests on one claim that must be made defensible, not
assumed: *a lock whose detector power clears the bound will reject a wrongly-regenerated candidate.*
This spec defines the loop and the safeguards that make that claim honest, **for the `U=0` tier
only** in this slice (§3).

## 2. The loop (mechanism)

A single repair cycle, given a lock `L@r` (subject, revision `r`) and a failing-invariant
**counterexample** `X` (a reproducible input + the invariant it violates — e.g. a replay case that
yields `changed` where the lock requires `preserved`, or a property violation):

1. **Admit the counterexample — lethal, then folded.** First (RL-9) `X` MUST be demonstrated to FAIL
   on the pre-repair implementation: the reproduction actually reopens the violation, so the new
   pinned line is lethal, not a strawman. Only then is `X` content-addressed and folded into a NEW
   lock revision `L@r+1` (behavior-lock plan: a later revision is a new artifact; it cannot rewrite
   `L@r`). `X` becomes a pinned, tested line of the new lock.
2. **Establish and verify detector power of `L@r+1`.** `L@r+1` is a new revision, so its power is
   **not** inherited from `L@r`. The loop **produces fresh execution-bound O-1 evidence for `L@r+1`**
   — a reproducible or (for slow embedded-postgres subjects such as `internal/twin`) a DP-8
   **attested** record (out-of-band sweep + named non-author re-runner + unexpired Law-4 advisory,
   O-1 CL-1) — and then `scripts/detector-power.sh` **verifies** that committed record (it verifies
   evidence; it does not itself sweep). Regeneration is authorized **only if** `L@r+1` clears the tier
   budget, **which in this slice is `U = 0`** (§3). If not, the loop **halts**; the remedy is authored
   lock-strengthening (O-1 §3), never regeneration against an under-powered lock. *Per-cycle cost:*
   for slow subjects this fresh evidence is an attested out-of-band sweep every cycle — a named
   latency/independence dependency (§5), not a cheap in-gate check.
3. **Re-derive, in isolation, by a distinct actor.** Discard the failing implementation and regenerate
   it from the authorized intent + `L@r+1`, under the **no-overfitting isolation** of §3. The
   regenerator MUST be distinct from the `L@r+1` author (RL-10). Output: a candidate implementation,
   declared as a candidate (behavior-lock plan: synthetic candidates never become observations).
4. **Verify against the lock.** Evaluate the candidate through **the same `L@r+1` digest** (RL-11):
   the full invariant set, the `vera.replay.v1` proof bindings, and the behavior-lock exit criteria.
   Any `UNVERIFIED`/`UNDECLARED`/mismatch fails closed.
5. **Accept iff** (a) `L@r+1` cleared `U=0` in step 2, (b) the candidate passed step 4 against the
   identical `L@r+1` digest, and (c) the isolation obligation (§3, attested by a non-author of the
   regeneration) holds. Acceptance is of the **candidate against `L@r+1`**, and is itself subject to
   Law 9 (a non-author verdict by an actor distinct from both the lock author and the regenerator,
   RL-10), never self-accepted by the loop.

Because this slice gates on `U = 0`, there is **no budgeted un-pinned surface** for a regenerated
candidate to satisfy-and-reopen: every class-U degree of freedom must have been killed or proven E/R
before `L@r+1` is regeneration-trusted. That is what makes "no satisfy-and-reopen" (RL-6) airtight
here; non-zero budgets are out of scope (§6) and guarded for the future by RL-12.

## 3. The two circularity dependencies the loop MUST discharge (detector-power spec §7)

These are the spec's hard core. A loop that skips them launders O-1's measurement into a guarantee it
does not carry.

- **D-a — operator representativeness.** Detector power is the lock's kill rate against **mutation
  operators**. The loop uses it as a proxy for the lock's kill rate against **regenerator** failure
  modes. The proxy holds only if the operator set plausibly covers the regenerator's failure classes.
  The spec requires, per loop deployment, an explicit **representativeness record** — either evidence
  that the operator basis spans the regenerator's *observed* failure classes, or a declared **bounded
  gap** registered as a Law-4 advisory (owner + ISO expiry) with a standing calibration obligation
  (vision's self-calibrating ladder: nightly promote/demote). **At genesis the evidence arm is
  unreachable** — no regenerator-failure corpus exists yet — so the first deployment necessarily uses
  the bounded-advisory path, and the evidence arm opens only as the loop accumulates observed failures
  (F8). An un-argued proxy fails closed: no representativeness record ⇒ no regeneration authority.
- **D-b — no overfitting against the measured tests.** The regenerator must not be trained or selected
  to satisfy the lock's specific test bodies (a vacuous, circular pass). The spec requires an
  **isolation attestation**: that the candidate was derived from the lock's *contract* (intent,
  invariants, counterexample) and not from the measured test implementations. The attestation MUST be
  made or independently checked by a **non-author of the regeneration** (RL-5; O-1 CL-1 analog) — not
  a self-sign-off. **Honest limit (the "lock IS the tests" seam, O-1 N2):** when the invariants and the
  measured test bodies are two renderings of one behavior, "derived from the contract, not the tests"
  is only thinly separable in a single-repo/single-context setting, so D-b **cannot be fully
  mechanically enforced there**; that residual is carried as a declared bounded Law-4 limitation
  (RL-8), never claimed airtight. The operator-unreachable blind spot (no survivor at all) is the same
  O-1 "necessary not sufficient" caveat and is likewise carried, not eliminated.

### 3.1 One home per record (Law 2)
The D-a representativeness record and the D-b isolation attestation are load-bearing data with a
single durable home: `docs/verification/regeneration/<lock-revision-digest>/representativeness.md` and
`.../isolation-attestation.md`, linked from the lock record by digest — never copied into the lock
or stored twice.

### 3.2 The gate: `U = 0` in this slice
`L@r+1` is **regeneration-trusted** in this slice only when O-1 reports `U = 0` at the top tier for
`L@r+1`. A non-zero `U_BUDGET` authorizes regeneration against a known satisfy-and-reopen surface
(O-1 CL-4) whose per-survivor "safe to vary" arguments are exactly the kind a counterexample can
falsify — out of scope here (§6), and future-guarded by RL-12.

## 4. Invariants (proposed contract — proving-test shapes TBD at implementation)

| # | Invariant | Proving test (shape) |
|---|---|---|
| RL-1 | A counterexample is admitted only as a new lock revision `L@r+1` (content-addressed, ancestry to `L@r`); it never rewrites `L@r` or its prior proofs/verdicts | `TestCounterexampleIsNewLockRevision` |
| RL-2 | Fresh execution-bound detector power is established and **verified on `L@r+1`** (no inheritance from `L@r`); regeneration is authorized iff `L@r+1` clears the tier budget | `TestRegenerationGatedOnNewLockPower` |
| RL-3 | A candidate is accepted iff it passes the full lock evaluation (invariants + `vera.replay.v1` bindings + behavior-lock exit criteria) with no `UNVERIFIED`/`UNDECLARED`/mismatch | `TestCandidateAcceptedOnlyOnFullLockPass` |
| RL-4 | **D-a:** absent a representativeness record (argument or unexpired Law-4 advisory) in its §3.1 home, the loop grants no regeneration authority (fails closed) | `TestNoRegenWithoutOperatorRepresentativeness` |
| RL-5 | **D-b:** acceptance requires an isolation attestation, **made or independently checked by a non-author of the regeneration**, that the candidate was derived from the contract not the test bodies; missing/self-only/false fails closed | `TestIsolationAttestationIsNonAuthor` |
| RL-6 | **No satisfy-and-reopen (`U=0` slice):** the loop authorizes regeneration only when `L@r+1` clears `U=0`, so there is no budgeted un-pinned surface to satisfy-and-reopen; RL-2 ∧ RL-3 enforced jointly | `TestNoSatisfyAndReopenAtU0` |
| RL-7 | Every acceptance is a Law-9 artifact: a non-author verdict of the candidate@`L@r+1` | `TestAcceptanceRequiresNonAuthorVerdict` |
| RL-8 | The residual limitations (operator blind spots; the "lock IS the tests" D-b seam) are **declared** as standing Law-4 advisories, never implied closed by a budget-clear | `TestResidualLimitationsDeclared` |
| RL-9 | **Counterexample lethality:** `X`'s pinned line must be demonstrated to FAIL on the pre-repair implementation before admission; a non-lethal `X` fails closed | `TestCounterexampleIsLethalPreRepair` |
| RL-10 | **Role separation (Law 9 at the lock level):** the `L@r+1` author, the regenerator, and the final verdict reviewer are three distinct actors; any coincidence fails closed | `TestLockAuthorRegeneratorVerdictDistinct` |
| RL-11 | **Lock-digest identity:** the `L@r+1` canonical SHA-256 folded (step 1), measured (step 2), and evaluated against (step 4) are byte-identical; a revision swap between measure and evaluate fails closed | `TestSameLockDigestThroughCycle` |
| RL-12 | **Class-U falsification propagation (future non-zero budgets):** a counterexample that falsifies one class-U "safe-to-vary" justification revokes every sibling budgeted-U justification on that lock; each must be re-argued before `L@r+1` is regeneration-trusted | `TestClassUFalsificationRevokesSiblings` |

## 5. Decisions required from the owner (OPEN)
1. **Regenerator boundary.** What performs re-derivation (a model-driven generator, a bounded
   transform, a human-in-the-loop), and what its D-b isolation attestation must contain and who
   (non-author) checks it.
2. **Representativeness bar (D-a).** The initial standard and who owns the nightly calibration;
   confirm the genesis bounded-advisory default (F8).
3. **First-slice scope.** Confirm the `U=0`-only slice and the subject (the `ReplayIsolated` behavior
   lock on the repo-native ledger feed), and whether draft-2 implements the full loop or only the
   **gate** (steps 1–2 + 5 authorize/deny) with re-derivation left manual.
4. **Halting policy.** On an under-powered `L@r+1` (step 2 fail), confirm the remedy is authored
   lock-strengthening and the loop records the halt as evidence (no silent retry).
5. **Counterexample-lethality standard (RL-9).** What constitutes an acceptable demonstration that `X`
   fails on the pre-repair impl (a failing replay proof? a reproduced property violation?).
6. **Per-cycle cost acceptance.** Confirm that for slow subjects each repair cycle carries a DP-8
   attested out-of-band sweep with an independent re-runner (§2 step 2), and who provides it.

## 6. Non-goals (this piece)
- **No non-zero-`U_BUDGET` loop.** This slice gates on `U = 0`; regeneration against a budgeted
  class-U (satisfy-and-reopen) surface is out of scope and future-guarded by RL-12.
- **No regenerator implementation and no new generation capability** — this spec defines the trust
  gate around re-derivation, not the generator.
- **No new event kind, provider, wire identity, or writable truth**; the lock, replay proof, and
  intent binding are the accepted behavior-lock artifacts.
- **No change to O-1** (detector-power gate) semantics, the zero-undeclared-survivors bar, or the
  mutation harness.
- **No automatic lock strengthening** — raising an under-powered lock remains authored work (O-1 §7).
- **No claim that a budget-clear is sufficient** — D-a/D-b and the declared residual limitations
  (RL-8) stand precisely because it is not.

## 7. Sequence & gates
1. **This draft 2 → non-author re-review** (deltas: the `U=0` scoping of RL-6, the step-2 fresh-
   evidence/cost wording, RL-9 lethality, RL-5 non-author D-b, RL-10 role separation, RL-11 digest
   identity, RL-12 forward-guard, §3.1 homes).
2. Owner settles §5.
3. An owner-approved **new change intent** for the governed loop work (separate from
   `CI-p7-detector-power-1a2b3c`, whose scope excludes O-2).
4. Implement spec-first (the loop gate + its fail-closed self-test proving each RL control
   load-bearing), reusing O-1's `scripts/detector-power.sh` for step 2's verification.
5. Package acceptance: calibrated `make mutants` of new code + a non-author verdict (Law 9).
