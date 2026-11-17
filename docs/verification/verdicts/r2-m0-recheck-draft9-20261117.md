---
schema: vera.verdict.v1
verdict_id: r2-m0-recheck-draft9-20261117
status: NEEDS_WORK
reviewed_commit: 991c9a9a0fac8719c02aae2e6db0c729448153ae
findings:
  - finding_id: r2-m0-recheck-draft9-20261117-F1
    severity: MED
  - finding_id: r2-m0-recheck-draft9-20261117-F2
    severity: LOW
  - finding_id: r2-m0-recheck-draft9-20261117-F3
    severity: LOW
  - finding_id: r2-m0-recheck-draft9-20261117-F4
    severity: LOW
  - finding_id: r2-m0-recheck-draft9-20261117-F5
    severity: LOW
  - finding_id: r2-m0-recheck-draft9-20261117-F6
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-recheck-draft9-20261117.md
artifact_sha: 650b226daad483d43c05b478d406383111c5afcb1b3cb0c75b97011e677a880f
---

# Release 2 M0 (SDD Warranty Connector): targeted non-author re-check of draft 9's changed sections

## Provenance

- **Under review:** HEAD `991c9a9a0fac8719c02aae2e6db0c729448153ae`.
  - `e333192` adds `docs/plans/R2-sdd-provider-spec-draft9.md` (supersedes draft 8). It also changes
    the `resolve-edges/07-edges` fixture (three plain files and the `rename-only` delta),
    `expected-resolve.json` (3 rows added, 21 in all), `expected-skips.json` (2 rows added),
    `testdata/README.md`, the probe's plain-file count, the plan's draft pointer and the intent SPEC §1
    pointer.
  - `991c9a9` is a one-line `@generated` pin refresh.
- **Scope, set by owner decision (2026-11-17):** a targeted re-check of what draft 9 changed: §0 (the
  round-8 table and the 1393 row), §2.1, §3, §4.1, §5, §6, §7, the §9 rows and statements, and §13.
  Unchanged sections were not re-reviewed from scratch. The owner decisions in force were taken as
  given, not re-litigated: M1 is OpenSpec only; profile paths are literal; `sync intent all` excludes
  `sdd`; a targeted re-check closes M0.
- **Inputs read in full:** `docs/verification/verdicts/r2-m0-review-round8-20261117.md`,
  `docs/decisions/VD-sdd-m1-openspec-first-2026-11-17.md`, `docs/plans/R2-sdd-provider-spec-draft9.md`,
  and `git diff --no-index` of drafts 8 and 9. I also read `fixture_test.go` and the gitcmd `Commits`
  `rev-list` call.
- **Out of scope:** the untracked `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`.
- **Evidence (all from my own commands; no repository file written except this verdict):**
  - `<scratch>/oracle/oracle.py` is my own oracle, written from draft 9's text. It replays every
    fixture the way `replayHistory` does: one commit per step, `PARENTS` / `REF`, fixed dates,
    `commit-tree`, `update-ref`. It lists commits with the gitcmd `rev-list --reverse --date-order
    --exclude=… --all` call. At every commit it then computes:
    - §2.1 strictness, including the exact pair and the direct-child archive;
    - the §3 state key and state-level causes;
    - §4.1 record units with the §4.2 parser;
    - §5 `changeOutcome` with rules 1-8 in order;
    - §6 `Resolve`, using parent archive listings;
    - the §7 Skips, de-duplicated by `(unit, cause, unit digest)` with the draft-9 digest.

    Output: `<scratch>/oracle-out.txt`.
  - `<scratch>/plain.txt`, `<scratch>/trees.txt` and `<scratch>/probe-out.txt` hold my scan and the
    probe run over the author's OpenSpec clone at `f085c850f9440f497ccc76ecbc2859c09e0e0ac7`.
  - `<scratch>/td-neutral/` is a copy of the testdata used for the neutral control.
- **HEAD at dispatch and at completion:** `991c9a9a0fac8719c02aae2e6db0c729448153ae`, unchanged.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral, as a targeted re-check. M1 itself (record identity, point-in-time
  resolution, a ledger-reaching source) should be reviewed at the foundational tier, as rounds 2-8
  recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m0-recheck9`, a **non-author**. It
  wrote none of the work under review and did not review rounds 4-8. This verdict counts under Law 9.
- **Redactions and tokens:**
  - No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-recheck9/` in the session scratchpad. It
    occurs 9 times in this verdict, including this line.
  - The OpenSpec clone is named only as "the author's clone".

