---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round4-20261117
status: NEEDS_WORK
reviewed_commit: 45a3bf8f8430c0802ddc5e31bb9840d1e2f7c8ed
findings:
  - finding_id: r2-m0-review-round4-20261117-F1
    severity: MED
  - finding_id: r2-m0-review-round4-20261117-F2
    severity: MED
  - finding_id: r2-m0-review-round4-20261117-F3
    severity: MED
  - finding_id: r2-m0-review-round4-20261117-F4
    severity: MED
  - finding_id: r2-m0-review-round4-20261117-F5
    severity: MED
  - finding_id: r2-m0-review-round4-20261117-F6
    severity: MED
  - finding_id: r2-m0-review-round4-20261117-F7
    severity: LOW
  - finding_id: r2-m0-review-round4-20261117-F8
    severity: LOW
  - finding_id: r2-m0-review-round4-20261117-F9
    severity: LOW
  - finding_id: r2-m0-review-round4-20261117-F10
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round4-20261117.md
artifact_sha: 7a978b13e753210b55e43aa00bcbf12300f8209e23ef413ec9531c8d7979c079
---

# Release 2 M0 (SDD Warranty Connector): draft-4 per-state, history-native provider contract. Review round 4

## Provenance

- **Under review:** HEAD `45a3bf8f8430c0802ddc5e31bb9840d1e2f7c8ed`. The content was introduced by
  `23e7246`, and `45a3bf8` is an `@generated` gate-pin refresh. Files:
  - `docs/plans/R2-sdd-provider-spec-draft4.md`, which supersedes draft 3.
  - `kernel/internal/connector/intent/sdd/testdata/`:
    - the `branches/` commit-graph fixture and the `fenced/` fixture;
    - `README.md` and the extended `derive/history_probe.py`;
    - the 29 newly tracked `.proofbound/sdd-profile.json` profiles, and the `.gitignore` rule that
      re-includes them.
  - The intent SPEC §1 pointer change.
  - The M0 status and M1 text in `docs/plans/sdd-warranty-connector-plan.md`.
- **Also reviewed: `kernel/internal/connector/intent/sdd/fixture_test.go`, as uncommitted
  working-tree content.** The owner decided to include it. It contains the replay helper's
  PARENTS/REF support, `TestBranchFixtureReplaysAsDocumented`, and `requireProfiles`. The committed
  `HEAD` version has none of these (see F10).
- **Out of scope:** `docs/plans/P5-intent-provenance-plan.md` and `kernel/cmd/p6dbprobe/`, which are
  untracked. `notes/state.md` is modified in the working tree by the owning session. I did not read or
  review that change.
- **Follows up:** `docs/verification/verdicts/r2-m0-review-round3-20261116.md` (NEEDS_WORK, F1..F5).
- **Judged against:** the CLAUDE.md Build Laws and this ground truth:
  - `intent/SPEC.md` and `intent.go`:
    - `Sync`, at lines 145-211;
    - `ValidateRevision`, at lines 214-278, whose source allowlist is at 215;
    - `validateBatch` and `orderBatch`;
    - `recordIDRE` and `obligationRE`;
    - `New`'s provider allowlist, at line 134.
  - `core/core.go`:
    - `Canonicalize`;
    - `NewEvent` and `Event.Validate`, with `MaxPayloadBytes` at line 64 and the payload check at
      line 322.
  - `connector/git/connector.go` (`resolvedCommit`, lines 180-214).
  - `git/gitcmd/gitcmd.go`:
    - `Commits`, at line 63;
    - `refusePartial` and `validateRefs`;
    - `run`, which passes `--no-replace-objects` at line 404.
  - `projections/projections.go`:
    - the commit-intent binding, at lines 216-227;
    - `eventRoute` and `supported`, at line 296.
  - `gates/gates.go`, line 304.
  - `BR-sdd-warranty-i9miv7/0002.md`.
