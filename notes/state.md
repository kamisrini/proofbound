# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-09

## Resume — 2026-11-09

**All five `/proofbound-next` items advanced; three non-author reviews (Law 9) completed and committed
(`0d36784`).** `make check` is GREEN at HEAD. Status:

- **① F1 (detector-power seed citations) — DONE + non-author re-confirmed.** Fix `25d13ad`; verdict
  schema corrected `9ee6c04`; **non-author re-confirmation ACCEPTABLE**
  (`verdicts/p7-2b-detector-power-f1-fix-reconfirm-round1-20261109.md`). F1 fully closed.
- **② state — DONE** (this file). **③ regen-loop spec — DRAFT 1 reviewed: NEEDS_WORK.**
  `verdicts/p7-2b-regeneration-loop-spec-review-round1-20261109.md`, 8 findings (F1 HIGH: RL-6 does
  not block satisfy-and-reopen on a NON-ZERO budget, and a counterexample falsifying one class-U
  "safe-to-vary" argument doesn't force re-validation of siblings; MEDs on step-2 mechanism/cost,
  counterexample lethality, D-b false-attestation, lock-level Law-9 separation; LOWs). **Next: author
  DRAFT 2** addressing all 8, then a non-author re-review.
- **⑤ kernel flake fix — DONE; non-author ACCEPTED (round-2); `DEF-KERNELCHECK-FLAKE-001` fixed + accepted, stays OPEN pending owner final close.**
  The ephemeral-port test change (`057bc07`) is correct/in-scope, but it
  shipped a stale `@generated` route matrix (Law 1): `057bc07` touched `kernel/internal/projections`
  and `.../gates`, which advances the derived "Code commit" pin in
  `docs/verification/p6-event-route-matrix.md` (derived via `git log -1 -- projections cli twin
  gates`), but it was not regenerated → `make check` RED at that HEAD. My pre-commit `make check`
  passed only on the working tree (`git log` can't see uncommitted changes). FIXED `8700760`
  (`scripts/p6-route-matrix.sh --write`, one-line pin refresh). Verdicts: round-1 NEEDS_WORK,
  **round-2 ACCEPTABLE** at `caffe71` (`verdicts/p7-flake-ephemeral-port-kernel-round{1,2}-20261109.md`).
  Deferral stays OPEN pending the owner's final close (Law 9: owner is the acceptor); non-author
  accepted. NOTE: `scripts/tests/deferrals.test.sh` hard-codes `DEF-KERNELCHECK-FLAKE-001 OPEN
  DEFERRED` as its `--open`-filter fixture, so the owner's close must also repoint that fixture to
  another OPEN row (governed edit) — a brittle-test coupling worth fixing when it is closed.
- **④ twin-root sweep (`DEF-P7-DP-TWINROOT-001`) — unblocked, AUTHORIZED, heavy; still queued.**

**LESSON (Law 7 candidate): committing a change under `kernel/internal/{projections,cli,twin,gates}`
re-stales `docs/verification/p6-event-route-matrix.md`** (its code-commit pin is derived via
`git log`, which can't see uncommitted work), so a PRE-commit `make check` passes and the POST-commit
gate is red. Always run `scripts/p6-route-matrix.sh --write` after committing such paths (precedent
refreshes `c83c807` / `a9f244f` / this `8700760`). Seen ≥3×; the pre-commit blind spot is structural —
a candidate for a staged-aware check or pre-commit hook.

**Still live / discoveries:**
- **`delivery-enforce` / `sync` is RED on PRE-EXISTING non-conforming verdicts** (status
  `ACCEPTABLE-WITH-NOTES`/`NEEDS_CHANGES`, `summary`/`verdict`/`id` finding fields). `make check` does
  NOT run `delivery-enforce`, so the local gate is unaffected. My own verdicts now conform (status ∈
  {ACCEPTABLE, NEEDS_WORK}, finding fields finding_id/severity). Repo-wide verdict cleanup is a
  SEPARATE task. See [[proofbound-make-check-vs-delivery-enforce]].
- Governance: recommendation (a) executed — gate-integrity `BD`/`BR` + `CI-flake-ephemeral-port-7d5f10`
  (`98cdeea`) give governed maintenance an intent home.

## Branch status
`main` → `origin/main`. Pushed through (to be pushed this session: `8700760`, `0d36784`, this state
commit — the earlier push of `4276179` left origin momentarily RED via the route-matrix slip; the
push of `8700760` restores green). Two permanent untracked paths remain
(`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`).

## Completed (2026-11-09)
- Detector-power gate (O-1) non-author verdict `ACCEPTABLE` + F1 (schema-corrected); F1 fixed AND
  non-author re-confirmed → F1 closed.
- Regen-loop spec draft 1 authored + non-author reviewed (NEEDS_WORK, draft 2 pending).
- ⑤ flake fixed (`057bc07`) under the gate-integrity intent chain (`98cdeea`); route-matrix Law-1 slip
  fixed (`8700760`); `make check` green. Three verdicts committed (`0d36784`).

## Open / deferred (see `docs/deferrals.tsv`)
- `DEF-KERNELCHECK-FLAKE-001` OPEN — fixed + non-author accepted (round-1+2); awaiting owner final close.
- `DEF-P7-REGENLOOP-001` OPEN — spec draft 1 reviewed NEEDS_WORK; draft 2 + re-review pending.
- `DEF-P7-DP-TWINROOT-001` OPEN — unblocked; heavy (④).
- `DEF-P7-001`/`-002`/`-003`/`DEF-METATAX-001` RESOLVED.

## Unverified / residual (Law 9 — each needs a NON-author lens)
- Regen-loop spec draft 2 (after authoring) → non-author re-review.
- `delivery-enforce` pre-existing verdict-schema cleanup (separate, repo-wide).

## Key facts
- `make check` = Makefile `check:` target; does NOT run `delivery-enforce` (sync + gates enforce).
  `kernel-check` = `go test ./... -count=1 -v` (no `-tags integration`).
- Verdict schema (`vera.verdict.v1`, reviews connector): status ∈ {ACCEPTABLE, NEEDS_WORK}; finding
  fields finding_id/severity(HIGH|MED|LOW)/defect_commit only. Self-digest: zero the `artifact_sha:`
  line, hash; validate `scripts/p6-artifact-integrity.sh --check`. Intent-record digests: see
  [[proofbound-intent-record-digests]].
- Detector-power refresh: `scripts/detector-power.sh <pkg> --refresh` reproduces `raw_sweep_sha256`.

## Next session's first action
Run `/proofbound-next`. Open: (b) author regen-loop spec DRAFT 2 (8 findings) → non-author re-review;
(c) ④ twin-root sweep (unblocked, heavy); (d) `delivery-enforce` verdict-schema cleanup. Push remains
the owner's closeout step.
