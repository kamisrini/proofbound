---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round6-20261117
status: NEEDS_WORK
reviewed_commit: 05b8e7dc5a1d7fe978e1d48eda74eed71bfff973
findings:
  - finding_id: r2-m0-review-round6-20261117-F1
    severity: MED
  - finding_id: r2-m0-review-round6-20261117-F2
    severity: MED
  - finding_id: r2-m0-review-round6-20261117-F3
    severity: LOW
  - finding_id: r2-m0-review-round6-20261117-F4
    severity: LOW
  - finding_id: r2-m0-review-round6-20261117-F5
    severity: LOW
  - finding_id: r2-m0-review-round6-20261117-F6
    severity: LOW
  - finding_id: r2-m0-review-round6-20261117-F7
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round6-20261117.md
artifact_sha: 8e72f2f634a0eaf8f0056a9d57b9fdf2e96690c6d968ee6655fbfc0034e303d9
---

# Release 2 M0 (SDD Warranty Connector): draft-6 derived-cause, parseDeltaSpec-pinned provider contract. Review round 6

## Provenance

- **Under review:** HEAD `05b8e7dc5a1d7fe978e1d48eda74eed71bfff973`.
  - `90dc1ac` introduced `docs/plans/R2-sdd-provider-spec-draft6.md` (SHA-256 `033212af…`), which
    supersedes draft 5, together with the testdata changes: the `fenced/fence-*` cases,
    `head-on-leftover/`, `expected-obligations.json`, the updated `expected-skips.json`, the Spec Kit
    profiles with `changes.status`, the READMEs, and the re-pinned `derive/history_probe.py`.
  - `4c8c838` committed `kernel/internal/connector/intent/sdd/fixture_test.go` (SHA-256 `5b850472…`),
    under `Intent: CI-sdd-warranty-m0b-ti5ga9` with the owner's `Authorized-By`.
    `scripts/intent-applicability.sh` reports it `applicable=true`, and `90dc1ac` and `05b8e7d`
    `applicable=false`.
  - `05b8e7d` is a one-line `@generated` pin refresh in `docs/verification/p6-connector-reality.md`,
    written by `scripts/p6-task2-close.sh --write`. It is not hand-written.
- **Out of scope:** `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`. Both are
  untracked and unchanged.
- **Follows up:**
  - `docs/verification/verdicts/r2-m0-review-round5-20261117.md` (NEEDS_WORK, F1..F7), read in full;
  - round 4, read for its frontmatter and the conditions that round 5 re-verified.
- **Owner decisions taken as given, not re-litigated:** `sync intent all` excludes `sdd`, and the
  helper is committed. I judged only whether the contract implements them consistently.
- **Judged against:** the CLAUDE.md Build Laws and this ground truth at HEAD:
  - **`intent.go`:** `New` (line 134), `Sync`, `ValidateRevision` (lines 214-278), `strict`,
    `empty`, `validateObligations`, `validateReferences`, `validateBatch` and `orderBatch`;
    `recordIDRE` and `obligationRE`.
  - **`core/core.go`:** `Canonicalize`, `inspectValue`, `NewEvent`, `Event.Validate` and
    `MaxPayloadBytes`, plus `core/SPEC.md` §2.1.
  - **`store/migrations/001_ledger.sql`.**
  - **`projections/projections.go`:**
    - the DDL for every intent view;
    - `eventRoute`, `reduce`, `reduceIntent` and `requireExactRecord`;
    - `Snapshot`, `readSnapshotRows` and `snapshotRowDigest`;
    - the `IntentRef` check at line 723.
  - **`projections/intent_report.go`** and **`gates/gates.go`** (lines 285 and 304).
  - **`cli/cli.go`:** lines 37, 53-75, 88-93, 236-262, 366-382, 513-528, 589-625 and 834-860.
  - **`connector/git/gitcmd/gitcmd.go`:** `Commits`, `Reachable`, `readCommit`, `files`,
    `refusePartial`, `validateRefs` and `run`. Also `connector/git/connector.go:185-200`.
  - **The intent providers:** `connector/intent/records/records.go` `parse`.
  - **The scripts:** `p6-route-matrix.sh`, `p6-census.sh`, `delivery-enforce.sh`,
    `intent-delivery-applicability.sh`, and `docs/releases/r1/seed-demo-app.sh`.
