# Proofbound — Current State

> Durable resume state. Rewrite this file in place; do not create a second state file.

As of: 2026-11-22

## Resume — 2026-11-22

- **U1 BUILT, not released: `proofbound ui`, the Audit room** (`a271e2d`, `CI-audit-room-u1-h5n9d2`,
  owner 2026-11-22 "lets build it, so we can ship it to clients"). U0 `381edea`: VD-audit-room-ui-q2v6x3,
  BD-audit-room-ui-k4m2p8 -> BR-audit-room-ui-r7t3w1 (O-1 Audit room, O-2 signed acts, `proposed`),
  spec `docs/plans/ui-u1-audit-room-spec.md` (draft 2), CLI-INV-37..41. Seven controls C1..C7 over
  base..HEAD; signed exception outcomes (`.proofbound/outcomes/`) and transcript attestations
  (`.proofbound/attestations/`) from the browser; token only via the terminal link (port-named
  SameSite=Strict cookie); signer key must be declared at the base and must not be the commit's author.
  Spec review `ui-u1-spec-review-20261122` NEEDS_WORK (F1, F2 HIGH, F3..F11 MED) and code review
  `ui-u1-code-review-20261122` NEEDS_WORK (F1..F4 MED): all fixed; mutants (Origin check, author rule,
  tag) killed. Gate GREEN before commit. Accepted on the tour's shop repo and the DEMO-MANUAL repo
  (all 19 blocks + new Step 15). Release notes v0.3.3 section marked unreleased.
  **Exact next action:** owner try-out (DEMO-MANUAL Step 15, or `proofbound ui` in any repo), then on
  owner OK cut v0.3.3 (tag, archives, GitHub release) as v0.3.2 was. Not in U1: other workspaces,
  time rail, in-browser signature check, `--bundle`, sample draw. New: `DEF-KERNELCHECK-SDD-TIMEOUT-001`
  (sdd package 442 s alone, near the 600 s go test limit under make check; failed the gate twice).
- **v0.3.2 PUBLISHED 2026-11-22** (owner: "i reviewed and accept, and do the rest"; owner accepted
  both CIs of `d8e0714`). Release commit `3c2995c` (`CI-mvp-release-v032-811d44`; review
  `mvp-release-v032-review-20261122` ACCEPTABLE, F1/F3 LOW fixed, F2 LOW recorded): install.sh warns on
  a shadowing older binary, QUICKSTART fixes, DEMO-MANUAL step 1b for the new hook, release notes
  (stricter check + hook upgrade step). Pushed main, tag + GitHub release `v0.3.2` (5 archives +
  SHA256SUMS); public `install.sh` installs v0.3.2 (verified); live Action re-run in
  `kamisrini/proofbound-action-test` installed v0.3.2 and reported WARRANTED. Deferrals
  `DEF-RELEASE-V032-001`, `DEF-QUICKSTART-STALE-001`, `DEF-QUICKSTART-EMPTY-REPO-001`,
  `DEF-INSTALL-PATH-SHADOW-001` RESOLVED. Landing page republished (card 05: fails the PR) and
  committed in the private repo. Owner's `~/.local/bin/proofbound` upgraded to v0.3.2.
- **Docs and demos updated for v0.3.2** (`b908ea5`, owner: "lets update all the docs, demo etc"): tour
  chapter 13 (hook refusal, `task start` stamping, `--no-verify` commit UNCOVERED), DEMO.md, DEMO-MANUAL,
  R2 DEMO usage, r3 README table; demo scripts put their own build on PATH for the hooks and drop the
  retired env var. Manual blocks, tour, demo-warranty (both kits), demo-share, demo-install all run
  clean on the published v0.3.2. Historical specs/plans under `docs/plans/` left as written.

- **Owner's two priorities built** (`d8e0714`, owner 2026-11-22 "lets do both one after other"):
  (1) `proofbound check` fails on UNCOVERED commits, i.e. unclaimed commits that change code
  (`CI-mvp-uncovered-code-581081`, CLI-INV-33; `DEF-CHECK-UNCLAIMED-CODE-001` RESOLVED); (2)
  `proofbound task start|stop` + init-installed prepare-commit-msg/commit-msg hooks stamp and require
  the claim; `PROOFBOUND_NO_INTENT` retired (`CI-mvp-agent-intent-3k7q2w`, CLI-INV-34..36). Review
  `mvp-uncovered-and-agent-intent-review-20261122` NEEDS_WORK (F1 HIGH, F2-F4 MED) all fixed by the
  author, F5-F7 LOW recorded; calibration mutant killed. Gate GREEN at `d8e0714`. Not released:
  ships with v0.3.2 (owner: no bump; release notes say check now fails PRs that passed before). Demo scripts still
  set the retired env var (harmless; owner: demos not to be fixed).
