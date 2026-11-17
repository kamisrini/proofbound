---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round8-20261117
status: NEEDS_WORK
reviewed_commit: 1f59ccd0944f2062e52e2ef66a199103a1c3abda
findings:
  - finding_id: r2-m0-review-round8-20261117-F1
    severity: MED
  - finding_id: r2-m0-review-round8-20261117-F2
    severity: MED
  - finding_id: r2-m0-review-round8-20261117-F3
    severity: LOW
  - finding_id: r2-m0-review-round8-20261117-F4
    severity: LOW
  - finding_id: r2-m0-review-round8-20261117-F5
    severity: LOW
  - finding_id: r2-m0-review-round8-20261117-F6
    severity: LOW
  - finding_id: r2-m0-review-round8-20261117-F7
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round8-20261117.md
artifact_sha: f36d305e255836eca68f9d000a5d7a602754794e09ace19684ba5c6dc21853f5
---

# Release 2 M0 (SDD Warranty Connector): draft-8 one-outcome-function, OpenSpec-only, literal-path provider contract. Review round 8

## Provenance

- **Under review:** HEAD `1f59ccd0944f2062e52e2ef66a199103a1c3abda`.
  - `0971e42` adds `docs/decisions/VD-sdd-m1-openspec-first-2026-11-17.md` and its `INDEX.md` line,
    under `Intent: CI-sdd-warranty-m0b-ti5ga9` with the owner's `Authorized-By`.
    `scripts/intent-applicability.sh` reports it `applicable=true`.
  - `33c8549` introduces `docs/plans/R2-sdd-provider-spec-draft8.md`, which supersedes draft 7. The
    same commit carries:
    - the `resolve-edges/` fixture (7 steps) and `expected-resolve.json`;
    - `id-vectors.json`, plus the `--pins` / `--vectors` modes of `derive/history_probe.py`;
    - the updated `expected-skips.json` and `expected-obligations.json` marker;
    - the READMEs, the plan's M1/M2 text, and the intent SPEC §1 pointer.

    It is `applicable=false`.
  - `835130f` changes `kernel/internal/connector/intent/sdd/fixture_test.go` (the self-test), under
    the same Intent with `Authorized-By`. It is `applicable=true`.
  - `1f59ccd` is a one-line `@generated` pin refresh in `docs/verification/p6-connector-reality.md`.
    `scripts/p6-task2-close.sh --check` exits 0 at HEAD.
- **Out of scope:** `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`. Both are
  untracked and unchanged.
- **Follows up:** `docs/verification/verdicts/r2-m0-review-round7-20261117.md` (NEEDS_WORK, F1..F8),
  read in full.
- **Owner decisions taken as given, not re-litigated:**
  - `sync intent all` excludes `sdd`;
  - the helper is committed;
  - `VD-sdd-m1-openspec-first-2026-11-17`: M1 is OpenSpec only, Spec Kit is parked in Appendix A for
    M2, and profile paths are literal.

  I judged only whether draft 8 implements them consistently. No Spec Kit gap that Appendix A parks is
  counted as an M1 finding.
- **Judged against:** the CLAUDE.md Build Laws and this ground truth at HEAD:
  - `connector/git/gitcmd/gitcmd.go`: `Commits` (`:53-75`), `readCommit` → `files` (`:159`,
    `:232-251`), `nulUTF8Strings` (`:430-443`);
  - `connector/intent/intent.go` and `core/core.go` `Event.Validate` (no U+0000 check on payload
    contents);
  - `.claude/hooks/block-generated-edit.sh`, `scripts/invariant-lint.sh`,
    `scripts/invariant-table-lint.sh`.

  `git diff fc9be6a 1f59ccd -- kernel/ scripts/` touches only `intent/SPEC.md` and `intent/sdd/`, so
  round 7's code ground truth (§10 I, §12 sites) still holds.
- **Real-history evidence:**
  - **Repository.** The author's clone of `Fission-AI/OpenSpec` (in the session scratchpad) at
    `f085c850f9440f497ccc76ecbc2859c09e0e0ac7`.
  - **The author's probe**, copied to `<scratch>/probe.py` and run over the clone. Its output is in
    `<scratch>/probe-out.txt`.
  - **My own scans:**
    - `<scratch>/rootblobs.txt`: every non-tree entry directly under `openspec/changes`, and every
      non-tree or undated entry directly under `openspec/changes/archive`, over `rev-list --all`;
    - `<scratch>/fpdual.txt`: the first-parent dual-presence run on `main`, with parent counts.
