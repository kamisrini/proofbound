# P7 Behavior Locks and Witnessed Regeneration — Planning Draft 1

**Status:** ACCEPTED FOR PLANNING — round-3 non-author review accepted the amended contract;
implementation remains separately gated. No implementation,
new event kind, provider, wire identity, or production capability is authorized by this document.

**Phase:** P7+ capability planning after P6 closure and the resolved snapshot-provider skip.

## Purpose

The Proofbound vision moves software changes from code-centered review toward a durable behavior
lock: typed contracts, generated property evidence, and recorded observations define what the system
must continue to do. Code may then be regenerated between lock revisions, while the lock and its
evidence remain the governing surface.

This draft scopes the smallest honest planning slice for that idea using existing feeds. It does
not assume an external snapshot provider, production-shaped twin, or unrecorded source of truth.

## Relationship to the project goal

Proofbound's central chain is:

business intent -> requirement obligations -> change intent -> exact commit -> independent
verification -> deployment -> observed outcome.

A behavior lock would strengthen the middle of that chain. It would make the observable behavior
being protected explicit and testable, so a code rewrite or regeneration can be judged against the
same durable contract rather than against the implementation that happened to exist before.

The lock is not a replacement for business intent, independent verification, or observed
deployment. It is a derived, evidence-bearing contract between intent and implementation.

## Accepted planning subject — implementation separately gated

The exact proposed product subject is `kernel/internal/twin.ReplayIsolated`, as recorded in the
proposed decision [VD-p7-behavior-lock-subject-2026-10-06](../decisions/VD-p7-behavior-lock-subject-2026-10-06.md).
The injected `Replay` function remains a supporting conformance seam, not the product boundary.

The bounded subject is:

- **Observed source input:** the selected append-only ledger prefix through `ThroughSeq`, including
  the exact source event identities, sequence numbers, and content digests.
- **Authored candidate input:** explicitly sequenced candidate events, with their exact canonical
  event bytes and declared role as candidate inputs rather than observed production events.
- **Baseline input:** a projection snapshot content-addressed to the same source prefix and the
  accepted projector revision. An arbitrary caller-supplied baseline is not acceptable evidence;
  if its provenance is absent, the result remains `UNVERIFIED` and cannot satisfy the lock.
- **Observed behavior boundary:** the isolated temporary-store projection, candidate snapshot,
  preserved/changed comparison, sequence map, and frozen `vera.replay.v1` proof payload.
- **Existing feed:** the repository-native ledger event stream. No external export feed, provider,
  vendor, or new event kind is needed.

The subject excludes calibration, prediction events, production traffic, a production-shaped twin,
any new persistence surface, and all regeneration implementation. `Replay` may prove that the
same request semantics hold through the injected seam, but it cannot substitute for the isolated
product-boundary evidence.

This subject is narrow, deterministic, non-mutating, and already implemented and evidenced. It is
The planning boundary and contract below were accepted by the fresh round-3 non-author review.
That acceptance does not authorize implementation, regeneration, a new wire identity, or a new
writable truth.

## Accepted planning lock contract — implementation separately gated

The behavior lock is a separate authored, content-addressed artifact. It is not the frozen
`vera.replay.v1` proof, not a projection table, and not a new ledger event or wire identity. The
planning profile below defines its canonical encoding and semantic content; a future implementation
addendum may only materialize this profile under separate authorization:

1. lock revision identity and exact subject (`ReplayIsolated`);
2. source-prefix boundary, ordered observed source event identities, and source content digests;
3. ordered authored candidate event identities, sequence numbers, and canonical event bytes;
4. baseline snapshot provenance, accepted projector revision, and baseline digest;
5. candidate snapshot digest, preserved/changed outcome, and complete sequence map;
6. the frozen replay-proof schema, exact proof bytes or digest, and its relation to the lock; and
7. the exact intent-obligation binding, or an explicit `UNDECLARED` state until an authorized
   requirement exists.

The lock revision digest covers the canonical lock bytes. A later lock revision is a new artifact;
it cannot reinterpret an earlier replay proof or verdict. The lock's authored declaration,
generated property corpus, observed ledger evidence, and derived replay result remain separate
roles. Synthetic candidates and fixtures may test the contract but must never be presented as
production observations.

## Round-3 contract amendments

