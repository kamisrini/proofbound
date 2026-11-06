# P7 #2b — detector-power metric + regeneration-trust gate — design spec (DRAFT 3)

**Status:** DRAFT 3 — revises draft 2 after the round-2 non-author verdict
([`verdicts/p7-2b-detector-power-spec-review-round2-20261105.md`](../verification/verdicts/p7-2b-detector-power-spec-review-round2-20261105.md),
NEEDS_CHANGES "narrow, near-acceptance" at `986d6d7`; all eight round-1 findings confirmed resolved,
E/R/U core accepted). Design spec only (Law 6 spec-before-code): no `kernel/*`, `scripts/*`,
`gates/*` code, no change intent accepted. First piece of vision build-order **#2b** (witnessed
regeneration); the regeneration **loop** is deferred to a follow-on spec (§7). Implementation is
gated on this spec's non-author acceptance + an owner-approved change intent.
**Date:** 2026-11-05
**Authorities:** [vision-2028.md](../../vision-2028.md) §9 item 2b (+ line ~156: the ladder can be
climbed while detector power → 0); completion plan
[product-completion-plan-2026-11-05.md](product-completion-plan-2026-11-05.md) §A. Builds on
`tools/mutants` / `scripts/mutants.sh`, `docs/allowed-survivors.txt`, `docs/gates.md`.

## 0. Revision history (findings addressed)

**Draft 3, after round-2 N1–N4:**
- **N1 (MED) — gate on the ABSOLUTE count of unproven-latitude survivors, not the ratio.** The ratio
  `K/(K+U)` is size-dependent (same `U` scores differently by `K`), paddable (well-tested trivial
  code raises the tolerated `U`), and — the real hole — an operator-unreachable blind spot produces
  **no survivor**, so `U=0 → ratio 1.0` with no mis-declaration. So the **trust gate keys on `U`
  directly** (§3); the ratio is demoted to a reported **secondary diagnostic** (§2). §2's "a perfect
  lock can reach 1.0" is withdrawn: `U=0` means *no mutation-detected unproven latitude* — necessary,
  not sufficient.
- **N2 (MED)** — DP-6 freshness digest now covers the **lock inputs**: non-test source **and
  `_test.go`** (the lock *is* the tests) **and** the package's E/R/U register entries **and** its
  SPEC invariant/lock rows — not just "package source" (§4, DP-6).
- **N3 (MED)** — the committed raw sweep is now **bound to a real execution**: either reproducible
  (a refresh re-runs `scripts/mutants.sh` at the pinned lock-inputs digest and reproduces
  `raw_sweep_sha256`), or, where a live re-run is infeasible in-gate (slow embedded-postgres
  packages), explicitly **human-attested with a named non-author re-runner**, registered as a Law-4
  advisory limitation with an expiry (DP-8). Closes the fabricated-`sweep.txt` gap (the B1 unit-3
  hand-attestation lesson).
- **N4 (LOW)** — each **R** justification must name the **specific pinning test/line** whose mutant
  dies (and each **E** its equivalence argument); DP-2+DP-3 enforce an **exact bijection** between
  sweep survivors and record-classified survivors (§5).

**Draft 2, after round-1 F1–F8:** introduced the E/R/U survivor classification and
`detector_power = K/(K+U)` counting only class-U; DP-6 freshness; necessary-not-sufficient; raw-sweep
re-derivation; named loop dependencies; per-tier threshold. (All confirmed resolved in round 2.)

## 1. Purpose — why detector power must exist before regeneration

#2b discards a failing implementation and **re-derives it against the lock + the new counterexample**
instead of patching. The vision gate: *"regeneration is only as good as the lock … [our best lock]
still shipped five residual items, any of which a regenerated implementation could
satisfy-and-reopen. So 2b unlocks only where aggregate detector power is measured and high."* The
prerequisite is a **measurement of lock strength**: can the lock reject a *wrong* implementation?
This spec defines that measurement and the trust gate.

## 2. The measurement: classify survivors, then count unproven latitude

Naive `killed/killable` is trivially 1.0 (undeclared survivors are already banned). The signal is in
the survivors, which are **three kinds the raw sweep cannot tell apart** — so each carries a
**declared, justified class** in the register, itself non-author-reviewed:

