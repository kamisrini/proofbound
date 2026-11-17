---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round7-20261117
status: NEEDS_WORK
reviewed_commit: fc9be6a85deaf6a1b55b28d62b7c7cb344049bf0
findings:
  - finding_id: r2-m0-review-round7-20261117-F1
    severity: MED
  - finding_id: r2-m0-review-round7-20261117-F2
    severity: MED
  - finding_id: r2-m0-review-round7-20261117-F3
    severity: MED
  - finding_id: r2-m0-review-round7-20261117-F4
    severity: LOW
  - finding_id: r2-m0-review-round7-20261117-F5
    severity: LOW
  - finding_id: r2-m0-review-round7-20261117-F6
    severity: LOW
  - finding_id: r2-m0-review-round7-20261117-F7
    severity: LOW
  - finding_id: r2-m0-review-round7-20261117-F8
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round7-20261117.md
artifact_sha: feb71d974c51d28dc7f60b6751d867320f648edaae12cc8bf7e8838e21a4b9dc
---

# Release 2 M0 (SDD Warranty Connector): draft-7 total-Resolve, provider-owned-NUL provider contract. Review round 7

## Provenance

- **Under review:** HEAD `fc9be6a85deaf6a1b55b28d62b7c7cb344049bf0`.
  - `ab7c4be` introduced `docs/plans/R2-sdd-provider-spec-draft7.md`, which supersedes draft 6, with
    the re-pinned `derive/history_probe.py` (naive-toggle mode, `--pins`), the `generated_by` field in
    `expected-obligations.json`, the README count fixes, the intent SPEC §1 pointer, and the plan status
    line. `scripts/intent-applicability.sh` reports it `applicable=false`.
  - `a1a104e` changed `kernel/internal/connector/intent/sdd/fixture_test.go` (the skip-cause check),
    under `Intent: CI-sdd-warranty-m0b-ti5ga9` with the owner's `Authorized-By`. It is
    `applicable=true`.
  - `fc9be6a` is a one-line `@generated` pin refresh in `docs/verification/p6-connector-reality.md`.
    `scripts/p6-task2-close.sh --check` exits 0 at HEAD, so it is script-written and current.
- **Out of scope:** `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`. Both are
  untracked and unchanged.
- **Follows up:** `docs/verification/verdicts/r2-m0-review-round6-20261117.md` (NEEDS_WORK, F1..F7),
  read in full; rounds 4 and 5 skimmed for their conditions.
- **Owner decisions taken as given, not re-litigated:** `sync intent all` excludes `sdd`, and the
  helper is committed. I judged only whether the contract implements them consistently.
- **Judged against:** the CLAUDE.md Build Laws and this ground truth at HEAD:
  - **`intent.go`:** `New` (`:134`), `Sync` (`:145-212`), `ValidateRevision` (`:214-278`), `strict`,
    `validateObligations`, `validateReferences`, `validateBatch` (`:361`) and `orderBatch`.
  - **`core/core.go`:** `MaxPayloadBytes` (`:64`), `NewEvent` (`:288`), `Event.Validate` (`:302-333`);
    `core/SPEC.md:75-83` and `:249`.
  - **Every ledger insert path:** `store.go` `Sync.Append` (`:256` validates, `:261` inserts),
    `ImportReplayRecords` (`:480` validates, `:503` inserts), `importRecords` (`:538` validates, `:563`
    inserts); `migration.go` `validateExpected` (`:139`); `twin/replay.go:340` and `:381`;
    `store/migrations/001_ledger.sql:9` (`payload JSON`). A repository-wide grep for `INTO events`
    finds no other insert.
  - **`projections/projections.go`:** `eventRoute` (`:296`) and `reduceIntent` (`:302-370`).
  - **`cli/cli.go`:** `:37`, `:57-60`, `:88-93`, `:513-528`, `:596`, `:835`, `:854`.
  - **`connector/git/gitcmd/gitcmd.go`:** `run` (`:401-416`), `gitEnvironment`, `gitCommandTimeout`.
  - **Hooks:** `.claude/hooks/block-generated-edit.sh`.
