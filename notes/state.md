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
- **③ regeneration-loop SPEC (`DEF-P7-REGENLOOP-001`, BR O-2) — AUTHORIZED, next action.** Both
  resume conditions are met (detector-power gate non-author accepted + owner authorization). Exact
  next action: author `docs/plans/P7-2b-regeneration-loop-spec-draft1.md` (a design spec in the
  detector-power-spec mold, NOT kernel code — Law 6), building on the accepted behavior-lock planning
  (`P7-behavior-locks-regeneration-plan-draft1.md`) and discharging the detector-power spec §7
  dependencies: (a) the mutation-operator failure distribution must plausibly represent the
  regenerator's failure modes; (b) the loop must not overfit/train against the measured tests. Then
  non-author review rounds, then implementation under a NEW owner-approved change intent.
- **④ twin-root detector-power sweep (`DEF-P7-DP-TWINROOT-001`) — AUTHORIZED, heavy.** Full-package
  `internal/twin` (root, embedded-postgres) mutation sweep → a reproducible-or-attested execution-bound
  record + classified survivors + a non-author verdict. Heavy compute; subject to ⑤'s flake.
- **⑤ kernel-check flake (`DEF-KERNELCHECK-FLAKE-001`) — AUTHORIZED; ROOT-CAUSED 2026-11-09.** The
  flake is embedded-postgres **fixed-port 55432 contention**: `kernel/internal/store/store.go:177`
  `embeddedPort(0)` returns a FIXED `55432` (asserted by `kernel/internal/store/surface_test.go:75`);
  `internal/migration` tests that use the default port collide under back-to-back server startup
  (`TestImportRefusesNonEmptyLedger`: "process already listening on port 55432"). Confirmed a flake —
  the package PASSES on isolated retry. Fix is a GOVERNED kernel change (`kernel/*`) → needs its own
  change intent + non-author verdict; the deferral's intended remedy is to isolate/serialize the
  offending test or give it a free ephemeral port without breaking the `surface_test` port contract.

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
Run `/proofbound-next`, then author `docs/plans/P7-2b-regeneration-loop-spec-draft1.md` (③) unless the
owner resequences toward ④ (twin-root sweep) or ⑤ (flake kernel fix). F1's non-author re-confirmation
remains outstanding and can be dispatched to a non-author lens at any time.
