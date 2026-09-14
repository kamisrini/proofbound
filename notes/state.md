# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-14

## Resume — 2026-09-14

P6 Task 8 is active. The frozen production universe is 16 packages from `go list ./internal/...`;
`internal/specfirst` is test-only and explicitly excluded by C3-017. P6 Task 7 remains skipped under
the founder ratification; no P7+ capability is being added.

The current frozen implementation hash is `4499fa5`. It includes the committed GitHub, gates, and
migration test remediations after the prior `a5ef803` freeze. Migration also has a committed
fixture-root correction and archive-validation invariants; focused host tests pass.

The complete pre-commit CLI mutation sweep was diagnostic and reported 234 candidates, 233 killed,
0 invalid, and one survivor at `cli.go:347#233`; the survivor was the committed-intent Markdown
filter because the fixture had no non-Markdown artifact. The fixture now includes `notes.txt`, the
focused test passes, and targeted candidate `233` is killed. The committed sweep is green: 234
candidates, 234 killed, 0 invalid, 0 survived. Completed package runs are core 58/58, CLI 234/234,
checks 36/36, Git 34/34, gitcmd 86/86, GitHub 58/58, intent 93/93, intent records 52/52, intent
specdir 47/47, reviews 114/114, sessions 62/62, gates 109/109, and migration 33/33; every listed
run had zero invalid and surviving mutants. Migration’s committed evidence is
`docs/verification/p6-task8-migration-mutation-4499fa5.md`. The projections diagnostic was
intentionally interrupted before a complete summary and is not accepted evidence.

**Exact next action:** execute a complete calibrated integration-tagged mutation sweep for
`internal/projections` against current frozen implementation `4499fa5`, using bounded isolated
disposable databases if needed; record no result until the final summary reports all candidates,
zero invalid, and zero survived. Then continue with `internal/store` and `internal/twin` in order.
Do not close C3, start Task 9, or claim P6 deep-complete until all package results, exact tree
objects, and a current-code non-author verdict are committed.

## Branch and repository status

- Branch: `main`; local history is ahead of `origin/main`. Migration evidence and the state/journal
  refresh are currently uncommitted; no push is being attempted in this turn.
- The only pre-existing untracked path is `docs/plans/P5-intent-provenance-plan.md`, SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`. It is untouched and must
  remain uncommitted without ownership resolution.

## Completed

- P5 remains accepted; its ratified plan, semantic VD, founder record, adjudications, independent
  obligation verdict, mutation evidence, and delivery-readiness evidence are preserved.
- P6 authorization, Task 0 census, Task 1 C1/C2 closure, Task 2 C4 closure, Task 3 historical
  portability, Task 4 applicability, Task 5 requirement review, Task 6 measurements/falsifiers,
  and the 52-cell C5 route matrix are complete. Task 7 is explicitly skipped to P7+.
- Native Windows evidence for Task 3 and the Linux fresh-clone evidence for C8 are preserved. The
  post-change host bare `PATH=/home/thamm/go/bin:$PATH make check` passed with exit 0 and `0 issues`;
  expected negative-control diagnostics were not failures.
- Task 8 `internal/core` mutation acceptance is green: 58 candidates, 58 killed, 0 invalid, 0
  survived, after invariant tests and behavior-preserving guard decomposition.
- CLI focused tests, tagged integration tests, and the complete committed mutation sweep are green:
  234 candidates, 234 killed, 0 invalid, 0 survived. Checks is 36/36, Git is 34/34, and gitcmd is
  86/86; all have zero invalid and surviving mutants. GitHub, intent, records, specdir, reviews,
  sessions, gates, and migration also completed with zero invalid and surviving mutants.

## Open work

- P6 Task 8 remains open for the other 3 production packages: projections, store, and twin. Final package evidence
  must bind one implementation commit, each exact package tree object, calibrated mutation counts,
  and a committed non-author verdict for that same commit/tree.
- Final C3 census closure, Task 9 round-C consolidation, and the P6 deep-complete claim remain open.
- The current state has no product implementation blocker. Projections integration mutation
  testing is operationally expensive; bounded isolated databases are required to finish it without
  treating interrupted chunks as evidence. No disposable DB service is left running.

## Blocked and unverified

- The pre-commit CLI sweep found and remediated one survivor; its 234/233/0/1 result is diagnostic.
- Completed package mutation runs are verified individually, but the final cross-package packet
  and current-code non-author verdict remain unverified. Projections, store, and twin remain
  unverified.
- No current non-author final package verdict covers the eventual frozen implementation commit.
- Push is not attempted; destination-specific authorization/authentication has not been freshly
  established for this turn.
- The pre-existing untracked P5 draft is intentionally untouched and excluded from all commits.

## Institutionalized improvement

Resume notes remain at the top with a dated exact next action. Mutation survivors now require a
focused invariant test or a semantics-preserving simplification before closure. Mutation evidence
is never accepted from an interrupted or pre-commit run; package acceptance requires the complete
calibrated summary after the frozen implementation commit. Disposable database runs are isolated,
and expected negative-test diagnostics are recorded separately from actual gate failures. When
integration mutation suites are slow, isolated databases may parallelize execution, but only a
complete per-package summary is durable evidence.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, all pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Use host execution for the Snap Go toolchain. Keep the pre-existing untracked P5 draft untouched.
