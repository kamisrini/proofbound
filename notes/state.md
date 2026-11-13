# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-13

## Resume — 2026-11-13

**Exact next action:** none required to ship — **Release 1 is written, verified, committed, AND
pushed** (`HEAD == origin/main == 316f1e9`). The only remaining items are **owner decisions**, none
started: (1) accept Release 1 as shipped; (2) tag `v0.1.0`
(`git tag -a v0.1.0 -m "Release 1 — the Trust Record"`); (3) start a single-dev corp pilot. Active
push (unchanged): make Release 1 (the Trust Record) usable on a real project under
`VD-agent-standing-authorization-5ouz7q` / [[proofbound-decision-commits-need-intent]].

**⚠ Known gate trap — read before running `make check`:** the two permanent-untracked paths
(`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) keep the worktree dirty, and
`scripts/commit-cadence.sh` fails once HEAD is older than 90 minutes while the tree is dirty. So on
any session that starts >90m after the last commit, bare `make check` goes **RED on
`commit-cadence`** by the passage of time alone — not from any code change. The rest of the gate is
green. Clear it with the normal wrap commit (makes HEAD <90m). The durable fix is an **owner
decision** on those two files: commit them, `.gitignore` them, or delete them.

**Docs layout note:** owner chose a single per-release folder (`docs/releases/r1/`) over the hybrid
I recommended; I flagged the Law 2 drift risk (a living user guide forked per release is a drift
seed). If/when R2's guide diverges, revisit — consolidate the living guide, version only the notes.
The three docs are NOT governed (no intent chain needed to edit them).

**R1 wrap-up status** (see [`RELEASE-PLAN.md`](../docs/plans/RELEASE-PLAN.md)) — **all DONE**:
- **R1-1** flaky test — RESOLVED.
- **R1-2** clean-clone proof — DONE + pushed ([`r1-2-fresh-clone-release1.md`](../docs/verification/r1-2-fresh-clone-release1.md)).
- **R1-3** `DEF-ROUTE-MATRIX-PRECOMMIT` — **RESOLVED** (`fa8e1af`, `a5a085b`, `CI-staged-pin-guard-7b3c9d`):
  route-matrix + `p6-task2-close` `--check` fail closed on pending pinned-path changes (seam-tested).
- **R1-4** `DEF-DELIVERY-ENFORCE-VERDICTS` — **RESOLVED** (`6b2961a`): `verify` runs green on a fresh clone.
- **README / quickstart + Finding A/B** — DONE.

**Immediate blockers:** none. R1 is shippable; it is now an owner-acceptance call. Governance
follow-ups still `proposed` (run `scripts/p6-task5-review.sh` when convenient):
`BR-release-line-governance-7ff96f`, `BR-self-hosting-correctness-4607a3`, `BR-gate-integrity-4b2e08`.

## This session (2026-11-13)

- **No product work** — a state-reconciliation + gate-verification wrap. No source commits between
  the prior wrap (`316f1e9`) and this one.
- **Reconciled drift:** the prior state.md next-action still said "push the docs commits (HEAD ahead
  of origin)", but the push had already completed (`HEAD == origin/main`). It also listed
  `DEF-ROUTE-MATRIX-PRECOMMIT` as OPEN though the tsv records it RESOLVED. Both corrected here.
- **make check:** RED only on `commit-cadence` (the idle-tree trap above); every other check green
  (16 production packages PASS, generated-freshness ok, all hostile/closure checks ok). Cleared by
  this wrap commit. See today's journal for the full LESSON.

## Branch status
`main` → `origin/main` **synced** at HEAD `316f1e9` (this notes wrap commit will put HEAD 1 ahead
until pushed). Two permanent untracked paths remain (`docs/plans/P5-intent-provenance-plan.md`,
`kernel/cmd/p6dbprobe/`) — owner decision pending (commit / ignore / delete); they are the cause of
the commit-cadence idle-trap above.

## Open / deferred (see `docs/deferrals.tsv`) — 2 OPEN (both dormant, owner-gated)
- `DEF-P7-REGENLOOP-001` OPEN (PENDING_PREREQUISITE) — regen full loop (owner-gated).
- `DEF-P7-DP-TWINROOT-001` OPEN (DEFERRED) — heavy twin-root detector-power sweep; `deferrals.test.sh` fixture anchor.
- RESOLVED previously: `DEF-ROUTE-MATRIX-PRECOMMIT`, `DEF-DELIVERY-ENFORCE-VERDICTS`.

## Key facts
- **Never pipe `make check`** (not even `| tail`) — the pipe swallows make's exit code; a red gate
  reads as green. `make check` includes `commit-cadence` (dirty-tree while HEAD > 90m old fails — see
  the idle-trap note under Resume), `state-freshness` (HEAD > 3 commits past state.md fails), and
  `delivery-enforce.test.sh` (needs an `Intent:` trailer when HEAD touches a governed path). See
  [[proofbound-make-check-vs-delivery-enforce]].
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
Run `/proofbound-next`. R1-1..R1-4 are all DONE and pushed; the open items are owner decisions only:
accept R1 as shipped, tag `v0.1.0`, start a single-dev corp pilot. If bare `make check` is red on
`commit-cadence` at start, that is the idle-tree trap (see Resume), not new breakage — a wrap commit
clears it. Release 1 user docs live under `docs/releases/r1/` (verified against the binary).
