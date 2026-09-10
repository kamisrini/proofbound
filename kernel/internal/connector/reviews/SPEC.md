# `internal/connector/reviews` SPEC

## Purpose

The reviews connector reads only artifacts supplied by an injected
`CommittedReader`. It does not inspect the working tree, open a database, or
modify verdict files. Each valid schema-bearing artifact emits one `review.verdict` event via
the normal store appender. The ledger's `(source, native_id, content_sha)` key
makes unchanged ingestion idempotent and changed content a revision.

The verdict directory also contains documentary adjudications and reviewed exhibits committed under
`VD-verdicts-are-artifacts-rl0rab`. A valid UTF-8 Markdown file whose first line is neither `---` nor
`schema:` is counted as documentary and does not claim to implement this wire contract. A file that
starts either marker is a wire candidate and must parse completely or fail closed. Invalid UTF-8 and
invalid paths fail before this distinction, so binary or path-hostile artifacts are never silently
ignored.

## Artifact contract: `vera.verdict.v1`

An artifact is UTF-8 Markdown with a YAML-like front matter block delimited by
`---` on its own line. The block must contain exactly these keys:

```text
---
schema: vera.verdict.v1
verdict_id: task8-current-round1
status: ACCEPTABLE
reviewed_commit: 0123456789abcdef0123456789abcdef01234567
findings:
  - finding_id: F-1
    severity: MED
    defect_commit: 0123456789abcdef0123456789abcdef01234567
artifact_path: docs/verification/verdicts/task8-current-round1.md
artifact_sha: <64 lowercase hexadecimal characters>
---
```

`findings` may be empty (`findings: []`). `defect_commit` is optional. Verdict
IDs and finding IDs are non-empty single-line identifiers. Status is exactly
`ACCEPTABLE` or `NEEDS_WORK`; severity is exactly `HIGH`, `MED`, or `LOW`.
Commits are 40- or 64-character lowercase hexadecimal strings. The artifact
path must be exactly the committed reader's path and must match
`docs/verification/verdicts/<filename>.md`; absolute paths, traversal, and
other extensions are rejected. `artifact_sha` is the lowercase SHA-256 digest
of the complete artifact after normalizing its own value to 64 zeroes. The
event payload retains it alongside the source path, so the event remains bound
to the committed artifact bytes without a self-referential digest.

Unknown, missing, duplicate, null, malformed, or out-of-order metadata fails
closed. Content after the closing delimiter is opaque Markdown and is not
parsed by this connector.

## Artifact contract: `proofbound.obligation-verdict.v2`

V2 is a separate schema, never an interpretation of v1. Its front matter is one compact JSON
object between `---` delimiters. The object has exactly these fields, in this order:
`schema`, `verdict_id`, `status`, `declared_reviewer`, `reviewed_commit`, `change_intent`,
`requirements`, `obligations`, `findings`, `artifact_path`, `artifact_sha256`.

`change_intent` is one canonical intent `Reference`; `requirements` is a non-empty sorted list of
exact requirement `Reference` values. Each obligation outcome contains exactly `source`,
`requirement_id`, `artifact_sha256`, `obligation_id`, `outcome`, and `evidence_event_ids`.
Outcome is `SATISFIED`, `NOT_SATISFIED`, or `INCONCLUSIVE`. Evidence IDs are sorted unique
canonical event IDs and may be empty only for a non-satisfied or inconclusive outcome. Every CI
target appears exactly once and no untargeted obligation may appear. `ACCEPTABLE` requires every
outcome to be `SATISFIED`; `NEEDS_WORK` preserves the complete per-obligation result set.

Findings retain the v1 fields and registries. The artifact digest is SHA-256 over the complete
artifact after replacing only its own 64-hex `artifact_sha256` value with zeroes. Projection-time
validation binds the exact commit, CI and BR revisions and requires cited evidence at a ledger
sequence no later than the verdict. Builder-produced evidence is a claim and cannot by itself
establish satisfaction; the declared reviewer owns the outcome.

## Artifact contract: `proofbound.requirement-review.v1`

Requirement reviews use the same strict compact-JSON front matter. The exact ordered fields are:
`schema`, `review_id`, `requirement` (an exact canonical Reference), `declared_reviewer`, `outcomes`,
`artifact_path`, `artifact_sha256`. Each non-empty outcome contains exactly `obligation_id`,
`outcome`, and optional `finding`; the closed outcomes are `VERIFIABLE`, `AMBIGUOUS`, `UNTESTABLE`,
and `CONTRADICTORY`. Every obligation in the bound requirement revision appears exactly once.

The connector performs syntactic validation. Projection validation rejects a dangling or wrong
revision digest, an absent obligation, and equality between `declared_reviewer` and the bound
requirement's `declared_owner`. Independence is declared, not authenticated. Reviews emit
`requirement.reviewed`; their native ID is `review_id`. Their result is a derived soft cap for all
chains, and is a hard self-hosting acceptance condition only for Proofbound's own P5 chain.

