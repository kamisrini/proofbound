# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-18

## Resume — 2026-09-18

P6 Task 8 remains active. The frozen implementation is
`e4c8e77407699f7e089d5c1a2b3ce58df5871fbf`; later commits are documentation-only. The requested
one-candidate `internal/cli` integration pilot passed calibration and killed its candidate in
15.22 seconds. Capped concurrency is two partitions, each with its own disposable PostgreSQL 16.4
instance. The complete CLI rerun passed: 234 killed, 0 invalid, 0 survived.

The complete `internal/projections` sweep passed: 404 killed, 0 invalid, 0 survived. Its four
calibrated ranges were intent/report 1–58 (58 candidates, 949.16 seconds), core 59–202 (144,
3,852.81 seconds), core 203–345 (143, 3,311.33 seconds), and report 346–404 (59, 701.70 seconds).
The two concurrent core partitions and final report range each had their own disposable PostgreSQL
instance. Only complete summaries count. The first filter attempt omitted the lethal calibration
test and stopped before candidate 1; an earlier unbalanced core attempt was interrupted at 26
candidates; and the first report calibration stopped before candidate 1 because its cluster lacked
the `proofbound` database. These attempts are excluded. Disposable database checkpoint settings
were raised to 30 minutes and 8 GB WAL after long file-sync pauses; transaction and durability
settings remained default.

The exact-tree 16-package result matrix and current-code non-author verdict have been drafted.
Acceptance checker, census wiring, and focused negative tests are in the worktree. C3 remains open
until the verdict is committed and the census rows cite both artifacts. Task 9 has not started.
Next: stage and commit the evidence packet plus verdict in the required order, verify all acceptance
checks, then close C3 in a separate census update. Keep the frozen Go implementation at `e4c8e77`.

## Branch and repository status

- Branch: `main`, tracking `origin/main`; current committed HEAD before this work was `0067dc6`
  (`docs: finalize durable EOD state`). The Task 8 acceptance checker, its test, evidence packet,
  verdict, and this status update are in progress. Commit only explicit Task 8 and state files;
  preserve both pre-existing user artifacts below.
- Preserve the pre-existing untracked `docs/plans/P5-intent-provenance-plan.md` (SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`) and
  `kernel/cmd/p6dbprobe/main.go` (SHA-256
  `f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0`). Keep both out of commits.
- Task-local frozen checkout and disposable database clusters are under `/tmp`; stop and remove
  them after the remaining projection partitions finish.
- The latest bare `make check` passed repository checks and Go build/tests. Kernel lint remains
  blocked because installed golangci-lint was built with Go 1.26 and panics on Go 1.27 source.
  Direct generated-index and invariant checks passed. The artifact-integrity audit was regenerated
  to account for the new schema-bearing verdict; the generated index and invariant table needed no
  change.

## Completed

- At frozen `e4c8e77`, 12 fresh calibrated package reruns passed with zero invalid or survivors:
  core 58, checks 36, Git 34, gitcmd 86, GitHub 58, intent 93, records 52, specdir 47, reviews
  114, sessions 62, gates 109, and migration 33 candidates.
- Frozen store and twin results are 133/133 and 31/31, each with zero invalid or survivors.
- The current frozen CLI rerun is 234/234/0/0. Two calibrated 117-candidate partitions used
  separate fresh PostgreSQL instances; elapsed times were 707.11 and 554.96 seconds.
- The package-universe check passed: 16 production packages and test-only `internal/specfirst`.
- Direct `scripts/index-check.sh` and `scripts/invariant-lint.sh` passed during EOD inspection.
- P5 remains accepted. P6 Tasks 0–6 are complete; Task 7 is explicitly skipped under ratification.
  Historical portability, route-matrix, platform, and prior acceptance evidence remain preserved.

## Open, blocked, and unverified

- The frozen result matrix and current-code non-author verdict are not yet committed. Do not close
  C3, start Task 9, or claim P6 deep-complete until the exact-tree packet and verdict are committed.
- The current linter compatibility issue is environmental, not a reported source lint error; do not
  claim kernel lint passed.
- The new schema-bearing verdict required updating the generated artifact-integrity summary; the
  generated index and invariant table remain unchanged.

## Institutionalized improvement

For DB-backed mutation work, first run a bounded calibrated pilot and measure throughput, then use
capped concurrency with a separate disposable database per partition. Exclude interrupted output;
record only complete summaries. Continue preserving untracked user artifacts, and refresh the
resume state alongside every durable checkpoint.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, all pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Do not close C3, start Task 9, or claim P6 deep-complete without exact frozen package evidence and
  the committed current-code non-author verdict.
