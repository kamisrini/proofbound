# internal/projections — SPEC

Task 6 contract for rebuildable, ledger-derived projections and the `vera verify` seam.
Projection tables are owned by this package and are never added to ledger migrations.

## 1. Boundary

The ledger is the sole source of truth. Projectors consume `store.Record` in ascending ledger
`seq` order and write only derived tables. They never import pgx or open a database; all database
access goes through `store.Store` and `store.Tx`.

Task 6 materializes `commits_view` and `checks_view`. `sessions_view` and `reviews_view` are
created as empty, versioned derived tables so later connectors have a stable destination; their
reducers are intentionally deferred to Tasks 7 and the verdict-ingestion work. Until those
reducers exist, a session or review event fails closed rather than being silently discarded.
Task 7 owns `sync sessions`; Task 8 owns `vera report week`. Task 6's `sync all` therefore means
all connectors implemented at this point: Git and checks.

## 2. Public API

```go
type Projector struct{}
func New() *Projector
func (p *Projector) Ensure(context.Context, *store.Store) error
func (p *Projector) Apply(context.Context, *store.Store) error
func (p *Projector) Rebuild(context.Context, *store.Store) error
func (p *Projector) Snapshot(context.Context, *store.Store) (Snapshot, error)
func CompareSnapshots(Snapshot, Snapshot) error
```

`Apply` consumes events after a derived `projection_meta.last_seq` checkpoint. Row updates and the
checkpoint advance are one transaction. `Rebuild` drops only projection tables, recreates them,
replays every ledger event, and never changes ledger rows. Revisions with the same natural key
fold last-write-wins by increasing `seq`.

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
10. **P-INV-10 — Empty future views are deterministic.** Session and review views exist and snapshot as empty until their connectors land; their events fail closed while deferred.
11. **P-INV-11 — Projection metadata is unique and versioned.** Exactly one named metadata row owns the checkpoint for projection version 1.

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
| P-INV-10 | Future views are present and empty; deferred events fail closed | projection_test.go::TestEnsure_CreatesFutureViews |
| P-INV-11 | Metadata has one versioned checkpoint row | projection_test.go::TestEnsure_MetadataIsUniqueAndVersioned |
