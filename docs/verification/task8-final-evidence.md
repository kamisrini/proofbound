# Task 8 — author-side evidence

**Frozen implementation:** `d91bd63` (`feat: add proof-bound weekly report`)
**Date:** 2026-08-26
**Acceptance status:** pending independent Law 9 verification

The first independent review returned `NEEDS_WORK` on a MED finding: inner joins could
silently omit a projection row whose event proof was missing. The remediation changes
the report queries to left joins and fails closed with an explicit missing-proof error;
`TestReportWeek_FailsClosedWhenProofEventIsMissing` covers the route. The amended review
target is the follow-up commit after this evidence update.

## Scope

Task 8 implements `vera report week`. The report reads the projection views for the
seven-day event window, renders commit/check/session summaries, retains event proof IDs
on every data entry, and marks ledger commits absent from current Git reachability as
`[superseded]`. Git reachability uses a lightweight `rev-list` query and also includes a
detached `HEAD` when it resolves to a commit.

## Evidence run

- `go test ./... -count=1` — PASS
- `golangci-lint run ./...` — PASS (`0 issues`)
- `DATABASE_URL=postgres://postgres:postgres@127.0.0.1:55433/vera?sslmode=disable go test -tags=integration ./internal/projections -count=1` — PASS
- bare `make check` — PASS (exit 0; known read-only `docs/decisions/INDEX.md` stale diagnostic)
- `bash scripts/tests/mutants.test.sh` — PASS
- real repository `vera report week` run — PASS; rendered ledger-backed commits with proof IDs and a superseded fixture
- PostgreSQL proof-loss regression test — PASS; missing event proof fails closed

## Mutation limitation

The integration-tagged package sweep was not accepted as evidence. Its calibration
reported the neutral baseline as killed because the full integration baseline exceeded
the harness's 30-second ceiling after cache initialization. The retained log showed a
stale synthetic lethal calibration failure, so no count was inferred from it. The
untagged sweep produced database-logic survivors because integration tests were absent;
that result was also not used as acceptance evidence.

## Independent verifier checklist

Review frozen commit `d91bd63` and this evidence against:

1. event-ID proof binding for commits, checks, and sessions;
2. inclusive/exclusive seven-day boundary behavior and event-time filtering;
3. retention and `[superseded]` marking after rewrite, branch deletion, and detached HEAD;
4. malformed projection JSON, missing proof rows, and writer/database failures;
5. whether the mutation calibration limitation is mechanism debt or a justified exception.

The author has not issued an acceptance verdict.
