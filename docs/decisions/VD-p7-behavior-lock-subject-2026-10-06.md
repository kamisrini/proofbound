# VD-p7-behavior-lock-subject-2026-10-06: proposed bounded replay lock subject

**Status:** Proposed — pending fresh non-author review
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
