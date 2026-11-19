---
schema: vera.verdict.v1
verdict_id: r2-m3-impl-recheck-20261119
status: ACCEPTABLE
reviewed_commit: eecd941fff903bc235321668c5eee99720961d4d
findings:
  - finding_id: r2-m3-impl-recheck-20261119-F1
    severity: LOW
  - finding_id: r2-m3-impl-recheck-20261119-F2
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-impl-recheck-20261119.md
artifact_sha: ce387188bc06cd5ea2df853ca06714f60f01151a2cdf4928ebd5640db7231f7d
---

# R2 M3 implementation re-check (non-author), 2026-11-19

## Provenance

- **Reviewer:** an independent Claude Code subagent (Opus 5.5). It authored none of the M3 code, none of the fixes under review (9464e06, 6f43108, af57109, 75c9eac), and neither of the two verdicts they answer. This satisfies Law 9.
- **HEAD:** eecd941fff903bc235321668c5eee99720961d4d at the start and at the end. It did not move.
- **Isolation:**
  - All work ran in a private detached worktree at `<scratch>/recheck/wt`, with GOCACHE and TMPDIR under `<scratch>/recheck`, DATABASE_URL unset, and Go 1.27.4.
  - The worktree was removed with `git worktree remove --force`. `git worktree list` now shows only the main checkout.
  - The main checkout was never modified. Its only untracked files are the two that were already there.
  - Nothing was committed or pushed. notes/ and scripts/p6-census.sh were not touched.
  - Nothing was listening on port 55432 after each sdd-package run. No process was killed.
- **What I read:**
  - Both verdicts.
  - The full diffs of 9464e06, 6f43108, af57109 and 75c9eac.
  - In cli.go: `resolveSDD`, `segmentNamesChange`, `readSDDLedgered`, `readLedgerView`, `storeLedgerView`, `printHolds`, `printClaimSkips`, and the `sync reviews` / `sync all` print paths.
  - The projections commit reducer.
  - `connectorintent.ValidateRevision`.
  - The git connector's `isSDDClaim`, `resolvedSDD` and `sddValueRE`.
  - `store/lock.go`, `store.Open`, and the paging in `ReadEvents`.
  - `port_home_test.go`.
  - M3 SPEC draft 5 §6/§7 (rows 10–12), draft 7 §4.6 and row 12, and draft 8 row 11.
  - Projections SPEC P-INV-34, reviews SPEC R-INV-13 with its new limitation note, and the git SPEC monotone limitation.
- **What I ran:**
  - Baselines of every new or changed test.
  - The full untagged `./internal/cli`, `./internal/projections` and `./internal/connector/...` packages (all pass). The sdd package ran on the fixed port, one process at a time.
  - `go vet`.
  - `gen-invariants-lock.sh`, which regenerates byte-identically.
  - `invariant-table-lint.sh` (clean) and `p6-route-matrix.sh --check` (exit 0).
  - 36 hand mutants of production and detector code, plus 6 probe files for the detector.
- **Production diffs match their claims:**
  - af57109 touches no production `.go` file: only four `_test.go` files plus reviews SPEC.md.
  - 9464e06's only production change is cli.go:
    - `ValidateRevision` replaces the local id and digest regexps.
    - The producer types `ChangeIntent` and `Requirement` replace the anonymous decoders.
  - Behaviour is unchanged, or fails closed:
    - `ValidateRevision` is stricter than the old regexps: it requires a canonical, strictly decoded payload, a non-zero ObservedAt, and a payload that binds the envelope.
    - It allows BD/BR prefixes, but the git connector's CI-only `intentIDRE` check on the returned ref still holds, and so do the Kind and segment checks.
    - Real sdd revisions still bind: the full sdd package passes, including its `cli.Run` sdd-claim tests.

## Calibration (stated before any count)

- **Positive control:** a commit-reducer mutant that runs `DELETE FROM commit_intents_view WHERE commit_sha=$1 AND seq<$2` on every commit event. It must die. **It was killed.**
- **Neutral controls:** one per test area, each a behaviour-preserving edit that must survive. All survived:
  - projections: `_ = len(intentRefs)`
  - view: `make(map…)` in place of `map…{}`
  - reviews print: holds error split into a local variable
  - resolver: `path.Clean` around ArtifactPath
  - printHolds: `string([]byte(path))`
  - git: `len(raw) <= 3`
  - detector: `var port string`
