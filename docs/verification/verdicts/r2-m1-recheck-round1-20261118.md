---
schema: vera.verdict.v1
verdict_id: r2-m1-recheck-round1-20261118
status: ACCEPTABLE
reviewed_commit: 77cc6ba57aa1d6e42ee1bf02401a4f55d6a6173e
findings:
  - finding_id: r2-m1-recheck-round1-20261118-F1
    severity: LOW
  - finding_id: r2-m1-recheck-round1-20261118-F2
    severity: LOW
  - finding_id: r2-m1-recheck-round1-20261118-F3
    severity: LOW
  - finding_id: r2-m1-recheck-round1-20261118-F4
    severity: LOW
  - finding_id: r2-m1-recheck-round1-20261118-F5
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m1-recheck-round1-20261118.md
artifact_sha: 85fad17d8d176e5d35237c4318676b5352ce701bfb84ea57930823ad639094d0
---

# Release 2 M1 (`intent.sdd` provider): targeted non-author re-check of round 1 (F1-F4)

## Provenance

- **Under review:** the fixes `5b69da1` (answering `r2-m1-review-round1-20261117`) and the pin refresh
  `77cc6ba`. HEAD `77cc6ba57aa1d6e42ee1bf02401a4f55d6a6173e`. `841231a` (between the round-1 HEAD
  `6221a8c` and the fixes) adds only the round-1 verdict, its integrity row and `notes/state.md`. The
  tree did not move during the re-check.
- **Contract:** `kernel/internal/connector/intent/sdd/SPEC.md`. I confirmed with `cmp` that it is
  byte-identical to `docs/plans/R2-sdd-provider-spec-draft10.md` at HEAD.
- **Out of scope:** the untracked `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`.
- **Scope of this round:** a targeted re-check, as round 1's acceptance decision allows. It covers
  F1-F4, re-runs round 1's surviving calibration mutants, checks the LOW fixes (F5, F6, F7, F9, F10),
  confirms F8's status, and judges whether F1's fix stays within "that one shape".
- **Reviewer identity:** `dispatched-non-author-reviewer-20261118-m1rc1`, a **non-author**. It wrote
  none of the work under review, none of the fixes, and not the round-1 verdict, so this verdict
  counts under Law 9.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent. The only
  repository file written is this verdict. Nothing was committed or pushed, and `notes/` was not
  written. `make check` was not run, because the owner is holding it.
