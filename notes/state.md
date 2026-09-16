# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-16

## Resume — 2026-09-16

P6 Task 8 remains active. The frozen implementation is `e4c8e77407699f7e089d5c1a2b3ce58df5871fbf`.
Today’s factual journal/state update is committed next; the remaining EOD closeout actions are to
create and clone-verify the dated bundle and push the completed checkpoint to authenticated
`origin`, as conditionally requested. Bare `make check` was rerun: repository checks passed, while
kernel lint is blocked by the installed Go-1.26-built linter panicking on Go 1.27 source. Direct
generated-index and invariant checks passed, so no derived artifact regeneration is indicated.

Tomorrow’s first implementation action: at `e4c8e77`, run a one-candidate calibrated integration
pilot for `internal/cli` against a fresh disposable PostgreSQL instance, measure elapsed time, then
choose a capped partition concurrency and finish the complete 234 CLI and 404 projections
candidates. Each concurrent partition must have its own disposable database; only complete
calibrated package summaries count. Obtain a current-code non-author verdict for all 16 package
rows before closing C3 or starting Task 9.

## Branch and repository status

- Branch: `main`, tracking `origin/main`. The remote main ref was `1d1112df2622abd54f07837335a7805c3c69145d`
  and an ancestor of local main when checked; GitHub authentication is active (`kamisrini`, `repo`
  scope). The local-only backup decision remains documented, but this EOD request conditionally
  authorizes pushing the committed history after the verified bundle is made.
- The two pre-existing untracked paths are `docs/plans/P5-intent-provenance-plan.md` (SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`) and
  `kernel/cmd/p6dbprobe/main.go` (SHA-256
  `f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0`). Preserve both and keep
  them out of commits unless ownership is resolved.
- Repository-scoped session artifact searches found no files in `notes/tmp`, `.codex`, `.agents`,
  or the expected user session directory. No mutation workers, PostgreSQL servers, or P6 temporary
  database directories remained at inspection.

## Completed

- Today’s committed documentation checkpoint `a199229` records 12 fresh complete calibrated
  mutation reruns at the frozen commit: core 58/58, checks 36/36, Git 34/34, gitcmd 86/86,
  GitHub 58/58, intent 93/93, records 52/52, specdir 47/47, reviews 114/114, sessions 62/62,
  gates 109/109, and migration 33/33; each has 0 invalid and 0 survivors.
- The fresh untagged gates probe’s 38 survivors are diagnostic because integration proving tests
  were omitted. Correctly tagged gates rerun passed 109/109/0/0. The package-universe test passed;
  the 16 production packages and test-only `internal/specfirst` classification remain explicit.
- Frozen store and twin results are 133/133/0/0 and 31/31/0/0. Earlier CLI evidence is 234/234/0/0
  at an earlier freeze; projections is 404/404/0/0 at an earlier freeze. Neither replaces the
  missing complete frozen-commit rerun.
- P5 remains accepted. P6 Tasks 0–6 are complete; Task 7 is explicitly skipped under ratification.
  Historical portability, route-matrix, platform, and prior acceptance evidence remain preserved.
- Direct `scripts/index-check.sh` and `scripts/invariant-lint.sh` passed during EOD inspection.

## Open, blocked, and unverified

- Task 8 still needs complete frozen-commit CLI and projections sweeps and committed exact-tree
  rows for all packages. Two attempts at those DB-backed sweeps were interrupted before complete
  summaries; their partial output is not evidence.
- No current-code non-author verdict covers all package rows. C3 closure, Task 9 round-C
  consolidation, and the P6 deep-complete claim remain prohibited until the package packet and
  independent verdict are committed.
- The first bare `make check` failed at `commit-cadence` because HEAD was over 90 minutes old while
  the only dirty paths were the two pre-existing untracked files. After a documentation checkpoint
  refreshed HEAD, a rerun passed all repository checks and Go build/tests, then golangci-lint v2.13.2
  panicked because it was built with Go 1.26 and encountered Go 1.27 source. This is an environment
  compatibility blocker, not a reported source lint error. An intermediate bare invocation also
  lacked the linter in PATH; the host-path rerun is the meaningful gate result.
- The `index stale; run make index` text came from an expected negative-control test. Direct
  `scripts/index-check.sh` and `scripts/invariant-lint.sh` passed, so no regeneration was needed.
- Bundle creation/clone verification and conditional push remain the final closeout steps. No
  current-code non-author package verdict exists.

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
