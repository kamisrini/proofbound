# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-13

## Resume — 2026-09-13

Founder ratified historical-evidence portability option 1. The strict migration-only archive now
contains five exact ledger envelopes: the two cited P5 evidence records and their three exact
intent-record prerequisites. The receipt is committed at
`docs/verification/verdicts/p6-historical-evidence-ratification.md`; the semantic VD is
`docs/decisions/VD-p6-historical-evidence-portability-2026-09-13.md`.

Task 3 acceptance has passed on both platforms at frozen commit
`b3303100f649c3987003a959668a261ae9b7b3f6`. Linux evidence is
`docs/verification/p6-fresh-clone-linux.md`; native Windows evidence is
`docs/verification/p6-windows-platform.md`. Both records bind migration `imported=5`, two-pass
sync with zero second-pass appends, witnessed/bare `make check`, verify, rebuild, and the two
self-hosted reports. C8-001 and C8-002 are now mechanically closed by their committed artifacts.

**Exact next action:** add the narrow applicability check to the existing `make delivery-enforce`
boundary and write its seeded missing-intent test. The complete post-`f426ca8` canary report is now
committed: 66 commits examined, 38 applicable, zero exact claims, no known false negatives, and no
threshold trigger. Do not start Task 5 or widen P6 scope.

## Branch and repository status

- Branch: `main`; Task 3 closure and the Task 4 matcher/canary slices are committed. Local history is ahead of
  `origin/main`; no push is being attempted in this turn.
- The only pre-existing untracked path is `docs/plans/P5-intent-provenance-plan.md`, SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`. It is untouched and must
  remain uncommitted without ownership resolution.

## Completed

- P5 remains accepted; its ratified plan, semantic VD, founder record, adjudications, independent
  obligation verdict, mutation evidence, and delivery-readiness evidence are preserved.
- P6 authorization, Task 0 census, Task 1 C1/C2 closure, Task 2 C4 closure, and the 52-cell C5
  route matrix are complete. Task 7 is explicitly skipped to P7+.
- Task 3 historical-evidence portability is implemented spec-first with exact archive validation,
  migration-only import permission, explicit CLI routing, exact-import/mutation/non-empty-ledger
  tests, and unchanged normal connector/projection/verify behavior.
- Native Windows passed the PowerShell entry point, focused replay test, bare and witnessed checks,
  five-record migration, two sync passes, verify, rebuild, and both reports. The earlier failures
  were recorded as process/PATH diagnostics, not allowed as evidence.
- Derived census and artifact-integrity outputs have been regenerated; C8-001 and C8-002 are
  evidence-bound. The post-change bare `make check` passed with expected negative diagnostics
  separated from actual failures. The pre-existing untracked P5 draft remains untouched.
- Task 4 now has a ratified path-only applicability SPEC, executable matcher, invariant tests, and
  complete post-anchor canary report; the matcher is intentionally not yet wired into delivery
  enforcement.

## Open work

- P6 Tasks 4, 5, 6, 8, and 9 remain open; Task 7 is deferred to P7+ by ratified decision.
- Census categories C3, C6, and C7 still contain open `close-in-P6` rows, including package
  acceptance, review completion, and measurements/falsifier evaluation.
- Final P6 package acceptance, mutation sweep, non-author current-code verdict, final census with
  zero open `close-in-P6` rows, and round-C close remain outstanding.

## Blocked and unverified

- No current implementation blocker remains for Task 3. The native Windows fixed-port collision
  and missing-`env` PATH issue were resolved by disposable-process cleanup and explicit tool PATH;
  neither changed repository behavior.
- Task 4 delivery-boundary enforcement and its seeded missing-intent proof remain open. The canary
  is observation only; no enforcement claim is made yet.
- Push is not attempted; destination-specific authorization/authentication has not been freshly
  established for this turn.

## Institutionalized improvement

Resume notes remain at the top of this file with a dated exact next action. Acceptance records now
name the frozen commit, toolchain, exit codes, expected negative diagnostics, and disposable-state
boundaries. Native platform runs must use one isolated process tree, an explicit PATH including Git
`usr\\bin`, and verified cleanup before a retry; no test skip is allowlisted merely to mask a fixed
port collision. Generated census/integrity artifacts are regenerated in the same coherent change as
the evidence they summarize.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, all pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Use host execution for the Snap Go toolchain. Keep the pre-existing untracked P5 draft untouched.
