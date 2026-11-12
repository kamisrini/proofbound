---
schema: vera.verdict.v1
verdict_id: p7-2b-detector-power-spec-review-round3-20261105
status: ACCEPTABLE
reviewed_commit: 73108b90e6e2769ad2ac5c33dcd7e294a5e7ccca
findings:
  - finding_id: p7-2b-detector-power-spec-review-round2-N1
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round2-N2
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round2-N3
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round2-N4
    severity: LOW
  - finding_id: p7-2b-detector-power-spec-review-round3-CL1
    severity: MED
  - finding_id: p7-2b-detector-power-spec-review-round3-CL2
    severity: LOW
  - finding_id: p7-2b-detector-power-spec-review-round3-CL3
    severity: LOW
  - finding_id: p7-2b-detector-power-spec-review-round3-CL4
    severity: LOW
artifact_path: docs/verification/verdicts/p7-2b-detector-power-spec-review-round3-20261105.md
artifact_sha: 2486d29d63d01d90b9b25fba1c42708a9e8ad5310982d69aefce366b0ef07dfe
---

# P7 #2b detector-power metric + regeneration-trust gate (spec draft 3) — non-author re-review, round 3 — ACCEPTABLE

## Provenance

- **Under review:** `docs/plans/P7-2b-detector-power-spec-draft1.md`, now **DRAFT 3** — the revision
  of the draft that round 2 returned NEEDS_CHANGES ("narrow, near-acceptance"). Design spec only
  (Law 6 spec-before-code): no `kernel/*`, `scripts/*`, or `gates/*` code exists or is under review.
- **Reviewed commit:** `73108b90e6e2769ad2ac5c33dcd7e294a5e7ccca` (`73108b9`). Round 1 reviewed
  `d110465`, round 2 `986d6d7`. HEAD unchanged at dispatch and at this write; no tree modification
  during review.
- **Driving mechanism:** `/proofbound-review`, round 3 — adversarial re-verify of round-2 N1–N4 plus
  a hunt for defects draft 3's additions (the absolute-U gate, `lock_inputs_sha256` scope, DP-8
  execution-binding modes, the survivor bijection) introduced. DESIGN-SPEC review: the thing to judge
  is the SOUNDNESS OF A METRIC.
- **Reviewer identity:** independent non-author reviewer dispatched via the Agent tool; same reviewer
  across rounds 1–3. I did NOT author any draft or revision (Build Law 9). This is an **acceptance of
  the design spec** by a non-author verdict (Law 9 satisfied). It accepts the spec as sound and
  implementation-ready; it does **not** accept any gate code, record, sweep, or `trusted` flag — each
  remains behind the governed implementation, `make mutants` + a non-author package verdict, and the
  owner-approved change intent (§8).
- **Cleanroom redaction:** none required; no absolute paths or usernames below.

## Overall verdict: ACCEPTABLE

Draft 3 resolves all four round-2 findings genuinely, not cosmetically. The metric is now **sound as a
design**: the trust gate keys on the absolute count of unproven-latitude survivors (`U`) against a
per-tier budget, the gameable ratio is demoted to a reported diagnostic, the freshness digest covers
the lock itself (tests + register + SPEC rows), and the raw sweep is bound to a real execution
(reproducible, or a named/expiring/fail-closed attestation). Combined with the round-1 E/R/U core and
round-2 necessary-not-sufficient framing, there is no path by which the gate passes something it
should not **without either violating Law 9 (self-attestation — forbidden repo-wide) or a semantic
mis-classification that survives non-author review** — and the latter is the correct, auditable,
only-possible trust boundary for a metric over a semantic judgment, the same discipline that has
survived nine adversarial rounds in `allowed-survivors.txt`.

Four clarifications (CL-1 MED, CL-2/3/4 LOW) are recorded below. **None is blocking**: CL-1 tightens
wording on a mechanism that is already correct and already governed by Law 9; the rest are
implementation/owner-decision guidance. They are applied as clarifying edits in the house pattern
(the round-6 B1 precedent: cosmetic/wording remedies applied in a follow-up without a new round), and
do **not** require a round 4.

## Round-2 findings — per-finding re-verify

