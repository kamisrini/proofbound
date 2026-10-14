# P7 `ReplayIsolated` lock — proposed fixture and evidence matrix

**Status:** DRAFT FOR REVIEW, 2026-10-14. Companion to the
[proposed twin SPEC addendum](P7-replay-lock-spec-addendum-draft1.md) and the accepted
[planning contract](P7-behavior-locks-regeneration-plan-draft1.md). Rows below are required
fixtures and checks, not claims that they have been run or accepted.

## Evidence roles and routes

| Role | Existing route | Required committed evidence |
|---|---|---|
| Authored declaration | This lock artifact and an exact P5 requirement revision if declared | Canonical lock bytes; parent digest; requirement artifact bytes and digest or explicit `UNDECLARED`. |
| Observed source | Append-only `store` ledger through `ThroughSeq` | Event IDs, sequence numbers, exact event payloads, content digests, source digest, and read witness. |
| Authored test input | Ordered `twin.Candidate` values | Canonical full event bytes, candidate digest, and explicit candidate role. No ledger append. |
| Derived baseline | Existing `projections.Projector` applied to the selected prefix | Accepted projector revision and manifest, snapshot evidence bytes, twin snapshot digest, and derivation witness. |
| Derived candidate behavior | `ReplayIsolated` in temporary `store` | Candidate snapshot bytes and twin digest, verdict, complete sequence map, source preservation and cleanup witnesses. |
| Replay proof | Existing `twin.Proof` in the returned result | Exact committed proof bytes under the ratified encoding, SHA-256, and all five matching bindings. |
| Independent intent evidence | Existing P5 review route for the exact requirement and obligation | Content-addressed review evidence and non-author verdict; no inferred satisfaction. |
| Gate result | Reviewed lock and evidence revision | Case manifest with lock digest, fixture digests, commands, outcomes, reviewer, and commit. |

The repository already has [13 registered source-kind routes](../verification/p6-event-route-matrix.md)
through the twin consumer. That route proof establishes coverage of the existing event vocabulary;
it does not by itself prove a behavior lock or baseline provenance. The first pilot may use a
synthetic GitHub deployment candidate from existing `ReplayIsolated` tests, but the fixture must
label it as authored test input, never an observed production event.

## Required cases

Each case records input artifact digests, expected outcome, actual outcome, check command, and the
reviewed evidence revision. `PASS` means the result matches the expectation; a deliberately
rejected lock is a passing negative control only if the gate actually rejects it.

| ID | Case | Expected result |
|---|---|---|
| L01 | Canonical lock round-trip, exact ten-member top level, nested shape, reproducible SHA-256 | Pass; changed byte, extra/null/missing member, or noncanonical encoding rejected. |
| L02 | Revision 1 `genesis` and revision 2 exact parent digest | Pass; skipped number, wrong parent, or rewritten prior artifact rejected. |
| L03 | Prefix with at least two observed records and a `ThroughSeq` that excludes a later record | Pass; excluded record absent; any event ID, sequence, content, or source digest substitution rejected. |
| L04 | Ordered authored candidate with canonical full event bytes and a gap after the prefix | Pass; changed payload, duplicate ID, invalid event, reordered or non-increasing sequence rejected; source ledger unchanged. |
| L05 | Baseline rebuilt from the exact prefix and accepted projector revision | Pass; arbitrary caller baseline, wrong source tuple, wrong snapshot bytes/digest, absent manifest, and stale projector revision produce `UNVERIFIED`. |
| L06 | Isolated candidate snapshot and `CompareSnapshots` yield both `preserved` and `changed` cases | Pass; swapped snapshot or claimed verdict rejected; temporary store removed on success and failure. |
| L07 | Complete source-plus-candidate sequence map | Pass; omitted, duplicate, reordered, or changed mapping rejected. |
| L08 | Retrievable frozen replay proof bytes match all five lock bindings | Pass; absent bytes, wrong digest, schema, through-sequence, source, candidate, baseline, or candidate snapshot binding rejected. |
| L09 | Exact registered P5 requirement revision and one obligation with complete independent review evidence | Pass only for exact source, record ID, artifact digest, obligation ID, and reviewer evidence; cross-provider, stale-revision, absent-obligation, or incomplete-review variants yield `UNVERIFIED`. |
| L10 | Explicit `UNDECLARED` intent binding | Derived `UNVERIFIED`; no satisfaction claim or stored status row. Missing binding is rejected. |
| L11 | Hostile JSON, duplicate keys, malformed digest/hex, path escape, spoofed source, and forged candidate-as-observation | All reject without modifying source ledger or accepting proof. |
| L12 | Replay from a clean clone using only committed artifacts and exact pinned revisions | Same lock digest, snapshots, proof bytes, verdict, and gate result; no uncommitted cache or private feed required. |
| L13 | Neutral, compile-invalid, and lethal calibration controls, then the declared P7 mutation set | Neutral survives, compile-invalid is invalid, lethal is killed; every declared mutation is accounted for with zero undeclared survivors. |

The declared positive intent case in L09 needs a requirement that genuinely applies to this
replay subject. Existing P5 fixtures may test the resolver, but may not be relabeled as a real
`ReplayIsolated` obligation. If no such requirement has been authorized, the first real lock must
use L10 and remain `UNVERIFIED`; it cannot be reported as a successful intent lock.

## Acceptance record

The evidence manifest must name the canonical lock SHA-256, the exact evidence revision SHA-256,
the Git commit containing the fixture corpus, projector revision and manifest digest, all fixture
file digests, toolchain/build identity, case outcomes, and calibrated mutation universe. The
independent reviewer records findings and `ACCEPT`, `AMEND`, or `REJECT` against those immutable
identifiers. The author records the outcome in a verdict and decision; an `AMEND` requires fresh
review of the amended revision. A separate non-author verdict is required for implementation.

No fixture, corpus, run result, or implementation acceptance evidence exists in this draft. The
proof serialization, projector manifest completeness, and real intent requirement remain review
questions in the [addendum](P7-replay-lock-spec-addendum-draft1.md).
