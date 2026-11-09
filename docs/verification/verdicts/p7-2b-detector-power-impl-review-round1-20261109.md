---
schema: vera.verdict.v1
verdict_id: p7-2b-detector-power-impl-review-round1-20261109
status: ACCEPTABLE
reviewed_commit: a8220886641a3a5c8d6e02303840932785dc23a2
findings:
  - finding_id: F1
    severity: LOW
artifact_path: docs/verification/verdicts/p7-2b-detector-power-impl-review-round1-20261109.md
artifact_sha: 79e0491e86cea201d3945dad7b842e9513e9459278620dc8bd9a72d249a0716a
---

# P7 #2b — detector-power / regeneration-trust gate (O-1) — implementation non-author verdict

**ACCEPTABLE** (with one LOW-severity finding, F1, recorded below). The gate mechanism, its
self-test, and the first execution-bound seed's `U=0 / trusted` conclusion are **accepted**. F1 is an
author correction to the seed record's justification *prose* — it does not change any survivor
classification or the trust outcome, and is kept separate below for the author to act on, with a
non-author to re-confirm the fix (`VD-verification-asymmetry-2dyjnd`: a finding transfers zero trust
to its remedy). (Status is `ACCEPTABLE` per the vera.verdict.v1 vocabulary — {ACCEPTABLE, NEEDS_WORK};
the earlier `ACCEPTABLE-WITH-NOTES` label was non-conformant and is corrected here. F1 was
subsequently fixed in `25d13ad`; its non-author re-confirmation remains outstanding.)

## Provenance