- **Real-history evidence:**
  - **Repository.** The author's clone of `Fission-AI/OpenSpec` at
    `f085c850f9440f497ccc76ecbc2859c09e0e0ac7`.
  - **OpenSpec's own reader, executed.** I did not port `parseDeltaSpec` for the reference side. I
    copied `requirement-blocks.ts`, `requirement-text.ts` and `code-fence.ts` verbatim into
    `<scratch>/osref/`. The only change is that their three relative imports are rewritten from
    `.js` to `.ts`. I ran them under a local Node.js v22.18.0 with built-in type stripping, through a
    10-line driver (`<scratch>/osref/run.ts`) that prints `parseDeltaSpec`'s
    `added`/`modified`/`removed`/`renamed`/`unpairedRenames`. This makes OpenSpec's actual code the
    reference, not a port of it.
  - **My draft-6 port.** `<scratch>/d6port/main.go` is a literal reading of draft 6 §4.2 under Go RE2,
    using the explicit `WS` / `DOT` classes and the ASCII-only fold. Its `MASK=naive|none` switch
    gives the two mutant masks.
  - **Probes.**
    - `<scratch>/hist.py`: per-state recounts.
    - `<scratch>/synth.py`: 48 synthetic delta files.
    - `<scratch>/blobs/`: all 401 distinct delta blobs, enumerated by `ls-tree -r` over all 1444
      distinct root trees.
    - I also ran the author's `history_probe.py` once, so its output could be compared with mine. I
      did not rely on it.
- **Synthetic evidence:**
  - `<scratch>/bdnul_probe_test.go`, injected into the intent package with `go test -overlay`.
  - `<scratch>/symrepo/`, a git repository with a symlinked SDD root and a tree at the profile path.
  - Throwaway module copies of the `sdd` package under `<scratch>/cal-*`.

  None of these wrote a repository file.
- **HEAD at dispatch and at completion:** `05b8e7dc5a1d7fe978e1d48eda74eed71bfff973`, unchanged.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. M1 owns record identity, point-in-time resolution and a new
  ledger-reaching source, so M1 itself should be reviewed at the foundational tier, as rounds 2-5
  recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m0-r6`, a **non-author**. It wrote
  none of the work under review, and it is not the reviewer of rounds 4 or 5. This verdict counts
  under Law 9.
- **Writes:** no repository file other than this verdict. No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m0r6/` in the session scratchpad. It occurs
    14 times in this verdict.
  - The OpenSpec clone is named only as "the author's clone", and the Node.js install only by its
    version.

## Round-5 findings re-verified

I checked against the text and the code, not against draft 6's §0 mapping table.

- **F1 (tenant bytes outside the closed list): FIXED for every `Revisions`-path check, with one
  state-level residual.**
  - **(a)** A blank Spec Kit statement → `statement-blank`. For OpenSpec, §4.2's "blank" is the union
    of `WS` and Go `unicode.IsSpace`. A chosen statement therefore always survives
    `validateObligations`' `TrimSpace`, and a whitespace-only line can never be chosen.
  - **(b)** An over-long Spec Kit id → `obligation-id-invalid`.
  - **(c)** The reader no longer calls `readCommit`/`files`, and SDD-INV-55 kills an implementation
    that does.
  - **(d)** A symlink or gitlink that a unit reads → `non-regular-file`.
  - Residual: the same entry kinds at the SDD **root** or the **profile path** are unplaced (new F3).
- **F2 (NUL passes to the projector): FIXED for `intent.sdd`, but the remedy wedges older ledgers.**
  - `nul-byte` is now a per-unit cause.
  - §12 adds a U+0000 rejection to `ValidateRevision` for every provider, and SDD-INV-52 projects a
    hostile history.
  - However, `reduceIntent` re-runs `ValidateRevision` over historical events on every rebuild. The
    amendment therefore turns an already-ledgered `records` event that is valid today into a
    permanent rebuild failure (new F2).
