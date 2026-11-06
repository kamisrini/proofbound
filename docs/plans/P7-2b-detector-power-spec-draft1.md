# P7 #2b — detector-power metric + regeneration-trust gate — design spec (DRAFT 2)

**Status:** DRAFT 2 — revises draft 1 after the round-1 non-author verdict
([`verdicts/p7-2b-detector-power-spec-review-round1-20261105.md`](../verification/verdicts/p7-2b-detector-power-spec-review-round1-20261105.md),
NEEDS_CHANGES at `d110465`). Design spec only (Law 6 spec-before-code): no `kernel/*`, `scripts/*`,
or `gates/*` code, no change intent accepted. First piece of vision build-order **#2b** (witnessed
regeneration); the regeneration **loop** is deferred to a follow-on spec (§7). Implementation is
gated on this spec's non-author acceptance + an owner-approved change intent.
**Date:** 2026-11-05
**Authorities:** [vision-2028.md](../../vision-2028.md) §9 item 2b (and its line ~156 — the ladder can
be climbed while detector power → 0); completion plan
[product-completion-plan-2026-11-05.md](product-completion-plan-2026-11-05.md) §A. Builds on
`tools/mutants` / `scripts/mutants.sh`, `docs/allowed-survivors.txt`, `docs/gates.md`.

## 0. What draft 2 changed (round-1 findings)

- **F1 (HIGH) — the metric was a category error** (the central fix). Draft 1 put *all* survivors in
  the denominator, which conflated three different things and **rewarded deleting defense-in-depth**
  (removing a redundant check raised the score). Draft 2 classifies each survivor and counts **only
  genuinely-unproven latitude** against detector power (§2, §5 DP-1, new survivor classes).
- **F2 (HIGH)** — the "is the denominator right" question is no longer a §6 open decision; DP-1
  resolves it (§2/§6).
- **F3 (HIGH)** — added a **freshness binding** (DP-6): a record is rejected unless its pinned package
  source digest equals the current source (§4, §5).
- **F4 (MED)** — detector power is now stated **necessary-not-sufficient**; the "trust the lock to
  catch a wrong candidate" over-claim is withdrawn (§3).
- **F5 (MED)** — the record binds to **committed raw sweep stdout** and C/K/V/S + calibration are
  **re-derived from it**, not copied into hand-authored JSON (§4, §5 DP-5).
- **F6 (MED)** — the two loop assumptions (operator distribution ≈ regenerator failure modes; no
  overfitting against the measured tests) are named as explicit **dependencies the follow-on loop
  spec MUST discharge** (§7).
- **F7/F8 (LOW)** — zero-denominator defined as rejected (§2); threshold explicitly provisional and
  per-tier (§3, §6).

## 1. Purpose — why detector power must exist before regeneration

#2b discards a failing implementation and **re-derives it against the lock + the new counterexample**
instead of patching. The vision gate: *"regeneration is only as good as the lock … [our best lock]
still shipped five residual items, any of which a regenerated implementation could
satisfy-and-reopen. So 2b unlocks only where aggregate detector power is measured and high."* The
prerequisite is a **measurement of lock strength**: can the lock reject a *wrong* implementation? A
regenerated impl is a different impl; where the lock has genuine blind spots, it can satisfy the lock
while being wrong. This spec defines that measurement and the threshold gate.

## 2. The metric

**Why the obvious definitions both fail.** Naive `killed/killable` is trivially 1.0 (undeclared
survivors are already banned), so it cannot discriminate. Draft 1's `killed/(candidates−invalid)`
over-corrected: it penalized every survivor, but a survivor is one of **three kinds the raw sweep
cannot distinguish** (round-1 F1):

1. **equivalent (E)** — semantically identical code, **un-killable by any test**; zero exploitable
   latitude.
2. **redundantly-pinned (R)** — the behavior **is** pinned, by a second tested line (defense-in-depth,
   e.g. a check also enforced by a downstream `Bind→Validate`); removing it changes nothing a test
   can see, but the behavior is not un-pinned. 4 of the 5 `internal/twin/manifest` survivors are this.
