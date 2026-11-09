---
schema: vera.verdict.v1
verdict_id: p7-2b-regeneration-loop-spec-review-round1-20261109
status: NEEDS_WORK
reviewed_commit: afbeb2010599070121061834e613f8855c1c130f
findings:
  - finding_id: F1
    severity: HIGH
  - finding_id: F2
    severity: MED
  - finding_id: F3
    severity: MED
  - finding_id: F4
    severity: MED
  - finding_id: F5
    severity: MED
  - finding_id: F6
    severity: MED
  - finding_id: F7
    severity: LOW
  - finding_id: F8
    severity: LOW
artifact_path: docs/verification/verdicts/p7-2b-regeneration-loop-spec-review-round1-20261109.md
artifact_sha: c7eee81921007e3a41521a9b11060179d6b0feae6bdedd4d1793e8acfdc3a416
---

# P7 #2b — regeneration-loop spec (BR O-2) draft 1 — non-author review

**NEEDS_WORK** (mapped to the `vera.verdict.v1` vocabulary; the reviewer's term was NEEDS_CHANGES).
An independent non-author reviewer adversarially reviewed `docs/plans/P7-2b-regeneration-loop-spec-draft1.md`
against the Build Laws, O-2 (`BR-regeneration-trust-5f8a03`), the detector-power spec §7, and the
behavior-lock contract. Findings transfer zero trust to any fix (`VD-verification-asymmetry-2dyjnd`);
they state harms, and the author (this session) owns the round-2 revision, which a non-author then
re-reviews.

**Confirmed sound (no finding):** RL-2's no-inheritance re-measurement; both §7 dependencies carried
as named fail-closed invariants (RL-4/D-a, RL-5/D-b); genuinely design-only (no code/provider/event
kind authorized); alignment with the behavior-lock contract (append-only revisions, UNVERIFIED/
UNDECLARED fail-closed); RL-7 applies Law 9 to the final verdict.

## Findings (for the author; round-2)

- **F1 (HIGH) — RL-6 does not prevent satisfy-and-reopen on a non-zero budget.** §4 RL-6 / §2 step 5
  / §3. RL-6 only blocks acceptance below budget. Under a non-zero `U_BUDGET`, a budgeted class-U
  survivor is a known satisfy-and-reopen surface tolerated by a per-survivor "safe to vary" argument
  (detector-power spec CL-4). Scenario uncaught: a counterexample X that falsifies one class-U safety
  argument is pinned into L@r+1; L@r+1 still clears budget on sibling class-U survivors whose safety
  arguments rest on the same flawed reasoning; the candidate passes (X now pinned) while varying a
  sibling surface → satisfy-and-reopen on a budgeted-U surface. No invariant forces re-validation/
  revocation of sibling class-U justifications when a counterexample falsifies one, nor restricts the
  first slice to U=0 tiers (where RL-6 would be airtight). The title "No satisfy-and-reopen" overclaims.
- **F2 (MED) — step 2 conflates "re-measure power" with what `detector-power.sh` does, and hides a
  per-cycle sweep cost.** §2 step 2 / RL-2. `scripts/detector-power.sh` VERIFIES a committed
  execution-bound record; it does not run a sweep. L@r+1 is a new revision minted mid-repair, so RL-2
  requires producing FRESH execution-bound evidence for L@r+1 first — for the named slow
  `internal/twin` slice, that is DP-8 "attested" mode (out-of-band sweep + named non-author re-runner
  + unexpired Law-4 advisory) every cycle. The "Run O-1" wording misrepresents the mechanism and omits
  this latency/independence dependency.
- **F3 (MED) — no invariant that the admitted counterexample is verified-lethal against the failing
  implementation.** §2 step 1 / missing RL. X is only "a reproducible input + the invariant it
  violates." Nothing requires X's pinned line to actually fail on the pre-repair impl. A weak/strawman
  counterexample could be folded as a line that passes even on the broken impl — pinning nothing,
  inflating L@r+1's apparent surface. O-1 requires lethal calibration controls (DP-7); the loop has no
  analog.
- **F4 (MED) — D-b's "false attestation fails closed" is undefined and can collapse to
  self-attestation.** §3 D-b / RL-5. The isolation attestation is a self-asserted claim; the spec
  gives no mechanism to detect a false one, yet asserts it fails closed. O-1 answered the same shape
  with CL-1 (independent non-author re-run, not a sign-off); the loop has no equivalent. Compounding:
  when "the lock IS the tests" (O-1 N2), "derived from the contract, not the tests" is thinly
  separable in a single-context setting — overfitting can re-enter through that seam undetected.
- **F5 (MED) — missing lock-level Law-9 separation between lock-revision authoring and regeneration.**
  §2 steps 1-3 / missing RL. RL-7 pins non-authorship of the final verdict only. Nothing prevents the
  actor that folds X / strengthens L@r+1 from also being the regenerator — grading a candidate against
  its own author's lock.
- **F6 (MED; reviewer: LOW/MED) — Law 2: no home assigned for the D-a representativeness record or the
  D-b isolation attestation.** §3 / §5 item 1. Both are new load-bearing data with no named single
  durable home → drift seed.
- **F7 (LOW; reviewer: LOW/MED) — no invariant binding steps 1, 2, 4 to the same lock-revision
  digest.** §2 / §4. Lock identity is the canonical-bytes SHA-256; nothing requires the L@r+1 measured
  (step 2) to be byte-identical to the one folded (step 1) and evaluated against (step 4). A revision
  swap between measure and evaluate is not explicitly closed.
- **F8 (LOW) — D-a's "evidence" arm is unreachable at genesis.** §3 D-a. Evidence that operators span
  the regenerator's failure classes presupposes a corpus only the running loop produces; every first
  deployment defaults to the bounded-advisory path. The spec should state this openly rather than
  present evidence-vs-advisory as a live genesis choice.

## Reviewer identity

Independent non-author reviewer (fresh-context subagent dispatched by the owning session, 2026-11-09).
Did not author the spec.
