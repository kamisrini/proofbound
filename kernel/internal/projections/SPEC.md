# internal/projections — SPEC

Task 6 contract for rebuildable, ledger-derived projections and the `proofbound verify` seam.
Projection tables are owned by this package and are never added to ledger migrations.

## 1. Boundary

The ledger is the sole source of truth. Projectors consume `store.Record` in ascending ledger
`seq` order and write only derived tables. They never import pgx or open a database; all database
access goes through `store.Store` and `store.Tx`.

Task 6 materializes `commits_view` and `checks_view`. Task 7 adds the best-effort sessions
connector and materializes `sessions_view`; `reviews_view` remains an empty stable destination
until verdict ingestion lands. Review events fail closed rather than being silently discarded.
Task 7 owns `sync sessions`; Task 8 owns `proofbound report week`; P3 owns the GitHub delivery view and
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
func (p *Projector) ReportIntent(context.Context, *store.Store, string, time.Time, io.Writer) error
func (p *Projector) ReportRequirement(context.Context, *store.Store, string, io.Writer) error
func (p *Projector) CheckIntent(context.Context, *store.Store, string, io.Writer) error
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

P5 adds `business_decisions_view`, `requirements_view`, `requirement_obligations_view`,
`change_intents_view`, `intent_targets_view`, `commit_intents_view`, `obligation_verdicts_view`, and
`requirement_reviews_view`. Revision tables key on provider source, record ID, and artifact digest;
relation tables retain both exact endpoints. `commits_view.intent_refs` is canonical JSON and old
git/1 payloads remain readable as no-claim commits. Every table row carries its originating event ID
and sequence.

Intent reports render the exact CI revision, each exact targeted requirement/obligation, declared
authority or `authorization: undeclared`, claiming commits, obligation verdict/evidence components,
spec-review component, exact-commit deployments/freshness, and event proof. Component states remain
distinct. A missing/non-verifiable requirement review caps a chain below green but does not block
authoring or targeting. `CheckIntent` validates only commits with explicit intent references and
fails closed on dangling exact revisions or obligations.

Review reduction dispatches by schema. `vera.verdict.v1` retains its existing finding-only meaning.
`proofbound.obligation-verdict.v2` must match one observed commit claim, its exact CI revision, every
exact targeted BR obligation, and only evidence event IDs whose ledger sequence is at or before the
verdict. Aggregate `ACCEPTABLE` requires complete `SATISFIED` outcomes. A
`proofbound.requirement-review.v1` must bind an existing exact requirement revision, cover every
obligation exactly once, and declare a reviewer different from the revision's declared owner.

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
21. **P-INV-21 — Intent revisions retain exact proof.** Every BD, BR, CI, obligation, target, and
    commit claim row retains provider, exact artifact digest, event ID, and ledger sequence.
22. **P-INV-22 — Intent references fail closed.** Dangling record revisions, obligation IDs, or
    commit claims abort the projection transaction.
23. **P-INV-23 — Intent replay is deterministic.** Incremental and from-genesis projection row sets
    match across every P5 table without consulting source files.
24. **P-INV-24 — Chain state is component-honest.** Missing, contradicted, inconclusive,
    superseded, deployed-unverified, and unreviewed/non-verifiable spec states remain explicit; no
    aggregate is green while one component is not.
25. **P-INV-25 — Reports are proof-bearing.** Every rendered record, commit, verdict, evidence, and
    deployment component includes event ID and sequence; missing proof fails closed.
26. **P-INV-26 — V1 review compatibility is frozen.** Existing v1 payloads continue to populate
    only `reviews_view` and never acquire obligation semantics.
27. **P-INV-27 — V2 verdict chains bind completely.** Commit, CI, BR, target obligation, aggregate
    status, and evidence sequence are all validated before any verdict row commits.
28. **P-INV-28 — Requirement review is exact and independent-by-declaration.** Wrong revisions,
    absent obligations, unknown outcomes, incomplete coverage, and equal owner/reviewer fail closed.
29. **P-INV-29 — Missing satisfaction is derived.** No stored verdict row represents UNVERIFIED;
    reports derive it when an exact targeted obligation has no applicable v2 outcome.
30. **P-INV-30 — Deployments join only by exact commit.** A deployment cannot attach through an
    intent ID, branch, or neighbouring revision; distinct commits and environments remain distinct.
31. **P-INV-31 — Deployment gaps are explicit.** Missing and stale deployment evidence render as
    components, and deployment without complete satisfaction is `DEPLOYED_UNVERIFIED`.
32. **P-INV-32 — Deployment proof is bounded.** Each joined deployment retains event/seq proof;
    missing proof or a future freshness timestamp fails closed.

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
| P-INV-21 | All intent projection rows retain exact revision and ledger proof | intent_test.go::TestIntentRowsRetainExactRevisionAndProof |
| P-INV-22 | Dangling intent relations and claims roll back | intent_test.go::TestIntentProjectionRejectsDanglingProof |
| P-INV-23 | Incremental and rebuilt P5 row sets match | intent_test.go::TestIntentProjectionRebuildMatchesIncremental |
| P-INV-24 | Report states and spec-review caps never hide a gap | intent_test.go::TestIntentReportRendersComponentStates |
| P-INV-25 | Intent reports carry proof and reject missing proof | intent_test.go::TestIntentReportProofFailsClosed |
| P-INV-26 | V1 bytes retain finding-only projection meaning | review_integration_test.go::TestApply_ReviewFindingsRetainProofAndRevision |
| P-INV-27 | V2 exact chain and evidence sequence fail closed | intent_review_test.go::TestObligationVerdictProjectionValidation |
| P-INV-28 | Requirement-review revision, coverage, and independence fail closed | intent_review_test.go::TestRequirementReviewProjectionValidation |
| P-INV-29 | UNVERIFIED is derived rather than stored | intent_review_test.go::TestUnverifiedIsDerived |
| P-INV-30 | Exact commit joins preserve multiple environments and revisions | intent_review_test.go::TestIntentDeploymentJoinsExactCommit |
| P-INV-31 | Missing, stale, and deployed-unverified states remain explicit | intent_review_test.go::TestIntentDeploymentStates |
| P-INV-32 | Deployment proof and freshness fail closed | intent_review_test.go::TestIntentDeploymentProofFailsClosed |
