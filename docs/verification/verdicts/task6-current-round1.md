# NEEDS_WORK

Frozen commit `b246ed5` is not acceptable under the Task 6 contract.

## HIGH findings

1. `vera verify` does not verify the latest witness.

The verifier only sets `found = true` when it sees any `checks/check.run` event:

[kernel/cmd/vera/main.go:231](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/cmd/vera/main.go:231)

It never identifies the latest spool run, compares its `run_id`, or ensures that the latest ledger event corresponds to the latest `make check-witnessed` run. A stale historical witness can therefore satisfy verification while the newest run is absent or different.

This violates the Task 6 DoD and creates a false “latest gate evidence exists” route.

2. Supported-but-unimplemented event kinds are silently discarded.

`session.observed` and `review.verdict` are declared supported:

[kernel/internal/projections/projections.go:134](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/projections/projections.go:134)

But `reduce` has no reducers for them and returns `nil`:

[kernel/internal/projections/projections.go:129](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/projections/projections.go:129)

The checkpoint then advances, permanently passing over those events. This directly violates P-INV-7 (“malformed or unsupported events fail closed”) and the SPEC’s claim that unsupported evidence is never silently discarded.

## MED findings

- The JSON decoder accepts malformed trailing bytes. After decoding the first value, any non-EOF error from the second decode is treated as success:

[kernel/internal/projections/projections.go:167](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/internal/projections/projections.go:167)

For example, a valid object followed by invalid garbage can be projected.

- Payload semantic validation is incomplete. Commit and check reducers validate only a few fields. They do not reject backward check timestamps, negative durations, invalid run IDs, invalid hashes, wrong schema/command, or missing scalar fields beyond the minimal checks. The upstream checks connector validates these, but the ledger projection boundary must fail closed for malformed ledger payloads independently.

- Snapshot canonicalization is not actually canonical for nested JSON columns. `files_touched`, `cited_decisions`, and `tool_versions` are scanned as database values and passed to `json.Marshal`; byte slices become base64 strings rather than parsed canonical JSON. Thus JSON formatting can affect the digest, contrary to P-INV-9.

- `projection_meta` has no uniqueness or primary-key constraint on its single checkpoint row. Corruption or duplicate rows can make checkpoint reads and updates ambiguous.

- The plan requires projection versioning, but the DDL creates no `projection_version` field or equivalent version mechanism.

## Scope inconsistencies

The Task 6 SPEC defers session and review reducers, while the P1 plan’s CLI/projection contract names:

- `sync sessions`
- `sync all`
- `sessions_view`
- `week_report`

The frozen CLI rejects `sync sessions`, and `sync all` only runs Git and checks:

[kernel/cmd/vera/main.go:172](/mnt/d/Users/thamm/Desktop/Projects/Vera/kernel/cmd/vera/main.go:172)

This may be valid sequencing if explicitly amended, but the current commit does not reconcile the scope claims.

## Invariant audit

- P-INV-1: Implementation checks increasing sequence; DB proof was unavailable.
- P-INV-2: Implementation uses sequence-based last-write-wins; the proving test is only an alias of the order test.
- P-INV-3: Reducer writes and checkpoint update share a transaction, but the test does not assert checkpoint rollback.
- P-INV-4: Rebuild drops only projection tables in code; the test checks only ledger row count, not complete ledger identity.
- P-INV-5: Implementation comparison exists; DB-backed proof was unavailable.
- P-INV-6: Code stores `event_id` and `seq`; DB-backed proof was unavailable.
- P-INV-7: Fails due to silently accepted session/review events and malformed trailing JSON.
- P-INV-8: No projection DDL was added to ledger migrations.
- P-INV-9: Comparison is multiset-based, but actual database-row canonicalization is incomplete.
- P-INV-10: DDL creates future tables; DB-backed proof was unavailable.

## Verification evidence

Passed:

- `GOCACHE=/tmp/vera-go-cache go test -race ./... -count=1`
- `make check`
- Non-integration command and projection package tests
- `git diff --check`
- Worktree remains clean

Unavailable or failed:

- DB-backed projection tests and `vera verify`: no `DATABASE_URL`, Docker, or available embedded PostgreSQL service.
- `go test -tags integration -race ...`: `cmd/vera` integration tests fail immediately with `DATABASE_URL is required`.
- No mutation sweep was run, per instruction.

The latest-witness defect and silent event-loss defect must be fixed before another acceptance review.
