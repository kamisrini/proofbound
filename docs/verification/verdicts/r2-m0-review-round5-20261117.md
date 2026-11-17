---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round5-20261117
status: NEEDS_WORK
reviewed_commit: f06c95f6f4a7eda16b3858a1ba983023c6344306
findings:
  - finding_id: r2-m0-review-round5-20261117-F1
    severity: MED
  - finding_id: r2-m0-review-round5-20261117-F2
    severity: MED
  - finding_id: r2-m0-review-round5-20261117-F3
    severity: MED
  - finding_id: r2-m0-review-round5-20261117-F4
    severity: LOW
  - finding_id: r2-m0-review-round5-20261117-F5
    severity: LOW
  - finding_id: r2-m0-review-round5-20261117-F6
    severity: LOW
  - finding_id: r2-m0-review-round5-20261117-F7
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round5-20261117.md
artifact_sha: 7c407c0c719aeecad593ef2d7e014f7a6dba953b981a446fedd403e43748be67
---

# Release 2 M0 (SDD Warranty Connector): draft-5 closed-cause, shared-reader provider contract. Review round 5

## Provenance

- **Under review:** HEAD `f06c95f6f4a7eda16b3858a1ba983023c6344306`. The content was introduced by
  `f1beb76`, and `f06c95f` is an `@generated` gate-pin refresh (`scripts/p6-task2-close.sh --write`).
  Files:
  - `docs/plans/R2-sdd-provider-spec-draft5.md`, which supersedes draft 4.
  - `kernel/internal/connector/intent/sdd/testdata/`:
    - `branches/14-stale-tip`;
    - the rebuilt `fenced/` (`improve-onboarding`, the discriminating case, and `document-renames`,
      the neutral control);
    - `expected-skips.json`;
    - `README.md`;
    - `derive/history_probe.py`.
  - The intent SPEC §1 pointer, and the status line in `docs/plans/sdd-warranty-connector-plan.md`.
- **Also reviewed: `kernel/internal/connector/intent/sdd/fixture_test.go`, as uncommitted
  working-tree content**, by owner decision. It has SHA-256 `3221e184…b276a371`, and it contains:
  - the replay helper with PARENTS/REF;
  - `requireProfiles`;
  - `TestBranchFixtureReplaysAsDocumented`, including the fenced checks.

  The committed `HEAD` version still has none of these.
- **Out of scope:** `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`, which are
  untracked.
- **Follows up:** `docs/verification/verdicts/r2-m0-review-round4-20261117.md` (NEEDS_WORK, F1..F10).
- **Judged against:** the CLAUDE.md Build Laws and this ground truth:
  - **`intent.go`:**
    - `New`, with the provider allowlist at line 134;
    - `Sync`, at lines 145-211;
    - `ValidateRevision`, at lines 214-278;
    - `validateObligations`, `validateBatch` and `orderBatch`;
    - `recordIDRE` and `obligationRE`.
  - **`core/core.go`:** `Canonicalize`, `NewEvent`, and `Event.Validate`, whose payload bound is
    `MaxPayloadBytes` at line 64.
  - **`gitcmd.go`:**
    - `Commits`, at lines 53-78;
    - `readCommit`, `files` and `nulUTF8Strings`;
    - `refusePartial`, `validateRefs` and `run`.
  - **`connector/git/connector.go:198`.**
  - **`projections/projections.go`:**
    - the DDL at lines 90-118;
    - `eventRoute`, at line 296;
    - `reduceIntent`, at lines 302-375;
    - the `IntentRef` check at line 723.
  - **`projections/intent_report.go`.**
  - **`gates/gates.go:304`.**
  - **`cli/cli.go`:** lines 37, 88-93, 208-262, 366-382, 513-528, 589-625 and 834-860.
  - **The scripts and store schema:**
    - `scripts/p6-route-matrix.sh` and `scripts/p6-census.sh`;
    - `scripts/delivery-enforce.sh`;
    - `store/migrations/001_ledger.sql`.
