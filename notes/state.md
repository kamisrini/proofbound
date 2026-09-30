# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-09-30

## Resume — 2026-09-30

The P6 final consolidation and P7 planning boundary are complete and pushed at 6902a79; no
implementation work was performed today. The exact next action is: resolve P7 Gate 1 by recording
the founder decision to retain the snapshot-provider skip, or naming one concrete lawful export
feed and explicitly authorizing the narrow width exception. Until that decision is recorded, do
not implement a provider or introduce a feed, vendor, schema, event kind, or other P7+ capability.

Closeout commit dc34cde is created; authenticated push and a verified dated Git bundle remain
pending. The two preserved untracked artifacts must remain untouched and excluded from commits.

## Branch and repository status

- Branch: main, tracking origin/main; HEAD is dc34cde and origin/main remains 6902a797.
- The documentation closeout commit dc34cde is today's only commit.
- Preserved untracked artifacts, excluded from all commits:
  docs/plans/P5-intent-provenance-plan.md (SHA-256
  eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439) and
  kernel/cmd/p6dbprobe/main.go (SHA-256
  f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0).
- Ignored runtime/session directories were inspected. No relevant PostgreSQL, Proofbound,
  mutation-worker, or p6dbprobe process remained active.

## Completed

- P6 Task 9 final consolidation is complete: round-C is ACCEPTABLE, the final census is 327/327
  closed with zero unclassified rows, and P1–P5 are annotated deep-complete.
- The P7 planning draft is committed. It authorizes planning only; the snapshot-provider decision
  is the first gate and no implementation is authorized before it is resolved.
- Today's audit inspected current status, history, journals, notes, repository laws, plans,
  ignored/session artifacts, and live processes. No implementation change was found.
- The protected artifact hashes match the prior durable handoff.

## Verification

- The first bare make check reached repository checks, generated freshness, P6 census and package
  acceptance checks, and the Windows contract before stopping at commit-cadence because the
  protected untracked artifacts kept the worktree dirty past the 90-minute window.
- After dc34cde, bare make check passed cadence, repository checks, generated freshness, P6 census,
  package acceptance, and the Windows contract. It stopped at kernel-check because
  golangci-lint was not available on PATH (exit 127); no source lint failure was reported.
- The line index stale; run make index was expected output from an intentional negative-control
  test. Generated-index and invariant-lock freshness passed; no regeneration was required.
- The compatible Go 1.27.1 / matching golangci-lint run recorded on 2026-09-22 passed make check
  and make verify; that result is historical, not a replacement for today's missing-linter stop.

## Open, blocked, and unverified

- Open decision: P7 Gate 1 must retain the snapshot-provider skip or name a lawful export feed and
  explicitly authorize the narrow width exception.
- Blocked: provider implementation and all downstream P7+ capability work until that decision is
  recorded.
- Unverified today: kernel build/test/lint and make verify were not reached after the missing-linter
  stop; the prior compatible-toolchain results remain the only evidence for those checks.
- Unverified assumption: the two untracked artifacts are user-owned or otherwise ownership-unclear;
  their matching hashes support preservation, not authorization to commit them.
- Native Windows acceptance remains an evidence requirement and is not assumed green from the
  Windows contract test.

## Institutionalized improvement

When preserved untracked artifacts trip the cadence guard, record their exact hashes in the
journal and state, make a documentation-only checkpoint, rerun the gate, and preserve the files.
Do not delete, stage, or absorb ownership-unclear artifacts. No new automation is justified by
this audit; the existing EOD checkpoint and cadence rule are the lightweight repeatable control.

## Standing cautions

- Preserve frozen vera.witness.v1, vera.verdict.v1, vera.replay.v1, pinned vectors, and
  historical artifacts.
- The archive is migration-only and exact; no broad import, new provider, event kind, platform
  promise, or P7+ capability is authorized.
- P7 planning is allowed after P6 closure, but implementation remains blocked at the
  snapshot-provider decision until the skip is reaffirmed or a lawful feed and width exception are
  explicitly named.

## Exact next action

Resolve P7 Gate 1: record the founder decision to retain the snapshot-provider skip, or name one
concrete lawful export feed and explicitly authorize the narrow width exception. Do not implement a
provider before that decision.
