# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-10-14

## Resume — 2026-10-14

Exact next action: obtain a fresh non-author review of the exact P5 business-decision and active
requirement revisions, `O-1`, the draft twin SPEC addendum, and fixture matrix. Record `ACCEPT`,
`AMEND`, or `REJECT` against the reviewed commit and evidence revision. Keep implementation gated
until the slice and later implementation have their own independent verdicts.

Immediate blockers: the P5 records and addendum/fixture slice have not been independently
reviewed; the full gate cannot complete because the installed linter crashes on a Go 1.27
version marker; committed fixture corpus, proof-byte serialization, projector revision evidence,
and implementation verdict are absent. Push
authorization/authentication remain unverified after the prior GitHub DNS failure.

## Branch and repository status

- Branch: `main`, tracking `origin/main`, ahead of the locally known remote-tracking ref.
  The remote tip is unverified.
- Preserved untracked user artifacts: `docs/plans/P5-intent-provenance-plan.md` and
  `kernel/cmd/p6dbprobe/main.go`. They are untouched and excluded from this checkpoint.
- The new P7 addendum and fixture matrix are draft planning artifacts only.

## Completed

- P6 consolidation and the P7 snapshot-provider skip remain accepted.
- Fresh round-3 non-author review accepted the amended P7 `ReplayIsolated` planning subject and
  contract; the verdict is `docs/verification/verdicts/p7-behavior-lock-plan-round3-20261013.md`.
- On 2026-10-14, the owner authorized drafting the lock-schema/addendum and fixture-planning slice.
  Drafts are `docs/plans/P7-replay-lock-spec-addendum-draft1.md` and
  `docs/plans/P7-replay-lock-fixture-matrix-draft1.md`. No product code, provider, feed, wire
  identity, event kind, or persistence surface changed.
- The owner requested a recommendation for the pilot's intent binding. The recommended path is a
  genuine P5 requirement, retaining `UNDECLARED` as a negative control. The owner authorized a
  draft proposal, then approved its business outcome and `O-1` wording and selected existing
  declared roles. Formal records are `BD-replay-evidence-7a91c4` and
  `BR-replay-evidence-6e3d2b/0001`; the active requirement binds the exact decision digest.

## Verification

- Bare `make check` ran on 2026-10-14. The `index stale; run make index` output was an expected
  negative control; generated freshness and hook tests passed. The actual stop was
  `commit-cadence: dirty worktree outlived 90m commit window`, so the main DAG did not reach
  kernel checks. A Snap Go confinement diagnostic in a self-test remains an environment issue.
- After documentation commit `7ca242b`, bare `make check` was rerun. Cadence and the repository,
  generated freshness, P6 package/verdict, and Windows contract checks passed. It reached
  `kernel-check` and stopped because this host's Snap Go confinement lacks
  `cap_dac_override` (`Function not implemented`). Go build/tests/lint remain unverified here;
  no source or generated-artifact failure was observed.
- After formal P5 record commit `9e0621f`, both normalized self-digests and the exact decision
  reference were verified; JSON metadata parses. The P5 records package tests passed with Go
  outside the sandbox. Bare `make check` passed repository checks and Go build/tests, then stopped
  because `golangci-lint` was absent from PATH. With its installed directory on PATH, the gate
  again passed repository checks and Go build/tests, then the linter panicked: a file requires
  Go 1.27 while the installed linter was built with Go 1.26.7. Lint remains unverified; this is
  a tooling failure, not evidence of a source defect.

## Open, blocked, and unverified

- Open: obtain non-author review of the exact requirement and draft slice; ratify proof-byte
  serialization and projector revision evidence; build committed fixtures and evidence manifest;
  obtain a separate non-author review of any later implementation.
- Blocked: behavior-lock, regeneration, snapshot-provider, and external-feed implementation until
  the applicable review and decision gate is satisfied. Push remains unverified.
- Unverified: independent verifiability of the new requirement, complete clean-clone replay,
  detector power for a P7 mutation set, lint with a compatible toolchain, native Windows
  acceptance, and remote-tip freshness.

## Institutionalized improvement

When preserved user-owned untracked artifacts trigger cadence, keep them untouched, record their
ownership, commit only a coherent documentation checkpoint, and rerun the full gate. No new
automation is justified.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and
  historical artifacts.
- Planning acceptance and draft review do not authorize behavior-lock or regeneration code.