3. **unproven latitude (U)** — a genuine **un-pinned degree of freedom**: the lock does not constrain
   this behavior and no second line does either. This is the only kind a regenerated impl can
   satisfy-and-reopen.

Only **U** is exploitable latitude. E and R are not, and penalizing R actively **rewards deleting
defense-in-depth** — a perverse incentive. So each survivor carries a **declared class** (E/R/U) in
the register (§4), each class independently justified, and:

> **detector_power(P) = K / (K + U)** , where `K` = killed, `U` = count of class-**U** survivors.
> Rejected (never trusted) when `K + U = 0` (no evidence) — F7.

E, R, and invalid (`V`) are excluded from the denominator. A lock with **no unproven latitude**
reaches `1.0` (a perfect lock now *can*, fixing draft 1's category error); each class-U survivor
drops it. Datapoints (illustrative, pending the real classified sweep): `internal/twin/manifest` —
all 5 survivors E/R → `U=0` → `1.0`; the `gitcmd` register has a survivor explicitly labeled "a
JUDGEMENT rather than a proof" → class U → `< 1.0`.

**The classification is the lock's own claim and is itself reviewed.** Declaring a survivor E or R
(to exclude it) is a falsifiable assertion a non-author checks — mis-declaring U as R to inflate the
score is a reviewable defect, exactly like today's equivalent-survivor justifications. Detector power
does not trust the class labels blindly; it makes them load-bearing and auditable.

## 3. The regeneration-trust gate (necessary, not sufficient)

`detector_power(P) ≥ THRESHOLD(tier)` is a **necessary precondition** for `P` to be
regeneration-trusted — **not a sufficient guarantee** (round-1 F4). Mutation operators are a finite
basis; a lock can kill every *generated* mutant and still miss a blind spot no operator reaches
(vision line ~156: the ladder can be climbed while true detector power falls). So this gate is one
required safeguard; the follow-on regeneration-loop spec must add others (§7). Below threshold,
regeneration is **not** authorized for `P`: strengthen the lock (add invariants / tighten tests)
until the class-U survivors fall or are proven E/R, and re-measure.

`THRESHOLD` is an **owner decision** (§6), **per-tier** not one global number (F8), and its first
value is an explicitly **provisional conservative placeholder** — it cannot be empirically validated
until the loop exists to correlate trusted locks against regenerated-impl escapes.

## 4. Interface (proposed — binds in the kernel SPEC at implementation)

A measurement over **committed, calibrated sweep evidence**, not a live re-sweep per `make check`
(sweeps run a server per mutant for embedded-postgres packages; see §7). Proposed shapes:

- **Raw sweep artifact** — the committed **verbatim stdout** of `scripts/mutants.sh <pkg>` (the
  per-mutant `path:line#idx killed|survived|invalid` lines, the `calibration …` line, and the
  `summary candidates=.. killed=.. invalid=.. survived=..` line), under
  `docs/verification/detector-power/<pkg>.sweep.txt`. This is the source of truth; C/K/V/S and the
  calibration outcome are **re-derived from it** (F5), never hand-copied.
- **Sweep-evidence record** `docs/verification/detector-power/<pkg>.json`:
  `{package, swept_at_commit, source_tree_sha256, raw_sweep_path, raw_sweep_sha256,
  survivors:[{id, class:"E"|"R"|"U", justification}], detector_power, threshold_tier, trusted}`.
- `scripts/detector-power.sh <pkg>` — re-hashes the raw sweep, re-derives C/K/V/S from it, recomputes
  `detector_power` from the declared classes, re-derives `source_tree_sha256` from the current
  package source and **rejects a stale record** (F3), cross-checks the survivor set against
  `docs/allowed-survivors.txt` (F-class drift), and fails closed on any mismatch, missing calibration,
  or a `trusted:true` below threshold.
- A `docs/gates.md` row (advisory first, Law 4: owner + ISO expiry).

Shell-gate vs a small `kernel/internal/detectorpower` Go package is a §6 open decision; the formula +
classification fold are unit-tested either way.

## 5. Invariants (proposed contract — become INV-<N> + `invariants.lock` on acceptance)

| # | Invariant | Proving test (shape) |
|---|---|---|
| DP-1 | `detector_power = K/(K+U)`; only class-**U** survivors count; E/R/invalid are excluded; `K+U=0` ⇒ rejected | `detectorpower_test.go::TestFormulaCountsOnlyUnprovenLatitude` |
| DP-2 | An **undeclared** survivor (present in the sweep, absent from the register) fails closed — never a low score, never trusted | `detectorpower_test.go::TestUndeclaredSurvivorFailsClosed` |
| DP-3 | Every survivor's class + justification is bound to a `docs/allowed-survivors.txt` entry; register/record drift fails closed | `detectorpower_test.go::TestClassesBoundToRegister` |
| DP-4 | `trusted` is true iff `detector_power ≥ threshold(tier)`; below-threshold is never trusted | `detectorpower_test.go::TestTrustGateRespectsThreshold` |
| DP-5 | C/K/V/S and the calibration outcome are **re-derived from the committed raw sweep stdout**, whose digest the record pins; a record disagreeing with its raw sweep fails closed | `detectorpower_test.go::TestReDerivesFromRawSweep` |
| DP-6 | A record whose pinned `source_tree_sha256` ≠ the current package source is **stale** and rejected (never trusted) | `detectorpower_test.go::TestStaleRecordRejected` |
| DP-7 | Calibration must show lethal=killed, neutral=survived, invalid=invalid in the raw sweep, or the record is rejected (an uncalibrated sweep proves nothing) | `detectorpower_test.go::TestRejectsUncalibratedSweep` |

(Permanent ids are assigned at the kernel-SPEC + lock append, under the change intent — not in this
draft.)

## 6. Open decisions (owner / review)
1. **THRESHOLD per tier** — owner decision; first values explicitly provisional/conservative (§3, F8).
2. **Shell-gate vs `kernel/internal/detectorpower` Go package** (§4).

*(Draft 1's "does the denominator penalize equivalents" item is now RESOLVED by DP-1, not open — F2.)*

## 7. Non-goals (this piece) + dependencies the loop spec MUST discharge
- **No regeneration loop.** This piece only *measures* lock strength. The follow-on loop spec MUST
  explicitly discharge two assumptions this gate silently rests on (round-1 F6): **(a)** the mutation
  **operator failure distribution must plausibly represent the regenerator's failure modes** (else
  detector power measures the wrong population of mistakes); **(b)** the loop must **not train or
  overfit against the measured tests** (else the trust is circular — a regenerator tuned to pass the
  exact tests that define detector power learns nothing about the un-pinned latitude). Until the loop
  spec addresses both, a high detector power authorizes nothing by itself.
- **No live re-sweep in `make check`** (sweeps are slow); detector power reads committed, calibrated,
  freshness-bound evidence refreshed out-of-band.
- **No change to the zero-undeclared-survivors bar** — detector power is an *additional*, higher bar
  for regeneration trust, not a replacement.
- **No automatic lock strengthening** — raising a below-threshold lock is authored work.

## 8. Sequence & gates
1. **Draft 2 → non-author re-review** (adversarial; the E/R/U classification soundness and the
   freshness/raw-sweep bindings are the new things to break).
2. Owner settles §6 (per-tier thresholds, shell-vs-package); owner-approved change intent for the
   governed `scripts/*`/`gates/*` (+ `kernel/*` if a package) work.
3. Implement spec-first (kernel SPEC.md with permanent INV ids + `invariants.lock` append under the
   intent, or the shell gate + self-test); add the E/R/U class field to the survivor register; run the
   deferred full-package `internal/twin` sweep and classify its survivors to seed the first records;
   register the advisory `docs/gates.md` row.
4. Package acceptance: calibrated `make mutants` of new code + non-author verdict.
5. **Only then** does the follow-on regeneration-loop spec begin, consuming `trusted` packages and
   discharging the §7 dependencies.
