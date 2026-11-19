---
schema: vera.verdict.v1
verdict_id: r2-m1-lowfix-recheck-20261118
status: ACCEPTABLE
reviewed_commit: 2d7f0f61239396beddad198e008e020f91b84984
findings:
  - finding_id: r2-m1-lowfix-recheck-20261118-F1
    severity: LOW
  - finding_id: r2-m1-lowfix-recheck-20261118-F2
    severity: LOW
  - finding_id: r2-m1-lowfix-recheck-20261118-F3
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m1-lowfix-recheck-20261118.md
artifact_sha: b6f69804e7c99f2ea30316cbc1c7e9fda76ebb289d8377ac956c857b1db4a249
---

# Release 2 M1 (`intent.sdd`): targeted non-author re-check of the LOW-finding fixes

## Scope and provenance

- **Under review:** `git diff 71128fe..2d7f0f6`. HEAD `2d7f0f61239396beddad198e008e020f91b84984`
  at both start and end of the review; the tree did not move.
- **Commits in range:**
  - substantive: `5aafb55` (new `recheck_kill_test.go`, INV-35 rework in `sdd_test.go`, cli
    `printableUnit` + `cli/printable_unit_test.go`, `history_probe.py`, SPEC), `b7bf296` and
    `04f7051` (SPEC §9 citation format only; the diff is confined to the "Also proven by" clauses
    of rows 47, 54, 56, 59, 60, 61, 65);
  - notes only: `affa0da`, `b50f61b` (each touches only `notes/state.md`);
  - generated-pin refreshes only: `9a774c3`, `25bcdb0`, `2d7f0f6`
    (`docs/verification/p6-connector-reality.md`) and `7b8e458`
    (`docs/verification/p6-event-route-matrix.md`). Each changes exactly the one `**Code commit:**`
    line under its `@generated` marker.
- **Contract:** `kernel/internal/connector/intent/sdd/SPEC.md`; `cmp` confirms it is byte-identical
  to `docs/plans/R2-sdd-provider-spec-draft10.md` at HEAD.
- **Out of scope:** the untracked `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261118-m1lowfix`, a **non-author**. It
  wrote none of the work under review and none of the prior verdicts, so this verdict counts under
  Law 9.
- **What I ran:** Go 1.27.4 with `$HOME/go/bin` on PATH.
  - Mutants: a `git archive` export of HEAD's `kernel/` at `<scratch>/base/`, mutated one at a time
    in `<scratch>/work/` by `<scratch>/mut/run.py` (exact-once substitution, `go vet` compile check,
    file restored from `<scratch>/base/` after each run; `<scratch>/work/` diffed clean against
    `<scratch>/base/` afterwards). Results: `<scratch>/mut/results*.tsv`, logs `<scratch>/mut/logs/`.
  - Whole-package cli mutants: `go test -overlay` against the repository's own package (the
    `kernel/`-only export lacks `docs/`, which some cli tests read), outputs in `<scratch>/ov/`.
  - Probes: `<scratch>/probe-k/` (exhaustive `printableUnit` sweep; pre-rework test vs HEAD-fatal
    mutant; shallow-clone construction probe).
  - `history_probe.py` re-run on the author's OpenSpec clone (`f085c850`): `<scratch>/probe.txt`.
  - Read-only gates: `invariant-lint`, `invariant-table-lint`, `spec-numbering-lint`,
    `law-citation-lint`, `prescription-lint`, `link-lint`, `cleanroom-lint`, and
    `scripts/p6-task2-close.sh --check`, `scripts/p6-route-matrix.sh --check`,
    `scripts/p6-census.sh --check`: all exit 0.
  - `go test -count=1` at HEAD in the repository: `internal/connector/intent/sdd` ok (66.9 s),
    `internal/cli` ok (67.6 s). `make check` was not run, per the dispatch.
- **Repository untouched:** no repository file written; `git status` at the end equals the start
  (only the two permanent untracked paths). `<scratch>` stands for the reviewer directory
  `reviewer-lowfix/` in the session scratchpad.