- **N1 (MED) — RESOLVED.** §3 makes `P` regeneration-trusted only when `U(P) ≤ U_BUDGET(tier)`, top
  tier `U=0`, each class-U survivor individually justified; §2 demotes `detector_power = K/(K+U)` to a
  reported diagnostic that "does **not** drive the gate." This closes the three round-2 holes:
  *size-dependence* (U is absolute, not scaled by K), *padding* (adding well-tested trivial code
  raises K but not U, and U is what's gated), and the *operator-unreachable blind spot* — which no
  mutation-based metric can close, and which §2's "Honest limit" paragraph now states explicitly
  ("`U=0` means *no mutation-detected unproven latitude* … necessary, not sufficient"). The
  withdrawn "a perfect lock can reach 1.0" is gone. **Does keying on U merely move the soft spot to
  the classification?** Yes — with the ratio gone, the only way to keep `U` low is the E/R/U labels,
  so the classification is now the sole lever. That is acceptable and correctly located: a mutation
  metric cannot eliminate the semantic judgment of whether a survivor is exploitable, only make it a
  falsifiable, non-author-reviewed, per-site claim — which draft 3 does, and mechanizes hardest on R
  (the dangerous class) via the named pinning line. Residual sharpenings are CL-3 (R subsumption) and
  CL-4 (per-U safety argument); neither undermines soundness.
- **N2 (MED) — RESOLVED.** DP-6 / §4 define `lock_inputs_sha256` to cover "non-test source + `_test.go`
  + the package's `allowed-survivors.txt` entries + its SPEC invariant/lock rows." This is exactly the
  set whose change can alter `K`, `U`, or a class: a test-only weakening (the most direct lock
  weakening) and a register relabel both now flip the digest and trigger stale-reject. The only input
  left outside the digest is the measurement tool itself (CL-2, LOW).
- **N3 (MED) — RESOLVED.** DP-8 binds the raw sweep to a real execution in two modes. *reproducible*
  (a refresh re-runs `scripts/mutants.sh` at the pinned digest and must reproduce `raw_sweep_sha256`)
  is a full closure of the fabricated-`sweep.txt` gap. *attested* — for slow embedded-postgres
  packages where an in-gate re-run is infeasible — is **not** a relabeled silent hand-attestation: it
  is (a) **named** (a specific non-author re-runner recorded, accountable, Law-9-bound), (b) **bounded**
  by a Law-4 advisory **expiry** (it cannot persist; it re-attests or graduates to reproducible), (c)
  **disclosed** as an explicit weaker limitation, and (d) **fail-closed** if neither mode holds. That
  is the honest ceiling for a non-reproducible evidence artifact and it invokes the correct precedent
  (the B1 unit-3 hand-attestation lesson, resolved by a named, source-reviewed producer + non-author
  inspection). The one residual is wording (CL-1): "non-author re-runner" must be read as non-author
  *of the sweep/record* performing an independent re-run — which Law 9 already compels repo-wide.
- **N4 (LOW) — RESOLVED.** §2 requires each R to "name that specific test/line whose mutant dies" and
  each E its equivalence argument; DP-3 states the "exact bijection sweep-survivors ↔
  record-classified survivors (extra record entry also fails closed)," and DP-2 fails closed on any
  unclassified sweep survivor. Together these pin every survivor to exactly one justified class bound
  to `allowed-survivors.txt`. Residual CL-3 (naming a pin proves it is tested, not that it subsumes
  the mutated behavior) is a LOW tightening.

## Clarifications to fold in (non-blocking; apply as clarifying edits, no round 4)

### CL-1 (MED, non-blocking) — DP-8 attested mode: re-runner independence
- **Where:** §4 record `execution_binding`, DP-8.
- **Why:** attested mode is the mode the packages #2b most cares about (embedded-postgres stores) will
  actually use. "Named non-author re-runner" is correct but terse; if "non-author" is read as merely
  non-author-of-the-package-code, the sweep's own author could attest their own sweep and the
  execution binding collapses to self-attestation, reopening N3 precisely where it bites.
- **Fix:** state that the attester is an actor who did **not** produce the sweep/record and who
  **independently re-ran** it (not a rubber-stamp), per Law 9. This restates an inherited law, so it
  is a wording tightening, not a new mechanism — hence non-blocking.

### CL-2 (LOW, non-blocking) — pin the measurement tool/operator basis
- **Where:** §4 record, DP-6.
- **Why:** the sweep result depends on the `tools/mutants` operator set (and the Go toolchain), neither
  of which is in `lock_inputs_sha256`. A change to the operator basis alters candidates/kills/U; in
  *reproducible* mode a refresh catches it, but in *attested* mode a stale record survives an
  operator-set change silently.
- **Fix:** pin the `tools/mutants` version/digest (and note the toolchain) in the record so an
  operator-basis change invalidates stale attested evidence.

### CL-3 (LOW, non-blocking) — R requires a subsumption argument, not just a tested pin
- **Where:** §2 R definition, DP-3.
- **Why:** an R justification that names a line whose mutant dies proves that line is *tested*; it does
  not prove that line *subsumes* the mutated behavior. A genuine U can be mislabeled R by citing an
  independently-tested but non-subsuming line, and the mechanical "its mutant dies" check still passes.
- **Fix:** require each R justification (and its DP-3 proving test) to argue that the named pin
  constrains the same behavior the mutated site would, so removing the site changes nothing the pin
  does not already catch.

### CL-4 (LOW, non-blocking) — per-U safety argument for a non-zero budget
- **Where:** §3, §6.1 (owner decision).
- **Why:** a non-zero `U_BUDGET` authorizes regeneration against a lock with a known, genuine un-pinned
  degree of freedom — a known satisfy-and-reopen surface. "Individually justified" should mean more
  than "we counted it."
- **Fix:** require each budgeted class-U survivor's justification to argue why that specific un-pinned
  behavior is safe for a regenerated implementation to vary (i.e. not correctness- or
  security-load-bearing). Owner guidance for §6.1.

## What is sound (the accepted design)

- **Round-1 core (E/R/U + survivor classification):** confirmed resolved round 2, unchanged and sound.
- **Round-2 absolute-U gate:** `U ≤ U_BUDGET(tier)`, top tier `U=0`, ratio demoted to diagnostic —
  closes size-dependence and padding; the operator-blind-spot limit is honestly disclosed.
- **Freshness (DP-6):** the digest covers the lock inputs that actually determine the measurement.
- **Execution binding (DP-8):** reproducible is a hard closure; attested is a named, expiring,
  fail-closed, disclosed advisory — the correct honest treatment and the correct precedent.
- **Bijection + R/E justification (DP-2/DP-3), necessary-not-sufficient (§3), loop dependencies (§7):**
  all sound. Spec-before-code discipline holds; permanent INV ids correctly deferred to the
  kernel-SPEC + `invariants.lock` append under the change intent.

## What remains (owner + implementation)

This acceptance is of the **spec**. Before anything is `trusted` in practice:

1. **Owner decisions (§6), still open by design:**
   - **`U_BUDGET` per tier** — first values explicitly provisional and conservative (§3); and per CL-4,
     a required per-U safety argument for any non-zero budget.
   - **Shell-gate vs a detector-power Go package** (§4).
2. **Owner-approved change intent** for the governed `scripts/*` / `gates/*` (and `kernel/*` if a
   package) work (§8.2).
3. **Implementation, spec-first (§8.3):** kernel SPEC.md with permanent INV ids + `invariants.lock`
   append under the intent (or the shell gate + self-test); add the E/R/U class + `pins` fields to the
   survivor register; run the deferred full-package `internal/twin` sweep and classify its survivors to
   seed the first execution-bound records; register the advisory `docs/gates.md` row (Law-4 owner +
   ISO expiry). Fold in CL-1..CL-4 here.
4. **Package acceptance (§8.4):** calibrated `make mutants` of any new code + a non-author ACCEPTABLE
   package verdict (ROADMAP P1 bar, Law 9).
5. **Only then (§8.5)** does the follow-on regeneration-loop spec begin, consuming budget-compliant
   packages and discharging the §7 loop dependencies (operator-distribution representativeness;
   no training/overfit against the measured tests).

## Recommendation

**ACCEPTABLE for implementation.** Round-1 F1–F8 and round-2 N1–N4 are all resolved; the detector-power
metric and regeneration-trust gate are sound as a design and implementation-ready. CL-1 (MED) and
CL-2..CL-4 (LOW) are non-blocking clarifications to apply as clarifying edits in the same pass that
opens implementation; they do not require a round 4. The owner can settle §6 (`U_BUDGET` per tier,
shell-vs-package) and flip an owner-approved change intent to authorize the governed work.

(Note: I am a non-author reviewer; per Law 9 this non-author verdict is the acceptance capacity for a
spec this session did not author. I made no edit to the spec.)


## Findings — original pre-schema block (R1-4 provenance)

> The structured `findings:` in the frontmatter were normalized to `vera.verdict.v1`
> (`finding_id` + `severity` only). The **original** block, including each finding's full
> `summary`/`verdict` text and its original severity word, is preserved verbatim below so no
> reviewer detail is lost. `info`→`LOW` and `low-medium`→`MED` were the only severity remaps.
> See `docs/deferrals.tsv` → `DEF-DELIVERY-ENFORCE-VERDICTS`.

```yaml
findings:
  - finding_id: p7-2b-detector-power-spec-review-round2-N1
    severity: MED
    summary: Ratio holes (size-dependence, padding, operator-blind-spot) — RESOLVED. The trust gate now keys on the absolute class-U count against a per-tier U_BUDGET (top tier U=0); the ratio is demoted to a reported diagnostic; the operator-unreachable gap is correctly disclosed as necessary-not-sufficient rather than hidden
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round2-N2
    severity: MED
    summary: Freshness digest scope — RESOLVED. DP-6 lock_inputs_sha256 covers non-test source + _test.go + the package register entries + the SPEC invariant rows; a test-only weakening or a class relabel now triggers stale-reject
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round2-N3
    severity: MED
    summary: Raw-sweep execution binding — RESOLVED. DP-8 reproducible mode (re-run matches raw_sweep_sha256) is a full closure; attested mode is an honest named + Law-4-expiring + fail-closed advisory, not a silent reopening. One non-blocking clarification (CL-1) on attested re-runner independence
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round2-N4
    severity: LOW
    summary: R/E justification + bijection — RESOLVED. §2/DP-3 require each R to name its pinning test/line and each E its equivalence argument; DP-2+DP-3 enforce an exact sweep↔record survivor bijection (unclassified survivor and extra record entry both fail closed)
    verdict: RESOLVED
  - finding_id: p7-2b-detector-power-spec-review-round3-CL1
    severity: MED
    summary: NON-BLOCKING clarification — DP-8 attested mode says "named non-author re-runner"; state that this means a non-author OF THE SWEEP/RECORD performing an independent re-run (not a sign-off), else attested collapses toward self-attestation for exactly the slow packages that use it. Law 9 already forbids self-grading repo-wide, so this is a wording tightening, not a new mechanism
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round3-CL2
    severity: LOW
    summary: NON-BLOCKING — lock_inputs_sha256 covers the lock but not the measurement tool; the sweep result also depends on the tools/mutants operator set (and Go toolchain). Pin the tools/mutants version/digest in the record so an operator-basis change invalidates stale evidence in attested mode (reproducible mode catches it on re-run)
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round3-CL3
    severity: LOW
    summary: NON-BLOCKING — DP-3's "R names a pinning test/line whose mutant dies" is necessary but not sufficient; the reviewer must also confirm the named pin SUBSUMES the mutated behavior, not merely that it is independently tested. Tighten the DP-3 proving-test shape to require the subsumption argument
    verdict: CONFIRMED
  - finding_id: p7-2b-detector-power-spec-review-round3-CL4
    severity: LOW
    summary: NON-BLOCKING — a non-zero U_BUDGET authorizes regeneration against a lock with known genuine un-pinned latitude (a known satisfy-and-reopen surface); require each budgeted class-U justification to argue why that specific behavior is safe for a regenerated impl to vary (not correctness/security-load-bearing). Guidance for the §6.1 owner decision
    verdict: CONFIRMED
```