These choices resolve the round-2 review findings for planning purposes. They define a
planning-only artifact profile; they do not introduce a shipped wire schema, event kind,
persistence surface, or implementation authorization.

### Baseline provenance and `UNVERIFIED`

The baseline provenance tuple is exactly:

`(ThroughSeq, ordered source event IDs and sequence numbers, source digest, accepted projector
revision, accepted projector revision digest, baseline snapshot digest)`.

The source event list and source digest must reproduce the selected append-only ledger prefix. The
projector revision and its digest must identify the accepted projector used to derive the baseline;
the baseline snapshot digest must reproduce the content-addressed snapshot for that same prefix and
projector revision. A caller-supplied snapshot without this tuple is not a valid baseline.

The existing twin `Verdict` remains only `preserved` or `changed`. `UNVERIFIED` is a derived lock
evaluation result, not a new twin verdict, ledger event, projection row, or authored satisfaction
outcome. Missing or mismatched baseline provenance produces `UNVERIFIED` and fails the lock gate;
it cannot be converted to `preserved` or `changed` by a caller.

### Canonical lock identity and proof binding

The planning lock artifact uses UTF-8 RFC 8785 canonical JSON with exactly these required top-level
members and no unknown or null members: `subject`, `lock_revision`, `source_prefix`, `candidates`,
`baseline`, `candidate_snapshot`, `verdict`, `sequence_map`, `replay_proof`, and `intent_binding`.
Arrays are ordered by their declared sequence; byte strings are lowercase hexadecimal; digests are
lowercase SHA-256 values. The lock identity is the SHA-256 digest of those canonical bytes. A
`lock_revision` contains the subject, a positive revision number, and either the prior lock digest
or the literal `genesis`; a later revision is a new artifact and cannot rewrite an earlier one.

`replay_proof` must identify exact committed proof bytes (or a content-addressed evidence artifact)
and carry `schema=vera.replay.v1`. Its `ThroughSeq`, source digest, candidate digest, baseline
snapshot digest, and candidate snapshot digest must equal the corresponding lock members. A proof
digest without retrievable bytes is insufficient evidence. This profile is documentary planning
syntax until a separately authorized implementation addendum maps it to code.

### Intent obligation binding and `UNDECLARED`

When declared, `intent_binding` must contain the exact P5 provider-scoped requirement revision:
`record_kind=requirement`, `source=intent.<provider>`, `record_id`, `artifact_sha256`, and one
`obligation_id`. The requirement artifact digest and obligation ID must resolve to the exact
revision; a different provider, revision, or obligation fails closed.

When no exact requirement is authorized, `intent_binding` must contain only the explicit state
`UNDECLARED`. `UNDECLARED` means that the lock author made no intent claim; it is not evidence of
satisfaction. The lock evaluation therefore remains derived `UNVERIFIED`, consistent with P5
`P-INV-29`, and no `UNVERIFIED` row is stored. A declared binding with missing, non-verifiable, or
incomplete review evidence has the same derived `UNVERIFIED` result.

### Explicit pass/fail exit criteria

Every criterion below is evaluated against the lock digest and a content-addressed reviewed
evidence revision. Any failure blocks planning acceptance or implementation authorization:

1. **Identity:** canonical bytes round-trip exactly, required members are complete, the lock digest
   is reproducible, and revision ancestry is valid.
2. **Source/candidates:** the source prefix, event identities, sequence numbers, candidate roles,
   canonical bytes, and digests match; synthetic candidates never become observations.
3. **Baseline:** the full provenance tuple verifies, including projector revision and snapshot
   digest; absent or stale provenance yields `UNVERIFIED` and fails.
4. **Proof:** the exact `vera.replay.v1` bytes are available and every source, candidate, sequence,
   baseline, and candidate-snapshot binding matches the lock.
5. **Intent:** the exact provider-scoped obligation resolves, or explicit `UNDECLARED` yields only
   derived `UNVERIFIED`; neither route may fabricate satisfaction.
6. **Evidence controls:** positive, negative, stale-revision, missing-evidence, hostile-input,
   clean-clone, and replay-from-committed-artifacts cases pass with failure-closed results.
7. **Detector power:** calibrated neutral, compile-invalid, and lethal controls behave as expected;
   the declared mutation set has no undeclared survivor, and the result is bound to the lock and
   reviewed evidence revision.
