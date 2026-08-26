# internal/connector/github — SPEC

The P3 v1 connector reads a bounded, read-only slice of one GitHub organization and an explicit
repository allowlist. It emits `github.workflow_run` and `github.deployment` events with upstream
IDs qualified by repository. Authentication is transport-only; tokens never enter payloads,
cursors, logs, or errors. The API is injected so deterministic HTTP fixtures can test pagination,
failure, and redaction before a live smoke test.

Invariants:

1. The organization and repository allowlist are required and repository names are validated.
2. Every upstream workflow/deployment ID is qualified by owner, repository, and record type.
3. Full normalized payloads carry upstream identity, status, upstream timestamps, and URL; sync
   freshness is carried by the sync cursor rather than changing event content on every replay.
4. Sync uses the ledger appender, so replay is idempotent and content revisions remain visible.
5. API errors stop the sync and return progress; partial results are never reported complete.
6. Authorization is sent only as a request header and is never persisted or logged.
7. The HTTP client accepts only HTTP(S) base URLs and non-2xx responses fail closed.
8. The connector is bounded to at most 100 records per upstream collection in v1.

| Invariant | Harm | Proving test |
|---|---|---|
| GITHUB-INV-1 | unsafe repository configuration enters the connector | github_test.go::TestNewRejectsUnsafeRepository |
| GITHUB-INV-2 | external records lose repository-qualified identity | github_test.go::TestSyncEmitsQualifiedWorkflowAndDeploymentEvents |
| GITHUB-INV-3 | transport query or authorization is malformed or leaked | github_test.go::TestHTTPClientPreservesQueryAndUsesHeaderAuth |
| GITHUB-INV-4 | replay or changed upstream content has correct ledger identity semantics | github_test.go::TestSyncChangedUpstreamRecordCreatesRevision |
| GITHUB-INV-5 | collection limits cannot exceed the v1 bound or be non-positive | github_test.go::TestHTTPClientClampsCollectionLimit |
| GITHUB-INV-6 | API records are emitted with normalized upstream fields | github_test.go::TestSyncEmitsQualifiedWorkflowAndDeploymentEvents |
| GITHUB-INV-7 | malformed repository configuration fails before sync | github_test.go::TestNewRejectsUnsafeRepository |
| GITHUB-INV-8 | authorization remains transport-only | github_test.go::TestHTTPClientPreservesQueryAndUsesHeaderAuth |