- **F3 (§4.2 is not OpenSpec's reader): FIXED, and verified against OpenSpec's own code.**
  - My literal draft-6 port and the real `parseDeltaSpec` agree on:
    - **0 of 401** real delta blobs;
    - **0 of 48** synthetic cases. These include all six rows of round 5's table, plus 42 more:
      - `##` followed by two spaces;
      - U+2028 inside a name and inside a title;
      - lone CR;
      - closing sequences after a tab and after an NBSP;
      - dotless ı and ſ;
      - a NEL, VT, ZWSP or mid-title BOM in a title;
      - one BOM and two BOMs;
      - an unclosed fence, an indented opener, a closer with info text, and an NBSP before a closer;
      - a backtick in a tilde fence's info string;
      - a 4-backtick fence with a 3-backtick inner line;
      - a backtick fence closed by tildes;
      - `*`/`+`/bare RENAMED lines, and lowercase `requirement:` in RENAMED;
      - a pending FROM across two RENAMED sections;
      - a fenced TO;
      - an HTML-commented header;
      - the REMOVED header form mixed with the bullet form;
      - repeated case-variant sections.
  - **Section fold.** The ASCII-only fold is equivalent to JavaScript's `toLowerCase` on the four
    titles, because no non-ASCII code point lower-cases to one of their ASCII letters. The only
    single-code-point ASCII lower-casing is K (U+212A) → k, and none of the titles contains a `k`.
  - **Header fold.** The ASCII-only `Requirement:` fold matches non-unicode `/i`.
  - **The labelled divergence.** The statement rule is honestly labelled as Proofbound's.
- **F4 (CLI rows; `all`): FIXED.**
  - `syncIntentOnStore` (`cli.go:366`) and the enum (`:58-60`) are listed. The parse is listed as
    "`:92`-area" (the actual lines are `:88-93`). The dispatch (`:513-514`) and the usage line
    (`:37`) are listed too.
  - `cli.go:247` is correctly moved to M3.
  - "`all` excludes `sdd`" is implemented consistently. `Sync("all")` iterates only the providers
    passed to `New` (`intent.go:154-159`), and §12 constructs `sdd` only when the selection is `sdd`.
    `verify` (`:835`, `:854`) and `sync all` (`:596`) pass `"all"`.
  - Two sites remain unlisted (new F5).
- **F5 (Spec Kit `accepted` fabricated): FIXED.**
  - `changes.status` is required for both kits (§2.1), and active changes take it (§5).
  - All **32** tracked fixture profiles conform to the §2.1 schema, checked by my independent script:
    29 OpenSpec and 3 Spec Kit, with no duplicate keys, the exact field sets, enums in range,
    `requirements.paths` free of `**` and under the root, and non-blank, NUL-free identities.
  - Archived → `accepted` remains a stated inference from OpenSpec's archive act, and is labelled as
    one.
- **F6 (self-tests weak): PARTIALLY FIXED.**
  - My calibration (below) confirms each claimed mechanisation: a renamed MODIFIED header fails, a
    bogus skip unit fails, a flipped `discriminates` fails, and a HEAD that is not dual-present
    fails. A neutral edit passes.
  - The `head-on-leftover` fixture kills a HEAD-fatal implementation.
  - The obligation lists are not derived from anything committed, and the self-test cannot detect a
    wrong pin (new F6).
  - My independent recomputation shows that every current pin is correct.
- **F7 (drift): PARTIALLY FIXED.**
  - Fixed: `unmappable/README.md`, the plan ("now draft 6"), the probe's statement comparison (exact
    line bytes), the probe's walk, and the Spec Kit roots (§2.1).
  - Still open: the top `testdata/README.md` table cites the counts that §0 corrected, among other
    items (new F7).

**Round-5 "Fitness for M1" conditions:**

1. **Every tenant byte rejected downstream is a listed cause or unreachable: MOSTLY MET.**
   - My own derivation (below) agrees with §7 for every check on the `Revisions` → append →
     projection path.
   - Residuals: state-level non-regular entries (F3), and the retroactive rebuild wedge created by
     the §12 remedy (F2).
2. **Pin §4.2 to the code it cites: MET.** It is verified against OpenSpec's executed reader.
3. **Correct §12's CLI rows and decide `all`: MET.** Minor omissions are in F5.
4. **Spec Kit change status declared: MET.**
5. **Mechanise the fixture claims, fix the drift, commit the helper: helper MET; the other two
   PARTIAL** (F6, F7).

## Evidence reproduced

All of these come from my own commands. The author's probe printed the same numbers when I ran it.