- **Real-history evidence:**
  - **Repository.** The author's clone of `Fission-AI/OpenSpec` at `f085c850`, in the session
    scratchpad. It is not shallow, and its refs are 1 local head, 98+ remote-tracking branches, and
    74 tags.
  - **My probes.** I wrote my own probe (`<scratch>/probe.py`) and two focused scripts
    (`<scratch>/fence2.py` and `<scratch>/inv39.py`, the latter a masked/unmasked delta parser using
    OpenSpec's own regexes). I read the author's probe but did not reuse its output.
  - **Hand checks.** Every quoted example was checked with `git show`, `git ls-tree` or
    `git rev-list`.
- **Synthetic evidence:**
  - **Replace-object walk.** A three-commit repository with a replace-object graft
    (`<scratch>/replace`) shows the walk divergence in F4.
  - **Oversize payload.** A probe test of the existing intent connector (F1) was injected with
    `go test -overlay` from `<scratch>/oversize_probe_test.go`. It wrote no repository file.
  - **Fixture calibration.** I copied the fixture package into four throwaway modules under
    `<scratch>/cal-*`.
- **HEAD at dispatch and at completion:** `45a3bf8f8430c0802ddc5e31bb9840d1e2f7c8ed`, unchanged.
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. M1 owns record identity and point-in-time resolution, so M1 itself
  should be reviewed at the foundational tier, as rounds 2 and 3 recommended.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261117-m0-r4`, a **non-author**. It wrote
  none of the work under review, so this verdict counts under Law 9.
- **Writes:** no repository file other than this verdict. No cleanroom redactions were needed.
  - `<scratch>` stands for the reviewer directory `reviewer-m0r4/` in the session scratchpad. It
    occurs 11 times in this verdict (7 in this section).
  - The OpenSpec clone is named only as "the author's clone".

## Round-3 findings re-verified

- **F1 (past fail-closed commits wedge every sync): FIXED IN DESIGN, with an incomplete
  enumeration.**
  - §7 makes every listed content condition per unit, reported as a `Skip`.
  - The reference repository's `7ced2a87` leftover no longer fails `Revisions`.
  - On real history the rule removes exactly one CI unit-state: the `adopt-verb-noun-cli-structure`
    dual presence.
  - Residual: §7's claim "tenant content can never trip the intent connector" is false for one
    condition it does not list, a payload over `core.MaxPayloadBytes`. That condition still wedges
    every sync, and it partially appends first (new F1).
  - Residual: the rejected "fail at a ref tip" alternative survives every listed test (new F6).
- **F2 (walk differs from the git connector's walk): FIXED for the argument list.**
  - §3 adopts the unsimplified `--all` list, and §12 gives it one exported home.
  - The branches fixture really separates the walks. Its self-test asserts that a path-limited walk
    from `main` misses steps 03 and 05 while `--all --full-history` keeps them. Step 09 is reachable
    only through `--all`.
  - Residual: "one home" covers only the `rev-list` arguments. It does not cover the git connector's
    read environment (`--no-replace-objects`) or its preconditions (shallow and graft refusal), so
    SDD-INV-31's "`Commits` equals the git connector's list" can still diverge (new F4).
- **F3 (O-11 unsatisfiable): FIXED.**
  - O-11 is withdrawn from the BR 0003 proposal, with the reason stated.
  - It is not reissued as a deferral row that could resurface it silently.
- **F4 (tests that cannot fail; titles that changed meaning): PARTIALLY FIXED.** Fixed:
  - INV-8, 11, 24 and 27 are retired and reissued as INV-34, 33, 31 and 32.
  - INV-26 gains its neutral direction (b).
  - INV-13 gains a mechanical form (b).
  - INV-32 covers the merge, root and dual-presence cases.

  New residuals:
  - INV-39's positive control does not die (F5).
  - INV-35 does not kill the tip-fatal implementation (F6).
  - INV-37's named kill cannot be reached under §7 (F2).
  - Three kept invariants changed meaning in place (F6).
- **F5 (drift): MOSTLY FIXED.** Fixed:
  - `slug` and `suffix` are now defined in the draft, with the `kit` and `kind` literals.
  - An empty name-part is unmappable.
  - RENAMED is read as `FROM`/`TO` pairs.
  - The plan's M1 text names `Resolve(ctx, commit, name)`.
  - §12 names the §4 amendment.

  Residuals:
  - The profile schema exists only in superseded draft 2, although draft 4 calls itself
    self-contained.
  - The obligation `state` value is unspecified.
  - The testdata README still cites the retired INV-14, 24 and 27 (F9).

**Round-3 "Fitness for M1" conditions:**

1. **Keep `sdd` out of the git connector: MET.** §1 and §12 say so explicitly.
2. **Test on real-shaped histories, with every probe count printed: MOSTLY MET.**
   - The fixture covers merge, TREESAME merge, unmerged claim, leftover-after-archive and root. Each
     shape is asserted by a self-test that I calibrated.
   - Two §0 counts are measured on a walk the draft itself rejects, or on a population other than
     the one the row describes (F8).
   - The fenced fixture is the wrong shape (F5).
3. **Amend intent SPEC §4 together with §1: MET as named.** The amendment is named, not yet made,
   which is acceptable at M0. The list it belongs to is incomplete (F3).
4. **Keep O-11 out of the BR: MET.**

## Evidence reproduced

My own probe and git commands give these results.

| §0 row | Author | Reviewer |
|---|---|---|
| Commits on `--all` (merges, roots) | 1624 (120, 1) | 1624 (120, 1) |
| Distinct SDD states | 338 | 338 (337 non-empty, plus the pre-OpenSpec "absent" state) |
| Touching `openspec/changes`: HEAD path-limited / `--all --full-history` | 256 / 451 | 256 / 451 |
| Dual presence: names; commits (HEAD path-limited / `--all --full-history`) | 1; 9 / 12 | 1 (`adopt-verb-noun-cli-structure`); 9 / 12 |
| **Dual presence on the provider's own walk (unsimplified `--all`)** | not stated | **23 commits**, 4 of them remote-branch tips (F8) |
| Archived twice at one commit | 0 | 0 |
| Tags at dual-presence commits | 0 | 0 tags. **4 branch tips:** `origin/archive-changes`, `origin/prepare-publish`, `origin/update-openspec-agent-instructions`, `origin/updating-agent-instructions` |
| RENAMED entries outside fences | 0 | 0 RENAMED pairs over all 401 distinct delta blobs, using OpenSpec's rules |
| Fenced `- FROM:` lines | "47 over the walk" | 8 over distinct blobs; 201 over the HEAD path-limited walk; 1534 over `--all` (all delta files). The author's 47 counts `- FROM:` anywhere in *active* deltas on the path-limited walk (F8) |

Other measurements:

- **Archive directory names.** Every one has a `YYYY-MM-DD-` prefix. No archive directory matches
  §6's glob `*-<n>` for a name other than its own, so F9's ambiguity has no instance on this history.
- **"Archived by", per §6.** "Absent from every parent" yields 150 events. "Absent from the first
  parent" yields 266. A first-parent rule would therefore attribute 116 archivings to merges, which
  supports §6's rule. The root commit `8614f7a4` carries 0 archives.
- **Code fences.** OpenSpec's fence rule (marker and length match; no backtick in the info string)
  and a naive toggle disagree on 0 blobs.
  - Exactly one distinct delta blob in the whole history yields different obligations when parsed
    masked and unmasked: `archive/2025-09-29-improve-init-onboarding/specs/cli-init/spec.md`
    (4 against 2; see F5).
  - The fenced `## RENAMED Requirements` example in `adopt-delta-based-changes` is **indented by two
    spaces** in every revision of that file.
- **Fixtures.**
  - `go test -count=1 ./internal/connector/intent/sdd/` passes, running both self-tests.
  - All 29 fixture profiles are tracked (`git ls-files`).
  - `git check-ignore` confirms that the re-include rule works.
- **Calibration of the fixture self-tests**, run on throwaway copies:
  - **Positive control: must fail, and fails.** Making step 05 keep step 03's delta bytes, so that
    06 is no longer TREESAME to 02, gives `step 06 is not TREESAME to 02 on openspec/`.
  - **Positive control 2: must fail, and fails.** Deleting step 09's `.proofbound/` gives `replayed
    commit 9 has no .proofbound/sdd-profile.json`.
  - **Neutral control: must survive, and survives.** Appending a line to step 04's `src/x.ts` leaves
    the tests passing.
  - **Unmodified baseline:** passes.

## Findings

### r2-m0-review-round4-20261117-F1 (MED): §7's "batch-safe by construction" is false. A tenant payload over `core.MaxPayloadBytes` passes every check the draft names, then fails `Sync` after a partial append, permanently

**Broken property:** SDD-INV-35 (in effect), and §7's claims "tenant content can never trip the
intent connector's batch check" and "INT-INV-4 holds unchanged". This is the same permanent-wedge
class as round-3 F1.

**Evidence:**
- **The gap.** `ValidateRevision` does not bound payload size. The bound lives in
  `core.Event.Validate` (`core.go:322`), which `Sync` reaches only inside the append loop
  (`intent.go:191`), after `validateBatch` and `orderBatch`.
- **The probe.** `<scratch>/oversize_probe_test.go` builds a canonical requirement with one 1 MiB+1
  statement. It was run against the unmodified intent package via `-overlay`. Result:
  - `ValidateRevision(big)` returns nil.
  - `Sync` returns `intent connector: BR-zz-zzzzzz: payload: must be a non-empty object`, with
    **1 event already appended**.
- **Why tenant content reaches it.**
  - An OpenSpec statement is "the first non-empty line, exact bytes", so one long line in a delta is
    enough.
  - A change intent's `targets` grows with every requirement and obligation of the change.
- **Why it is permanent.** Under history-native ingestion, one such commit anywhere on any ref is in
  every future batch.
- **It is missing from the cause list.** The §7 list of per-unit causes does not include it, and
  neither does any fixture.
- **A related unstated condition.** Go's `encoding/json` silently replaces invalid UTF-8 with U+FFFD
  when marshalling. That contradicts §4.2's "exact bytes" statement with no unmappable rule. Payload
  bytes would then depend on the encoder, not on the commit.

### r2-m0-review-round4-20261117-F2 (MED): §7 turns every `ValidateRevision` failure of a candidate into a tenant `Skip`, so provider defects become silent exit-0 skips, and SDD-INV-37's named kill cannot be reached

**Broken property:**
- §7's own sentence "If the batch check does fire on `sdd` revisions, that is a provider bug and
  fails the sync".
- SDD-INV-37's row ("kills: returning a revision that fails ValidateRevision").
- The Law 9 corollary: a test whose positive control cannot die.

**Evidence:**
- **The catch-all.** §7's cause list ends with "a candidate revision that fails
  `intent.ValidateRevision`". Most of what `ValidateRevision` checks is provider-produced structure,
  which no tenant byte controls:
  - canonical JSON;
  - `strict` decoding;
  - strictly sorted obligations;
  - sorted unique references with sorted obligation ids;
  - source and kind registry;
  - envelope-to-payload binding.
- **What the catch-all does to a provider bug.** An implementation that, for example, emits
  obligations in map order does not fail. It skips every affected record. Every dependent change
  intent then cascades, `Revisions` returns fewer revisions, and the CLI exits 0 (§7).
- **The INV-37 kill is unreachable.** INV-37's mutant ("returning a revision that fails
  ValidateRevision") cannot occur in a conforming implementation, because §7 requires pre-validating
  and skipping. The row's positive control therefore never dies, and the class it guards is now
  masked rather than caught.
- **Nothing pins the clean case.** No invariant states that the reference fixtures (the
  openspec-history, speckit-history and branches histories) yield **exactly** their expected `Skip`
  set, which for the clean histories is empty except for the step 11-12 dual presence. Without that,
  a partial-skip provider bug can survive any test that only checks the revisions it does find.

### r2-m0-review-round4-20261117-F3 (MED): §12's "amendments that land with M1" omits the kernel registries that today reject `intent.sdd`. One omission degrades into F2's silent skips; another, once events are ledgered, fails the projector permanently

**Broken property:** §12's role as the complete list of contract changes M1 must make, and the intent
SPEC's INT-INV-1 and INT-INV-12.

**Evidence.** §12 names intent SPEC §1, §2 and §4, plus `gitcmd`. It does not name these:
- **`intent.go:215`.** `ValidateRevision` rejects any source other than `intent.records` and
  `intent.specdir`. Combined with §7's catch-all (F2), every `sdd` candidate would become a `Skip`,
  and the sync would exit 0 with nothing emitted.
- **`intent.go:134`.** `New` rejects a provider named other than `records` or `specdir`.
  INT-INV-1 and INT-INV-12 are the invariants affected.
- **`projections.go:296`.** `eventRoute` routes intent kinds only for those two sources, and
  `supported()` returns `unsupported event` (`projections.go:200-201`) for anything else.
  - Once `sync intent sdd` appends `intent.sdd` events, the append-only ledger holds events that the
    projector rejects on every rebuild.
  - The plan's M1 deliverable is provider conformance. It contains no projection test that would
    notice this.
- **`gates.go:304`.** The `intent-reference-integrity` matcher lists only those two sources.
- **The family STOP meaning.** Intent SPEC §6 and INT-INV-11 define the meaning of
  `ErrUnmappable`: a STOP when the canonical model cannot represent the source. §7 now wraps tenant
  *data* defects (an uncovered owner, an unpaired line) in the same sentinel and delivers them as
  skips. The draft amends neither the family sentence nor the invariant, nor does it state that they
  are unchanged for `sdd`.

The plan's M1 line names `core.Source` and "target grammars". That is not the list above, and the
draft is the contract.

### r2-m0-review-round4-20261117-F4 (MED): the walk's "one home" is only the `rev-list` argument list. The git connector's walk also depends on `--no-replace-objects`, shallow and graft refusal, and ref validation, so SDD-INV-31's equality can fail and §6's root rule fabricates "archived by"

**Broken property:**
- SDD-INV-31: `Commits` equals the git connector's list.
- §3's "closes round-3 F2 by construction".
- §6's root case.

**Evidence:**
- **What the git connector actually does.** It never runs bare `git rev-list`:
  - `gitcmd.run` prefixes every call with `--no-replace-objects`, extra `-c` options, and a scrubbed
    `GIT_*` environment (`gitcmd.go:404-406`);
  - `Commits` first calls `refusePartial`, which refuses shallow repositories and grafts, and then
    `validateRefs`.

  §3 and §12 move only "the `rev-list` argument list" to one home.
- **Reproduction** (`<scratch>/replace`): three linear commits plus `git replace --graft` on the
  middle commit, run with the identical shared arguments.
  - The default reader lists 2 commits, and the middle commit has no parents.
  - With `gitcmd`'s base options, it lists 3, and the middle commit has its real parent.
- **What this breaks.**
  - The commit sets differ, so INV-31 is false.
  - The grafted commit becomes a **root**. Under §6's root rule, a root "archives" every archive
    directory it carries, so `Resolve` succeeds for every archived change at that commit.
  - A shallow clone, which the git connector refuses (`ErrShallow`), produces the same fabrication at
    its boundary commit if the `sdd` reader does not refuse it too.
- **The tests cannot see it.** The INV-31 test runs only over a fixture with no replace refs and no
  shallow boundary, so it cannot detect this.

### r2-m0-review-round4-20261117-F5 (MED): SDD-INV-39's fixture cannot discriminate a masked parser from an unmasked one, its "fence-removed" control is false under §4.2's own header rule, and §0's rationale for masking is wrong about the real case

**Broken property:**
- The INV-39 row: "kills: an unmasked parser; the fence-removed variant proves the case discriminates".
- §0's and §4.2's claim that the fenced RENAMED example, "read unmasked, would mint rename
  obligations nobody declared".

**Evidence:**
- **The fixture, and the real example, are indented.** In `testdata/fenced/…/conventions/spec.md`,
  and in every real revision of `adopt-delta-based-changes/specs/openspec-conventions/spec.md`
  (checked with `git show` across the walk), the fenced example's header is `  ## RENAMED
  Requirements`, indented by two spaces.
- **That indentation makes masking irrelevant.** §4.2 adopts OpenSpec's header rule. OpenSpec
  splits top-level sections only on `^##\s+` at column 0 (`requirement-blocks.ts:75` and `:491`;
  `DELTA_HEADER` in `spec-structure.ts`). An indented line is never a section header, masked or not.
