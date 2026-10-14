# P7 `ReplayIsolated` lock — proposed twin SPEC addendum

**Status:** AMEND REQUIRED after fresh non-author round-1 review on 2026-10-14; see
[`p7-replay-lock-slice-round1-20261014.md`](../verification/verdicts/p7-replay-lock-slice-round1-20261014.md).
This is a proposed addendum to
[`internal/twin/SPEC.md`](../../kernel/internal/twin/SPEC.md), not a live package contract.
The bounded subject and planning profile were accepted in
[`VD-p7-behavior-lock-subject-2026-10-06`](../decisions/VD-p7-behavior-lock-subject-2026-10-06.md).
Adoption, implementation, and acceptance remain separate decisions.

## Scope and ownership

One authored lock describes one call to `ReplayIsolated` over one observed ledger prefix and one
ordered candidate stream. The lock is a committed, content-addressed artifact; the source ledger
owns observations, the P5 intent provider owns requirement revisions, and `twin` owns replay
results. The lock does not write any of those stores. `Replay` is a conformance seam only.

The existing `ReplayIsolated` signature accepts a caller-supplied baseline and its `Proof` exists
only as an in-memory struct. It does not verify baseline provenance, emit committed proof bytes, or
evaluate a lock. These are implementation gaps, not present-tense guarantees. Any code design must
first be reviewed against this addendum and preserve the accepted P4 behavior.

## Proposed artifact profile

The artifact is UTF-8 RFC 8785 canonical JSON. Its top level contains exactly `subject`,
`lock_revision`, `source_prefix`, `candidates`, `baseline`, `candidate_snapshot`, `verdict`,
`sequence_map`, `replay_proof`, and `intent_binding`; unknown, missing, or null members fail.
All byte strings use lowercase hexadecimal and all digests are lowercase SHA-256. The lock ID is
SHA-256 of the complete canonical artifact bytes, including its intent binding. No field contains
the lock ID itself.

The nested shape proposed for review is:

| Member | Required value and verification |
|---|---|
| `subject` | Literal `ReplayIsolated`. |
| `lock_revision` | `{subject, number, prior_lock_digest}`. `subject` repeats the exact subject; `number` is a positive integer. Revision 1 uses literal `genesis`; later revisions use the immediately preceding lock digest and increment by one. |
| `source_prefix` | `{through_seq, events, source_digest}`. `events` is the complete ordered list of `{event_id, seq, content_sha}` from the selected ledger prefix. Verify identities, sequence order, content, and the existing twin `digestRecords` value against a fresh read of that prefix. A digest supplied by the author alone is insufficient. |
| `candidates` | `{events, candidate_digest}`. Each ordered event has `{seq, event_id, canonical_event_hex}`; decode the complete canonical `core.Event` JSON, validate it, and recompute the existing twin candidate digest from the resulting records. Candidate sequences strictly increase after `through_seq`. These are authored inputs, never observations. |
| `baseline` | `{through_seq, source_events, source_digest, projector_revision, projector_revision_digest, snapshot_digest, snapshot_evidence_sha256}`. The first six entries are the accepted provenance tuple; `source_events` is the ordered `{event_id, seq}` list. The evidence digest points to retrievable canonical snapshot bytes. Rebuild the baseline from the same ledger prefix with the accepted projector revision and match both evidence bytes and the existing twin `digestSnapshot` value. |
| `candidate_snapshot` | `{snapshot_digest, snapshot_evidence_sha256}`. The committed snapshot evidence must match the isolated result and its existing twin digest. |
| `verdict` | Exactly `preserved` or `changed`, reproduced by `CompareSnapshots` on the verified baseline and isolated candidate snapshots. Derived `UNVERIFIED` is never stored here. |
| `sequence_map` | Complete ordered `{event_id, requested, effective}` list for source then candidate records; reproduce from `Result.SequenceMap`. |
| `replay_proof` | `{schema, proof_evidence_sha256}`. `schema` is the frozen `vera.replay.v1` identity. The SHA-256 points to retrievable committed bytes of the exact `Proof` value returned by the isolated replay, serialized under the reviewed evidence encoding. Decode those bytes and compare every field to the lock and recomputed result: `ThroughSeq`, source digest, candidate digest, baseline snapshot digest, and candidate snapshot digest. The evidence encoding must be ratified before a fixture is accepted; a digest without bytes fails. |
| `intent_binding` | Either `{state: "UNDECLARED"}` only, or `{record_kind: "requirement", source, record_id, artifact_sha256, obligation_id}`. In the declared branch `source` is one registered `intent.<provider>` source, and the exact immutable requirement revision and obligation must resolve. No inferred applicability. |

