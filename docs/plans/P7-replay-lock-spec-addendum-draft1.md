# P7 `ReplayIsolated` lock — proposed twin SPEC addendum

**Status:** AMENDED FOR ROUND 3; fresh non-author review required. The round-2 `AMEND` is recorded in
[`p7-replay-lock-slice-round2-20261014.md`](../verification/verdicts/p7-replay-lock-slice-round2-20261014.md).
The round-1 `AMEND` is recorded in
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
| `source_prefix` | `{through_seq, events, source_digest, full_source_sha256, ledger_before_sha256, ledger_after_sha256}`. `events` is the complete ordered prefix list of `{seq, event_id, content_sha, canonical_event_hex}`. Each hex value decodes to the exact canonical full `core.Event` object read from the ledger, including both timestamps and connector version. `full_source_sha256` binds the complete prefix using the ledger-evidence encoding below. Verify every field against a fresh ledger read and independently recompute the existing twin `digestRecords` value as `source_digest`; the latter retains its frozen P4 meaning and is not a substitute for the full-record binding. The two ledger digests bind complete before/after reads, including records beyond `through_seq`. |
| `candidates` | `{events, candidate_digest}`. Each ordered event has `{seq, event_id, canonical_event_hex}`; decode the complete canonical `core.Event` JSON, validate it, independently recompute SHA-256 of its RFC 8785 canonical payload and require equality with `content_sha`, then recompute the existing twin candidate digest from the resulting records. Candidate sequences strictly increase after `through_seq`. These are authored inputs, never observations. |
| `baseline` | `{through_seq, source_events, source_digest, projector_revision, projector_revision_digest, snapshot_digest, snapshot_evidence_sha256}`. The first six entries are the accepted provenance tuple; `source_events` is the ordered `{event_id, seq}` list. The evidence digest points to retrievable canonical snapshot bytes. Rebuild the baseline from the same ledger prefix with the accepted projector revision and match both evidence bytes and the existing twin `digestSnapshot` value. |
| `candidate_snapshot` | `{snapshot_digest, snapshot_evidence_sha256}`. The committed snapshot evidence must match the isolated result and its existing twin digest. |
| `verdict` | Exactly `preserved` or `changed`, reproduced by `CompareSnapshots` on the verified baseline and isolated candidate snapshots. Derived `UNVERIFIED` is never stored here. |
| `sequence_map` | Complete ordered `{event_id, requested, effective}` list for source then candidate records; reproduce from `Result.SequenceMap`. |
| `replay_proof` | `{schema, proof_evidence_sha256}`. `schema` is the frozen `vera.replay.v1` identity. The SHA-256 points to retrievable committed bytes of the exact `Proof` value returned by the isolated replay, serialized exactly as specified below. Decode those bytes and compare every field to the lock and recomputed result: `ThroughSeq`, source digest, candidate digest, baseline snapshot digest, and candidate snapshot digest. A digest without bytes fails. |
| `intent_binding` | Either `{state: "UNDECLARED"}` only, or `{record_kind: "requirement", source, record_id, artifact_sha256, obligation_id}`. In the declared branch `source` is one registered `intent.<provider>` source, and the exact immutable requirement revision and obligation must resolve. No inferred applicability. |

Nested objects also reject unknown, missing, and null members. Empty source and candidate lists are
valid only when their digests, sequence map, snapshots, and proof all verify for that exact empty
case. The source prefix may contain gaps; `through_seq` is a boundary, not an assertion that an
event exists at that number. The candidate sequence map must preserve the requested and effective
values actually returned by P4; a proposed remapping would require a separate decision.

### Full-event and whole-ledger evidence

`canonical_event_hex` is lowercase hex of RFC 8785 canonical JSON for the **complete** existing
`core.Event` shape: `event_id`, `source`, `native_id`, `kind`, `occurred_at`, `recorded_at`,
`payload`, `content_sha`, and `connector_version`. Both times are normalized to UTC and encoded
with Go `time.RFC3339Nano` (including its omission of trailing fractional zeroes); no field may
be omitted or null. The payload is the decoded JSON value, not a
quoted JSON byte string. On both observed and candidate events, require the decoded event to
round-trip to the same canonical bytes and independently check
`content_sha == SHA256(RFC8785(payload))`. The observed fields must also equal the stored record's
fields, not merely an author's digest. This is an additional P7 evidence check; it does not change
P4's `Proof.SourceDigest` or `digestRecords` algorithm.

The ledger evidence bytes are UTF-8 RFC 8785 canonical JSON with exactly one member, `records`,
an array of `{seq, event}` objects sorted by ascending sequence. `event` is the complete event
object encoded as above. An empty ledger is `{"records":[]}`. `full_source_sha256` is SHA-256 of
these bytes for records with `seq <= through_seq`; `ledger_before_sha256` and
`ledger_after_sha256` are SHA-256 of the complete ledger bytes before and after the call,
respectively. Each digest must resolve to committed bytes in the reviewed evidence revision.
Immediately before and after **every** success or failure call, read the complete source ledger
under the fixture's exclusive-writer guarantee; require byte-for-byte equality of the two reads,
equality of their digests, and exact agreement of the selected prefix with `source_prefix.events`.
If an exclusive read/writer boundary cannot be established, preservation is `UNVERIFIED`. This
also catches mutation of records after `through_seq` and must not be replaced by a prefix-only
count or digest.

### Projector revision and build manifest

