# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-09

## Resume — 2026-11-09

**Owner authorized (2026-11-09) proceeding on all five `/proofbound-next` items; all five advanced.**
Status:

- **① F1 (detector-power seed citations) — DONE + verdict schema corrected.** Citations fixed
  (`25d13ad`). The round-1 verdict was ALSO corrected to conform to `vera.verdict.v1` (`9ee6c04`):
  the reviews connector accepts status only `{ACCEPTABLE, NEEDS_WORK}` and finding fields only
  `{finding_id, severity∈HIGH|MED|LOW, defect_commit}`; the verdict had used `ACCEPTABLE-WITH-NOTES`
  + `id/status/summary` and was rejected by `sync`. Now status `ACCEPTABLE`, findings `[finding_id:
  F1, severity: LOW]`, self-digest re-pinned, `p6-artifact-integrity --check` green. **Outstanding
  (Law 9):** non-author re-confirmation of the F1 fix (I authored finding + fix).
- **② state refresh — DONE** (this file).
- **③ regeneration-loop SPEC (`DEF-P7-REGENLOOP-001`, BR O-2) — DRAFT 1 DONE** (`afbeb20`,
  `docs/plans/P7-2b-regeneration-loop-spec-draft1.md`). **Outstanding:** non-author review (I authored
  it); then owner §5 decisions; then implementation under a new change intent.
- **⑤ kernel-check flake (`DEF-KERNELCHECK-FLAKE-001`) — IMPLEMENTED; make check GREEN + deterministic.**
  Root cause: embedded-postgres fixed-port 55432 contention. Fix (`057bc07`, `Intent:
  CI-flake-ephemeral-port-7d5f10`): a `freeTestPort(t)` helper (listen `:0`, mirrors
  `twin.temporaryPort`) in the four offender packages — `internal/migration` (make-check path),
  `internal/projections` / `internal/store` / `internal/gates` (`//go:build integration`) — passing
  `Port: freeTestPort(t)` at each embedded-store call site. Product fixed-port contract +
  `store/surface_test.go` untouched. Verified: migration `-count=3` clean, `go vet` clean (default +
  `-tags integration`), **full `make check` green (0 issues), no 55432 contention**. **Outstanding
  (Law 9):** non-author package verdict of the kernel change before it counts as accepted.
  **Governance (owner chose recommendation (a) 2026-11-09):** authored the intent chain (`98cdeea`) —
  `BD-self-hosted-gate-integrity-9a7c31` (accepted) → `BR-gate-integrity-4b2e08` (proposed; O-1 = the
  gate is a deterministic witness) → `CI-flake-ephemeral-port-7d5f10` (accepted). This gives governed
  maintenance a real intent home, closing the governed-without-intent gap for future infra fixes.
- **④ twin-root detector-power sweep (`DEF-P7-DP-TWINROOT-001`) — now UNBLOCKED (⑤ done), AUTHORIZED,
  heavy.** Full-package `internal/twin` (root, embedded-postgres) mutation sweep → a
  reproducible-or-attested execution-bound record + classified survivors + a non-author verdict.

**⚠ NEW DISCOVERY — `delivery-enforce` / `proofbound sync` is RED on PRE-EXISTING verdicts.** `make
check` does NOT run `delivery-enforce` (not in the `check` target), so the local gate is green, but
`make delivery-enforce` runs `sync all`, and the reviews connector rejects many committed verdicts:
status `ACCEPTABLE-WITH-NOTES` (×2) and `NEEDS_CHANGES` (×9) are not in `{ACCEPTABLE, NEEDS_WORK}`,
and several carry `summary`/`verdict`/`id` finding fields the connector forbids. My own verdict is now
conformant; the rest is a SEPARATE cleanup (not ⑤, not the intent work). Until it is done,
`delivery-enforce` cannot fully validate the ⑤ intent binding end-to-end (it aborts at the first
non-conforming verdict). The ⑤ intent records themselves are well-formed (they parsed; `sync` failed
only at the reviews connector).

**Immediate blockers / notes (still live):**
- Governed `kernel/*` commits carry an `Intent:` trailer; governance is enforced by
  `delivery-enforce` (NOT `make check`). `make check` = `hooks-test commit-cadence state-freshness …
  kernel-check` (Makefile line 3); `delivery-enforce` (sync + gates enforce) is a separate target.
- `kernel-check` runs `go test ./... -count=1 -v` (NO `-tags integration`, no `-short`): only
  `internal/migration` opens embedded postgres under `make check`; the other three ⑤ packages are
  integration-tagged (delivery/CI + the ④ sweep).
- Adding a `docs/gates.md` row forces a matching C2 census row + `p6-task1-close.sh` bump + regen of
  `p6-census.md` + `p6-c1-c2-closure.md`.
- `block-secrets` hook: a command naming `docs/decisions/INDEX.md` AND containing `>` is blocked.

## Branch status
`main`, tracking `origin/main`. Pushed through `eda55c4`. Unpushed: `9ee6c04` (verdict schema),
`98cdeea` (intent chain), `057bc07` (⑤ kernel fix), and this state commit — push is the owner's
closeout step. Two permanent untracked paths remain (`docs/plans/P5-intent-provenance-plan.md`,
`kernel/cmd/p6dbprobe/`).

## Completed (2026-11-09)
- Non-author verdict of the detector-power gate (O-1): `ACCEPTABLE` + finding F1 (`b8dc948`, schema
  corrected `9ee6c04`); reviewer independently reproduced the sweep via `--refresh`.
- F1 fix (`25d13ad`). Regen-loop spec draft 1 (`afbeb20`).
- ⑤ flake root-caused, intent chain authored (`98cdeea`), and FIXED (`057bc07`); `make check` green.
- Recommendation (a) executed: gate-integrity BD/BR give governed maintenance an intent home.

## Open / deferred (see `docs/deferrals.tsv`)
- `DEF-KERNELCHECK-FLAKE-001` OPEN — IMPLEMENTED (`057bc07`), make check green; stays open pending the
  non-author kernel verdict (Law 9), then resolve.
- `DEF-P7-REGENLOOP-001` OPEN — spec draft 1 done; pending non-author spec review + owner §5.
- `DEF-P7-DP-TWINROOT-001` OPEN — now unblocked; heavy (④).
- `DEF-P7-001`/`-002`/`-003`/`DEF-METATAX-001` RESOLVED.

## Unverified / residual (Law 9 — each needs a NON-author lens)
- F1 fix re-confirmation; regen-loop spec draft review; ⑤ kernel-fix package verdict.
- `delivery-enforce` verdict-schema cleanup (pre-existing, repo-wide) — separate task.

## Key facts
- Toolchain: Go 1.27.3; `golangci-lint` 2.15.0 at `$(go env GOPATH)/bin`; gcc present.
- Intent-record self-digest: zero the record's OWN (last) `"artifact_sha256"` in the whole file,
  sha256 the file; cross-refs cascade BD→BR→CI. Verdict self-digest: zero the `artifact_sha:` line.
- Detector-power refresh: `scripts/detector-power.sh <pkg> --refresh` re-runs the sweep and asserts it
  reproduces `raw_sweep_sha256`; `mutants.sh` exits 1 when survivors exist (gate accepts exit <=1).

## Next session's first action
Run `/proofbound-next`. Open, all Law-9-pending or owner-steered: (a) dispatch NON-author lenses for
the F1 fix, the regen-loop spec, and the ⑤ kernel fix; (b) ④ twin-root sweep (now unblocked); (c) the
`delivery-enforce` pre-existing verdict-schema cleanup. Push remains the owner's closeout step.
