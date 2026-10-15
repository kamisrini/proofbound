# P7 `ReplayIsolated` lock — proposed fixture and evidence matrix

**Status:** ACCEPTED FOR DOCUMENTARY PLANNING by fresh non-author round-4 review; runtime fixtures
and implementation remain separately gated. See
[`p7-replay-lock-slice-round4-20261015.md`](../verification/verdicts/p7-replay-lock-slice-round4-20261015.md).
The round-3 `AMEND` is recorded in
[`p7-replay-lock-slice-round3-20261014.md`](../verification/verdicts/p7-replay-lock-slice-round3-20261014.md).
The round-2 `AMEND` is recorded in
[`p7-replay-lock-slice-round2-20261014.md`](../verification/verdicts/p7-replay-lock-slice-round2-20261014.md).
The round-1 `AMEND` is recorded in
[`p7-replay-lock-slice-round1-20261014.md`](../verification/verdicts/p7-replay-lock-slice-round1-20261014.md).
Companion to the
[proposed twin SPEC addendum](P7-replay-lock-spec-addendum-draft1.md) and the accepted
[planning contract](P7-behavior-locks-regeneration-plan-draft1.md). Rows below are required
fixtures and checks, not claims that they have been run or accepted.

## Evidence roles and routes

| Role | Existing route | Required committed evidence |
|---|---|---|
| Authored declaration | This lock artifact and an exact P5 requirement revision if declared | Canonical lock bytes; parent digest; requirement artifact bytes and digest or explicit `UNDECLARED`. |
| Observed source | Append-only `store` ledger through `ThroughSeq`, plus complete-ledger preservation | Canonical complete event bytes (including times and connector version), independently checked payload digests, full-prefix SHA-256, frozen twin source digest, and complete-ledger before/after canonical bytes and SHA-256 on success and failure. |
| Authored test input | Ordered `twin.Candidate` values | Canonical full event bytes, candidate digest, and explicit candidate role. No ledger append. |
| Derived baseline | Existing `projections.Projector` applied to the selected prefix | Accepted projector revision; canonical whole-kernel, toolchain, selected-module, executable, and runtime manifest bytes; Git-tracked Go toolchain and embedded-Postgres archives, toolchain manifest, `.mod`, `.zip`, and executable bytes and digests; exact snapshot-map evidence bytes; frozen twin snapshot digest; derivation witness. |
| Derived candidate behavior | `ReplayIsolated` in temporary `store` | Candidate snapshot bytes and twin digest, verdict, complete sequence map, source preservation and cleanup witnesses. |
| Replay proof | Existing `twin.Proof` in the returned result | Exact committed six-field proof bytes under the pinned encoding, SHA-256, all five matching bindings, and byte-exact serialization control. |
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
| L03 | Prefix with at least two observed records and a `ThroughSeq` that excludes a later record | Pass; excluded record absent from prefix but included in full-ledger evidence. Mutate `occurred_at`, `recorded_at`, or `connector_version` while preserving P4 `source_digest`: full-event and full-prefix binding rejects. Any other event ID, sequence, content, or source digest substitution rejects. |
| L04 | Ordered authored candidate with canonical full event bytes and a gap after the prefix | Pass; changed payload with unchanged `content_sha`, duplicate ID, invalid event, reordered or non-increasing sequence rejected by independent payload-hash and shape checks. Candidate remains authored input. |
| L05 | Baseline rebuilt from the exact prefix and accepted projector revision | Pass; arbitrary caller baseline, wrong source tuple, wrong snapshot bytes/digest, absent manifest, changed kernel blob, changed build/toolchain, selected module, executable or Postgres bytes, changed effective settings, and stale projector revision produce `UNVERIFIED`. Manifest must match its canonical bytes and SHA-256. |
| L06 | Isolated candidate snapshot and `CompareSnapshots` yield both `preserved` and `changed` cases | Pass; swapped snapshot or claimed verdict rejected; temporary store removed on success and failure. Capture entire source ledger before and after each successful and failing replay, including the record beyond `ThroughSeq`; byte-for-byte equality required. A changed excluded record or missing exclusive-writer guarantee yields `UNVERIFIED`. |
| L07 | Complete source-plus-candidate sequence map | Pass; omitted, duplicate, reordered, or changed mapping rejected. |
| L08 | Retrievable frozen replay proof bytes match all five lock bindings | Pass; the addendum's no-newline control hashes to `ec7960eda35be3798ac6cb25e0853608937c6d24a2025e8135a146ac8771ca79`, and actual emitted proof bytes round-trip to the six existing fields. Absent bytes, wrong digest, schema, through-sequence, source, candidate, baseline, candidate snapshot, or added field rejected. |
| L09 | Exact registered P5 requirement revision and one obligation with complete independent review evidence | Pass only for exact source, record ID, artifact digest, obligation ID, and reviewer evidence; cross-provider, stale-revision, absent-obligation, or incomplete-review variants yield `UNVERIFIED`. |
| L10 | Explicit `UNDECLARED` intent binding | Derived `UNVERIFIED`; no satisfaction claim or stored status row. Missing binding is rejected. |
| L11 | Hostile JSON, duplicate keys, malformed digest/hex, path escape, spoofed source, and forged candidate-as-observation | All reject without modifying source ledger or accepting proof. |
| L12 | Replay from a clean clone using only committed artifacts and exact pinned revisions | With network disabled and empty build/runtime caches, resolve every build/runtime digest to a regular Git-tracked byte artifact in the reviewed evidence revision. Safely unpack the committed Go archive, match the exact nested toolchain manifest, build from committed `.mod`/`.zip` bytes, match and run the committed executable, and supply the committed embedded-Postgres archive to the temporary store. Same lock digest, snapshots, proof bytes, verdict, and gate result. If the current P4 temporary store cannot use the verified archive without a separately reviewed implementation route, L12 remains `UNVERIFIED`; no external download or private feed fills the gap. |
| L13 | Neutral, compile-invalid, and lethal calibration controls, then the declared P7 mutation set | Neutral survives, compile-invalid is invalid, lethal is killed; every declared mutation is accounted for with zero undeclared survivors. |
| L14 | Empty snapshot-map encoding versus an injected absent-table key | An actual empty `Snapshot.Tables` map encodes as no table keys and hashes to the addendum's `d4fdaf3e…bb82badc` control; adding `"commits_view":[]` changes evidence bytes and frozen `digestSnapshot` and is rejected. A genuinely present empty key, if produced, remains present; absence is never filled from the query list. |
| L15 | Executable and dependency-byte substitution with unchanged `kernel/` Git blobs | Changed built executable, changed `.mod` or `.zip` bytes, missing module archive, unexpected selected graph entry, or unrecorded module-cache input yields `UNVERIFIED`; raw SHA-256 checks and applicable `go.sum` cross-checks must both pass before rebuild. |
| L16 | Canonical nested Go toolchain manifest and archive closure | Exact `{"files":[{"mode":"100755","path":"bin/go","sha256":"<sha256>"}]}` shape (values illustrative), sorted unique relative paths, complete extracted regular-file set, modes, and raw-byte digests pass. Changed envelope/member, ordering, path escape, missing/extra file, unsafe link, altered archive bytes, absent committed archive, or a hash-only manifest yields `UNVERIFIED`. |
| L17 | Embedded-Postgres cold-start archive closure | The exact committed `txz` bytes safely provision the temporary store offline and the launched server executable matches `postgres_binary_sha256`. Missing or changed archive, unsafe extraction, unbound launched binary, network fetch, or inability to supply the archive through a reviewed route yields `UNVERIFIED`. |