| §0 row | Author | Reviewer (own method) |
|---|---|---|
| Walk / merges / roots | 1624 / 120 / 1 | **1624 / 120 / 1** (`rev-list --all`, `--merges`, `--max-parents=0`) |
| Distinct SDD states | 338 | **338**: 337 distinct `openspec/changes` tree ids, plus "absent" (`<scratch>/hist.py`) |
| Changes observed active / archived | 135 / 108 | **135 / 108** |
| Archived, never observed active | 12 / 108 | **12** |
| Zero requirement entries at some state | 38 | **38** (active changes; entry counts from my draft-6 port) |
| Dual-presence commits | 23 (1 name) | **23**, all `adopt-verb-noun-cli-structure`; archived twice: 0 |
| `main` on the dual state | "9 commits" | **13** consecutive first-parent commits, `332816cc` … `acd1ca28`. Of these, **9** are non-merge and 4 are merges (F7) |
| Distinct delta blobs / changed by masking | 401 / 1 | **401** (`ls-tree -r` over 1444 distinct root trees; `rev-list --all --objects` alone yields 400, because one blob is first seen at another path) / **1**. All 509 delta entries are mode `100644` |
| Parser fidelity | 6 rows | **6 of 6**, plus 42 more synthetic cases and 401 real blobs. That is 0 differences between my draft-6 port and OpenSpec's executed `parseDeltaSpec`. The author's probe parser also matches on all 48 synthetic cases for names |

**`expected-obligations.json`.**
- **The `openspec_mask` lists** equal OpenSpec's executed reader for all 5 cases.
- **The `naive_toggle` lists** equal my port with "any `` ``` `` / `~~~` line toggles" on all 5
  cases.
- **The `unmasked` lists** equal my port with no mask on all 5 cases.

| Case | Mask | Naive | None |
|---|---|---|---|
| `improve-onboarding` | 3 | 3 | 1 |
| `fence-marker` | Alpha, Beta | Alpha, Hidden | Alpha, Hidden, Beta |
| `fence-length` | Alpha, Beta | Alpha, Hidden | Alpha, Hidden, Beta |
| `fence-info-backtick` | Alpha, Beta | Alpha | Alpha, Beta |
| `document-renames` (neutral) | 1 | 1 | 1 |

**Fixtures:** `go test -count=1 -v ./internal/connector/intent/sdd/` passes all three tests at HEAD,
under Go 1.27.4.

**Calibration**, on throwaway module copies of the `sdd` package:

| Control | Mutation | Expected | Result |
|---|---|---|---|
| Baseline | none | pass | passes |
| Positive 1 | rename `## MODIFIED Requirements` to `## Notes` in `improve-onboarding` | fail | fails: `lost its column-0 MODIFIED section header` |
| Positive 2 | change the `branches` skip unit to `change add-nope` | fail | fails: `does not name an existing unit` |
| Positive 3 | rename the step-02 archive entry of `head-on-leftover` to `…-add-q` | fail | fails: `HEAD is not on the dual-presence state` |
| Positive 4 | set `fence-marker` `discriminates` to false | fail | fails: `discriminates=false but lists say true` |
| Neutral 1 | append prose to the archived `add-w/proposal.md` in `head-on-leftover` | pass | passes |
| Neutral 2 | reword Alpha's statement in `fence-length` | pass | passes |
| Gap 1 | delete `fence-length`'s fence lines, so OpenSpec now yields `[Alpha]` against the pinned `[Alpha, Beta]` | — | **passes** (F6) |
| Gap 2 | set the `branches` skips to `[]`, and the `head-on-leftover` cause to `bogus-cause` | — | **passes** (F6) |
| Gap 3 | replace `fence-length`'s `naive_toggle` with `[[MODIFIED, Zzz]]` | — | **passes** (F6) |

### §7 re-derived: every check tenant bytes can reach

These are the checks an `intent.sdd` revision meets between the tree and a projected row.

- **`ValidateRevision`:**
  - the source and kind allowlists, which are provider-only;
  - `recordIDRE`. This is provider-only: the §4.3 slug is `[a-z0-9-]`, trimmed, 2-40 bytes, and the
    suffix is 6 characters of `[a-z0-9]`, so `BR-`/`CI-` + slug + `-` + suffix always matches
    `^(BD|BR|CI)-[a-z0-9][a-z0-9-]{1,62}-[a-z0-9]{6}$`;
  - `digestRE`;
  - `ArtifactPath != ""`. A tenant path is never empty, and git paths cannot hold NUL;
  - canonical equality;
  - `strict`;
  - the schema and status enums;
  - `DeclaredOwner == ""` / `DeclaredSponsor == ""`, which the profile rule covers;
  - `len(Obligations) == 0` / `len(Targets) == 0`. These are covered by "zero entries yields no
    record" and by INV-28;
  - `validateObligations`:
    - `obligationRE`. OpenSpec ids are at most `O-` + 3 + `-` + 28 = 32 characters after `O-`, by the
      §4.2 truncation. Spec Kit ids → `obligation-id-invalid`;
    - `TrimSpace`. Spec Kit → `statement-blank`; OpenSpec cannot select a blank statement;
    - the `active` state;
    - strict ordering, which is provider structure, with duplicates → `obligation-collision`;
  - `validateReferences`, which is provider structure.
