---
schema: vera.verdict.v1
verdict_id: p7-2b-detector-power-spec-review-round2-20261105
status: NEEDS_CHANGES
reviewed_commit: 986d6d787bc56b3e93e38bf268abd1a7ba619cb9
findings:
  - finding_id: p7-2b-detector-power-spec-review-round1-F1
    severity: HIGH
    summary: Category error / perverse defense-in-depth incentive — RESOLVED. E/R/U classification, detector_power = K/(K+U) counting only class-U, E/R excluded; deleting a redundant check no longer raises the score
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F2
    severity: HIGH
    summary: §6.3 contradiction — RESOLVED. The denominator question is decided by DP-1; §6 explicitly records it as RESOLVED, not open
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F3
    severity: HIGH
    summary: Freshness binding — RESOLVED IN PRINCIPLE by DP-6 / source_tree_sha256 stale-reject, but the digest SCOPE is under-specified (see N2)
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F4
    severity: MED
    summary: Necessary-not-sufficient — RESOLVED. §3 states it explicitly; the "trust the lock to catch a wrong candidate" over-claim is withdrawn
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F5
    severity: MED
    summary: Hand-transcription — RESOLVED IN PART by committed raw sweep stdout + re-derivation (DP-5), but the raw sweep is not bound to a real/reproducible execution (see N3)
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F6
    severity: MED
    summary: Loop-dependence — RESOLVED. §7 names both assumptions (operator distribution ≈ regenerator failure modes; no overfit against measured tests) as dependencies the loop spec MUST discharge
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F7
    severity: LOW
    summary: Zero denominator — RESOLVED. K+U=0 ⇒ rejected (§2, DP-1)
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round1-F8
    severity: LOW
    summary: Threshold — RESOLVED. Per-tier, explicitly provisional/conservative (§3, §6). But the ratio form has a within-tier size-dependence the per-tier split does not cure (see N1)
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round2-N1
    severity: MED
    summary: §2/§3 the trust criterion is a RATIO K/(K+U) but the exploitable quantity is the ABSOLUTE count U of unproven-latitude blind spots; the ratio dilutes U by K, so the same blind spot passes in a large/verbose package and fails in a small one, detector power can be inflated by adding easily-killed code, and a genuinely weak lock whose blind spots fall outside the mutation-operator basis scores 1.0 with no mis-declaration. Gate on U (U=0 or a per-tier U-budget, each U justified) with the ratio as a secondary diagnostic, or justify the ratio and add an anti-padding guard
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round2-N2
    severity: MED
    summary: DP-6 freshness digest scope is under-specified — "current package source" plausibly excludes the two inputs that most change the result: the test files (the lock IS the tests) and the E/R/U survivor register. A test-only weakening or a stale register then escapes stale-reject, defeating F3 for its most sensitive input. State that source_tree_sha256 covers _test.go and the register (and the invariant/lock definition)
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round2-N3
    severity: MED
    summary: §4/DP-5 the committed raw sweep stdout is re-derived-from but not bound to a real execution — a hand-authored sweep.txt with fabricated per-mutant lines, a calibration line, and an all-E/R survivor set re-derives self-consistently and passes. The regeneration-TRUST gate then rests on forgeable evidence (the B1 unit-3 hand-transcription class). Require the raw sweep be REPRODUCIBLE from source_tree_sha256 (a refresh/audit re-runs scripts/mutants.sh and matches raw_sweep_sha256), or state explicitly it is a human-attested artifact whose integrity depends on a named non-author re-run
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round2-N4
    severity: LOW
    summary: §2/DP-3 the R class carries the most judgment and is the mis-declaration vector (label a U as R to exclude it); draft 2's own "falsifiable claim" promise is strongest if each R justification names the specific pinning line/test so the reviewer can confirm that line's mutant dies. Also state DP-2+DP-3 enforce an exact bijection sweep-survivors ↔ record-classified-survivors (no extra, no missing)
    verdict: CONFIRMED
artifact_path: docs/verification/verdicts/p7-2b-detector-power-spec-review-round2-20261105.md
artifact_sha: 08af1955f8815994a3f80c17d7bfd491720616297633d52865f992eda2cf634b
---

