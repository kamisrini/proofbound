# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-19

## Resume — 2026-09-19

P6 Task 8 package acceptance is complete. Frozen implementation:
`e4c8e77407699f7e089d5c1a2b3ce58df5871fbf`. Today’s durability closeout is complete: the factual
journal/state checkpoint was committed, the documented gate passed repository checks, the dated
bundle was clone-verified, and the authenticated push succeeded. Kernel lint remains blocked by
the known Go-1.26-built golangci-lint versus Go 1.27 compatibility mismatch; it is not claimed as
passed.

Tomorrow’s first action is Task 9’s independent final consolidation round. Do not claim P6
deep-complete until that round is committed. Preserve the two untracked user artifacts below.

## Branch and repository status

- Branch: `main`, tracking `origin/main`; HEAD at inspection was `054da9db9921dbc67f87cd1c13c03576ff0224a3`.
- No commits were made earlier on 2026-09-19. The only worktree entries were the pre-existing
  untracked `docs/plans/P5-intent-provenance-plan.md` (SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`) and
  `kernel/cmd/p6dbprobe/main.go` (SHA-256
  `f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0`).
- No PostgreSQL server, mutation worker, task-local database, or repository session fragment was
  active. Existing `/tmp` build/mutation caches are disposable and not evidence.

## Completed

- P6 Task 8 is accepted at `e4c8e77`: all 16 production packages have exact tree objects,
  complete calibrated mutation counts totaling 1,584 killed, 0 invalid, 0 survived, and a
  committed ACCEPTABLE current-code non-author verdict. `internal/specfirst` remains the explicit
  test-only exclusion.
- C3 is closed. Task 9 has not started. P5 and P6 Tasks 0–6 remain complete; Task 7 remains
  explicitly skipped under ratification.
- The 2026-09-18 acceptance checkpoint recorded the frozen CLI 234/234/0/0 rerun, projections
  404/404/0/0 rerun, exact matrix, verdict, refreshed C6 coverage, and negative-test coverage.
- Today’s gate reached and passed the Task 8 acceptance checker, all package-universe checks, and
  repository checks before stopping at commit cadence. After this checkpoint, the complete rerun
  passed cadence, repository checks, the 16-package acceptance checker, Go build, and Go tests.
  Direct `scripts/index-check.sh` and `scripts/invariant-lint.sh` passed; the index freshness text
  was an expected negative control.

## Open, blocked, and unverified

- Task 9’s independent final consolidation round remains open. No P6 deep-complete claim is made.
- The first `make check` stopped at `commit-cadence` because HEAD was over 90 minutes old while the
  preserved untracked user artifacts kept the worktree dirty; the documentation checkpoint then
  refreshed cadence.
- The final gate passed repository checks, Go build, Go tests, and Task 8 acceptance, but kernel lint remains
  blocked: golangci-lint 2.13.2 was built with Go 1.26.7 and panicked on Go 1.27 source. This is
  an environment/toolchain blocker, not a reported source lint failure.
- No derived artifact regeneration is indicated: generated freshness, direct index check, and
  invariant lint passed.
- Final documentation commit `04d60c12c55040f10ff4b373d2551e7e64ed7016` was pushed to
  `origin/main`. Final bundle: `/home/thamm/Backups/proofbound-20260919T144637Z.bundle`,
  SHA-256 `0995034c77d7ae1220bb40d9854106bac21867c9d736a26710918f3947f5e68f`; `git bundle verify`
  and a bare clone resolved main/HEAD to `04d60c1`.

## Institutionalized improvement

At EOD, if preserved user-owned untracked artifacts trip commit cadence, make a documentation-only
journal/state checkpoint, rerun the gate, and preserve the artifacts. Do not delete, add, or absorb
ownership-unclear files. Continue recording only complete calibrated mutation summaries.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Do not start or close Task 9 or claim P6 deep-complete without the independent final consolidation
  round.
