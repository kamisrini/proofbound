# VD-p7-behavior-lock-subject-2026-10-06: proposed bounded replay lock subject

**Status:** ACCEPTED FOR PLANNING — fresh non-author review round 3 on 2026-10-13; implementation separately gated
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
projection table, or a ledger event. The planning-only artifact profile below now defines its
canonical encoding, revision identity, provenance, and proof binding; a separately authorized
implementation addendum is still required and may not introduce a new wire identity or writable
truth.

## Round-3 planning contract amendment

The amended planning profile uses UTF-8 RFC 8785 canonical JSON with exactly these required
top-level members and no unknown or null members: `subject`, `lock_revision`, `source_prefix`,
`candidates`, `baseline`, `candidate_snapshot`, `verdict`, `sequence_map`, `replay_proof`, and
`intent_binding`. Arrays are sequence-ordered; byte strings are lowercase hexadecimal; digests are
lowercase SHA-256 values. The lock identity is the SHA-256 digest of its canonical bytes. A
revision contains the subject, a positive revision number, and either the prior lock digest or
`genesis`.

The exact baseline provenance tuple is `(ThroughSeq, ordered source event IDs and sequence
numbers, source digest, accepted projector revision, accepted projector revision digest, baseline
snapshot digest)`. The source and snapshot values must reproduce the same append-only prefix and
projector derivation. Missing or mismatched provenance yields derived `UNVERIFIED`; it is not a
twin verdict, ledger event, projection row, or authored satisfaction outcome.

The exact replay-proof binding requires retrievable committed `vera.replay.v1` proof bytes (or a
content-addressed evidence artifact). Its through-sequence, source, candidate, baseline, and
candidate-snapshot digests must equal the lock. A digest without retrievable proof bytes is
insufficient.

When intent is declared, `intent_binding` must resolve the P5 provider-scoped tuple
`record_kind=requirement`, `source=intent.<provider>`, `record_id`, `artifact_sha256`, and one
`obligation_id`. When no exact requirement is authorized, it must be explicit `UNDECLARED`; that
means no intent claim was made and results in derived `UNVERIFIED`, consistent with P5 P-INV-29.
No `UNVERIFIED` row is stored and neither route fabricates satisfaction.

Planning acceptance is pass/fail against the lock digest and a reviewed evidence revision:
canonical identity and ancestry; source/candidate role and digest binding; complete baseline
provenance; exact replay-proof binding; exact intent resolution or explicit `UNDECLARED`;
positive/negative, stale-revision, missing-evidence, hostile, clean-clone, and replay controls;
calibrated neutral/compile-invalid/lethal controls; no undeclared mutation survivor; and a fresh
non-author acceptance verdict. Any failure blocks acceptance and implementation.

## Boundaries and acceptance

The subject excludes calibration, prediction events, production traffic, a production-shaped twin,
external snapshot providers, new feeds, new event kinds, new persistence surfaces, and
regeneration implementation. Synthetic candidates and fixtures are test inputs, never observed
production behavior.

Before implementation, a fresh non-author review must accept, amend, or reject this exact subject
and the lock contract. Acceptance also requires provenance, stale-revision, missing-evidence,
clean-clone, hostile, and calibrated mutation controls, followed by an independent verdict on any
implementation.

The planning review has accepted this boundary. The snapshot-provider skip and all existing frozen
wire identities remain unchanged; implementation remains separately gated.

## Fresh non-author review outcome — round 2 — 2026-10-13

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
external-feed implementation is authorized. Another fresh non-author review is required for this
round-3 amendment.

## Fresh non-author review outcome — round 3 — 2026-10-13

The fresh independent review is recorded in
[`p7-behavior-lock-plan-round3-20261013.md`](../verification/verdicts/p7-behavior-lock-plan-round3-20261013.md)
against amended planning commit `d9dd491c0140beb110885195a607ed754eec7556`.

**Outcome: ACCEPT for planning.** The reviewer found no blocking, high, medium, or low findings.
The amended contract coherently specifies baseline provenance, derived `UNVERIFIED` semantics,
canonical lock identity and revision ancestry, exact `vera.replay.v1` proof binding, provider-
scoped intent obligation or explicit `UNDECLARED`, and pass/fail acceptance criteria.

This acceptance is limited to the planning subject and contract. It does not authorize behavior-
lock implementation, regeneration, a new wire identity, a new event kind, a provider, an external
feed, or a second writable truth. Proceed only to a separately authorized lock-schema/addendum and
fixture-planning slice, followed by independent implementation review.
