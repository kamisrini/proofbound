# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-05

## Resume — 2026-11-05

**Exact next action: implement the P7 #2b detector-power shell gate, spec-first, under the
owner-accepted change intent `CI-p7-detector-power-1a2b3c`.** The design spec
[`docs/plans/P7-2b-detector-power-spec-draft1.md`](../docs/plans/P7-2b-detector-power-spec-draft1.md)
is **ACCEPTED** (three-round non-author loop, round-3 ACCEPTABLE `7e29032`), the intent chain is
**owner-accepted** (`830f6b3`: BD-witnessed-regeneration-9c4e21 accepted, BR-regeneration-trust-5f8a03
active, CI-p7-detector-power-1a2b3c accepted), and the owner's §6 decisions are fixed: **shell gate**
(not a kernel package) + **measure-first provisional `U=0`**. Build it:
- `scripts/detector-power.sh <pkg>` + a `scripts/tests/` self-test whose controls prove each §5
  invariant DP-1..DP-8 is load-bearing (gate keys on the ABSOLUTE class-U count vs a per-tier
  `U_BUDGET`, top tier `U=0`; `detector_power=K/(K+U)` is a reported diagnostic only; freshness over
  lock inputs incl. `_test.go` + register + SPEC rows + `tools/mutants` operator-set; raw sweep bound
  to a real execution — reproducible or named independent non-author attested; exact survivor↔record
  bijection; E/R/U classes with R carrying a subsumption argument).
- Extend `docs/allowed-survivors.txt` with per-survivor **class (E/R/U)** + `pins` fields.
- Run the deferred full-package `internal/twin` mutation sweep, classify its survivors, and commit
  the first execution-bound records under `docs/verification/detector-power/`.
- Register an advisory `docs/gates.md` row (Law 4: owner + ISO expiry). Fold CL-1..CL-4.
- This is **governed** work: each commit carries `Intent: CI-p7-detector-power-1a2b3c`; the
  implementation passes a **non-author package verdict** (`make mutants` calibrated + ACCEPTABLE)
  before it counts as accepted (Law 9).

**Immediate blockers / notes:**
- None blocking the above — the capability is authorized. The follow-on **regeneration loop**
  (BR O-2) stays deferred to its own spec (`DEF-P7-REGENLOOP-001`).
- `scripts/p6-census.sh` (NOT a `make check` gate) flags the rejected `kernel/internal/detectorpower`
  token that remains only in the committed round-3 verdict's history — expected, ignore for `make check`.
- `block-secrets` hook: a command that names `docs/decisions/INDEX.md` AND contains `>` (incl. `->`)
  is blocked — stage via directory, keep `->` out of such commands.
- Intent-record self-digest = zero the LAST `"artifact_sha256":"<64hex>"` in the whole file, sha256
  the file; BD→BR→CI references cascade.

## Branch status
`main`, tracking `origin/main`. Pushed through `af0467b` (the B1 batch) earlier today; this session's
later work (`c3ff35d`..`830f6b3`, incl. the completion plan, the #2b spec + 3 verdicts, and the
accepted intent chain) is pushed at the closeout (owner authorized push). Two permanent untracked
paths remain (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) — they keep the
tree dirty between commits, so run `make check` promptly after a commit (commit-cadence else trips).

## Completed (this session)
- **P7 B1 COMPLETE**: unit-3 runtime fixture (TWIN-INV-18..22), real LOCKED+PASS, non-author verdict
  + non-author §9.3a live run; **DEF-P7-001 RESOLVED** ([[VD-p7-def001-resolve-2026-11-05]]).
- **P7 #2b started**: completion plan; detector-power design spec ACCEPTED (3 rounds); intent chain
  (BD/BR/CI) owner-accepted.

## Open / deferred (see `docs/deferrals.tsv`)
- `DEF-KERNELCHECK-FLAKE-001` OPEN/DEFERRED — embedded-postgres port-contention flake in full
  `make check`; unrelated; isolate the offending test to make the gate deterministic.
- `DEF-P7-REGENLOOP-001` OPEN/PENDING_PREREQUISITE — the #2b regeneration loop (BR O-2); begins only
  after the detector-power gate is implemented + accepted and its own follow-on spec is reviewed.
- `DEF-P7-001`/`-002`/`-003`/`DEF-METATAX-001` RESOLVED.

## Unverified / residual
- Detector-power per-tier `U_BUDGET` values beyond the provisional `U=0` are an owner decision to be
  made from the real classified sweeps (with a CL-4 safety argument per tolerated class-U survivor).
- Full-package `internal/twin` mutation sweep still deferred (feeds the detector-power seed records).
- Optional unit-3 LOW cleanups (verdict F1/F2 — each needs a re-verify; unit-2 three cosmetic nits).

## Key facts
- Prepared postgres archive (NOT committed): `~/proofbound-p7-build/postgres-16.4-subset.txz`
  (`PROOFBOUND_P7_PG_ARCHIVE`); digests pin this prefix-specific build.
- Toolchain: Go 1.27.3; `golangci-lint` 2.15.0 at `$(go env GOPATH)/bin` (resolved by
  `kernel-check.sh`); gcc 11.4.0 + bison/flex present. See [[proofbound-wsl-toolchain-gap]].

## Tomorrow's first action
Run `/proofbound-next`, then implement the detector-power shell gate spec-first under
`CI-p7-detector-power-1a2b3c` (see Resume).