1. **equivalent (E)** — semantically identical; **un-killable by any test**; zero exploitable latitude.
2. **redundantly-pinned (R)** — the behavior **is** pinned by a second tested line (defense-in-depth);
   the R justification MUST name that specific test/line whose mutant dies (N4). Not latitude, and
   penalizing it would reward deleting defense-in-depth.
3. **unproven latitude (U)** — a genuine **un-pinned degree of freedom** no test and no second line
   constrains. The only kind a regenerated impl can satisfy-and-reopen.

**The gated quantity is the absolute count `U` of class-U survivors** (N1). A **reported secondary
diagnostic** `detector_power = K/(K+U)` (K = killed; rejected when `K+U=0`, F7) is published for
trend/scale context but does **not** drive the gate — because the ratio is size-dependent, paddable,
and blind to operator-unreachable gaps.

**Honest limit:** `U=0` means *no mutation-detected unproven latitude*. A blind spot no mutation
operator reaches produces no survivor and cannot raise `U`; detector power is therefore **necessary,
not sufficient** (§3), and the loop spec must add the safeguards in §7.

## 3. The regeneration-trust gate (necessary, not sufficient)

`P` is **regeneration-trusted** only when `U(P) ≤ U_BUDGET(tier)` — the absolute count of class-U
(unproven-latitude) survivors is within a **per-tier budget**, with the highest tier requiring
`U = 0`, and **every** class-U survivor individually justified as a tolerated, bounded risk. This is
a **necessary precondition, not a sufficient guarantee**: finite mutation operators miss blind spots
(vision line ~156). Below budget-compliance, regeneration is **not** authorized: strengthen the lock
(add invariants / tighten tests) until class-U survivors are killed or proven E/R, then re-measure.

`U_BUDGET(tier)` is an **owner decision** (§6), per-tier, first values explicitly **provisional and
conservative** — not empirically validatable until the loop exists to correlate trusted locks
against regenerated-impl escapes. The `detector_power` ratio is reported alongside as diagnostic only.

## 4. Interface (proposed — binds in the kernel SPEC at implementation)

A measurement over **committed, calibrated, execution-bound sweep evidence**, not a live re-sweep per
`make check` (embedded-postgres packages run a server per mutant). Proposed shapes:

- **Raw sweep artifact** — verbatim stdout of `scripts/mutants.sh <pkg>` (per-mutant
  `path:line#idx killed|survived|invalid` lines, the `calibration …` line, the `summary candidates=..
  killed=.. invalid=.. survived=..` line), committed at `docs/verification/detector-power/<pkg>.sweep.txt`.
  Source of truth; C/K/V/S + calibration are **re-derived from it** (F5), never hand-copied.
- **Sweep-evidence record** `docs/verification/detector-power/<pkg>.json`:
  `{package, swept_at_commit, lock_inputs_sha256, raw_sweep_path, raw_sweep_sha256,
  execution_binding:{mode:"reproducible"|"attested", re_runner?}, survivors:[{id, class:"E"|"R"|"U",
  justification, pins?}], unproven_latitude_count, detector_power_diagnostic, u_budget_tier, trusted}`.
  `lock_inputs_sha256` covers **non-test source + `_test.go` + this package's `allowed-survivors.txt`
  entries + its SPEC invariant/lock rows** (N2).
- `scripts/detector-power.sh <pkg>` — re-hashes the raw sweep; re-derives C/K/V/S + calibration from
  it; recomputes `U` and the diagnostic from the declared classes; recomputes `lock_inputs_sha256`
  from the current inputs and **rejects a stale record** (N2); enforces the survivor↔record bijection
  against the sweep and `allowed-survivors.txt`; verifies the **execution binding** (mode
  `reproducible`: re-run matches `raw_sweep_sha256`; mode `attested`: a named non-author re-runner is
  recorded and the Law-4 advisory row is unexpired); fails closed on any mismatch, missing
  calibration, unclassified survivor, or `trusted:true` above budget.
- A `docs/gates.md` row (advisory first, Law 4: owner + ISO expiry).

Shell-gate vs a small `kernel/internal/detectorpower` Go package is a §6 open decision; the fold +
re-derivation are unit-tested either way.