- **Synthetic evidence:**
  - `<scratch>/oracle/`: my own replay plus a §5/§6/§7 oracle, written from draft 8's text. Its output
    is in `<scratch>/oracle-out.txt`.
  - `<scratch>/vec/`: an independent Go re-derivation of §4.2/§4.3 ids.
  - `<scratch>/td/`: a copy of `testdata/` with no enclosing git repository, for `--pins` / `--vectors`.
  - `<scratch>/ps/`: a git repository for `--literal-pathspecs` semantics, under git 2.34.1.
  - `<scratch>/cal-*`: throwaway module copies of the `sdd` package for calibration.

  None of these wrote a repository file.
- **HEAD at dispatch and at completion:** `1f59ccd0944f2062e52e2ef66a199103a1c3abda`, unchanged. The
  working tree held only the two out-of-scope untracked paths, plus this verdict.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. M1 owns record identity, point-in-time resolution and a new
  ledger-reaching source, so M1 itself should be reviewed at the foundational tier, as rounds 2-7
  recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m0-r8`, a **non-author**. It wrote
  none of the work under review, and it did not review rounds 4-7. This verdict counts under Law 9.
- **Writes:** no repository file other than this verdict. No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m0r8/` in the session scratchpad. It occurs
    21 times in this verdict, including this line.
  - The OpenSpec clone is named only as "the author's clone".

## Round-7 findings re-verified

I checked against the text and the code, not against draft 8's §0 mapping table.

- **F1 (`Resolve` not a partition): MOSTLY FIXED.**
  - Fixed:
    - One function, `changeOutcome`, with a stated first-match precedence decides both `Resolve` and
      the change `Skip` (§5). §6's outcome list defers to it.
    - Archived twice by `c`, re-archive of a reused name, a `profile-invalid` state, a zero-entry
      leftover, and the `ErrNoChangeIntent` / `ErrUnmappable` precedence each have a rule, and each
      except the non-tree cases is pinned in `resolve-edges`.
    - "Has records counts skipped records" is stated.
    - A non-tree root or archive is state-level (§5 rule 1). A non-tree active directory is rule 4.
    - An absent profile has no roots, so every `n` falls to `ErrUnknownChange`. That is defined by
      implication, though no row pins it.
  - Not fixed:
    - Rule 5's "has an entry" predicate is undefined for delta files skipped before parsing (new F1).
    - Rule 4 is applied to non-directory entries that are not changes, and is not applied to non-tree
      archive entries (new F2).
    - `changeOutcome` runs over names that §6's name domain excludes (new F3).
    - Round 7's condition asked for a fixture for each corner. The absent-profile, non-tree-root,
      non-tree-change and skipped-record corners are not in SDD-INV-59's row (F7).
- **F2 (Spec Kit has no resolve rule or id): FIXED BY SCOPE.**
  - `kit: speckit` is `profile-invalid` consistently in §2.1, §9 (SDD-INV-40's row, SDD-INV-64),
    `expected-skips.json` (one `profile` / `profile-invalid` skip at `01-feature-spec`; the three
    profiles are byte-identical, so de-duplication gives exactly one) and §13.
  - SDD-INV-10, 13, 42 and 58 are tagged M2, with no M1 row. The three Spec Kit causes are marked
    "unreachable in M1" in §7.
  - No normative M1 text still requires Spec Kit behaviour. §4.3's `suffix` keeps `kit` in the hash,
    which is harmless.
  - The plan's M1/M2 text matches the decision.
- **F3 (pathspec magic): FIXED.**
  - The §2.1 grammar rejects every form that round 7 showed reaching rc 128: `:(glob)…`, `:!x`, `../x`,
    `/x`, and the empty prefix of `*/spec.md`. It also rejects the non-literal reads (`:(top)`, trailing
    `/`, `./`, `//`).
  - §3 adds `--literal-pathspecs`, and I tested its effect below. Residual: the literal-pathspec half of
    SDD-INV-62 has no reachable kill (new F5).
- **F4 (`invalid-utf8` scope; archive unit form): FIXED.**
  - §3, §7, SDD-INV-55 and SDD-INV-61 now state one scope: a file a unit parses, or a name or path a
    payload carries. SDD-INV-55's row no longer contradicts SDD-INV-61.
  - `archive <entry>` is a fourth `Skip.Unit` form (§2, §6), and the self-test accepts it.
- **F5 (profile mode missing from the state key): FIXED.**
  - §2's interface comment and §3 now give the same key: mode and oid of the profile entry, then
    `changes.root`, `changes.archive` and further roots, with `-` for absent and `""` for no profile.
    That is sufficient for SDD-INV-31: every root lies under `changes.root`, so its tree id covers
    every input.
  - SDD-INV-63's kill is reachable: a symlink whose target bytes equal a valid profile shares the blob
    oid. Residual: SDD-INV-63's statement says more than the key gives (new F6).
- **F6 (no non-ASCII id vector): FIXED.**
  - `id-vectors.json` holds 15 vectors, 6 with non-ASCII input. I recomputed all 15 from §4.2/§4.3's
    text in my own Go, and all 15 match.
  - A `strings.ToUpper` value (`O-MOD-EXPOSE`) is rejected by that re-derivation, so SDD-INV-7's new
    kill is reachable.
- **F7 (pins generator; Law 1): FIXED as conditioned.**
  - `--pins` and `--vectors` run with no enclosing git repository (`GIT_CEILING_DIRECTORIES` set; `git
    rev-parse` fails there). Both regenerate their files byte for byte.
  - Both files carry `@generated by`, and `block-generated-edit.sh` exits 2 for each. It exits 0 for
    the hand-authored `expected-resolve.json` and `expected-skips.json`, which is correct.
  - Residual, as round 7 bounded it: no `make check` target regenerates them before M1. Python is not a
    declared prerequisite, so the freshness half of Law 1 waits for M1's Go re-derivation (SDD-INV-7,
    SDD-INV-39).
