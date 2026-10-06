# P7 Behavior-Lock Planning Review — Round 1

**Reviewed artifact:** `docs/plans/P7-behavior-locks-regeneration-plan-draft1.md`
**Reviewed subject:** proposed deterministic replay path in `kernel/internal/twin`
**Reviewed commit:** `c1ba62907cfc019a0143ee0cc46096ae4d96113f`
**Plan author identity in Git:** `VERA Maintainer <maintainer@vera.local>`
**Reviewer identity:** Codex independent review

## VERDICT: AMEND

The proposed subject is directionally sound for a bounded planning slice: it uses existing
repository-native evidence, introduces no provider or event kind, and keeps implementation blocked.
It is not ready for acceptance as a behavior-lock subject because the exact lock surface and
provenance boundary are still unresolved. No behavior-lock, regeneration, snapshot-provider, or
external-feed implementation is authorized by this verdict.

## Evidence reviewed

- The draft's proposed inputs, outputs, exclusions, unresolved questions, deliverables, and
  falsifiers (`docs/plans/P7-behavior-locks-regeneration-plan-draft1.md:32-52,67-107`).
- The twin interface and proving-test matrix, including the distinct injected `Replay` and real
  projector `ReplayIsolated` paths (`kernel/internal/twin/SPEC.md:8-10,14-37,42-71`).
- The accepted P4 boundary, which originally treated the injected projector as the first slice and
  deferred the isolated adapter until separate evidence (`docs/decisions/VD-p4-twin-replay-calibration-2026-08-26.md:8-24`).
- The later accepted isolated implementation and current P6 mechanical evidence. The twin result
  is 31 candidates, 31 killed, 0 invalid, 0 survived, but that artifact explicitly identifies
  itself as author-side mechanical evidence rather than a non-author semantic verdict
  (`docs/verification/p6-task8-store-twin-mutation-e4c8e77.md:3-16,45-47`).

## Required amendments before implementation

1. **Choose the exact subject and boundary.** The draft currently names a “deterministic replay
   path,” lists an injected projector as an input, excludes the isolated database-fork adapter,
   and then claims an end-to-end Proofbound product boundary. Select one exact subject. The review
   recommendation is `ReplayIsolated` as the bounded product behavior, with injected `Replay`
   retained as a supporting conformance seam; alternatively narrow the claim explicitly to the
   injected seam and remove the end-to-end language.
2. **Define the lock artifact before implementation.** Specify its canonical bytes, revision
   identity, source bindings, and exact observable tuple. At minimum, settle whether the lock binds
   the prefix and ordered candidate events, baseline provenance/digest, candidate snapshot,
   verdict, sequence map, and frozen `vera.replay.v1` proof bytes. A supplied baseline must not be
   an unproven arbitrary input that makes the lock vacuous. The lock must remain distinct from the
   frozen replay proof and must not silently create a second writable truth.
3. **Separate fact roles and authorization.** The ledger prefix is observed source evidence;
   candidate events are explicitly supplied inputs, not observations; properties are generated
   artifacts; and replay output is a derived result. Add the exact intent-obligation binding or an
   explicit undeclared state for the lock claim. The route/evidence matrix must prevent synthetic
   candidates or fixtures from being presented as production behavior.
4. **Turn the acceptance bar into exit criteria.** Keep stale-revision, missing-evidence,
   clean-clone, positive, negative, and calibrated mutation controls as required gates. Treat the
   historical 31/31 result only as baseline mechanical evidence bound to `e4c8e77`, not as proof
   that a future behavior-lock detector has power or that the lock matches human intent. Require a
   fresh non-author verdict for the ratified lock and any regenerated implementation.

## Next gate

Amend the planning draft and record the exact bounded subject and lock contract in a decision
record. Obtain a fresh non-author review of that amended boundary. Until then, the repository
remains in planning-only state and implementation is blocked.
