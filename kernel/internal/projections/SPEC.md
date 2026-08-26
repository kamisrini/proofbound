# internal/projections — SPEC

Task 6 contract for rebuildable, ledger-derived projections and the `vera verify` seam.
Projection tables are owned by this package and are never added to ledger migrations.

## 1. Boundary

The ledger is the sole source of truth. Projectors consume `store.Record` in ascending ledger
`seq` order and write only derived tables. They never import pgx or open a database; all database
access goes through `store.Store` and `store.Tx`.

Task 6 materializes `commits_view` and `checks_view`. Task 7 adds the best-effort sessions
connector and materializes `sessions_view`; `reviews_view` remains an empty stable destination
until verdict ingestion lands. Review events fail closed rather than being silently discarded.
Task 7 owns `sync sessions`; Task 8 owns `vera report week`; P3 owns the GitHub delivery view and
`ReportGitHub`.

## 2. Public API

```go
type Projector struct{}
func New() *Projector
func (p *Projector) Ensure(context.Context, *store.Store) error
func (p *Projector) Apply(context.Context, *store.Store) error
func (p *Projector) Rebuild(context.Context, *store.Store) error
func (p *Projector) Snapshot(context.Context, *store.Store) (Snapshot, error)
func CompareSnapshots(Snapshot, Snapshot) error
func (p *Projector) ReportWeek(context.Context, *store.Store, time.Time, map[string]bool, io.Writer) error
func (p *Projector) ReportGitHub(context.Context, *store.Store, time.Time, io.Writer) error
```

`Apply` consumes events after a derived `projection_meta.last_seq` checkpoint. Row updates and the
checkpoint advance are one transaction. `Rebuild` drops only projection tables, recreates them,
replays every ledger event, and never changes ledger rows. Revisions with the same natural key
fold last-write-wins by increasing `seq`.

`ReportGitHub` groups `github_delivery_view` rows by `(repository, commit_sha)`. A missing workflow
or deployment is rendered as `missing`; a workflow with a non-success completed conclusion is
`failed`; a non-completed workflow is `running`. Deployment status is `observed` because the v1
connector does not ingest deployment status. Freshness is the oldest `freshness_at` in the group:
`fresh` means no older than 24 hours at report time, otherwise `stale`. Every rendered proof is
`event_id/seq`, and a missing event proof fails closed.

## 3. Derived schema

All tables use natural keys and retain `event_id` and `seq` as proof links. No serial or wall-clock
mutation columns are permitted. `files_touched`, `cited_decisions`, and `tool_versions` remain
canonical JSON columns.

## 4. Invariants

1. **P-INV-1 — Ledger order is projection order.** Reducers consume records in ascending `seq`.
2. **P-INV-2 — Revisions are last-write-wins.** A later event for one natural key replaces its row.
3. **P-INV-3 — Projection writes are resumable.** Row updates and `last_seq` advance commit atomically.
4. **P-INV-4 — Rebuild is ledger-preserving.** Rebuild changes only derived tables.
5. **P-INV-5 — Incremental and rebuilt row sets are identical.** Full-row canonical digests compare equal.
6. **P-INV-6 — Projection rows retain proof identity.** Every row stores its originating event ID and seq.
7. **P-INV-7 — Malformed or unsupported events fail closed.** No partial projection transaction commits.
8. **P-INV-8 — Projection DDL is not ledger migration.** Derived tables are created only by this package.
9. **P-INV-9 — Snapshots are natural-key canonical multisets.** Database order and JSON formatting do not affect comparison.
10. **P-INV-10 — Review findings are deterministic.** Valid verdict findings retain their verdict and event proof; malformed review events fail closed.
11. **P-INV-11 — Projection metadata is unique and versioned.** Exactly one named metadata row owns the checkpoint for projection version 1.
12. **P-INV-12 — Session metadata is projected without content.** Session rows contain only the connector's bounded metadata and preserve the event proof links.
13. **P-INV-13 — Week report entries carry proof identity.** Every commit, check, and session entry renders its originating event ID.
14. **P-INV-14 — Unreachable commits are retained and marked superseded.** A commit absent from the supplied current reachability set is not omitted from the report.
15. **P-INV-15 — Missing proof rows fail closed.** A projection row whose event proof is absent causes the week report to fail rather than silently disappearing.
16. **P-INV-16 — GitHub events are normalized by kind.** Workflow and deployment payloads validate their repository, commit, identity, status, and timestamps before materialization.
17. **P-INV-17 — GitHub revisions are last-write-wins.** A later event for one qualified upstream `native_id` replaces its delivery row.
18. **P-INV-18 — GitHub delivery rows retain proof.** Each normalized row stores its source event ID and ledger sequence.
19. **P-INV-19 — GitHub report semantics are explicit.** Missing, failed, observed, and stale states are rendered rather than inferred as success.
20. **P-INV-20 — GitHub report proof is fail-closed.** A delivery row whose event proof is absent or whose freshness is in the future causes reporting to fail.