- **F8 (drift; unrecorded wedge): FIXED.**
  - §0 and §10 D say 13 / 8. The probe prints `longest run … 13 (8 non-merge)`.
  - My own first-parent walk of `main` finds 13 consecutive dual-presence commits (`332816cc` …
    `acd1ca28`). By parent count, 5 are merges and **8** are non-merge, so round 7's "9 non-merge" was
    wrong and draft 8 corrects it.
  - The "draft 6" labels are gone. Only historical citations remain.
  - The `records` / `specdir` U+0000 wedge is now in §14.

**Round-7 "Fitness for M1" conditions:**

1. **A stated partition with a precedence, a fixture per corner, and `Skipped()` / `Resolve` agreement:
   MOSTLY MET.** The precedence is stated, and the archived-twice, re-archive, profile-invalid and
   zero-entry corners are pinned and agree. Not met: the pre-parse skipped-record case (F1), non-tree
   entries (F2), the name-domain edges (F3), and the missing rows (F7).
2. **Spec Kit rule and id, or out of R2/M1: MET** by the owner decision. The duplicate-feature-key case
   is parked in Appendix A's open items.
3. **Close the path syntax or make the lookup literal: MET.** Both were done. One INV-62 sub-row cannot
   fire (F5).
4. **One `invalid-utf8` scope, an archive unit form, the profile mode in the key, §2 = §3: MET.**
5. **Non-ASCII id vector, git-independent pins under the Law 1 marker, F8 drift: MET.**

## Evidence reproduced

All of this comes from my own commands.

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
- 401 / 1 delta blobs;
- 23 dual-presence commits for 1 name, 4 remote refs, 0 names archived twice;
- the 13 (8 non-merge) first-parent run.

**Entries under the roots, real history** (`<scratch>/rootblobs.txt`):
- **1393 of the 1624 walked commits** carry a regular file (`100644 blob`) directly under
  `openspec/changes`: `IMPLEMENTATION_ORDER.md` at 1393 commits, and `skills-only-migration.napkin.md`
  at 4 of them. That is 2 distinct blob ids. HEAD carries neither.
- No non-tree entry and no undated entry exists directly under `openspec/changes/archive`, so §6's
  "Real history has none" for undated archives holds.
- The probe counts a change only when a path continues below `openspec/changes/<name>/`, so its 135
  excludes these files. §5 does not (F2).

**Fixtures.** Under Go 1.27.4, `go test -count=1 -v ./internal/connector/intent/sdd/` passes all three
tests at HEAD: `TestFixturesReplayAsDocumented`, `TestBranchFixtureReplaysAsDocumented` and
`TestPinnedExpectationsMatchFixtures`.

**My replay of the pins against draft 8's text** (`<scratch>/oracle/`; output in
`<scratch>/oracle-out.txt`).
- **How it works.** It replays each fixture into git the way `replayHistory` does. At every commit it
  then computes, from §2.1, §5, §6 and §7:
  - profile strictness and the literal grammar;
  - the state-level causes;
  - `changeOutcome` with rules 1-8 in order;
  - `Resolve`, with parent archive listings via `rev-list --parents`;
  - de-duplicated Skips, keyed by `(unit, cause, digest)` and ordered by the provider's walk.
- **`expected-resolve.json`:** all 18 rows match (15 for `resolve-edges`, 3 for `branches`). Each
  row's `rule` text names the §6 line that decides it. In particular:
  - `05-rearchive-again`/`again` is `name-reuse`, not `ErrArchivedEarlier`, because §6 line 4 precedes
    line 5;
  - `zero-left` is `name-reuse` (rule 2 precedes rule 5);
  - `a` is `slug-too-short` (rule 3 precedes rule 5);
  - `nope` at `06` is `profile-invalid` (§6 line 2 precedes line 3).
