# P7 Behavior Locks and Witnessed Regeneration — Planning Draft 1

**Status:** DRAFT — amended after P7 planning review rounds 1 and 2; planning only. No implementation,
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

## Proposed bounded subject — amended, pending resolution of round-2 review

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
still a proposal: the fresh non-author review must accept, amend, or reject this exact boundary.

## Proposed lock contract — pending decision

The behavior lock is a separate authored, content-addressed artifact. It is not the frozen
`vera.replay.v1` proof, not a projection table, and not a new ledger event or wire identity. A
future lock addendum must define the canonical encoding, but its semantic content is fixed here:

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

## Questions the plan must settle before implementation

1. What canonical encoding and revision digest will the lock addendum use, without introducing a
   new wire identity before a separate authorization?
2. Which existing intent obligation authorizes the lock, or how will the explicit `UNDECLARED`
   route be represented until one exists?
3. What generated property corpus covers source bounds, candidate ordering, baseline provenance,
   snapshot comparison, proof binding, stale revisions, and missing evidence?
4. What detector-power and mutation controls show that the generated corpus distinguishes a broken
   `ReplayIsolated` implementation from a passing one?
5. What is the smallest regeneration boundary, and what evidence proves that regeneration preserves
   the lock, ledger history, and replay-proof semantics?
6. What fresh non-author verdict is required before the lock or any regenerated implementation is
   accepted?

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

## Current stop point

Round-2 independent review returned `AMEND`. Amend the planning draft and proposed decision to
resolve those four findings, then obtain another fresh non-author review. Until that review
records `ACCEPT`, no behavior-lock or regeneration implementation is authorized.
