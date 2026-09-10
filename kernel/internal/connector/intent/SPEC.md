# Intent provider family contract

## 1. Purpose and boundary

This package owns Proofbound's provider-independent intent payload model and ingestion coordinator.
Providers observe immutable source revisions and normalize them into business decisions,
requirements, or change intents. The ledger payload—not a provider file format—is the semantic
contract consumed by projections, reports, gates, and verdicts.

P5 production providers are exactly `records` and `specdir`. A synthetic provider may exist only in
tests and must not be configurable by the CLI. Missing source segments remain missing; no provider
may fabricate a business decision, requirement, intent, relation, authority, or obligation.

## 2. Interface lock

```go
const Version = "intent/1"

type Revision struct {
    Source         core.Source
    Kind           core.Kind
    NativeID       string
    ArtifactPath   string
    ArtifactSHA256 string
    ObservedAt     time.Time
    Payload        json.RawMessage
}

type Provider interface {
    Name() string
    Revisions(context.Context) ([]Revision, error)
}

type Appender interface {
    Append(context.Context, core.Event) (store.Record, bool, error)
}

type Deps struct { Providers []Provider; IDs *core.IDGenerator; Logger *slog.Logger }
type Result struct { Listed, Appended, Existing int; ByProvider map[string]ProviderResult; Cursor json.RawMessage }
func New(*Deps) (*Connector, error)
func (c *Connector) Sync(context.Context, string, Appender) (Result, error)
func ValidateRevision(Revision) error
```

`Sync` accepts one configured provider name or `all`. It validates the entire batch, then orders it
dependency-topologically (BD before the BR it authorizes, BR before the CI that targets it, and
predecessor before successor), with source, kind, record id, artifact digest, and path as stable
tie-breakers. Cross-provider ordering and malformed input therefore cannot create partial trust.

## 3. Canonical payload contract

All payloads are strict JSON: UTF-8, no duplicate/unknown/null/missing fields, and stable RFC 8785
canonicalization. Every field required to rebuild relation and obligation meaning is embedded.

```go
type Reference struct {
    RecordKind    string   `json:"record_kind"`
    Source        string   `json:"source"`
    RecordID      string   `json:"record_id"`
    ArtifactSHA256 string  `json:"artifact_sha256"`
    Relation      string   `json:"relation"`
    ObligationIDs []string `json:"obligation_ids,omitempty"`
}
type Obligation struct { ID string `json:"id"`; Statement string `json:"statement"`; State string `json:"state"` }
```

Business decision payload (`proofbound.business-decision.v1`) contains, in canonical field order:
`schema`, `decision_id`, `status`, `outcome`, `declared_owner`, `declared_approver`, `decided_at`,
`supersedes`, `artifact_path`, `artifact_sha256`.

Requirement payload (`proofbound.requirement.v1`) contains: `schema`, `requirement_id`, `status`,
`authorized_by`, `declared_owner`, `obligations`, `supersedes`, `artifact_path`, `artifact_sha256`.
`authorized_by` may be empty only for a provider that genuinely has no authorization segment; this
renders `authorization: undeclared`.

Change-intent payload (`proofbound.change-intent.v1`) contains: `schema`, `intent_id`, `status`,
`declared_sponsor`, `targets`, `constrained_by`, `supersedes`, `artifact_path`, `artifact_sha256`.

Closed statuses and relations are those ratified in
`VD-p5-intent-provenance-2026-09-10`. References always bind source + kind + id + exact artifact
digest. Requirement targets carry one or more sorted unique obligation ids; other references carry
none. Provider-scoped record identity is `(source, record_id)`.
The closed relation vocabulary is `authorizes`, `implements`, `modifies`, `repairs`, `retires`,
`supersedes`, `constrained_by`, `reviews`, and `evaluates`.

## 4. Two digests and point-in-time resolution

`artifact_sha256` is the SHA-256 of the provider's immutable revision bytes: normalized-self-digest
bytes for the native provider and exact committed file bytes for `specdir`. `content_sha` is computed
independently by core over canonical normalized payload JSON. Removing either comparison is a defect.

Committed-file providers resolve at the commit's own tree. A later API provider would record a
snapshot observed at ingest, but no API provider ships in P5. A resolution returns the exact digest;
an upstream mutation between resolution and ingest fails before append.

## 5. Lifecycle and replay

Revisions are append-only. A successor may supersede exact earlier revisions, including across
providers. Active obligation IDs cannot disappear or change statement/state in place: meaning change
uses a new ID and retirement leaves a tombstone. All ids, states, references, statements, paths, and
digests are carried in payloads, so projection deletion and ledger replay are source-independent.

## 6. Non-goals

- No mutable API provider, authenticated approval, inferred applicability, graph store, or fabricated
  chain segment.
- No schema change to accommodate the deliberately unmappable foreign fixture. Encountering it is a
  phase STOP result.
- No prose-body ingestion beyond its source artifact digest.

## 7. Invariants and test derivation

| Invariant | Statement | Proving test |
|---|---|---|
| INT-INV-1 | Only registered canonical kinds and `intent.<provider>` sources pass | intent_test.go::TestValidateRevisionRegistry |
| INT-INV-2 | Artifact digest and canonical payload digest are independent and both enforced | intent_test.go::TestTwoDigestsAreIndependentlyRequired |
| INT-INV-3 | Every reference binds source, kind, id, digest, relation, and valid obligation ids | intent_test.go::TestReferenceValidationFailsClosed |
| INT-INV-4 | One batch is completely validated and deterministically ordered before append | intent_test.go::TestSyncValidatesBeforeAppendAndOrdersProviders |
| INT-INV-5 | Re-ingest appends zero and one changed immutable revision appends one | intent_test.go::TestSyncIdempotenceAndRevision |
| INT-INV-6 | Provider-scoped identity rejects cross-provider collision and prefix spoofing | conformance_test.go::TestProviderScopedIdentityAndPrefixSpoofing |
| INT-INV-7 | Cross-provider supersedes retains both exact lineages | conformance_test.go::TestCrossProviderSupersedes |
| INT-INV-8 | Mutation between resolution and ingest fails on digest mismatch | conformance_test.go::TestUpstreamMutationDigestMismatch |
| INT-INV-9 | Provider ordering ingests revisions before commit consumers | conformance_test.go::TestCrossProviderSyncOrdering |
| INT-INV-10 | Payloads reconstruct relation and obligation meaning without source files | conformance_test.go::TestReplaySufficiency |
| INT-INV-11 | The unmappable foreign fixture returns the distinguished STOP error | conformance_test.go::TestUnmappableFixtureStops |
| INT-INV-12 | No synthetic provider can be selected by a production constructor | intent_test.go::TestSyntheticProviderIsTestOnly |