- **Real-history evidence:**
  - **Repository.** The author's clone of `Fission-AI/OpenSpec` at
    `f085c850f9440f497ccc76ecbc2859c09e0e0ac7`. It is not shallow.
  - **Parser source.** I read OpenSpec's `src/core/parsers/` in that clone: `code-fence.ts`,
    `requirement-blocks.ts`, `spec-structure.ts` and `requirement-text.ts`.
  - **My own port.** I wrote `<scratch>/osparse.py`, a port of OpenSpec's `parseDeltaSpec`, together
    with a literal reading of draft 5 §4.2 under Go RE2 semantics.
  - **Probes.** `<scratch>/blobs.py`, `<scratch>/dual.py`, `<scratch>/neveractive.py` and
    `<scratch>/synth.py`. I did not reuse the author's probe output.
- **Synthetic evidence:**
  - **`ValidateRevision` probe.** `<scratch>/tenantbytes_probe_test.go`, injected into the intent
    package with `go test -overlay`.
  - **`gitcmd.Commits` probe.** `<scratch>/latin1_probe_test.go`, injected into `gitcmd` the same way.
  - **Postgres.** A single-user PostgreSQL 16.4, extracted from the local embedded-postgres cache into
    `<scratch>/pg`.
  - **Go regexp semantics.** `<scratch>/gore`.
  - **Fixture calibration.** Throwaway module copies under `<scratch>/cal-*`.

  None of these wrote a repository file.
- **HEAD at dispatch and at completion:** `f06c95f6f4a7eda16b3858a1ba983023c6344306`, unchanged.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. M1 owns record identity, point-in-time resolution and a new
  ledger-reaching source, so M1 itself should be reviewed at the foundational tier, as rounds 2-4
  recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m0-r5`, a **non-author**. It wrote
  none of the work under review and is not the round-4 reviewer, so this verdict counts under Law 9.
- **Writes:** no repository file other than this verdict. No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m0r5/` in the session scratchpad. It
    occurs 15 times in this verdict.
  - The OpenSpec clone is named only as "the author's clone".

## Round-4 findings re-verified

I checked draft 5's §0 mapping table against the text and the code, not against the table itself.

- **F1 (oversize payload wedges after a partial append): FIXED for size, but the class
  remains.**
  - `payload-too-large` and `invalid-utf8` are now §7 data causes.
  - §12 moves the `MaxPayloadBytes` bound into `ValidateRevision` for every provider.
  - Residual: a NUL byte is valid UTF-8, under the size bound, and accepted by `ValidateRevision`,
    `Event.Validate` and the ledger's `JSON` column. The `TEXT` column of the projection then rejects
    it on every rebuild. The wedge has moved from append to projection (new F2).
- **F2 (the catch-all turns provider defects into skips): FIXED in design. The list is not
  closed.**
  - §7 now ends with "everything else fails `Revisions`".
  - INV-49 kills the draft-4 catch-all.
  - INV-46 pins exact Skip sets.
  - Residual: several tenant byte patterns pass every listed cause and then fail
    `ValidateRevision`. Under the new rule they fail `Revisions` permanently (new F1).
- **F3 (incomplete amendment list): MOSTLY FIXED.**
  - §12 names `New`, `ValidateRevision`, `eventRoute`, the gates matcher, intent SPEC §6/INT-INV-11,
    both scripts and the M3 sites.
  - SDD-INV-52 projects a sync.
  - Residuals (new F4): one CLI line is mis-assigned, the M1 construction site is missing, and the
    consequences of "`all` includes it" are unexamined.
- **F4 (the reader shares only the arguments): FIXED.**
  - The Reader is `gitcmd.Repo`. Every call goes through `run`, so `--no-replace-objects`, the `-c`
    options and the scrubbed `GIT_*` environment are shared.
  - `Commits` is the git connector's own function, so `refusePartial` and `validateRefs` are shared.
  - SDD-INV-51 tests a graft, a shallow clone and `info/grafts`.
  - Consequence: the reader also inherits `Commits`' failure on a non-UTF-8 path anywhere in history
    (new F1(c)).
