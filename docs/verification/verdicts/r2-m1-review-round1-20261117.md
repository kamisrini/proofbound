---
schema: vera.verdict.v1
verdict_id: r2-m1-review-round1-20261117
status: NEEDS_WORK
reviewed_commit: 6221a8c71136f541e4f31da0cc6b604924874540
findings:
  - finding_id: r2-m1-review-round1-20261117-F1
    severity: MED
  - finding_id: r2-m1-review-round1-20261117-F2
    severity: MED
  - finding_id: r2-m1-review-round1-20261117-F3
    severity: MED
  - finding_id: r2-m1-review-round1-20261117-F4
    severity: MED
  - finding_id: r2-m1-review-round1-20261117-F5
    severity: LOW
  - finding_id: r2-m1-review-round1-20261117-F6
    severity: LOW
  - finding_id: r2-m1-review-round1-20261117-F7
    severity: LOW
  - finding_id: r2-m1-review-round1-20261117-F8
    severity: LOW
  - finding_id: r2-m1-review-round1-20261117-F9
    severity: LOW
  - finding_id: r2-m1-review-round1-20261117-F10
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m1-review-round1-20261117.md
artifact_sha: 4eeb8b32e48a2b9a7e3c6694440a25e98b89c6a531f6d443f42e91fccf2b5022
---

# Release 2 M1 (`intent.sdd` provider, OpenSpec kit): foundational-tier non-author review, round 1

## Provenance

- **Under review:** commits `61aecab^..6221a8c`, HEAD `6221a8c71136f541e4f31da0cc6b604924874540`
  (16 commits, 40 files, +4381/-57). The tree did not move between dispatch and completion.
- **Contract:** `kernel/internal/connector/intent/sdd/SPEC.md`, which I confirmed is byte-identical
  (`cmp`) to `docs/plans/R2-sdd-provider-spec-draft10.md`. Scope: `CI-sdd-warranty-m1-6op6vl`
  (revision 0001) and `VD-sdd-m1-openspec-first-2026-11-17`.
- **Out of scope:** the untracked `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`.
- **Tier:** foundational, per the `/proofbound-review` procedure. One reviewer ran the four lenses
  (spec-truth, mutation adequacy, wedge/concurrency, build laws) in sequence, instead of
  dispatching four parallel reviewers. Adjudication rule: every finding below has a reproduction
  in `<scratch>`, or quotes the contract text it relies on; I dropped candidates that had neither.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m1r1`, a **non-author**. It wrote
  none of the work under review and none of the M0 rounds, so this verdict counts under Law 9.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent. No repository
  file was written except this verdict. I ran nothing that commits, pushes, or writes `notes/`.
- **Evidence (all from my own commands):**
  - Toolchain: Go 1.27.4, golangci-lint from `$HOME/go/bin`.
  - Scratch copy of `kernel/` at `<scratch>/kcopy/`. Every mutant went through a copy of
    `scripts/mutant-apply.sh` from `<scratch>/mut/run.sh`. Results: `<scratch>/mut/results.tsv`. Logs:
    `<scratch>/mut/logs/`.
  - Real-history harness: `<scratch>/overlay/zz_realhistory_test.go`, run with `go test -overlay`
    against the repository's own package. Output: `<scratch>/real-history.txt`.
  - Probes: `<scratch>/overlay/zz_probes_test.go`. Output: `<scratch>/probes.txt`.
  - Tests: `go test ./...` and the integration run, output in `<scratch>/integration.txt`.
  - A baseline export of `61aecab^` at `<scratch>/base/`, used to show a pre-existing integration
    failure is not M1's.
- **Redactions and tokens:**
  - No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m1r1/` in the session scratchpad. It
    occurs 21 times in this verdict, including this line.
  - The OpenSpec clone is named only as "the author's clone" (`f085c850`).

## M1 entry conditions (re-check of draft 9) re-verified

I checked each against draft 10's text and against the code, not against §0's answer table.

- **F1 (symlink/gitlink inside a change directory): CLOSED for symlinks and gitlinks.**
  - §4.1 fixes the layout types.
  - `provider.go` skips a non-tree `specs` or `specs/<cap>`, and a non-regular `spec.md`, as
    `non-regular-file` record units. A symlinked `design.md` and a nested gitlink enter the change
    manifest as `<mode> <oid>  <path>` (`contentLine`).
  - Pinned by `TestNonRegularFilesAreUnits` (the `designlink` change maps).
  - **New problem:** the same answer's clause "a **blob** occupies `specs/<cap>`" reaches real history
    (F1 below).