- **`expected-skips.json`:** the exact Skip set matches for all 8 fixtures:
  - `openspec-history` 0;
  - `speckit-history` 1;
  - `branches` 1;
  - `head-on-leftover` 1;
  - `resolve-edges` 6;
  - `duplicate-obligation` 2;
  - `reused-name` 1;
  - `fenced` 0.

  The de-duplication rule decides `resolve-edges`:
  - `change twice` stays at `01`, because its entries and the profile are unchanged through `05` and
    again at `07`;
  - `change again` has two skips, at `04` and `05`, because the entry set differs;
  - neither recurs at `07`, because the profiles of `05` and `07` are byte-identical (SHA-256
    `6acb4340…`) and the entries are unchanged.

  The pins do not depend on how F6's undefined digests are read.
- **Agreement.** Over every replayed commit and every name with a change entry, `Resolve`'s
  `ErrUnmappable` cause equals `changeOutcome`'s Skip cause.

**`--literal-pathspecs`, git 2.34.1** (`<scratch>/ps/`, a tree holding `openspec/changes/{*x, ax,
x[1], :(top)y}`):

| Path argument to `ls-tree <c> --` | default | `--literal-pathspecs` |
|---|---|---|
| `:(glob)openspec/*`, `:!top` | rc 128 (magic unsupported) | rc 0, nothing (literal) |
| `:(top)top` | rc 0, matches `top` (magic applied) | rc 0, nothing (literal) |
| `openspec/changes/*x` | rc 0, the literal `*x` | rc 0, the literal `*x` |
| `openspec/changes/*`, `openspec/changes/x?1?` | rc 0, nothing (no globbing) | rc 0, nothing |
| `openspec/changes/:(top)y` | rc 0, the literal entry | rc 0, the literal entry |
| `../x` | rc 128 | **rc 128** |
| `` (empty) | rc 128 | **rc 128** |
| `openspec/changes/` | the children | **the children** |
| `./openspec/changes`, `openspec//changes` | normalised to the entry | **normalised** |

- `GIT_LITERAL_PATHSPECS=1` behaves like the flag. `cat-file` takes `<rev>:<path>`, not a pathspec, so
  the flag does not change it.
- **Conclusion:** §3's claim holds *in combination*. The flag removes magic, and the grammar removes
  the four forms the flag does not cover (`..`, empty, trailing `/`, `./` and `//`). With the grammar
  in place the flag changes no observable result for `ls-tree` (F5).

**Pins and vectors, outside git** (`<scratch>/td/`, no enclosing repository):
- `python3 -I <scratch>/probe.py --pins` and `--vectors` both exit 0.
- `expected-obligations.json` (SHA-256 `f4c21e6f…`) and `id-vectors.json` (`abdc9f75…`) are
  **byte-identical** to the committed files.
- My Go re-derivation (`<scratch>/vec/`) reproduces all 15 vectors:
  - `expoſe` → `EXPO-E`;
  - `ıd check` → `D-CHECK`;
  - Kelvin sign → `ELVIN`;
  - `Café menu` → `CAF-MENU`;
  - `ünïcode-only` → `n-code-only`;
  - the 28-byte truncation with `29A7A8`;
  - the 40-byte slug truncation;
  - `!!!` → `empty-name-part` and `a` → `slug-too-short`;
  - both BR ids, under the reading that `artifact_sha256` is its lower-case hex payload form.

**Calibration**, on throwaway module copies of the package (`<scratch>/cal-*`):

| Control | Mutation | Expected | Result |
|---|---|---|---|
| Baseline | none | pass | passes |
| Positive 1 | a `branches` resolve outcome → `ErrBogus` | fail | fails: `bad row` |
| Positive 2 | remove `@generated by` from `id-vectors.json` | fail | fails: `lacks an @generated-by marker` |
| Positive 3 | delete `resolve-edges/07-edges` | fail | fails: the pinned `change a` names no unit |
| Positive 4 | a skip unit `archive nope` (no such entry) | fail | fails: `does not name an existing unit` |
| Neutral | reword `again`'s statement at `02` | pass; oracle unchanged | passes; oracle 0 mismatches |
| Stated gap | `05`/`again` → `ErrArchivedEarlier` (round 7 F1(b) regressed, but a listed outcome) | — | self-test **passes**; my oracle reports 1 mismatch. The values are M1's SDD-INV-59 |
| Stated gap | `O-MOD-EXPO-E` → `O-MOD-EXPOSE` in `id-vectors.json` | — | self-test **passes**; my re-derivation reports 1 mismatch. The values are M1's SDD-INV-7 |

Both stated gaps match §13 ("Pin *values* are checked by M1's … tests").

**Records.**
- `link-lint`, `index-check` and `p6-artifact-integrity.sh --check` exit 0 at HEAD, before this
  verdict.
- `invariant-lint` and `invariant-table-lint` check only that the citations a row makes resolve, so
  the four rowless M2-tagged invariants will not break `make check` when the draft becomes
  `sdd/SPEC.md`.

## Findings