- **F5 (INV-39 cannot discriminate): FIXED for the fixture. The pinning claim is false.**
  - The rebuilt fixture discriminates. Measured independently with my own port of OpenSpec's
    parser:

    | Fixture | Masked | Unmasked |
    |---|---|---|
    | `improve-onboarding` | 3 | 1 |
    | `document-renames` | 1 | 1 |

  - §0's rationale is corrected.
  - Residual: §4.2 says it is "OpenSpec's rules, pinned to `src/core/parsers/`", but it diverges
    from `parseDeltaSpec` in five concrete ways (new F3).
  - Residual: the self-test does not guard the discrimination it claims (new F6).
- **F6 (tip-fatal survives; meanings flipped in place): FIXED.**
  - `branches/14-stale-tip` puts the `stale` ref on the leftover state. My calibration shows that the
    self-test catches its loss.
  - INV-1, 3, 4, 6, 25, 29, 36, 37 and 38 are retired and re-issued as INV-40 to 52.
  - I diffed every kept row (2, 5, 7, 9, 10, 12, 13, 18, 20, 26, 28, 30-35, 39) against draft 4. Each
    edit is wording, or a test strengthened under the same statement. None changes its meaning.
  - The tombstone reasons of INV-15 to 23 are restored, matching draft 3 except for dropped section
    cross-references.
  - Residual: a HEAD-only-fatal variant still survives (F6).
- **F7 (INV-38 against INV-31/33): FIXED.**
  - The collision check is per state, so it is decidable from `c`'s state alone.
  - Across states, two `(key, manifest)` pairs have different digests, because the manifest carries
    the path. So `(id, digest)` stays unique, and INV-44 holds.
  - I checked every view and every report query in `projections.go` and `intent_report.go`. Each
    keys on `(source, id, artifact_sha256)`, except the two `report … <id>` listings. Those print
    every revision with its source and digest, so a shared id is displayed, never merged.
- **F8 (mis-scoped counts): FIXED for the rows round 4 named, with one residual.**
  - "23" and "4 remote branches" reproduce exactly.
  - The "47 lines" row is removed.
  - Residual: other §0 rows are still path-limited, for example "16 / 108" (F7).
- **F9 (not self-contained; drift): MOSTLY FIXED.**
  - The profile schema is inlined. All **30** tracked fixture profiles conform to it (the brief said
    29; `14-stale-tip` adds one).
  - Obligation `state` is defined as `active`. That is consistent with `projections.go:364`, which
    requires `state='active'` for a target.
  - The archive-name regex, BOM and CRLF are specified.
  - The top README is fixed.
  - Residuals (F7): `testdata/unmappable/README.md` still cites retired INV-14, -6 and -29, and the
    plan says "superseded by draft 3 (now draft 4)".
- **F10 (hand-offs): FIXED for the sentinels. Still open for the helper.**
  - `ErrUnknownChange`, `ErrArchivedEarlier` and the wrapped `ErrUnmappable` are present (INV-50).
  - The helper is still uncommitted. §13 makes committing it an M1 precondition, which is acceptable
    at M0.

**Round-4 "Fitness for M1" conditions:**

1. **Close §7's enumeration: PARTIALLY MET.**
   - Size and encoding are now causes.
   - The catch-all is gone, with INV-49 and INV-46 as its tests.
   - Tenant bytes still reach non-listed failures (F1) and a post-append failure (F2).
2. **List every registry, with its test: MOSTLY MET.** The registries are listed, and INV-52 projects
   a sync. The CLI rows are wrong (F4).
3. **Share the full read contract: MET.**
4. **Rebuild INV-39, add a tip test, settle the meanings: MET.** The residuals are F3 and F6.
5. **INV-33 against INV-38, the inlined profile, obligation `state`: MET.** **Commit the helper: NOT
   YET**, but it is now an explicit M1 precondition.

