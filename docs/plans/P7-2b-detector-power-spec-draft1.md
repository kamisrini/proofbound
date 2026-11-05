# P7 #2b — detector-power metric + regeneration-trust gate — design spec (DRAFT 1)

**Status:** DRAFT for non-author review — design spec only (Law 6 spec-before-code). No `kernel/*`,
`scripts/*`, or `gates/*` code is written by this document, and no change intent is accepted yet.
This is the **first** piece of vision build-order **#2b** (witnessed regeneration /
regeneration-as-repair); the regeneration loop itself is **explicitly deferred to a follow-on spec**
(§7). Implementation is gated on this spec's non-author acceptance + an owner-approved change intent.
**Date:** 2026-11-05
**Authorities:** [vision-2028.md](../../vision-2028.md) §9 item 2b; the completion plan
[product-completion-plan-2026-11-05.md](product-completion-plan-2026-11-05.md) §A. Builds on the
existing mutation harness (`tools/mutants`, `scripts/mutants.sh`), the survivor register
(`docs/allowed-survivors.txt`), and the gate registry (`docs/gates.md`).

## 1. Purpose — why detector power must exist before regeneration

Build-order #2b discards a failing implementation and **re-derives it against the lock + the new
counterexample** instead of patching. The vision states the gate plainly: *"regeneration is only as
good as the lock … our best lock — 1,609 lines, 70 invariant rows — still shipped five residual
items, any of which a regenerated implementation could satisfy-and-reopen. So 2b unlocks only where
aggregate detector power is measured and high."*

So the prerequisite is a **measurement of lock strength**: can the lock (the spec invariants and
their derived tests) reject a *wrong* implementation? A regenerated implementation is a different
implementation; if the lock has blind spots, a regenerated impl can satisfy the lock while being
wrong (satisfy-and-reopen). This spec defines that measurement and the threshold gate that decides
whether a given package/lock is **regeneration-trusted**.

## 2. The metric (and why the naive one is useless)

**Naive kill-rate is trivially 1.0.** An accepted package already has **zero undeclared survivors**
(the P1/Task-8 acceptance bar + `make mutants` gate require it), so `killed / killable` is 100% by
construction for everything already shipped — it cannot discriminate lock strength.

**Detector power penalizes survivors (the blind spots).** Define, from one calibrated mutation sweep
of a package `P`:

```
candidates C, killed K, invalid V, survived S          (from the mutants summary)
```

> **detector_power(P) = K / (C − V)**

Survivors — including the *declared-equivalent* ones in `allowed-survivors.txt` — stay in the
denominator, because **an equivalent survivor is a behavior the lock does not pin**, i.e. exactly the
latitude a regenerated implementation could exploit. Invalid mutants (`V`, uncompilable) are excluded
as not-semantically-meaningful. Worked example (the accepted `internal/twin/manifest` sweep,
2026-10-31): `C=190, V=0, S=5, K=185 → detector_power = 185/190 = 0.9737`.

A lock with many equivalent survivors scores lower and is **not** regeneration-trusted even though it
passes the zero-undeclared-survivors bar; a lock that pins nearly every semantic mutation scores high.

## 3. The regeneration-trust gate