### r2-m0-review-round8-20261117-F1 (MED): rule 5's "has an entry" is undefined for delta files skipped before parsing, and contradicts §4.2 for an unpaired-only RENAMED file

**Broken property:**
- SDD-INV-59 ("exactly one" outcome per `(c, n)`).
- SDD-INV-46 ("each fixture's Skip set is exactly the pinned one").
- Round 7 condition 1's last bullet ("including whether a skipped record counts").

**Evidence.** All of this follows from the text.
- **How the rule decides.** §5 rule 5 is "none of `n`'s delta files has an entry", and "'Has records'
  counts records whether or not they are skipped". §4.1 defines a record as a matched delta file "that
  has at least one requirement entry"; "A file with zero entries yields no record".
- **It works for causes found after parsing.** `obligation-collision`, `empty-name-part`,
  `owner-uncovered`, `record-id-collision` and `payload-too-large` all leave a known entry count, so
  rule 5 versus rule 6 is decided. `unmappable/duplicate-obligation` (cascade) is that case, and it is
  pinned.
- **It is undecidable for causes found before or instead of parsing.** Take a change whose only delta
  file is:
  - a **symlink** or **gitlink**. The `Artifact` carries `Regular=false` and **no Bytes** (§2), so the
    file has no entries. SDD-INV-54 nevertheless makes it a `requirement` unit skipped with
    `non-regular-file`. Is it "a record that is skipped" (rule 6 → `cascade`) or "a file with no entry"
    (rule 5 → `ErrNoChangeIntent`, and no change Skip)?
  - **not valid UTF-8**, or **carrying U+0000**. §4.2's Input step rejects the file before the
    parser's rules run, so its entry count is not defined by §4.2.
  - **a RENAMED section holding only an unpaired `FROM` line.** It has zero entries, because only
    pairs are entries. §4.1 and SDD-INV-43's row ("a zero-entry file yields no record and no skip")
    say no record and no skip. §4.2 ("Any unpaired line makes the record unmappable
    (`unpaired-rename`)") and the same row's "unpaired renames are skipped with their cause" say a
    skipped record.
- **Two conforming M1s therefore differ** on:
  - `Resolve(c, n)`: `ErrNoChangeIntent` against `ErrUnmappable:cascade`;
  - whether `Skipped()` holds a `change n` / `cascade` unit, and, for the unpaired-only file, a
    `requirement` unit at all.
- **The tests cannot see it.** These are exactly the byte-level cases §13 leaves to M1's synthesized
  tests (SDD-INV-52..55). SDD-INV-46 requires their skip sets to be exact, but no text fixes them.
  SDD-INV-59's row has no such case.

### r2-m0-review-round8-20261117-F2 (MED): §5 makes every non-directory entry under `changes.root` a change unit skipped as `non-regular-file`, which §7 does not define and real history hits at 1393 of 1624 commits; non-tree archive entries escape rule 4

**Broken property:**
- §7's closed cause definition: `non-regular-file` is "a symlink (mode `120000`) or gitlink (`160000`)
  where a unit reads a file or directory; at state level, a non-tree root or archive".
- SDD-INV-61 ("files a unit neither parses nor carries are never a cause").
- SDD-INV-54 (a symlinked change directory skips with `non-regular-file`).
- SDD-INV-46's exactness on real-shaped trees.
- §0's evidence-first claim that every consequence was measured.

**Evidence:**
- **What §5 says.**
  - A change entry is "the active directory `<root>/<n>` (**of any object type**)".
  - `changeOutcome` runs "for every name that has at least one change entry".
  - Rule 4 is "`non-regular-file`: the active entry is not a tree".
- **What §5 then does to a regular file.** A **regular** file directly under the root is a change
  entry, and it gets `non-regular-file`. Examples: `openspec/changes/README.md` or a `.gitkeep`. The
  Skip is `change README.md`, and `Resolve(c, "README.md")` is `ErrUnmappable:non-regular-file`.
- **§7 and SDD-INV-61 disagree.** §7's definition does not cover a `100644` blob. Under SDD-INV-61, a
  file no unit parses or carries is never a cause.
- **The reference repository hits it.**
  - `openspec/changes/IMPLEMENTATION_ORDER.md` is a `100644` blob directly under the root at **1393 of
    the 1624** commits on the provider's walk. `skills-only-migration.napkin.md` is there at 4 of them
    (`<scratch>/rootblobs.txt`).
  - A literal M1 therefore reports a `non-regular-file` change Skip for a Markdown file on 86% of the
    commits of the one repository the contract was measured on. A §7-literal M1 reports nothing.
  - Neither §0 nor the probe mentions this. The probe counts a change only from paths under
    `openspec/changes/<name>/`.
  - No fixture has a file directly under the root, and my oracle confirms the pins are silent on it.
- **The converse gap.** Rule 4 tests only "the **active** entry". A dated archive entry
  (`<archive>/2026-01-01-x`) that is a symlink, gitlink or blob is still a change entry for `x` (§6
  names any entry). Rule 4 does not fire, and rule 5 sees no delta files, so the outcome is "absent":
  `ErrNoChangeIntent` with no Skip. SDD-INV-54 says a symlinked change directory skips with
  `non-regular-file`, and the archived entry is that change's directory (§5: `artifact_path` is the
  change directory).