# P7 #2b detector-power metric + regeneration-trust gate (spec draft 2) — non-author re-review, round 2 — NEEDS_CHANGES

## Provenance

- **Under review:** `docs/plans/P7-2b-detector-power-spec-draft1.md`, now **DRAFT 2** — a revision in
  place of the draft that round 1 returned NEEDS_CHANGES. Design spec only (Law 6 spec-before-code):
  no `kernel/*`, `scripts/*`, or `gates/*` code exists or is under review.
- **Reviewed commit:** `986d6d787bc56b3e93e38bf268abd1a7ba619cb9` (`986d6d7`). Round-1 reviewed
  `d110465`. HEAD unchanged at dispatch and at this write; no tree modification during review.
- **Driving mechanism:** `/proofbound-review`, round 2 — adversarial re-verify of the round-1
  findings F1–F8 plus a hunt for defects draft 2 introduced. DESIGN-SPEC review: the thing to break
  is the SOUNDNESS OF A METRIC.
- **Reviewer identity:** independent non-author reviewer dispatched via the Agent tool; same reviewer
  as round 1. I did NOT author draft 1, draft 2, or any revision (Build Law 9). This verdict is a
  rejection-with-findings (NEEDS_CHANGES); it is not an acceptance, so it raises no self-acceptance
  concern.
- **Cleanroom redaction:** none required; no absolute paths or usernames below.

## Overall verdict: NEEDS_CHANGES — narrow, near-acceptance

**All eight round-1 findings are genuinely (not cosmetically) resolved.** The central fix is real:
the E/R/U classification with `detector_power = K/(K+U)` counting only unproven-latitude survivors is
exactly the "killable-or-unproven denominator" the round-1 position demanded, it removes the
perverse defense-in-depth incentive, and a perfect lock can now reach 1.0. The metric's **core is
sound**.

Draft 2 is held back only by three MED refinements that the F3/F5 fixes left incomplete and one
soundness choice the spec asserts but does not justify: **N1** (the trust criterion is a ratio, but
the exploitable quantity is the absolute count U — the ratio is dilutable and size-dependent),
**N2** (the new freshness digest's scope is under-specified and may exclude the tests that ARE the
lock), and **N3** (the new raw-sweep evidence is re-derived-from but not bound to a real execution,
so a trust gate rests on forgeable evidence). None is a category error or a contradiction; each is a
specific binding to tighten. A round 3 addressing N1–N3 (N4 is a LOW tightening) should accept.

## Round-1 findings — per-finding re-verify

