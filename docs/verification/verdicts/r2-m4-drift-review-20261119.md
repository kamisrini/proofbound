---
schema: vera.verdict.v1
verdict_id: r2-m4-drift-review-20261119
status: NEEDS_WORK
reviewed_commit: 46dbf2748a84ae25d0c51b1c08e7cbd75e44b56a
findings:
  - finding_id: r2-m4-drift-review-20261119-F1
    severity: HIGH
    defect_commit: 46dbf2748a84ae25d0c51b1c08e7cbd75e44b56a
    summary: "current revision = highest seq is wrong after a revert; a never-verified promise is reported CURRENT"
artifact_path: docs/verification/verdicts/r2-m4-drift-review-20261119.md
artifact_sha: 8e2a6376bee034231ad183b088b0cc0cbff9b565b7c871bc2d08b08649de2855
---

# R2 M4 drift: non-author review (connector tier)

## Provenance
- Reviewer: independent non-author (Build Law 9), one connector-tier review under Law 11. Did not write 46dbf27.
- Read: `docs/plans/R2-m4-drift-spec.md`, `kernel/internal/cli/drift.go`, the `cli.go` route, `drift_test.go`, `sdd/drift_e2e_test.go`, the projection and store SQL (`projections.go` change-intent, target and verdict handlers; `store.go` event insert), and BR-sdd-warranty-i9miv7 0003 and 0004.
- Ran: a text diff of 0003 and 0004, and a recompute of the 0004 self-digest. No tests were run and no worktree was created, because the main finding follows from reading the SQL. The main checkout is untouched.

## Findings

### F1 (HIGH): after a revert, "current" is stale and a never-verified promise reads CURRENT
`drift.go` takes the current revision to be the one with the highest `seq` in `change_intents_view`. But `store.go:262` inserts events with `ON CONFLICT(source,native_id,content_sha) DO NOTHING`. A revert re-emits a payload whose bytes are identical to an earlier revision, so it is deduped and gets no new seq. `change_intents_view.seq` is therefore the seq of the first time that digest was ledgered, and a revert never raises it.

Failure scenario, in the dangerous direction:
1. A is ledgered at seq 10.
2. B is ledgered at seq 20 and verified (warranty on B).
3. The change is reverted to A and synced. Event A is deduped, so A keeps seq 10 and current stays B.
4. drift prints `state=CURRENT` against B's warranty, but the change now promises A, which was never verified.

The mirror case gives a false STALE: warranty on A, edit to B, revert to A. Current stays B, so it prints STALE with a diff against B, although the tree promises exactly what was verified.

Why the tests miss it: the unit case "edit reverted" feeds a fresh `{"v1-again", 40}` revision, which the real ledger never produces. The e2e test has no revert step. The same issue applies to archive then unarchive.

Fix direction: stop using the ledger seq as the current-revision signal. Options are to resolve current from the commit graph (the intent revision at HEAD of the synced history, or the latest git-ingested commit's `commit_intents_view`/SDD state), or to record a monotonic observation for each sync. Then add an e2e revert step (verify on A, change to B, revert to A: expect CURRENT) and one for B verified then reverted to A (expect STALE). Also amend the spec text "highest ledger seq" and O-13's "latest ledgered revision" in a new BR revision, since the claim as worded is exactly the defective one.

### Low (one line each)
- L1: Spec choice, not a bug. The warranty omits the VERIFIABLE requirement-review that the readiness gate requires (O-8). It is defensible because drift is information only, but state it in the spec's non-goals so CURRENT is not read as "gate-warranted".
- L2: `warranty.seq` is the max over a verdict's rows, so a re-ledgered verdict is ordered by its latest outcome row. This is harmless.
- L3: The e2e `appended=[1-9]` check proves the archive step creates revisions. That part is good.

## Questions answered
- Same verdict spanning several requirements, and several verdicts per commit: the SQL groups by (verdict, commit, intent, digest) with `bool_and(outcome='SATISFIED')`, which is correct. The projection also enforces exact coverage of every target (`len(targets)==len(v.Obligations)`), so a partial verdict cannot be ingested.
- Missing join row shrinking a promise: ingest rejects a change intent whose target obligation does not exist and is not active (dangling check in `projections.go`). The join therefore cannot silently drop a target, and rows are keyed by immutable digest. No problem.
- Duplicate obligation ids across requirements: statements are sorted and joined with " | ", which is deterministic, and a duplicate cannot hide a difference. Fine.
- BR 0004: the diff against 0003 adds only O-13, correctly supersedes 0003 (`3c30b498...`), and leaves O-1 through O-12 unchanged. The self-digest `83d98dbe...4024` recomputes correctly, with the last `"artifact_sha256"` value zeroed. No issue.
- CLI-INV-21 and the e2e flow (verify, reword gives STALE with the exact diff, re-verify gives CURRENT, archive still CURRENT with new revisions) prove their stated steps, except for revert (F1).

## Verdict
NEEDS_WORK. One HIGH: a stale warranty can be reported CURRENT after a revert, so the "current revision" rule needs a source that a re-ledgered old digest cannot defeat. Everything else is sound.
