# Repo-native records provider contract

## 1. Source and committed-reader boundary

The provider name is `records`; emitted source is `intent.records`. It reads only bytes returned by
an injected committed-tree reader. The production reader uses `git ls-tree` and `git show` at an
explicit revision, never working-tree file reads. P5 sync uses `HEAD`; commit-trailer resolution uses
the referenced commit tree.

```go
type Artifact struct { Path string; Bytes []byte }
type Reader interface { ReadIntentArtifacts(context.Context, string) ([]Artifact, error) }
type Provider struct { /* unexported */ }
func New(Reader, time.Time) (*Provider, error)
func (p *Provider) Revisions(context.Context) ([]intent.Revision, error)
func (p *Provider) Resolve(context.Context, string, string) (intent.Revision, error)
```

## 2. File and digest contract

Paths are exactly one of:

```text
docs/intent/records/business-decisions/<BD-id>/<four-digit-revision>.md
docs/intent/records/requirements/<BR-id>/<four-digit-revision>.md
docs/intent/records/change-intents/<CI-id>/<four-digit-revision>.md
```

IDs match `^(BD|BR|CI)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$` and the prefix/path kind must match.
Revisions start at `0001`; files are never replaced, and every successor contains an exact
`supersedes` reference to its predecessor. Paths are slash-separated, clean, relative, and contain
no traversal, backslash, NUL, or normalization ambiguity.

Each UTF-8 Markdown file starts with exactly:

```text
```proofbound-json
<one compact JSON object>
```
```

The object has exactly the canonical fields and schema for its record kind in the family SPEC. The
field order shown there is required for native artifacts. No BOM, indentation, duplicate/unknown
field, null, trailing JSON, or second metadata block is accepted. The remaining Markdown body is
opaque.

Native `artifact_sha256` solves self-reference by hashing the complete file after replacing only the
64 lowercase hexadecimal value of that metadata field with 64 ASCII zeroes. The declared digest
must equal that value. Core separately hashes the normalized payload after the real artifact digest
is restored.

## 3. Lifecycle and lineage

Business statuses: `proposed`, `accepted`, `superseded`, `withdrawn`; requirement statuses:
`proposed`, `active`, `superseded`, `retired`; intent statuses: `proposed`, `accepted`, `superseded`,
`withdrawn`. A successor revision may keep status or advance, but cannot reactivate a superseded,
retired, or withdrawn lineage. Every earlier accepted revision remains a distinct file.

Obligation IDs match `^O-[A-Z0-9][A-Z0-9-]{0,31}$`, are sorted and unique, and never disappear.
Statement changes require a new ID; retired IDs retain their last statement with state `retired`.

## 4. Hostile vectors

Fixtures under `testdata/` are contract artifacts. `valid/` pins one record of each kind and exact
normalized payload bytes. `hostile/` covers malformed UTF-8, duplicate and unknown fields, nulls,
digest mismatch, traversal, ID/path mismatch, invalid lifecycle, missing predecessor, dangling
relation, obligation removal/reuse, and canonicalization instability.

## 5. Invariants and test derivation

| Invariant | Statement | Proving test |
|---|---|---|
| REC-INV-1 | Only explicit committed-tree artifacts are read; working bytes cannot enter | records_test.go::TestCommittedReaderBoundary |
| REC-INV-2 | Path, UTF-8, metadata shape, order, and digest fail closed | records_test.go::TestHostileFixturesFailClosed |
| REC-INV-3 | Valid BD, BR, and CI vectors map to exact canonical payload bytes | records_test.go::TestValidVectors |
| REC-INV-4 | Same bytes re-ingest identically; a new immutable revision is distinct | records_test.go::TestRevisionIdentity |
| REC-INV-5 | Lifecycle and predecessor links are append-only and exact | records_test.go::TestLifecycleTransitions |
| REC-INV-6 | Obligation IDs cannot disappear, be reused, or change meaning | records_test.go::TestObligationLineage |
| REC-INV-7 | Resolve uses the named commit tree, never the working tree | records_test.go::TestResolveUsesCommitTree |
| REC-INV-8 | Missing, malformed, withdrawn, and ambiguous resolutions fail closed | records_test.go::TestResolveFailsClosed |