- **F2 (parent archive check by name only): CLOSED in code, but untested.**
  - `Resolve` requires `isChangeEntryMode(e.Mode)` on the parent's entry.
  - Deleting that conjunct survives the whole package (F8).
- **F3 (tree at a delta path): CLOSED.** It is a `non-regular-file` record unit and its change is
  `cascade`, pinned by `tree-delta`.
- **F4 (profile digest from an invalid profile): CLOSED by construction.**
  - For an invalid profile, `GitReader.Read` returns only the profile entry, and `StateKey` is the
    profile line alone.
  - So roots are never read, and one invalid profile oid is one state and one Skip.
  - My mutant that adds a root line to that digest is therefore equivalent, and I did not count it.
- **F5 (INV-65 cache-by-key kill): CLOSED.** The kill is gone from INV-65's row. The `branches` 08
  row in `expected-resolve.json` carries it.
- **F6 (drift): CLOSED.**
  - §7's `archive-name-invalid` row and §2's `Skip.Unit` comment match §6.
  - INV-62 is retired and re-issued as INV-67.
  - `fixture_test.go` now says "draft 10".

## Conformance, section by section

- **§2 interface lock.** `Entry`, `Artifact`, `Reader` (5 methods), `New`, `Name`, `Revisions`,
  `Skipped`, `Resolve` and the three sentinels match, sentinel texts included. Two things differ:
  - The package also exports `UnmappableError` (the only carrier of `ErrUnmappable:<cause>`),
    `GitReader`/`NewGitReader`, `ProfilePath` and `Kit`, which the lock does not list (F10).
  - `Read` returns ancestor trees of the root (F7).
- **§2.1 profile.**
  - These rules are implemented: the exact pair, direct-child archive, literal grammar, `*`/`**`
    patterns, non-blank and NUL-free values (raw and decoded), `kit`/`mode`/`status` enums, and
    null/duplicate/trailing-data refusal.
  - **Deviation:** key matching is case-insensitive (F5).
- **§3 reader contract.**
  - `CommitSHAs` is the single home of the walk. It has the same `refusePartial`/`validateRefs` and
    `rev-list` as `Commits` (Law 2 holds).
  - Every git call goes through `run` (`--no-replace-objects`, the same `-c` options, a scrubbed
    environment). `TreeEntries` carries `--literal-pathspecs` and never dereferences `<c>:<path>`.
  - The state key is as specified. For an unparseable profile it is the profile line alone.
- **§4.1-4.3.**
  - Record units, the layout types, the "zero entries and no unpaired line" non-unit, slug, suffix,
    ids and both manifests match the text.
  - The parser matches `parseDeltaSpec`, `splitTopLevelSections`, `getSectionsCaseInsensitive`,
    `parseRequirementBlocksFromSection`, `parseRemovedNames` and `parseRenamedPairs` at `f085c850`.
    I read them side by side:
    - `FROM` is checked before `TO`;
    - a REMOVED header precedes the bullet, and a bullet ends an open block;
    - a one-space `## ` line ends a block but starts no section;
    - the closing-sequence strip and `trim`;
    - ASCII-only title folding (no Unicode lower-casing reaches an ASCII title letter except
      U+212A, and no title contains `k`).
  - One divergence from OpenSpec at `f085c850`: its `spec-discovery.ts` ignores `specs/*.md` that has
    no delta sections, while §4.1 makes such a file a unit (F1).
- **§5 `changeOutcome`.** The precedence is implemented in rule order 1-8 in `mapState`.
  - Change entries are trees, symlinks and gitlinks only.
  - Name-reuse records are still emitted.
- **§6 `Resolve`.** Lines 1-7 are implemented in order: the typed parent check, a root resolving, and
  `ErrNoChangeIntent` after `ErrArchivedEarlier`. One deviation: the archive's own name at a
  state-level-cause state (F6).
- **§7.**
  - The 15 M1-reachable causes are all produced.
  - Skips are de-duplicated by `(unit, cause, digest)`. The digest follows the per-kind rule (the
    change digest also hashes the change entry's own tree line, which is a function of the same
    contents).
  - `Commit` is the first carrying commit.
  - A `ValidateRevision` failure after the data checks is returned, not skipped.