Nested objects also reject unknown, missing, and null members. Empty source and candidate lists are
valid only when their digests, sequence map, snapshots, and proof all verify for that exact empty
case. The source prefix may contain gaps; `through_seq` is a boundary, not an assertion that an
event exists at that number. The candidate sequence map must preserve the requested and effective
values actually returned by P4; a proposed remapping would require a separate decision.

### Projector revision proposal

`projector_revision` identifies an accepted committed Git revision. Its digest is proposed as
SHA-256 of a canonical manifest of every tracked file under `kernel/` at that revision, with paths
sorted by UTF-8 bytes and each entry containing the path and SHA-256 of its exact blob bytes. This
conservative whole-kernel scope includes projection code, ledger import behavior, migrations, and
the module dependency files. The evidence must include the manifest bytes and the accepted
revision's full commit ID. A code edit creates a new revision digest. The reviewer must confirm
that the build/toolchain identity and any runtime dependency needed for reproducible projection are
also bound before this proposal is adopted.

### Snapshot and proof evidence proposal

Snapshot evidence is a canonical JSON object of table names to ordered row-digest arrays, using
the existing `projections.Snapshot` result. The reviewer must verify that this encoding reproduces
the current `digestSnapshot` algorithm (sorted names and sorted rows, length-framed bytes); the
evidence file SHA-256 and the twin snapshot digest are distinct values and both are checked.

The frozen proof currently has no committed byte artifact or defined serialization in `twin`.
Propose recording canonical JSON of the six existing `Proof` fields as returned, with no new proof
field or schema name. Acceptance requires a byte-exact fixture and compatibility review of this
serialization against the frozen identity and any pinned vectors. Until that review succeeds,
proof-byte binding remains an open implementation prerequisite.

## Evaluation and failure behavior

Evaluate against a lock digest and a content-addressed reviewed evidence revision. Recompute the
lock ID, ancestry, source and candidate digests, baseline derivation, candidate snapshot, verdict,
sequence map, and proof bindings from retrievable inputs. A missing input, conflicting digest,
stale projector revision, invalid candidate, ambiguous intent reference, or incomplete independent
review yields derived `UNVERIFIED` and fails the lock gate. This state is not a `twin.Verdict`,
ledger event, projection row, or authored satisfaction verdict.

`UNDECLARED` explicitly makes no intent claim and also yields derived `UNVERIFIED`. A declared
binding resolves the exact P5 provider-scoped requirement and one obligation. A green intent claim
requires the separately accepted P5 review evidence for that exact revision and obligation; a lock
cannot infer it from a replay verdict. The lock remains failed if an exact declared binding or its
review evidence is missing.

## Proposed proving obligations

The companion [fixture matrix](P7-replay-lock-fixture-matrix-draft1.md) defines the cases and
evidence route for identity, provenance, proof, intent, hostile input, clean-clone replay, and
calibrated mutation controls. Before adoption, a non-author review must resolve the proposed
projector manifest and proof serialization, inspect the committed fixture corpus, and accept,
amend, or reject the exact addendum and evidence revision. An independent implementation verdict
is required after any code is written.

## Non-goals

No snapshot provider, external feed, production traffic, prediction event, calibration change,
new event kind, new wire identity, projection persistence, second writable truth, or regeneration
implementation is part of this addendum.
