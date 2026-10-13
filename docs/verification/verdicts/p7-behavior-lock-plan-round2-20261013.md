# P7 Behavior-Lock Planning Review — Round 2

**Reviewed artifacts:**

- `docs/plans/P7-behavior-locks-regeneration-plan-draft1.md`
- `docs/decisions/VD-p7-behavior-lock-subject-2026-10-06.md`

**Reviewed commit:** `d8d0339bb5a94f58f6436c555b535da19053e63e`
**Reviewer identity:** independent non-author review context (`Dewey`)

## VERDICT: AMEND

The `ReplayIsolated` boundary is coherent and correctly separated from injected `Replay`. The
exclusions are consistent with the accepted P4/P6 boundaries: no provider, new event kind, wire
identity, production-shaped twin, or regeneration implementation is authorized.

The proposal is not yet sufficient for planning acceptance because material contract requirements
remain underspecified.

## Findings

1. **HIGH — Baseline provenance is not enforceably bound.** The current boundary accepts a
   caller-supplied baseline snapshot, while the proposal requires provenance tied to the same
   source prefix and projector revision. Define the exact provenance tuple, verification relation,
   and representation of `UNVERIFIED`; the existing twin verdict domain only defines `preserved`
   and `changed`.
2. **HIGH — Content-addressed lock identity remains unresolved.** Define canonical encoding,
   required-field semantics, revision identity, and the exact relationship between lock bytes and
   the frozen `vera.replay.v1` proof.
3. **HIGH — Intent-obligation binding is incomplete.** Define the provider-scoped requirement
   revision, artifact digest, obligation ID, and semantics of an undeclared lock; reconcile this
   with the accepted P5 rule that absent satisfaction is derived as `UNVERIFIED`, not stored as an
   authored state.
4. **MEDIUM — Acceptance gates are named but not fully operational.** Make provenance,
   stale-revision, missing-evidence, clean-clone, hostile, positive/negative, detector-power, and
   mutation controls explicit pass/fail exit criteria bound to the lock contract and reviewed
   evidence revision.

The historical P6 result of 31/31 twin mutations killed remains mechanical baseline evidence; it
does not prove future behavior-lock detector power or intent alignment.

## Decision recommendation

Amend the proposal before recording planning acceptance. The boundary and exclusions may remain,
but the lock identity, baseline provenance, intent-obligation/`UNDECLARED` semantics, and complete
acceptance gates must be made explicit. Implementation remains blocked.