**Why it matters.** This is not a wedge, because `Revisions` still succeeds. But it is a
tenant-reachable self-contradiction in the cause table that M1 encodes. The real corpus sits on the
contradicted branch.

### r2-m0-review-round8-20261117-F3 (LOW): `changeOutcome` runs over names that §6's name domain excludes, so `Skipped()` and `Resolve` can disagree

**Broken property:** SDD-INV-59 ("`Skipped()` and `Resolve` agree on the cause").

**Evidence:**
- **The two sections use different name sets.**
  - §6's name domain makes `.`, `..` and, when the archive lies directly under the root, `archive`
    `ErrUnknownChange` before any other rule.
  - §5 computes `changeOutcome` for every name with a change entry, and applies no domain.
  - `<root>/archive` is itself an active entry of type tree under §5's definition.
  - The archive capture `(.+)` admits `.`, `..` and `archive`.
- **The disagreement.** At a state carrying `<archive>/2026-01-01-archive`:
  - name `archive` has an active entry (the archive tree) and an archive entry, so rule 2 gives
    `name-reuse`;
  - `Skipped()` reports `change archive` / `name-reuse`;
  - `Resolve(c, "archive")` is `ErrUnknownChange`.

  Likewise `<archive>/2026-01-01-.` yields `change .` / `slug-too-short`, while `Resolve(c, ".")` is
  `ErrUnknownChange`.
- **A smaller gap.** An archive nested deeper than one level under the root (allowed: "must lie
  under") makes its first intermediate directory an active change.
- **Severity.** All of this is contrived. Real history has no such entries (my scan finds no undated or
  odd archive entries).

### r2-m0-review-round8-20261117-F4 (LOW): the literal glob grammar admits globs whose matches have no §4.1 shape, and a glob with no `*`

**Broken property:**
- SDD-INV-7 / SDD-INV-44 (deterministic, content-addressed record ids).
- §3 (a root is a tree or the state is unmappable).

**Evidence:**
- **What the grammar allows.** §2.1 admits any path under `changes.root` in which a segment "**may**"
  be `*`.
- **Matches outside §4.1's shapes.** `openspec/changes/*/spec.md` and
  `openspec/changes/*/*/specs/*/spec.md` are both profile-valid. Neither has §4.1's shape
  (`<root>/<name>/specs/<capability>/spec.md` or `<archive>/YYYY-MM-DD-<name>/specs/<capability>/spec.md`).
  §4.3's key `<name>/<capability>`, and so the BR id, and the owning change for rule 5/6 are then
  undefined.
- **A glob with no `*`.** `openspec/changes/x/specs/core/spec.md` is profile-valid, and its "literal
  prefix" is the whole path. That makes an SDD root a blob, so by §3 **every state carrying the file**
  is `non-regular-file` at state level.
- **Context.** §2.1 already says "Kit rules are code, not profile". The profile can still select shapes
  the code has no rule for.
- **Not reachable from fixtures or real history.** Every committed OpenSpec profile (35 valid plus 1
  `unknown`-field profile) uses the two standard globs, and all of them pass the grammar by my check.

### r2-m0-review-round8-20261117-F5 (LOW): SDD-INV-62's literal-pathspec half has no reachable kill, and its `*x` case rests on a false premise

**Broken property:** §9's rule that each proving test names the wrong implementation it kills, and
Law 9's calibration corollary.

**Evidence** (table above):
- **The listed cases never reach git.** Every malformed root in SDD-INV-62's row (`:(glob)…`, `:!x`,
  `../x`, `/x`, `*/spec.md`, a trailing `/`, `./`, `//`, a space) fails the §2.1 grammar first, so it
  is `profile-invalid`.
- **So the named kill cannot fire.** A mutant that drops `--literal-pathspecs` but keeps the grammar
  passes every case. The kill is "passing tenant strings as pathspecs (rc 128)".
- **The `*x` premise is false.** The row says a change directory literally named `*x` is "valid only
  through `--literal-pathspecs`". On git 2.34.1, `ls-tree <c> -- openspec/changes/*x` returns the
  literal entry **without** the flag. `ls-tree` never expands wildcards: `openspec/changes/*` lists
  nothing in either mode.
- **No grammar-valid path is affected.** No grammar-valid path starts with `:`, so for `ls-tree` and
  `cat-file` the flag is defence in depth with no observable effect.