- **`core.NewEvent` / `Event.Validate`:**
  - `NativeID` is at most 72 bytes of `[A-Za-z0-9-]`;
  - the payload bound → `payload-too-large`;
  - `jcs.Transform` cannot fail on Go-marshalled, valid-UTF-8 strings.
- **The ledger:**
  - `payload JSON` accepts any valid JSON, and `\u0000` only reaches it if the new rule is absent;
  - the only unique index is `(source, native_id, content_sha)`, whose columns are all bounded.
- **The projections:**
  - the `TEXT` columns `status`, `declared_owner`, `statement`, `state`, `declared_sponsor`,
    `relation` and the ids reject only U+0000 and invalid UTF-8, and both are now causes;
  - `artifact_path` lives only in the `payload JSON` columns;
  - **no primary key or index contains a variable-length tenant column**, so very long paths, names
    or statements cannot hit a btree row-size limit. They are bounded only by `payload-too-large`;
  - the target-obligation existence check requires `state='active'`, which every `sdd` obligation
    has;
  - `Snapshot` marshals rows with `json.Marshal` and `Canonicalize`, which is safe for valid,
    NUL-free strings.
- **The report and gates:**
  - `intent_report.go` prints statements with `%q`;
  - `gates.go` reads no tenant string.

**Conclusion:** for the `Revisions` → `Sync` → append → rebuild path, §7's list is complete. Nothing
is wrongly classed as provider-only. The record-id regex claim holds. `artifact_path` contents and
very long paths or names reach no unlisted failure. The two gaps lie outside that path: state-level
reads (F3) and replay of older events (F2).

## Findings

### r2-m0-review-round6-20261117-F1 (MED): §6 gives no outcome for a change that exists but has no change intent, so `Resolve` classifies a tenant state as a "reader failure"

**Broken property:**
- §6's error classification, and SDD-INV-50 ("Resolve distinguishes unknown, archived-earlier, and
  unmappable by sentinel"). Every outcome must be one of these classes; "anything else is a reader
  failure".
- §14 M3 entry condition 3, which relies on that split to report a bad claim per commit instead of
  failing the sync.

**Evidence:**
- **Resolve rule 1 has nothing to return.** It returns a change intent when "`<root>/<n>/` exists at
  `c` and `n` is mappable there". However, SDD-INV-28 and §5 say that a change with no requirement
  records has **no change intent and no Skip** ("not an error").
- **None of the errors applies.**
  - `ErrUnknownChange` requires "no active or archive entry names `n`", but the active directory
    does.
  - `ErrArchivedEarlier` does not apply.
  - `ErrUnmappable` requires `n` to be unmappable, and no §7 cause fires.
  - So by the text, `Resolve(c, n)` on a proposal-only change is a **reader failure**.
- **The same holds for an archive.** An archive entry archived by `c` whose change has zero
  requirement entries behaves the same way.
- **This state is common.** It is `openspec-history` step `02-draft-proposal-only`. On real history,
  **38 of 135** changes were in it at some state (reproduced above). A claim committed while a change
  is still proposal-only (scaffolding before deltas) is ordinary tenant behaviour.
- **The name domain is also unspecified.** Rule 1 tests "`<root>/<n>/` exists", but `n` is never
  constrained:
  - `n = "."` names `<root>/./`, which exists at every commit;
  - `n = "archive"` names the archive directory itself;
  - `n = "archive/2026-01-01-x"` names an archive entry through the active rule.

  None of these is a sentinel case either.

**Why it matters:**
- **M1 encodes it.** `Resolve` and SDD-INV-50 are M1's interface and test. An M1 written to the
  letter must class these as reader failures, which is the one class M3 is told to treat as fatal.
- **M3 inherits a wedge.** A tenant claim on a proposal-only change would then fail every git sync,
  which is the wedge class that §14 item 3 exists to prevent.

### r2-m0-review-round6-20261117-F2 (MED): the §12 U+0000 rule, placed in `ValidateRevision` "for every provider", makes already-ledgered `records` events fail every rebuild

**Broken property:**
- §12's claim that the rule "closes the latent partial-append and projection-wedge paths for
  `records` / `specdir` too".
- The append-only ledger's rebuildability (SDD-INV-52's "projects and rebuilds", applied to the
  whole ledger).

**Evidence:**
- **The projector re-validates history.** `reduceIntent` calls `connectorintent.ValidateRevision` on
  **every historical intent event during every rebuild** (`projections.go:311`, `:327` and `:350`).
  Tightening `ValidateRevision` is therefore retroactive.