- **Also 2026-11-22:** gap deferrals `DEF-NO-INTENT-SELF-WAIVER-001`, `DEF-CLAIMED-CODE-SCOPE-001`,
  `DEF-AGENT-NATIVE-INTENT-001` (minimal slice now built; rest open); `DEF-P7-REGENLOOP-001` now
  requires an off-by-default opt-in flag for token-spending steps; landing page (private repo) gained
  a problem section with five measurable metrics (uncommitted there); DEMO-MANUAL demo folder moved
  under `~/projects`.

## Earlier — 2026-11-21

**Vertical MVP is under way (`VD-vertical-mvp-first-ai3dhn`, owner "go" 2026-11-20): five steps in
order — 1 installable, 2 PR check, 3 reviewer-signed warranty, 4 shareable proof, 5 commercial (kept private).
Deeper R3 organs are frozen until real users ask.** Chain: `BD-vertical-mvp-qbacah` →
`BR-vertical-mvp-xc31gp` O-1..O-5 (one per step).

- **R3 slice 1** (portable signed warranty) owner-accepted 2026-11-20 without the optional LOW fixes.
- **Step 1 DONE** (`a1794af`, pin refresh `a0ffc04`, `CI-mvp-install-k99zpl` → O-1): spec
  `docs/plans/mvp-1-install-spec.md`; `scripts/release-build.sh`, `install.sh`, `proofbound version`
  (CLI-INV-25), `docs/releases/mvp/{QUICKSTART.md,demo-install.sh}`. Install → first PASS 27 s
  (cold DB cache 23 s for the loop). Review `mvp-1-install-review-20261120` ACCEPTABLE; F1/F2 MED fixed;
  F4-F8 LOW recorded. Nothing published (pushes held).
- **Step 2 DONE** (`e9a0d48`, pin refresh `e0b3121`, `CI-mvp-pr-check-8hq2rm` → O-2): `proofbound
  check <base> [--markdown] [--require-claims]` (CLI-INV-26), `action.yml` + `scripts/action-check.sh`,
  `init` writes the tenant workflow (CLI-INV-18 amended), QUICKSTART §4. Review
  `mvp-2-pr-check-review-20261120` ACCEPTABLE, 6 LOW recorded (F1 worth the owner's eye: a mistyped
  `intent: SDD:x` claim is not listed, so such a PR reads NO CLAIMS and passes without
  `--require-claims`). The action's `${{ }}` expressions are unexercised until a real PR run.
