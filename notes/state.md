# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-11 (EOD wrap executed early 2026-11-12)

## Resume — 2026-11-11

**Exact next action:** nothing is mid-flight — everything is committed and `make check` is GREEN
(verified bare, true exit 0, this session). `origin/main` is **1 commit behind** HEAD `93bebd2`:
**push is the owner's closeout step** and was deliberately NOT done this session. Active push
(unchanged): **make Release 1 (the Trust Record) usable on a real project** under standing
authorization `VD-agent-standing-authorization-5ouz7q` / [[proofbound-decision-commits-need-intent]].
The highest-value unblocked R1 item is **R1-2: the clean-clone proof** (fresh clone builds + `make
check` green on a clean machine; mind the documented WSL prereqs — `~/go/bin` on PATH, Go-1.27
golangci-lint, `PROOFBOUND_BACKUP_DIR`). After that, in priority order: **(a)** continue probing the
live surface for real-data blockers (`verify`, `report github`, `intent check`, `gates enforce`) —
self-hosting is the acceptance test (`BD-self-hosting-correctness-11bcf2`); **(b)** refresh the stale
`README.md` (still says "not product-ready / P6 / P7-planning, 2026-09-22") + a real end-user
quickstart; **(c)** R1-3 `DEF-ROUTE-MATRIX-PRECOMMIT`; **(d)** R1-4 `DEF-DELIVERY-ENFORCE-VERDICTS`
(owner-optional). See [`RELEASE-PLAN.md`](../docs/plans/RELEASE-PLAN.md). **Immediate blockers:** none.

**Governance follow-ups (both normal lifecycle, run when convenient):** two requirements sit
`proposed` pending a non-author requirement review (`scripts/p6-task5-review.sh`):
`BR-release-line-governance-7ff96f` and `BR-self-hosting-correctness-4607a3`.

## This session (2026-11-11)

- **Cleared the last Law-9 debt on the self-hosting fix.** Produced the non-author **ACCEPTABLE**
  verdict for `internal/projections` on `CI-fix-vd-citation-9958b1` (the date-suffixed `VD-*` decision
  citation fix, authored 2026-11-10 in `ac76f88` — this session did not author it, so the review is
  acceptance, not a self pre-check). Focused calibrated mutation sweep on the changed surface:
  positive controls A–D died, declared-equivalent neutral E survived, lethal injected-test F killed,
  zero undeclared survivors. P-INV-7 fail-closed preserved; Laws 1/6/7 hold. One non-blocking advisory
  (date branch is a shape guard, accepts calendar-invalid dates; pre-existing, attack surface
  unchanged). Verdict `docs/verification/verdicts/p7-projections-vd-citation-round1-20261111.md`
  (self-digest validated) + regenerated `@generated` audit (`docs/verification/p6-artifact-integrity.md`,
  52 schema-bearing) committed as `93bebd2`. `docs/verification/*` is ungoverned → no `Intent:` trailer.
- **Process slip caught in-session:** piped `make check` through `tail`, which masked a red
  `commit-cadence` gate (tail's exit 0 ≠ make's). Caught it, committed to refresh HEAD, re-ran the
  gate bare → genuinely green. See the 2026-11-11 journal `LESSON: class=unverified-gate-claim`.

## Branch status
`main` → `origin/main` **[ahead 1]** (HEAD `93bebd2`, NOT pushed — owner closeout). Two permanent
untracked paths remain (`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`); these
keep the worktree permanently "dirty," so `commit-cadence` passes only within 90m of the last commit.

## Open / deferred (see `docs/deferrals.tsv`) — 4 OPEN (unchanged this session)
- `DEF-P7-REGENLOOP-001` OPEN — regen gate-only slice CI-complete + accepted; remaining is owner final
  acceptance + the FULL loop (step 3/4), which reopens the §5.1 regenerator-boundary.
- `DEF-P7-DP-TWINROOT-001` OPEN — heavy twin-root detector-power sweep; also the `deferrals.test.sh`
  fixture anchor (keep OPEN, or repoint the fixture if it is ever closed).
- `DEF-DELIVERY-ENFORCE-VERDICTS` OPEN — repo-wide verdict-schema cleanup (R1-4, owner-optional).
- `DEF-ROUTE-MATRIX-PRECOMMIT` OPEN — staged-aware derived-pin freshness (R1-3).
- RESOLVED: `DEF-P7-001`/`-002`/`-003`, `DEF-METATAX-001`, `DEF-KERNELCHECK-FLAKE-001`.

## Key facts
- `make check` = Makefile `check:` target. It includes `commit-cadence` (dirty-tree > 90m fails) and
  `state-freshness`. It does NOT run `delivery-enforce`, BUT `make hooks-test` (inside it) runs
  `delivery-enforce.test.sh`, which requires an `Intent:` trailer when HEAD touches a governed path.
  **Never pipe `make check`** (not even `| tail`) — the pipe swallows make's exit code and a red gate
  reads as green. See [[proofbound-make-check-vs-delivery-enforce]], [[proofbound-decision-commits-need-intent]].
- Governed paths (need an `Intent:` trailer): `scripts/*`, `tools/*`, `gates/*`, `kernel/*` (except
  SPEC.md/testdata), `Makefile`, `CLAUDE.md`, `docs/gates.md`, `docs/decisions/INDEX.md`,
  `docs/{invariants,laws}.lock`, `docs/allowed-{skips,survivors}.txt`. `docs/verification/*`,
  `docs/intent/records/*`, `docs/decisions/VD-*.md`, `docs/plans/*`, `ROADMAP.md`, `notes/*` are NOT governed.
- Verdict/intent schemas + digests: [[proofbound-intent-record-digests]]. After adding a verdict,
  regenerate `docs/verification/p6-artifact-integrity.md` with `scripts/p6-artifact-integrity.sh --write`.
- Route-matrix trap: committing `kernel/internal/{projections,cli,twin,gates}` re-stales the
  `@generated` route matrix; regenerate with `scripts/p6-route-matrix.sh --write` AFTER committing.
  See [[proofbound-route-matrix-staleness-trap]].

## Next session's first action
Run `/proofbound-next`. Highest-value unblocked item: **R1-2 clean-clone proof**. If the owner wants
this session's commit upstream first, **push `93bebd2`** (origin is 1 behind).