- **Measured** (`<scratch>/inv39.py`, using OpenSpec's regexes) on the fixture:

  | Parse | Obligations |
  |---|---|
  | Masked | `[MODIFIED Rename Format]` |
  | Unmasked | `[MODIFIED Rename Format]` |
  | Fence lines removed | `[MODIFIED Rename Format]` |

  An unmasked implementation therefore passes INV-39 as specified. A correct implementation
  *fails* the row's "removing the fence lines yields a REN obligation" control. Only a parser with a
  looser, non-OpenSpec header rule satisfies that control.
- **Where masking really matters.** Over all 401 distinct real delta blobs, exactly one changes under
  masking: `archive/2025-09-29-improve-init-onboarding/specs/cli-init/spec.md`.
  - A **column-0** `## TL;DR Quick Checklist` sits inside a fence at line 37.
  - Unmasked, that line ends the MODIFIED section, so the requirements "Interactive Mode" and
    "Success Output" are silently dropped (4 obligations masked, 2 unmasked).

  That is under-mapping, not fabrication. It is neither the fixture shape nor the shape §0 cites.
- **The fence rule is not specified to OpenSpec's precision.** §4.2 says "from an opening ``` or
  ~~~ fence through its closing fence". OpenSpec's `buildCodeFenceMask` additionally requires:
  - that the closing fence uses the same marker and is at least as long as the opener;
  - that a backtick fence's info string contains no backtick.
- **Section termination is not specified.** §4.2 also never says what ends a delta section, namely a
  non-delta `## ` header. Two conforming implementations can therefore disagree on the one real
  blob where masking matters.

### r2-m0-review-round4-20261117-F6 (MED): the alternative rejected in §10 D survives every test, and three kept invariants changed meaning in place, contrary to §9's own retire-and-reissue rule

**Broken property:**
- The Law 9 corollary: decision D has no killing test.
- §9: "an invariant whose meaning changes is retired and re-issued under a new number".

**Evidence:**
- **The tip-fatal implementation survives.**
  - §10 D rejects "fail only when the bad state sits at a ref tip".
  - INV-35's test places each cause "at a non-tip commit and fixed later". §13's fail-closed tables
    say the same.
  - In the branches fixture, every ref tip is clean: `main` at 13, `feature` at 05, `side` at 07,
    `wip` at 09.
  - A tip-fatal implementation therefore passes INV-35, INV-36 and INV-20.
- **The case is not hypothetical.** Real OpenSpec has **four remote-branch tips at dual-presence
  commits** (Evidence). A tip-fatal implementation would fail every sync of the reference repository
  today.
- **Kept numbers whose meaning inverted.**
  - In draft 3, INV-1 (strict profile, "rejected"), INV-6 ("ErrUnmappable") and INV-29 ("a reused
    name is ErrUnmappable") meant sync-fatal, through INV-14 "stops". Round 3 F1's evidence was
    exactly INV-29's test asserting that failure.
  - Draft 4 keeps all three numbers and titles but changes their statements to "that state
    unmappable", and for INV-29 "for that change at that state only".
  - A test written to the draft-3 row expects `Revisions` to error. The draft-4 row requires it not
    to. That is a changed meaning under a kept number, which is the pattern round 3 F4 flagged for
    INV-8 and INV-11.
- **Tombstone reasons were deleted in place.** The tombstones of INV-15, 16, 17, 19, 21, 22 and 23
  lost their retirement reasons. For example, "RETIRED (draft 3): archived changes are now ingested"
  became "RETIRED (draft 3)." That contradicts §9's statement that the numbered list *is* the
  tombstone, and it contradicts the draft's self-contained claim.

### r2-m0-review-round4-20261117-F7 (LOW): SDD-INV-38's cross-state collision rule contradicts SDD-INV-33's read set and SDD-INV-31's membership guarantee

**Broken property:** the joint satisfiability of INV-31, INV-33 and INV-36 together with INV-38.

**Evidence:**
- **The cross-state dependency.** INV-38 makes **both** colliding states unmappable, and INV-36
  cascades every change intent that targets them. So the mappability of a change-intent state at
  commit `c` depends on whether some *other* state anywhere in the history mints the same record id.
- **The read-set clash.** INV-33 restricts `Resolve(c, n)` to `c`'s state and its parents' archive
  listings, and INV-33's test fails the reader on any other commit.
  - If `Resolve` obeys INV-33, it cannot see the collision. It returns the cascaded revision, which
    is absent from `Revisions()`, so INV-31 is false.
  - If `Resolve` sees the collision, INV-33 is false.
- **The draft concedes the contradiction.** §4.3's stated residual ("the first stays in the ledger
  but no longer resolves") presupposes that `Resolve` sees beyond `c`.
- **No test runs the conflicting cases together.** The INV-31 test runs only over the branches
  fixture. The INV-38 pinned pair is never run through INV-31 or INV-33.
- **Severity.** The probability is about 2⁻³¹ per same-slug pair, which is why this is LOW. The
  contract is still unsatisfiable as written.

### r2-m0-review-round4-20261117-F8 (LOW): two §0 rows are measured on a walk the draft rejects, or on a population other than the one the row describes

**Broken property:** §0's "each count is printed by the probe that claims it", read as evidence for
the draft's own consequences.

**Evidence:**
- **The "12" is a path-limited count.**
  - The "12 dual-presence commits", repeated in §6 and §14.3, come from
    `rev-list --all --full-history -- openspec/changes`, which is path-limited.
  - The provider walks unsimplified `--all`. There, **23** commits carry the dual-presence state
    (`<scratch>/dual23.txt`), from `7ced2a87` through `acd1ca28`.
  - The draft's own §3 says path-limited counts miss states.
- **The tip question has the wrong answer.** The "Release tags … 0" row exists so that "no tip is
  wedged" is checkable. Its answer is true for tags, but four branch tips sit on those commits (F6).
- **The "47 lines" population is mislabelled.** The probe counts
  `re.match(r"^\s*-\s*FROM:")` lines anywhere in *active* delta files, fences included, summed over
  the HEAD path-limited walk. The row describes "lines … inside a code fence … over the walk".
  Measured by fence and by walk, the numbers are 8, 201 or 1534 (Evidence), never 47.
- **The probe's fence mask is not OpenSpec's.** It is a naive toggle. It happens to agree with
  OpenSpec on this history (0 disagreements), but §0 attributes OpenSpec's masking to it.

### r2-m0-review-round4-20261117-F9 (LOW): the draft calls itself self-contained but is not, and drift remains

**Broken property:** the "Self-contained: no definition is inherited from a superseded draft" claim,
spec exactness (two conforming implementations must emit identical bytes), and Law 2.

**Evidence:**
- **The profile is undefined in draft 4.**
  - `proofbound.sdd-profile.v1` (its fields, glob semantics, the Spec Kit line and status regexes
    with their named groups, `absent`, `map`, `sibling`, and the allowed `changes.status` values) is
    defined only in superseded draft 2 §3.
  - Draft 4 relies on all of it: INV-1's "regexes lacking required groups", §4.2's "statement group",
    §5's `changes.status`, and the `owners`/`sponsors` matching.
  - Draft 2's profile also carries `changes.sections` and `changes.renamed` regexes. Draft 4 replaces
    those with hard-coded OpenSpec rules, and the fixture profiles omit them. Which shape INV-1
    accepts is therefore undefined.
- **The obligation `state` is unspecified.** §4.2 never says what `state` an obligation carries, in
  particular a REMOVED entry (`active` or `retired`). Yet it is in the payload bytes and in
  `requirement_obligations_view.state` (`projections.go:339`). The same applies to the BR
  `authorized_by`/`supersedes` and the CI `constrained_by` values, which are left implicit.
- **The archive glob is ambiguous.** §6 matches archives by the glob `changes/archive/*-<n>/`, where
  `*` matches `2026-01-01-add` in `2026-01-01-add-x`. So name `x` "has" `add-x`'s archive, while
  §4.3 strips an exact `YYYY-MM-DD-` prefix. OpenSpec has no instance of this (Evidence), but a
  two-change synthetic repository does.
- **The README still cites retired invariants.** `testdata/README.md` still maps fixture rows to
  INV-14, INV-24 and INV-27 (rows openspec 04, 06, 07 and unmappable/duplicate-obligation), and still
  heads its column "draft 3 §0 row".
- **"Exact bytes without the line terminator" is undefined for CRLF files.** OpenSpec normalises
  `\r\n` to `\n` before parsing.

### r2-m0-review-round4-20261117-F10 (LOW): two hand-offs are not yet durable. Resolve's error classes have no sentinels for M3, and the committed fixture helper does not implement the format the committed fixtures use

**Broken property:** §14.3, the M3 entry condition (M1 encodes something M3 needs), and the rule that
"if losing a file would cost more than 30 minutes, it goes in git".

**Evidence:**
- **Resolve's error classes.**
  - §6 gives only the unmappable case a sentinel (`intent.ErrUnmappable`). "Unknown name" and
    "archived earlier" are plain errors.
  - §14.3 requires M3 to make an unresolvable claim per-commit and reported. To do that, it must tell
    tenant-caused refusals apart from reader or I/O failures.
  - As drafted, M3 can only do this by changing M1's `Resolve` contract. The deferral is honest in
    naming the wedge, but it leaves M1 to encode a signature M3 cannot use as is.
- **The committed fixture helper.** At HEAD the committed `fixture_test.go` has no PARENTS/REF
  support, no `requireProfiles`, and no `TestBranchFixtureReplaysAsDocumented` (`git show HEAD:…`
  greps 0). The committed `branches/` fixture and README depend on that format.
  - On a fresh clone the helper replays `branches/` as a linear chain on `main`, and nothing checks
    it.
  - Every claim above about the fixture's shapes holds only for the working-tree file reviewed here.

## What holds

- **Per-state failure is the right design.** It closes the real wedge: `7ced2a87`'s leftover costs
  one unit-state, not the repository. The rejection in §10 D is better supported than §0 says, since
  four live branch tips sit on that state.
- **State-key de-duplication is sound** (§3, the question asked).
  - Requirement and change-intent payloads depend only on manifest files and their paths, plus the
    profile. Status follows from the path, owner and sponsor from the profile, and targets from the
    in-directory deltas. All of these are inside the state key.
  - So equal `(kind, id, digest)` implies equal payload, and de-duplication by `(kind, id, digest)`
    never merges distinct payloads.
  - `Resolve`'s parent dependency changes only *whether* it succeeds, never *which* revision it
    returns. `Revisions` emits every mappable unit regardless of resolvability, so INV-31's
    membership holds, except for F7.
- **The §6 merge and root rules are right for real history.** A first-parent rule would misattribute
  116 of 266 apparent archivings to merges. The root case is consistent and tested, though F4 shows
  that what counts as a root must be fixed by the same reader options the git connector uses.
- **The walk.** The unsimplified `--all` list closes round-3 F2's walk-set gap. On OpenSpec it is
  1624 commits and 338 states, so cost tracks states.
- **The batch-level checks hold apart from F1.**
  - De-duplication removes duplicates.
  - The cascade prevents dangling same-source targets.
  - `sdd` revisions reference only `intent.sdd`, so no cross-provider reference exists.
  - Requirement records carry no references, so `orderBatch` cannot cycle.
  - The id grammar fits: `BR-`/`CI-` + `slug` (2..40) + `-` + 6 is within `recordIDRE`;
    `O-<SEC>-<np>` is at most 32 characters after `O-`; and the empty name-part is closed.
- **Skips that are derived rather than ledgered are an acceptable trust posture.**
  - A skip can only withhold trust: no revision means no claim resolves, which means no verdict
    binding and no readiness PASS.
  - It hides no STOP that INT-INV-11 requires, because that invariant concerns the
    `records`/`specdir` foreign fixture.

  This holds on two conditions: provider defects must not be classified as skips (F2), and the
  family sentence must be amended explicitly (F3).
- **BR 0003 wording.**
  - O-5 and O-9 are retired with their statements kept, as intent SPEC §5 requires.
  - O-10 matches §3, §6 and §7, and its final sentence correctly carries the per-state qualification.
  - O-12 is observable through INV-30.
  - O-11 is dropped with an honest reason.
- **M3 entry conditions** are honest. §14.3 names the git-connector wedge as M3's, and no M1
  behaviour depends on it, apart from F10's error taxonomy.
- **The fixtures hold what they claim** (within F5 and F10).
  - All five round-3 shapes are present and asserted.
  - The PARENTS/REF replay is correct.
  - The `.gitignore` re-include fixes a real vacuity, since the profiles had never been committed.
  - My positive controls died, and my neutral control survived.
- **Build Laws.**
  - **Law 6:** spec before code. No provider code exists.
  - **Law 8:** no new dependency.
  - **Law 1:** `45a3bf8` is a script-regenerated `@generated` pin.
  - **Law 2:** the intent SPEC §1 and the plan point at draft 4. The remaining drift is in F9.

## Fitness for M1

**Not as written, but close.** No finding is HIGH, and none requires reshaping the contract. Each is
an omission or contradiction that M1 code written to the draft would encode. The fixes can land as
draft 5, or directly in the `sdd/SPEC.md` that M1 commits, provided a non-author re-verifies the
changed sections. M1 itself should be reviewed at the foundational tier.

Conditions:

1. **Close §7's enumeration.**
   - Every condition that can fail `Sync` after validation (payload size, encoding) must be a
     per-unit cause (F1).
   - The `ValidateRevision` catch-all must not convert provider-structure failures into skips. An
     invariant must pin each reference fixture to its exact expected `Skip` set (F2).
2. **List every registry M1 must change, with its test:** `ValidateRevision`, `New`, projections'
   `eventRoute`, the gates matcher, and intent SPEC §6/INT-INV-11. M1's acceptance must include
   projecting a sync that contains `intent.sdd` events (F3).
3. **Share the git connector's full read contract with the `sdd` reader, not only its argument
   list:** `--no-replace-objects`, environment scrubbing, and shallow, graft and ref refusal (F4).
4. **Rebuild INV-39 on the shape that discriminates.** That is the column-0 fenced header from
   `improve-init-onboarding` (F5). A test must also place a skip cause at a ref tip (F6), and the
   meanings of INV-1, 6 and 29 must be settled under §9's rule (F6).
5. **Resolve INV-33 against INV-38** (F7). **Inline the profile schema and the obligation `state`
   rule** (F9). **Commit the fixture helper** before M1 builds on it (F10).
