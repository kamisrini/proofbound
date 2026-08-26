---
schema: vera.verdict.v1
verdict_id: task6-current-round2
status: NEEDS_WORK
reviewed_commit: abf4344b5ac1d1306ed9527e59820b4704d217a6
findings:
  - finding_id: task6-current-round2-finding
    severity: MED
artifact_path: docs/verification/verdicts/task6-current-round2.md
artifact_sha: 4eb65605fa2a5f850e268d09597ad823d4d699139a01b17cd686f1b3243e177f
---

# NEEDS_WORK

Frozen commit `abf4344` is not acceptable under the full Task 6 / P1 contract.

## Findings

### HIGH — P1 scope remains unreconciled

The implementation still omits required P1 review evidence:

- `review.verdict` is part of the ROADMAP P1 DoD and must expose the red-verdict → later-fix/check chain.
- The projections SPEC explicitly defers `reviews_view`.
- `sync sessions` is rejected as “not implemented,” and `sync all` runs only Git and checks.
- `vera report week` is absent from the CLI despite being specified by the plan.

Evidence:

- `ROADMAP.md:25-30`
- `docs/plans/P1-flight-recorder-plan.md:82-89`
- `kernel/cmd/vera/main.go:25,40-62`

Deferral may be valid for Tasks 7–8, but the frozen Task 6 work does not amend or reconcile the conflicting contract. As written, the P1 DoD cannot be satisfied and the resulting view cannot truthfully represent all required evidence.

### MED — Commit payload validation remains incomplete

`commitPayload.validate` checks only SHA, timestamp, and basic file-path properties:

[kernel/internal/projections/projections.go:163-172]

It accepts empty author/committer identities, empty subjects, malformed emails, nil/invalid cited-decision values, and other structurally invalid commit payloads. Since projection decoding is a ledger trust boundary, malformed commit evidence can still become durable derived state.

## Previously reported Round 1 findings

Closed in code:

- Latest witness is selected and compared with the latest ledger `check.run` by sequence.
- Deferred session/review events now fail closed.
- Trailing JSON is rejected.
- Nested JSON snapshot values are canonicalized.
- Projection metadata is unique and versioned.

## Invariant audit

- P-INV-1: Code enforces increasing sequence order; DB proof unavailable.
- P-INV-2: Last-write-wins is implemented with sequence comparison; DB proof unavailable.
- P-INV-3: Projection writes and checkpoint update share a transaction; rollback DB proof unavailable.
- P-INV-4: Rebuild drops projection tables only; complete DB ledger-identity proof unavailable.
- P-INV-5: Rebuild/incremental comparison exists, but the integration proving test could not run.
- P-INV-6: Rows retain `event_id` and `seq`; DB proof unavailable.
- P-INV-7: Unsupported/deferred and malformed events fail closed in code; DB rollback proof unavailable.
- P-INV-8: Projection DDL remains outside ledger migrations.
- P-INV-9: Snapshot comparison is multiset-based and canonicalizes nested JSON in code; DB proof unavailable.
- P-INV-10: Future views are created and deferred events fail closed.
- P-INV-11: Metadata has a primary key and version column.

The SPEC’s proving table uses `integration_test.go::TestRebuild_RowSetMatchesIncremental`, but the repository’s test is in `projection_test.go`; the cited proving cell is therefore inaccurate.

## Verification evidence

Passed:

- `GOCACHE=/tmp/vera-review-cache go test -race ./... -count=1`
- Bare `make check`
- `git diff --check`
- CLI and non-database projection tests
- Worktree remained clean

`make check` printed `index stale; run make index` but completed with `0 issues` and exit code 0.

Unavailable:

- DB-backed projection tests
- End-to-end `vera verify`
- Integration sync tests

Targeted integration execution failed because `DATABASE_URL` is unset:

```text
TestSyncChecksIngestsAndDeduplicates: DATABASE_URL is required
TestSyncChecksReportsMalformedWitness: DATABASE_URL is required
```

No mutation sweep was run, per instruction.

The Round 1 HIGH implementation defects are closed, but the unresolved P1 scope and incomplete commit validation require another remediation round before acceptance.
