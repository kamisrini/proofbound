# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-10

## Resume — 2026-11-10

**Exact next action:** nothing is mid-flight — everything is committed and `make check` is GREEN.
Active push: **make Release 1 (the Trust Record) usable on a real project** (owner, 2026-11-10),
under standing authorization [[proofbound-decision-commits-need-intent]] /
`VD-agent-standing-authorization-5ouz7q`. **Proven live this session:** the binary builds (CGO off,
no gcc/Java/docker), `proofbound sync git` ingests this repo (596 commits, idempotent), and after a
blocker fix `proofbound report week` renders the weekly proof view (85 commits, exit 0). Next
Release-1-usability steps, in priority order: **(1)** probe the rest of the live surface for more
real-data blockers (`verify`, `report github`, `intent check`, `gates enforce`) — self-hosting is the
acceptance test (`BD-self-hosting-correctness-11bcf2`); **(2)** refresh the stale `README.md` (it
still says "not product-ready / P6-complete / P7-planning, as of 2026-09-22" — the literal front door)
+ write a real end-user quickstart (set `DATABASE_URL` OR embedded dev-mode; `sync` → `report`);
**(3)** R1-2 clean-clone proof; **(4)** R1-3 `DEF-ROUTE-MATRIX-PRECOMMIT`; **(5)** R1-4
`DEF-DELIVERY-ENFORCE-VERDICTS` (owner-optional). See [`RELEASE-PLAN.md`](../docs/plans/RELEASE-PLAN.md).
**Immediate blockers:** none. `make check` green; push this session's commits as the closeout step.
**Pending non-author verdict (Law 9):** `internal/projections` on the `CI-fix-vd-citation-9958b1`
revision (+ calibrated mutation sweep) before that fix counts as accepted.

**One open governance follow-up:** `BR-release-line-governance-7ff96f` is `proposed` — it goes
`active` only after a non-author requirement review (`scripts/p6-task5-review.sh`). Normal lifecycle
(the gate-integrity BR sits the same way); run that review when convenient. See
[[proofbound-decision-commits-need-intent]].

## This session (2026-11-10, continued)

- **Release 1 proven to RUN on real data + first real-data blocker FIXED.** Built the `proofbound`
  binary (pure Go, CGO off — the toolchain gap only blocks the embedded-PG-from-source/B1 path, not
  the Flight Recorder). Live-ran against this repo (embedded PG fallback when `DATABASE_URL` unset):
  `sync git` = listed 596 / appended 19 / existing 577 (idempotency proven). `report week` first
  CRASHED (`invalid cited_decisions`): the commit projection's `decisionPattern` only accepted the
  6-char VD suffix, rejecting all 17/31 date-suffixed `VD-*` decisions, failing the projection apply
  closed. **FIXED** (`CI-fix-vd-citation-9958b1` under `BD-self-hosting-correctness-11bcf2` →
  `BR-self-hosting-correctness-4607a3` O-1): broadened `decisionPattern` to accept the `YYYY-MM-DD`
  form + a NON-integration-tagged regression test (`citation_test.go`, so it runs in `make check` —
  the old commit-validation tests were integration-tagged and skipped by the gate) + SPEC §3 note.
  Re-ran live: `report week` exits 0, renders 85 commits with proof. Commit + route-matrix refresh
  follow-up. `make check` green. **Needs a non-author `internal/projections` verdict (Law 9).**