- **Runner hygiene:** the runner applies exactly one occurrence of each pattern and restores with `git checkout -- kernel`. Worktree status was clean after every mutant.
- **One discarded run:** a first form of F4b failed to build. It was re-expressed so that it compiles, and it was then killed. That build failure is not counted as a kill.

With the harness calibrated:
- 27 mutants expected to die were killed.
- 9 expected survivors survived: 7 neutrals, plus 2 documented equivalents.
  - **F4e** drops the Kind check. The author already declared it equivalent, and I confirm it survives.
  - **P4** narrows printableUnit's C1 range. The new test uses only `\n`, so it is out of scope for that test; older printableUnit tests own it.
- 1 mutant I expected to survive did survive, and that exposes a gap: R2 (new F2).

## Closure table

| Finding | Closed? | Evidence (my mutants) | Regression / spec drift |
|---|---|---|---|
| A-F1 (MED) monotone bindings untested | **Closed** | `TestBindingsAreMonotone` (untagged, ephemeral port) killed the positive control and a "delete all rows when the newer event has no refs" mutant. Neutral survived. The test follows draft 5 row 10's sequence: change intent → binding → later unbound event → v2 verdict → Apply + Rebuild → row present. | P-INV-34 restates M3-INV-10 faithfully. invariants.lock regenerates byte-identically. Pre-existing nuance, not introduced: the git SPEC says "insert-only", but the upsert moves `event_id`/`seq` to a later event that repeats the ref. The row still survives, so monotonicity holds. |
| A-F2 (LOW) undecodable git / schema-less reviews | **Closed** | V1 (schema-less reviews skipped) and V2 (undecodable git payload skipped) killed. | None. |
| A-F3 (LOW) CLI half of malformed-after-refusal | **Closed** for the stated scenario | R1 killed: holds print only when the refusal is the sole leaf error. | See new F2 for an adjacent clause that is still unpinned. |
| A-F4 (LOW) Law 2 grammar/decoder copies | **Closed** | The local regexps are gone, and the producer types are used. F4a (drop ValidateRevision), F4b (drop segment check), F4c (dated segment accepts any name), F4d (prefix match for active segment) and F4f (drop Source check) all killed. F4e (drop Kind check) is equivalent, as the author states. | `sddDatedRE` stays as the archive-date prefix, which is not a grammar copy. |
| A-F5 (LOW) detector narrower than its lesson | **Closed for the named shapes** (multi-line literal, `Port: 0`, unstubbed `runCommand`); residual in new F1 | Probes: a multi-line portless `store.Config` in projections is FLAGGED, and an unstubbed `runCommand` in a cli file is FLAGGED. D-zero-is-set and D-no-runCommand killed the calibration test. The 16 in-test probes run on every `make check`. | — |
| B-F1 (LOW) fail-closed view and L | **Closed** | V1–V3 killed. V4 (keep partial sets after a failed load, i.e. C49) killed by the second-call assertion. V5 (L decode skipped, i.e. C22) killed. V6 and V7 (CommitClaimed / RequirementLedgered ignore load errors) killed. | None. |
| B-F2 (LOW) printableUnit on claim-skip / held / refused | **Closed** | P1 (raw hold path), P2 (raw claim-skip value) and P3 (printHolds swallows write error) killed. | None. |
| B-F3 (LOW) detector and D1 edges | **Closed** | G1 (`len<3`), G2 (drop colon), G3 (colon only), G4 (drop byte 1), G5 (slug ≤98), G6 (slug ≤100) and G7 (slug min 0, which admits `sdd:a`) killed. | None. |
| B-F4 (LOW) evadable default-port guard | **Closed** for the multi-line evasion and for `Unsetenv`; residual in new F1 | Multi-line probe flagged. The in-test calibration includes `Unsetenv("DATABASE_URL")` and an allowed `Unsetenv("HOME")`. | — |
| B-F5 (LOW) non-linearizable held vs refused | **Closed as an accepted limitation**, with an accurate note | Checked against the code. `acquireLock` runs in `store.Open` before the DatabaseURL branch, so it applies even with an external URL. The lock is `<DataDir>.lock`, so it is per state root. `ReadEvents` pages with `seq>$cursor ORDER BY seq LIMIT n` and no snapshot. Prerequisite sets only grow, because the ledger is append-only and the claims set is a union over every git event. So a stale or paged view can only under-report, and "nothing is appended ahead of its prerequisites" holds. | Minor imprecision, not a finding: "the next sync refuses it" holds only if the prerequisite is still missing then. If it has since been ledgered, the next sync appends the revision, which is also correct. |

