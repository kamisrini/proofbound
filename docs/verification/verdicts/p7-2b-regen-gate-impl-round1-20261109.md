---
schema: vera.verdict.v1
verdict_id: p7-2b-regen-gate-impl-round1-20261109
status: ACCEPTABLE
reviewed_commit: 635afd3cf2771e5fff30a804b9fe4db797a67a8c
findings:
  - finding_id: F1
    severity: MED
  - finding_id: F2
    severity: MED
  - finding_id: F3
    severity: LOW
  - finding_id: F4
    severity: LOW
artifact_path: docs/verification/verdicts/p7-2b-regen-gate-impl-round1-20261109.md
artifact_sha: 9854068d5789f40f5ffab392820561825d280727f8138ec39cc84c06697ce98c
---

# P7 #2b — regeneration-as-repair GATE (gate-only slice) — non-author package verdict (round 1)

**ACCEPTABLE** (non-blocking findings). An independent non-author reviewer (fresh-context subagent;
authored neither `scripts/regen-gate.sh` nor its self-test) verified commit `635afd3` against the
ACCEPTED spec `docs/plans/P7-2b-regeneration-loop-spec-draft1.md` §4, the change intent
`CI-regen-loop-gate-1a2b3c`, and the Build Laws. The gate-only slice of O-2 is **non-author
accepted**; the owner remains the final acceptor. Findings transfer zero trust to any fix
(`VD-verification-asymmetry-2dyjnd`).

## Basis
Read the full diff and both scripts; read the spec §4, the change intent, and the reused
`scripts/detector-power.sh` (line 281) output format; ran `scripts/tests/regen-gate.test.sh` (ok); ran
`make check` alone → exit 0, "0 issues", with `regen-gate.test: ok` firing in `make hooks-test`.
Confirmed every RL control fails closed, the NEUTRAL fixture AUTHORIZEs, each POSITIVE control is
DENIED with its intended RL tag, and the gate's `U=0`/`trusted=true` greps genuinely match the real
`detector-power.sh` OK line (printed only after DP-4's trusted-consistency check). Within scope (a
`scripts/` shell gate + self-test); no exclusion violated (no regenerator/step-3, no step-4, no
non-zero `U_BUDGET`). The declared design seam (O-1 detector-power is per-package, the lock is
per-revision; bound via the admission/dp-record naming the `L@r+1` digest, checked structurally and
carried as an RL-8 limitation) is honest and adequately bounded for a gate-only slice; it does not
overclaim RL-2's trust.

## Findings (harness-completeness; non-blocking; addressed in `a785420`)
- **F1 (MED) — RL-11's D-b-digest leg was not proven load-bearing.** The gate checks the `L@r+1`
  digest against admission, the detector-power record, AND the D-b attestation
  (`scripts/regen-gate.sh:62-63`), but only the admission and detector-power legs had positive
  controls; deleting the D-b-digest check left the self-test green — the "control that cannot fail"
  class the acceptance boundary targets. *Fixed:* a reject control mutating the D-b
  `lock_revision_digest` (tagged RL-11), proven load-bearing.
- **F2 (MED; reviewer LOW-MED) — RL-2's "not U=0" control mutated `U` and `trusted` together**, so
  neither grep was isolated. *Fixed:* split into `U=1 trusted=true` and `U=0 trusted=false`.
- **F3 (LOW) — several RL checks have multiple legs with only one controlled** (RL-9 `ce_id`
  emptiness, RL-5 empty-isolation/empty-checker, RL-4 non-JSON grep fallback). Minor completeness.
- **F4 (LOW) — `db`/`dp_record` were not pre-validated as JSON**, so a malformed input exited via
  `set -e` with a raw `jq` error instead of a tagged `DENY:`. Still fail-closed, but inconsistent with
  the `die_deny` discipline. *Fixed:* `dp_record` and `db` now pre-validated as JSON with a tagged
  DENY.

## Outstanding CI-scope item (not a code defect)
The CI Allowed scope also lists a `docs/gates.md` Law-4 advisory row + the p6-census C2 coupling; that
registration is a follow-up (deferred to avoid a derived-staleness error). Until it lands, the gate's
RL-8 realization is partial (it checks a `declared_limitations` field but there is no standing
`gates.md` advisory yet).

## Reviewer identity
Independent non-author reviewer (fresh-context subagent dispatched by the owning session, 2026-11-09).
Authored neither the gate nor its self-test.