The grammar half of the row does kill a lenient grammar. For example, `:(glob)openspec/changes` under
the flag reads as an absent root rather than `profile-invalid`.

### r2-m0-review-round8-20261117-F6 (LOW): SDD-INV-63 promises equal outcomes for equal keys, which the pinned fixture contradicts, and the Skip de-dup digest is undefined for several unit kinds

**Broken property:**
- SDD-INV-63's statement.
- §7's de-duplication rule, which SDD-INV-46's exactness depends on.

**Evidence:**
- **SDD-INV-63 versus the pins.**
  - SDD-INV-63 reads "equal keys imply equal outcomes".
  - In `branches`, steps `07-side-archive` and `08-merge-side` have the same `openspec` tree
    (`f00d36fb…`) and the same profile blob (`97b8bcfe…`), so the same state key.
  - `expected-resolve.json` pins `Resolve(07, add-x) = revision` and
    `Resolve(08, add-x) = ErrArchivedEarlier`.
  - §2's comment ("byte-identical SDD inputs and equal object types") and SDD-INV-31 (state *and
    parents' archive listings*) are right. SDD-INV-63's "outcomes" holds only for `changeOutcome`.
  - An M1 that caches `Resolve` by key, as the statement licenses, fails SDD-INV-59's `08` row. The
    failure is caught, but the invariant text points the implementer the wrong way.
- **The unit digest is undefined in four cases.** The digest is "the manifest digest of the unit's
  inputs, or of the profile entry for `profile`". It has no definition for:
  - a `name-reuse` change: it has two or more directories, and §5 says only that "`artifact_path` is
    the change directory";
  - a non-tree change entry, which has no files;
  - an `archive <entry>` unit, whose "delta files are not read";
  - a state-level `non-regular-file`. Its cause lies in a root or the archive, but the `profile` digest
    covers only the profile entry. So a symlinked root at one state and a blob archive at a later state
    under the same profile collapse to one Skip, and the second defect is never reported.

  The committed pins do not depend on any of these readings (above).

### r2-m0-review-round8-20261117-F7 (LOW): SDD-INV-59's row misses corners round 7 asked for; residual drift

**Evidence:**
- **SDD-INV-59's row omits four corners.** It pins `resolve-edges` and `branches` only. These corners,
  listed in round 7 condition 1, have no row and no agreement check:
  - an absent profile;
  - a non-tree root or archive (SDD-INV-61 checks the Skip, not `Resolve`);
  - a non-tree change directory (SDD-INV-54 likewise);
  - a pre-parse skipped record (F1).

  The replay helper cannot commit symlinks, but §13 already assigns synthesized hostile cases to M1's
  tests, and the row could name them.
- **SDD-INV-52's list is partial.** Its parenthetical gives 8 units, covering 7 causes, as "one unit
  per M1-reachable §7 cause". It omits 8 of §7's 15 M1-reachable causes: `profile-invalid`,
  `owner-uncovered`, `sponsor-uncovered`, `obligation-collision`, `empty-name-part`, `unpaired-rename`,
  `record-id-collision` and `cascade`. Whether the list is meant to be exhaustive is unclear.
- **The README has stale rows.**
  - `testdata/README.md` still maps `openspec 02-draft-proposal-only` to "INV-59 (Resolve →
    ErrNoChangeIntent)", but no `expected-resolve.json` row covers `openspec-history`.
  - The `speckit` rows cite INV-10 / INV-42 without the M2 tag, and do not cite INV-64.

## What holds

- **The precedence is a real total order, and it is applied once.**
  - §6's outcome list routes every `ErrUnmappable` through `changeOutcome`, so `Resolve` and
    `Skipped()` cannot disagree on any name §5 and §6 both cover. My oracle confirms this on every
    replayed commit.
  - Merges and roots are right:
    - a side-branch archive merged in gives `ErrArchivedEarlier` at the merge, and the side commit
      resolves;
    - a root archives what it carries;
    - re-dating an archive entry is "archived earlier";
    - two branches archiving one name and merging is `name-reuse`.
- **Every `expected-resolve.json` and `expected-skips.json` row is correct by the text,** including the
  de-dup outcomes at `07-edges`.
- **The literal grammar is complete and consistent with every committed profile.** All 36 OpenSpec
  profiles (`openspec-history`, `branches`, `head-on-leftover`, `resolve-edges`, `unmappable/*`,
  `fenced`) use grammar-valid roots, archive and globs. `resolve-edges/06` is invalid only for its
  `unknown` field, and the 3 `speckit` profiles are invalid for `kit`.
- **Literal pathspecs plus the grammar close round 7's F3.** No tenant string can reach git as magic or
  as a fatal pathspec.
- **The state key is coherent and sufficient for SDD-INV-31.** It is consistent between §2 and §3, and
  SDD-INV-63's symlink kill is reachable.