- **Step 3 DONE** (`210d08d`, pin `8c7e2ef`, `CI-mvp-reviewer-sign-5tqv2c` → O-3; owner: "first A and
  then B"): verify check 9 — every named reviewer proven by their declared key via `.sig` beside their
  artifacts (A, written by `review seal`) or a cosignature (B, `warranty cosign`); else UNREVIEWED
  (exit 3); identity is `KeyName`; `PROOFBOUND_KEY`. Spec rounds: r1 NEEDS_WORK (1 HIGH, 5 MED, all
  answered) → r2 ACCEPTABLE; code review ACCEPTABLE (5 LOW); `internal/warranty` sweep 148/148 killed.
  R3 demo shows UNREVIEWED → cosign → VALID on both kits.

- **Step 4 DONE, locally** (`66ffd36`, pin `95b5589`, `CI-mvp-share-proof-w7c4na` → O-4; owner: "build
  it locally for now"): `proofbound warranty bundle <base> [-o dir]` writes signed warranty files plus
  `VERIFY.md` / `index.json` (both say unsigned; fingerprint out of band) and `badge.svg`; exit 0/3/1
  (CLI-INV-28..30). Demo `docs/releases/mvp/demo-share.sh`. Spec r1 NEEDS_WORK (3 MED) → r2 ACCEPTABLE;
  code review ACCEPTABLE (4 LOW). Nothing published.

- **Step 5 (commercial):** drafts exist, kept OUTSIDE this public repository by the owner's decision
  (2026-11-20: commercial material and contact details stay private); not tracked here.

- **Claim-typo fix** (`043afd7`, `CI-mvp-claim-typo-r4n8xk`): a mistyped `Intent:` claim now fails
  `proofbound check` (UNBOUND) instead of reading as NO CLAIMS. Review ACCEPTABLE (3 LOW).
- **v0.3.0 PUBLISHED 2026-11-20:** main pushed (`b783b61`), tag `v0.3.0`, GitHub release with 5
  archives + SHA256SUMS; `curl …/install.sh | sh` from GitHub installs v0.3.0 (verified). Before the push
  the owner had commercial material kept private: 25 unpushed commits were rewritten (filter-branch) to
  drop it; old history on local branch `backup/pre-scrub-20261120` (never push it); the material lives in
  `~/projects/proofbound-commercial/` (local git, never pushed).
- **Live Action test FAILED (real finding):** private throwaway repo `kamisrini/proofbound-action-test`
  PR #1. Install step worked; `proofbound sync all` on the runner failed: `v2 verdict evidence … is
  dangling` — `DEF-VERDICT-EVIDENCE-PORTABILITY-001` (verdict evidence = per-ledger random event id). The
  PR check cannot pass in CI for any reviewed commit until this is fixed. Known issue added to the
  release notes.

- **v0.3.1 PUBLISHED 2026-11-20 — the Action now passes in CI.** Owner authorized
  `DEF-VERDICT-EVIDENCE-PORTABILITY-001` ("go with recommendation"); fix `bddf9c0` (pin `0da19a8`,
  `CI-mvp-evidence-portable-6dq3ve` → O-2): seal cites `evidence_refs: ["git:commit.recorded:<sha>"]`,
  resolved on whichever ledger projects it (R-INV-14, CLI-INV-20 amended, CLI-INV-31); legacy event ids
  keep their ledger-local meaning (re-seal). Spec `docs/plans/mvp-evidence-portability-spec.md`: r1
  NEEDS_WORK (4 MED answered) → r2 ACCEPTABLE. Code review NEEDS_WORK (F1 MED: no test could fail on
  resolution) fixed by the author — reducer refuses SATISFIED with no resolved evidence, e2e asserts clone
  B's own event id; skip-resolution mutant killed; 2 LOW recorded. Row RESOLVED. Tag + release `v0.3.1`
  (5 archives + SHA256SUMS); public install gives v0.3.1. **Live re-test:** PR #1 in
  `kamisrini/proofbound-action-test` re-sealed with v0.3.1 → Action run installed v0.3.1 and reported
  WARRANTED (PASS). The test repo is kept.
- **Found, not fixed:** the integration-tagged projections harness (`go test -tags integration
  ./internal/projections/`) fails at `testStore` with `store: ledger tables are append-only` even at
  `98b33a7` (pre-existing; `make check` does not run integration tests).

- **R3 docs consolidated** (`25e8e9a`, owner 2026-11-20): `docs/releases/r3/README.md` is the one front
  door (nothing moved); `docs/releases/r3/tour.sh` + `DEMO.md` walk every v0.3.1 feature in 12 chapters
  with real output and why each matters. Found: `report week` fails on tenants with v2 verdicts
  (`DEF-REPORT-WEEK-V2-001`, OPEN, dormant). Gate GREEN at `25e8e9a`.

- **DEMO-MANUAL.md** (`de53647`, owner 2026-11-21): the tour typed by hand, 14 steps with business value;
  all 20 bash blocks run from the page against v0.3.1.
- **Owner trying the R3 README (2026-11-21), findings so far:** (1) FIXED `986bf32` (pin `190c102`,
  `CI-mvp-reviewer-frontmatter-05p936` → O-1; owner: "fix it now"): the `init`-written reviewer agent's
  front matter was invalid YAML (unquoted `Intent: sdd:` in `description`; VS Code preview rejected it);
  now quoted, CLI-INV-32. Review `mvp-1-reviewer-frontmatter-review-20261121` ACCEPTABLE, 2 LOW. Not
  released: users get it only with a v0.3.2 (needs the owner's OK to publish). (2) QUICKSTART §1/§4
  still say "Not published yet". (3) QUICKSTART assumes an existing OpenSpec/Spec Kit repo; nothing for
  an empty folder. (4) `install.sh` does not warn when an older `proofbound` earlier on PATH shadows
  the new one. (2)-(4) not fixed: `DEF-QUICKSTART-STALE-001`, `DEF-QUICKSTART-EMPTY-REPO-001`,
  `DEF-INSTALL-PATH-SHADOW-001`. The owner's old `~/bin/proofbound` was renamed `.old-20261114`.

**Exact next action:** the owner continues the DEMO-MANUAL trial on v0.3.2 and adds gaps as
deferrals. Fix new trial findings when asked. MVP step 5 (commercial) continues outside this
repository. Ask before starting other work.

**Gate:** GREEN at `b908ea5` (2026-11-22). main is pushed (in sync with origin). Untracked
`docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/` are the owner's to keep or
drop. **Immediate blockers:** none.

## R2 build log — 2026-11-19

**Now: R2 M3 implementation, under `CI-sdd-warranty-m3-xaqocp`** (owner-approved 2026-11-19, `260b563`).
- **Decision record:** `VD-sdd-m3-git-binding-2026-11-19` records D1-D6.
- **Contract:** the M3 SPEC, `docs/plans/R2-sdd-m3-git-binding-spec-draft5.md` as amended by drafts
  6-8. Non-author round 7 found it ACCEPTABLE (`fe8d116`).
- **Deferrals filed:** `DEF-VERDICT-STALE-OUTCOME-001` and `DEF-VERDICT-EVIDENCE-PORTABILITY-001`
  (owner-approved), beside `DEF-INTENT-GIT-ORDERING-001`.

**Done (units, each calibrated with hand mutants):**
1. **Git connector** (`9600441`, pin `033f5b7`):
   - sdd claims become per-commit `ClaimSkip`s with all-or-nothing refs;
   - git SPEC INV-29 is amended and INV-31/32 appended;
   - 3 sdd corpus rows are added, and the oracle counts a bound ref;
   - 9/9 mutants are killed.
2. **Projections** (`b993d0b`, pin `9cb920c`): provider `sdd` is accepted, under P-INV-33 with an
   untagged test, and the invariants lock is regenerated. 2/2 mutants are killed.
3. **CLI resolver** (`e596e06`):
   - the `sdd` arm, with the lazy provider, lazy `L` and the §4.2 checks;
   - claim-skip lines;
   - 14 mutants killed.
4. **Reviews hold and refusal, plus the CLI `storeLedgerView`** (`fc96b9c`):
   - D5/D6 per draft 8 §4.6: R-INV-13, CLI-INV-15/16;
   - 27 mutants killed;
   - the view is checked against this repo's own archived p5 claim.
- **Lesson compiled** (`7ffe671`): `port_home_test.go` fails `make check` if an untagged test outside
  the sdd package opens the default embedded port. Every end-to-end embedded test lives in the sdd
  package; cli tests use ephemeral ports.

5. **End-to-end procedures** (`6ac05c6`) in sdd `m3_e2e_test.go`:
   - §7.1, row 8: x1 PASS, x2/x3 BLOCKED on one ledger, and X3's verdict held;
   - §7.2: interleavings A and B with the five-record archive;
   - §7.3: in-place revisions refused, the rebuild green, then recovery to PASS;
   - 7 mutants killed.
6. **sdd SPEC M3 note** (`075a1dd`):
   - §12 records M3 as landed and SDD-INV-56 in substance;
   - §14 gets the M2 entry condition (monotone bindings);
   - the draft 10 copy stays byte-identical.

**Implementation review (non-author, foundational tier, two lenses):**
- **Lens A, spec-truth and laws** (`ed1744d`): NEEDS_WORK, 1 MED and 4 LOW. All five are answered in
  `9464e06`:
  - row 10 / M3-INV-10 is proven by `TestBindingsAreMonotone` (P-INV-34);
  - the view's error paths are tested;
  - the CLI half of malformed-after-refusal is tested;
  - cli uses `ValidateRevision` and the producer decoders (Law 2);
  - the port detector is rebuilt on go/ast, with 14 in-test probes.
- **Lens B, mutation and concurrency** (`8d6dbde`): ACCEPTABLE, 5 LOW. All five are answered in
  `af57109` (pins `75c9eac`), test-only plus one SPEC note, with every new test calibrated by hand
  mutants:
  - F1: fail-closed decode paths of the view and of L (C22/C39/C43/C46/C49 killed);
  - F2: printableUnit on claim-skip values and held/refused paths, and write errors (C28/C29/C54);
  - F3: the D1 100/101 bound and near-sdd values (sweep #39/#40, G1/G7/G8);
  - F4: the port guard also flags `Unsetenv("DATABASE_URL")` (multi-line evasion was already closed);
  - F5: the reviews SPEC R-INV-13 states the cross-root held-vs-refused limitation with a revisit trigger.

- **Non-author re-check** (`b9eda29`, over `eecd941`): ACCEPTABLE. All ten lens A/B findings are
  closed (B-F5 as an accepted limitation). Two new LOWs, **left on record by the owner (2026-11-19)**:
  - F1: residual blind spots in the default-port guard: a file-scoped `openStore` stub, the
    historical-evidence opener, an aliased store import, and a non-literal `store.Config`;
  - F2: held lines printed before a NON-refusal error are unpinned (mutant R2 survives).
- **Gate:** `go test ./...` (19 ok) and `make check` both exit 0 at `b9eda29`.

**M3 is accepted.** Lens B and the re-check are ACCEPTABLE, and lens A's findings are closed.

**Process changed 2026-11-19 (owner, `CI-review-budget-fs8f5h`):** Law 11 "Payoff first" and the
review budget: a spec gets at most 2 review rounds, code gets 1, and LOW findings are recorded, not
fixed. Pins now digest the staged code, so they commit WITH the change. The state-note limit is 10
commits.

**R2 step 1 done: the real-repo demo** (`CI-openspec-demo-mju3te`). Run `bash docs/releases/r2/demo-openspec.sh`
(about 4 minutes). It clones `Fission-AI/OpenSpec` at a pin and commits the profile into every commit
of the local clone (O-3 needs a committed profile). It ingests 648 spec revisions from 1,059 real
commits, binds a demo commit to `install-skills-globally`, and shows both demo gates BLOCKED. The
claimed change shows real obligations, all UNVERIFIED. No kernel code changed.

**Rough edges fixed** (`CI-demo-polish-i3ew84`): the embedded server's output goes to `.proofbound/db.log`
(store INV-52); readiness gates print `reason="..."` with their counts (GATE-INV-16);
`report intent sdd:<change-name>` looks a change up by name (CLI-INV-17). The demo moved to
`docs/releases/r2/` (`demo-openspec.sh` + `DEMO.md`, R1's layout) and is linked from the README.

**Starter kit and reviewer workflow built** (`CI-starter-kit-68z7m8`, spec
`docs/plans/R2-starter-kit-spec.md`):
- `proofbound init openspec` writes a tenant's profile, `.gitignore` lines, AGENTS.md block, Claude
  Code reviewer subagent, and commit-msg hook;
- `proofbound review draft <commit>` / `review seal <dir>` turn a reviewer's decisions into review
  and verdict artifacts plus that commit's readiness gate (CLI-INV-18..20).
- Proven end to end on the embedded ledger (`TestStarterKitReviewFlow`: SATISFIED gives PASS,
  INCONCLUSIVE gives BLOCKED), and by hand on the OpenSpec demo clone, where a sealed review turned
  the claimed commit's gate to PASS.
- Usage: `docs/releases/r2/DEMO.md` "Use it on your own OpenSpec project".

**Review:** `r2-starter-kit-review-20261119` (`592a52e`), NEEDS_WORK on one MED: seal trusted the
editable draft. Fixed: seal now rebuilds all but the decisions from the ledger. 3 LOWs recorded; the
mid-write rollback LOW was fixed with it.

**Re-check** `r2-starter-kit-recheck-20261119`: ACCEPTABLE, F1 closed. **The starter kit is accepted.**

**M4 drift built** (`CI-sdd-warranty-m4-dcom6f`, spec `docs/plans/R2-m4-drift-spec.md`, BR rev 0004 O-13):
`proofbound report drift` compares a change's promise (obligation ids and statements), not digests.
This was measured on OpenSpec's real history: every archive move changes every digest, yet 60 of 73
promise the same thing. Proven by `TestDriftComparesPromises` and `TestDriftEndToEnd` (verify,
reword gives STALE, re-verify gives CURRENT, archive stays CURRENT), and on the OpenSpec demo clone
(a reworded obligation is reported as `changed` with the old and new text).

**M4 review** `r2-m4-drift-review-20261119` (`441577c`) found one HIGH: "current = highest ledger seq"
fails after a revert, because the ledger dedupes the re-emitted bytes, so a never-verified promise read
CURRENT. **Fixed** (`CI-sdd-warranty-m4-head-4ut53f`, BR rev 0005 O-14 replaces O-13): drift reads each
change's revision at HEAD via the new `sdd.Provider.StateRevision` (SDD-INV-68). Tests cover both
revert directions, and the e2e asserts the revert appends nothing.

**Re-check** `r2-m4-drift-recheck-20261119`: ACCEPTABLE, F1 closed. **M4 is accepted.**

**Spec Kit built** (`CI-sdd-warranty-speckit-eiis5s`, spec `docs/plans/R2-speckit-spec.md`):
- the provider gains a `speckit` profile shape and mapping path: a feature `specs/<name>/` gives one
  requirement record from `- **FR-001**:` lines, and a change intent once `tasks.md` exists;
- ids hash the kit, and OpenSpec ids are unchanged;
- `proofbound init speckit`;
- SDD-INV-58/64 retired, SDD-INV-68..71 numbered.
- Proven by `TestSpecKitHistoryMapsExactly`, `TestSpecKitProfileIsStrict`, `TestSpecKitResolveAndIDs`
  (three mutants killed) and `TestSpecKitEndToEnd` (init, claim, review, gate PASS, drift CURRENT,
  then STALE on an in-place edit).
- Spec Kit's own repo does not use its layout, so no real-repo run; its template confirms the format.

**Review** `r2-speckit-review-20261119`: ACCEPTABLE. OpenSpec ids, digests, skip order and Resolve are
unchanged, and a 15-shape hostile probe gave no wrong record and no panic. 3 LOWs recorded, not fixed: a
near-miss FR line format is dropped silently; the sibling can be any segment; the init profile test
checks JSON only (the e2e covers acceptance). **Spec Kit is accepted.**

**Docs audit (2026-11-19, owner request):**
- New `docs/releases/r2/demo-loop.sh {openspec|speckit}` runs the full loop (claim, review, PASS,
  drift) on a tiny project in about 2 minutes.
- DEMO.md rewritten: setup, three demos, own-project steps, a command reference.
- README, RELEASE-PLAN, ROADMAP and the R2 plan say R2 is built (accepted later the same day).
- Literal walkthroughs found and fixed two things:
  1. `sync all` does not read specs (SDD-INV-56, by design), so every instruction now says
     `sync intent sdd && sync all`, including the AGENTS block `init` writes.
  2. **`sync all` failed on every Spec Kit project:** the P5 `specdir` provider is also rooted at
     `specs/`. Fixed as CLI-INV-22 (omit `specdir` when HEAD's profile is Spec Kit at `specs/`).
- All three demos and the documented own-project path were re-run clean. The CLI-INV-22 fix passed
  its scoped non-author review (`r2-speckit-specdir-review-20261119`, ACCEPTABLE).

**R2 release notes drafted** (`docs/releases/r2/RELEASE-NOTES.md`, suggested tag `v0.2.0`), linked from
the README and the R2 demo hub.

**Owner accepted Release 2 (2026-11-19).**

**Commit trailers:** every M3 commit carries `Authorized-By: Srinivas Kamisetty` and
`Intent: CI-sdd-warranty-m3-xaqocp`, adjacent to `Co-Authored-By`.

**Immediate blockers:** pushes are held (DEF-PUSH-HOLD-20261116).

**M1 is closed and accepted.** `sdd` is the 18th Task-8 package. The M1 LOW findings and the
re-check's answers were accepted, ending with the owner's verdict `r2-m1-lowfix2-owner-20261118` in
`b18e672`.

**Gotchas:**
- invariant-lint resolves `file_test.go::Test` citations only inside the SPEC's own package.
- invariant-table-lint takes exactly one citation per §9 row.

**Immediate blockers:** pushes are held (DEF-PUSH-HOLD-20261116). Do not push without the owner's word.

**Draft 9 (2026-11-17)** answers round 8 (NEEDS_WORK, 2 MED + 5 LOW): record units (pre-parse skips and
unpaired-only deltas count → `cascade`); change entries are trees/symlinks/gitlinks only — plain files
under root/archive ignored (real `IMPLEMENTATION_ORDER.md` at 1393/1624 commits; INV-66); name domain
inside §5; archive = direct child of root; `requirements.paths` fixed to the exact pair; literal
pathspecs = defence in depth, not claimed tested; INV-63 → INV-65; uniform Skip-digest definition.

**Draft 8 (2026-11-17)** answers round 7 under **`VD-sdd-m1-openspec-first-2026-11-17`** (owner: M1 =
OpenSpec only, Spec Kit → M2 parked in Appendix A; literal profile path grammar + `--literal-pathspecs`).
One `changeOutcome` function with precedence drives both `Resolve` and `Skipped` (name-reuse covers
archived-twice and re-archive); new `resolve-edges` fixture + `expected-resolve.json`; state key =
mode+oid of profile and roots (INV-63); `id-vectors.json` with non-ASCII inputs (`--vectors`);
`--pins`/`--vectors` standalone with `@generated by` markers; skip de-dup by (unit, cause, digest).
**Raised to owner, not filed:** (1) git connector fails on non-UTF-8 paths in history; (2) a
`records`/`specdir` U+0000 value passes ValidateRevision and wedges every rebuild (draft 8 §14).

**Draft 7 (2026-11-17)** answers round 6 (NEEDS_WORK, 2 MED + 5 LOW): `ErrNoChangeIntent` + a change-name
domain make `Resolve` total (INV-59 re-issues INV-50); the U+0000 rule moved out of `ValidateRevision`
(rebuild re-validates history — §10 I: only retroactively-safe checks go there); state-level object
types (INV-61); byte-wise ASCII case mapping; `archive-name-invalid`; CLI prints skips (INV-60);
naive toggle defined + `history_probe.py --pins` regenerates `expected-obligations.json`; self-test
rejects causes outside §7's list.

**Draft 6 (2026-11-17)** answers round 5 (NEEDS_WORK, 3 MED + 4 LOW, no HIGH). The §7 skip-cause table is
derived from every check tenant bytes reach (adds `nul-byte`, `statement-blank`, `obligation-id-invalid`,
`non-regular-file`); `ValidateRevision` rejects U+0000 + oversize for all providers; the reader shares
gitcmd's listing but not its per-commit file lists; §4.2 is re-pinned to OpenSpec's `parseDeltaSpec`
with JS regex classes made explicit; Spec Kit `changes.status` is declared. **Owner decisions
2026-11-17:** `sync intent all` excludes `sdd` (SDD-INV-56); the replay helper is committed. Fixtures:
fence-precision cases, `head-on-leftover`, `expected-obligations.json`, a mechanised pin test.
**Raised to owner, not filed:** the git connector itself fails on any non-UTF-8 path in history
(pre-existing; draft 6 §14).

**Draft 5 (2026-11-17)** answers round 4 (NEEDS_WORK, 6 MED + 4 LOW, no HIGH): closed list of tenant-data
skip causes (incl. `payload-too-large`, `invalid-utf8`); everything else fails `Revisions`;
`ValidateRevision` gains the size bound for all providers; complete grep-derived amendment list (§12,
M1 vs M3 sites); the reader is `gitcmd.Repo` (replace/graft/shallow contract); fence/section rules
pinned to OpenSpec's code; INV-1/3/4/6/25/29/36/37/38 retired and re-issued as INV-40..52; tombstone
reasons restored; profile schema inlined; Resolve sentinels. Fixtures: `branches/14-stale-tip`, the
column-0 fenced case (masked 3 / unmasked 1), `expected-skips.json`. Probe re-measured: 23
dual-presence commits on the provider walk, 4 branch tips, 1 of 401 delta blobs changes under masking.

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

**Immediate blockers:** none for draft 5. M1 code is blocked until the round-4 findings are fixed and
a non-author re-check passes (Law 9).

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

`main` == `origin/main` after this note's commit (public, `kamisrini/proofbound`); tags `v0.3.0`, `v0.3.1`.
Local-only branch `backup/pre-scrub-20261120` must never be pushed.

## Open / deferred (`docs/deferrals.tsv`) — all dormant until the owner names them

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