## Calibration (mutation evidence)

Harness: `<scratch>/mut/run.sh`. Every substitution went through `mutant-apply.sh`, and each mutant
was compile-checked before its named test ran. An uncompilable mutant is `INVALID` and is never
counted.

| Control | Mutation | Must | Result |
|---|---|---|---|
| Positive | `obligationID` drops the `-` after the section code | die | **killed** (`TestRecordIDsAreDeterministic`) |
| Neutral | reword the "provider is not initialized" error, whole package | survive | **survived** |
| Lethal | an injected `t.Fatal` test | die | **killed** |

**Invariant mutants: 32 counted, 27 killed, 5 survived.** One more was `INVALID` and re-issued
(INV-39 length), and one was equivalent and not counted (entry-condition F4).

Killed:
- INV-31 HEAD-only walk;
- INV-33 parent full-state read;
- INV-35 HEAD-fatal and batch-fatal;
- INV-39 closer length ignored, and info-string backtick ignored;
- INV-43 unpaired-only treated as zero entries;
- INV-46 de-dup without the digest;
- INV-47 oversize emitted;
- INV-48 first-wins collision;
- INV-49 defect turned into a Skip;
- INV-53 NUL tolerated, and a decoded `\u0000` owner;
- INV-54 symlinked delta followed, and active-only rule 4;
- INV-59 plain files as changes, `ErrNoChangeIntent` before `cascade`, and `ErrArchivedEarlier` on
  reuse;
- INV-61 dereferencing root, and over-broad `invalid-utf8`;
- INV-66 any object type;
- INT-INV-4 bound removed;
- INV-56 `sdd` in `all`;
- INV-60 silent skips;
- INV-52 unregistered route (integration: `unsupported event intent.sdd/requirement.recorded`);
- INV-52 NUL reaching `TEXT` (integration: `invalid byte sequence for encoding "UTF8": 0x00`);
- G-INV-26 fabricated root.

Survived (each an undeclared survivor; see F2 and F8):
1. **INV-59:** a state-level `non-regular-file` state resolved as `ErrUnknownChange`. Survives the
   whole package.
2. **INV-59:** a reader error returned as the `ErrUnknownChange` sentinel. Survives the whole
   package.
3. **INV-65(a):** an oid-only state key.
4. **INV-54:** the non-regular manifest line replaced by a content-style hash line.
5. **Entry-condition F2:** the parent-archive type check removed. Survives the whole package.

## Real-history run (the author's clone, `f085c850`)

The clone carries no profile, so the harness wraps `GitReader`. It injects the canonical M1 profile
(`testProfile`) as the profile entry at every commit. Every other read goes through `GitReader`'s own
`own`/`artifact`/`TreeEntries` path.

- **`Revisions`:** no error. It covered **1624** commits and **338** distinct states (equal to §0's
  338), and returned **810** revisions (481 requirement, 329 change intent).
  - Every revision passes `ValidateRevision`.
  - Runtime **42.6 s**.
- **`Sync`** through the intent connector (validation, ordering, a memory appender): listed 810,
  appended 810, no error, 34.3 s.
