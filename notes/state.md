# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-12 (P6 Task 1 in progress; legacy-storage decision required)

## Exact resume point

P5 is accepted and must not be restarted. The founder ratified all four recommended P6 inputs:
P6 is consolidation/deep completion; the exact draft-2 path predicate and zero-false-negative /
greater-than-10%-false-positive threshold govern; the snapshot provider is skipped until P7+; and
the stop-early ceiling is 40 active execution hours after Task 0. The record is durable in
`28d8716`; the semantic VD and roadmap authorization are durable in `708ad12`.

Task 0 followed spec-first order. The census SPEC is `1ba2336`; the closed scanner and hostile
fixtures are `7046b42` plus discovery-universe tightening in `7bb9a66`. The first generated census
contains 188 classified rows across C1–C8, with zero unclassified subjects. Task 1's mechanism
contract is durable in `cce0a5d`. Its first implementation slice is `dd6154e`: `make short` now runs
the short Go suite, and cadence, state freshness/rendering, laws-lock generation, meta-tax, explicit
bundle backup, and wrap verification exist with hostile tests. Commits `7fd328d`, `7d840ba`, and
`4de1439` then restored cleanroom, lesson recurrence, figure provenance, skip/prescription lint,
skip-aware kernel checking, the four Claude hooks, and blocking exact package-SPEC citation
resolution. The full host gate passed after each code-bearing slice. The census is now 160 open / 28
closed.

STOP at the legacy-alias row before changing storage. The only local ledger is the migrated embedded
cluster `.proofbound/db`, whose private identity marker is `vera-v1`. The accepted P5 migration
document says this private compatibility identity is removed with the aliases at P6 start, but
removing support makes the existing ledger inaccessible. Founder choice required: either
(recommended) retain this private on-disk identity as frozen migration compatibility while removing
all external CLI/environment aliases, or authorize a one-time in-place PostgreSQL role/database
migration before removing it.

Current verification results:

- `git log -1 --oneline`: the documentation close-out commit immediately after `f426ca8`.
- Bare `PATH=/home/thamm/go/bin:$PATH make check`: exit 0; all packages and `golangci-lint: 0 issues`.
  The sandbox-only invocation still fails before Go starts because Snap lacks `cap_dac_override`;
  host execution is required on this machine.
- `make verify`: exit 0 against the moved `.proofbound` ledger.
- `make delivery-enforce`: exit 0; all configured gates PASS, including
  `intent-delivery-readiness` at proof event `01M29HMPE5V977AR3VMW47DVDE` (seq 1834); the bad-chain
  BLOCKED control remains recorded.
- Intent report for `CI-implement-p5-0a1b2c`: `SATISFIED`; `O-1` and `O-2` are independently
  reviewed `VERIFIABLE` and `SATISFIED`.
- Requirement report for `BR-intent-chain-d4e5f6`: exact active revision, declared authorization.
- Exact commit check for reviewed commit `c29bb3b`: resolves to the exact CI revision and proof
  event `01M29HMPE5V977AR3VMW47DVDE/1834`.

## Active mutation evidence

Already green, with calibrated positive/neutral controls and no survivors:

- `internal/connector/intent`: 93/93 killed.
- `internal/connector/intent/records`: 52/52 killed.
- `internal/connector/intent/specdir`: 47/47 killed.
- `internal/connector/git`: 34/34 killed.
- `internal/connector/reviews`: 114/114 killed.

Final mutation evidence is recorded in `docs/verification/p5-mutation-evidence.md`: all connector,
projection, gate, and CLI candidates were killed with calibrated controls; focused integration
suites pass. The mutation runner exports the original repository root for scratch-tree fixtures.

## Worktree ownership and cautions

- The P6 work queue is `docs/plans/p6-census.md`; its stable classified source is
  `docs/plans/p6-census-rows.tsv`. Never hand-edit the generated Markdown result; regenerate it with
  `scripts/p6-census.sh --write` and prove freshness with `--check`.
- Task 1 is only partially complete. Do not treat the presence of its closure-artifact pathname in
  open-row probes as evidence; `docs/verification/p6-c1-c2-closure.md` does not exist yet and may be
  created only after its checker proves every C1/C2 row.
- No storage mutation or alias deletion has been performed while the private legacy-identity choice
  is unresolved.
- `docs/plans/P5-intent-provenance-plan.md` is an older untracked draft with digest prefix
  `eacb`; it is not the committed adjudicated exhibit (digest prefix `9d203`). Treat it as
  pre-existing user material and do not delete or commit it without resolving its ownership.
- P6 authority is `docs/decisions/VD-p6-consolidation-2026-09-12.md`, citing the founder record,
  round-1 adjudication, and exact draft-2 digest. Wider vision capability remains P7+.
- Preserve the frozen wire identities `vera.witness.v1`, `vera.verdict.v1`, and `vera.replay.v1`,
  and all pinned vector bytes. Live product identity is Proofbound.
- Run `make check` bare. Use `PATH=/home/thamm/go/bin:$PATH` so the installed linter is found. The
  Snap-provided Go toolchain and local PostgreSQL sockets can require host approval.
- Do not claim package acceptance until mutation sweeps are green and the non-author artifacts are
  committed verbatim. Do not push unless separately authorized.

## Canonical P5 authority and chain

- Plan: `docs/plans/P5-PROOFBOUND-INTENT-PROVENANCE-v3.md`.
- Semantic VD: `docs/decisions/VD-p5-intent-provenance-2026-09-10.md`.
- Founder record and adjudication:
  `docs/verification/verdicts/p5-founder-ratification.md` and
  `docs/verification/verdicts/p5-adjudication-round1.md`.
- BD: `intent.records:BD-proofbound-p5-a1b2c3@810c2932ed868ada9c0b9da389dc0c89880b1a6c658ba29fce5ab8aefc32745c`.
- BR: `intent.records:BR-intent-chain-d4e5f6@913d6751b8a3717254beda42ba889061fc2d247fc6a1d8505c69f1c27c314562`.
- CI: `intent.records:CI-implement-p5-0a1b2c@f90b216573f3f88eee61029f81d180c5ae1dd70dcb919d9662f97cfd76395424`.
- Requirement obligations: `O-1` and `O-2`; requirement author/owner is
  `proofbound-maintainer`, so the declared independent reviewer must differ.

## Standing product laws

`CLAUDE.md` is authoritative. In particular: commits are durability; evidence must be
ledger-backed; package acceptance requires calibrated mutation green plus a non-author verdict;
received verdict artifacts are committed verbatim; and this state file stays live under Law 10.