## Evidence reproduced

My own probe and git commands give these results.

| §0 row | Author | Reviewer |
|---|---|---|
| `--all` walk (merges, roots) | 1624 (120, 1) | 1624 (120, 1) |
| Distinct SDD states | 338 | 337 non-empty `openspec/changes` trees, plus "absent" = 338 |
| HEAD path-limited / `--all --full-history` | 256 / 451 | 256 / 451 |
| Dual-presence commits on the unsimplified `--all` walk | 23 | **23** (`7ced2a87` … `acd1ca28`, one name: `adopt-verb-noun-cli-structure`); archived twice: 0 |
| Refs at those commits | 4 remote branches, 0 tags | **4**: `origin/archive-changes`, `origin/prepare-publish`, `origin/update-openspec-agent-instructions`, `origin/updating-agent-instructions` |
| Distinct delta blobs / blobs changed by masking | 401 / 1 | **401 / 1**: `archive/2025-09-29-improve-init-onboarding/specs/cli-init/spec.md`, 4 masked against 2 unmasked. This uses my port of `parseDeltaSpec`. The blobs are enumerated per distinct tree over the unsimplified walk, so the count does not depend on `rev-list --objects` simplification. No blob sits at a nested capability path |
| Archived without ever being observed active | 16 / 108 | **16 / 108 on the HEAD path-limited walk; 12 / 108 on the provider's `--all` walk** (F7) |
| Draft-5 §4.2, read literally (Go RE2), against `parseDeltaSpec` over the 401 blobs | — | 0 blobs differ, so the F3 divergences are latent on this history |

**Fixtures:**

- `go test -count=1 ./internal/connector/intent/sdd/` passes, with both self-tests run against the
  working-tree helper.
- My own scan of every fixture state, using my port, finds exactly one skip-cause state:
  `add-z` name reuse at branches 11, 12 and 14, which is the same state. That agrees with
  `expected-skips.json`. The unmappable pins (`Login!`/`Login?` → `O-ADD-LOGIN`, and `add-login`
  reuse) also check out by hand.

**Calibration**, on throwaway module copies:

| Control | Mutation | Expected | Result |
|---|---|---|---|
| Baseline | none | pass | passes |
| Positive 1 | append a line to the `add-z` leftover in `14-stale-tip` | fail | fails: `step 14 is not a ref tip at the leftover state` |
| Positive 2 | indent `## TL;DR` in `improve-onboarding` | fail | fails: `fenced fixture lacks a column-0 header inside a fence…` |
| Neutral | edit `14-stale-tip/tree/src/x.ts` | pass | passes |
| Gap probe 1 | rename `## MODIFIED Requirements` to `## Notes` in `improve-onboarding` | — | **passes**, although masked and unmasked both yield 0 (F6) |
| Gap probe 2 | set `expected-skips.json` `branches` to `[]` and `fenced` to a bogus entry | — | **passes** (F6) |

## Findings

### r2-m0-review-round5-20261117-F1 (MED): §7's closed list is not closed. Tenant bytes that pass every listed cause still fail `ValidateRevision` or the shared reader, so `Revisions` fails permanently

**Broken property:**
- SDD-INV-35: "No tenant content at any commit makes Revisions fail".
- §10 D.
- §7's claim that a non-listed failure "is recovered by fixing the provider or the clone, never by
  editing history".

**Evidence** (`<scratch>/tenantbytes_probe_test.go`, run against the unmodified intent package):

- **(a) A whitespace-only statement.**
  - §4.2's Spec Kit rule `^- \*\*(?P<id>…)\*\*:? (?P<statement>.+)$` reads the ASCII line
    `- **FR-001**:  ` (a colon and two spaces) as statement `" "`.
  - `ValidateRevision` then returns `intent: invalid obligation`, because `validateObligations`
    rejects `strings.TrimSpace(statement) == ""`.
  - The OpenSpec rule "the first non-blank line" leaves "blank" undefined. A line holding only
    U+00A0 is non-blank under an ASCII reading, but Go's `TrimSpace` empties it, with the same
    result.
  - None of these is a listed cause, so under §7 each fails `Revisions`.
