# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-10-14

## Resume — 2026-10-14

Exact next action: obtain a fresh non-author round-2 review of the amended exact revision of
`docs/plans/P7-replay-lock-spec-addendum-draft1.md` and
`docs/plans/P7-replay-lock-fixture-matrix-draft1.md`; record `ACCEPT`, `AMEND`, or `REJECT` before
any behavior-lock, regeneration, snapshot-provider, or external-feed implementation.

Immediate blockers: the round-1 addendum/fixture review outcome was `AMEND`; the amended revision
has no round-2 verdict. The full gate cannot complete because the installed linter crashes on a
Go 1.27 version marker. Committed runtime fixture corpus, projector revision evidence, and
implementation verdict are absent. Push
authorization/authentication remain unverified after the prior GitHub DNS failure.

## Branch and repository status

- Branch: `main`, tracking `origin/main`, ahead of the locally known remote-tracking ref.
  The remote tip is unverified.
- Preserved untracked user artifacts: `docs/plans/P5-intent-provenance-plan.md` and
  `kernel/cmd/p6dbprobe/main.go`. They are untouched and excluded from this checkpoint.
- The amended P7 addendum and fixture matrix remain draft planning artifacts until independently
  reviewed; no product implementation is authorized.

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
- Fresh non-author reviewer `Fermat` judged the exact active P5 `O-1` `VERIFIABLE`; the P7 addendum
  and fixture planning slice received `AMEND` with one high and three medium findings. The
  requirement review and documentary slice verdict are recorded under `docs/verification/verdicts/`.
- Amended the addendum and fixture matrix for all four round-1 findings: complete event evidence
  beyond P4's source digest, whole-ledger before/after witnesses on success and failure, pinned
  projector/build/runtime and proof evidence encodings with a byte-exact proof control, and
  independent candidate payload-hash verification. These are proposed contracts, not accepted
  behavior or runtime fixtures.
- The generated P6 requirement-review packet now includes both active requirements, three active
  obligations, and two exact reviews. The generated P6 artifact-integrity packet now enumerates
  26 schema-bearing verdicts and 12 documentary verdict artifacts.

## Verification

- Bare `make check` ran on 2026-10-14. The `index stale; run make index` output was an expected
  negative control; generated freshness passed. The hook loop's exit status is not sufficient to
  claim every self-test passed, as discovered later. The actual stop was
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
- For the amended P7 draft, `git diff --check`, `make link-lint`, `make index-check`,
  `scripts/p6-task5-review.sh --check`, and `scripts/p6-artifact-integrity.sh --check` passed.
  The proof encoding control hashed to its documented SHA-256. After commit `210830e`, bare
  `make check` passed repository/generated checks but stopped at `kernel-check` because sandboxed
  Snap Go lacks `cap_dac_override`. With the installed linter on PATH outside that confinement,
  the same gate passed repository checks and Go build/tests, then the linter panicked on a file
  marked Go 1.27 while it was built with Go 1.26.7. The `index stale` line was an expected
  negative-test message, not a freshness failure. Full lint remains unverified.

## Open, blocked, and unverified

- Open: obtain another non-author planning review of the amended exact revision; ratify the
  proposed proof/projector evidence bytes; build committed runtime fixtures and evidence manifest; obtain a
  separate non-author review of any later implementation. The `hooks-test` Makefile loop can
  mask a failing intermediate self-test because it returns only the last test's status; repair
  this independently of the P7 contract work.
- Blocked: behavior-lock, regeneration, snapshot-provider, and external-feed implementation until
  the applicable review and decision gate is satisfied. Push remains unverified.
- Unverified: satisfaction of the new requirement, complete clean-clone replay, detector power
  for a P7 mutation set, lint with a compatible toolchain, native Windows
  acceptance, and remote-tip freshness.

## Institutionalized improvement

When preserved user-owned untracked artifacts trigger cadence, keep them untouched, record their
ownership, commit only a coherent documentation checkpoint, and rerun the full gate. For replay
evidence, check both complete event fields and complete-ledger preservation independently of
existing P4 digests. No new automation is justified.

## Standing cautions

- Preserve frozen `vera.witness.v1`, `vera.verdict.v1`, `vera.replay.v1`, pinned vectors, and
  historical artifacts.
- Planning acceptance and draft review do not authorize behavior-lock or regeneration code.
