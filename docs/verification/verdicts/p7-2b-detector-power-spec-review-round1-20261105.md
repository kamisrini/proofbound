---
schema: vera.verdict.v1
verdict_id: p7-2b-detector-power-spec-review-round1-20261105
status: NEEDS_WORK
reviewed_commit: d11046559a3dfe25119b5bdc78d7e7a7a5743a98
findings:
  - finding_id: p7-2b-detector-power-spec-review-round1-F1
    severity: HIGH
  - finding_id: p7-2b-detector-power-spec-review-round1-F2
    severity: HIGH
  - finding_id: p7-2b-detector-power-spec-review-round1-F3
    severity: HIGH
  - finding_id: p7-2b-detector-power-spec-review-round1-F4
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round1-F5
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round1-F6
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round1-F7
    severity: LOW
  - finding_id: p7-2b-detector-power-spec-review-round1-F8
    severity: LOW
artifact_path: docs/verification/verdicts/p7-2b-detector-power-spec-review-round1-20261105.md
artifact_sha: b22df2bb544ae1a094abd87bc661c88559daa9c2b2378a869d47749290961f6c
---

# P7 #2b detector-power metric + regeneration-trust gate (spec draft 1) — non-author review, round 1 — NEEDS_CHANGES

## Provenance

- **Under review:** `docs/plans/P7-2b-detector-power-spec-draft1.md` — DRAFT 1, the first piece of
  vision build-order #2b (witnessed regeneration). Design spec only; spec-before-code per Build
  Law 6 — no `kernel/*`, `scripts/*`, or `gates/*` code exists or is under review.
- **Reviewed commit:** `d11046559a3dfe25119b5bdc78d7e7a7a5743a98` (the commit that introduced the
  draft, `d110465`). HEAD unchanged at dispatch and at this write; no tree modification during review.
- **Driving mechanism:** `/proofbound-review` adversarial pass against the Build Laws + the stated
  authorities (vision-2028.md §9 item 2b; `allowed-survivors.txt`; `tools/mutants`). DESIGN-SPEC
  review — the thing to break is the SOUNDNESS OF A METRIC, not code.
- **Reviewer identity:** independent non-author reviewer dispatched via the Agent tool. I did NOT
  author the draft or any part of it (Build Law 9). This verdict is a **rejection with findings**
  (NEEDS_CHANGES); it is not an acceptance, so it raises no self-acceptance concern — the author may
  revise, and a fresh non-author verdict adjudicates the revision.
- **Cleanroom redaction:** none required; no absolute paths or usernames below.

## Overall verdict: NEEDS_CHANGES

The spec is disciplined (spec-first, explicit non-goals, calibration-aware) and its headline critique
is correct: naive kill-rate is 1.0 by construction for any accepted package (§2) and cannot
discriminate lock strength, so a survivor-sensitive metric is the right instrument. But the specific
metric it defines is **not yet sound as a measure of regeneration-trustworthiness**, and it freezes
the one decision that would fix it into an invariant while nominally leaving it open. Three findings
are HIGH (one of them a category error visible on the spec's own flagship example, one an internal
contradiction, one a silent-staleness gate bypass). The metric's *intent* survives review; its
*definition* does not. A round 2 is required before this binds as a kernel SPEC + invariants.lock
append.

## Position on the central question (Q1) — the equivalent-survivor denominator

**Stated unambiguously: keeping declared-equivalent survivors in the denominator, as DP-1 does, is a
category error as written, and must be replaced by a classification step. It should NOT, however, be
"fixed" by removing all survivors from the denominator.**