- **Evidence (all from my own commands):**
  - Toolchain: Go 1.27.4, golangci-lint from `$HOME/go/bin`.
  - Mutation harness: `<scratch>/mut/run.sh`, run over a `git archive` export of HEAD at
    `<scratch>/kcopy/`. Every substitution goes through that copy's `scripts/mutant-apply.sh`, and
    every mutant is compile-checked before its test runs. Results: `<scratch>/mut/results.tsv`. Logs:
    `<scratch>/mut/logs/`.
  - Real history: round 1's harness, re-run with `go test -overlay` against the repository's own
    package, plus one targeted test (`TestReviewerRecheckF1`). Output: `<scratch>/real-history.txt`.
  - Probes: `<scratch>/overlay/zz_probes_test.go` (round 1's probes plus RC1-RC3). Output:
    `<scratch>/probes.txt`.
  - `testdata/derive/history_probe.py` re-run on the clone. Output: `<scratch>/probe.txt`.
  - Gates and tests: `<scratch>/lint-*.txt`, `<scratch>/p6-*.txt`, `<scratch>/gotest.txt`,
    `<scratch>/golangci.txt`, `<scratch>/integration.txt`.
- **Redactions and tokens:**
  - No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m1rc1/` in the session scratchpad. It
    occurs 17 times in this verdict, including this line.
  - The OpenSpec clone is named only as "the author's clone" (`f085c850`).

## Round-1 MED findings

### Round-1 F1 (regular file under `specs/`): CLOSED

- **Contract.** §4.1 now gives only a symlink or gitlink at `specs` or `specs/<cap>` the
  `non-regular-file` outcome. A new bullet says a **regular file** at `specs`, or directly under
  `specs/`, "is ignored … never a record unit and never a cause". §5 and §6 are untouched.
- **Code.** `provider.go` adds `if !specs.Regular` and `if !capDir.Regular` guards around the two
  `addRecord(… "non-regular-file" …)` calls. `Regular` is `regularMode` (`100644`/`100755` only), so
  symlinks (`120000`) and gitlinks (`160000`) still produce the record unit.
- **Real history** (the author's clone, canonical profile injected at every commit, as in round 1):
  - Commits 1624, states 338, revisions 810 (481 requirement, 329 change intent): unchanged.
  - Skips fell from 5 to 3. The two removed are exactly F1's `non-regular-file
    …/specs/no-changes.md` and `cascade change opencode-command-references`. The remaining three
    (`name-reuse`, `obligation-collision` and its `cascade`) are unchanged.
  - `Sync`: listed 810, appended 810, no error.
  - `Resolve(697738bc, "opencode-command-references")` is now **`ErrNoChangeIntent`** (it was
    `ErrUnmappable:cascade`). At the clone's HEAD it is `ErrArchivedEarlier`, which is §6's correct
    outcome once the parent already holds the archive entry. Before the fix, `cascade` preceded
    archived-earlier at every later state.
- **Other types at the same positions still skip** (`<scratch>/probes.txt` RC1):
  - a symlink at `specs/<cap>` gives `requirement …/specs/link|non-regular-file` and `cascade`;
  - a symlinked `specs` gives `non-regular-file` and `cascade`;
  - a gitlink at `specs/<cap>` and a gitlink at `specs` each give `non-regular-file` and `cascade`.
  - A blob at `specs` gives `ErrNoChangeIntent`.
  - A `100755` file directly under `specs/` is ignored, and the sibling capability still emits.
- **§0 row against the probe.** `history_probe.py` prints `regular files directly under a change's
  specs/ …: 1 ['openspec/changes/archive/2026-01-30-opencode-command-references/specs/no-changes.md']`.
  That matches §0's new row. Independently:
  - `ls-tree -r` over every `rev-list --all` commit finds the same single `100644` path.
  - No non-tree `…/specs` entry exists anywhere in the history.
- **Pinned.** `TestNonRegularFilesAreUnits` gains a `notes-only` change. Mutant `F1-revert-cap` is
  killed, and both over-broad mutants (`F1-overbroad-cap` and `F1-overbroad-specs`, which ignore
  symlinks too) are killed. A regular file at `specs` itself is unpinned (F1 below).
- **Shape.** The fix stays within "that one shape", so no full round is needed:
  - §4.1 changes only the outcome for a regular file at the two layout positions, which round 1's
    remedy named.
  - §5's precedence, §6's lines and §7's cause list are unchanged.
  - Symlink and gitlink behaviour is unchanged and still killed.
  - Real history moves only in F1's two skips.

### Round-1 F2 (SDD-INV-59 proof): CLOSED

- **What changed.** `TestResolveIsTotal` gains a `corners` subtest with three parts:
  - The row's synthesized cases: absent profile → `ErrUnknownChange`; symlinked root and blob
    archive → `ErrUnmappable:non-regular-file`, with `archive` → `ErrUnknownChange` at both; a
    symlinked change directory → `non-regular-file`; a symlink-only delta → `cascade`; a clean
    sibling → revision.
  - An agreement sweep over every commit of the four replayed histories and the synthesized one.
    The sweep is real: it fails when it checks nothing, and the `V-change-cause-swapped` mutant dies.
  - A `failingReader`.
- **Both round-1 survivors now die, under the whole package:**

  | Mutant | Round 1 | Now | Kill line |
  |---|---|---|---|
  | `INV59-statelevel-nonregular-unknown` | survived | **killed** | `Resolve(x): got ErrUnknownChange want ErrUnmappable:non-regular-file` |
  | `INV59-reader-error-as-sentinel` | survived | **killed** | `a reader failure surfaced as a tenant outcome` |

- **Variants also die:**
  - a reader error returned as `ErrNoChangeIntent`, or as `ErrUnmappable:cascade`;
  - every state cause returned as `ErrUnknownChange`.
- **Limits** (F2 below): the failing reader fails only `Read`. The agreement sweep checks the
  skip → `Resolve` direction over `mapState`'s skips, not `Skipped()` itself.

### Round-1 F3 (SDD-INV-52 coverage): CLOSED

- **Coverage.** `TestSDDEventsProject` now builds three extra states on top of `hostile()`:
  - the pinned record-id collision pair;
  - a `profile-invalid` state of its own;
  - a narrowed profile that yields `owner-uncovered` and `sponsor-uncovered`.
- **Assertions.**
  - The test asserts each of the 15 causes in the row by name on the synthesized history's
    `Skipped`.
  - It replaces the `count(*) > 0` check with a per-unit query on `payload->>'artifact_path'`, for
    five clean units across both views.
- **Run.** `env -u DATABASE_URL go test -tags integration -count=1
  ./internal/connector/intent/sdd/` passes (`<scratch>/integration.txt`).
- **Kill re-confirmed.** The unregistered-route mutant (`F3-INV52-no-route`) is still killed:
  `projections: unsupported event intent.sdd/requirement.recorded`.

### Round-1 F4 (`prescription-lint`): CLOSED

- `make -s prescription-lint` exits 0.
- §0's evidence row now names the walk options in words. SPEC and draft 10 stay byte-identical.

## Round-1 LOW findings

| Round-1 finding | Status | Evidence |
|---|---|---|
| F5 case-insensitive profile keys | **Closed** | `strictShape` accepts only §2.1's 14 keys, in exact case, at every depth. The key set matches the decoder's `json:` tags. Probes reject `KIT`, `kit`+`Kit` and `OWNERS` (P1), and nested `Pattern`, `Mode`, and a `kit` key inside an owner entry (RC2). `F5-revert` is killed by the new `case-twin keys` case |
| F6 `Resolve("archive")` order | **Closed** | `archive` is checked against `st.archiveBase` before `stateSkip`. P3 now gives `ErrUnknownChange` at a blob-archive state. `F6-revert` is killed |
| F7 ancestors in children listings | **Closed** | Children mode now keeps only `name` with prefix `path+"/"`. `TreeEntries` already refuses `""`, a leading `/` and a trailing `/`. P4 no longer reports `openspec`. `F7-revert` is killed by the new nested-path assertion. `sdd/reader.go` is the only caller with `children=true` |
| F8 §9 rows narrower than their kills | **Not addressed (confirmed)** | `INV65-oid-only`, `INV54-hash-symlink-target` and `F2-untyped-parent-archive` all still **survive the whole package**. The other F8 items are unchanged (INV-61 count/gitlink root, INV-35 hostile set and head variant, INV-47 just-under-limit, INV-56 `verify`, INT-INV-4 replay row). See F5 below |
| F9 raw tenant bytes / unstable order | **Mostly closed** | `printableUnit` quotes invalid UTF-8 and C0/DEL bytes. `TestSyncIntentSDDQuotesHostileUnits` proves that a newline name cannot forge a `skip` line, and `F9-unquoted` is killed. `mapState` now sorts skips and revisions, and probes show 1 order over 8 `Revisions` runs (P6) and 1 order over 12 `Skipped()` runs (RC3). Residuals are in F3 below |
| F10 scope and law drift | **Closed** | §2's lock now lists `UnmappableError`, `ProfilePath`, `Kit`, `GitReader` and `NewGitReader`, which with the earlier names are every exported package identifier. §12 records the census rows as really done (C3-019, C4-013 and C5-053..060 exist in `docs/plans/p6-census-rows.tsv`; `scripts/p6-census.sh`'s last commit is still `13503c7`) and the `p6-task2-close.sh` universe entry (`intent/sdd`, C4-013 synthetic, present). The "code before SPEC" point was historical, and the kind-blind verify cells were pre-existing; neither was claimed |

## Calibration (mutation evidence)

Harness `<scratch>/mut/run.sh`. Each substitution goes through `mutant-apply.sh`, and each mutant is
compile-checked before its test runs. An `INVALID` mutant is never counted, and none occurred.

| Control | Mutation | Must | Result |
|---|---|---|---|
| Positive | `obligationID` drops the `-` after the section code | die | **killed** |
| Neutral | reword "provider is not initialized", whole package | survive | **survived** |
| Lethal | an injected failing test (`if false` → `if true` before `t.Fatal`) | die | **killed** |

**Re-run of round 1's five survivors** (whole package):

| Mutant | Now |
|---|---|
| INV-59: a state-level `non-regular-file` → `ErrUnknownChange` | **killed** |
| INV-59: a reader error → `ErrUnknownChange` | **killed** |
| INV-65(a): an oid-only state key | survived (F8, unaddressed) |
| INV-54: the symlink manifest line hashed as content | survived (F8, unaddressed) |
| entry-condition F2: the parent-archive type check removed | survived (F8, unaddressed) |

**New mutants: 16 counted, 12 killed, 4 survived.**

- **Killed:**
  - reader error → `ErrNoChangeIntent`, and reader error → `ErrUnmappable:cascade`;
  - every state cause → `ErrUnknownChange`;
  - change cause swapped to `cascade`;
  - `F1-revert-cap`, `F1-overbroad-cap` and `F1-overbroad-specs`;
  - `F5-revert`, `F6-revert`, `F7-revert` and `F9-unquoted`;
  - INV-52 unregistered route (integration).
- **Survived:**
  - `F1-revert-specs` (F1);
  - `V-parents-error-as-unknown` (F2);
  - `F9-no-skip-sort` and `F9-no-rev-sort` (F3).

No kill was caused by the shared embedded-Postgres port. Every kill line in `results.tsv` is the
named assertion. The serial re-runs of `go test ./...` and the integration package were made after
the harness finished, because a run concurrent with the harness had collided on port 55432.

## Gates (run individually; `make check` not run, per the owner's hold)

- **All exit 0:**
  - `invariant-lint`, `invariant-table-lint`, `spec-numbering-lint`, `link-lint`,
    `prescription-lint`, `law-citation-lint`, `cleanroom-lint`;
  - `scripts/p6-census.sh --check`, `scripts/p6-route-matrix.sh --check`,
    `scripts/p6-task2-close.sh --check`.
- `golangci-lint run ./...`: 0 issues.
- `go test -count=1 ./...` (from `kernel/`): all packages ok.
- Integration (`sdd` package, `DATABASE_URL` unset): ok.
- **Will go stale.** Adding this verdict makes `docs/verification/p6-artifact-integrity.md` stale. Its
  `--check` will fail until it is regenerated in the owning session that commits this file.

## Findings (all LOW; none blocks M1)

### r2-m1-recheck-round1-20261118-F1 (LOW): the F1 contract change is partly unpinned and partly not reflected elsewhere in the SPEC

**Broken property:** §9 ("each proving test names the wrong implementation it kills"), and the
SPEC's self-consistency.

**Evidence:**
- **A regular file at `specs` itself is unpinned.** `F1-revert-specs` restores the old outcome for a
  blob at `<change>/specs`, making it `non-regular-file` and the change `cascade`. It **survives the
  whole package**. Only the "directly under `specs/`" case is tested. Real history has no such entry,
  so this is a test gap, not a behaviour defect. Probe RC1 shows the current code gives
  `ErrNoChangeIntent`.
- **Text elsewhere still states the old rule.**
  - §0's answer to re-check F1 (SPEC line 79) still says "**Any other type** at those positions is a
    record unit skipped with `non-regular-file`", which now contradicts §4.1.
  - SDD-INV-54's §9 row does not mention the ignored-regular-file case its test now pins.
  - §7's `non-regular-file` row ("a wrong type at a delta-path layout position (§4.1)") is correct
    only by deferring to §4.1.

### r2-m1-recheck-round1-20261118-F2 (LOW): SDD-INV-59's "a reader failing on purpose" is proven only for `Read`

**Broken property:** SDD-INV-59 ("any other error comes only from the reader"), Law 9's
calibration corollary.

**Evidence:**
- **The test's limit.** `failingReader` fails only `Read`. `Resolve` also calls `Parents` and `List`
  on the `singleArchive` path.
- **The survivor.** `V-parents-error-as-unknown` turns a `Parents` failure into `ErrUnknownChange`
  and **survives the whole package**. The `List` path has the same shape.
- **The sweep's limit.** The agreement sweep compares `Resolve` with `mapState`'s skips, not with
  `Skipped()`. It runs only skip → `Resolve`, so a divergence in `Revisions`' de-duplication or
  stamping is outside it.
- **Why it is not MED.** The current code returns wrapped reader errors on all three paths, which I
  read directly. Round 1's two named survivors die.

### r2-m1-recheck-round1-20261118-F3 (LOW): F9's fixes are partial

**Broken property:** SDD-INV-60 ("one line per Skip"), and the claim "skip and revision order is
deterministic".

**Evidence:**
- **The ordering is not pinned.** Replacing either new `sort.Slice` comparator with an always-false
  one survives the whole package (`F9-no-skip-sort`, `F9-no-rev-sort`). Round 1's probe measured
  6 orders in 8 runs without the sort, so nothing would catch a regression.
- **Some control characters still pass unquoted.** `printableUnit` quotes C0 controls and DEL, but
  not C1 controls (U+0080-U+009F, including NEL and CSI) or U+2028/U+2029. Some terminals treat
  these as line breaks or escapes.
- **The regex-dialect half was not addressed.** `datedRE`'s `.` still excludes `\n`. An archived
  `2026-01-01-x\ny` is `archive-name-invalid`, while the same in-domain name maps when active
  (P2, unchanged). Draft 10 still fixes no dialect for §6's archive pattern. That is now the only
  route by which a newline name becomes a Skip.

### r2-m1-recheck-round1-20261118-F4 (LOW): the probe's new count does not discriminate type

**Broken property:** §0's "probe-printed" provenance for the new row.

**Evidence:**
- The added `history_probe.py` block counts `ls-tree` entries whose type is `blob`. Git lists
  symlinks (`120000`) as type `blob` too. So the row's "regular files" count would include a
  symlink at that position, which §4.1 now treats differently.
- On the reference corpus the count is right: the one entry is `100644`, checked independently.
  The probe would mislabel a future corpus.
- The probe also does not count a blob at `<change>/specs` itself, which §4.1 also covers. There is
  none in history.

### r2-m1-recheck-round1-20261118-F5 (LOW): round-1 F8's survivors remain undeclared

**Broken property:** the `/proofbound-review` rule "undeclared survivors fail; declare any survivor
in `docs/allowed-survivors.txt`", and Law 9's corollary.

**Evidence:**
- `INV65-oid-only`, `INV54-hash-symlink-target` and `F2-untyped-parent-archive` still survive.
- With this round's `F1-revert-specs`, `V-parents-error-as-unknown` and the two sort mutants, they
  appear in neither `docs/allowed-survivors.txt` (no `sdd` entry) nor `docs/deferrals.tsv`.
- **Why this does not block.** Round 1 placed F8 outside M1's acceptance boundary, and every one of
  these is a test-adequacy gap over behaviour I confirmed directly. Left untracked, though, they are
  exactly the "prose lessons decay" class (Law 7).

## Acceptance decision for M1

**ACCEPTABLE, subject only to `make check` passing when the owner lifts the hold.**

- Round-1 F1-F4 are closed, each with the reproduction that opened it now giving the contract's
  answer.
  - F1 holds on real history: `opencode-command-references` is `ErrNoChangeIntent`, and only F1's two
    skips moved.
  - F2's two survivors die, F3's integration test carries all 15 causes and per-unit rows, and
    `prescription-lint` passes.
- The F1 change stays within the one shape round 1 named, so no fresh full round is needed.
- The LOW fixes F5, F6, F7, F9 (quoting) and F10 are correct, and each is killed by a mutant. They
  introduce no regression: every gate, `go test ./...`, golangci-lint and the integration package
  are green.
- F8 was not addressed (confirmed). It does not block, as round 1 decided.
- The five LOWs above are test-adequacy and text-consistency gaps. None is a behaviour defect on any
  input I tried. They may be fixed later, or the survivors declared or tracked (F5), without another
  review of M1's acceptance.
- **Expected red.** `docs/verification/p6-artifact-integrity.md` will be stale once this verdict
  lands, and must be regenerated in the session that commits it. Otherwise `make check` will be red
  for that reason alone.
