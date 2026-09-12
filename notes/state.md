# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-12 (P6 draft 2 awaiting founder ratification; implementation not authorized)

## Exact resume point

P5 is accepted and must not be restarted. The off-machine P6 consolidation draft 1 and the
build-machine non-author `NEEDS_WORK` adjudication are preserved in `ceeae16`. Draft 2 folds all ten
findings and is awaiting exactly four founder inputs: phase number, the closed path-based intent
applicability rule (including its canary tolerance), snapshot-provider feed or explicit skip, and a
countable stop-early effort ceiling. STOP here: do not mint the semantic VD, amend `ROADMAP.md`, or
begin Task 0 until those four inputs are ratified and recorded.

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

- The tracked changes are P5 hardening work: fail-closed tests and simplifications in intent,
  records, specdir, reviews, projections, gates, CLI, delivery enforcement, invariant lock, and the
  calibrated mutation harness.
- `docs/verification/p5-independent-verifier-handoff.md` records the frozen reviewed commit and
  independent-verifier instructions.
- `docs/plans/P5-intent-provenance-plan.md` is an older untracked draft with digest prefix
  `eacb`; it is not the committed adjudicated exhibit (digest prefix `9d203`). Treat it as
  pre-existing user material and do not delete or commit it without resolving its ownership.
- P6 authority is currently draft-only: `docs/plans/P6-CONSOLIDATION-PLAN-draft2.md` plus
  `docs/verification/verdicts/p6-consolidation-plan-round1.md`. No P6 execution task is authorized.
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
