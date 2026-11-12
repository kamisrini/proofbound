# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-12

## Resume — 2026-11-12

**Exact next action:** nothing is mid-flight — all work is committed, pushed (`origin/main` =
`c7a8a39`), and `make check` is GREEN (verified bare, true exit 0, this session). Active push
(unchanged): **make Release 1 (the Trust Record) usable on a real project** under
`VD-agent-standing-authorization-5ouz7q` / [[proofbound-decision-commits-need-intent]]. The last
required R1 item is **R1-3 `DEF-ROUTE-MATRIX-PRECOMMIT`** (governed; author a change intent under the
gate-integrity chain, as done this session for `CI-readme-r1-status-b7f2a9` / `CI-gate-expiry-fresh-c4e2a1`).

**R1 wrap-up status** (see [`RELEASE-PLAN.md`](../docs/plans/RELEASE-PLAN.md)):
- **R1-1** flaky test — RESOLVED (prior session).
- **R1-2** clean-clone proof — **DONE + pushed** ([`r1-2-fresh-clone-release1.md`](../docs/verification/r1-2-fresh-clone-release1.md)).
- **R1-3** `DEF-ROUTE-MATRIX-PRECOMMIT` — **OPEN, not started.** The one remaining required R1 item.
  It is a GOVERNED (`scripts/*`) change → needs an owner-approved change intent (author a CI under
  the gate-integrity chain, as was just done for `CI-readme-r1-status-b7f2a9`).
- **R1-4** `DEF-DELIVERY-ENFORCE-VERDICTS` — **DONE this session** (owner made it required): `verify`
  now runs on a fresh clone. **RESOLVED.**
- **README / quickstart + Finding A** — **DONE this session.**

**Immediate blockers:** none. **Remaining for R1:** R1-3 (route-matrix pre-commit), then owner
accepts R1 as shipped. Governance follow-ups still `proposed` (run `scripts/p6-task5-review.sh` when
convenient): `BR-release-line-governance-7ff96f`, `BR-self-hosting-correctness-4607a3`,
`BR-gate-integrity-4b2e08`.

## This session (2026-11-12)

- **R1-2 clean-clone proof** (`6f0a1b0`): fresh `git clone --no-local` of HEAD with fresh embedded-PG
  + fresh build/lint caches builds (CGO off) and `make check` green (cold). Live surface proven
  end-to-end in the correct bootstrap order. Surfaced two findings → both now addressed.
- **Finding A (bootstrap order undocumented)** → **README refresh** (`a26e957`, `b3477c2`): replaced
  the stale "P6 / not product-ready" front door with Release-1 reality + a real end-user quickstart
  documenting the load-bearing order (`migrate historical-evidence` FIRST on an empty ledger → `sync
  intent all` → `sync git` → `report`/`verify`). Re-added operator `make` targets for the p6-census C1.
- **Finding B / R1-4 (`verify` red on fresh clone)** → reclassified R1-4 required (`074da6c`), then
  **migrated 15 pre-schema verdicts to `vera.verdict.v1`** (`421938c`): status synonyms mapped,
  findings reduced to `{finding_id,severity[,defect_commit]}` (info→LOW, low-medium→MED) with the
  original block preserved **verbatim in-body**; NEEDS_WORK verdicts keep ≥1 finding (projection
  invariant). **Proven on a clean-bootstrapped ledger: `sync reviews` malformed=0, `sync all` exit 0,
  `verify` exit 0.** `DEF-DELIVERY-ENFORCE-VERDICTS` RESOLVED (`6b2961a`).
- **Gate-integrity fix** (`c6aa4db`, governed, `Intent: CI-readme-r1-status-b7f2a9`): the README
  refresh broke `p6-task1-close` (C1-001), which hard-coded the stale phase sentence. Retargeted the
  assertion to the stable phrase `feature-complete and independently accepted`. See the
  `hardcoded-fixture-coupling` lesson.
- **Corp-readiness live-verb probe** (self-hosting) — tested the untested surface. Working:
  `sync checks/sessions`, `report github`, `gates canary`. Two findings, both fixed:
  - **Finding A (docs gap, `b252c0a`):** `sync github` needs `PROOFBOUND_GITHUB_OWNER/REPOS/API_BASE_URL`
    + `GITHUB_TOKEN` (the CLI wires them; they were undocumented). Documented in the README.
  - **Finding B (real R1 defect, `c7a8a39`, governed `Intent: CI-gate-expiry-fresh-c4e2a1`):** all 7
    enforce-mode gate definitions had expired 2026-10-16, so `gates enforce` was silently dead (make
    check never runs enforce). Renewed to **2027-02-16**; `gates enforce` exits 0 again. Added
    `scripts/gate-expiry.sh` + self-test (auto-run by hooks-test) so an expired/dateless gate can
    never pass make check unseen (`LESSON class=silent-gate-expiry`).

## Branch status
`main` → `origin/main` **synced** at HEAD `c7a8a39` (this state commit will put HEAD 1 ahead until
pushed). Two permanent untracked paths remain (`docs/plans/P5-intent-provenance-plan.md`,
`kernel/cmd/p6dbprobe/`). New this session: `gates enforce` restored + `scripts/gate-expiry.sh` guard.

## Open / deferred (see `docs/deferrals.tsv`) — 3 OPEN
- `DEF-P7-REGENLOOP-001` OPEN — regen full loop (owner-gated).
- `DEF-P7-DP-TWINROOT-001` OPEN — heavy twin-root detector-power sweep; `deferrals.test.sh` fixture anchor.
- `DEF-ROUTE-MATRIX-PRECOMMIT` OPEN — R1-3, the last required R1 item (governed; needs a change intent).
- RESOLVED this session: `DEF-DELIVERY-ENFORCE-VERDICTS`.

## Key facts
- **Never pipe `make check`** (not even `| tail`) — the pipe swallows make's exit code; a red gate
  reads as green. `make check` includes `commit-cadence` (dirty-tree > 90m fails), `state-freshness`
  (HEAD > 3 commits past state.md fails), and `delivery-enforce.test.sh` (needs an `Intent:` trailer
  when HEAD touches a governed path). See [[proofbound-make-check-vs-delivery-enforce]].
- **Authoring an intent chain** (for a governed change): ID `^CI-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$`;
  self-digest = zero the record's OWN (last) `"artifact_sha256"` then sha256 the file; CI
  `targets[].artifact_sha256` = the BR's self-digest; must target ≥1 obligation. Home for governed
  maintenance: `BR-gate-integrity-4b2e08` O-1. See [[proofbound-intent-record-digests]].
- Governed paths need an `Intent:` trailer; `docs/verification/*`, `docs/intent/records/*`,
  `docs/plans/*`, `README.md`, `notes/*` are NOT governed.
- Embedded-PG dev mode: `DATABASE_URL` unset → embedded PG under `./.proofbound` (persistent, pure
  Go). The ledger is append-only — a bad `sync` pollutes it; reset the dev DB by moving
  `.proofbound/db` aside (don't `rm -rf`; the classifier blocks it).
- Route-matrix trap: committing `kernel/internal/{projections,cli,twin,gates}` re-stales the
  `@generated` route matrix; regenerate with `scripts/p6-route-matrix.sh --write` AFTER committing.

## Next session's first action
Run `/proofbound-next`. Push the 5 unpushed commits first (owner closeout). Then the last required R1
item is **R1-3 `DEF-ROUTE-MATRIX-PRECOMMIT`** (governed; author a change intent under the
gate-integrity chain).
