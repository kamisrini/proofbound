# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads `CLAUDE.md`, then this file, before acting.

**As of:** 2026-09-11 (P5 Task 8 acceptance hardening in progress)

## Exact resume point

P5 Tasks 0 through 8 are implemented through commit `01f41d6` (`docs: bind P5 self-hosting
delivery`). The ratified plan, received adjudication/exhibit/founder record, baseline evidence,
semantic VD, roadmap DoD, Proofbound identity migration, schemas/providers, commit binding,
projections/reports, verdict/review contracts, gates, and self-host chain are already durable in the
preceding P5 commits. The current worktree is the uncommitted final hardening and acceptance packet.

Do not restart P5. Resume with these steps:

1. Finish the three integration-tagged mutation sweeps already being hardened:
   `internal/projections`, `internal/gates`, and `internal/cli`.
2. Record all final calibrated green counts in
   `docs/verification/p5-mutation-evidence.md`.
3. Demonstrate that the promoted `intent-delivery-readiness` gate BLOCKS an intentionally bad
   self-hosted chain at the real `make delivery-enforce` boundary and record the isolated,
   reproducible result in `docs/verification/p5-delivery-readiness.md`.
4. Run bare `make check`, freeze the author implementation/evidence commit with trailer
   `Intent: CI-implement-p5-0a1b2c`, then request a genuinely non-author requirement review and v2
   obligation verdict against that exact commit.
5. Commit both independent artifacts verbatim on receipt under `docs/verification/verdicts/`, sync
   the complete chain, run `proofbound verify`, the intent report, and the exact commit intent check,
   then close the roadmap/state/journal without rewriting received verdicts.

## Active mutation evidence

Already green, with calibrated positive/neutral controls and no survivors:

- `internal/connector/intent`: 93/93 killed.
- `internal/connector/intent/records`: 52/52 killed.
- `internal/connector/intent/specdir`: 47/47 killed.
- `internal/connector/git`: 34/34 killed.
- `internal/connector/reviews`: 114/114 killed.

Projection continuation began at global candidate 127 after adding a direct test for the prior
`isJSONColumn` survivor at `projections.go:975#69`; candidates through the resumed early range are
being killed. Gates and CLI need their final counts. The mutation runner now exports the original
repository root to scratch-tree integration tests because the gates package reads the real
top-level gate definitions; its clean integration suite passes.

Disposable PostgreSQL instances from the resumed run listen on localhost ports 55434
(projections), 55435 (gates), and 55436 (CLI). They are temporary test infrastructure only.

## Worktree ownership and cautions

- The tracked changes are P5 hardening work: fail-closed tests and simplifications in intent,
  records, specdir, reviews, projections, gates, CLI, delivery enforcement, invariant lock, and the
  calibrated mutation harness.
- `docs/verification/p5-independent-verifier-handoff.md` is intentional untracked P5 work and must
  be completed with the frozen reviewed commit before the independent handoff.
- `docs/plans/P5-intent-provenance-plan.md` is an older untracked draft with digest prefix
  `eacb`; it is not the committed adjudicated exhibit (digest prefix `9d203`). Treat it as
  pre-existing user material and do not delete or commit it without resolving its ownership.
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
