---
schema: vera.verdict.v1
verdict_id: p7-2b-regeneration-loop-spec-review-round2-20261109
status: ACCEPTABLE
reviewed_commit: 81e1d2128870f1161762d370c1e45eef87da737e
findings:
  - finding_id: N1
    severity: MED
  - finding_id: N2
    severity: MED
  - finding_id: N3
    severity: LOW
artifact_path: docs/verification/verdicts/p7-2b-regeneration-loop-spec-review-round2-20261109.md
artifact_sha: a1a82e41a06fb212b794896f93152b47054be0178f5fc36e69fd68d397f7943d
---

# P7 #2b — regeneration-loop spec (BR O-2) draft 2 — non-author re-review (round 2)

**ACCEPTABLE** (with three non-blocking clarifications N1–N3 to fold before the implementation change
intent). An independent non-author reviewer (the round-1 reviewer, re-engaged; did not author the
spec) re-reviewed `docs/plans/P7-2b-regeneration-loop-spec-draft1.md` at `81e1d21` and confirmed **all
8 round-1 findings substantively resolved** — F1 (the HIGH) genuinely fixed by the `U=0` scoping, not
merely deferred. The remaining items are MED-and-below refinements; the design-only posture holds.
This mirrors O-1's round-3 ACCEPTABLE with applied CL-1..CL-4. Findings transfer zero trust to any fix
(`VD-verification-asymmetry-2dyjnd`).

## Round-1 findings — all resolved

F1 (HIGH, `U=0` scoping removes the hazard, not defers it — airtight modulo the operator-unreachable
residual honestly carried in RL-8; aligns with O-1 §6 decision 1); F2 (step-2 produce-vs-verify +
per-cycle cost named); F3 (RL-9 lethality, coherent with RL-3); F4 (RL-5 non-author D-b attestation +
"lock IS the tests" seam as a declared Law-4 limit); F5 (RL-10 role separation — see N1); F6 (§3.1
single home — see N3); F7 (RL-11 digest identity, clean); F8 (genesis bounded-advisory default
stated).

## Clarifications to fold before the implementation change intent

- **N1 (MED) — "actor" identity undefined; the three Law-9 roles (RL-10, and RL-5/RL-7) can silently
  collapse in a single-operator/agent context.** RL-10 says the `L@r+1` author, regenerator, and
  verdict reviewer are "three distinct actors; any coincidence fails closed," but "actor" is undefined
  (human / agent session / signing identity), so three roles can be one human in three hats, which
  "fails closed" cannot detect without a defined identity basis. O-1 CL-1 set a usable precedent (a
  *named* non-author re-runner recorded in verdict files); the spec should cite it, make the
  actor-identity basis an owner decision, and — until defined — carry the collapse risk as a declared
  Law-4 limitation, the same honest treatment given the D-b seam (RL-8).
- **N2 (MED) — RL-12 sits in the active proving-test table while gating out-of-scope functionality:
  risk of a vacuously-passing self-test (the foundational Law-9 / "test that cannot fail" lesson).**
  RL-12 guards non-zero budgets, which §6 makes a non-goal; listing `TestClassUFalsificationRevokesSiblings`
  beside the load-bearing RL-1..RL-11 invites compiling that control in the `U=0` slice where no
  budgeted-U surface exists for it to act on — an un-failable control (Law 3: controls must prove they
  can fail). Separate RL-12 from the active contract and state explicitly that no self-test is
  compiled for it in this slice.
- **N3 (LOW) — §3.1 keys the D-a representativeness record per `<lock-revision-digest>`, but
  representativeness is a per-(subject, regenerator) deployment property, not per-revision — storing
  it per revision duplicates one datum (Law 2).** D-b's isolation attestation is legitimately
  per-revision/candidate; D-a's representativeness likely warrants a per-deployment key.

## Confirmed sound (not findings)
`U=0` scoping consistent with O-1 §6 decision 1 (not a conflict); RL-2 general-budget vs §3.2 `U=0` is
deliberate instantiation; design-only posture intact (owner §5 → change intent → implementation →
package verdict); RL-9 ∧ RL-3 together prove the repair fixed the reported defect.

## Reviewer identity
Independent non-author reviewer (fresh-context subagent dispatched by the owning session, 2026-11-09;
same reviewer as round 1, re-engaged). Did not author the spec.