## Round-8 findings re-verified

I checked against the text, the fixtures and real history, not against draft 9's §0 table.

- **F1 (rule 5 undefined for pre-parse skips and unpaired-only files): CLOSED.**
  - §4.1 now defines a record unit. A unit is a matched delta path that is either skipped before
    parsing (`non-regular-file`, `invalid-utf8`, `nul-byte`) or that parses to at least one entry or
    at least one unpaired RENAMED line.
  - §5 rule 5 reads "none of `n`'s delta paths is a record unit", and rule 6 covers a unit skipped
    "for any reason, including before parsing". Every case round 8 named now has one outcome:
    `cascade`.
  - SDD-INV-43's row now agrees with §4.2 on the unpaired-only file.
  - Pinned by `resolve-edges` `rename-only` in both pin files. SDD-INV-59 names the symlinked-delta
    case as an M1-synthesized row.
  - Residual: a **tree** at a delta path is not classified (F3 below). A non-delta non-regular entry
    inside a change directory has two readings (F1 below).
- **F2 (plain files under the root become change Skips; rule 4 active-only): CLOSED.**
  - §5 now says change entries are "a tree, symlink, or gitlink (never a regular file)". Regular files
    directly under either directory are "not changes, not units, and never a cause" (SDD-INV-66).
  - Rule 4 reads "a change entry of `n`, whether active or archived, is not a tree". SDD-INV-54's row
    adds the symlinked dated archive entry.
  - §0 has the 1393 row. I reproduced it independently:
    - **1393** of the 1624 walked commits carry `openspec/changes/IMPLEMENTATION_ORDER.md` as a `100644`
      blob directly under the root;
    - `skills-only-migration.napkin.md` is there at 4 commits;
    - the archive has no non-tree or undated entry at any commit;
    - the probe prints `… 1393 ['IMPLEMENTATION_ORDER.md', 'skills-only-migration.napkin.md']`.
  - `resolve-edges/07-edges` carries both shapes, and my oracle's draft-8 mutant (F2 control below) is
    caught.
  - Residual: §6's parent-archive check still reads a name-only `List` (F2 below).
- **F3 (name domain not applied in §5): CLOSED.**
  - `changeOutcome` runs only over in-domain names (§5).
  - `changes.archive` must be a direct child of `changes.root` (§2.1).
  - A dated tree, symlink or gitlink archive entry whose capture is `.`, `..` or the archive's own name
    is an `archive <entry>` unit with `archive-name-invalid` (§6).
  - `Resolve` and `Skipped()` therefore use the same name set, and my oracle finds no disagreement.
  - Residual: drift in §7's cause row and §2's unit comment (F6).
