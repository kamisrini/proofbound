# `internal/connector/reviews` SPEC

## Purpose

The reviews connector reads only artifacts supplied by an injected
`CommittedReader`. It does not inspect the working tree, open a database, or
modify verdict files. Each valid artifact emits one `review.verdict` event via
the normal store appender. The ledger's `(source, native_id, content_sha)` key
makes unchanged ingestion idempotent and changed content a revision.

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

## Proving table

| Invariant | Proving test |
|---|---|
| R-INV-1 | reviews_test.go::TestSyncUsesOnlyInjectedCommittedReader |
| R-INV-2 | reviews_test.go::TestParseStrictFrontMatter |
| R-INV-3 | reviews_test.go::TestSyncMintsReviewVerdictEvent |
| R-INV-4 | reviews_test.go::TestSyncIsIdempotentAndRevisionSafe |
| R-INV-5 | reviews_test.go::TestSyncSortsAndFailsClosed |
| R-INV-6 | reviews_test.go::TestNewRequiresDependencies |
| R-INV-7 | reviews_test.go::TestParseBindsPathAndDigest |
