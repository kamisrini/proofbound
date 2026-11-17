# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-16 (EOD)

## Resume — 2026-11-16

**Exact next action:** write **SPEC draft 4** of the R2 `intent.sdd` provider
(`docs/plans/R2-sdd-provider-spec-draft4.md`, superseding draft 3) addressing M0 review round 3
(`docs/verification/verdicts/r2-m0-review-round3-20261116.md`, NEEDS_WORK, 2 HIGH):
1. **Per-state unmappability** (F1): a fail-closed condition at a *past* commit makes only that
   state unmappable — skipped and reported, never a batch error — so no historical commit can block
   every future sync. Only a *claim* on such a state fails, at resolution.
2. **Walk = the git connector's commit set** (F2): `rev-list --all` semantics with `--full-history`,
   plus an invariant that every `Resolve` result is a member of `Revisions()`.
3. **Drop O-11** (change drift) from the proposed BR revision 0003 (F3); keep O-10 and O-12.
4. **Discriminating tests** for SDD-INV-8/11/13/24/26/27 (F4); fix F5 drift (slug/suffix defined in
   draft 4 itself, empty name-part, RENAMED as FROM/TO bullets, plan M1 text, name the intent SPEC §4
   amendment).
5. **Fixtures** (round-3 conditions): add a merge (incl. TREESAME), an unmerged-branch claim, the
   leftover-after-archive state (OpenSpec `7ced2a87`..`df12368f`), a root commit.
Then dispatch **M0 review round 4** (fresh non-author reviewer, same pattern as rounds 1-3). If
ACCEPTABLE → bring BR revision 0003 wording to the owner → **M1 code** (fixtures + replay helper
already committed in `kernel/internal/connector/intent/sdd/`).

**Owner priorities (2026-11-16):** build activities first; **checks, hooks, and pushes are held for a
few days** (`DEF-PUSH-HOLD-20261116`). Do **not** push without the owner's word.

**Immediate blockers:** none for draft 4. M1 code is blocked until an M0 review returns ACCEPTABLE
(Law 9).

**⚠ Known gate trap:** the two permanent-untracked paths (`docs/plans/P5-intent-provenance-plan.md`,
`kernel/cmd/p6dbprobe/`) keep the tree dirty, so bare `make check` goes RED on `commit-cadence` once
HEAD is >90 min old — by the clock alone. A commit clears it. Owner decision pending on those files.

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
  (change-scoped, history-native, measured on OpenSpec's 256 change-touching commits) rejected on
  2 HIGH, both tractable. Drafts kept as reviewed artifacts in `docs/plans/`.
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

`main` is **17 commits ahead of `origin/main` (`1de1b21`), unpushed by owner instruction.** Today's
work exists only locally and in the dated git bundle under `~/Backups/proofbound`.

## Open / deferred (`docs/deferrals.tsv`) — all dormant until the owner names them

- `DEF-PUSH-HOLD-20261116` — local commits held unpushed (owner, 2026-11-16).
- `DEF-INTENT-TRAILER-BLOCK-001` — gate: pre-push hook + round-3 LOWs.
- `DEF-INTENT-CI-REVISION-001` — change intents cannot supersede themselves (kernel limitation).
- `DEF-P7-REGENLOOP-001`, `DEF-P7-DP-TWINROOT-001` — unchanged.

## Unverified assumptions

- That draft 4's per-state unmappability is consistent with the intent connector's batch-validation
  contract (INT-INV-4) — likely needs an intent SPEC amendment; check before writing.
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