## 1. Closure of prior findings

### Re-check `r2-m1-recheck-round1-20261118`

| Finding | Status | Evidence |
|---|---|---|
| F1 regular file AT `specs` unpinned; old rule in §0/§7/§9 text | **Closed** | `survivors_test.go::TestRegularFileNamedSpecsIsIgnored` kills `F1-revert-specs` (`a regular file named specs became [change flat\|cascade requirement …/specs\|non-regular-file]`). §0's F1 answer now says a symlink/gitlink at `specs` or `specs/<cap>` is a unit and a regular file there is ignored; §7's `non-regular-file` row names the exception; SDD-INV-54's row states the ignored case and `ErrNoChangeIntent`. All three now agree with §4.1 and with `provider.go` (`if !specs.Regular`, `if !capDir.Regular`) |
| F2 INV-59 reader failure proven only for `Read` | **Closed** | `Parents`: `TestParentsFailureIsNotASentinel` kills `V-parents-error-as-unknown`. `List`: `TestResolveParentListFailureIsAReaderError` kills three hand mutants (List error → `ErrUnknownChange`, → `continue`, → `ErrUnmappable:cascade`). The sweep-direction limit is closed by narrowing: row 59 now says "`mapState`'s skips and `Resolve` agree", which is what the sweep checks |
| F3 ordering unpinned; C1/U+2028 unquoted; regex dialect | **Closed** | Both sort mutants (`return false` comparators) are killed by `TestOutputOrderIsDeterministic` (`-count=3`, every run failed). `printableUnit` quotes C1 and U+2028/U+2029 (section 3). §6 now fixes the dialect: Go RE2, `.` excludes `\n` — true (section 5) |
| F4 probe count not type-discriminating | **Closed** | `history_probe.py` types by mode (`REGULAR = ("100644","100755")`) and adds the "AT `specs`" and "symlinks or gitlinks" counts (section 6) |
| F5 round-1 F8 survivors undeclared | **Closed** | All seven named survivors die now (table below), so nothing needs declaring. `docs/allowed-survivors.txt` declares four other, REDUNDANTLY-PINNED sweep survivors, which are outside this finding |

Seven survivors added by `48614c3` (before this range), re-killed here by me:

| Mutant | Killing test | Result |
|---|---|---|
| INV65-oid-only (`entryLine` drops the mode) | `TestStateKeyModeOnlyDifference` | killed |
| INV54-hash-symlink-as-content (`contentLine` always hashes bytes) | `TestSymlinkManifestLineIsGitIdentity` | killed |
| F2-untyped-parent-archive (drop `isChangeEntryMode`) | `TestParentArchiveFileDoesNotArchive` | killed |
| F1-revert-specs (`if !specs.Regular` → `if true`) | `TestRegularFileNamedSpecsIsIgnored` | killed |
| V-parents-error-as-unknown | `TestParentsFailureIsNotASentinel` | killed |
| F9-no-skip-sort | `TestOutputOrderIsDeterministic` (×3) | killed |
| F9-no-rev-sort | `TestOutputOrderIsDeterministic` (×3) | killed |

### Round-1 `r2-m1-review-round1-20261117` F8 bullets