- **F4 (globs without §4.1's shape): CLOSED.**
  - §2.1 requires `requirements.paths` to be exactly `["<root>/*/specs/*/spec.md",
    "<archive>/*/specs/*/spec.md"]`, in that order.
  - §2.1's roots are now only `changes.root` and `changes.archive`, which removes the blob-root case.
- **F5 (literal-pathspec kill cannot fire): MOSTLY CLOSED.**
  - §3 states the flag is defence in depth with no observable effect.
  - SDD-INV-62's row drops the false `*x` claim and says the flag "is not claimed tested". Its kill is
    now the grammar's, which is reachable.
  - Residual: SDD-INV-62's *statement* still asserts "every git path lookup uses literal pathspecs"
    (F6).
- **F6 (INV-63 over-promises; digest undefined): MOSTLY CLOSED.**
  - INV-63 is retired with its reason and re-issued as INV-65 ("equal `changeOutcome` and equal
    Skips"). That follows §9's rule.
  - I checked INV-65(b) over every replayed commit of every fixture: equal keys always gave equal
    `changeOutcome` maps and equal Skip triples. `branches` 07 and 08 share a key under draft 9's key,
    and their `Resolve` differs, as the row says.
  - The digest is now defined per unit kind. Residual: the `profile` digest reads roots from a profile
    that may be invalid (F4), and INV-65's "caching `Resolve` by key" kill is not reachable from its
    own row (F5).
- **F7 (corners and drift): CLOSED.**
  - SDD-INV-59 names the four synthesized corners: absent profile, non-tree root or archive,
    symlinked change directory, symlink-only delta. Each outcome follows from §6 lines 2 and 4 and
    §5 rules 4 and 6.
  - SDD-INV-52 lists 15 causes. I counted §7's M1-reachable causes, and there are exactly these 15.
  - The README's `openspec 02` row now points at `resolve-edges` `draft-only`. The `speckit` rows read
    "M1: INV-64; M2: INV-10/42".

## Checks

**2. One outcome per (state, name), using my oracle (`<scratch>/oracle-out.txt`).**
- **`expected-resolve.json`:** all **21** rows match, 18 for `resolve-edges` and 3 for `branches`. The
  new rows:
  - `IMPLEMENTATION_ORDER.md` → `ErrUnknownChange`;
  - `README.md` → `ErrUnknownChange`;
  - `rename-only` → `ErrUnmappable:cascade`.
- **`expected-skips.json`:** the exact set matches for all 8 fixtures:

  | Fixture | Skips |
  |---|---|
  | `openspec-history` | 0 |
  | `speckit-history` | 1 |
  | `branches` | 1 |
  | `unmappable/duplicate-obligation` | 2 |
  | `unmappable/reused-name` | 1 |
  | `fenced` | 0 |
  | `head-on-leftover` | 1 |
  | `resolve-edges` | 8, including the new `requirement …/rename-only/specs/core/spec.md` / `unpaired-rename` and `change rename-only` / `cascade` at `07-edges` |

  Under the new unit digest:
  - `change twice` de-duplicates to step 01 (its archive entries and the profile are unchanged at 07);
  - `change again` keeps two Skips (04 and 05, different entry sets).
- **Agreement:**
  - at every replayed commit, for every in-domain name with a change entry, `Resolve`'s
    `ErrUnmappable` cause equals `changeOutcome`'s;
  - that commit's Skip set carries the matching `change <n>` triple.
- **Corners from the text (not fixture-pinned):**
  - **Record units:** decided as F1 above, except the tree-at-delta-path shape (F3).
  - **Tree, symlink or gitlink change entries:** rule 4 applies to active and archived entries
    alike. A dated symlink archive entry gives `non-regular-file` in both `Skipped()` and `Resolve`
    (§6 line 4 precedes line 5).
  - **Out-of-domain archive captures:** an `archive <entry>` Skip. `Resolve` gives `ErrUnknownChange`
    for both the capture and the entry name, so they agree.
  - **Absent profile:** §6 line 2 now gives `ErrUnknownChange` explicitly, with no Skip.
  - **Rule 4 on archive entries:** a regular file with a date prefix in the archive is ignored by §5,
    but not by §6 line 5 (F2).
- **Oracle calibration:**

  | Control | Mutation | Expected | Result |
  |---|---|---|---|
  | Positive | regular files counted as change entries (draft 8) | must fail | fails: an extra `change IMPLEMENTATION_ORDER.md` / `non-regular-file` Skip, and its `Resolve` row is wrong |
  | Positive | an unpaired-only file treated as zero entries | must fail | fails: 2 Skips missing; `rename-only` → `ErrNoChangeIntent` |
  | Neutral | reword `IMPLEMENTATION_ORDER.md` | must survive | 0 mismatches |
  | Non-discriminating (F4) | `profile` digest = the profile entry only (draft 8) | — | 0 mismatches: no pin depends on the new `profile` digest |

**3. The exact-pair rule against every committed profile.**
- 39 profile files: 35 OpenSpec profiles are valid and exactly the standard pair, with the archive a
  direct child of the root.
- `resolve-edges/06` is the standard pair, invalid only for its `unknown` field, as intended.
- The 3 `speckit` profiles are invalid by `kit` (and also by the pair and the missing root). That is
  intended (SDD-INV-64).
- No profile became `profile-invalid` unintentionally.

**4. §9 discipline.**
- Statements 43, 52, 54, 59 and 62 are unchanged. Only their rows changed, and each change narrows or
  strengthens. INV-28's and INV-43's statements remain true under record units, because an
  unpaired-only file yields no *emitted* record.
- INV-63 has a tombstone with its reason. INV-65 and 66 are appended, and no number is reused.
- **Named kills, all reachable:**
  - INV-43: "unpaired-only as zero entries" (my control);
  - INV-54: "an active-only rule 4", via the synthesized symlinked dated archive entry;
  - INV-59: "`ErrNoChangeIntent` before `cascade`" and "plain files as changes" (my controls);
  - INV-62: "a lenient grammar", for example `openspec/changes/*/spec.md`;
  - INV-65(a): "an oid-only key";
  - INV-66: "any object type" (my control).

  The exception is INV-65's "caching `Resolve` by key" (F5).

**5. Tests and probe.**
- Under Go 1.27.4, `go test -count=1 -v ./internal/connector/intent/sdd/` passes all three tests.
- My own plain-file scan reproduces 1393 / 1624 (above).
- `scripts/p6-artifact-integrity.sh --check` exited 0 at HEAD, before this verdict existed.

## Findings

### r2-m0-recheck-draft9-20261117-F1 (MED): a symlink or gitlink inside a change directory, other than a delta file, has two contract outcomes

**Broken property:**
- SDD-INV-59 ("exactly one" outcome per `(c, n)`).
- SDD-INV-54 against SDD-INV-61.
- SDD-INV-34 (the manifest covers exact committed bytes).

**Evidence (text only):**
- **The case.** Take an active change `x` whose directory is a tree, with one clean delta file and a
  symlink `x/design.md`, or a gitlink, or a symlinked `x/specs` or `x/specs/<cap>`.
- **What §5 decides.** Rule 4 tests only "a change entry of `n`", and `x` itself is a tree. So rule 4
  does not fire, and rules 5-8 give **a change intent**. For a symlinked `specs` directory, no delta
  path matches, so the outcome is **`ErrNoChangeIntent`** with no Skip.
- **The texts that say otherwise:**
  - SDD-INV-54's statement: "a symlink or gitlink a unit reads makes that unit unmappable".
  - §7's `non-regular-file` row: "where a unit reads a file or directory".
  - §5's change manifest "covers every file in it", so the change unit does read `design.md`. These
    texts give `ErrUnmappable:non-regular-file`.
- **The texts that agree with §5:**
  - SDD-INV-61: "files a unit neither parses nor carries are never a cause";
  - §7's "Other files … are never a cause".
- **The manifest cannot be computed.** §2's `Artifact` gives symlinks and gitlinks "no Bytes". The
  manifest line needs `<sha256-hex-of-exact-bytes>`, so the change intent §5 selects has no
  computable digest from the `Reader`.
- **Two conforming M1s therefore differ** on `Resolve(c, x)`, on `Skipped()`, and on the emitted
  revision's digest.
- **Scope.**
  - The real corpus has no symlink or gitlink anywhere under `openspec/changes` at any of 337 distinct
    trees (`<scratch>/trees.txt`).
  - The gap predates draft 9: draft 8's rule 4 was narrower still.
  - It is the same class as round 8 F1 (a byte-level case with two outcomes), and draft 9 changed
    rule 4 without covering it.
  - No fixture or SDD-INV-59 synthesized row pins it.

### r2-m0-recheck-draft9-20261117-F2 (LOW): §6's parent-archive check reads names only, so a regular file in a parent's archive can still "archive" a change that §5 and SDD-INV-66 say it cannot

**Broken property:** SDD-INV-66 ("never a change, a unit, or a cause"), SDD-INV-32 and SDD-INV-33.

**Evidence:**
- **The interface.** `List(ctx, commit, dir) []string` returns "Names of the entries directly under
  dir" (§2), with no type.
- **The rule.** §6 line 5 fires when "some parent of `c` has an archive entry naming `n`".
- **The case.** A parent carries a regular file `archive/2026-01-01-x`. That file is not a change
  entry (§5, SDD-INV-66). A child then archives `x` as a tree.
- **The two outcomes:**
  - an implementation using `List` as specified gives `ErrArchivedEarlier`;
  - one applying §5's type rule to the parent gives the revision.
- **Scope.** Contrived, and not in real history: no non-tree or undated archive entry exists. Draft 9's
  type restriction was not carried into §6 line 5 or `List`.

### r2-m0-recheck-draft9-20261117-F3 (LOW): §4.1 does not classify a tree at a delta path

**Broken property:** SDD-INV-59 (rule 5 against rule 6).

**Evidence:**
- **The case.** A tree at `<root>/x/specs/core/spec.md` textually matches the exact glob.
- **Why §4.1 cannot decide it.** §2's `Read` returns "every tree entry", which may or may not list
  trees. §4.1's pre-parse causes do not include a tree, because §7's `non-regular-file` is a symlink or
  gitlink. A tree cannot be parsed either.
- **The two outcomes.** If it is ignored, `x` resolves to `ErrNoChangeIntent`. If it counts as a
  non-regular delta, `x` resolves to `ErrUnmappable:cascade`.
- **Scope.** The real corpus has no tree at any `specs/*/spec.md` path.

### r2-m0-recheck-draft9-20261117-F4 (LOW): the `profile` unit digest reads roots from a profile that may be invalid

**Broken property:** §7's de-duplication rule, and SDD-INV-46's exactness.

**Evidence:**
- **The rule.** §7's `profile` digest covers "every SDD root's and the archive's own entry". For a
  `profile-invalid` state, the roots come from the invalid profile.
- **Unparseable profile.** For invalid JSON, or the speckit profile, which has no `changes.root`, the
  roots are undefined.
- **Parseable profile.** For a profile invalid only by an extra field (`resolve-edges/06`), the roots
  parse. The digest then changes with the root tree oid, so a long-lived invalid profile yields one
  Skip per distinct `changes` tree rather than one Skip. A persistent blob archive does the same.
  That works against §7's stated purpose ("one `Skip`, not one per state").
- **The pins cannot see it.** My draft-8-digest control gives 0 mismatches: `speckit-history` has no
  root, and `06` is a single state.

### r2-m0-recheck-draft9-20261117-F5 (LOW): SDD-INV-65's "caching `Resolve` by key" kill cannot fire from its own row

**Broken property:** §9 ("each proving test names the wrong implementation it kills"), and Law 9's
calibration corollary.

**Evidence:**
- **What the row asserts.** Part (a) asserts keys and mapping only. Part (b) asserts `changeOutcome`
  and Skips, and says `Resolve` "may differ".
- **Why the kill cannot fire.** A mutant that caches `Resolve` by key leaves every assertion of (a)
  and (b) unchanged. It is killed only by SDD-INV-59's `branches` `08-merge-side` row.
- **What I confirmed.** The `branches` 07 and 08 keys are equal under draft 9's key.

### r2-m0-recheck-draft9-20261117-F6 (LOW): residual drift in sections draft 9 changed

- **§7's `archive-name-invalid` row** still reads "an archive entry without a date prefix". Two of §6's
  draft-9 rules are not reflected in it:
  - it does not cover a dated entry whose capture is out of domain;
  - it does not exclude regular files.
- **§2's `Skip.Unit` comment** still says "`archive <entry>` (an undated archive entry, §6)".
- **SDD-INV-62's statement** still asserts "every git path lookup uses literal pathspecs". §3 says
  that this property is not separately testable and that "no invariant claims it is", and INV-62's row
  disclaims it.
- **`fixture_test.go`** says "draft 8 §7's closed list" in a failure message. This is cosmetic, and the
  file was not changed by `e333192`.

## What holds

- Every round-8 MED is closed as round 8's conditions 1 and 2 asked:
  - a record-unit predicate that covers pre-parse and unpaired-only files;
  - change-entry types that agree with §7 and SDD-INV-61 for regular files;
  - rule 4 applied to archive entries;
  - the real `IMPLEMENTATION_ORDER.md` shape in a fixture and in §0.
- The exact-pair and direct-child rules remove round 8 F3 and F4 by construction. Every committed
  OpenSpec profile still conforms.
- Both pin files are exactly right under draft 9's text, by an oracle with calibrated positive and
  neutral controls. `Resolve` and `Skipped()` agree on every replayed commit, and INV-65(b) holds on
  every fixture.
- §9's retire-and-reissue rule is followed (INV-63 → INV-65; 66 appended).
- Law 6: no provider code exists. Law 8: no new dependency. Law 1: generated pins unchanged, and
  `991c9a9` is a marker-carrying refresh. Law 9: this re-check is non-author.

## Fitness for M1

**Fit for M1 code under one condition.** No HIGH finding. The one MED (F1) is narrow: one sentence in §5
or §7 and a synthesized SDD-INV-54/59 row close it. It is not reachable on the reference corpus. It can
be fixed in the `sdd/SPEC.md` that M1 commits, verified by M1's foundational-tier non-author review,
without another M0 paper round. That is consistent with the owner's decision.

Conditions:
1. **F1:** decide in the text, once, whether a symlink or gitlink inside a change directory (a
   non-delta file, or a `specs/` or capability directory) makes the change `non-regular-file` or is
   ignored. Reconcile SDD-INV-54 with SDD-INV-61 and §7's "Other files". Say what the manifest records
   for such an entry. Pin one case in SDD-INV-59's synthesized rows.
2. Land F2-F6 in the same `sdd/SPEC.md`:
   - **F2:** `List`, or §6 line 5, applies §5's entry types.
   - **F3:** classify a tree at a delta path.
   - **F4:** define the `profile` digest for an invalid profile, or restrict it to the profile entry
     for `profile-invalid`.
   - **F5:** add a `Resolve` assertion to INV-65's row, or drop the kill.
   - **F6:** fix the drift.
3. M1 is reviewed at the foundational tier, with the mutation calibration §13 already requires.
