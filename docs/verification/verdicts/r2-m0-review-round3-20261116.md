---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round3-20261116
status: NEEDS_WORK
reviewed_commit: d440bd4156f25861556589cde01fd10c044fd545
findings:
  - finding_id: r2-m0-review-round3-20261116-F1
    severity: HIGH
  - finding_id: r2-m0-review-round3-20261116-F2
    severity: HIGH
  - finding_id: r2-m0-review-round3-20261116-F3
    severity: MED
  - finding_id: r2-m0-review-round3-20261116-F4
    severity: MED
  - finding_id: r2-m0-review-round3-20261116-F5
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round3-20261116.md
artifact_sha: 3c65f710c1a49f76240196ee00824c49fda16a328a428a9bf4beaacf9c9c1283
---

# Release 2 M0 (SDD Warranty Connector): draft-3 change-scoped, history-native provider contract. Review round 3

## Provenance

- **Under review:** the documentary M0 round-3 deliverables at HEAD `d440bd4`, introduced by
  `d4dc7c1` (draft 3, the scope VD, the CI record, and the drift fixes) and `d440bd4` (a gate pin
  refresh). Files:
  - `docs/plans/R2-sdd-provider-spec-draft3.md`
  - `docs/decisions/VD-sdd-change-scoped-r2-k7q3wz.md`
  - `docs/intent/records/change-intents/CI-sdd-warranty-m0b-ti5ga9/0001.md`
  - the amendment paragraph in `kernel/internal/connector/intent/SPEC.md` §1
  - `docs/plans/sdd-warranty-connector-plan.md` (M0 status)
  - the appended correction in `docs/verification/r2-m0-mapping-spike.md`
- **Follows up:** `docs/verification/verdicts/r2-m0-review-round2-20261116.md` (NEEDS_WORK, F1..F10).
- **Judged against:** CLAUDE.md Build Laws, and this ground truth:
  - `kernel/internal/connector/intent/SPEC.md` §1-§5 and `intent.go` (`ValidateRevision`,
    `validateBatch`, `orderBatch`, the `Sync` sort at lines 176-181, `recordIDRE` / `obligationRE`)
  - `kernel/internal/connector/git/connector.go` (`resolvedCommit`, lines 180-214), its SPEC (INV-29 /
    INV-30), and `gitcmd/gitcmd.go` (the commit walk, line 63)
  - `kernel/internal/projections/projections.go` (commit-intent binding at 216-227, change-intent
    targets at 358-372, `reduceObligationVerdict` at 452-544, `reduceRequirementReview` at 546-590)
  - `kernel/internal/gates/gates.go:285` and `gates/SPEC.md` (intent-delivery-readiness)
  - `core/core.go` (`IdempotencyKey`)
  - `docs/intent/records/requirements/BR-sdd-warranty-i9miv7/0002.md`
- **Real-history evidence:** a full clone of `Fission-AI/OpenSpec` at
  `8bd4b3557b5bf1d231cd2f165b28f24e2d90a2a0` (1052 commits, 62 merges). I wrote my own probe
  (`<scratch>/reviewer-m0r3/probe.py`), run it under both git walk modes (default path-limited
  simplification and `--full-history`), and also ran the author's probe unchanged for comparison. I
  read the author's probe and checked its method; one defect is reported in F1. Every quoted example
  was checked by hand with `git ls-tree`, `git show`, or `diff`. A synthetic five-commit repository
  (`<scratch>/reviewer-m0r3/syn`) reproduces the walk defect in F2.