- **A records event with U+0000 is valid today.** A `records` business decision whose `outcome`
  carries U+0000 passes every check at HEAD:
  - `<scratch>/bdnul_probe_test.go` builds the canonical payload
    `…"outcome":"ship\u0000it"…`, and `ValidateRevision` returns **`<nil>`**;
  - `records.go` `parse` checks only UTF-8 validity, which U+0000 passes;
  - `outcome` has no `TEXT` column: `business_decisions_view` stores it only inside `payload JSON`.
    Round 5 showed on PostgreSQL 16.4 that a `JSON` column accepts `\u0000`.

  Such an event is ledgered and projects cleanly at HEAD.
- **After M1 it can never project again.** Once M1 lands §12 as written, the next `rebuild`
  re-validates that event, fails, and the ledger cannot drop it.
- **Today's exposure.** No intent record in this repository carries `\u0000` today. I grepped every
  `proofbound-json` record. The hazard is to any existing ledger whose history ever held one, and to
  tenants.
- **No test sees it.** No §9 row or INT-INV-4 test replays an event that was ledgered before the
  tightening.

### r2-m0-review-round6-20261117-F3 (LOW): non-regular or non-tree entries at the SDD roots, the archive directory, or the profile path have no cause, and the two encoding causes have different scopes

**Broken property:**
- SDD-INV-35 ("No tenant content at any commit makes Revisions fail").
- §7's "Everything else fails `Revisions`".

**Evidence:**
- **The table only covers unit reads.** The `non-regular-file` row covers an entry "where a unit
  reads a file or directory", and SDD-INV-54 tests only entries *under* a root. The **state-level**
  reads are not unit reads:
  - `StateKey` reads the root tree ids;
  - `Read` reads the profile;
  - `List` reads `changes.archive`.
- **A literal implementation fails.** In `<scratch>/symrepo/`, with `openspec/changes` committed as a
  symlink and `.proofbound/sdd-profile.json` committed as a directory:
  - `git ls-tree -r HEAD:openspec/changes` → `fatal: not a tree object`;
  - `git cat-file blob HEAD:.proofbound/sdd-profile.json` → `fatal: … bad file`.

  An implementation that reads a root as `<c>:<root>`, or the profile as a blob, gets a reader
  failure. Under §7 that is a permanent `Revisions` failure.
- **Other implementations differ.** Implemented as `ls-tree -r <c> -- <root>`, the same state yields
  nothing. The contract does not choose.
- **Real repositories can have these shapes.** A symlinked `specs/` in a Spec Kit repository is a
  plausible tenant shape. A profile path that is a symlink has no stated cause (`profile-invalid` or
  `non-regular-file`), so SDD-INV-46's exact pin depends on the implementation.
- **The encoding causes have different scopes.**
  - `invalid-utf8` covers "a source file **or path under an SDD root**".
  - `nul-byte` covers "a file a unit **parses**".
  - Read literally, a binary asset (an image) in a change or feature directory skips the whole change
    intent as `invalid-utf8`, while a NUL-bearing binary would not.
  - This is latent: all 1069 distinct blobs under `openspec/changes` in the reference history are
    valid UTF-8. But the scope is unstated for Spec Kit feature directories.

### r2-m0-review-round6-20261117-F4 (LOW): residual spec-exactness gaps that let two conforming implementations mint different ids or causes

**Broken property:**
- SDD-INV-7 ("Record ids are deterministic") across implementations.
- SDD-INV-46's exact pins.
- The family's exact-bytes rule.

**Evidence:**
- **ASCII case mapping (§4.2 np, §4.3 slug).** "upper-case ASCII" / "lower-case ASCII" does not say
  whether non-ASCII letters are mapped first. Go's idiomatic calls do map them (`<scratch>/fold`):
  - `strings.ToUpper("expoſe") == "EXPOSE"` and `strings.ToUpper("ıd") == "ID"`, where the ASCII-only
    reading gives `EXPO-E` and `-D`;
  - `strings.ToLower("Key")` with a Kelvin sign gives `key`.

  The result is different obligation ids and slugs for the same bytes. testdata has no id vector yet,
  so nothing pins the choice.
- **Undated archive entries.** An archive entry that does not match `^[0-9]{4}-[0-9]{2}-[0-9]{2}-(.+)$`
  still matches the archive glob `openspec/changes/archive/*/specs/*/spec.md`. For such an entry,
  §4.3's key `<name>/<capability>`, §5's change membership, and name reuse are all undefined. Real
  history has none.