| Bullet | Status | Evidence |
|---|---|---|
| INV-65(a) oid-only key | **Closed** | Killed (above); row 65 adds clause (c) for the mode-only case |
| INV-54 symlink hashed as content | **Closed** | Killed (above), plus `TestSymlinkManifestLineFormatIsPinned` recomputes the `<mode> <oid>  <path>` line from git |
| entry-condition F2 (untyped parent archive) | **Closed** | Killed (above) |
| INV-61 count and gitlink root | **Closed** | `TestStateLevelSkipsAreCounted` counts three `profile\|non-regular-file` Skips. It kills: de-dup ignoring the digest (got 1), gitlink root not a state skip (got 2), state digest omitting the offending entry (got 1). The same de-dup mutant **survives** the older `TestStateLevelObjectTypes`, so the new test adds real power |
| INV-35 hostile set and HEAD variant | **Closed** (with F2 below) | The rework adds the collision pair, `owner-uncovered`, `sponsor-uncovered`, and five variants including three HEAD variants. HEAD-fatal mutants for `owner-uncovered`, `record-id-collision`, `nul-byte`, any non-profile cause, and `profile-invalid` are all killed. The pre-rework test (from `71128fe`) **survives** "HEAD-fatal for any data cause". The "every other unit emitted" half is weaker (F2) |
| INV-47 just-under-limit | **Closed** | `TestExactBoundPayloadSyncs` builds a requirement payload of exactly `core.MaxPayloadBytes`, self-calibrated by a size measurement. It kills `>`→`>=` in `buildRecord`, `>`→`>=` in `core.Event.Validate`, `>`→`>=` in the intent connector's guard, and a `+1` slack in `buildRecord` |
| INV-56 `verify` | **Partly closed** | Row 56 now says honestly what is tested: the provider set of `all`. `verify`'s two call sites pass `"all"` (`cli.go` lines 902 and 921 at HEAD), as the row says. But neither that call-site property nor "all never constructs sdd" in a constructible repository is pinned by any test (F1) |
| INT-INV-4 replay row | **Closed** | §12's wording is corrected from "replay" to the exact-bound argument. `intent_test.go::TestSyncRejectsOversizeBeforeAnyAppend` already has the exact-bound row (intent SPEC row 143). The bound `MaxPayloadBytes = 1 << 20` and `len(e.Payload) > MaxPayloadBytes` in `core.Event.Validate` are unchanged since `33d46d4`, the commit that introduced core events, and `store.Append` validates every event. So "no event ledgered before M1 exceeds it" is true |

## 2. Production tree

- `git rev-parse 77cc6ba:kernel/internal/connector/intent/sdd` = `066f8b01d873986fc645d4e402a13739408ce55f`,
  as the admission addendum records.
- All six production files (`build.go`, `ids.go`, `parse.go`, `profile.go`, `provider.go`,
  `reader.go`) have the **same blob id** at `77cc6ba` and HEAD. The set of non-test `.go` files is the
  same (6). The package tree at HEAD is `b5165e82…`. It differs only in test files, the SPEC and
  `testdata/derive/history_probe.py`.
- The admission therefore still binds the production code. The addendum's parenthetical about
  what later commits changed is now incomplete (F3).

## 3. cli `printableUnit` (C1, U+2028, U+2029)

- **Correctness.** The new condition is `r < 0x20 || (0x7f <= r <= 0x9f) || r == 0x2028 || r == 0x2029`.
  - An exhaustive sweep runs `printableUnit("x"+r+"y")` for every scalar value U+0000-U+10FFFF
    except surrogates. Exactly 67 runes are quoted (32 C0 + DEL + 32 C1 + 2 separators).
  - No output, quoted or not, contains a rune from that set or any of `\n\r\v\f`.
  - Each of the 128 single invalid bytes 0x80-0xFF yields a quoted, valid-UTF-8 string.
  - So `strconv.Quote` output stays on one line for every input it now quotes. This holds because
    `strconv.Quote` escapes every rune that is not `strconv.IsPrint`, and none of the 67 is printable.
- **Regressions.**
  - Only names containing a C1 control or U+2028/U+2029 change form. Those were unquoted before and
    could break lines in some terminals.
  - NBSP (U+00A0, just past the C1 range) and `café` stay unquoted, and the test pins both.
  - `printableUnit` has one caller, the `skip` line of `sync intent sdd`.
  - `go test ./internal/cli/` and `TestSyncIntentSDDQuotesHostileUnits` pass.
- **Residuals (not findings).**
  - Bidi controls (e.g. U+202E) stay unquoted. They can reorder a line visually but cannot split it,
    which is outside SDD-INV-60's one-line claim.
  - An unquoted name that itself begins with a `"` can look like a quoted one. This predates the
    change and does not split lines.

## 4. Calibration of every new or reworked test

Each row is one hand mutant run against the named test alone, unless marked "whole package".

