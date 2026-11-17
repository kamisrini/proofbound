# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-17

## Resume — 2026-11-17

**Exact next action:** get the owner's word on committing the extended replay helper
(`kernel/internal/connector/intent/sdd/fixture_test.go` — governed path, needs an `Intent:` +
`Authorized-By:` trailer; precedent `4d91413` used `CI-sdd-warranty-m0b-ti5ga9`), then dispatch
**M0 review round 4** (fresh non-author reviewer, same pattern as rounds 1-3) on
`docs/plans/R2-sdd-provider-spec-draft4.md` + the new fixtures. If ACCEPTABLE → bring BR revision
0003 wording (draft 4 §11: O-10 reworded, O-12; **O-11 dropped**) to the owner → **M1 code**.

**Draft 4 (written 2026-11-17)** answers round 3: fail-closed **per state** with reported `Skip`s and a
cascade (never batch-fatal; INT-INV-4 holds unchanged because the provider pre-validates); the walk is
the git connector's own `rev-list --all`, unsimplified, deduplicated by SDD state key (OpenSpec: 1624
commits → 338 states); "archived by" = absent from **every** parent (merge / root cases defined); drifted
invariants retired and re-issued (INV-8/11/14/24/27 → INV-31..39); slug/suffix (`kit`, `kind` literals)
defined in-draft; RENAMED read as FROM/TO pairs; intent SPEC §1/§2/§4 amendments named (§12); M3 entry
conditions named (§14, incl. the git connector's own unresolvable-claim wedge). **New beyond round 3:**
re-running the probe found fenced documentation examples (`## RENAMED Requirements` inside a code fence,
47 lines) that an unmasked parser would mint as renames → code fences are masked (SDD-INV-39).

**Fixtures added:** `testdata/branches/` (13-step commit graph: root-with-archive, TREESAME `--no-ff`
merge, side-branch archive + merge, unmerged-branch claim, leftover-after-archive) and `testdata/fenced/`;
the replay helper gained `PARENTS`/`REF`; `TestBranchFixtureReplaysAsDocumented` passes and was
calibrated (breaking the TREESAME revert makes it fail). The probe now prints every draft-4 §0 count.
**Found + fixed:** root `.gitignore` (`.proofbound/`) had kept every fixture profile out of git since
`4d91413`; a re-include rule now tracks all 29, and `requireProfiles` (in the held helper) guards it.

**Owner priorities (2026-11-16):** build activities first; **checks, hooks, and pushes are held for a
few days** (`DEF-PUSH-HOLD-20261116`). Do **not** push without the owner's word.

**Immediate blockers:** the helper commit needs owner authorization (governed path). M1 code is blocked
until an M0 review returns ACCEPTABLE (Law 9).

**⚠ Known gate trap:** the two permanent-untracked paths (`docs/plans/P5-intent-provenance-plan.md`,
`kernel/cmd/p6dbprobe/`) keep the tree dirty, so bare `make check` goes RED on `commit-cadence` once
HEAD is >90 min old — by the clock alone. A commit clears it. Owner decision pending on those files.
Committing under `kernel/internal/connector/` re-stales the `p6-task2-close` pin: run
`scripts/p6-task2-close.sh --write` and commit the refresh.

## Where R2 stands

- **Release line:** R1 (Trust Record) accepted as shipped; **R2 = SDD Warranty Connector**; R3 = the
  revenue organs (`VD-release-line-r2-sdd-t5j34p`, amended).
- **R2 scope (owner, option 2):** warrant OpenSpec changes / Spec Kit features against the
  requirements they declare, read **history-natively**; main-spec lifecycle deferred
  (`VD-sdd-change-scoped-r2-k7q3wz`). Accountable identity from the tenant's committed profile
  (`VD-sdd-accountable-identity-ypcx7x`).
- **Intent home:** `BD-sdd-warranty-jdyjhh` → `BR-sdd-warranty-i9miv7` rev 0002 (O-1..O-5, O-8, O-9
  active; O-6/O-7 retired; `proposed`) → `CI-sdd-warranty-m0-a49f5b`, `CI-sdd-warranty-m0b-ti5ga9`.
- **Contract history:** draft 1 (snapshot) and draft 2 (fold/tombstones) rejected by review; draft 3
  (change-scoped, history-native) rejected on 2 HIGH (round 3); **draft 4** (per-state fail-closed,
  git-connector walk) awaits round 4. Drafts kept as reviewed artifacts in `docs/plans/`.
- **M1 test assets:** `kernel/internal/connector/intent/sdd/testdata/` (OpenSpec 10-step and Spec Kit
  3-step commit-sequence fixtures, hostile trees, derivation probe) + `fixture_test.go` replay helper
  and calibrated self-test. **Profile shape in the fixtures is provisional** until the contract lands.

## Intent-trailer gate

`DEF-INTENT-TRAILER-BLOCK-001` OPEN. Shipped: kernel-grammar port, shared 32-case corpus with
end-to-end kernel test, unpushed-range checking, strict UTF-8, explicit `make check` step
`intent-trailer-gate`, proven `delivery-enforce` invocation, `scripts/mutant-apply.sh`. Round 3
(`intent-trailer-gate-review-round3-20261116`) NEEDS_WORK: any pre-push check is skipped by pushing
first; owner chose a **local pre-push hook** (residual: `--no-verify`, uninstalled clones) — held
with the other checks. Round-3 LOWs open: six unpinned mechanisms, Perl rejects noncharacters Go
accepts, quoted paths in `intent-applicability.sh`, `mutant-apply` comment classification.
**Convention:** `Intent:` directly above `Co-Authored-By:`, no blank line.

## Branch status

`main` is **18+ commits ahead of `origin/main` (`1de1b21`), unpushed by owner instruction.** Today's
work exists only locally and in the dated git bundle under `~/Backups/proofbound`.

## Open / deferred (`docs/deferrals.tsv`) — all dormant until the owner names them

- `DEF-PUSH-HOLD-20261116` — local commits held unpushed (owner, 2026-11-16).
- `DEF-INTENT-TRAILER-BLOCK-001` — gate: pre-push hook + round-3 LOWs.
- `DEF-INTENT-CI-REVISION-001` — change intents cannot supersede themselves (kernel limitation).
- `DEF-P7-REGENLOOP-001`, `DEF-P7-DP-TWINROOT-001` — unchanged.

## Unverified assumptions

- Checked 2026-11-17: per-state unmappability needs **no** INT-INV-4 change (the provider pre-validates
  and cascades); it needs an additive intent SPEC §2 `Skip` report (draft 4 §12).
- Round 4 has not yet confirmed that skips being derived-not-ledgered is acceptable (draft 4 §7).
- That a history walk over `rev-list --all --full-history` stays fast on real tenants (OpenSpec:
  ~7 s, reviewer-measured).

## Institutionalized today

- `scripts/mutant-apply.sh` (+ self-test in `make check`): refuses no-op / comment-only mutants —
  compiled response to `LESSON class=fixture-not-what-it-claims` (3 sightings).
- Trailer gate checks the unpushed range and runs as an explicit `make check` step.
- Evidence-probe rule (journal): every metric a document cites must be a line the probe prints, and a
  zero must be shown reachable by a run that finds a non-zero.

## Key facts

- **Never pipe `make check`** — the pipe swallows make's exit code.
- **Validate uncommitted intent records** with the overlay test pattern, not `proofbound sync` (sync
  reads committed HEAD only).
- `block-secrets` refuses a command containing both `docs/decisions/INDEX.md` and a redirect / `>`
  (including `<noreply@…>` in a heredoc) — stage `docs/decisions/` as a directory; run `make index`
  alone.
- Committing under `kernel/internal/connector`, `scripts/delivery-enforce.sh`, or its test requires
  `scripts/p6-task2-close.sh --write` + a pin commit afterwards.
- Census: live docs must not backtick a kernel path that is not a tracked file; kernel SPECs may not
  cite tests that do not exist (spec-first drafts live in `docs/plans/`); new `docs/gates.md` rows
  need a census row + `p6-task1-close.sh` count bumps + `--write` regeneration.
- `specfirst` walks into `testdata/` — fixture stand-in code must not use `.go`.