`projector_revision` identifies an accepted committed Git revision. Its digest is proposed as
SHA-256 of the following UTF-8 RFC 8785 canonical JSON manifest bytes, with exactly these members
and no nulls:

```json
{"build":{"build_command":"<exact-command>","cgo_enabled":false,"dependencies":[{"mod_sha256":"<sha256>","module":"<module-path>","version":"<version>","zip_sha256":"<sha256>"}],"executable_sha256":"<sha256>","gcflags":"","go_version":"go1.26.7","goarch":"amd64","goos":"linux","ldflags":"","tags":[],"toolchain_sha256":"<sha256>"},"files":[{"mode":"100644","path":"kernel/go.mod","sha256":"<sha256>"}],"projector_revision":"<full-commit-id>","runtime":{"database_collation":"<value>","database_encoding":"UTF8","database_locale":"<value>","database_timezone":"UTC","postgres_binary_sha256":"<sha256>","postgres_version":"<value>"}}
```

The displayed file is shape only: `files` must enumerate **every** tracked Git-tree entry under
`kernel/` at that commit, sorted by UTF-8 path bytes, each with exact Git mode, repo-relative path,
and SHA-256 of exact blob bytes. Symlinks, submodules, missing blobs, and untracked runtime inputs
are unsupported and fail this profile. `toolchain_sha256` hashes a separately committed canonical
manifest of all regular Go toolchain files under the selected `GOROOT`, sorted by relative UTF-8
path with mode and SHA-256 of exact bytes; symlinks are unsupported there too. The `build` values
are the actual verified values, not hard-coded example values; nonempty tags are sorted strings,
and unsupported build flags or environment overrides fail. `dependencies` enumerates the complete
selected external Go module graph for that build (`go list -m all`, excluding the main module),
sorted by UTF-8 module path then version. Each `mod_sha256` hashes the exact downloaded `.mod`
bytes and each `zip_sha256` hashes the exact raw `.zip` bytes, **not** the `go.sum` `h1` value.
Both byte artifacts must be retrievable from the reviewed evidence revision; compare the declared
graph, both SHA-256 values, and the applicable `go.sum` entries before building with those
verified bytes. A missing archive, unexpected selected module, local `replace`, vendor tree,
workspace override, or unrecorded module-cache input yields `UNVERIFIED`. `executable_sha256`
hashes the exact built executable bytes, also retained in that evidence revision. Build with the
declared command and inputs, require its output to match those bytes, and run that exact executable
for baseline and candidate projection; `go run` or an unbound binary is insufficient.

The runtime values are captured from the same isolated replay environment, including the
Postgres executable's exact bytes and effective database settings. If any additional
projection-affecting input is discovered, the manifest is incomplete and the lock is `UNVERIFIED`
pending amendment. The manifest, selected module graph, `.mod` and `.zip` files, full commit ID,
toolchain manifest, and built executable must be retrievable in the reviewed evidence revision.
A code, dependency, build, or runtime change creates a different revision digest; Git commit
alone does not identify executable behavior.

### Snapshot and proof evidence bytes

Snapshot evidence bytes are UTF-8 RFC 8785 canonical JSON with exactly one member, `tables`,
whose value has **exactly the table-key set actually present** in the returned
`projections.Snapshot.Tables` map. A missing table key stays absent; never synthesize it from a
schema or query list. A present key with zero rows, if encountered, remains an empty array; it is
not equivalent to an absent key for the frozen `digestSnapshot`. Each present value is an array
of lowercase row-digest strings sorted by UTF-8 bytes. No duplicate table names or nulls are
allowed. The empty map encodes as `{"tables":{}}`, without a trailing newline, with SHA-256
`d4fdaf3ef6e7b78dc1a224d36c8f3530f79170316cdaadf83a757322bb82badc`. By contrast,
`{"tables":{"commits_view":[]}}` hashes to
`b495215a7b37edf6e10fd4f6715f232b768928426678ebc84d6607c78320ee1a`; that extra key
must be rejected when the projector omitted it. Independently decode the evidence, require an
exact key-set and row-list match to the returned snapshot, and recompute the current
`digestSnapshot` (sorted present table names and rows, length-framed bytes) without changing that
P4 algorithm. The evidence SHA-256 and twin snapshot digest are separate checks.

Proof evidence bytes are UTF-8 RFC 8785 canonical JSON with exactly six members:
`baseline_snapshot_digest`, `candidate_digest`, `candidate_snapshot_digest`, `schema`,
`source_digest`, and `through_seq`. The four digests are lowercase SHA-256; `through_seq` is a
nonnegative JSON integer; `schema` is exactly `vera.replay.v1`. The names are an evidence encoding
of the six existing in-memory `twin.Proof` fields, **not** a new wire proof or schema. The bytes
must round-trip to that struct and exactly match the result returned by `ReplayIsolated`. The
byte-exact encoding control below uses the all-empty digest value
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` and zero sequence;
the single JSON line, **without a trailing newline**, is the fixture input:

```json
{"baseline_snapshot_digest":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","candidate_digest":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","candidate_snapshot_digest":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","schema":"vera.replay.v1","source_digest":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","through_seq":0}
```

The SHA-256 of precisely that line is
`ec7960eda35be3798ac6cb25e0853608937c6d24a2025e8135a146ac8771ca79`.
This is a serialization control, not a claim that a product replay emitted that proof. The
committed corpus must additionally pin actual `ReplayIsolated` proof bytes and compare their
SHA-256 with `proof_evidence_sha256` before implementation acceptance.

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