The declared positive intent case in L09 needs a requirement that genuinely applies to this
replay subject. Existing P5 fixtures may test the resolver, but may not be relabeled as a real
`ReplayIsolated` obligation. If no such requirement has been authorized, the first real lock must
use L10 and remain `UNVERIFIED`; it cannot be reported as a successful intent lock.
The owner-approved [replay evidence requirement](../intent/records/requirements/BR-replay-evidence-6e3d2b/0001.md)
is an active declared P5 revision. Its exact `O-1` [independent requirement review](../verification/verdicts/p7-replay-requirement-review-round1-20261014.md)
found it `VERIFIABLE`, not satisfied. An independent satisfaction verdict and real lock evidence
remain pending; the L09 positive case cannot be reported green yet.

## Acceptance record

The evidence manifest must name the canonical lock SHA-256, the exact evidence revision SHA-256,
the Git commit containing the fixture corpus, projector revision and manifest digest, all fixture
file digests, Go toolchain and Postgres archive digests, nested toolchain manifest digest,
selected module `.mod`/`.zip` and executable digests,
case outcomes, and calibrated mutation universe. The
independent reviewer records findings and `ACCEPT`, `AMEND`, or `REJECT` against those immutable
identifiers. The author records the outcome in a verdict and decision; an `AMEND` requires fresh
review of the amended revision. A separate non-author verdict is required for implementation.

No runtime fixture corpus, run result, or implementation acceptance evidence exists in this draft.
The pinned proof encoding and projector manifest are still proposals awaiting fresh independent
review; the encoding control is not an emitted replay proof. The real intent requirement is
reviewed for verifiability, but its satisfaction remains unverified.