A mutation survivor is one of three things, and the harness cannot tell them apart — "declared
equivalent" in `allowed-survivors.txt` is a human prose judgment, sometimes explicitly hedged
(gitcmd.go:75#29 is annotated "this one is a JUDGEMENT rather than a proof"):

1. **Irreducibly equivalent** — the mutated program is observationally identical to the original on
   *every* input (e.g. gitcmd.go:312 `|| -> &&`, equivalent by operator precedence). It is
   un-killable by *any* test, present or perfect. It therefore represents **no latitude a regenerated
   implementation could exploit**: an impl differing only this way is behaviorally identical, which is
   exactly what regeneration should be free to do. Penalizing it does not measure a regeneration risk.
2. **Redundantly pinned** — the mutant survives only because a *second, tested* line enforces the same
   behavior (defense in depth). This is **4 of the 5 survivors in the spec's own flagship manifest
   example**: declaration.go:189/192/216/234 all survive because `DeclareCommittedManifest` re-runs the
   same check via `Bind→Validate`, whose mutant *dies*. The accept/reject behavior **is** pinned — a
   regenerated impl that gets it wrong is killed at the downstream test. This too is **not** exploitable
   latitude.
3. **Unproven / judgment latitude** — a survivor declared equivalent that is actually a real behavioral
   degree of freedom the lock does not pin (the ancient-git shallow-check risk, gitcmd.go:75#29). **This
   is genuine regeneration latitude**, and the spec's instinct that "declared equivalent should not be a
   free pass for regeneration trust" is right *for this category only*.

DP-1 counts all three identically. The consequence is concrete and perverse: a lock with **more
defense-in-depth scores lower**. On the manifest example, deleting the four redundant input-layer
checks removes four survivors from both S and (C−V) while K is unchanged, so
`detector_power = K/(C−V)` *rises toward 1.0*. The metric therefore **rewards removing redundant safety
checks** — the opposite of what a trust gate should incentivize, and a direct contradiction of the
defense-in-depth that `allowed-survivors.txt` itself documents as virtuous ("Defence in depth, not a
gap"). A perfect, fully-pinning lock cannot reach 1.0 solely because it is redundant, which is not a
blind spot at all.

The sound design: the regeneration-trust bar must force each declared survivor to be **re-classified**
as "provably equivalent or redundantly-pinned" (does not count against power) versus "unproven /
judgment latitude" (counts, or must be driven to zero). Mechanically counting *all* declared survivors
conflates "irreducibly equivalent" with "lock blind spot" (category error); mechanically counting
*none* restores the trivial 1.0 and discards the category-3 signal that is the whole point of the gate.
The denominator must be **killable-or-unproven candidates**, not **all non-invalid candidates**.

## Findings

### F1 (HIGH) — §2 / §5 DP-1: survivor-in-denominator is a category error; perverse defense-in-depth incentive
- **Where:** §2 ("Survivors — including the declared-equivalent ones … stay in the denominator"),
  invariant DP-1 (§5), worked example §2 (`185/190 = 0.9737`).
- **Why:** see the Q1 position above. The justification "an equivalent survivor is a behavior the lock
  does not pin" is false for categories 1 and 2, and category 2 is 4 of the 5 survivors in the spec's
  own flagship example. The metric penalizes fully-pinned behavior and rewards deleting redundant
  checks.
- **Fix:** redefine the denominator as killable-or-unproven candidates. Require a per-survivor
  classification (`provably-equivalent | redundantly-pinned | unproven-latitude`) in the evidence
  record; only `unproven-latitude` survivors lower detector power (or must be zero for trust). State in
  §2 that irreducible/redundant equivalence is not a blind spot. Re-derive the manifest example under
  the corrected formula so the flagship datapoint is not itself a demonstration of the bug.

### F2 (HIGH) — §6.3 contradicts invariant DP-1: the core soundness question is frozen while called "open"
- **Where:** §6 open decision 3 ("does declaring equivalence still exempt them, or must a
  regeneration-trusted lock drive them to zero? (Leaning: they still count against power …)") vs. DP-1
  / §2, which already fix "survivors (declared or not) are in the denominator."
- **Why:** an invariant is a contract; it cannot be a contract while its determining question is still
  an open option. §6.3 is a load-bearing soundness decision masquerading as a deferred choice, and it
  is pre-frozen to the answer F1 shows is wrong. Whichever way §6.3 resolves changes DP-1's text.
- **Fix:** resolve §6.3 *in this spec* (it is not deferrable past acceptance), adopt the F1
  classification answer, and make DP-1 state the resolved rule. Remove the "leaning" contradiction.

### F3 (HIGH) — §4 / §7 / DP-2: no freshness binding; "trusted" can rest on a stale sweep
- **Where:** §4 evidence record (`swept_at_commit`), §7 ("no live re-sweep … reads committed …
  evidence"), DP-2 (undeclared survivor fails closed).
- **Why:** DP-2's fail-closed is only as live as the committed record. Nothing binds `swept_at_commit`
  to the current package source, so a package changed after its sweep (introducing a new undeclared
  survivor, exactly the gitcmd refresh story in `allowed-survivors.txt` lines 24–28) stays "trusted" on
  evidence from before the change. This is silent satisfy-and-reopen at the evidence layer, and the same
  staleness class Law 10 and the allowed-survivors line-drift notes already recognize.
- **Fix:** add an invariant: `trusted` is valid only if `swept_at_commit` equals the current package
  source hash (or the package tree is unchanged since that commit); otherwise the record is STALE and
  the package is not trusted (fail closed), forcing a re-sweep. The gate must detect a package edit
  since its evidence, not just read the number.

### F4 (MED) — §3 over-claims sufficiency; detector power is necessary-not-sufficient
- **Where:** §3 ("makes P regeneration-trusted (the follow-on regeneration loop may re-derive P's
  implementation and **trust the lock to catch a wrong candidate**)").
- **Why:** mutation operators are a finite basis (`tools/mutants`: string/int/binary/cond/if-cond/
  if-stmt …). A lock can kill every *generated* mutant and still miss behaviors no operator reaches
  (whole-function logic, ordering, concurrency, boundaries the operators don't synthesize). High
  detector power is a *lower bound on lock strength over the operator basis* — necessary, never
  sufficient. Vision line 156 concedes exactly this ("the whole ladder can be climbed while detector
  power goes to zero"). As written, §3 reopens satisfy-and-reopen one level up.
- **Fix:** state explicitly that detector power is NECESSARY-NOT-SUFFICIENT; that "regeneration-trusted"
  means regeneration is *permitted pending the loop's own output verification*, not that a
  lock-satisfying regenerated impl is correct. Move this from an implicit §7 non-goal to a stated
  limitation in §3.

### F5 (MED) — DP-5: calibration "presence" verifies a copied constant, not this sweep's calibration
- **Where:** DP-5 ("Calibration must be present in the sweep evidence … or the record is rejected").
- **Why:** `tools/mutants` prints `calibration neutral=survived invalid=invalid lethal=killed` as a
  **fixed string**, unconditionally, only after a passing scratch-tree `calibrate()` (main.go:67–71);
  a run that fails calibration exits before any summary. So in genuine stdout, calibration is implied by
  the summary existing. DP-5's added value is only against a **hand-fabricated** record — and requiring
  the *presence* of a boilerplate string guards against fabrication poorly (copy the constant). The
  C/K/V/S the metric consumes are hand-transcribed into JSON (§4), unbound to the sweep that produced
  them.
- **Fix:** commit the raw sweep stdout as the evidence and have `detector-power.sh` parse the
  `summary …` and `calibration …` lines *from that transcript*, re-deriving C/K/V/S rather than trusting
  hand-entered numbers. Then calibration and counts share one unforgeable source. (This mirrors the B1
  unit-3 fight over hand-copied digests vs. a single bound producer.)

### F6 (MED) — §7 non-goal is not clean: metric validity depends on unspecified loop details
- **Where:** §7 ("No regeneration loop") and §1 ("measures whether a lock is strong enough to trust that
  loop").
- **Why:** detector power is measured against mutation operators on a *human-written* impl. Its transfer
  to a *machine-regenerated* impl rests on two unstated assumptions: (a) the regenerator's failure
  modes resemble the operator distribution (generators produce plausible-but-wrong code, not random
  operator mutations — a different error cluster), and (b) the loop does not generate against the very
  tests detector power measures (teaching-to-the-test: tests fit by the generator no longer
  independently validate it). Both are loop details deferred to the follow-on spec, yet the gate's
  validity depends on them.
- **Fix:** add a §7 (or §6) note making these assumptions explicit as preconditions the regeneration-loop
  spec MUST satisfy (e.g. the loop must not train against the measured tests), and flag that the metric
  is a proxy whose predictive validity is unestablished until the loop yields correlatable evidence.

### F7 (LOW) — §2: formula undefined when C−V = 0
- **Where:** §2 formula `K/(C−V)`.
- **Why:** a package whose every candidate is invalid yields 0/0. Undefined.
- **Fix:** define C−V = 0 as rejected / never trusted (no semantic mutants means nothing was measured).

### F8 (LOW) — §3 / §6.1: single cross-package THRESHOLD and its unvalidatability (Q3)
- **Where:** §3 threshold gate, §6 open decision 1.
- **Why:** detector power is a ratio over a per-package operator sample whose difficulty is not uniform
  across heterogeneous packages, so one global threshold compares unlike quantities (the §6.1 "likely
  per-tier" instinct is right). Separately, the threshold cannot be empirically validated until the loop
  exists to correlate detector power against observed regeneration safety — so measuring lock strength
  first (Q3) is the correct ordering, but the first threshold is necessarily an *unvalidated proxy
  cut-point*, not a calibrated one.
- **Fix:** state that the first threshold is an explicitly provisional, conservative placeholder to be
  revisited once the loop produces regeneration attempts to correlate against; prefer per-tier.

## What is sound (survives review)

- §2's demolition of naive kill-rate (1.0 by construction for accepted packages) is correct and is the
  right reason to build a survivor-sensitive metric.
- **DP-2's direction is correct**: an *undeclared* survivor is a genuine untested behavior and belongs
  as a hard fail-closed error, not a low score. (F3 only concerns its liveness against stale evidence.)
- Excluding invalid mutants `V` from the denominator is correct — uncompilable mutants carry no
  semantic signal.
- DP-3 (declared count bound to `allowed-survivors.txt`) is a sound anti-drift invariant.
- The non-goals "no live re-sweep in make check" (sweeps are slow) and "detector power is additional,
  not a replacement for the zero-undeclared-survivors bar" are pragmatic and correct.
- Spec-before-code discipline (Law 6) is followed: no code, permanent INV ids correctly deferred to the
  kernel-SPEC append under an owner-approved change intent.

## Recommendation

**NEEDS_CHANGES — round 2 required.** The blocking items are F1 (redefine the denominator via
survivor classification; re-derive the flagship example), F2 (resolve §6.3 in-spec and align DP-1), and
F3 (freshness/staleness invariant binding `swept_at_commit` to current source). F4–F6 (necessary-not-
sufficient framing; calibration/counts bound to committed sweep stdout; explicit loop-precondition
assumptions) should land in the same revision. F7–F8 are low-cost tightenings. Once revised, a fresh
non-author verdict adjudicates; only then do §6 owner decisions (threshold value, shell-vs-package) and
an owner-approved change intent authorize implementation.

(Note: I am a non-author reviewer. This NEEDS_CHANGES verdict is a rejection with findings; per Law 9
the author may revise and a non-author adjudicates the revision. I made no edit to the spec.)


## Findings — original pre-schema block (R1-4 provenance)

> The structured `findings:` in the frontmatter were normalized to `vera.verdict.v1`
> (`finding_id` + `severity` only). The **original** block, including each finding's full
> `summary`/`verdict` text and its original severity word, is preserved verbatim below so no
> reviewer detail is lost. `info`→`LOW` and `low-medium`→`MED` were the only severity remaps.
> See `docs/deferrals.tsv` → `DEF-DELIVERY-ENFORCE-VERDICTS`.

```yaml
findings:
  - finding_id: p7-2b-detector-power-spec-review-round1-F1
    severity: HIGH
    summary: §2/§5 DP-1 survivor-in-denominator is a category error as stated — it conflates irreducibly-equivalent and redundantly-pinned survivors (neither is exploitable latitude) with genuine blind spots; the flagship manifest example (4 of 5 survivors are redundant-with-Bind→Validate, i.e. doubly-pinned) is penalized though fully pinned, and the metric rewards deleting defense-in-depth
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F2
    severity: HIGH
    summary: §6.3 "open decision" contradicts invariant DP-1 — the core soundness question (does declared-equivalence still count against power) is presented as deferrable but is already frozen by DP-1/§2, and frozen to the answer F1 shows is wrong; an invariant cannot be a contract while its basis is an open option
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F3
    severity: HIGH
    summary: §4/§7/DP-2 bind no freshness — "trusted" reads committed evidence with swept_at_commit but nothing binds that commit to the current package source, so a package can stay trusted on a sweep taken before a trust-relevant change; the undeclared-survivor fail-closed (DP-2) is only as live as a possibly-stale record (same staleness class as Law 10 / allowed-survivors line drift)
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F4
    severity: MED
    summary: §3 over-claims sufficiency — "trust the lock to catch a wrong candidate" treats high detector power as sufficient; over a finite mutation-operator basis it is necessary-not-sufficient (vision line 156 concedes the ladder can be climbed while detector power → 0). Satisfy-and-reopen reappears through operator blind spots the metric cannot see
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F5
    severity: MED
    summary: DP-5 calibration check is weak — the sweep's calibration line is a fixed boilerplate string printed unconditionally after a passing scratch-tree calibrate(); requiring its "presence" in a hand-authored JSON record verifies only that a constant was copied, not that THIS sweep's numbers came from a calibrated run. Bind the record to committed raw sweep stdout and re-derive C/K/V/S from it
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F6
    severity: MED
    summary: §7 "no regeneration loop" non-goal is not clean — the metric's validity as a regeneration-trust gate silently assumes the mutation-operator failure distribution represents the regenerator's failure modes, and that the loop does not generate against the very tests measured (overfitting); both are unspecified loop details the gate depends on
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F7
    severity: LOW
    summary: §2 formula undefined when C−V = 0 (all candidates invalid) — division by zero unspecified; define as rejected / never trusted
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round1-F8
    severity: LOW
    summary: §3/§6.1 a single cross-package THRESHOLD assumes operator-sampling difficulty is comparable across heterogeneous packages (it is not), and the threshold cannot be empirically validated until the loop exists to correlate detector power against observed regeneration safety; the first value should be an explicitly provisional, conservative placeholder pending loop evidence
    verdict: CONFIRMED
```