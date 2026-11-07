# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-06

## Resume — 2026-11-06

**Exact next action: obtain a NON-AUTHOR verdict (Law 9) of the P7 #2b detector-power gate as
implemented this session, then (per that verdict) proceed.** The gate, its self-test, the first real
seed record, the `docs/gates.md` advisory row + census/closure coupling, the twin-root deferral, and
the governance intent-linkage reconciliation are all BUILT, self-verified, and committed under
`Intent: CI-p7-detector-power-1a2b3c`. I am the author, so none of it counts as accepted until an
independent verifier runs `/proofbound-review` against the Build Laws + the ACCEPTED spec
(`docs/plans/P7-2b-detector-power-spec-draft1.md`) and produces an ACCEPTABLE verdict. Review focus:
- `scripts/detector-power.sh` realizes DP-1..DP-8 faithfully; the two realization choices the spec
  left to the gate are **(a)** the committed `<slug>.sweep.txt` is `mutants.sh` stdout with the
  host-specific repo-root prefix stripped to kernel-relative (so `raw_sweep_sha256` is host-portable —
  a precondition of DP-8 reproducibility — and keys match `allowed-survivors.txt`), and **(b)** the
  register carries `{dp:class=E|R|U}` which the gate CROSS-CHECKS against the record (Law 2 drift
  guard; `pins` live once, in the record).
- `scripts/tests/detector-power.test.sh` — each DP control is load-bearing AND asserts its intended DP
  tag in the failure message (so a control can't pass by tripping a different check).
- The manifest seed's **E→R reclassification**: the 5 survivors are R (redundantly-pinned), proven by
  the subsuming Validate-layer lines being killed in the same sweep (not among the survivors).

**Immediate blockers / notes:**
- **GOVERNANCE DEBT reconciled** (owner decision 2026-11-06, intent-linkage):
  `docs/verification/governance/2026-11-06-governed-without-intent-reconciliation.md` links `daf9fe4`
  + `6b04f0a` to their authorizing work. Law 7 follow-on (owner): range-level intent-trailer
  enforcement is a SEPARATE enforcement-design question — `delivery-enforce` is HEAD-only by design —
  not closed here.
- This session's governed commits carry `Intent: CI-p7-detector-power-1a2b3c`; **run full `make check`
  (not just `kernel-check`) after any governed commit** — `delivery-enforce` + `hooks-test` only run
  in the full gate.
- `scripts/tests/p6-census-current.test.sh` is a `make check` gate: a `scripts/*` / `kernel/internal/*`
  backtick token in a census-scanned file (CLAUDE/README/ROADMAP/state.md/gates.md + accepted VDs +
  SPECs) must resolve to a tracked path. `scripts/detector-power.sh` now EXISTS, so it is safe to name.
- Adding ANY `docs/gates.md` row forces: a matching C2 row in `docs/plans/p6-census-rows.tsv`
  (`gate-row:<exact Check cell>`), a `p6-task1-close.sh` estate-count bump, and regen of
  `docs/plans/p6-census.md` + `docs/verification/p6-c1-c2-closure.md`.
- `block-secrets` hook: a command naming `docs/decisions/INDEX.md` AND containing `>` (incl. `->`) is
  blocked — stage via directory, keep `->` out of such commands.

## Branch status
`main`, tracking `origin/main`. This session's detector-power work is committed at **`3f446eb`**
(`Intent: CI-p7-detector-power-1a2b3c`); **bare `make check` GREEN at that HEAD** (hooks-test +
kernel-check, 0 issues). NOT pushed — push is the owner's closeout step. Two permanent untracked paths
remain (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) — they keep the tree dirty
between commits, so run `make check` promptly after a commit (commit-cadence else trips).

## Completed (2026-11-06)
- **P7 #2b detector-power gate (O-1) BUILT + self-verified**: `scripts/detector-power.sh` +
  `scripts/tests/detector-power.test.sh` (DP-1..DP-8 + Law-2 drift guard, all load-bearing); first
  real reproducible seed `docs/verification/detector-power/internal-twin-manifest.{sweep.txt,json}`
  (K=330, U=0, trusted at tier top); `allowed-survivors.txt` `{dp:class}` extension + E→R
  reclassification; advisory `docs/gates.md` row + census/closure coupling; governance reconciliation.
- Owner decisions 2026-11-06: intent-linkage reconciliation; manifest-now/twin-root-deferred; classify
  manifest survivors as R.

## Open / deferred (see `docs/deferrals.tsv`)
- `DEF-P7-DP-TWINROOT-001` OPEN/DEFERRED — the heavy postgres-dependent full-package `internal/twin`
  (root) detector-power sweep; needs a reproducible-or-attested record + non-author verdict.
- `DEF-P7-REGENLOOP-001` OPEN/PENDING_PREREQUISITE — the #2b regeneration LOOP (BR O-2); its own spec,
  after the detector-power gate is non-author accepted.
- `DEF-KERNELCHECK-FLAKE-001` OPEN/DEFERRED — embedded-postgres port-contention flake in full
  `make check`; unrelated; isolate the offending test to make the gate deterministic.
- `DEF-P7-001`/`-002`/`-003`/`DEF-METATAX-001` RESOLVED.

## Unverified / residual
- Non-author verdict of the detector-power work is OUTSTANDING (see Resume).
- Detector-power per-tier `U_BUDGET` beyond provisional `U=0` is an owner decision from the real
  classified sweeps (CL-4 safety argument per tolerated class-U survivor).
- `internal/twin` (root) full-package sweep still deferred (DEF-P7-DP-TWINROOT-001).

## Key facts
- Toolchain: Go 1.27.3; `golangci-lint` 2.15.0 at `$(go env GOPATH)/bin` (resolved by
  `kernel-check.sh`); gcc 11.4.0 + bison/flex present. See [[proofbound-wsl-toolchain-gap]].
- Detector-power record refresh (reproducible mode): `scripts/detector-power.sh <pkg> --refresh`
  re-runs the sweep and asserts it reproduces `raw_sweep_sha256`. `mutants.sh` exits 1 when survivors
  exist (normal) — the gate accepts exit <=1 and rejects a real tool error (exit 2).
- Intent-record self-digest = zero the LAST `"artifact_sha256":"<64hex>"` in the whole file, sha256
  the file; BD→BR→CI references cascade.

## Next session's first action
Run `/proofbound-next`, then a non-author `/proofbound-review` of the detector-power gate + self-test +
the manifest-R classification against the Build Laws and the ACCEPTED spec (see Resume).