| Test | Positive control(s): must die | Result | Neutral control: must survive | Result |
|---|---|---|---|---|
| `TestResolveParentListFailureIsAReaderError` | List error → `ErrUnknownChange`; → `continue`; → `unmappable("cascade")` | all killed | reword the List error message | survived |
| `TestStateLevelSkipsAreCounted` | de-dup by `(unit,cause)` only; gitlink root (`160000`) exempted from the state check; state digest = profile line only | all killed | swap the order of the two digest lines (`digestLines` sorts) | survived |
| `TestExactBoundPayloadSyncs` | `buildRecord` `>`→`>=`; `buildRecord` `> MaxPayloadBytes+1`; `core.Event.Validate` `>`→`>=`; intent-connector guard `>`→`>=` | all killed | `> MaxPayloadBytes+0` | survived |
| `TestHistoricalFailClosedIsRecoverable` (reworked) | HEAD-fatal for `owner-uncovered`; for `record-id-collision`; for `nul-byte`; for any non-profile cause; for `profile-invalid`; `sponsor-uncovered` fatal anywhere; `record-id-collision` fatal off HEAD | all killed | reword the `read` error in `Revisions` | survived |
| ″ (key-variable check) | a state carrying any Skip silently emits **no** revisions | **survived** this test; killed by the whole package (`sdd_test.go:623`, `:679`) | — | — (F2) |
| `TestPrintableUnitQuotesEveryLineBreaker` | revert to C0+DEL only; drop U+2029; C1 upper bound 0x9e; over-broad (quote every non-ASCII rune) | all killed | reorder the four disjuncts | survived |
| `TestAllSelectionNeverConstructsSDD` | `all` unconditionally constructs `sdd` too | killed | assign the provider slice to a local, then return it | survived |
| ″ (key-variable check) | `all` adds `sdd` **when it can be constructed** | **survived** this test, `TestIntentAllExcludesSDD`, and the whole `internal/cli` package | — | — (F1) |
| ″ | `verify`'s first, or both, intent syncs pass `"sdd"` instead of `"all"` | **survived** the whole `internal/cli` package and the sdd CLI tests | — | — (F1) |

**Mechanics:**
- Each HEAD mutant identifies HEAD as `commits[len(commits)-1]`, because `CommitSHAs` walks
  `--reverse --date-order`.
- My first batch used `commits[0]` (the root commit, which carries no Skip). I discarded it as a
  harness error, not counted it.
- Two first-try sort mutants failed `go vet` (unused variables). They were marked INVALID and re-run
  as `_, _ = a, b; return false`.
- No other mutant was invalid.