- **Skips: 5.**

  | Cause | Unit | Plausible? |
  |---|---|---|
  | `name-reuse` | `change adopt-verb-noun-cli-structure` | expected (§0's one dual-presence name) |
  | `obligation-collision` | `requirement openspec/changes/add-non-interactive-init-options/specs/cli-init/spec.md` | plausible: the delta repeats its ADDED and MODIFIED sections with the same requirement names |
  | `cascade` | `change add-non-interactive-init-options` | plausible, follows from the row above |
  | `non-regular-file` | `requirement openspec/changes/archive/2026-01-30-opencode-command-references/specs/no-changes.md` | **not plausible** (F1) |
  | `cascade` | `change opencode-command-references` | **not plausible** (F1) |

  - Nothing for `IMPLEMENTATION_ORDER.md` or `skills-only-migration.napkin.md` (SDD-INV-66 holds on
    real history).
  - §0 predicted only the first of these. The third and fifth follow from the second and fourth.
- **`Resolve` sweep:** a full sweep over all 1624 commits did not finish within the 1-hour test
  timeout, so I sampled.
  - **Sample:** 439 commits, namely the first commit of every distinct state plus every merge. At each,
    I resolved every in-domain name present, plus `nope`, `archive` and `IMPLEMENTATION_ORDER.md`.
  - **Calls:** 22587.
  - **Outcomes:**

    | Outcome | Count |
    |---|---|
    | `ErrArchivedEarlier` | 17521 |
    | revision | 2955 |
    | `ErrUnknownChange` | 1317 |
    | `ErrNoChangeIntent` | 678 |
    | `ErrUnmappable:cascade` | 102 |
    | `ErrUnmappable:name-reuse` | 14 |

    The 1317 `ErrUnknownChange` results are exactly the 439 × 3 probe names, so no name that was
    present resolved as unknown.
  - **Results:**
    - zero reader errors;
    - every successful `Resolve` was a member of `Revisions()` (SDD-INV-31 on real history);
    - `Resolve`'s cause matched `changeOutcome`'s cause at every skipped change (0 disagreements).
  - **Runtime:** 29.5 min, about 78 ms per call, because every call re-reads the commit's state.
    That is a cost note for M3's per-claim `Resolve`, not a finding.
  - **Attribution:** the 102 `cascade` outcomes are F1's change and the `obligation-collision` change.

## Findings

### r2-m1-review-round1-20261117-F1 (MED): the §4.1 layout rule makes a real, legitimate spec-less OpenSpec change unmappable at every state since 2026-01-30, HEAD included, and contradicts SDD-INV-61

**Broken property:**
- SDD-INV-61's second clause ("files a unit neither parses nor carries are never a cause").
- §7's "Other files in a change directory … are never a cause".
- SDD-INV-35's intent (no withheld trust from conforming tenant content).
- §0's claim to measure every consequence on the reference corpus.

**Evidence:**
- **What is in the corpus.** At `697738bc` and at HEAD of the author's clone,
  `openspec/changes/archive/2026-01-30-opencode-command-references/specs/no-changes.md` is a
  `100644` blob directly under `specs/`. It is the only such file in the whole history (`git log
  --all -- ':(glob)openspec/changes/**/specs/*.md'`). Its text reads "No spec files are needed for
  this implementation-only fix."
- **What OpenSpec does with it.** OpenSpec's `spec-discovery.ts` at `f085c850` reads only
  `…/spec.md` under a capability directory. It ignores any other `.md` under `specs/` that has no
  delta section. So OpenSpec treats the change as spec-less.
- **What the contract says.** §4.1 ("If a symlink, gitlink, or **blob** occupies `specs` or
  `specs/<cap>` … skipped with `non-regular-file`") makes the file a record unit, and §5 rule 6 then
  makes the change `cascade`.
- **What the code does.** The provider does exactly this. A regular file is reported as
  `non-regular-file`, so the cause name contradicts §7's own row ("Nothing else is this cause").
- **The outcome.** At every later state, `Resolve(c, "opencode-command-references")` is
  `ErrUnmappable:cascade` instead of `ErrNoChangeIntent`. The archive is permanent, so this never
  clears.
- **Reproduction:**
  - the real-history table above;
  - `<scratch>/probes.txt` P5: a synthesized `fix/specs/no-changes.md` gives
    `requirement …/specs/no-changes.md|non-regular-file` and `change fix|cascade`, and `Resolve` gives
    `ErrUnmappable:cascade`.
- **Origin.** Draft 10 introduced this to answer re-check F1/F3. It was not re-measured on the
  corpus, and §0 has no row for it.

### r2-m1-review-round1-20261117-F2 (MED): SDD-INV-59's proving test omits three of its row's four obligations, and two calibration mutants survive the whole package

**Broken property:**
- §9 ("each proving test names the wrong implementation it kills").
- SDD-INV-59 ("any other error comes only from the reader").
- Law 9's calibration corollary.

**Evidence:**
- **What the row requires.** SDD-INV-59's row promises:
  - rows "**Synthesized by the M1 test**" (absent profile; symlinked root or blob archive →
    `ErrUnmappable:non-regular-file`; symlinked change directory; symlink-only delta → `cascade`);
  - "Over every listed commit and every name, `Skipped()` and `Resolve` agree on the cause";
  - "A reader failing on purpose yields none of the sentinels".
- **What the test does.** `TestResolveIsTotal` (`sdd_test.go:172-202`) only replays
  `expected-resolve.json`. No test in the package synthesizes those `Resolve` rows, runs an agreement
  sweep, or uses a failing reader.
- **Mutant 1.** Before the `stateSkip` return in `Resolve`, insert
  `if st.stateSkip == "non-regular-file" { return …, ErrUnknownChange }`. It **survives all tests in
  the package**.
- **Mutant 2.** Replace `Resolve`'s read-error return with `ErrUnknownChange`. It **survives all
  tests** (`<scratch>/mut/results.tsv`).
- **Why it matters.** M3 relies on exactly this totality, to tell a bad claim from a reader
  failure (§6, §14).

### r2-m1-review-round1-20261117-F3 (MED): SDD-INV-52, an acceptance-boundary test, covers 11 of the 15 causes its row names

**Broken property:** SDD-INV-52's row, and the CI acceptance boundary ("including a sync-then-rebuild
over OpenSpec-shaped history (SDD-INV-52)").

**Evidence:**
- **What the row requires.** A synthesized history with one unit for **each** of 15 causes.
- **What the test uses.** `TestSDDEventsProject` uses `hostile()` (`sdd_test.go:628-640`), which
  carries:
  - `nul-byte`, `invalid-utf8`, `unpaired-rename`, `empty-name-part`, `obligation-collision`;
  - `name-reuse`, `slug-too-short`, `archive-name-invalid`, `non-regular-file`;
  - `payload-too-large` and `cascade`.
- **What is missing:** `profile-invalid`, `owner-uncovered`, `sponsor-uncovered` and
  `record-id-collision`.
- **What it asserts.** "Rows exist for the clean units" is asserted only as `count(*) > 0` per
  view, not per clean unit.
- **What holds.** The two kills the row names do fire: an unregistered route, and a NUL reaching
  `TEXT`. Both are integration mutants, both killed.

### r2-m1-review-round1-20261117-F4 (MED): `make check` will be red: `prescription-lint` fails on `sdd/SPEC.md`

**Broken property:** the CI acceptance boundary ("`make check` is green"), and the intent SPEC §1
condition for `sdd` becoming a production provider.

**Evidence:**
- **The failure.** `make -s prescription-lint` exits 2:
  `kernel/internal/connector/intent/sdd/SPEC.md prescribes absent literal --all` and `… --full-history`.
- **The trigger.** §0's evidence row quotes `` `--all --full-history` `` in backticks, and the lint
  requires every flag in such a span to appear as a Go string literal in the package.
- **Why the fix is not free.** The SPEC must remain a byte copy of draft 10. Any fix therefore
  changes both files, or the lint's rules.
- **No other gate reproduced as red.** I ran individually, all exit 0:
  - `invariant-lint`, `invariant-table-lint`, `spec-numbering-lint`, `link-lint`,
    `law-citation-lint`, `cleanroom-lint`;
  - `identity-inventory`, `p6-task1-close`, `figure-provenance`, `lesson-recurrence`, `index-check`,
    `intent-trailer-gate`;
  - `p6-census.sh`, `p6-route-matrix.sh`, `p6-task2-close.sh` and `p6-artifact-integrity.sh`, each
    with `--check`;
  - `golangci-lint` (0 issues), and `go test ./...` (all packages ok).

  I did not run `make check` itself, which the owner is holding.

### r2-m1-review-round1-20261117-F5 (LOW): the profile decoder matches keys case-insensitively, so §2.1's "no unknown or duplicate fields" and SDD-INV-64 can be bypassed

**Broken property:** SDD-INV-40 ("strict JSON"; its named kill is "a lenient decoder") and
SDD-INV-64.

**Evidence:**
- **The cause.** `encoding/json` folds key case, and `strictShape`'s duplicate check is
  case-sensitive.
- **What parses as valid** (`<scratch>/probes.txt` P1):
  - `"KIT": "openspec"`;
  - `"OWNERS": […]`;
  - `"kit": "speckit", "Kit": "openspec"`, a speckit-declaring profile that maps as OpenSpec.
- **Why no test catches it.** `TestProfileIsStrict` tests same-case duplicates only.

### r2-m1-review-round1-20261117-F6 (LOW): `Resolve` checks the archive's own name after the state-level cause, against §6's line order

**Broken property:** SDD-INV-59 (§6 line 1, "outside the name domain → `ErrUnknownChange`", precedes
line 2).

**Evidence:**
- **The code.** `Resolve` tests `name == st.archiveBase` only after returning a state-level
  `ErrUnmappable`.
- **The probe.** At a state whose `changes.archive` is a blob, the profile is valid, so the archive
  path is defined. `Resolve(c, "archive")` gives `ErrUnmappable:non-regular-file`, where §6 gives
  `ErrUnknownChange` (`<scratch>/probes.txt` P3).
- **Scope.** Contrived, and no real-history state has a state-level cause.

### r2-m1-review-round1-20261117-F7 (LOW): `TreeEntries(…, children=true, recursive=true)` also returns the path's ancestor trees, so `Read` returns entries outside the SDD roots

**Broken property:**
- §2's `Read` contract ("every tree entry … under the profile's SDD roots, plus the profile").
- G-INV-24's description in the gitcmd SPEC ("returns the entries under path").

**Evidence:**
- **Why it happens.** `ls-tree -r -t … -- openspec/changes/` also prints `openspec`, and the filter
  drops only `name == path`. I reproduced this with plain git in `<scratch>/t1`.
- **What it does.** `GitReader.Read` returns `openspec` (`040000`) (`<scratch>/probes.txt` P4).
- **Scope.** That entry's oid moves with main specs, which SDD-INV-55 says never affect the provider.
  Today `mapState` never reads an ancestor, so there is no observable effect.

### r2-m1-review-round1-20261117-F8 (LOW): further §9 rows whose named kill cannot fire, or whose test is narrower than the row

**Broken property:** §9, and Law 9's calibration corollary.

**Evidence** (mutants in `<scratch>/mut/results.tsv`):
- **INV-65(a): an oid-only key survives.** An invalid profile's key is one line and a valid one's
  is three, so the symlink-versus-regular keys differ under any key format.
- **INV-54: "hashing a symlink target as content" cannot be detected.** The `<mode> <oid>` manifest
  line is never pinned, so the mutant survives.
- **Entry-condition F2: the parent-archive type check is unpinned.** The mutant survives the whole
  package.
- **INV-61:**
  - The test asserts exactly one `profile|non-regular-file` through a `unit|cause` set. That set
    cannot count the two Skips §7 says a symlinked root and a later blob archive produce.
  - The row's gitlink root is absent.
- **INV-35:**
  - The hostile set lacks `owner-uncovered`, `sponsor-uncovered` and `record-id-collision`.
  - The "head" variant places only `profile-invalid` at HEAD.
  - The row's `branches` step 14 and `head-on-leftover` are not used by this test. They are pinned
    elsewhere, through INV-46.
- **INV-47:** the row's "just-under-limit payload" is absent.
- **INV-56:** `proofbound verify` is not exercised.
- **INT-INV-4:** draft 10 §12's "rebuild-replay row: an event ledgered before M1 still re-validates
  after it" became an exact-bound validation row, which is not a replay of a ledgered event.

### r2-m1-review-round1-20261117-F9 (LOW): skip output carries raw tenant bytes and has no stable order

**Broken property:** SDD-INV-60 ("one line per Skip"). Draft 10 also never fixes a regex dialect for
§6's archive pattern.

**Evidence:**
- **Newlines in names.** Go's `.` excludes `\n`. So an archive entry `2026-01-01-x\ny` becomes
  `archive-name-invalid`, while the same in-domain name maps when active (`<scratch>/probes.txt` P2).
  Its `Unit` contains a newline, and `writeSkips` prints it unescaped across two lines.
- **Spoofable lines.** A tenant can likewise emit lines that look like `skip …` or `listed=…`.
- **Unstable order.** `Revisions` and `Skipped()` iterate a Go map, so their order varies run to run:
  6 distinct orders in 8 runs (probe P6). The CLI prints skips in that order. Payload bytes are
  stable, so SDD-INV-12 holds.

### r2-m1-review-round1-20261117-F10 (LOW): scope and law drift

- **§12's census site was not done.** §12 lists `scripts/p6-census.sh:233-235` ("add the
  `intent.sdd` rows") as an M1 site, and it was left untouched on purpose, to keep the C6 window.
  - §12 was not amended to say so.
  - `p6-census.sh --check` does not require C5-053..060. Only `p6-route-matrix.sh`'s count of 60
    guards them.
- **The code landed before the SPEC.**
  - The provider landed in `11f1916` and `e110b6c`, before `sdd/SPEC.md` landed in `fbd722f`. The CI
    says the SPEC lands "in the same commit as the code and tests it cites".
  - Law 6's substance holds, because draft 10 existed first under `docs/plans/`.
- **The interface lock is incomplete.** §2's lock omits the exported `UnmappableError` (the only
  carrier of `ErrUnmappable:<cause>`, which M3 will consume), `GitReader`/`NewGitReader`,
  `ProfilePath` and `Kit`.
- **The verify cells cite a kind-blind test.** The route-matrix "verify consume" cells C5-054 and
  C5-058 cite `TestVerifyChecksEveryStageError`, a source-text count that is blind to source and
  kind. This is a pre-existing pattern, and nothing in it exercises `intent.sdd`.

## What holds

- **Scope.**
  - Every governed commit (`intent-applicability.sh` → `applicable=true`) carries
    `Intent: CI-sdd-warranty-m1-6op6vl`. Those are 61aecab, b92adc4, 11f1916, 5d1caef, e110b6c,
    ddeea93, a094e4d and fbd722f.
  - None of the exclusions changed: `connector.go:198`, the `cli.go` resolver switch (now line 251),
    the `projections.go` `IntentRef` check (now line 727), and `intent-delivery-applicability.sh`.
    `all` still means `records` + `specdir`, and the `INV56-sdd-in-all` mutant is killed. No Spec Kit
    mapping was added, and no BR status changed.
- **Law 1.** Every `@generated` output passes its `--check`. `id-vectors.json` and
  `expected-obligations.json` were not touched.
- **Law 2.** `CommitSHAs` is the single home of the walk, and the SPEC is cited, not paraphrased.
- **Law 8.** No new dependency.
- **Wedge classes.** On fixtures, synthesized hostile bytes and real history, no tenant byte, path,
  object type or profile shape made `Revisions` fail. The shapes were:
  - CRLF and BOM (parser);
  - `\u0000` JSON escapes (decoded check, killed mutant);
  - NUL and invalid UTF-8 in deltas or names;
  - symlinks, gitlinks, blob roots and archives, a tree profile, and newline names.

  Payload strings that reach `TEXT`/`JSON` columns are UTF-8 and NUL-free by construction:
  statement, name fallback, `TO` line, owner, sponsor, `artifact_path`, and the ids. An oversize
  payload is a Skip before append (INT-INV-4 bound, killed mutant).
- **Batch and replay safety.** Real history (810 revisions) passes the intent connector's `Sync`.
  The integration sync-then-rebuild passes (`TestSDDEventsProject`), and its two named kills fire.
- **Tests.**
  - `go test ./...` passes.
  - Integration with `DATABASE_URL` unset: the `sdd`, `twin` and `gates` packages pass.
  - `projections` fails 48 tests at setup with "ledger tables are append-only". That failure is
    pre-existing: the same failure reproduces on the `61aecab^` export, because the suite's
    embedded-store path `TRUNCATE`s an append-only ledger. It is not an M1 regression.

## Acceptance decision for M1

**NEEDS_WORK.** No HIGH finding. Four MEDs block M1 from counting as done under the CI's acceptance
boundary:
1. **F1.** The contract must give a regular file directly under `specs/` (and a regular file at a
   `specs/<cap>` position) a single outcome that agrees with SDD-INV-61, §7 and OpenSpec's reader. The
   real `opencode-command-references` shape must be measured in §0 and pinned by a test. This is a
   text change to draft 10 and `sdd/SPEC.md` together.
2. **F2.** SDD-INV-59's row must be proven as written, or narrowed through §9's retire-and-reissue
   rule. Both surviving mutants must die.
3. **F3.** SDD-INV-52's test must carry all 15 causes its row names, or the row must be narrowed the
   same way.
4. **F4.** `prescription-lint` must pass on `sdd/SPEC.md`, keeping it a byte copy of draft 10, and
   `make check` must be green once the hold lifts.

The LOW findings (F5-F10) may land in the same round. A non-author re-check of F1-F4, with the
survivors re-run, is sufficient. A full foundational round is not needed again unless the fix for F1
changes §4.1/§5 beyond that one shape.
