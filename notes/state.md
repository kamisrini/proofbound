# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-09

## Resume — 2026-11-09

**Owner authorized (2026-11-09) proceeding on all five items surfaced by `/proofbound-next`.** Status
and exact next action per item:

- **① F1 (seed-citation fix) — DONE.** Verdict `p7-2b-detector-power-impl-review-round1-20261109`
  finding F1 corrected: the seed record and the governed `docs/allowed-survivors.txt` register no
  longer cite mutant-free lines (`manifest.go:299`, `:273`, the required-field loop) as kill
  evidence. Classifications unchanged (all R, `U=0`, trusted). Committed `25d13ad` with
  `Intent: CI-p7-detector-power-1a2b3c` (within that accepted intent's allowed scope); record
  `lock_inputs_sha256` re-pinned; gate re-verified green. **OUTSTANDING (Law 9):** a non-author
  re-confirmation of this prose fix — I authored both the finding and the fix, so it does not count
  as accepted until a non-author re-verifies. The fix is citation-accuracy only (no reclassification).
- **② state refresh — DONE** (this file).
- **③ regeneration-loop SPEC (`DEF-P7-REGENLOOP-001`, BR O-2) — DRAFT 1 AUTHORED.**
  `docs/plans/P7-2b-regeneration-loop-spec-draft1.md` committed `afbeb20` (design-only, Law 6). It
  defines the repair cycle (admit counterexample as a NEW lock revision → re-measure detector power
  on the new lock, NO inheritance → re-derive in isolation → verify against the lock → accept iff
  budget-clear ∧ full-lock-pass ∧ isolation-attested), RL-1..RL-8, and makes the two detector-power
  §7 circularity dependencies first-class (D-a operator representativeness, D-b no-overfitting). §5
  owner decisions left OPEN. **Next action: a NON-AUTHOR review of this draft** (I authored it — Law 9),
  then owner settles §5, then implementation under a NEW change intent.
- **⑤ kernel-check flake (`DEF-KERNELCHECK-FLAKE-001`) — ROOT-CAUSED; do BEFORE ④.** Embedded-postgres
  **fixed-port 55432 contention**: `kernel/internal/store/store.go:177` `embeddedPort(0)` returns a
  FIXED `55432` (an intentional product contract — stable local CLI DB port — asserted by
  `kernel/internal/store/surface_test.go:75`); `internal/migration` tests that use the default port
  collide under back-to-back server startup (`TestImportRefusesNonEmptyLedger`: "process already
  listening on port 55432"). Confirmed a flake — the package PASSES on isolated retry. **Fix is a
  GOVERNED kernel change** (`kernel/*`, incl. `_test.go`) → needs an owner-approved change intent + a
  non-author verdict, AND an owner decision on approach (recommended: give the embedded-postgres
  *tests* a free ephemeral port instead of the default 0→55432, leaving the product's fixed-port
  contract and `surface_test` intact; alternatives: serialize startups on a cross-process lock, or
  change the default). **Sequence ⑤ before ④:** the twin-root sweep spins up many embedded-postgres
  instances, so the fixed-port flake would make that heavy sweep unreliable.
- **④ twin-root detector-power sweep (`DEF-P7-DP-TWINROOT-001`) — AUTHORIZED, heavy; AFTER ⑤.**
  Full-package `internal/twin` (root, embedded-postgres) mutation sweep → a reproducible-or-attested
  execution-bound record + classified survivors + a non-author verdict. Heavy compute; reliability
  depends on ⑤.

**Gate state (2026-11-09):** full `make check` at `25d13ad` passed `hooks-test` (incl.
`detector-power.test`) + `delivery-enforce` (F1's `Intent` trailer accepted) + link/freshness; it
FAILED `kernel-check` ONLY on the `DEF-KERNELCHECK-FLAKE-001` flake above, which passed on isolated
retry. So the gate is green modulo the known flake; F1 is not implicated.

**Immediate blockers / notes (unchanged, still live):**
- This session's governed commits carry `Intent: CI-p7-detector-power-1a2b3c`; **run full `make check`
  after any governed commit** — `delivery-enforce` + `hooks-test` only run in the full gate.
- Adding ANY `docs/gates.md` row forces a matching C2 row in `docs/plans/p6-census-rows.tsv`
  (`gate-row:<exact Check cell>`), a `p6-task1-close.sh` estate-count bump, and regen of
  `docs/plans/p6-census.md` + `docs/verification/p6-c1-c2-closure.md`.
- `scripts/tests/p6-census-current.test.sh`: a `scripts/*` / `kernel/internal/*` backtick token in a
  census-scanned file (CLAUDE/README/ROADMAP/state.md/gates.md + accepted VDs + SPECs) must resolve
  to a tracked path.
- `block-secrets` hook: a command naming `docs/decisions/INDEX.md` AND containing `>` (incl. `->`) is
  blocked — stage via directory, keep `->` out of such commands.

## Branch status
`main`, tracking `origin/main`. Pushed through `b8dc948` (the non-author verdict). Unpushed:
`25d13ad` (F1 fix) and this state commit — push is the owner's closeout step. Two permanent untracked
paths remain (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`).

## Completed (2026-11-09)
- **Non-author verdict of the P7 #2b detector-power gate (O-1):** `ACCEPTABLE-WITH-NOTES`,
  `docs/verification/verdicts/p7-2b-detector-power-impl-review-round1-20261109.md` (`b8dc948`). Gate +
  self-test + `U=0` conclusion accepted; reviewer independently re-ran the full sweep (`--refresh`) and
  reproduced `raw_sweep_sha256` exactly (closing the fabricated-sweep gap by live reproduction); all 5
  survivors traced to genuine downstream checks. One open finding F1 (now fixed, re-confirm pending).
- **F1 fix (`25d13ad`).**
- **`DEF-KERNELCHECK-FLAKE-001` root-caused** (fixed-port 55432 contention; see ⑤).

## Completed (2026-11-06)
- **P7 #2b detector-power gate (O-1) BUILT + self-verified**: `scripts/detector-power.sh` +
  `scripts/tests/detector-power.test.sh` (DP-1..DP-8 + Law-2 drift guard); first real reproducible
  seed `docs/verification/detector-power/internal-twin-manifest.{sweep.txt,json}`; `allowed-survivors`
  `{dp:class}` extension; advisory `docs/gates.md` row + census/closure coupling.

## Open / deferred (see `docs/deferrals.tsv`)
- `DEF-P7-REGENLOOP-001` OPEN — prerequisite now MET; owner-authorized; next action is the spec (③).
- `DEF-P7-DP-TWINROOT-001` OPEN — owner-authorized; heavy (④).
- `DEF-KERNELCHECK-FLAKE-001` OPEN — owner-authorized; root-caused (⑤); governed kernel fix pending.
- `DEF-P7-001`/`-002`/`-003`/`DEF-METATAX-001` RESOLVED.

## Unverified / residual
- **Non-author re-confirmation of F1's prose fix is OUTSTANDING** (Law 9; author==fixer here).
- Detector-power per-tier `U_BUDGET` beyond provisional `U=0` is an owner decision from the real
  classified sweeps.

## Key facts
- Toolchain: Go 1.27.3; `golangci-lint` 2.15.0 at `$(go env GOPATH)/bin`; gcc 11.4.0 present.
- Detector-power refresh (reproducible): `scripts/detector-power.sh <pkg> --refresh` re-runs the sweep
  and asserts it reproduces `raw_sweep_sha256`. `mutants.sh` exits 1 when survivors exist (normal) —
  the gate accepts exit <=1 and rejects a real tool error (exit 2).

## Next session's first action
Run `/proofbound-next`. Open threads from 2026-11-09, all owner-authorized: (a) NON-AUTHOR review of
the regen-loop spec draft (③, `afbeb20`); (b) ⑤ flake fix — owner picks the approach (recommended:
free ephemeral port for the embedded-postgres tests), then scribe a change intent + implement +
non-author verdict; (c) ④ twin-root sweep AFTER ⑤. Also outstanding: F1's non-author re-confirmation
(Law 9), dispatchable to a non-author lens at any time.