Unknown, duplicate, missing, null, out-of-order, malformed, or non-canonical fields fail closed in
both new schemas. A v2 or requirement-review artifact can never fall back to the v1 parser.

## Interface

```go
const Version = "reviews/1"

type Artifact struct { Path string; Bytes []byte }
type CommittedReader interface {
    ReadCommittedVerdicts(context.Context) ([]Artifact, error)
}
type Appender interface {
    Append(context.Context, core.Event) (store.Record, bool, error)
}
type Deps struct { Reader CommittedReader; IDs *core.IDGenerator; Logger *slog.Logger; Now func() time.Time }
type Connector struct { /* unexported */ }
func New(*Deps) (*Connector, error)
func (c *Connector) Sync(context.Context, Appender) (Result, error)
func Parse(path string, data []byte) (Verdict, error)
func ParseObligationVerdict(path string, data []byte) (ObligationVerdict, error)
func ParseRequirementReview(path string, data []byte) (RequirementReview, error)
```

## Invariants

1. **R-INV-1 — Committed-only input:** only artifacts returned by
   `ReadCommittedVerdicts` are considered; the connector performs no filesystem
   fallback.
2. **R-INV-2 — Exact schema:** front matter has exactly the v1 fields and
   closed registries; malformed UTF-8, delimiters, duplicates, unknown fields,
   nulls, bad paths, or bad hashes fail closed.
3. **R-INV-3 — Event mapping:** source is `reviews`, kind is `review.verdict`,
   native ID is `verdict_id`, payload is the validated verdict, and version is
   `reviews/1`.
4. **R-INV-4 — Idempotent and revision-safe:** unchanged content is absorbed by
   the appender; changed content with the same verdict ID produces a new event.
5. **R-INV-5 — Determinism and preservation:** artifacts are processed by path,
   bytes are never changed, and an error stops later processing with explicit
   progress in `Result`.
6. **R-INV-6 — Dependency safety:** nil dependencies, typed-nil appenders, and
   nil clocks are rejected at construction or sync.
7. **R-INV-7 — Binding:** event payload retains both the committed artifact path
   and artifact SHA; the reader path must equal front matter `artifact_path`.
8. **R-INV-8 — Documentary artifacts are not wire verdicts:** valid plain Markdown is counted and
   skipped without error; anything declaring a front-matter or `schema:` start remains a strict
   candidate and malformed candidates fail closed.
9. **R-INV-9 — Schema dispatch is closed:** v1, v2, and requirement-review artifacts dispatch to
   distinct parsers and event mappings; stored v1 meaning and bytes are unchanged.
10. **R-INV-10 — V2 aggregation is honest:** exact references and every targeted obligation are
    present; `ACCEPTABLE` cannot coexist with a non-satisfied outcome.
11. **R-INV-11 — Evidence is bounded:** every evidence event exists before the verdict and authored
    evidence cannot grade itself; dangling or future evidence fails projection.
12. **R-INV-12 — Requirement review binds exactly:** wrong revision digests, unknown outcomes,
    absent obligations, and equal declared author/reviewer identities fail closed.

## Proving table

| Invariant | Statement | Proving test |
|---|---|---|
| R-INV-1 | Only injected committed artifacts are considered | reviews_test.go::TestSyncUsesOnlyInjectedCommittedReader |
| R-INV-2 | Strict v1 metadata fails closed | reviews_test.go::TestParseStrictFrontMatter |
| R-INV-3 | Valid artifacts emit review verdict events | reviews_test.go::TestSyncMintsReviewVerdictEvent |
| R-INV-4 | Unchanged artifacts are idempotent and changes revise | reviews_test.go::TestSyncIsIdempotentAndRevisionSafe |
| R-INV-5 | Processing is sorted and stops on malformed input | reviews_test.go::TestSyncSortsAndFailsClosed |
| R-INV-6 | Required dependencies are enforced | reviews_test.go::TestNewRequiresDependencies |
| R-INV-7 | Reader path and payload path remain bound | reviews_test.go::TestParseBindsPathAndDigest |
| R-INV-8 | Documentary artifacts are distinct from strict wire candidates | reviews_test.go::TestSyncDistinguishesDocumentaryArtifacts |
| R-INV-9 | Three schemas dispatch separately and v1 remains byte-semantic | reviews_test.go::TestSchemaDispatchAndV1Compatibility |
| R-INV-10 | V2 exact targets and aggregate status fail closed | reviews_test.go::TestObligationVerdictAggregation |
| R-INV-11 | Evidence exists, precedes verdict, and does not self-grade | reviews_test.go::TestObligationVerdictEvidenceValidation |
| R-INV-12 | Requirement reviews bind revision, obligations, and independent identity | reviews_test.go::TestRequirementReviewValidation |
