# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-22

## Resume — 2026-09-22

P6 Task 9’s independent final consolidation round is complete in commit `52d41ca`. The frozen
implementation remains `e4c8e77407699f7e089d5c1a2b3ce58df5871fbf`; the round-C verdict is
ACCEPTABLE, the final census is 327/327 closed with zero unclassified rows, and P1–P5 are annotated
deep-complete in `ROADMAP.md`. `proofbound verify` passed. The final repository gate passed
repository checks, Go build, and Go tests, while kernel lint remains blocked by the known
Go-1.26-built golangci-lint versus Go 1.27 compatibility mismatch; it is not claimed as passed.

Preserve the two untracked user artifacts below; they remain excluded from all commits.

## Branch and repository status

- Branch: `main`, tracking `origin/main`; the Task 9 consolidation packet is at `52d41ca` before
  this state checkpoint.
- The only worktree entries are the preserved
  untracked `docs/plans/P5-intent-provenance-plan.md` (SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`) and
  `kernel/cmd/p6dbprobe/main.go` (SHA-256
  `f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0`).
- No PostgreSQL server, mutation worker, task-local database, or repository session fragment was
  left active. Existing `/tmp` build/mutation caches are disposable and not evidence.

## Completed

- P6 Task 8 is accepted at `e4c8e77`: all 16 production packages have exact tree objects,
  complete calibrated mutation counts totaling 1,584 killed, 0 invalid, 0 survived, and a
  committed ACCEPTABLE current-code non-author verdict. `internal/specfirst` remains the explicit
  test-only exclusion.
- C3 is closed. Task 9 is closed by the ACCEPTABLE round-C verdict
  docs/verification/verdicts/p6-consolidation-roundC-20260922.md; P5 and P6 Tasks 0–8 remain
  complete, and Task 7 remains explicitly skipped under ratification.
- The final evidence packet is docs/verification/p6-task9-consolidation.md. The final census
  generation input is `13503c7c96bad63ce1b95405ee8a83f09ccdab5a`; its 327 rows are all closed.
- The 2026-09-18 acceptance checkpoint recorded the frozen CLI 234/234/0/0 rerun, projections
  404/404/0/0 rerun, exact matrix, verdict, refreshed C6 coverage, and negative-test coverage.
- The final gate passed all repository checks, the Task 8 acceptance checker and hostile tests,
  package-universe checks, Go build, and Go tests. The expected `index stale; run make index`
  text came only from the intentional negative control. Direct index and invariant checks passed.
- The Task 9 close also corrected the strict `vera.verdict.v1` metadata boundary by keeping its
  front matter exact and moving reviewer metadata into Markdown body text consumed only by the
  package-acceptance checker. `proofbound verify` then passed against the committed verdict set.

## Open, blocked, and unverified

- Kernel lint remains blocked: golangci-lint 2.13.2 was built with Go 1.26.7 and panicked on Go
  1.27 source. This is an environment/toolchain blocker, not a reported source lint failure.
- Final closeout commit `b9d816f` was authenticated-pushed to `origin/main`. The dated bundle
  `/home/thamm/Backups/proofbound-20260922T134451Z.bundle` was bundle-verified at that tip with
  SHA-256 `05e7bbd62f867fbe5dbb6b9ebf51e071f0e426e557407cd6f92a090e78970e88`. The two preserved
  untracked artifacts remain untouched.

## Institutionalized improvement

At EOD, if preserved user-owned untracked artifacts trip commit cadence, make a documentation-only
journal/state checkpoint, rerun the gate, and preserve the artifacts. Do not delete, add, or absorb
ownership-unclear files. Continue recording only complete calibrated mutation summaries.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Do not add P7+ capability, widen the provider boundary, or alter frozen wire identities without
  new authorization. Preserve the two untracked user artifacts.