- **(b) An obligation id too long for `obligationRE`.** Spec Kit ids are `O-<id>` with an unbounded
  `[0-9]+`. `O-FR-` followed by 30 digits is 33 characters after `O-`, so `obligationRE` rejects it
  with `intent: invalid obligation`. This is not a listed cause either.
- **(c) A non-UTF-8 path anywhere in history.**
  - §3 makes the reader's `Commits` "the same function the git connector calls". That function
    calls `readCommit` and then `files`, and `files` fails on any non-UTF-8 path in a commit's
    diff (`nulUTF8Strings`).
  - Reproduced in `<scratch>/latin1_probe_test.go`, with a file named `caf\xe9.md` added outside any
    SDD root: `gitcmd: list files for "ff905f…": path is not valid UTF-8`.
  - Under §7 that is "a reader failure", so it fails `Revisions`. It is a tenant byte in a past
    commit, and it can only be fixed by editing history, which contradicts §7's recovery sentence.
  - The repository needs no profile at all for this to happen (see F4).
  - A Latin-1 author name does not trigger it, because git re-encodes it.
- **(d) Non-blob entries are unspecified.** A gitlink or a symlink under an SDD root has no
  specified meaning. `Read`'s "every committed file" leaves open whether a gitlink fails the reader,
  which would be a permanent `Revisions` failure.

**Why it matters:** this is the same permanent-wedge class as round-3 F1 and round-4 F1. INV-49
would *require* each of these failures, and no INV-35 row places any of them.

### r2-m0-review-round5-20261117-F2 (MED): a NUL byte in tenant content passes every check through append, and then fails every projection rebuild

**Broken property:**
- §7's "Batch-safe … `Sync` validates, orders, and appends every `sdd` batch end to end".
- SDD-INV-52 ("A sync containing intent.sdd events projects and rebuilds").
- The append-only ledger's recoverability.

**Evidence:**
- **The provider and connector accept it.** U+0000 is valid UTF-8, so `invalid-utf8` does not fire.
  A statement containing it is not "blank", so `ValidateRevision` returns `<nil>` (probe). The
  canonical payload carries `\u0000`, and `Event.Validate` accepts it.
- **The ledger accepts it.** `events.payload` is `JSON` (`001_ledger.sql:9`). On PostgreSQL 16.4 in
  `<scratch>/pg`, `INSERT … VALUES ('{"statement":"a\u0000b"}')` into a `JSON` column **succeeds**.
- **The projection rejects it.** `reduceIntent` inserts `o.Statement` into
  `requirement_obligations_view.statement TEXT` (`projections.go:102-105`, `:339`). On the same
  server, a `TEXT` insert containing 0x00 fails with `invalid byte sequence for encoding "UTF8":
  0x00`, and `payload->>'statement'` fails with `\u0000 cannot be converted to text`.
- **Every route reaches it.**
  - An OpenSpec statement line.
  - The fallback "statement is the name".
  - A RENAMED `TO:` line.
  - A Spec Kit statement.
  - A profile `owner` or `sponsor` written as `"\u0000"` in valid JSON, which lands in
    `declared_owner TEXT`.
- **Why it is permanent.** Once one such event is ledgered, every `rebuild` fails, and the ledger
  cannot drop it.
- **No test sees it.** SDD-INV-47 stops at append, and SDD-INV-52 projects only the clean
  `openspec-history`.

### r2-m0-review-round5-20261117-F3 (MED): §4.2 is not OpenSpec's parser, although it says it is pinned to it. Two conforming implementations, one following the draft's regexes and one following the code it cites, emit different obligations

**Broken property:**
- The §4.2 heading "OpenSpec's rules, pinned to `src/core/parsers/` at `f085c850`".
- Spec exactness.
- The family's no-fabrication rule.
- SDD-INV-9 ("kills: dropping REMOVED").

