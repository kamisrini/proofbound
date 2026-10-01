# P7 Snapshot Provider — Planning Draft 1

**Status:** GATE 1 RESOLVED — skip retained; no provider implementation authorized.

**Phase:** P7+ capability planning after P6 closure.

## Purpose

P6 deliberately skipped the conditional snapshot-class intent provider. Its provider-independent
contract remains pinned, but a mutable API-backed provider must not be invented from a generic
vendor assumption. This draft records the resolved Gate 1 boundary and preserves the acceptance
shape for a future, separately authorized phase.

## Authority and boundary

The P6 semantic decision and the founder decision recorded in
VD-p7-snapshot-provider-skip-2026-10-01 retain the no-width doctrine and defer the snapshot
provider. The P5 intent-provider contract remains the source-neutral authority: immutable revision
bytes, artifact digest, canonical mapping, stable provider-scoped identity, point-in-time
resolution, and replay-sufficient ledger payloads are mandatory.

No provider, vendor, export feed, schema, event kind, or implementation is selected here. This
plan does not infer a lawful export license, change a frozen wire identity, or authorize production
code.

## Gate 1 — resolved

The founder decision is to retain the snapshot-provider skip. The pinned contract remains available
for a later explicitly authorized phase. No concrete lawful export feed was named, and no narrow
width exception was granted.

P7 planning may continue only within existing feeds and separately authorized slices. The agent
must not fill in a vendor, API, export format, retention guarantee, or legal authority on the
founder's behalf.

## Conditional provider acceptance shape

This acceptance shape is dormant and does not authorize work. If a later decision authorizes a
narrow exception, the next artifact is a provider SPEC and semantic decision record. Before
implementation, it must pin and fixture-test:

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

## Planning sequence after Gate 1

1. The founder decision and semantic decision record are committed here.
2. Continue only with a separately authorized P7 planning slice that uses existing feeds.
3. The next bounded planning candidate is behavior locks and regeneration; it remains planning until
   its own scope and authorization are recorded.
4. A future provider, if needed, requires a new lawful-feed decision before its SPEC or
   implementation.
5. Any provider acceptance still requires conformance, hostile fixtures, calibrated mutation
   acceptance, an independent verdict, route-matrix coverage, and clean-clone evidence.
6. The remaining P7+ ladder—production-shaped twin, governor, warranties/marketplace, and UI—
   remains separately authorized work.

## Current stop point

Gate 1 is resolved in favor of retaining the skip. The snapshot provider remains deferred and the
repository remains implementation-frozen with respect to this capability. The next planning step
is a bounded behavior-locks/regeneration plan within existing feeds.
