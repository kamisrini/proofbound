# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-16

## Resume — 2026-11-16

**Exact next action:** the **release-line renumber is committed and pushed** (one governed commit,
trailer `Intent: CI-release-line-r2-sdd-t5j34p`; owner authorized commit + push 2026-11-16). What it
landed: **R1 ACCEPTED as shipped** + **release line redrawn into three cuts** — R1 = the Trust Record (shipped), **R2 = the
SDD Warranty Connector**, **R3 = the revenue organs + turnkey redistribution** (the former R2 scope,
unchanged). Authority: `VD-release-line-r2-sdd-t5j34p` (supersedes `VD-release-line-kwdh4v`).
**R2 M0 in progress (owner-guided, one blocker at a time).** Step 1 DONE 2026-11-16: mapping spike
= **GO** ([`r2-m0-mapping-spike.md`](../docs/verification/r2-m0-mapping-spike.md)); governance → BD
is NO-GO (→ `authorization: undeclared`); owner chose **(a) profile-declared accountable
identity** → `VD-sdd-accountable-identity-ypcx7x`. Step 2 DONE: R2 intent home
`BD-sdd-warranty-jdyjhh` → `BR-sdd-warranty-i9miv7` (O-1..O-7 approved verbatim by the owner;
`proposed` until non-author review) → `CI-sdd-warranty-m0-a49f5b` (covers M0 records + step 3
SPECs). Step 3 DONE (pending owner OK to commit): the `intent/sdd` SPEC is written as a **draft**,
`docs/plans/R2-sdd-provider-spec-draft1.md` (SDD-INV-1..14; it becomes the kernel `SPEC.md` with
M1's code and tests, per census row C2-036), and the intent connector SPEC's provider paragraph
amended to admit `sdd` (production only once its proofs pass). **Next:
step 4** (review round 1) DONE: both verdicts NEEDS_WORK (`a2cc920`). Response in progress: SPEC
**draft 2** (`docs/plans/R2-sdd-provider-spec-draft2.md`): active-changes-only, native-name ids,
manifest digests, a Linkage-and-lifecycle section (native-name trailers `Intent: sdd:<change>`,
satisfaction via R1's obligation-verdict.v2 + readiness gate, read-only feedback). **Owner decisions
open: A** (warranty across archive; author recommends A2 re-attest), **B** (obligation lifecycle vs
intent SPEC §5; author recommends B1 deltas-as-lifecycle), **O-6 rewording.** Gate findings
(verdict 2 F1/F3/F4) being fixed; F2's central claim was disproven by reproduction (journal).

**⚠ Known gate trap — read before running `make check`:** the two permanent-untracked paths
(`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) keep the worktree dirty, and
`scripts/commit-cadence.sh` fails once HEAD is older than 90 minutes while the tree is dirty. So on
any session that starts >90m after the last commit, bare `make check` goes **RED on
`commit-cadence`** by the passage of time alone — not from any code change. The rest of the gate is
green. Clear it with the normal wrap commit (makes HEAD <90m). The durable fix is an **owner
decision** on those two files: commit them, `.gitignore` them, or delete them.

**Docs layout note:** owner chose a single per-release folder (`docs/releases/r1/`) over the hybrid
I recommended; I flagged the Law 2 drift risk (a living user guide forked per release is a drift
seed). Now a three-release line (R1 shipped / R2 SDD connector / R3 organs); if/when a later
release's user guide diverges, revisit — consolidate the living guide, version only the notes. The
R1 release docs are NOT governed (no intent chain needed to edit them).

**R1 = ACCEPTED as shipped (owner, 2026-11-16)** — recorded in `VD-release-line-r2-sdd-t5j34p`. The
wrap-up set R1-1..R1-4 was all DONE and the clean-clone proven before acceptance. Follow-ups
(unstarted): tag the release, add a demo script, run a single-dev corp pilot.

**R2 M0 review round 2 = NEEDS_WORK (6 HIGH): M1 must not start.** Draft 2's lifecycle fails against real OpenSpec history (verdict `r2-m0-review-round2-20261116`). **Owner decision pending on R2 direction:** (1) history-native provider, full lifecycle; (2) narrow R2 to change-scoped warranty, history-native for change folders, main-spec lifecycle deferred (author recommends); (3) pause R2.

**Trailer gate round 3 = NEEDS_WORK (1 MED, 5 LOW).** Range checking works; the remaining harm is structural: a pre-push check is skipped by pushing first. **Owner decision pending:** a push- or server-side check mechanism. Push of local commits is HELD by the owner.

**Immediate blockers:** none for planning. Owed: a non-author verdict on the gate change (closes the
deferral) and on the R2 plan (before any M1 code). Next work: R2 M0 (the mapping spike). After the
gate-fix commit, run `scripts/p6-task2-close.sh --write` and commit the refreshed pin (it pins
`kernel/internal/connector`). Governance follow-ups
still `proposed` (run `scripts/p6-task5-review.sh` when convenient):
`BR-release-line-governance-7ff96f` (its O-1 governs this renumber — a CI may target a proposed
obligation), `BR-self-hosting-correctness-4607a3`, `BR-gate-integrity-4b2e08`.

## This session (2026-11-16)

- **Governed release-line renumber (documentary only; no connector code).** Owner directed: accept
  R1 shipped; insert **R2 = the SDD Warranty Connector** (warrant the output of any spec-driven-dev
  tool — OpenSpec reference, Spec Kit subset); push former R2 (revenue organs + turnkey) to **R3**.
- **New governance records:** `VD-release-line-r2-sdd-t5j34p` (supersedes `VD-release-line-kwdh4v`,
  marked Superseded) + `CI-release-line-r2-sdd-t5j34p` (targets `BR-release-line-governance-7ff96f`
  O-1; validated by feeding the working tree through the real `records` provider + `intent.Sync`
  batch validation in a throwaway overlay test — all 36 records pass, the new CI ingests as
  `change_intent.recorded`). **LESSON:** `proofbound sync` reads the committed `HEAD` tree, so it
  silently ignores an uncommitted record — a green `sync` is NOT validation of an unstaged file.
  `make index` regenerated the governed INDEX.
- **R2 plan written:** [`sdd-warranty-connector-plan.md`](../docs/plans/sdd-warranty-connector-plan.md)
  (M0–M6, roles, non-goals). Citation sweep R2→R3 across README / ROADMAP / RELEASE-PLAN / r1 notes;
  committed preview `docs/releases/r2/regen-repair/` → `r3/` (git mv). No stale `releases/r2` refs.
- **make check:** RED only on `commit-cadence` (the idle-tree trap above); every substantive check
  green (16 production packages PASS, generated-freshness ok, delivery-enforce invariants ok,
  non-author-identity ok). Clears on the commit below. A second full gate over the post-validation
  bytes went RED on `p6-census` (unclassified mechanism: this file had backticked the not-yet-existing
  `intent/sdd` SPEC *path*) — fixed by naming the package, not the path; see today's journal LESSON.
  Third full gate over the final bytes: **green except `commit-cadence`** (22 checks ok, census
  passes) — committed and pushed on owner go-ahead.

## Branch status
`main` → `origin/main` **synced** after this session's governed release-line commit (pushed
2026-11-16). Two permanent untracked paths remain
(`docs/plans/P5-intent-provenance-plan.md`, `kernel/cmd/p6dbprobe/`) — owner decision pending
(commit / ignore / delete); they are the cause of the commit-cadence idle-trap above.

## Open / deferred (see `docs/deferrals.tsv`) — 3 OPEN (all dormant, owner-gated)
- `DEF-INTENT-TRAILER-BLOCK-001` OPEN (DEFERRED) — 17/30 governed commits carry `Intent:` outside
  git's final trailer block, so the kernel binds no claim; fix-forward under `BR-gate-integrity-4b2e08`
  O-1 (convention + gate parses the trailer block + kernel-side bind check). Added 2026-11-16.
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