- **The id vectors are an independent, discriminating oracle.** All 15 reproduce from the text, and the
  `ToUpper` / `ToLower` mutant fails `expoſe` and the Kelvin sign.
- **Generated pins.** Both regenerate byte for byte outside git, carry the Law 1 marker, and are
  hook-blocked against hand edits.
- **§9 discipline.**
  - No statement changed meaning without a retirement. The only statement edits are the M2 tags on
    INV-10, 13, 42 and 58; INV-62, 63 and 64 are appended.
  - The row edits narrow or strengthen: INV-7, 40, 52, 53, 55 and 59.
  - The new kills for INV-7, 59, 63 and 64 are reachable, and INV-59's three kills are exactly the
    round-7 F1 regressions my calibration reproduces.
  - The neutral controls are neutral: INV-55's non-parsed file, INV-39's `document-renames`, and my
    reword.
- **The decision record is well-formed.**
  - It has the title shape, Status, Date, Context, a testable Decision, Consequences (easier, harder,
    given up) and a concrete "Revisit when", and it uses the date suffix of a phase decision.
  - It appears in the regenerated `INDEX.md`, and it was committed under an Intent with
    `Authorized-By`.
  - Its history claims are accurate: 2 HIGH in round 3, 2 MED in round 6, 3 MED in round 7, and seven
    rounds.
  - Draft 8 implements its regex, its no-`.`/`..` rule, its whole-segment `*`, and its
    literal-pathspec lookup. It chooses the flag form, which `gitEnvironment`'s `GIT_*` scrub cannot
    strip.
- **Scope consistency.**
  - The plan's M1 and M2 paragraphs, the intent SPEC §1 pointer (now draft 8), §11's milestone split,
    §13 and `expected-skips.json` all say OpenSpec-only M1.
  - Appendix A is labelled non-normative and lists round 7 F2's items, including the multi-`*`
    feature-key collision.
- **§14's two "recorded, not owned here" items are accurate.**
  - **The git connector's non-UTF-8 wedge.** `gitcmd.Commits` calls `readCommit` → `files`
    (`:159`), and `files` rejects any non-UTF-8 path in a commit's diff through `nulUTF8Strings`
    (`:437`). Every path enters history through some walked commit's diff, so "anywhere in history"
    holds.
  - **The `records` / `specdir` U+0000 wedge.** `ValidateRevision` and `Event.Validate` have no check
    on U+0000 in payload contents. I did not re-execute the `TEXT` insert failure against Postgres; it
    is round 6/7's evidence.
  - Both are recorded as "raised to owner, not filed" in `notes/state.md` and in the 2026-11-17
    journal. Neither has a `docs/deferrals.tsv` row, so the journal is their only append-only record.
- **Build Laws:**
  - **Law 1:** markers and hook enforcement are now in place. The freshness half waits for M1, as
    bounded.
  - **Law 2:** pointers updated; no second status file.
  - **Law 6:** no provider code exists.
  - **Law 8:** no new dependency, and Python stays a documentation-only probe.
  - **Law 9:** this round is non-author.

## Fitness for M1

**Not as written, but close.** No finding is HIGH. Round 7's three MED findings are fixed: F2 by owner
scope and F3 by construction, and F1 is mostly fixed. The two remaining MED findings are narrow text
defects in the outcome function M1 encodes. One sits on the real reference corpus (F2), and one is the
last clause of round 7's partition condition (F1). Both can be fixed in the `sdd/SPEC.md` that M1
commits, or in a draft 9, provided a non-author re-verifies the changed §4.1, §5, §6 and §7 text. M1
itself should be reviewed at the foundational tier.

Conditions:

1. **Define rule 5's predicate for every requirement unit** (F1). Say whether a delta skipped with
   `non-regular-file`, `invalid-utf8` or `nul-byte`, or holding only unpaired RENAMED lines, "has
   records". Reconcile §4.1 and SDD-INV-43 with §4.2 for the unpaired-only file. Pin one such case in
   SDD-INV-59 and SDD-INV-46.
2. **Make a change entry's type rule agree with §7 and SDD-INV-61** (F2). Say what a regular file
   directly under `changes.root` is (the real corpus has one at 1393 commits). Apply the non-tree rule
   to dated archive entries as SDD-INV-54 requires. Add the `IMPLEMENTATION_ORDER.md` shape to a
   fixture and to §0.
3. **Apply §6's name domain inside §5** (F3). **Constrain `requirements.paths` to the shapes §4.1 keys**,
   or define the key for any shape the grammar admits, including a glob without `*` (F4).
4. **Give SDD-INV-62's literal-pathspec half a kill that can fire, or state that it is defence in depth
   with no row** (F5). **Make SDD-INV-63 say `changeOutcome`, and define the de-dup digest for
   multi-entry, non-tree, `archive` and state-level units** (F6). **Fix the F7 rows and drift.**