## 5. Proving table

| Invariant | Statement | Proving test |
|---|---|---|
| P-INV-1 | Records are reduced by ascending seq | projection_test.go::TestApply_UsesLedgerOrder |
| P-INV-2 | Newer revisions replace older rows | projection_test.go::TestApply_RevisionLastWriteWins |
| P-INV-3 | Checkpoint and rows commit atomically | projection_test.go::TestApply_MalformedPayloadRollsBack |
| P-INV-4 | Rebuild leaves ledger unchanged | projection_test.go::TestRebuild_DoesNotModifyLedger |
| P-INV-5 | Rebuild matches incremental canonical row sets | projection_test.go::TestRebuild_RowSetMatchesIncremental |
| P-INV-6 | Rows retain event ID and seq | projection_test.go::TestRows_RetainProofIdentity |
| P-INV-7 | Unsupported or malformed events fail closed | projection_test.go::TestApply_RejectsUnsupportedEvent |
| P-INV-8 | Projection DDL is absent from ledger migration | projection_test.go::TestDDL_IsNotLedgerMigration |
| P-INV-9 | Snapshot comparison ignores row order and formatting | projection_test.go::TestSnapshot_CanonicalMultisets |
| P-INV-10 | Review verdict events validate and materialize finding proof rows | review_integration_test.go::TestApply_ReviewFindingsRetainProofAndRevision |
| P-INV-11 | Metadata has one versioned checkpoint row | projection_test.go::TestEnsure_MetadataIsUniqueAndVersioned |
| P-INV-12 | Session metadata is materialized with proof identity | projection_test.go::TestApply_Session |
| P-INV-13 | Week report entries render their originating event IDs | report_test.go::TestRenderWeekReport_ProofAndSupersededFixture |
| P-INV-14 | Unreachable commit fixture is retained and marked superseded | report_test.go::TestRenderWeekReport_ProofAndSupersededFixture |
| P-INV-15 | Missing projection proof fails closed | report_integration_test.go::TestReportWeek_FailsClosedWhenProofEventIsMissing |
| P-INV-16 | GitHub workflow/deployment rows validate and normalize fields | projection_test.go::TestGitHubPayloadValidationRejectsMalformedFields |
| P-INV-17 | GitHub delivery rows preserve the newest qualified event | projection_test.go::TestApply_GitHubDeliveryRetainsNormalizedFieldsAndProof |
| P-INV-18 | GitHub rows retain event ID and seq | projection_test.go::TestApply_GitHubDeliveryRetainsNormalizedFieldsAndProof |
| P-INV-19 | GitHub report renders missing, failed, observed, and stale states | report_test.go::TestRenderGitHubReport_StatesMissingFailedAndStaleExplicitly |
| P-INV-20 | GitHub report renders proof and rejects missing proof integration path | report_integration_test.go::TestReportGitHub_RendersJoinStatesFreshnessAndProof |