- **HEAD at dispatch:** `d440bd4156f25861556589cde01fd10c044fd545`. **HEAD at completion:**
  `d440bd4156f25861556589cde01fd10c044fd545` (unchanged). The untracked
  `kernel/internal/connector/intent/sdd/` fixtures that the author was writing concurrently are not
  under review.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. The subject owns record identity and point-in-time resolution, so M1
  should be reviewed at the foundational tier, as round 2 recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261116-m0-r3`, a **non-author**. It wrote
  none of the work under review, so this verdict counts under Law 9.
- **Writes:** no repository file other than this verdict. No cleanroom redactions. The session scratch
  directory is written as `<scratch>` (2 occurrences in this section).

## Round-2 findings re-verified

- **F1 (fail-closed fold makes real repositories unsyncable): WITHDRAWN BY SCOPE, but the class
  recurs.** No fold exists any more. However, the history-native walk makes any fail-closed condition
  at any past commit permanent, and the reference repository contains one such condition (new F1).
- **F2 (tombstones need text that is absent from the tree): FIXED (by scope).** Records are
  content-addressed, so nothing needs a tombstone. Intent SPEC §5 holds trivially: each record has
  one revision.
- **F3 (fingerprinted statement undefined): FIXED (by scope).**
  - Obligation ids no longer carry a fingerprint.
  - The statement is now defined (§4.2: first non-empty line, exact bytes, else the name).
  - Residual: the rule for RENAMED does not match OpenSpec's real format (F5).
- **F4 (A2 and O-9 unsatisfiable; drift blind): PARTIALLY FIXED.**
  - A2 is withdrawn, and O-9 is proposed for retirement.
  - The proposed replacement, O-11, reproduces round 2's "stale and can never clear" defect for
    archived changes, and its "current revision" cannot be observed in the ledger (new F3).
- **F5 (HEAD-only sync leaves commits dangling): PARTIALLY FIXED.**
  - History-native ingestion closes the linear case. On OpenSpec, every change-directory state at
    all 1052 reachable commits occurs at some walked commit.
  - The walk is still not the git connector's walk: HEAD versus `--all`, and the history
    simplification mode is unspecified (new F2).
- **F6 (digest does not pin payload): FIXED.**
  - The profile is in both manifests (§4.3, §5).
  - A change intent's targets derive only from files inside its own directory plus the profile, so
    the projector's last-write-wins keys `(source, id, artifact_sha256)` can no longer change meaning
    under one key.
  - Residual: SDD-INV-26's test proves only one direction (F4).
- **F7 (mid-change OpenSpec flows): FIXED (by scope).**
  - Main specs are not targeted, so early sync and rename-plus-modify have nothing to break.
  - A change with no deltas yields no intent (SDD-INV-28).
- **F8 (anti-spoofing; locked `Resolve` signature): DEFERRED TO M3, ACCEPTABLE WITH CONDITIONS.**
  - §6 states the right principle: the compared native name comes from the connector's own read of
    the tree, never from the provider.
  - M1's `Resolve(ctx, commit, name) (intent.Revision, error)` does not foreclose that, provided M1
    does not wire `sdd` into the git connector (see Fitness for M1).
- **F9 (cross-document drift): MOSTLY FIXED.**
  - Fixed: the intent SPEC amendment, the plan's M0 status, and the spike now point at draft 3. The
    spike's correction note is accurate.
  - Residual: the plan's M1 interface text is stale, and the suffix `kind` literal is still
    undefined (F5).
  - New: §0 states one count that is false (F1).
- **F10 (BR not fit for review): PARTIALLY FIXED.**
  - Revision 0003 is proposed, not made, so O-5's stale wording is still the live record.
  - SDD-INV-13's untestable half ("kit read only from the profile") is unchanged.
  - Change-name reuse is now addressed, but the measurement behind it is wrong (F1).

## §0 evidence reproduced

My probe was written independently. It parses delta files with OpenSpec's own case-insensitive
section regex (`spec-structure.ts:8`) and also recognises RENAMED `FROM:` lines.

| §0 row | Author | Reviewer (default walk) | Reviewer (`--full-history`) |
|---|---|---|---|
| Commits / commits touching `openspec/changes` | 1052 / 256 | 1052 / 256 | 1052 / 296 |
| Changes observed active / archived | 111 / 108 | 111 / 108 | 111 / 108 |
| Archived, never observed active | 16 / 108 | 16 | 16 |
| Changes with zero requirement entries at some commit | 34 | 34 | 34 |
| Within-change requirement removals | 47 events, 10 changes | 47 / 10 | 47 / 10 |
| Within-change statement changes | 10 events, 5 changes | 10 / 5 | 10 / 5 |
| Requirements with no statement line | 4 | 4 | 4 |
| MODIFIED / REMOVED targets missing from the main spec | 1044 / 6 | not measured (main specs are out of scope) | — |
| **Change names reused** | **0** | **1 name, active and archived together at 9 walked commits** | **12 commits** |

Further measurements:

- **The 16 never-active changes.** I sampled 7 of their archiving commits. All are single-parent and
  carry 2-24 files outside `openspec/` (for example `38d23568`: 24 files, `af273b8e`: 19 files). That
  supports the squash-merge reading.
- **Obligation ids.** There are no collisions under the §4.2 rule in any delta file at any walked
  commit. No name-part is empty or over-length.
- **Walk coverage.** Every change-directory state at every reachable commit appears at a walked
  commit, in both walk modes. The F2 gap does not occur in this repository.
- **Ledger size.** The repository's history yields 390 distinct requirement records and 259 distinct
  change-intent states before any profile edit. The walk takes about 7 s in Python.

All counts except one reproduce. The exception is "change names reused: 0".

## Findings

### r2-m0-review-round3-20261116-F1 (HIGH): §0's "0 name reuses" is false, and under draft 3 the reference kit's own repository becomes permanently unsyncable. History-native ingestion makes every past fail-closed condition permanent

**Broken property:** SDD-INV-20 ("archiving a change never makes a later sync fail"), the §0
evidence claim, and the claim in the VD that §0 "was measured on the real history before it was
written". This is the same failure class as round-2 F1.

**Evidence:**
- **The case in OpenSpec.** `7ced2a87` (2025-08-20, "archive: apply delta-based spec updates and
  archive changes") moved `adopt-verb-noun-cli-structure` into
  `changes/archive/2025-08-19-adopt-verb-noun-cli-structure/`. It left
  `changes/adopt-verb-noun-cli-structure/specs/cli-list/spec.md` behind. The leftover file is
  byte-identical to the archived copy (`diff` is empty) and holds three `MODIFIED` requirement
  entries.
- **How long it lasted.** The leftover was removed only by `df12368f` (2025-09-06). In between, the
  name is present both active and archived at 9 walked commits (12 under `--full-history`).
- **Why the batch fails.** §5 makes that state `intent.ErrUnmappable`, and SDD-INV-29 tests exactly
  this. §12 and SDD-INV-14 require that an `ErrUnmappable` returns no revisions, so the whole sync
  fails.
- **Why it is permanent.** `Revisions` walks the whole history on every sync (§3), so the same past
  commits fail every sync forever. The tenant cannot edit a past commit, so a history rewrite is the
  only remedy.
- **Why the probe missed it.** The author's probe never measured reuse. Its line 94 is
  `reused = {…} & set()`, which is always empty, and the value is never printed. The probe's
  `archived_names` tracking also cannot see a name that is active and archived at once.
- **The general case.** A HEAD-only provider could recover from a fail-closed error by a new commit.
  Under history-native ingestion, every fail-closed rule in draft 3 becomes a permanent wedge once one
  commit exhibits it:
  - an invalid profile at any commit (§2);
  - an owner or sponsor that the profile does not cover;
  - an obligation-id collision;
  - an unmapped Spec Kit status;
  - name reuse.

  Draft 3 has no invariant and no test for "a historical fail-closed commit can be recovered from".
- **The test that should catch this cannot.** SDD-INV-20's test uses an author-built history, so it
  passes while the property is false for the reference repository.

### r2-m0-review-round3-20261116-F2 (HIGH): the walk §2/§3 lock is not the set of commits the git connector binds, so a tenant's claim can resolve to a revision absent from the batch and stop the projector

**Broken property:** §3's "Every commit whose trailer the git connector will resolve (§6) has its
exact revision in the batch". This claim carries the fix for round-2 F5, and it rests on
intent SPEC §4 point-in-time resolution.

**Evidence:**
- **The two walks differ.**
  - The git connector ingests `git rev-list --reverse --date-order … --all`
    (`gitcmd/gitcmd.go:63`). That is every branch and tag, not just HEAD.
  - §2 locks `Reader.Commits` to "commits reachable from HEAD that touch any pattern".
  - A commit on an unmerged feature branch carrying `Intent: sdd:X` is therefore resolved by the git
    connector at its own tree, while its change-directory state has never been walked. This is the
    normal workflow: propose and implement on a branch, then run verdict CI before merge.
  - The resulting `CI-X@digest` is absent from the ledger. `projections.go:216-225` returns
    `dangling intent revision`, which fails the projection transaction, not just one row.
- **Even within HEAD, the walk is underspecified.**
  - "Touch any pattern" names no history-simplification mode. The author's probe uses git's default
    path-limited simplification (`rev-list … HEAD -- openspec/changes`).
  - That mode prunes the side of a merge when the merge is TREESAME to one parent.
  - Reproduction (`syn`): c0 creates change `x` on main; on a branch, `b1` edits `x`'s delta, then a
    commit claims `sdd:x`, then `b2` reverts the edit; a `--no-ff` merge follows.
    - `git log -- openspec/changes` lists only `c0`.
    - `--full-history` lists `c0`, `b1` and `b2`.
  - So the claiming commit's revision (the tree at `b1`) is never emitted, even though the commit is
    reachable from HEAD.
- **Not visible on OpenSpec.** Neither gap occurs in OpenSpec (0 missing states in both modes), so
  §0 could not have shown it.
- **The tests cannot catch it either.**
  - No SDD invariant states the property that matters: for every commit the git connector ingests,
    and every name, the result of `Resolve` is a member of `Revisions()`.
  - SDD-INV-24's test is linear (F4).

### r2-m0-review-round3-20261116-F3 (MED): proposed O-11 (change drift) flags most archived changes stale permanently, and "current revision" cannot be observed in the ledger

**Broken property:** BR revision 0003 O-11 as proposed (§10), and §9.A's "the change's own verdicts
remain valid history". This repeats round-2 F4 ("its stale flag can never clear").

**Evidence:**
- **Archiving restales the change.**
  - Requirement ids are content-addressed over a manifest that includes the path and the profile
    (§4.3), so archiving mints new requirement ids. §6 step 8 shows this ("archived-location
    records").
  - CI-X's archived revision therefore targets records that no pre-archive verdict covers. Under
    O-11 the change is stale.
- **Nothing can clear it.**
  - A verdict needs a commit claim that resolves to that revision (`projections.go:465-475`).
  - By SDD-INV-27, after the archiving commit no commit can ever resolve X.
  - In OpenSpec, 92 of 108 changes were archived in a commit separate from their active life, so
    they would be reported stale permanently.
- **Profile edits restale everything.** Any profile edit, for example adding a sponsor, re-mints the
  records and the change-intent revision of every archived change, with the same permanent
  staleness.
- **"Current revision" is not observable.**
  - Successive change-intent revisions carry empty `supersedes` (§5).
  - `Sync` appends a batch in `(source, kind, id, digest, path)` order (`intent.go:176-181`), not
    history order. A revert from state A to B and back to A is de-duplicated to two revisions.
  - Neither ledger sequence nor payload records which revision is current, so O-11 cannot be
    evaluated from ledger data, as the plan requires (M4 "a projection/report over existing data").
- **Scope.** O-11 belongs to M4, and the BR wording awaits the owner. This does not block M1 code,
  but O-11 is not yet a satisfiable obligation.

### r2-m0-review-round3-20261116-F4 (MED): several §8 proving tests cannot fail while their property is false, and two kept invariant titles changed meaning in place

**Broken property:** the Law 9 corollary (a test that cannot fail proves nothing) and the permanence
rule for invariant identifiers in §8.

**Evidence:**
- **SDD-INV-24** ("a change opened and archived between two commits of the walk is still emitted").
  - Archived directories persist at HEAD, so a HEAD-only implementation emits that change (as
    archived) and passes.
  - The property that needs proving is that the *active-window* revision is emitted. Concretely:
    `Resolve(c, X)` for a commit `c` inside the window is a member of `Revisions()`.
  - The test also has no merge, side-branch, or non-HEAD-ref case (F2).
- **SDD-INV-26** ("payload is a pure function of the manifest").
  - The test mutates the owner and expects a different digest. That is only the "manifest covers the
    profile" direction.
  - An implementation that reads an input outside the manifest passes, for example a main spec,
    another change's files, or the git author. The test needs a case that varies a non-manifest
    input and asserts that the revision is unchanged.
- **SDD-INV-27** checks four cases (active, archived-by, archived-earlier, unknown). It does not
  cover the cases where "archived by this commit" is ambiguous:
  - **A merge commit.** Under `--full-history`, 13 OpenSpec merges introduce an archive directory
    relative to their first parent. One is `98d90cc0`, which merges PR #61 after `ccaa5ad0` archived
    `add-view-dashboard-command` on the side branch. By §6 it "archives" that change.
  - **A root commit.** `FirstParent` returns `ok=false`, and the expected result is undefined.
  - **Dual presence**, which is F1's case.
- **SDD-INV-13's** "kit read only from the profile" still has no mechanical test form (round-2 F10).
- **Two titles changed meaning in place.**
  - SDD-INV-8 is still titled "the artifact digest is the exact committed bytes", but the digest is now
    a synthesized manifest digest.
  - SDD-INV-11 is still titled "resolution reads only the named tree", while its proving test now
    admits the first parent.

  §8 itself says new meaning takes a new number.

### r2-m0-review-round3-20261116-F5 (LOW): residual specification gaps and drift

**Broken property:** spec exactness (SDD-INV-5/7 conformance) and Law 2 drift.

**Evidence:**
- **The id functions live in a superseded document.** §4.3 and §5 define `slug` and `suffix` "as in
  draft 2 §4.1", a document marked superseded. That section still never gives the `kind` literal
  hashed into `suffix` (round-2 F9 residual). Two conforming implementations can therefore mint
  different record ids.
- **The empty name-part case was dropped.** Draft 2 made an empty name-part `ErrUnmappable`. Draft 3
  is silent, and `O-ADD-` matches `obligationRE` (`intent.go:25`). A name with no `[A-Za-z0-9]`
  bytes therefore yields a dangling-dash id.
- **The RENAMED rule does not match OpenSpec's format.** §4.2's RENAMED rule (name = FROM, statement =
  "the `TO` line" after "the requirement heading") presumes a heading. OpenSpec's RENAMED entries are
  `- FROM:` / `- TO:` line pairs with no heading (`src/core/parsers/requirement-blocks.ts:560`).
  No real delta has used RENAMED yet (0 seen), so the rule is untested against the format it targets.
- **The plan's M1 text is stale.** `sdd-warranty-connector-plan.md:109` still specifies "parse of
  committed artifacts at a tree" and "`Resolve(tree, token)`". The status note supersedes only the
  plan's "main-spec mapping, governance, and drift" sections, not M1's interface.
- **Intent SPEC §4 needs amending, unnamed.** Intent SPEC §4 defines `artifact_sha256` per provider
  (native self-digest, specdir exact file bytes) and says committed-file providers "resolve at the
  commit's own tree". Draft 3 introduces a manifest digest and a first-parent read, and does not name
  the §4 amendment that M1 needs alongside the §1 one.

## What holds

- **Content addressing works with the existing machinery.**
  - Projections key requirement and change-intent rows on `(source, id, artifact_sha256)`.
  - Readiness (`gates.go:285`) evaluates the scoped commit's exact targets against exact-revision
    verdicts and reviews.
  - A verdict bound to change-intent revision N remains valid for its commit when revision N+1
    exists. Revisions with empty `supersedes` pass `ValidateRevision` / `validateBatch` as specdir's
    do.
  - Every edit of a delta file requires a fresh `requirement-review.v1` on a new record id. That is
    the same per-revision cost the records provider already has, and it is workable. The extra
    re-reviews come from archive moves and profile edits (F3).
- **The id grammar fits.**
  - `BR-`/`CI-` + slug (2..40) + `-` + 6 matches `recordIDRE`.
  - `O-<SEC>-<name-part>` is at most 32 bytes after `O-` (3 + 1 + 28), within `obligationRE`.
  - There are no obligation-id collisions on the real history.
- **The payload is pinned (round-2 F6).** The profile is in both manifests, and change-intent targets
  derive only from in-directory files. One byte-level check: the change-intent payload's status and
  targets depend on the change directory path, and the manifest lines carry that path.
- **Records and decisions are well-formed.**
  - `CI-sdd-warranty-m0b-ti5ga9/0001.md` self-digest reproduces (`84786b71…`) and targets BR 0002
    at its real digest (`fbf813f6…`).
  - The VD is consistent with draft 3 and names its costs and revisit triggers.
- **Law 6:** spec, no production code. **Law 8:** no new dependency (git is already the git
  connector's tool).
- **The O-12 scope claim is observable** through SDD-INV-30.
- **Invariant numbering is append-only, with retirements kept as tombstones:** INV-1..14 titles are
  byte-identical to draft 2, and INV-15/16/17/19/21/22/23 are retired in place. F4 notes the two
  titles whose meaning drifted.
- **Most of §0 is honest.** Every count I measured reproduces exactly except name reuse. The
  squash-merge reading of the 16 never-active changes is supported.
- **History-native ingestion is the right shape.** On OpenSpec its cost is small (256-296 commits,
  about 650 distinct revisions, seconds), and it fully covers that repository's reachable states.

## Fitness for M1

**No.** F1 and F2 are defects in what M1 itself builds:

- **F1:** the reference repository wedges permanently under the specified fail-closed rules, and
  history-native ingestion turns every fail-closed rule into a permanent one.
- **F2:** the locked `Reader.Commits` walk is not the set of commits the git connector binds.

Both are contract-level. Code written to the contract as it stands would encode them.

Once the contract is revised, M1 should respect these conditions:

1. **Keep `sdd` out of the git connector.** M1 must not register `sdd` in the git connector's
   provider allowlist (`connector.go:198`) or in the CLI's `commitIntentResolver`. Anti-spoofing is
   M3's, and until then no tenant trailer may resolve through the provider. M3's independent check
   must read the profile's `kit` at the commit tree, because the change-intent suffix hashes it.
2. **Test the lifecycle on real histories.** The lifecycle tests should run over a fixture that
   includes:
   - a merge (including a TREESAME merge);
   - an unmerged branch claim;
   - the leftover-after-archive state from `7ced2a87`;
   - a root commit.

   Each probe count used as evidence should be printed by the probe that claims it.
3. **Amend intent SPEC §4 with §1.** The §4 amendment (manifest digest, first-parent read) lands in
   the same M1 commit as the §1 amendment.
4. **Keep O-11 out of the BR until it is satisfiable.** O-11 should not enter BR revision 0003 until
   it can be satisfied and evaluated from ledger data (F3).