**Evidence.** `<scratch>/synth.py` compares my port of `parseDeltaSpec` (OpenSpec) with a literal
reading of draft 5 under Go RE2. Go semantics were confirmed in `<scratch>/gore`.

| Input | OpenSpec | Draft 5 |
|---|---|---|
| `## REMOVED Requirements` + ``- `### Requirement: Legacy Export` `` | REMOVED `Legacy Export` | **nothing** (silent drop) |
| `## ADDED  Requirements` (two spaces) + one requirement | nothing (OpenSpec warns/refuses) | **ADDED `Foo`** (fabricated) |
| `##` + U+00A0 + `MODIFIED Requirements` | MODIFIED `Foo` | nothing |
| a `## ` line (one trailing space) between two requirements | A and B | **A only** |
| RENAMED `FROM a`, `FROM b`, `TO x`, `TO y` | 1 pair plus 2 unpaired, so the delta is refused | "pair up in order" is ambiguous: positional gives 2 pairs, OpenSpec's rule gives `unpaired-rename` |

The causes:
- **The wrong source regex.** The section regex the draft quotes is `DELTA_HEADER` from
  `spec-structure.ts`, which is a main-spec diagnostic. The reader, `parseDeltaSpec`, works
  differently:
  - it splits on `^(##)\s+(.+)$`;
  - it strips the closing sequence;
  - it matches the title by an exact lower-case fold (`requirement-blocks.ts:430-482`).

  Its own comment (`:293-297`) warns that "a looser test (say, any run of whitespace) would treat
  `## ADDED  Requirements` as a delta section … while the reader ignores it".
- **The REMOVED bullet form.** `parseRemovedNames` accepts the bullet form (`:535-557`). §4.2 omits
  it.
- **The regex dialect is unstated.** JavaScript's `\s` includes U+00A0, `\v` and U+FEFF. Go's `\s` is
  `[\t\n\f\r ]`. Go's `(?i)` folds `ſ` to `s`, so `## ADDED Requirementſ` matches in Go and not in
  OpenSpec.
- **The RENAMED pairing is ambiguous.** "Pair up in order" does not state OpenSpec's displacement
  rule (`:580-614`).
- **The statement is not OpenSpec's requirement text.** OpenSpec skips `**metadata**:` lines and
  stops at `^#{1,6}\s`. The draft takes "the first non-blank line before any line starting with `#`".
  That may be a deliberate choice, but §4.2 presents it under "OpenSpec's rules".

**Severity:** the divergences are latent on the reference history (0 of 401 blobs differ). They are
MED because round 4 asked for exactly this pinning, and the contract is what M1 codes against.

### r2-m0-review-round5-20261117-F4 (LOW): §12's CLI rows are mis-assigned and incomplete, and "`all` includes it" changes every tenant's `sync intent all` without saying so

**Broken property:** §12 "complete list … assigned to M1 or M3", and "M1 must not touch" the M3
sites.

**Evidence:**
- **Line 247 is the resolver, not sync.** `cli.go:247` is `case "specdir":` inside
  `commitIntentResolver.Resolve` (`cli.go:236-262`). That is the git connector's claim resolver, so
  it is M3 work, yet §12 lists it as M1 "`sync intent sdd`".
- **The M1 construction site is missing.** The site where M1 must build the provider is
  `syncIntentOnStore` (`cli.go:366-382`): it constructs `records` and `specdir` and passes them to
  `connectorintent.New`. It is not listed. Neither are the command enum (`:57-60`) or the second
  dispatch (`:513-514`). Line 514 is listed, but only as part of a list headed "sync, all, usage".
- **"`all` includes it" has unexamined consequences.** `sync intent all` is called by:
  - `sync all` (`cli.go:596`);
  - `proofbound verify` (`:835`, `:854`, through `make verify`);
  - `scripts/delivery-enforce.sh:38`;
  - `docs/releases/r1/seed-demo-app.sh:94`.

  With `sdd` in `all`, every such run walks the full history through `gitcmd.Commits`, even in a
  repository with no profile. It therefore inherits `ErrShallow` and F1(c), and the first provider
  error aborts the whole `Sync` (`intent.go:166-168`). A standalone `sync intent all` in a shallow CI
  checkout, which works today, would start failing. §12 and §14 do not mention this.