- **INV-53's row asks for unreachable cases.** It requires a blank-statement skip on the "OpenSpec
  statement, name fallback, RENAMED `TO`" routes. By §4.2's own rules none of these can be blank:
  - blank lines are never selected;
  - a name that survives `empty-name-part` contains `[A-Za-z0-9]`;
  - a `TO` line contains `TO:`.

  So a test written to the row cannot be built as stated.

### r2-m0-review-round6-20261117-F5 (LOW): §12's "complete list" still omits sites M1 must touch

**Broken property:** §12 "complete list"; Law 6 (the interface stays in the spec); §7's "The CLI
prints them and exits 0".

**Evidence:**
- **The core interface lock.** `kernel/internal/core/SPEC.md:75-83` lists the `Source` constants in
  core's interface lock, including `SourceIntentRecords` and `SourceIntentSpecdir`. §12 adds
  `SourceIntentSDD` to `core.go:56` but not to core's SPEC. The §12 grep itself prints
  `core/SPEC.md:82`.
- **Printing the Skips.** §7 makes the CLI print Skips and exit 0, but the only print site
  (`cli.go:525`, `listed=… appended=… existing=…`) is not listed, and no §9 row tests the printing
  or the exit code.
- **The user guide.** `docs/releases/r1/USER-GUIDE.md:269` documents
  `sync intent {records|specdir|all}`.
- **A loose description.** The "unchanged on purpose" list says `delivery-enforce.sh:38` calls
  `sync intent all`. It calls `sync all`, which reaches `"all"` through `cli.go:596`. The conclusion
  is unaffected.

### r2-m0-review-round6-20261117-F6 (LOW): the fenced-case obligation lists cannot be regenerated from committed code, and the self-test cannot see a wrong pin

**Broken property:**
- §13's "The lists were computed by the probe's §4.2 parser".
- Law 1's spirit for derived pins.
- The LESSON class fixture-not-what-it-claims.

**Evidence:**
- **The probe cannot produce one of the lists.** `history_probe.py`'s `parse_delta` has only
  `masked=True/False`, and "naive toggle" is defined nowhere in the draft or the probe. The
  `naive_toggle` lists therefore have no committed derivation.
- **The current pins are right.** I re-derived them with an explicit definition, and they are correct
  today (table above).
- **The self-test passes with wrong pins.**
  - **Gap 1:** a `fence-length` edit makes OpenSpec yield `[Alpha]` against the pinned
    `[Alpha, Beta]`.
  - **Gap 3:** a bogus `naive_toggle` list.

  M1's SDD-INV-39 test will catch a stale `openspec_mask`. Nothing will catch a stale `naive_toggle`
  or `unmasked` list, yet those lists are the evidence for the "kills a naive toggle" claim.
- **Skip pins.** `TestPinnedExpectationsMatchFixtures` checks only that a skip's unit exists and that
  its cause is non-empty (Gap 2: `bogus-cause` passes). The draft claims no more than that, and M1's
  SDD-INV-46 compares exact sets, so this is noted, not counted.

### r2-m0-review-round6-20261117-F7 (LOW): residual drift and evidence labels

**Evidence:**
- **The README table still has the corrected counts.** `testdata/README.md`'s table still cites:
  - "34 real" against §0's 38;
  - "47 / 10 real" against 83 / 39;
  - "16 / 108 real" against 12 / 108.

  These are the path-limited values that round-5 F7 named and that §0 says are fixed.
- **The probe docstring is stale.** It still says it "Walks every commit that touches
  openspec/changes".
- **The probe's statement rule differs from §4.2 in one case.** In a REMOVED section, a bullet line
  after a header block ends that block in the probe (`close(); entry = None`). §4.2 does not end a
  block there, so the bullet line becomes the header's statement.
  - Synthetic `removed-header-and-bullet`: the probe gives `(REMOVED, A, "A")`, and the draft gives
    `(REMOVED, A, "- ### Requirement: B")`.
  - The probe's header still claims it "implements draft 6 §4.2". On the 401 real blobs the two agree.
- **"9 commits" is ambiguous.** "`main` itself sat on this state for 9 commits" is 13 consecutive
  first-parent commits, of which 9 are non-merge. The conclusion holds either way.

## What holds

