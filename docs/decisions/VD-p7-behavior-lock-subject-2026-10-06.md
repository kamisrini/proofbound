# VD-p7-behavior-lock-subject-2026-10-06: proposed bounded replay lock subject

**Status:** AMEND — fresh non-author review round 2 on 2026-10-13
**Date:** 2026-10-06

## Context

The first independent planning review of
[`P7-behavior-locks-regeneration-plan-draft1.md`](../plans/P7-behavior-locks-regeneration-plan-draft1.md)
returned `AMEND` in
[`p7-behavior-lock-plan-round1.md`](../verification/verdicts/p7-behavior-lock-plan-round1.md).
The review found that the draft mixed the injected test seam with the isolated product boundary,
left baseline provenance and lock identity unresolved, and did not yet separate all fact roles.

This is a proposed planning decision only. It does not authorize implementation, a new event kind,
provider, vendor, wire identity, or second writable truth.

## Proposed decision

Use `kernel/internal/twin.ReplayIsolated` as the first behavior-lock subject. Retain injected
`Replay` as a supporting conformance seam, but do not present it as the product boundary.

The proposed lock binds one exact bounded replay:

- an observed append-only ledger prefix through `ThroughSeq`, with event identities and content
  digests;
- explicitly authored, ordered candidate events with canonical event bytes;
- a baseline snapshot derived from the same source prefix and accepted projector revision, with
  content-addressed provenance;
- the isolated candidate projection, snapshot comparison, complete sequence map, and frozen
  `vera.replay.v1` proof bytes or digest; and
- the exact intent-obligation binding, or explicit `UNDECLARED` until an authorized requirement
  exists.

The lock is a separate authored, content-addressed artifact. It is not the replay proof, a
projection table, or a ledger event. Its canonical encoding and revision digest require a future
lock addendum; no wire identity is introduced by this proposal.

## Boundaries and acceptance

The subject excludes calibration, prediction events, production traffic, a production-shaped twin,
external snapshot providers, new feeds, new event kinds, new persistence surfaces, and
regeneration implementation. Synthetic candidates and fixtures are test inputs, never observed
production behavior.

Before implementation, a fresh non-author review must accept, amend, or reject this exact subject
and the lock contract. Acceptance also requires provenance, stale-revision, missing-evidence,
clean-clone, hostile, and calibrated mutation controls, followed by an independent verdict on any
implementation.

Until that review accepts the proposal, the snapshot-provider skip and all existing frozen wire
identities remain unchanged and implementation is blocked.

## Fresh non-author review outcome — 2026-10-13

The fresh independent review is recorded in
[`p7-behavior-lock-plan-round2-20261013.md`](../verification/verdicts/p7-behavior-lock-plan-round2-20261013.md)
against commit `d8d0339bb5a94f58f6436c555b535da19053e63e`.

**Outcome: AMEND.** The `ReplayIsolated` boundary and exclusions are coherent, but planning
acceptance requires these amendments:

1. Define enforceable baseline provenance, its verification relation to the source prefix and
   projector revision, and the representation of `UNVERIFIED`.
2. Define canonical lock encoding, required-field semantics, revision identity, and the exact
   binding between lock bytes and the frozen `vera.replay.v1` proof.
3. Define the provider-scoped intent-obligation revision/artifact/obligation binding and reconcile
   `UNDECLARED` with the accepted P5 `UNVERIFIED` semantics.
4. Make provenance, stale-revision, missing-evidence, clean-clone, hostile, positive/negative,
   detector-power, and mutation controls explicit pass/fail exit criteria bound to the lock and
   reviewed evidence revision.

The proposal remains planning-only. No behavior-lock, regeneration, snapshot-provider, or
external-feed implementation is authorized. Another fresh non-author review is required after
these amendments.