8. **Independent acceptance:** the amended plan receives a fresh non-author `ACCEPT`, and any
   later implementation or regeneration receives its own independent verdict.

## Planning boundary

This slice may use only existing repository-native intent records, witnesses, checks, projections,
and accepted feeds. It may not:

- select or implement a snapshot provider or external export feed;
- add a vendor, schema, event kind, wire identity, or second writable truth;
- claim production traces or external behavior that the repository has not observed;
- build a production-shaped twin, governor, warranty, marketplace, or user interface;
- regenerate production code before the lock contract and acceptance bar are independently reviewed;
  or
- treat generated tests or a green implementation as proof that the lock matches human intent.

## Deferred implementation questions after planning acceptance

1. Which implementation artifacts will materialize this already-defined planning profile without
   creating a new wire identity or writable truth?
2. Which committed fixture corpus and existing provider routes can prove every criterion without
   adding a provider, feed, or event kind?
3. What is the smallest regeneration boundary, and what evidence proves that regeneration preserves
   the lock, ledger history, and replay-proof semantics?
4. How will the implementation bind the exact independent verdict and reviewed evidence revision
   while keeping planning acceptance separate from implementation acceptance?

## Candidate deliverables

A separately authorized implementation slice would need, at minimum:

- a package SPEC addendum defining the lock schema, source bindings, revision identity, invariants,
  non-goals, and failure-closed behavior;
- a proposed decision record naming `ReplayIsolated` as the bounded product subject;
- a fixture matrix separating authored intent, generated properties, and observed witnesses;
- a route and evidence matrix showing which existing feeds support each lock claim;
- positive, negative, provenance, mutation, stale-revision, missing-evidence, and clean-clone
  fixtures;
- a measured detector-power and regeneration baseline;
- an implementation boundary that can be rebuilt from committed artifacts and the ledger; and
- a non-author independent verdict before acceptance.

## Falsifiers

Stop and re-scope if:

- the selected replay surface has no complete existing evidence feed;
- the lock cannot distinguish authored claims from observed facts;
- generated properties pass while calibrated mutations survive;
- regeneration can satisfy the lock while losing ledger history, migration safety, or replay proof
  evidence;
- the lock requires an external provider, production trace, or new event kind before its boundary
  can be stated honestly; or
- authoring and maintaining the lock costs more than the verification risk it removes.

## Planning sequence

1. Obtain a fresh non-author independent review of this amended draft and the proposed decision.
2. Record accept, amend, or reject; implementation remains blocked unless the exact boundary is
   accepted by that review.
3. Write the lock addendum and fixture matrix for `ReplayIsolated`.
4. Validate the acceptance bar with provenance, hostile, and mutation controls.
5. Implement only the ratified slice, then obtain the independent verdict.
6. Keep regeneration-as-repair separately gated on measured detector power.

## Fresh review outcome — round 2

The fresh non-author review recorded in
[`p7-behavior-lock-plan-round2-20261013.md`](../verification/verdicts/p7-behavior-lock-plan-round2-20261013.md)
reviewed this amended draft and the proposed decision at commit `d8d0339`. Its verdict was
`AMEND`: the `ReplayIsolated` boundary and exclusions are coherent, but baseline provenance and
`UNVERIFIED` semantics, canonical lock identity and proof binding, intent-obligation/
`UNDECLARED` semantics, and explicit pass/fail acceptance gates remain underspecified.

## Fresh review outcome — round 3

The fresh non-author review recorded in
[`p7-behavior-lock-plan-round3-20261013.md`](../verification/verdicts/p7-behavior-lock-plan-round3-20261013.md)
reviewed this amended draft and the proposed decision at commit
`d9dd491c0140beb110885195a607ed754eec7556`. Its verdict was `ACCEPT`, with no blocking, high,
medium, or low findings. The reviewer accepted the planning boundary and contract while noting
that the current implementation does not materialize the lock artifact or baseline-provenance
mechanism.

## Current stop point

Planning acceptance is complete. The next stage is a separately authorized lock-schema/addendum and
fixture-planning slice for `ReplayIsolated`. No behavior-lock or regeneration implementation may
begin until that slice is authorized, its acceptance controls are evidenced, and the resulting
implementation receives an independent verdict.