- **Reviewed HEAD:** `a8220886641a3a5c8d6e02303840932785dc23a2` (`main`). The detector-power work
  under review landed at `3f446eb` (`feat(p7-2b): detector-power / regeneration-trust gate (O-1) + first real seed`);
  `a822088` on top is a notes/EOD commit only. The tree at review time was clean apart from the two
  permanent untracked paths (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`),
  neither of which this review touches.
- **Tier (`VD-review-ceremony-tiers-v2b6kp`):** connector/peripheral — a `scripts/` shell gate plus
  its `scripts/tests/` self-test. No concurrency, filesystem-identity, crash-safety, or ledger
  write-path surface, so no escalation to the foundational tier was triggered. One adversarial review.
- **Reviewer identity:** independent non-author reviewer (`claude-opus-nonauthor-reviewer-20261109`).
  The reviewed change was authored in a prior session; this session did not write the gate, the
  self-test, or the seed record. This is therefore a Law-9 **acceptance**, not a self pre-check.
- **Basis:** the ACCEPTED design spec `docs/plans/P7-2b-detector-power-spec-draft1.md` (DP-1..DP-8,
  §5), the Build Laws in `CLAUDE.md`, and the committed artifacts: `scripts/detector-power.sh`,
  `scripts/tests/detector-power.test.sh`, `docs/verification/detector-power/internal-twin-manifest.{json,sweep.txt}`,
  the `docs/allowed-survivors.txt` `{dp:class}` entries, the `docs/gates.md` advisory row, and the
  p6-census coupling (`docs/plans/p6-census-rows.tsv` C2-042).
- **Redaction:** this verdict was authored fresh, not transcribed from raw tool output; no forbidden
  strings (absolute paths, usernames) are included. Redaction count: 0.
- **Driving evidence run:** reviewer-executed reproducible refresh, `scripts/detector-power.sh
  internal/twin/manifest --refresh` (background run id `br33o80st`, exit 0).

## What was verified (accepted)

**The gate realizes DP-1..DP-8 faithfully and fails closed.** Read line by line, `detector-power.sh`
re-derives C/K/V/S from the committed sweep stdout and requires the summary to agree (DP-5);
pins the sweep bytes by digest and verifies the execution binding (DP-8); requires the calibration
line (DP-7); enforces an exact sweep↔record survivor bijection with duplicate-id rejection and binds
every classified survivor to a `docs/allowed-survivors.txt` entry whose `{dp:class}` must equal the
record's class, a Law-2 drift guard (DP-2/DP-3); keys trust on the absolute class-U count with the
`K+U=0` rejection and the diagnostic held to computed `K/(K+U)` (DP-1); gates `trusted` on the
per-tier budget with an unknown tier failing closed (DP-4); and recomputes `lock_inputs_sha256` over
the current source + `_test.go` + register entries + SPEC + `tools/mutants` operator set, rejecting a
stale record (DP-6).

**The self-test is load-bearing and fires.** `detector-power.test.sh` is hermetic (synthetic
fixtures in a temp dir, no repo package, no real sweep). Each DP control pairs a NEUTRAL fixture that
must be accepted with a POSITIVE control that must be rejected **and** whose failure message must
carry the intended `DP-N:` tag (`reject()` greps for the tag) — so a control cannot pass by tripping
a different check, the exact "detector-inert / scope-blind" failure shape this project legislates
against. Both the self-test and the gate-over-seed pass at the reviewed HEAD:
`detector-power.test: ok` and `OK pkg=internal/twin/manifest K=330 U=0 detector_power=330/330
tier=top budget=0 trusted=true mode=reproducible`.

**The committed sweep is a real execution, reviewer-reproduced.** I independently re-ran the full
mutation sweep via `--refresh`; it reproduced the pinned `raw_sweep_sha256`
(`4968444f…013f69`) exactly (`refresh OK`). This closes the fabricated-`sweep.txt` gap (the N3 /
B1-unit-3 hand-attestation lesson) for this seed by live reproduction — stronger than the design's
out-of-band requirement, which only pins bytes by digest in `make check`.

**`U = 0` is a correct conclusion.** All five survivors are genuine redundancy, not unproven
latitude. I traced each to a tested downstream check in source: survivors #1/#2 (declaration-layer
`validateRevision`) are subsumed by the Bind→Validate calls at `manifest.go:209`/`:212`, whose
mutants are killed in the reproduced sweep; #3 (empty module/version) by `validateModuleValues`
(`:305` killed); #4 (`validateToolchainValue(input.Toolchain)`) by the **same function** called in
the Validate path at `manifest.go:218` (killed); #5 (the `object == nil` nil-guard) by the
required-field loop, which rejects a JSON `null` object (nil map → missing-field error) regardless of
the guard. None is a class-U degree of freedom, so the top-tier `U=0` budget and `trusted=true` hold.

**Build-Law walk (gate + seed).** Law 1: the `@generated` audit artifact regenerated by this commit
carries its marker; the seed record is evidence, not a generated artifact. Law 2: the register↔record
class agreement is enforced, not copied; `pins` live once in the record. Law 3: the self-test proves
each control fires and is in `make hooks-test`. Law 4: the `docs/gates.md` row is ADVISORY with owner
+ `2026-12-31` expiry and a graduate-or-delete condition. Laws 6/7/8/9: spec-before-code honored
(round-3 ACCEPTED spec); DP-8 compiles the fabricated-sweep lesson into an enforced binding; shell +
`jq`/`sha256sum`, no new dependency; this verdict supplies the required non-author lens. No violation
found.

## Finding (for the author; separate from the acceptance above)

**F1 — low-medium — seed record overstates its mutation support (open).**
The gate cannot verify subsumption: for a class-R survivor it checks only that `.pins` is non-empty
and that the justification contains the substring "subsum". The justification *prose* is therefore
the sole evidence a future non-author relies on — and in the canonical first seed it is partly
inaccurate. Three survivors cite subsuming lines that carry **zero mutants** in the committed
(and reviewer-reproduced) sweep, while asserting those lines' mutants are "KILLED in this sweep":

- survivor #3 (`declaration.go:216#41`) record cites `manifest.go:299` as a killed subsuming operand
  — `:299` has no mutant; the real killed evidence is `:305`.
- survivor #4 (`declaration.go:234#37`) register entry cites `manifest.go:273` as the redundant line
  — `:273` has no mutant; the real killed evidence is `:218`.
- survivor #5 (`manifest.go:382#59`) record states "the required-field loop's mutants are KILLED in
  this sweep" — lines `:388`/`:389` carry no mutant at all; the subsumption is a sound pure-reasoning
  downstream argument (which the register states correctly) with no mutation backing to cite.

*Harm:* the exemplar every future detector-power record will imitate asserts mutation evidence that
does not exist on the named lines, in the "declaration checked, truth unchecked" shape this repo has
repeatedly had to gate against (cf. `skip-lint`/`figure-provenance` known limits). It does not break
the trust decision — the classifications are correct via the adjacent killed line (`:305`, `:218`) or
pure downstream subsumption (#5), so `U=0` is unaffected — but it erodes the one thing the gate
leaves to human review.

*This is a statement of harm, not an instruction.* The author owns the remedy (e.g. repoint each
citation to the line whose mutant is actually killed, and restate #5 as pure downstream subsumption
as the register already does); a non-author then re-confirms the corrected prose. Trust does not
transfer from this finding to any fix.

## Disposition

- **Accepted:** the detector-power gate (`scripts/detector-power.sh`) + self-test
  (`scripts/tests/detector-power.test.sh`), the DP-1..DP-8 mechanism, and the
  `internal/twin/manifest` seed's `U=0 / trusted=true / tier=top` trust conclusion — realizing
  `BR-regeneration-trust-5f8a03` O-1 under `CI-p7-detector-power-1a2b3c`.
- **Open author action:** F1, a justification-prose correction to the seed record (and matching
  register entry), to be re-confirmed by a non-author. It gates whether the seed is cited as the
  taxonomy exemplar; it does not gate the mechanism's acceptance.
- Whether `DEF-P7-REGENLOOP-001` may proceed on this acceptance (gate accepted; one open prose note)
  is an owner decision.
</content>
</invoke>