### r2-m0-review-round5-20261117-F5 (LOW): §5's Spec Kit change-intent status `accepted` is fabricated

**Broken property:** §1's family rule "no … status … is invented".

**Evidence:**
- **Nothing in the kit records acceptance.** In Spec Kit, `tasks.md` is generated by the `/tasks`
  step. It records no approval, by anyone. §5's own justification says only that work was "committed
  to".
- **Draft 2 recorded the opposite rule.** Draft 2 §3: "The kit records no approval, so the tenant
  declares what an active change means in its workflow". Draft 4 applied the profile's
  `changes.status` to active changes.
- **The tenant can no longer declare it.** Draft 5 removes `changes.status` from the Spec Kit
  profile (§2.1) and hard-codes `accepted`.
- **Severity.** It is LOW only because no kernel path reads change-intent status except
  `superseded`/`withdrawn` (`intent_report.go:186`; `gates.go` has no status predicate). The report
  still prints `status=accepted` for a feature nobody accepted.

### r2-m0-review-round5-20261117-F6 (LOW): the fixture self-tests do not guard the properties they claim, and two rejected implementations still have no killing fixture

**Broken property:**
- The LESSON class "fixture-not-what-it-claims", which the helper names.
- The Law 9 corollary: each named kill must have a case that discriminates.

**Evidence:**
- **The fenced check tests string positions, not discrimination.** It survives a mutation that
  removes the discrimination entirely: renaming the MODIFIED header gives 0 obligations masked and
  0 unmasked, and the test passes (Gap probe 1). §13's "verified with the probe's §4.2 parser:
  masked 3, unmasked 1" is not mechanised.
- **`expected-skips.json` is checked only for JSON validity** (Gap probe 2). Nothing ties its steps
  or unit paths to the fixtures.
- **§4.2's mask precision has no killing fixture.** No fixture separates OpenSpec's mask from a naive
  toggle. Specifically, none tests:
  - the same marker for the closer;
  - a closer at least as long as the opener;
  - no backtick in a backtick fence's info string.

  This is the precision round 4 asked to be specified.
- **A HEAD-only-fatal implementation survives.** INV-35(b)'s tip is the non-HEAD ref `stale`, and
  `main` ends on a clean state. On real history, `main` itself sat on the dual-presence state for 9
  commits, so a HEAD-fatal provider would have failed every sync in that window.

### r2-m0-review-round5-20261117-F7 (LOW): residual drift and mis-scoped evidence

**Evidence:**
- **A testdata README still cites retired invariants.** `testdata/unmappable/README.md` still says
  "SDD-INV-14, -6, -29". All three are retired. §0 lists the README among the round-4 F9 items it
  fixed.
- **The plan names the wrong draft.** `sdd-warranty-connector-plan.md` still says "superseded by
  draft 3 (now draft 4)".