`detector_power(P) ≥ THRESHOLD` makes `P` **regeneration-trusted** (the follow-on regeneration loop
may re-derive `P`'s implementation and trust the lock to catch a wrong candidate). Below threshold,
regeneration is **not** authorized for `P` — the lock must first be strengthened (more invariants /
tighter tests) until the blind-spot survivors fall or are proven genuinely equivalent **and** the
residual power still clears the bar.

`THRESHOLD` is an **owner decision** (§6 open decision) — not chosen by this draft. It must be high
enough that the residual latitude is tolerable; the manifest example (0.97) is offered only as a
real datapoint, not a proposed value.

## 4. Interface (proposed — binds in the kernel SPEC at implementation)

A measurement over committed sweep evidence, not a live re-sweep per `make check` (sweeps are slow;
see §Non-goals). Proposed shapes:

- A **sweep-evidence record** per measured package, committed under
  `docs/verification/detector-power/<pkg>.json`: `{package, swept_at_commit, candidates, killed,
  invalid, survived, declared_survivors, detector_power, threshold, trusted: bool}` — the non-binary
  record of a calibrated sweep (the sweep itself runs out-of-band, like the §9.2 build evidence).
- `scripts/detector-power.sh <pkg>` — parses the committed record (or a fresh sweep when invoked
  with the sweep output), recomputes `detector_power`, cross-checks the declared-survivor count
  against `allowed-survivors.txt`, and fails closed on any mismatch or below-threshold "trusted".
- A `docs/gates.md` row registering it (advisory first, per Law 4, with an owner + ISO expiry).

Whether the computation lives in a shell gate or a small `kernel/internal/detectorpower` Go package
is a §6 open decision; this draft leans shell+gate (it orchestrates `tools/mutants` + reads
`allowed-survivors.txt`, mirroring `meta-tax.sh`), with the formula unit-tested either way.

## 5. Invariants (proposed contract — become INV-<N> in the kernel SPEC on acceptance)

| # | Invariant | Proving test (shape) |
|---|---|---|
| DP-1 | `detector_power = killed/(candidates−invalid)`; survivors (declared or not) are in the denominator | `detectorpower_test.go::TestFormulaPenalizesSurvivors` |
| DP-2 | A package with an **undeclared** survivor fails closed (never "trusted"); undeclared survivors are an error, not a low score | `detectorpower_test.go::TestUndeclaredSurvivorFailsClosed` |
| DP-3 | The declared-survivor count in the record must equal the `allowed-survivors.txt` entries for that package (no silent drift) | `detectorpower_test.go::TestDeclaredCountBoundToRegister` |
| DP-4 | `trusted` is true iff `detector_power ≥ threshold`; a below-threshold package is never trusted | `detectorpower_test.go::TestTrustGateRespectsThreshold` |
| DP-5 | Calibration must be present in the sweep evidence (lethal killed, neutral survived, invalid invalid) or the record is rejected — an uncalibrated sweep proves nothing | `detectorpower_test.go::TestRejectsUncalibratedSweep` |

(Permanent ids are assigned when this becomes the kernel SPEC + `invariants.lock` append, under the
change intent — not in this draft.)

## 6. Open decisions (for the owner / review)
1. **THRESHOLD value** — owner decision; needs a rationale tying residual latitude to acceptable
   regeneration risk. Likely per-tier, not one global number.
2. **Shell-gate vs kernel Go package** for the computation (§4).
3. **Equivalent-survivor re-audit:** today a declared-equivalent survivor is justified in prose.
   Detector power makes those survivors *cost* score — does declaring equivalence still exempt them,
   or must a regeneration-trusted lock drive them to zero? (Leaning: they still count against power,
   so "declared equivalent" stops being a free pass for 2b — only for the existing kill gate.)
4. **Advisory→blocking** timeline for the gate (Law 4 expiry).

## 7. Non-goals (this piece)
- **No regeneration loop.** Re-deriving an implementation from a lock + counterexample is the
  follow-on spec; this piece only *measures* whether a lock is strong enough to trust that loop.
- **No live re-sweep in `make check`.** Full sweeps are slow (embedded-postgres packages run a server
  per mutant); detector power reads committed, calibrated sweep evidence, refreshed out-of-band.
- **No change to the zero-undeclared-survivors acceptance bar** — detector power is an *additional*,
  higher bar for regeneration trust, not a replacement.
- **No automatic lock strengthening** — raising a below-threshold lock is authored work, not derived.

## 8. Sequence & gates
1. **This draft → non-author review** (adversarial; the metric's soundness — especially the
   survivor-in-denominator choice and the undeclared-survivor fail-closed — is the thing to break).
2. Owner settles §6 (threshold, shell-vs-package). Owner-approved change intent for the governed
   `scripts/*`/`gates/*` (and `kernel/*` if a package) work.
3. Implement spec-first (kernel SPEC.md with permanent INV ids + `invariants.lock` append under the
   intent, or the shell gate + its self-test); measure the already-accepted packages to seed the
   evidence records; register the `docs/gates.md` row (advisory).
4. Package acceptance: calibrated `make mutants` of any new code + non-author verdict.
5. **Only then** does the follow-on regeneration-loop spec begin, consuming `trusted` packages.
