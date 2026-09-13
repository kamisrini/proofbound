# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-13

## Resume — 2026-09-13

Founder ratified historical-evidence portability option 1: a strict migration-only archive of the
two exact ledger envelopes cited by the accepted P5 obligation verdict. The receipt is committed at
`docs/verification/verdicts/p6-historical-evidence-ratification.md`; the semantic VD is
`docs/decisions/VD-p6-historical-evidence-portability-2026-09-13.md`.

The archive SPEC and exact fixture/tests were committed before implementation. The migration,
store permission, explicit `proofbound migrate historical-evidence` command, package SPEC, README
operator note, and identity-classification fix are committed in `7273311`; the Task 3 route artifact
and archive reference are committed in `d89f423`. Normal sync, projection, verify, reports, and
failed-closed dangling-reference behavior remain unchanged.

The verifier showed that the two cited evidence events also require three exact P5 intent-record
prerequisites in ledger order; those five records are now the archive's minimal transitive closure.

**Exact next action:** commit the five-record archive/spec/VD amendment, then rerun a clean detached
Linux acceptance: bare `make check`; `proofbound migrate historical-evidence`; first and second
`sync all`; `proofbound verify`; rebuild equality; and self-hosted reports. Bind the result, then
repeat the already-required native Windows acceptance before closing C8-001/C8-002. Do not start
Task 4 or relax referential integrity.

## Branch and repository status

- Branch: `main`; current HEAD when this note was written: `d89f423`; local branch is ahead of
  `origin/main` and has not been pushed.
- Push remains unauthorized at the destination-specific safety boundary for
  `https://github.com/kamisrini/proofbound.git`; do not infer write authorization from remote read
  access.
- The only pre-existing untracked path is `docs/plans/P5-intent-provenance-plan.md`, SHA-256
  `eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439`. It is untouched and must
  not be committed without ownership resolution.

## Completed

- P5 remains accepted; its ratified plan, semantic VD, founder record, adjudications, independent
  obligation verdict, mutation evidence, and delivery-readiness evidence are preserved.
- P6 authorization, Task 0 census, Task 1 C1/C2 closure, Task 2 C4 closure, and the 52-cell C5
  route matrix are complete. Task 7 is explicitly skipped to P7+.
- Native Windows portability work is implemented and the prior native run passed the PowerShell
  gate, witnessed check, and two-pass sync; its verify was blocked only by the shared historical
  dangling event.
- The option-1 archive contains the three exact intent prerequisites at sequences 1227–1229 plus
  sequence 1670/event `01M28TPW9C8R7ND19MNDCJ9GDG` and sequence 1834/event
  `01M29HMPE5V977AR3VMW47DVDE`. Focused parser, mutation, exact-import, non-empty-ledger,
  CLI-routing, and package tests pass.

## Verification and current blockers

- Prior author-side bare `make check` and `proofbound verify` are green on the existing migrated
  local ledger; the earlier Linux linter result is historical evidence only.
- The current bare check reached the gates but reported the expected generated-artifact/census
  follow-ups: the Task 2 artifact still names the pre-migration code commit, and the census lacks
  the new `internal/migration` C3 row. No implementation test failure was reported.
- A clean detached clone at `d89f423` reached the same two follow-ups and then stopped at state
  freshness because this note had not yet been refreshed. That acceptance is unverified and must be
  rerun after the next coherent documentation commit.
- C8-001 Linux and C8-002 native Windows remain open until migration-enabled fresh-clone verify,
  rebuild equality, and reports are green and evidence-bound. Tasks 4–9 remain unstarted; C3, C6,
  and C7 remain open.

## Unverified assumptions

- The archive was extracted from the local migrated ledger and its complete line digests are bound
  in the migration implementation; the fresh-clone end-to-end proof has not yet established that
  the archive makes the accepted P5 chain fully reproducible.
- Native Windows must be rerun after the final migration-enabled code/artifact commit; the previous
  green-through-sync run is not evidence for this implementation.

## Institutionalized improvement

Resume state is maintained at the top of this file with a dated exact next action. New production
packages must be added to the P6 census registry in the same coherent change, and generated closure
artifacts must be regenerated only after the code commit they bind. Focused tests are not treated as
package acceptance; fresh-clone commands and non-author verdicts remain required.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, all pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- Use host execution for the Snap Go toolchain. Keep the pre-existing untracked P5 draft untouched.
