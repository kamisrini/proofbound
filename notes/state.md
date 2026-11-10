# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-09

## Resume — 2026-11-09

**Exact next action:** nothing is mid-flight — all of today's work is committed, pushed, and `make
check` is GREEN at HEAD (`e6e9e2d`). Run `/proofbound-next` to pick the next AUTHORIZED task. The
natural continuations, all owner-steered (every item below is a `status=OPEN` deferral — dormant until
the owner names it): **(a)** the regen-gate `docs/gates.md` advisory registration (the one CI-scope
follow-up; 5-file census cascade, recipe in the `DEF-P7-REGENLOOP-001` row + below) then owner final
acceptance of the gate; **(b)** owner's final close of `DEF-KERNELCHECK-FLAKE-001` (also repoint the
`deferrals.test.sh` fixture); **(c)** `DEF-P7-DP-TWINROOT-001` twin-root sweep; **(d)**
`DEF-DELIVERY-ENFORCE-VERDICTS` repo-wide verdict-schema cleanup; **(e)** the regen-loop full loop
(step 3/4), which reopens the §5.1 regenerator-boundary decision.
**Immediate blockers:** none. `make check` green; origin up to date.

**Background (today, 2026-11-09):** all five `/proofbound-next` items advanced; six non-author
verdicts landed; the O-2 regeneration-gate (gate-only slice) was implemented and non-author accepted.
Status detail:

- **① F1 (detector-power seed citations) — DONE + non-author re-confirmed.** Fix `25d13ad`; verdict
  schema corrected `9ee6c04`; **non-author re-confirmation ACCEPTABLE**
  (`verdicts/p7-2b-detector-power-f1-fix-reconfirm-round1-20261109.md`). F1 fully closed.
- **② state — DONE** (this file). **③ regen-loop spec — DRAFT 2 ACCEPTED (round-2 non-author verdict
  ACCEPTABLE); CL-1..CL-3 applied. DONE pending owner §5 + implementation.** Round-1 NEEDS_WORK (8
  findings) → draft 2 addressed all 8 (F1 resolved by scoping the first slice to `U=0` tiers; RL-9/10/11
  added; D-b strengthened; RL-12 future-guard). Round-2 verdict ACCEPTABLE
  (`verdicts/p7-2b-regeneration-loop-spec-review-round2-20261109.md`) with 3 non-blocking
  clarifications now APPLIED as CL-1..CL-3: CL-1 defines the Law-9 actor-identity basis (owner decision
  §5.7; until then a declared RL-8 limitation); CL-2 moves RL-12 out of the active contract into §4.1
  with NO self-test (avoids a vacuous control, Law 3); CL-3 keys the D-a record per (subject,
  regenerator) deployment vs D-b per revision (Law 2). **Next on ③: owner settles §5, then a NEW
  change intent + spec-first implementation + non-author package verdict.**
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
- Regen-loop spec: draft 1 (NEEDS_WORK) → draft 2 addressing all 8 findings → **round-2 ACCEPTED**
  (ACCEPTABLE) + CL-1..CL-3 applied. Spec design done; implementation gated on owner §5.
- ⑤ flake fixed (`057bc07`) under the gate-integrity intent chain (`98cdeea`); route-matrix Law-1 slip
  fixed (`8700760`); `make check` green. Three verdicts committed (`0d36784`).

## Open / deferred (see `docs/deferrals.tsv`)
- `DEF-KERNELCHECK-FLAKE-001` OPEN — fixed + non-author accepted (round-1+2); awaiting owner final close.
- `DEF-P7-REGENLOOP-001` OPEN — spec ACCEPTED; §5 owner-accepted (gate-only + defaults); CI
  `CI-regen-loop-gate-1a2b3c` accepted (targets BR O-2). **Gate-only slice IMPLEMENTED + non-author
  ACCEPTED (round-1):** `scripts/regen-gate.sh` + `scripts/tests/regen-gate.test.sh` (RL-1/2/4/5/8/9/10/11,
  AUTHORIZE/DENY, fail-closed; reuses `scripts/detector-power.sh` for RL-2), committed `635afd3`,
  hardened `a785420` (round-1 verdict findings: RL-11 D-b-leg control, RL-2 split, JSON pre-validation).
  Verdict `verdicts/p7-2b-regen-gate-impl-round1-20261109.md` ACCEPTABLE. `make check` green.
  **ONE CI-scope follow-up remains: the `docs/gates.md` advisory row + census coupling** (RL-8 standing
  advisory). Recipe (a 5-file cascade; careful — `p6-census.md` carries a commit-pin like the route
  matrix): (1) add a `gates.md` row [Check|Tier(ADVISORY …)|Enforced by|Self-test|Expiry (ISO + `Owner:`)]
  naming `scripts/regen-gate.sh`; (2) add `C2-043` to `docs/plans/p6-census-rows.tsv`
  (`gate-row:<exact Check cell>`, `path:scripts/regen-gate.sh`, files `scripts/regen-gate.sh,scripts/tests/regen-gate.test.sh`);
  (3) bump `scripts/p6-task1-close.sh` `-eq 42`→43 and `seq 1 42`→43; (4) `scripts/p6-census.sh --write`
  + `scripts/p6-task1-close.sh --write`; (5) `make check`. Regenerator (step 3) + full loop remain
  out of scope (future CI). Owner is the final acceptor.
- `DEF-P7-DP-TWINROOT-001` OPEN — unblocked; heavy (④).
- `DEF-P7-001`/`-002`/`-003`/`DEF-METATAX-001` RESOLVED.

## Unverified / residual (Law 9 — each needs a NON-author lens)
- Regen-loop spec implementation (after owner §5) → non-author package verdict (future).
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
Run `/proofbound-next`. Open: (b) regen-gate — the `docs/gates.md` advisory registration cascade
(recipe in the DEF-P7-REGENLOOP-001 row above), then owner final acceptance; the full loop (step 3/4)
is a future CI;
(c) ④ twin-root sweep (unblocked, heavy); (d) `delivery-enforce` verdict-schema cleanup. Push remains
the owner's closeout step.