- **Release line drawn (owner decision).** `VD-release-line-kwdh4v`: **Release 1 = "the Trust Record"**
  (P1 Flight Recorder + P2 gates-as-data + P3 GitHub connector + P5 intent provenance + P7 B1 behavior
  locks/runtime replay — records + verifies, run against a user-supplied Postgres); **Release 2** = the
  revenue organs (#2b full loop, #3 Twin product, #4 Governor, #5 Trust Engine) + embedded-DB
  redistribution. Maintained scope split in `docs/plans/RELEASE-PLAN.md`; `ROADMAP.md` P7+ points to it.
  Committed `34af46d` under a NEW honest governance intent chain
  (`BD-release-line-governance-a01de6` → `BR-release-line-governance-7ff96f` O-1 →
  `CI-release-line-4069b0`; `Intent:` trailer). **Why the chain:** a VD commit regenerates the governed
  `@generated` decisions index, so `make check`'s delivery-enforce self-test fails closed without an
  `Intent:` trailer; the precedent `daf9fe4` skipped it and shipped a red gate. See
  [[proofbound-decision-commits-need-intent]].
- **`DEF-KERNELCHECK-FLAKE-001` CLOSED (owner final close 2026-11-10, Law 9).** Row flipped
  OPEN→RESOLVED; the `deferrals.test.sh` `--open`/`--hold-type` fixture (which hard-coded this row) was
  repointed to `DEF-P7-DP-TWINROOT-001` in the same governed commit (`Intent:
  CI-flake-ephemeral-port-7d5f10`). The ephemeral-port fix (`057bc07`; route-matrix slip `8700760`) was
  non-author accepted round-1+2; `make check` green + deterministic.

## Branch status
`main` → `origin/main`. Commits this session: `34af46d` (release line) + the flake-close commit (this
state file rides with it). **Push is the owner's closeout step.** Two permanent untracked paths remain
(`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`).

## Open / deferred (see `docs/deferrals.tsv`) — 4 OPEN
- `DEF-P7-REGENLOOP-001` OPEN — regen gate-only slice FULLY CI-COMPLETE + non-author accepted; remaining
  is owner final acceptance + the FULL loop (step 3/4), which reopens the §5.1 regenerator-boundary.
- `DEF-P7-DP-TWINROOT-001` OPEN — heavy twin-root detector-power sweep; also now the `deferrals.test.sh`
  fixture anchor (keep OPEN, or repoint the fixture if it is ever closed).
- `DEF-DELIVERY-ENFORCE-VERDICTS` OPEN — repo-wide verdict-schema cleanup (R1-4, owner-optional).
- `DEF-ROUTE-MATRIX-PRECOMMIT` OPEN — staged-aware derived-pin freshness (R1-3).
- RESOLVED: `DEF-P7-001`/`-002`/`-003`, `DEF-METATAX-001`, and now `DEF-KERNELCHECK-FLAKE-001`.

## Key facts
- `make check` = Makefile `check:` target; does NOT run `delivery-enforce` (sync + gates enforce), BUT
  `make hooks-test` (in `make check`) runs `delivery-enforce.test.sh`, which requires HEAD carry an
  `Intent:` trailer when HEAD touches a governed path. See [[proofbound-decision-commits-need-intent]]
  and [[proofbound-make-check-vs-delivery-enforce]].
- Governed paths (need an `Intent:` trailer): `scripts/*`, `tools/*`, `gates/*`, `kernel/*` (except
  SPEC.md/testdata), `Makefile`, `CLAUDE.md`, `docs/gates.md`, `docs/decisions/INDEX.md`,
  `docs/{invariants,laws}.lock`, `docs/allowed-{skips,survivors}.txt`. `docs/intent/records/*`,
  `docs/decisions/VD-*.md`, `docs/plans/*`, `ROADMAP.md`, `notes/*` are NOT governed.
- Verdict/intent schemas + digests: see [[proofbound-intent-record-digests]].
- Detector-power refresh: `scripts/detector-power.sh <pkg> --refresh` reproduces `raw_sweep_sha256`.
- Route-matrix trap: committing `kernel/internal/{projections,cli,twin,gates}` re-stales the
  `@generated` route matrix; regenerate with `scripts/p6-route-matrix.sh --write` AFTER committing.
  See [[proofbound-route-matrix-staleness-trap]].

## Next session's first action
Run `/proofbound-next`. The highest-value R1 item is **R1-2 clean-clone proof**. Push this session's
commits first (owner closeout).
