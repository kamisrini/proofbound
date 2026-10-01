# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-10-01

## Resume — 2026-10-01

P7 Gate 1 is resolved: retain the P6 snapshot-provider skip. The exact next action is to obtain a
non-author independent review of docs/plans/P7-behavior-locks-regeneration-plan-draft1.md and its
proposed kernel/internal/twin replay subject, then record accept, amend, or reject before any
implementation. Do not implement a snapshot provider, behavior lock, or regeneration yet.

## Branch and repository status

- Branch: main, tracking origin/main; local main remains ahead because prior push publication was
  blocked by policy.
- Today's changes are documentation-only. The preserved untracked artifacts remain excluded:
  docs/plans/P5-intent-provenance-plan.md (SHA-256
  eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439) and
  kernel/cmd/p6dbprobe/main.go (SHA-256
  f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0).

## Completed

- P6 final consolidation remains complete: round-C ACCEPTABLE, 327/327 census rows closed, and
  P1–P5 deep-complete.
- Recorded the founder decision in
  docs/decisions/VD-p7-snapshot-provider-skip-2026-10-01.md.
- Updated the P7 snapshot-provider plan, roadmap, and decision index.
- Drafted the behavior-locks/regeneration plan and proposed the bounded twin replay subject.
  No implementation or frozen wire identity changed.

## Verification

- Focused index, state, lesson, invariant, spec-numbering, and link checks passed.
- Bare make check passed cadence, repository checks, generated freshness, P6 census, package
  acceptance, and platform stages, then stopped at kernel-check because golangci-lint was unavailable
  (exit 127). The expected index stale; run make index output was an intentional negative-control
  message; no source lint failure was reported.
- Kernel build/test/lint and make verify remain unverified today because the compatible linter is
  unavailable. Native Windows acceptance remains an evidence requirement and is not assumed green.

## Open, blocked, and unverified

- Open: non-author independent review of the proposed twin replay subject and the behavior-lock plan.
- Blocked: all behavior-lock, regeneration, snapshot-provider, and external-feed implementation
  until the relevant scope and acceptance decision are recorded.
- Unverified assumption: the existing twin replay evidence is sufficient for a behavior-lock pilot;
  independent review must test this before implementation.

## Institutionalized improvement

Keep capability boundaries decision-first: record the source, scope, and acceptance bar before
writing capability code. When no real feed is needed, retain the skip and continue only within
existing feeds. No new automation is justified.

## Standing cautions

- Preserve frozen vera.witness.v1, vera.verdict.v1, vera.replay.v1, pinned vectors, and
  historical artifacts.
- No provider, feed, vendor, schema, event kind, or P7+ implementation is authorized by the
  snapshot-provider decision.
- Planning drafts are not implementation authorization.

## Exact next action

Obtain a non-author independent review of docs/plans/P7-behavior-locks-regeneration-plan-draft1.md
and its proposed kernel/internal/twin replay subject. Record accept, amend, or reject before any
implementation.
