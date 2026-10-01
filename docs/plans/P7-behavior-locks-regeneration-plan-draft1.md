# P7 Behavior Locks and Witnessed Regeneration — Planning Draft 1

**Status:** DRAFT — planning only. No implementation, new event kind, provider, or production
capability is authorized by this document.

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

## Proposed bounded subject — pending independent review

The recommended first subject is the deterministic replay path in kernel/internal/twin:

- Inputs are a selected ledger prefix, explicitly sequenced candidate events, a baseline
  projection snapshot, and an injected projection function.
- Outputs are the replay comparison and the frozen vera.replay.v1 proof payload.
- Existing evidence includes kernel/internal/twin/SPEC.md, replay proving tests, the accepted P4
  decision and verdict, and calibrated P6 mutation evidence of 31 killed, 0 invalid, and 0 survived.
- The existing ledger event stream is the input feed. No external export feed or new event kind is
  needed.
- The first subject excludes pure calibration, an isolated database-fork adapter, prediction
  events, production traffic, and any regeneration implementation.

This is preferred over selecting the whole twin package because it is deterministic, non-mutating,
already evidence-bearing, and narrow enough to expose whether a behavior lock can protect a real
Proofbound product boundary. The core package already has an interface-lock mechanism, but its
low-level primitives are less representative of the end-to-end behavior-lock goal.

This is a scope recommendation, not an independent acceptance. Law 9 requires a non-author review
before the subject or any implementation is accepted.

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

1. What exact replay behavior is the lock: candidate ordering, prefix selection, projection output,
   proof bytes, or the complete bounded combination?
2. What exact artifact is the lock, how is its revision content-addressed, and which intent
   obligations authorize it?
3. Which portions are authored declarations, which are generated artifacts, and which are observed
   witnesses? The three must remain distinguishable.
4. What property corpus can be generated from the replay lock without pretending that synthetic
   fixtures are production observations?
5. What detector-power and mutation controls show that the generated corpus can distinguish a
   broken implementation from a passing one?
6. What is the smallest regeneration boundary, and what evidence proves that regeneration preserves
   the lock, ledger history, and replay proof semantics?
7. What independent verdict is required before a lock or regenerated implementation is accepted?

## Candidate deliverables

A separately authorized implementation slice would need, at minimum:

- a package SPEC addendum defining the lock schema, source bindings, revision identity, invariants,
  non-goals, and failure-closed behavior;
- a fixture matrix separating authored intent, generated properties, and observed witnesses;
- a route and evidence matrix showing which existing feeds support each lock claim;
- positive, negative, mutation, stale-revision, missing-evidence, and clean-clone fixtures;
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

1. Obtain a non-author independent review of this draft and the proposed twin-replay subject.
2. Record the selected bounded subject and any amendments as a decision before implementation.
3. Write the lock addendum and fixture matrix for the replay path.
4. Validate the acceptance bar with hostile and mutation controls.
5. Implement only the ratified slice, then obtain the independent verdict.
6. Keep regeneration-as-repair separately gated on measured detector power.

## Current stop point

The repository-grounded preflight recommends the twin replay path, but independent review and scope
decision are still required. Until they occur, no behavior-lock or regeneration implementation is
authorized.