- **Real-history evidence:**
  - **Repository.** The author's clone of `Fission-AI/OpenSpec` at
    `f085c850f9440f497ccc76ecbc2859c09e0e0ac7`, branch `main`.
  - **OpenSpec's reader source.** `src/core/parsers/requirement-blocks.ts` `parseRemovedNames`
    (`:535-557`) and `parseRequirementBlocksFromSection` (`:484-526`), read against §4.2. No Node.js
    runtime is installed on this host, so I did **not** re-execute OpenSpec's reader. Round 6 executed
    it over the same five fenced fixtures and the same 401 blobs, and none of those inputs changed
    (`git diff 05b8e7d fc9be6a -- kernel/` touches only the probe, the test, the README, the intent
    SPEC and the pins file's two metadata lines).
  - **The author's probe**, copied to `<scratch>/probe.py` and run over the clone. Its output is in
    `<scratch>/probe-out.txt`.
  - **My own recounts:** `<scratch>/dual.txt` (the 23 dual-presence commits by my own
    `ls-tree` loop), `<scratch>/fp.txt` and `<scratch>/fpdual.txt` (the first-parent walk of `main`).
- **Synthetic evidence:**
  - `<scratch>/td/`: a copy of `testdata/` for the `--pins` regeneration.
  - `<scratch>/cal-*`: throwaway module copies of the `sdd` package for calibration.
  - `<scratch>/ps/`: a three-file git repository for `git ls-tree` path semantics, under git 2.34.1.

  None of these wrote a repository file.
- **HEAD at dispatch and at completion:** `fc9be6a85deaf6a1b55b28d62b7c7cb344049bf0`, unchanged. The
  working tree held only the two out-of-scope untracked paths.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. M1 owns record identity, point-in-time resolution and a new
  ledger-reaching source, so M1 itself should be reviewed at the foundational tier, as rounds 2-6
  recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m0-r7`, a **non-author**. It wrote
  none of the work under review, and it is not the reviewer of rounds 4, 5 or 6. This verdict counts
  under Law 9.
- **Writes:** no repository file other than this verdict. No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m0r7/` in the session scratchpad. It occurs
    16 times in this verdict, including this line.
  - The OpenSpec clone is named only as "the author's clone".

## Round-6 findings re-verified

I checked against the text and the code, not against draft 7's §0 mapping table.

- **F1 (Resolve has no outcome for a change without an intent): PARTIALLY FIXED.**
  - Fixed: `ErrNoChangeIntent` gives proposal-only changes and zero-entry archives an outcome
    (`02-draft-proposal-only` holds only `proposal.md`, so SDD-INV-59's row is buildable). The name
    domain is defined, and `.`, `..`, `archive` and `archive/2026-01-02-add-login` all fall to
    `ErrUnknownChange` by the text.
  - Not fixed: the five outcomes are still neither exhaustive nor exclusive (new F1), and Spec Kit has
    no resolution rule at all (new F2).
- **F2 (U+0000 in `ValidateRevision` is retroactive): FIXED.** Verified in code, below under "§10 I".
  The `nul-byte` rule now lives only in the provider, and `ValidateRevision` gains only the size
  bound, which no ledgered event can fail. INT-INV-4 gains the rebuild-replay row that round 6
  condition 2 asked for.
- **F3 (state-level object types; encoding scopes): PARTIALLY FIXED.**
  - Fixed: §3 places the profile path, every root and the archive as absent, tree, blob, symlink or
    gitlink, and SDD-INV-61 tests it. §7 now gives `invalid-utf8` and `nul-byte` the same two scopes.
  - Not fixed: §3 and SDD-INV-55 still apply `invalid-utf8` to any non-UTF-8 path inside a unit, which
    §7 and SDD-INV-61 now forbid (new F4). The state key omits the profile's entry mode (new F5). The
    `ls-tree <c> -- <path>` semantics expose tenant profile strings to git's pathspec parser (new F3).
- **F4 (case mapping; undated archives; INV-53 row): MOSTLY FIXED.**
  - The byte-wise ASCII rule for `np` and `slug` is unambiguous, and the `expoſe` → `EXPO-E` example
    pins the reading in prose.
  - Undated archive entries are `archive-name-invalid` and name no change.
  - INV-53's row now lists only reachable routes. `statement-blank` is Spec Kit-only, which matches
    §4.2: a chosen OpenSpec statement is never blank, a surviving name contains `[A-Za-z0-9]`, and a
    `TO` line contains `TO:`.
  - Residual: round 6 condition 4 asked for a **non-ASCII id vector**. None exists in `testdata/`, and
    SDD-INV-7's row still kills only "a suffix omitting `kind`" (new F6).
- **F5 (§12 omissions): FIXED.**
  - `core/SPEC.md:75-83` is listed. Line 82 is `SourceIntentSpecdir`, which matches the grep.
  - The skip print site (`cli.go:~528`; the `Fprintf` is at `:527`) and SDD-INV-60 are added.
  - The `delivery-enforce` description is now correct: `scripts/delivery-enforce.sh:38` runs
    `sync all`.
  - The user guide is assigned to R2's release documentation, which is a defensible boundary.
  - Re-running §12's grep at HEAD finds the same registration sites and no new one.
- **F6 (pins not regenerable; self-test gaps): MOSTLY FIXED.**
  - The naive toggle is defined in §13, and the probe implements that definition literally.
  - `--pins` regenerates `expected-obligations.json` **byte-identically** (SHA-256 `cd9319b3…` for
    both). The `naive_toggle` and `unmasked` lists therefore have a committed derivation now.
  - The self-test rejects causes outside §7's list, and its `skipCauses` set equals §7's 18 causes
    exactly.
  - SDD-INV-39's row now re-derives all three lists, so a stale `naive_toggle` or `unmasked` pin is an
    M1 kill.
  - Residuals: `--pins` only runs inside some git work tree, and the file carries no Law 1 marker
    (new F7).
- **F7 (drift): MOSTLY FIXED.**
  - Fixed: the README counts (38 of 135; 83 / 17 and 39 / 9; 12 / 108), the probe docstring, the
    "13 / 9" label in §0, and §4.2's REMOVED-bullet rule (which now matches the probe).
  - Residual: §10 D still says "`main` itself for 9 commits", and several "draft 6" labels remain
    (new F8).

**Round-6 "Fitness for M1" conditions:**

1. **Every `Resolve` input has a defined outcome, tested by SDD-INV-59: NOT MET.** The proposal-only
   and name-domain cases are met. Archived-twice, profile-invalid and non-tree states have no outcome,
   several states have two, and Spec Kit has no rule (F1, F2).
2. **The U+0000 rule versus already-ledgered events, with a rebuild test: MET.**
3. **Place every state-level read; one scope for the encoding causes: PARTIALLY MET.** The reads are
   placed, but the scope still contradicts itself across §3, §7, INV-55 and INV-61 (F4). The
   state-level lookup brings a new tenant-caused reader error (F3), and the profile's mode is missing
   from the key (F5).
4. **Fix the case-mapping reading and pin it with a non-ASCII id vector; define undated archives:
   PARTIALLY MET.** The reading is fixed and undated archives are defined. No vector is pinned (F6).
5. **Complete §12, make the pins derivable, fix the drift: MOSTLY MET** (F7, F8).

## Evidence reproduced

All of these come from my own commands.

**§0 counts.** The probe's output (`<scratch>/probe-out.txt`) matches every §0 row:
- 1624 / 120 / 1 commits, merges and roots;
- 338 distinct states;
- 256 / 451 commits touching `openspec/changes`;
- 135 / 108 changes observed active / archived;
- 12 / 108 archived without ever being observed active;
- 38 changes with zero entries at some state;
- 83 (17) removals and 39 (9) statement changes;
- 4 requirements with no statement line;
- 0 RENAMED `FROM` entries;
- 401 / 1 delta blobs (`2025-09-29-improve-init-onboarding/specs/cli-init`);
- 23 dual-presence commits for 1 name, 4 remote refs, 0 names archived twice.

Independent checks:
- **Dual-presence set.** My own `cat-file -e` + archive `ls-tree` loop over `rev-list --all` finds the
  same **23** commits (`<scratch>/dual.txt`). The leftover active directory holds one file,
  `specs/cli-list/spec.md`, with **one** tree id across all 23 commits. So the real leftover has
  entries, and the zero-entry name-reuse overlap in F1(c) is not exercised by real history.
- **"13 / 9".** The first-parent walk of `main` contains **13** consecutive dual-presence commits, at
  positions 826-838 (`332816cc` … `acd1ca28`). By their parent lists, **9** are non-merge and 4 are
  merges. §0 is correct. The probe does not print this row (F8).

**Fixtures.** Under Go 1.27.4, `go test -count=1 -v ./internal/connector/intent/sdd/` passes all three
tests at HEAD: `TestFixturesReplayAsDocumented`, `TestBranchFixtureReplaysAsDocumented` and
`TestPinnedExpectationsMatchFixtures`.

**Pin regeneration.** I ran `python3 -I <scratch>/probe.py --pins <scratch>/td` on a copy of `testdata/`:
- **On the plain copy:** it **fails**. The module-level `git rev-list … --all` in that directory raises
  `CalledProcessError` (exit 128, "not a git repository") before `--pins` is examined (F7).
- **After `git init` of the copy:** it exits 0, and the regenerated `expected-obligations.json` is
  **byte-identical** to the committed one.

**Calibration**, on throwaway module copies of the `sdd` package (`<scratch>/cal-*`), and for the
pins on copies run through `--pins`:

| Control | Mutation | Expected | Result |
|---|---|---|---|
| Baseline | none | pass | passes |
| Positive 1 | `obligation-collision` → `bogus-cause` in `expected-skips.json` | fail | fails: `has a cause outside draft 7 §7's closed list` |
| Positive 2 | the `branches` cause `name-reuse` → `Name-Reuse` | fail | fails, with the same message |
| Positive 3 (pins) | the probe's fence closer ignores length (`len(…) >= active[1]` → `True`) | `--pins` differs | differs: `fence-length`'s `openspec_mask` becomes `[Alpha, Hidden]` instead of `[Alpha, Beta]` |
| Neutral 1 | reword Alpha's statement in `fence-length` | pass; `--pins` identical | passes; identical |
| Neutral 2 (pins) | edit the probe's docstring | `--pins` identical | identical |
| Stated gap | `obligation-collision` → `unpaired-rename` (wrong but listed) | — | passes. Exact causes are M1's SDD-INV-46, as §13 says |
| Stated gap | delete `fence-length`'s 4-backtick lines | — | self-test **passes**, and `--pins` **differs** (OpenSpec now yields `[Alpha]`). Nothing in `make check` regenerates pins before M1 (F7) |

**§10 I: is the size bound retroactively safe? Yes.**
- Every path that writes `events` validates first:
  - `Sync.Append` (`store.go:256`);
  - `ImportReplayRecords` (`:480`);
  - `importRecords` (`:538`), whose historical-evidence caller also runs `migration.go:139`.

  The twin replay validates too (`replay.go:340`, `:381`). No other `INSERT INTO events` exists in Go,
  SQL or shell.
- `Event.Validate` rejects `len(Payload) > MaxPayloadBytes` (`core.go:322`). `MaxPayloadBytes` has been
  `1 << 20` since `33d46d4` introduced it, so no earlier, larger bound ever admitted bigger events.
- `payload` is a `JSON` column (`001_ledger.sql:9`), which stores the input text verbatim. A rebuild
  therefore reads back the exact bytes that were validated.
- `reduceIntent` passes `r.Event.Payload` as the revision payload (`projections.go:303`).
  `ValidateRevision` already requires `Canonicalize(payload) == payload`, so the bound is applied to
  the same bytes in `Sync` and in a rebuild.
- **Conclusion:** no ledgered event can exceed the new bound. The rule also closes the partial-append
  path: today an oversize item fails at `NewEvent` inside the append loop (`intent.go:191-195`), after
  earlier items were appended. With the bound, it fails in `validateBatch` (`:366`) before any append.
- **Other M1 changes are additive for history:** admitting `intent.sdd` in `ValidateRevision`,
  `eventRoute` and `gates.go:304`. No `intent.sdd` event can exist before M1. No other M1 change
  tightens a check that rebuild re-runs.

**§4.2 against OpenSpec and the probe.**
- `parseRemovedNames` collects header and bullet names in document order, and has no notion of a
  block. "A REMOVED bullet ends any open block" therefore affects only Proofbound's own statement
  rule, and the probe's `close(); entry = None` implements it.
- The probe's `WS`, `DOT`, `JS_WS` and `GO_SPACE` sets are, by code point, exactly §4.2's `WS`, `DOT`,
  JavaScript `trim`, and Go `unicode.IsSpace`.
- The statement rule matches §4.2: the first unmasked, non-blank line of the block that comes before
  any `#` line; RENAMED takes the `TO` line's bytes. A `##` line ends statement selection, and the
  next header or bullet closes the entry. The output is the same as ending the block at that line.
- **Conclusion:** the probe now implements §4.2 for names, sections, masks and statements. The pins
  carry only `(section, name)`, so statements are not pinned. That is consistent with §13.

**`git ls-tree <c> -- <path>` semantics**, under git 2.34.1 through `run`'s flags, in `<scratch>/ps/`:

| Path argument | Result |
|---|---|
| `openspec/changes`, `./openspec/changes`, `openspec//changes`, `openspec/./changes` | the tree entry itself (rc 0) |
| `openspec/changes/` (trailing slash) | the **children** of the tree, not the entry (rc 0) |
| `:(top)openspec/changes` | the entry `openspec/changes`, although no path with that literal name exists (rc 0) |
| `:(glob)openspec/*`, `:!x1`, `:(attr:foo)x1` | **fatal**, "pathspec magic not supported by this command" (rc 128) |
| `../x`, `/x` | **fatal**, "outside repository" (rc 128) |
| `` (empty) | **fatal**, "empty string is not a valid pathspec" (rc 128) |
| `.` | every top-level entry (rc 0) |
| `x[1]`, `x?`, `open*` | literal match only (`x[1]` found; the others list nothing) |

`run` sets neither `GIT_LITERAL_PATHSPECS` nor `--literal-pathspecs`, and `gitEnvironment` strips every
`GIT_*` variable.

**Law 1 hook.** `block-generated-edit.sh` exits **0** for `expected-obligations.json`, so the file is
hand-editable. It exits **2** for the control `docs/verification/p6-event-route-matrix.md`.

## Findings

### r2-m0-review-round7-20261117-F1 (MED): §6's outcomes are still not a partition of tenant states, so `Resolve` is not total, and some states have two outcomes

**Broken property:**
- SDD-INV-59 ("Resolve is total over tenant states … exactly one of").
- §6's "An error that is none of these comes **only** from the `Reader` … No tenant state reaches it".
- §14 M3 entry condition 3, which relies on that split to report a bad claim per commit instead of
  failing the sync.
- §0's row "Names archived twice at one commit: … Still unmappable if it occurs (§5)".

**Evidence.** All of this follows from the text of §5 and §6.

**(a) A name archived twice by `c` has no outcome.** Take two archive entries naming `n` at `c`, no
parent with either, and no active `n`.
- Rule 1 fails, because nothing is active.
- Rule 2 fails, because it needs "**exactly one** archive entry".
- So the revision, `ErrNoChangeIntent` and `ErrUnmappable` outcomes are excluded, because all three
  require "rule 1 or 2 holds".
- `ErrArchivedEarlier` needs an entry "at some parent", and there is none.
- `ErrUnknownChange` needs "no entry names it", but two do.

By §6's own last paragraph, the result is a reader-class error. §0 and §5 say this state is
`name-reuse`, an `ErrUnmappable`. It is reachable by tenant bytes:
- a root commit that imports an existing OpenSpec tree in which a name was archived on two dates (§6
  itself says "a root commit archives what it carries");
- any single commit that adds two dated copies.

**(b) The commit that archives a reused name is classed "archived earlier".** Change `n` is archived
(`2026-01-05-n`), and a new `n` is later proposed and archived (`2026-03-01-n`) at `c`.
- At `c`, two entries name `n`, and the parent has `2026-01-05-n`.
- Rule 2 fails, and `ErrArchivedEarlier` holds, although `c` is exactly the commit that archived the
  second `n`.
- Meanwhile `Revisions` skips `change n` at that state as `name-reuse` (§5, "archived twice").
- So the `Resolve` sentinel and the `Skipped()` cause disagree for the same `(c, n)`.

**(c) Two outcomes hold at once.** `ErrNoChangeIntent` ("rule 1 or 2 holds, but `n` has no requirement
records") and `ErrUnmappable` ("rule 1 or 2 holds and `n` is unmappable at `c` (a §7 cause, including
`name-reuse`)") overlap, and no precedence is stated:
- a zero-entry leftover directory beside its archive entry, which is `name-reuse` with no records;
- a proposal-only change named `a` (`slug-too-short`) or with a non-UTF-8 name (`invalid-utf8`);
- a change whose only requirement record is skipped. Does the skipped record count as "has
  requirement records"? `expected-skips.json` pins `unmappable/duplicate-obligation`'s `change dup`
  as `cascade`, which reads "yes", but §6's examples ("a proposal-only change, or a zero-entry
  archive") read "records emitted".

The same ambiguity decides whether such a change is a `Skip` unit at all. §7's units are "a
requirement record, a change intent at a state, or a whole state", and a change with no records has
no change intent. That makes SDD-INV-46's exact skip set implementation-defined.

**(d) States whose roots come from a profile that is absent, invalid, or names a non-tree.**
- Rules 1 and 2 read `<root>` and `changes.archive` from the profile at `c`.
- At a `profile-invalid` state there is no root to read. At a state whose root or archive is a
  symlink, blob or gitlink, §3 makes "the whole state unmappable".
- §6 does not say whether `Resolve(c, n)` is then `ErrUnmappable` (the state is skipped) or
  `ErrUnknownChange` ("no entry names it"). The natural literal reading is `ErrUnknownChange`, which
  tells a claimant their change does not exist when the state is in fact skipped.
- The same holds for a change directory that is itself a symlink or gitlink. SDD-INV-54 skips that
  unit as `non-regular-file`, but rule 1's "`<root>/<n>/` exists" is false for a non-tree, so `Resolve`
  says `ErrUnknownChange`.

**(e) The proving row cannot see any of this.** SDD-INV-59's row runs over `openspec-history` and
`branches` only. Neither holds an archived-twice name, a reused-then-rearchived name, a
`profile-invalid` state, a non-tree root, or a zero-entry leftover. The test can pass while (a)-(d)
stand.

**Why it matters:**
- (a) is round 6's F1 harm again, in a smaller corner. M1 written to the letter returns a reader-class
  error for a tenant state. That is the one class M3 is told to treat as fatal.
- (b)-(d) make the sentinel, and so the per-commit report M3 builds on it, depend on which reading an
  implementation picks.

### r2-m0-review-round7-20261117-F2 (MED): Spec Kit has no resolution rule and no change-intent id

**Broken property:**
- SDD-INV-59 (totality "over tenant states", which includes Spec Kit tenants).
- SDD-INV-7 ("Record ids are deterministic"), for Spec Kit change intents.
- SDD-INV-13 (two kits on one provider).
- §1's scope: "R2 warrants OpenSpec changes **and Spec Kit features**".

**Evidence:**
- **§6 is written only for OpenSpec.** Rules 1 and 2, the archive-name rule and the name domain are
  stated entirely in terms of `changes.root` and `changes.archive`. The Spec Kit profile (§2.1) has
  neither: its `changes` object is `{mode, sibling, status}`. For a `speckit` state, rule 1's
  `<root>/<n>/` is undefined, so every `Resolve(c, n)` is undefined.
- **The change-intent id is OpenSpec-only too.** §5 gives the id formula
  `CI-<slug(name)>-<suffix("change-intent", name)>` under the "**OpenSpec:**" bullet only. The
  "**Spec Kit:**" bullet says "one change intent per feature directory" and gives no id.
- **The feature key can collide.** §2.1 says the feature key is "that directory's name", and
  `requirements.paths` may contain more than one `*`. With `specs/*/*/spec.md`, the features
  `specs/a/x/` and `specs/b/x/` share the key `x` at one state. That gives two change intents with
  one name. §5's `name-reuse` is defined only for active/archived OpenSpec names, and §4.3's
  `record-id-collision` only for *distinct* keys.
- **No test touches it.** SDD-INV-59's row names only OpenSpec fixtures, and `speckit-history` is in
  no `Resolve` row.
- **Pre-existing, not new.** Draft 6 §5 and draft 4 §5 have the same text, and rounds 4-6 did not
  raise it. It matters now because draft 7 makes an explicit totality claim (SDD-INV-59) that this
  contradicts.

**Why it matters:** M1 implements both kits (SDD-INV-13). To resolve a Spec Kit claim, or to mint a
Spec Kit change-intent id, M1 has to invent the rule, and two conforming implementations can mint
different ids for the same tree.

### r2-m0-review-round7-20261117-F3 (MED): profile path strings reach git's pathspec parser, so tenant bytes in any historical profile can make `Revisions` fail

**Broken property:**
- SDD-INV-35 ("No tenant content at any commit makes Revisions fail").
- SDD-INV-61 ("… never reader errors").
- §7's "None of these [unlisted failures] can be caused by tenant bytes in history".
- §3's "an entry of an unexpected type is a value, not a reader error".

**Evidence:**
- **The lookup takes tenant strings.** §3 prescribes path lookup "the equivalent of
  `git ls-tree <c> -- <path>`". The paths are the profile's SDD roots and `changes.archive`, and the
  roots come from tenant JSON.
- **§2.1 does not constrain path syntax.** It constrains glob wildcards (`*`, and no `**` in
  `requirements.paths`) and requires OpenSpec globs to "lie under `changes.root`". It does not forbid
  a leading `:`, `..`, a leading `/`, an empty prefix, a trailing `/`, or repeated separators. Each of
  these is a valid profile under §2.1, so it is not `profile-invalid`.
- **`run` passes paths as pathspecs.** It does not set literal pathspecs, and `gitEnvironment` strips
  any `GIT_LITERAL_PATHSPECS` the operator sets. My `<scratch>/ps/` table above shows:
  - **Reader errors (rc 128):**
    - a root `:(glob)openspec/changes` or `:!x`;
    - `../openspec/changes` or `/openspec/changes`;
    - an **empty** root. That one needs no hostile intent: a Spec Kit glob `*/spec.md` has the empty
      string as its "longest wildcard-free directory prefix" (§2.1).

    Under §7, each is a `Reader` failure, so `Revisions` fails at that commit, at every sync, forever.
    That is the permanent-wedge class §7 exists to prevent.
  - **Non-literal reads (rc 0):**
    - `:(top)openspec/changes` reads `openspec/changes`, although no entry has that literal name;
    - `openspec/changes/` lists the tree's *children*, so §3's type test ("a tree → read; a blob … →
      non-regular-file") sees child entries instead of the root's own entry;
    - `./openspec/changes` and `openspec//changes` read `openspec/changes`, while the glob-matching
      rule compares literal path bytes against `./openspec/changes/*/…`, so the root is read and
      nothing in it matches.
- **The kill misses it.** SDD-INV-61's row tests only object types at well-formed paths. Its kill
  ("dereferencing `<c>:<path>`") does not cover a pathspec-interpreting lookup, and no SDD-INV-40 row
  lists a malformed path as a profile violation.

### r2-m0-review-round7-20261117-F4 (LOW): the encoding-cause scope still contradicts itself; an undated archive directory has no unit form

**Broken property:**
- Round 6 condition 3 ("give the two encoding causes one stated scope").
- SDD-INV-55 against SDD-INV-61.
- §2's closed `Skip.Unit` forms.

**Evidence:**
- **Two sections apply `invalid-utf8` to every path.**
  - §3: a non-UTF-8 path "**Inside** the roots … makes its unit unmappable (`invalid-utf8`)".
  - SDD-INV-55's row: `caf\xe9.md` "inside a change directory skips that change with
    `invalid-utf8`".
- **Two others exempt that same file.**
  - §7: "Other files in a change or feature directory … enter the unit's manifest as raw bytes and raw
    path bytes, and are **never a cause**".
  - SDD-INV-61: "files a unit neither parses nor carries are never a cause".

  `caf\xe9.md` in a change directory is neither parsed nor carried: a change intent's
  `artifact_path` is the directory. So an M1 cannot satisfy both rows. The INV-61 kill "an
  over-broad `invalid-utf8`" is exactly what INV-55's row requires.
- **The undated archive directory has no unit form.** §6 makes an undated archive entry's "delta
  files **and its directory** … units skipped with `archive-name-invalid`". §2 allows only
  `requirement <path>`, `change <name>` and `profile`. The directory names no change, so there is no
  `<name>` for it, and its `Skip.Unit` is undefined.

### r2-m0-review-round7-20261117-F5 (LOW): the state key omits the profile's entry mode, and §2 and §3 describe the key differently

**Broken property:**
- SDD-INV-31 (`Resolve` and `Revisions` are "the same function of `c`'s state").
- §2's "Equal keys ⇒ byte-identical SDD inputs", read as "equal outcomes".

**Evidence:**
- **The two definitions differ.**
  - §2: `StateKey` is "the object ids of the profile blob and of every SDD root tree".
  - §3: "The state key includes each root's entry mode and object id".
  - Neither puts the **profile's** mode in the key.
- **Two states can share a key and differ in outcome.** By §3, a symlink or gitlink at the profile
  path is `profile-invalid`, and a regular blob is parsed.
  - A symlink's object is a blob holding its target bytes, and a gitlink records any object id.
  - So a symlink whose target text equals a valid profile, or a gitlink naming the profile blob's id,
    gives the same key as a regular profile, with a different outcome.
- **The provider then disagrees with itself.** It maps each distinct key once (§3), so whichever state
  is mapped first decides both. `Resolve`, which reads `c` directly, can then succeed at a commit
  whose state was mapped as `profile-invalid`. That breaks SDD-INV-31's membership.
- **Severity.** Contrived, but tenant-reachable. Roots are protected only by §3's sentence, not by
  §2's interface comment, which is the one M1 copies into code.

### r2-m0-review-round7-20261117-F6 (LOW): the byte-wise case-mapping rule has no proving vector

**Broken property:**
- Round 6 condition 4 ("pin it with a non-ASCII id vector").
- SDD-INV-7's kill list.

**Evidence:**
- **No vector exists yet.** §4.2 and §4.3 now state the rule and give `expoſe` → `EXPO-E`. But
  `testdata/` holds no id vectors at all; a grep for `vector`, `EXPO` or `expo` finds none.
- **The row would not catch the wrong mapping.** SDD-INV-7's row ("The §4.3 / §5 id vectors in
  testdata … match exactly") names only "a suffix omitting `kind`" as its kill. An M1 that uses
  `strings.ToUpper` / `strings.ToLower`, the idiomatic Go that §4.2 itself warns against, passes
  every named test unless a vector contains `ſ`, `ı` or a Kelvin sign. Nothing requires one.

### r2-m0-review-round7-20261117-F7 (LOW): the pins generator depends on an incidental git context, and the generated file is outside Law 1's enforcement

**Broken property:**
- §13's regeneration claim.
- Law 1 ("carries an `@generated` marker + its regeneration command, and is never hand-edited
  (hook-blocked). `make check` fails if regeneration produces a diff").

**Evidence:**
- **`--pins` needs a git work tree it does not use.** `history_probe.py` runs
  `git -C <arg> rev-list … --all` and starts `git cat-file --batch` at module level, before it checks
  `--pins`.
  - On a plain copy of `testdata/`, `--pins` crashes with exit 128.
  - Run as §13 documents, from the repository, it succeeds only because `testdata/` lies inside
    Proofbound's own work tree. It then walks Proofbound's whole history for nothing.
  - The regenerated bytes are correct (identical, above). The generator's precondition is accidental.
- **The file has no Law 1 marker.** `expected-obligations.json` carries `"generated_by"`, not the
  `@generated by` form. `block-generated-edit.sh` exits 0 for it (the control exits 2), so a hand edit
  is not blocked.
- **No freshness check runs before M1.** No `make check` target regenerates it. My stated-gap row
  shows a fixture edit that changes the true pins while `make check`'s Go tests still pass. §13 assigns
  value-checking to M1's SDD-INV-39, so this is bounded to the M0 → M1 interval, but the file's own
  "never hand-edit" is prose, not enforcement.

### r2-m0-review-round7-20261117-F8 (LOW): residual drift, one unprinted §0 row, and an unrecorded pre-existing wedge

**Evidence:**
- **§10 D** still says "`main` itself for 9 commits". §0 now says 13 consecutive first-parent commits,
  9 of them non-merge.
- **One §0 row is not printed by the probe.** §0 says "Every row is printed by
  `testdata/derive/history_probe.py`", but the probe prints no first-parent count, so the "13 / 9"
  row is not. I reproduced it independently, and it is correct.
- **"Draft 6" labels remain** in:
  - `expected-skips.json`'s `note` ("Draft 6 SDD-INV-46");
  - `unmappable/README.md:1,3`;
  - `history_probe.py:19,29`;
  - `fixture_test.go:322`.
- **A pre-existing wedge is no longer claimed closed, and is not recorded.** Withdrawing the U+0000
  rule from `ValidateRevision` is correct (F2 above). But draft 6 had claimed that the rule closed the
  latent U+0000 projection wedge for `records` / `specdir`. With the rule gone, that wedge is open
  again: a `records` statement carrying U+0000 passes `ValidateRevision`, appends, then fails the
  `requirement_obligations_view.statement` `TEXT` insert on every rebuild.
  - Draft 7 drops the claim but does not record the residual in §14's "Recorded, not owned here", as
    it does for the git connector's non-UTF-8 wedge.
  - No ledger in this repository is exposed today; round 6 grepped every record.

## What holds

- **§10 I is right, and the retroactive hazard is gone.** Every ledger insert path validates, the
  bound has never changed, the `JSON` column returns exact bytes, and the projector validates the same
  bytes. No other M1 change tightens a check that rebuild re-runs.
- **`ErrNoChangeIntent` and the name domain answer round 6 F1 for the cases it named.** The
  proposal-only step, `.`, `..`, `archive` and an `archive/…` path each get a defined outcome.
- **Rules 1 and 2 handle merges correctly.** A merge whose side branch archived `n` gets
  `ErrArchivedEarlier`, and the side commit resolves. Active-at-`c` with an archive entry at a parent
  is rule 1 plus `name-reuse` (`ErrUnmappable`), exclusive of `ErrArchivedEarlier` because of its "no
  active entry" clause.
- **§3's state-level placement covers every object type at a well-formed path.** That placement is
  coherent with a key of mode plus object id for the roots: a tree's id covers every mode and id
  beneath it.
- **§4.2 is OpenSpec's reader for names, sections, REMOVED bullets, RENAMED pairing and fences.** The
  REMOVED-bullet clause touches only Proofbound's labelled statement rule. The probe implements §4.2
  in full, statements included.
- **§4.2 and §4.3's byte-wise mapping is unambiguous and ASCII-safe for truncation.** Its output is
  ASCII, so "first 21 bytes" and "40 bytes" never split a code point.
- **The §9 renumbering is clean.**
  - INV-50 is retired with a reason and re-issued as INV-59. INV-60 and INV-61 are appended.
  - A diff of every statement and row against draft 6 shows no other statement changed.
  - INV-53's row only removes unreachable routes. INV-39's row only strengthens its test (adding "a
    stale pin"). INV-54's statement and row are unchanged.
- **The named kills are reachable.**
  - INV-59: proposal-only → `ErrNoChangeIntent`, plus a failing reader.
  - INV-60: the reused-name fixture.
  - INV-61: a directory at the profile path, which `cat-file blob <c>:<path>` turns into a reader
    error; and a binary in a change directory.
  - INV-39: the stale pin, demonstrated by my stated-gap row.
- **The neutral controls are neutral.** `document-renames` and my neutral edits pass. For a symlink
  at the profile path, a `<c>:<path>`-dereferencing mutant would read the link text and still yield
  `profile-invalid`, so that sub-case does not discriminate. The directory sub-case does.
- **§12 is complete.** The registration grep at HEAD matches §12 exactly, and the M1 / M3 split is
  unchanged and right.
- **The owner decisions are implemented consistently.**
  - `all` excludes `sdd`: `Sync("all")` iterates only the constructed providers, and §12 constructs
    `sdd` only for `sync intent sdd`.
  - Skip printing is defined only on that path, and SDD-INV-56 / SDD-INV-60 test both sides.
  - The helper is committed under an Intent chain.
- **Build Laws:**
  - **Law 1:** `fc9be6a` is `@generated` and `--check`-current. The pins file is the exception (F7).
  - **Law 2:** the plan and the intent SPEC §1 point at draft 7. `skipCauses` copies §7's list into a
    test, which is acceptable for a test oracle until `sdd/SPEC.md` exists.
  - **Law 6:** no provider code exists.
  - **Law 8:** no new dependency.
  - **Law 9:** this round is non-author.

## Fitness for M1

**Not as written.** No finding is HIGH, and the parser and ledger questions are settled. But all three
MED findings sit in the contract M1 encodes, and each lets tenant data produce either a reader-class
failure or an implementation-defined result:

- F1 is the round-6 F1 property, still not a partition.
- F2 is a whole kit without a resolution rule or a change-intent id.
- F3 is a reader error reachable from profile strings.

All three can be fixed in draft 8, or in the `sdd/SPEC.md` that M1 commits, provided a non-author
re-verifies the changed sections. M1 itself should be reviewed at the foundational tier.

Conditions:

1. **Make §6's outcomes a stated partition, with a precedence** (F1). The partition must cover:
   - archived twice by `c`;
   - the re-archive of a reused name;
   - `profile-invalid`, absent-profile and non-tree-root states;
   - a non-tree change directory;
   - `ErrNoChangeIntent` against `ErrUnmappable`, including whether a skipped record counts.

   Extend SDD-INV-59's row with a fixture for each, and keep `Skipped()` and `Resolve` in agreement on
   each.
2. **Give Spec Kit a resolution rule and a change-intent id, or state that Spec Kit changes are not
   resolvable in R2** (F2). Either way, cover the duplicate-feature-key case.
3. **Close §2.1 over path syntax, or make the lookup literal** (F3). Either a malformed root is
   `profile-invalid` (an SDD-INV-40 row), or the reader cannot interpret pathspecs (an SDD-INV-61
   row with `:(glob)…`, `../…`, and the empty prefix of `*/spec.md`).
4. **Pick one `invalid-utf8` scope across §3, §7, INV-55 and INV-61, and give the undated-archive
   directory a unit form** (F4). **Put the profile's mode in the key, and make §2 say what §3 says**
   (F5).
5. **Name a non-ASCII id vector in SDD-INV-7** (F6). **Make `--pins` independent of git and bring the
   pins file under Law 1's marker** (F7). **Fix the F8 drift, and record the `records` / `specdir`
   U+0000 projection residual where the git-connector wedge is recorded.**