**Fixtures holding the key variable constant (question 4).**
- **`TestAllSelectionNeverConstructsSDD`** checks `all` only outside a repository. **`TestIntentAllExcludesSDD`**
  checks it only on a shallow clone, which `gitcmd.New` itself refuses ("repository history is
  incomplete"). So both tests hold "sdd is constructible" at *false* (F1).
- **`TestHistoricalFailClosedIsRecoverable`'s** emission assertion uses `good`. `good` is already emitted
  by the clean first commit, so the assertion cannot tell whether the defect states emit their
  clean units (F2).
- **`TestExactBoundPayloadSyncs`** self-calibrates its size, so nothing is held constant. The
  `TestStateLevelSkipsAreCounted` emission check uses a change that exists only at the final clean
  state, so it is not held constant either.

## 5. SPEC truth checks

- **§0 rows.** The rows "Regular files AT a change's `specs`: 0" and "Symlinks or gitlinks at or
  directly under a change's `specs`: 0" match the re-run probe, line for line.
  - Independent check: `ls-tree -r` over every `rev-list --all` commit of the clone finds **no**
    entry under `openspec/changes` with a mode other than `100644`, and exactly one regular file
    directly under a `specs/` (`…/2026-01-30-opencode-command-references/specs/no-changes.md`).
  - So all three counts are true.
- **§6 archive-pattern dialect paragraph.** True.
  - `datedRE` is `regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}-(.+)$`)`, with no `(?s)` flag. In
    Go RE2, `.` excludes `\n` unless the `s` flag is set, and `$` without `m` anchors at the end of
    text, so a capture holding a newline cannot match.
  - `inDomain` admits `\n`, so the same name maps when active.
  - `TestSyncIntentSDDQuotesHostileUnits` exercises the archived case (`archive-name-invalid`).
  - `Resolve`'s parent check uses the same `datedRE`, so it behaves consistently.
- **§7 `non-regular-file` row.** It matches `provider.go`:
  - a `specs`/`specs/<cap>` that is not a tree is a unit only when `!Regular`;
  - a `spec.md` that is not regular (a symlink, gitlink or tree) is a unit;
  - a non-tree root or archive is a state skip.
- **§9 rows.**
  - **35:** true; branches step 14 `stale` is the same tree as step 11, which carries `name-reuse`.
  - **47:** true. `core.Event.Validate` bounds the payload with `>`, `store.Append` calls it, and
    `NewEvent` calls it.
  - **54:** true.
  - **56:** true as code (both `verify` sites pass `"all"`), but partly unpinned (F1).
  - **59:** true.
  - **60:** true.
  - **61:** true.
  - **65:** true.
  - Every "Also proven by" test named exists. The cross-package names are written in prose, which
    `invariant-lint` accepts (exit 0).
- **§12 INT-INV-4.** True, as shown in section 1.

## 6. Probe check

- **Pins and vectors.** I ran `history_probe.py --pins` and `--vectors` against a scratch **copy** of
  `testdata/`. They regenerated `expected-obligations.json` and `id-vectors.json` byte-identically.
  `diff -r` of the whole copy against the repository's `testdata/` is empty, and nothing in the
  repository was written.
- **Full run.** The full probe on the clone exits 0 with an empty stderr.
- **Old counts unchanged.** Moving the root-file count from object type to mode leaves it at
  `1393 ['IMPLEMENTATION_ORDER.md', 'skills-only-migration.napkin.md']`, and the under-`specs/`
  count stays at 1. Both are as §0 and `notes/state.md` state.
- **New counts.** AT `specs` is 0, and symlinks or gitlinks is 0.

## 7. Law walk

- **Law 1:** the four pin refreshes are produced by the scripts named in their `@generated` markers,
  and all three `--check`s exit 0 at HEAD. No generated file was hand-edited.
- **Law 2:**
  - The SPEC and draft 10 are byte-identical, the existing paired copy.
  - The new INV-35 helpers reuse `hostile()`, `hostileSkips` and the pinned
    `record-id-collision.json`, instead of re-listing them.
  - One drift: the Task-8 addendum's description of later commits (F3).
- **Law 6:** every behaviour change (`printableUnit`) and every new proof landed in the same commit as
  its SPEC row (SDD-INV-60, 47, 59, 61, 35). `kernel/internal/cli/SPEC.md` does not own skip
  printing, and never did, so SDD-INV-60 remains its home.
- **Law 7:** no `LESSON:` was added in range. The state note's "Gotcha learned" is about
  cross-package citations in `invariant-lint`, and is not a journal LESSON. Not a violation, but it
  is uncompiled if it recurs.
- **Law 9:**
  - The author self-calibrated (`notes/state.md`: "fails under its own hand mutant"). This verdict
    is the independent calibration.
  - Every new test's claimed mutant dies, and each has a surviving neutral control.
  - Two key-variable gaps survive: F1 and F2.
- **Law 10:** `notes/state.md` was last touched at `b50f61b`, 2 commits before HEAD, within the
  default of 3. Its Resume section names this re-check as the next action.

## Findings

### r2-m1-lowfix-recheck-20261118-F1 (LOW): SDD-INV-56's "never `sdd`" is proven only where `sdd` cannot be constructed

**Broken property:** SDD-INV-56 ("only `sync intent sdd` constructs the provider"; "`all` … never
`sdd`"; kills "adding `sdd` to `all`"), and Law 9's calibration corollary. The fixture holds the key
variable constant.

**Evidence:**
- **The fixtures.** `TestAllSelectionNeverConstructsSDD` runs `intentProviders(…, "all", …)` only in an
  empty temporary directory. `TestIntentAllExcludesSDD` runs `sync intent all` only on a shallow
  clone, which `gitcmd.New` refuses. My probe of an unconditional mutant panicked with "gitcmd:
  repository history is incomplete".
- **The survivor.** A mutant that appends the `sdd` provider to `all` whenever `gitcmd.New`,
  `NewGitReader` and `New` succeed survives all of these:
  - `TestAllSelectionNeverConstructsSDD`;
  - the sdd package's three CLI tests;
  - the whole `internal/cli` package (`go test -overlay`, rc 0).

  Under that mutant, `sync intent all`, `sync all`, `verify` and delivery-enforce walk full history
  on every ordinary repository, which is the harm SDD-INV-56 exists to prevent.
- **The verify call sites.** Changing `verify`'s initial intent sync (and, separately, both of its
  intent syncs) from `"all"` to `"sdd"` also survives the whole `internal/cli` package. The row's
  "verify reaches intent sync only through `syncIntentOnStore(…, "all")`" is true by reading
  (`cli.go` at HEAD, lines 902 and 921), but no test pins it.
- **Why LOW.** The code at HEAD is correct, and the unconditional add-to-`all` mutant is killed. This
  is a test-adequacy gap, not a behaviour defect.

### r2-m1-lowfix-recheck-20261118-F2 (LOW): SDD-INV-35's "every other unit emitted" is not discriminated by its proving test

**Broken property:** SDD-INV-35 ("… leaves `Revisions` without error **and every other unit
emitted**"). The fixture holds the key variable constant.

**Evidence:**
- **The assertion.** `TestHistoricalFailClosedIsRecoverable` checks emission only through
  `revs[ciKind+"openspec/changes/good"]`.
- **Why it cannot fail.** Every variant commits `good` at the clean first commit, before any defect
  state. So `good` is emitted whether or not a defect state emits its own clean units.
- **The survivor.** A mutant that drops every revision of any state carrying a Skip survives this
  test. The whole package kills it (`sdd_test.go:623` "owner change did not change the digest",
  `:679`).
- **Why LOW.** The fail-closed half (no error at a non-tip, a tip, or HEAD) is now fully
  discriminated: five HEAD-fatal mutants die, and the pre-rework test survives one of them.

### r2-m1-lowfix-recheck-20261118-F3 (LOW): the Task-8 admission addendum's account of later commits is no longer complete

**Broken property:** Law 2 (a statement about the admitted package's history that no longer matches
the repository).

**Evidence:**
- **The claim.** `docs/plans/p6-task8-package-acceptance-SPEC.md` (the addendum for the 18th
  package) says: production `.go` byte-identical to the reviewed tree, "later commits add only
  `mutation_kill_test.go` and `survivors_test.go`".
- **What actually changed.** Since then, `5aafb55`, `b7bf296` and `04f7051` have added
  `recheck_kill_test.go` and changed `sdd_test.go`, `SPEC.md` and `testdata/derive/history_probe.py`.
- **The tree object.** The package tree object `066f8b01…` recorded there is no longer HEAD's tree
  (`b5165e82…`).
- **Why LOW.** The binding property, that production `.go` is byte-identical, still holds: all six
  blob ids are equal. A reader who checks the tree object at HEAD, or trusts the parenthetical,
  would be misled about what changed.

## Acceptance decision

**ACCEPTABLE.**
- **Prior findings.** Re-check F1-F5 are closed. Round-1 F8 is closed in seven of its eight bullets;
  INV-56 `verify` is partly closed (F1).
- **Production code.** The sdd production code is byte-identical to the admitted tree, so the
  18th-package admission still binds.
- **cli change.** The `printableUnit` change is correct on every Unicode scalar value and every
  invalid byte, and introduces no regression.
- **Calibration.** Every new or reworked test kills its claimed mutant and survives its neutral
  control.
- **Findings.** The three LOWs are two key-variable test gaps over code I confirmed directly, and one
  stale descriptive sentence. None blocks.