## 5. Invariants (proposed contract — become INV-<N> + `invariants.lock` on acceptance)

| # | Invariant | Proving test (shape) |
|---|---|---|
| DP-1 | The gate keys on the **absolute class-U count** `U`; `detector_power = K/(K+U)` is a reported diagnostic only and never drives trust; `K+U=0` ⇒ rejected | `detectorpower_test.go::TestGateKeysOnAbsoluteU` |
| DP-2 | A survivor in the sweep absent from the record (unclassified) fails closed — never a score, never trusted | `detectorpower_test.go::TestUnclassifiedSurvivorFailsClosed` |
| DP-3 | Exact bijection sweep-survivors ↔ record-classified survivors (extra record entry also fails closed); each R names a pinning test/line, each E an equivalence argument, bound to `allowed-survivors.txt` | `detectorpower_test.go::TestSurvivorRecordBijection` |
| DP-4 | `trusted` is true iff `U ≤ U_BUDGET(tier)` with every class-U survivor justified; above budget is never trusted | `detectorpower_test.go::TestTrustGateRespectsUBudget` |
| DP-5 | C/K/V/S + calibration are **re-derived from the committed raw sweep stdout**, whose digest the record pins; a record disagreeing with its raw sweep fails closed | `detectorpower_test.go::TestReDerivesFromRawSweep` |
| DP-6 | A record whose pinned `lock_inputs_sha256` (non-test source + `_test.go` + register entries + SPEC invariant rows) ≠ the current inputs is **stale** and rejected | `detectorpower_test.go::TestStaleRecordRejected` |
| DP-7 | Calibration (lethal=killed, neutral=survived, invalid=invalid) must appear in the raw sweep or the record is rejected | `detectorpower_test.go::TestRejectsUncalibratedSweep` |
| DP-8 | The raw sweep is **execution-bound**: mode `reproducible` ⇒ a refresh re-runs the sweep at `lock_inputs_sha256` and reproduces `raw_sweep_sha256`; mode `attested` ⇒ a named non-author re-runner is recorded and the Law-4 advisory limitation row is unexpired. A sweep bound to neither fails closed | `detectorpower_test.go::TestRawSweepExecutionBound` |

(Permanent ids assigned at the kernel-SPEC + lock append under the change intent — not in this draft.)

## 6. Open decisions (owner / review)
1. **`U_BUDGET` per tier** — owner decision; first values explicitly provisional/conservative (§3).
2. **Shell-gate vs `kernel/internal/detectorpower` Go package** (§4).

## 7. Non-goals (this piece) + dependencies the loop spec MUST discharge
- **No regeneration loop.** This piece only *measures* lock strength. The follow-on loop spec MUST
  discharge two assumptions this gate rests on (F6): **(a)** the mutation **operator failure
  distribution must plausibly represent the regenerator's failure modes**; **(b)** the loop must
  **not train/overfit against the measured tests** (else the trust is circular). Until then, a
  budget-compliant `U` authorizes nothing by itself.
- **No live re-sweep in `make check`** — detector power reads committed, calibrated, freshness- and
  execution-bound evidence refreshed out-of-band.
- **No change to the zero-undeclared-survivors bar** — this is an *additional*, higher bar.
- **No automatic lock strengthening** — raising a below-budget lock is authored work.

## 8. Sequence & gates
1. **Draft 3 → non-author re-review** (round 3; N1–N4 are the deltas — the absolute-U gate, the
   freshness scope, the execution binding, and the bijection/R-pinning).
2. Owner settles §6 (per-tier `U_BUDGET`, shell-vs-package); owner-approved change intent for the
   governed `scripts/*`/`gates/*` (+ `kernel/*` if a package) work.
3. Implement spec-first (kernel SPEC.md + `invariants.lock` append under the intent, or the shell gate
   + self-test); add the E/R/U class + `pins` fields to the survivor register; run the deferred
   full-package `internal/twin` sweep and classify its survivors to seed the first execution-bound
   records; register the advisory `docs/gates.md` row.
4. Package acceptance: calibrated `make mutants` of new code + non-author verdict.
5. **Only then** does the follow-on regeneration-loop spec begin, consuming budget-compliant packages
   and discharging the §7 dependencies.
