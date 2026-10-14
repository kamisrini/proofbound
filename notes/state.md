# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-10-14

## Resume — 2026-10-14

Exact next action: ask the owner whether the first real `ReplayIsolated` lock should explicitly
remain `UNDECLARED` or be bound to a newly authorized, genuinely applicable P5 requirement. Then
submit the draft twin SPEC addendum and fixture matrix for fresh non-author review; record
`ACCEPT`, `AMEND`, or `REJECT` against their exact revision before any implementation.

Immediate blockers: the addendum and fixtures have not been independently accepted; no real
`ReplayIsolated` intent obligation has been authorized; committed fixture corpus, proof-byte
serialization, projector revision evidence, and implementation verdict are absent. Push
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

## Open, blocked, and unverified

- Open: choose the pilot's honest intent binding; ratify proof-byte serialization and projector
  revision evidence; build committed fixtures and evidence manifest; obtain non-author review of
  the draft slice and later of any implementation.
- Blocked: behavior-lock, regeneration, snapshot-provider, and external-feed implementation until
  the applicable review and decision gate is satisfied. Push remains unverified.
- Unverified: a real requirement applicable to this replay subject, complete clean-clone replay,
  detector power for a P7 mutation set, kernel build/tests/lint in this host, native Windows
  acceptance, and remote-tip freshness.

## Institutionalized improvement

When preserved user-owned untracked artifacts trigger cadence, keep them untouched, record their
ownership, commit only a coherent documentation checkpoint, and rerun the full gate. No new
automation is justified.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and
  historical artifacts.
- Planning acceptance and draft review do not authorize behavior-lock or regeneration code.
