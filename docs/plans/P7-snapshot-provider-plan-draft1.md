# P7 Snapshot Provider — Planning Draft 1

**Status:** DRAFT — planning only. No provider, feed, schema, event kind, or implementation is
authorized by this document.

**Phase:** P7+ capability planning after P6 closure.

## Purpose

P6 deliberately skipped the conditional snapshot-class intent provider. Its provider-independent
contract remains pinned, but a mutable API-backed provider must not be invented from a generic
vendor assumption. This draft establishes the first P7 decision gate and the acceptance shape that
would apply only after that gate is explicitly resolved.

## Authority and boundary

The P6 semantic decision preserves the no-width doctrine and says the snapshot provider is skipped
unless a real export feed is designated. The P5 intent-provider contract remains the source-neutral
authority: immutable revision bytes, artifact digest, canonical mapping, stable provider-scoped
identity, point-in-time resolution, and replay-sufficient ledger payloads are mandatory.

This draft does not amend either authority. It does not select a provider, infer a lawful export
license, add an event kind, change a frozen wire identity, or authorize production code.

## Gate 1 — founder decision

One of these decisions must be recorded before any provider SPEC is minted:

1. **Retain skip:** reaffirm that no snapshot provider is selected; keep the pinned contract for a
   later explicitly authorized phase and proceed to other P7 planning only within existing feeds.
2. **Authorize a narrow exception:** name one concrete lawful export feed, identify its permitted
   access/export boundary, and explicitly amend the no-width doctrine for that provider task.

The agent must not fill in a vendor, API, export format, retention guarantee, or legal authority on
the founder’s behalf. A feed named only as a product category or hypothetical example is not a gate
decision.

## Conditional provider acceptance shape

If the narrow exception is authorized, the next artifact is a provider SPEC and semantic decision
record. Before implementation, it must pin and fixture-test:

- the lawful export boundary and exact source bytes;
- immutable revision identity and artifact SHA-256;
- stable provider-scoped record identity and canonical mapping;
- point-in-time resolution for the referenced commit;
- canonical-payload and artifact-digest separation;
- replay sufficiency from a fresh clone and ledger alone;
- ordering, upstream mutation, digest mismatch, missing revision, cross-provider collision, and
  unmappable-source failure behavior;
- conformance, hostile fixtures, mutation acceptance, independent verdict, route-matrix coverage,
  and clean-clone evidence; and
- the explicit limits of the provider, including any observation-at-ingest rule it cannot improve.

The provider must stop closed on unsupported or unverified source claims. It may not synthesize a
canonical record to make an unmappable export appear complete.

## P7 planning sequence after Gate 1

1. Record the founder decision and mint the semantic decision record.
2. Write and independently review the provider SPEC, fixture matrix, and acceptance bar.
3. Implement only the ratified contract and run conformance and hostile tests.
4. Run calibrated mutation acceptance and the independent verdict round.
5. Update the route matrix, clean-clone evidence, roadmap, and state only after committed proof.
6. Plan the remaining P7+ ladder separately: behavior locks/regeneration, production-shaped twin,
   governor, warranties/marketplace, and UI remain distinct authorization gates.

## Current stop point

This draft stops at Gate 1. The next required input is either a reaffirmation of the P6 skip or a
concrete lawful export feed plus explicit permission for the narrow width exception. Until then,
the repository remains implementation-frozen with respect to this capability.