- **§4.2 now *is* OpenSpec's reader for names, sections, REMOVED bullets, RENAMED pairing and fence
  masking.** This is established against OpenSpec's executed source, not a port of it, over 401 real
  and 48 adversarial inputs. The explicit `WS`, `DOT`, ASCII-fold, `trim` and closing-sequence rules
  are each necessary and sufficient on that corpus.
- **The §7 derivation is right for the `Revisions` path.** Every check that tenant bytes can reach
  through `ValidateRevision`, `Event.Validate`, the ledger, the projection DDL, the snapshot code, the
  reports and the gates is either listed or provider-only. No index holds a variable-length tenant
  column.
- **§3 keeps SDD-INV-31 and SDD-INV-51.**
  - The shared listing is the existing `refusePartial` + `validateRefs` + `rev-list` prefix of
    `gitcmd.Commits`. `gitcmd.Reachable` already reuses that prefix.
  - All calls go through `run` (`--no-replace-objects`, the `-c` options, the scrubbed environment).
  - `Resolve` and `Revisions` read the same state per commit.
  - The SDD roots cover every file a unit reads, for both kits, so equal state keys imply equal
    inputs.
- **`Artifact.Regular` is coherent with the manifest.** A non-regular entry is never hashed, because
  its unit is skipped. A mode-only change alters the state key but not the manifest, and
  de-duplication absorbs it.
- **The owner decisions are implemented consistently.**
  - `all` excludes `sdd` by construction in `syncIntentOnStore`, and SDD-INV-56's kill (adding `sdd`
    to `all`) is reachable through `sync intent all` on a shallow clone.
  - The helper is committed under an Intent chain with `Authorized-By`.
- **The M1 / M3 split is right.** `connector.go:198`, `cli.go:247`, `projections.go:723` and
  `intent-delivery-applicability.sh:93` are the claim-resolution sites, and M1 leaves them alone.
  The re-run grep finds no other registration site, apart from the documentation and spec sites in
  F5.
- **The §9 numbering is clean.**
  - INV-1..52 are unchanged from draft 5, apart from the INV-35, -39 and -52 test rows. Their
    statements are identical and their tests are only strengthened.
  - INV-53..58 are appended.
  - No meaning change hides under a kept number.
- **The named kills are reachable:**
  - HEAD-fatal (`head-on-leftover`, confirmed by Positive 3);
  - naive toggle (all three precision cases, re-derived);
  - per-commit-diff reader (INV-55);
  - Go `\s` / `(?i)` (the NBSP and ſ rows);
  - an id that `ValidateRevision` rejects (INV-58).
- **The neutral controls are neutral:** `document-renames` gives 1/1/1, and my neutral edits passed.
- **Every §0 count I re-measured reproduces exactly.** The exception is the "9" label (F7).
- **Build Laws:**
  - **Law 1:** `05b8e7d` is an `@generated`, script-written pin.
  - **Law 6:** no provider code exists; the `sdd` package holds only `fixture_test.go`.
  - **Law 8:** no new dependency; the helper uses only the standard library.
  - **Law 9:** this round is non-author.
  - **Law 2:** the plan and the intent SPEC §1 point at draft 6. The residual drift is in F7.

## Fitness for M1

**Not as written, but close.** No finding is HIGH, and none reshapes the contract. The parser
question that dominated rounds 3-5 is settled. The two MED findings are narrow, and M1 code written
to the letter would encode them:

- F1 is a hole in `Resolve`'s closed classification, on a common tenant state.
- F2 is a retroactive rebuild wedge created by the round-5 remedy itself.

They can be fixed in draft 7, or directly in the `sdd/SPEC.md` that M1 commits, provided a non-author
re-verifies the changed sections. M1 itself should be reviewed at the foundational tier.

Conditions:

1. **Give every `Resolve` input a defined outcome** (F1). That covers:
   - a change that exists with no change intent, whether active or archived by `c`;
   - names that are not single path segments, or that name the archive directory.

   SDD-INV-50 must test each.
2. **State how the U+0000 rule meets events already in the ledger** (F2), and test a rebuild over an
   event ledgered before the rule. This holds wherever the rule lives (`Sync` only, `ValidateRevision`
   at replay, or elsewhere).
3. **Place every state-level read**: the root, the archive directory and the profile path, as a
   symlink, gitlink, file or tree. **Give the two encoding causes one stated scope** (F3).
4. **Fix the case-mapping reading and pin it with a non-ASCII id vector**, and define undated archive
   entries (F4).
5. **Complete §12** with `core/SPEC.md` and the Skip print site and its test (F5). **Make the
   naive/unmasked pins derivable from committed code** (F6), and fix the F7 drift.
