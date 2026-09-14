# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-14

## Resume — 2026-09-14

P6 Task 8 is active. The frozen production universe is 16 packages from `go list ./internal/...`;
`internal/specfirst` is test-only and explicitly excluded by C3-017. `internal/core` is accepted by
its calibrated result of 58 candidates, 58 killed, 0 invalid, 0 survived. P6 Task 7 remains skipped
under the founder ratification; no P7+ capability is being added.

The CLI remediation is committed at `3f115ea`. It adds behavior-preserving
dispatch/control-flow simplifications and narrow seams for store opening, sync creation, gate
evaluation, projection application, ID generation, event reading, and close handling. Focused host
`go test ./internal/cli -count=1` passes, and the DB-tagged integration suite passed before mutation.

The complete pre-commit CLI mutation sweep was diagnostic and reported 234 candidates, 233 killed,
0 invalid, and one survivor at `cli.go:347#233`; the survivor was the committed-intent Markdown
filter because the fixture had no non-Markdown artifact. The fixture now includes `notes.txt`, the
focused test passes, and targeted candidate `233` is killed. The required complete sweep against
the frozen commit is now green: 234 candidates, 234 killed, 0 invalid, 0 survived. Evidence is in
`docs/verification/p6-task8-cli-mutation-3f115ea.md` and the exact CLI tree is
`6a1a1804018be77c5a17a3d70c024c2978511683`.

**Exact next action:** run the calibrated mutation sweep for the next production package in frozen
order, `internal/connector/checks`, against commit `3f115ea`; record its counts and tree object,
then continue package-by-package. Do not close C3, start Task 9, or claim P6 deep-complete until all
package results, exact tree objects, and a current-code non-author verdict are committed.

## Branch and repository status

- Branch: `main`; local history is ahead of `origin/main`. The CLI acceptance evidence and state/
  journal refresh are currently uncommitted; no push is being attempted in this turn.
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
  234 candidates, 234 killed, 0 invalid, 0 survived.

## Open work

- P6 Task 8 remains open for the other 15 production packages. Final package evidence
  must bind one implementation commit, each exact package tree object, calibrated mutation counts,
  and a committed non-author verdict for that same commit/tree.
- Final C3 census closure, Task 9 round-C consolidation, and the P6 deep-complete claim remain open.
- The current state has no implementation blocker; the disposable DB service was stopped after the
  CLI sweep. Remaining package sweeps may need their own disposable DB service if tagged tests need it.

## Blocked and unverified

- The pre-commit CLI sweep found and remediated one survivor; its 234/233/0/1 result is diagnostic.
- CLI committed acceptance is now verified, but the other package rows and current-code non-author
  verdict remain unverified.
- No current non-author final package verdict covers the eventual frozen implementation commit.
- Push is not attempted; destination-specific authorization/authentication has not been freshly
  established for this turn.
- The pre-existing untracked P5 draft is intentionally untouched and excluded from all commits.

## Institutionalized improvement

Resume notes remain at the top with a dated exact next action. Mutation survivors now require a
focused invariant test or a semantics-preserving simplification before closure. Mutation evidence
is never accepted from an interrupted or pre-commit run; package acceptance requires the complete
calibrated summary after the frozen implementation commit. Disposable database runs are isolated,
and expected negative-test diagnostics are recorded separately from actual gate failures.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, all pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Use host execution for the Snap Go toolchain. Keep the pre-existing untracked P5 draft untouched.