- **More §0 rows are path-limited.** §0 corrects only the "12". The "16 / 108", "34", "47 (10) / 10
  (5)" and "4" rows still come from the probe's first loop, `rev-list --topo-order HEAD --
  openspec/changes`. On the provider's walk, "archived without ever being observed active" is 12,
  not 16. The conclusion still holds.
- **The probe does not implement §4.2's statement rule.** It compares statements after `.strip()`,
  not as exact bytes, so the header claim that it "implements draft 5 §4.2" is false for statements.
- **The Spec Kit state key is undefined.** `StateKey` is "every SDD root tree the profile names",
  but a Spec Kit profile names no root, only the glob `specs/*/spec.md`. The change manifest covers
  every file of the feature directory. So unless the root is derived from the whole glob prefix, an
  edit to `plan.md` alone would leave the state key unchanged, and the change intent's new revision
  would be missed (INV-18, INV-31). Spec Kit's key, "the feature directory name", is likewise
  undefined for a `**` glob.

## What holds

- **The design shape is right, and it is now coherent.**
  - Failure is per state.
  - Skips come from a closed list, and defects fail.
  - The git connector's own reader is used.
  - Collisions are checked per state.
- **INV-31, 33 and 44 are jointly satisfiable under §4.3.** `(id, digest)` stays unique across states,
  because the manifest includes the path.
- **De-duplication by state key** remains sound for OpenSpec (round 4's argument still applies),
  apart from F7's Spec Kit root.
- **Every §0 count I reproduced matches:** 1624/120/1, 338, 256/451, 23, 4 tips, and 401/1.
- **Retirement is clean.**
  - Every meaning change follows retire-and-reissue.
  - The kept rows changed only in wording, or gained stronger tests.
  - The tombstone reasons are back.
- **Every named kill I checked is reachable:**
  - INV-35(b) kills tip-fatal.
  - INV-39 kills unmasked parsing (3 against 1).
  - INV-49 kills the catch-all.
  - INV-47's oversize mutant now fails at `ValidateRevision`, inside the provider. It never reaches
    append, so the row's "(partial append)" describes the old mechanism, but the kill still occurs.
  - INV-48's colliding pair is findable offline. A birthday search over two same-slug keys is about
    2^16 work.
- **The neutral controls are neutral.**
  - `document-renames` yields 1 obligation both masked and unmasked, under OpenSpec and under the
    draft alike.
  - My `src/` edit survived.
- **Obligation `state = active` for REMOVED** is required by the kernel. `projections.go:364`
  rejects a target whose obligation is not active.
- **The profiles conform.** All 30 tracked profiles match the inlined §2.1 schema (`<scratch>/profiles.py`).
- **The pinned skip set is right.** `expected-skips.json` agrees with an independent scan of every
  fixture state.
- **The M3 split is honest.** `connector.go:198`, `projections.go:723` and
  `intent-delivery-applicability.sh:93` are correctly kept out of M1.
- **Build Laws.**
  - **Law 6:** spec before code. No provider code exists.
  - **Law 8:** no new dependency.
  - **Law 1:** `f06c95f` is an `@generated`, script-written pin.
  - **Law 2:** the intent SPEC §1 and the plan point at draft 5. The remaining drift is in F7.
  - **Law 9:** this round is non-author.

## Fitness for M1

**Not as written, but close.** No finding is HIGH, and none reshapes the contract. The three MED
findings are omissions that M1 code written to the letter would encode:

- F1 and F2 are two more members of the wedge class this series exists to close.
- F3 is a pinning claim that the cited code contradicts.

They can be fixed in draft 6, or directly in the `sdd/SPEC.md` that M1 commits, provided a non-author
re-verifies the changed sections. M1 itself should be reviewed at the foundational tier.

Conditions:

1. **Every tenant byte that `ValidateRevision`, the shared reader, the store or the projector can
   reject must be either a listed cause or provably unreachable** (F1, F2). That covers at least:
   - a whitespace-only or NUL-bearing statement;
   - a NUL in an identity value;
   - an over-long Spec Kit id;
   - a non-UTF-8 path, with an honest statement of which reader failures are tenant-caused;
   - a non-blob entry.

   INV-52 must project a fixture that carries those bytes.
2. **Pin §4.2 to the code it cites** (F3):
   - the reader's section split and title fold;
   - REMOVED bullets;
   - the pairing rule;
   - the regex dialect.

   Alternatively, drop the "pinned" claim and list each deliberate divergence.
3. **Correct §12's CLI rows, and decide whether `all` includes `sdd`,** stating what that does to
   repositories without a profile (F4).
4. **Make Spec Kit change-intent status declared, not inferred** (F5).
5. **Mechanise the fixture claims** (F6), fix the F7 drift, and **commit the fixture helper** before
   M1 builds on it.