## New findings

### F1 (LOW): the compiled default-port guard still misses four syntactic shapes

The detector's own doc comment says it "returns each syntactic way it can open the embedded store at the default port", and 9464e06 says it "catches ... unstubbed Run or runCommand in package cli". Neither statement is fully true. Each probe below compiled, and each one was **not flagged** by `TestDefaultPortStoreTestsLiveInOnePackage`:

1. **One stub exempts the whole file.** A cli test file that assigns `openStore` anywhere, even `openStore = old` in a cleanup or in a different test, exempts every `runCommand` / `Run` in that file. The stub check is file-scoped, not call-scoped. `ledger_view_test.go` and `cli_test.go` already contain such assignments, so any new unstubbed test appended to either file passes the guard.
2. **A second opener is not covered.** A cli test that stubs `openStore` but runs `commandMigrateHistoricalEvidence` reaches `openHistoricalEvidenceStore`. That opener calls `store.Config{Root, DatabaseURL, AllowHistoricalEvidenceImport}` with no Port, i.e. the default port. Today's tests stub it.
3. **An aliased store import evades the match.** For example, `st.Config{Root: "x"}`: `isStoreConfig` matches the identifier name `store` literally.
4. **A Config built without a composite literal evades the match.** For example, `var c store.Config; c.Root = …; store.Open(ctx, c)`.

**Failure scenario:** a future untagged test appended to `ledger_view_test.go` calls `runCommand(ctx, commandSyncGit, nil, root, "", out)` without its own stub. It opens 55432 concurrently with the sdd package under parallel `go test ./...`, which brings back the 2026-11-09 / 2026-11-19 flake while the compiled lesson stays green.

**Current state:** no offender exists today. Every current call site is stubbed or uses an ephemeral port. Hence LOW, the same class as A-F5 / B-F4 but narrower.

### F2 (LOW): "held lines are printed before ANY error" is unpinned when the error is not a refusal

Draft 8 §4.6 and the cli.go comment both say counts, held and refused lines are printed before any error is returned.

**Mutant R2:** in `commandSyncReviews`, return `syncErr` without printing holds when it does not wrap `ErrRevisionRefused`. It **survives**:
- the targeted test;
- the full `./internal/cli` package;
- the full `./internal/connector/intent/sdd` package, which holds the `cli.Run` M3 tests.

A-F3's new case covers only malformed *after a refusal*, where the joined error still wraps `ErrRevisionRefused`.

**Failure scenario:** a sync lists one v2 verdict that is held (claim not yet ledgered) and one malformed artifact later in path order. Under an R2-style refactor, the operator sees the malformed error and a counts line but no `held claim-not-ledgered <path>` line. The held verdict then goes unexplained until a later clean sync.

The exit status is unaffected, since the run is non-zero for the malformed artifact anyway, and nothing is ledgered out of order. Hence LOW.

## Verdict

**ACCEPTABLE.**

- All ten findings answered by 9464e06 and af57109 are closed, or recorded as an accepted limitation (B-F5) whose SPEC note matches the store lock and `ReadEvents`.
- Every new or changed test kills my own production mutants for its property, under a calibrated harness: the positive control died and every neutral survived.
- af57109 changed no production code. 9464e06's production change in cli.go is behaviour-preserving or stricter, and it fails closed.
- The SPEC amendments (P-INV-34 and the R-INV-13 limitation) are accurate, and the invariants lock and route-matrix pin are fresh.
- Two new LOW findings remain:
  - **F1:** the residual blind spots of the default-port guard.
  - **F2:** held lines on a non-refusal error are unpinned.
- Neither finding leaves a path where something is ledgered ahead of its prerequisites, or where a gate passes wrongly.