- **F1 (HIGH) — RESOLVED.** §2 classifies survivors E (irreducibly equivalent, un-killable) / R
  (redundantly-pinned, behavior pinned by a second tested line) / U (genuine un-pinned latitude), and
  `detector_power = K/(K+U)` counts **only class-U**. This is the round-1 "killable-or-unproven"
  direction. The perverse incentive is gone: deleting a redundant R check now removes R survivors
  (uncounted) and some killed mutants (K drops), so the score is neutral-to-slightly-lowered, never
  raised. The flagship manifest example is re-derived correctly (5 survivors E/R → U=0 → 1.0), so the
  datapoint no longer demonstrates the bug. **Does the classification merely relocate the judgment?**
  Yes — from "is this equivalent" to "is this E, R, or U" — and draft 2 is honest about this (§2 final
  paragraph: the class is "a falsifiable assertion a non-author checks … mis-declaring U as R … is a
  reviewable defect"). That is the correct and only available trust model for a semantic judgment, and
  it matches today's governed equivalent-survivor justifications. The relocation is acceptable; the
  residual sharpenings are N1 (the honest-1.0-on-a-weak-lock path, which needs no mis-declaration) and
  N4 (make R maximally falsifiable). F1 itself is resolved.
- **F2 (HIGH) — RESOLVED.** DP-1 fixes `K/(K+U)`; §6 carries only the threshold and shell-vs-package
  decisions and states explicitly that the denominator question "is now RESOLVED by DP-1, not open."
  No invariant is left contradicting an open option.
- **F3 (HIGH) — RESOLVED in principle; scope gap → N2.** DP-6 rejects a record whose pinned
  `source_tree_sha256` ≠ the current source, and §4 has `detector-power.sh` re-derive the digest and
  stale-reject. The mechanism is correct; only its **scope** is under-specified (N2).
- **F4 (MED) — RESOLVED.** §3 is titled "necessary, not sufficient," states high detector power is "a
  **necessary precondition** … **not a sufficient guarantee**," cites vision line ~156, and defers the
  other safeguards to the loop spec. The round-1 over-claim is withdrawn. §2's "a perfect lock now can
  [reach 1.0]" is consistent with this (it claims reachability, not sufficiency).
- **F5 (MED) — RESOLVED in part; execution-binding gap → N3.** §4 commits the verbatim sweep stdout
  (`<pkg>.sweep.txt`), pins its digest, and DP-5 re-derives C/K/V/S and the calibration outcome **from
  it** rather than from hand-copied JSON. This is a real improvement. The residual is that the raw
  stdout itself is not bound to an actual tool run (N3).
- **F6 (MED) — RESOLVED.** §7 names both loop assumptions verbatim — (a) operator failure
  distribution ≈ regenerator failure modes, (b) no training/overfit against the measured tests — as
  dependencies the follow-on loop spec MUST discharge, and states "a high detector power authorizes
  nothing by itself" until it does.
- **F7 (LOW) — RESOLVED.** `K+U=0 ⇒ rejected (never trusted)` in both §2 and DP-1.
- **F8 (LOW) — RESOLVED.** Threshold is per-tier and explicitly provisional/conservative (§3, §6.1),
  with the honest note that it "cannot be empirically validated until the loop exists." (The ratio's
  within-tier size-dependence is a separate issue — N1.)

## New findings (draft 2)

### N1 (MED) — the trust criterion is a ratio, but the exploitable quantity is the absolute U
- **Where:** §2 (`detector_power = K/(K+U)`), §3 (threshold gate).
- **Why:** U is the count of genuine un-pinned degrees of freedom a regenerated impl can
  satisfy-and-reopen — that is the risk, in absolute terms. The gate instead keys on the **ratio**
  `K/(K+U) = 1 − U/(K+U)`. Three consequences the spec does not address:
  1. **Size-dependence.** One U survivor scores 0.91 in a package with 10 killed mutants and 0.999 in
     one with 1000 — identical blind spot, identical regeneration risk, opposite verdict against any
     fixed per-tier threshold. Per-tier (F8) does not cure within-tier verbosity differences.
  2. **Dilution / padding.** Adding well-tested trivial code raises K while U is unchanged, so the
     ratio rises toward 1.0 and the *tolerated* number of unproven blind spots grows with K. Detector
     power is improvable without reducing real latitude.
  3. **Honest 1.0 on a weak lock (the answer to "construct a weak lock scoring 1.0 without
     mis-declaration").** A blind spot that no mutation operator reaches produces **no survivor at
     all**, so U=0 and the score is 1.0 though the lock is genuinely weak there — no mis-declaration
     required. §3's necessary-not-sufficient caveat covers this in prose, but §2 still presents 1.0 as
     the strong end of the scale ("a perfect lock now can").
- **Fix:** make the gate key on **U directly** — `U = 0`, or a small per-tier **U-budget** with each
  budgeted U separately justified — and keep `detector_power` as a secondary diagnostic. If the ratio
  is retained as the gate, justify why a size-scaled tolerance for unproven latitude is acceptable and
  add an anti-padding guard (e.g. measure U against the count of *public-contract* mutants, not all
  killed mutants). Soften §2's "a perfect lock now can reach 1.0" to "can reach 1.0 over the operator
  basis," so 1.0 is not read as strength.

### N2 (MED) — DP-6 freshness digest scope is under-specified; it likely excludes the lock itself
- **Where:** §4 / §5 DP-6 (`source_tree_sha256`, "current package source").
- **Why:** detector power is a property of (implementation shape × **test suite** × declared classes),
  and the lock IS the tests + invariants, not the implementation. If `source_tree_sha256` hashes only
  non-test implementation source, then **weakening a test** (the most direct way to make a lock weaker)
  leaves the digest unchanged, the prior high-scoring sweep stays "fresh," and the stale-reject never
  fires — defeating F3 for exactly the input it most needs to cover. Likewise a change to the E/R/U
  **register** (a U→R relabel) should force re-validation. "Package source" is ambiguous: in Go the
  package directory includes `_test.go`, so the natural reading *might* cover tests — but on a
  load-bearing freshness invariant this must be stated, not inferred.
- **Fix:** state in DP-6 that `source_tree_sha256` covers the whole package including `_test.go`, the
  survivor register entries for the package, and the invariant/lock definition — i.e. every input whose
  change could change K, U, or a class. Anything excluded must be argued not to affect the measurement.

### N3 (MED) — the raw sweep is re-derived-from but not bound to a real execution (forgeable trust evidence)
- **Where:** §4 raw sweep artifact + record, DP-5.
- **Why:** committing verbatim stdout and re-deriving C/K/V/S from it (F5) closes hand-copied-number
  drift, but nothing binds `<pkg>.sweep.txt` to an actual `scripts/mutants.sh` run. A hand-authored
  sweep file with fabricated per-mutant `path:line#idx killed` lines, a `calibration …` line, and a
  summary whose survivors are all declared E/R will **re-derive self-consistently**, pass DP-5 and
  DP-7 (calibration lines are present because they were typed), and yield detector_power = 1.0. This is
  the same hand-attestation gap the B1 unit-3 review treated as load-bearing, and here it sits under
  the word *trust* in "regeneration-trust gate."
- **Fix:** require the raw sweep be **reproducible** from `source_tree_sha256` — a refresh/audit re-runs
  `scripts/mutants.sh <pkg>` and asserts the stdout matches `raw_sweep_sha256` (mutation sweeps are
  deterministic for deterministic tests; state the determinism assumption). If full reproduction is out
  of scope for this piece, state explicitly that the raw sweep is a **human-attested** artifact whose
  integrity depends on a **named** non-author having re-run it, and register that as a known limitation
  (Law 4) rather than implying the re-derivation makes the evidence self-proving.

### N4 (LOW) — make the R class maximally falsifiable; state the survivor bijection
- **Where:** §2 classification, DP-2, DP-3.
- **Why:** R is the mis-declaration vector (relabel a U as R to exclude it from the denominator). Draft
  2's own "falsifiable claim" promise is strongest when an R justification is mechanically checkable.
  Separately, soundness of `K/(K+U)` requires that every sweep survivor is classified exactly once and
  no register entry lacks a matching survivor.
- **Fix:** require each R justification to name the specific pinning line/test whose mutant dies (so the
  reviewer confirms the downstream mutant is actually killed), and state that DP-2 + DP-3 together
  enforce an exact bijection between the raw sweep's survivors and the record's classified survivors
  (an extra register entry or an unclassified survivor both fail closed).

## What is sound (survives review)

- The E/R/U classification + `K/(K+U)` is the correct resolution of the round-1 category error; the
  perverse defense-in-depth incentive is removed and the flagship example is re-derived honestly.
- §3's necessary-not-sufficient framing and the §7 loop-dependency discharge (F4/F6) are correct and
  honest.
- DP-6 (freshness), DP-5 (re-derive from raw sweep), and DP-7 (calibration in the raw sweep) are the
  right *mechanisms*; N2/N3 only tighten their scope and their binding to reality.
- Spec-before-code discipline holds: no code, permanent INV ids correctly deferred to the kernel-SPEC
  append under an owner-approved change intent; §8 correctly flags the register class-field addition as
  governed implementation work.

## Recommendation

**NEEDS_CHANGES — round 3, narrow.** Round-1 F1–F8 are all resolved; the metric's core is accepted as
sound. The blocking items are N1 (key the gate on absolute U or justify + guard the ratio), N2 (state
the freshness digest covers the tests and the register), and N3 (make the raw sweep reproducible or
label it human-attested with a named re-runner). N4 is a low-cost tightening to fold into the same
revision. None requires re-opening the E/R/U design. Once revised, a fresh non-author verdict
adjudicates; only then do the §6 owner decisions and an owner-approved change intent authorize
implementation.

(Note: I am a non-author reviewer. This NEEDS_CHANGES verdict is a rejection with findings; per Law 9
the author may revise and a non-author adjudicates the revision. I made no edit to the spec.)
