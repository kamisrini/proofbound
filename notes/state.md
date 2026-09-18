# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-18

## Resume — 2026-09-18

P6 Task 8 package acceptance is complete. The frozen implementation is
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

The exact-tree 16-package result matrix and current-code non-author verdict are committed. The
acceptance checker validates the frozen commit, production package set, tree objects, complete
calibrated counts, verdict identity/schema/digest, and census evidence links. Its negative tests
passed, including author impersonation with the multiword code-author identity. The census now
binds all 16 C3 production package rows to the matrix and committed verdict; C3-017 remains the
explicit `internal/specfirst` test-only exclusion. The rendered census has 7 open, 309 closed, and
0 unclassified rows. Task 9 has not started.

Updating the census generator advanced its input to `8bef29020d0799c7a71efba1230c85e0860a3e4f`.
Per the C6 census SPEC, the input interval includes every post-`f426ca8` commit through that hash.
The 134-commit interval is now represented by exact C6 commit rows. A refreshed path-only canary at
current HEAD `e02dbb1c05695451961f0cad4a669501c47c95b3` examined 135 commits, 81 applicable, with
zero explicit Intent trailers. The current C6 summary is recorded in
`docs/verification/p6-intent-coverage.md` and its full table in the new canary artifact.

The final `make check` was run with `PATH="/snap/go/current/bin:/home/thamm/go/bin:$PATH"`
and `GOFLAGS=-p=1`, with loopback access enabled for embedded PostgreSQL tests. Repository checks,
Go build, and Go tests passed. Kernel lint did not run to completion: installed golangci-lint
2.13.2 was built with Go 1.26.7 and panicked on Go 1.27 source with `file requires newer Go
version go1.27`. Direct generated-index and invariant checks passed. The C3 closure and this result
are committed; Task 9 has not started.

## Branch and repository status

- Branch: `main`; the Task 8 acceptance, checker-fix, C6 refresh, and C3 closure commits through
  `624e704` have been pushed to `origin/main`. That closure commit is included in verified bundle
  `/home/thamm/Backups/proofbound-20260918T230921Z.bundle` (SHA-256
  `2cc596921755ccc890b76994ca9b51a936df57901e15faa81e130ea7897a4252`); `git bundle verify` and
  a bare clone confirmed its `main` and `HEAD` refs. The final notes-only handoff checkpoint follows
  the closure commit and will be included in a fresh verified bundle and push before handoff.
  Preserve both pre-existing user artifacts below.
- Preserve the pre-existing untracked `docs/plans/P5-intent-provenance-plan.md` (SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`) and
  `kernel/cmd/p6dbprobe/main.go` (SHA-256
  `f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0`). Keep both out of commits.
- Task-local disposable PostgreSQL clusters were stopped after all partitions completed. Mutation
  checkouts and per-run database directories were removed from `/tmp`.
- The final `make check` passed repository checks, Go build, and Go tests. Kernel lint remains
  blocked because installed golangci-lint 2.13.2 was built with Go 1.26.7 and panics on Go 1.27
  source. Direct generated-index and invariant checks passed. The artifact-integrity audit was
  regenerated to account for the new schema-bearing verdict; the generated index and invariant
  table needed no change.

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

- Do not start Task 9 or claim P6 deep-complete; the independent final consolidation round remains
  separate.
- The current linter compatibility issue is environmental, not a reported source lint error; do not
  claim kernel lint passed.
- The new schema-bearing verdict required updating the generated artifact-integrity summary. The
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
