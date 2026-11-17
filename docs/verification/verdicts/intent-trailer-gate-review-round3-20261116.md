---
schema: vera.verdict.v1
verdict_id: intent-trailer-gate-review-round3-20261116
status: NEEDS_WORK
reviewed_commit: 0dc9e087454a58748b82b72d531797ff1966a612
findings:
  - finding_id: intent-trailer-gate-review-round3-20261116-F1
    severity: MED
  - finding_id: intent-trailer-gate-review-round3-20261116-F2
    severity: LOW
  - finding_id: intent-trailer-gate-review-round3-20261116-F3
    severity: LOW
  - finding_id: intent-trailer-gate-review-round3-20261116-F4
    severity: LOW
  - finding_id: intent-trailer-gate-review-round3-20261116-F5
    severity: LOW
  - finding_id: intent-trailer-gate-review-round3-20261116-F6
    severity: LOW
artifact_path: docs/verification/verdicts/intent-trailer-gate-review-round3-20261116.md
artifact_sha: bb07f2c754c6594b99deea4777b4710846789eb7fcdf43575b6ae8e712a50af9
---

# Intent-trailer gate (CI-intent-trailer-range-gwi9uu / DEF-INTENT-TRAILER-BLOCK-001): review round 3

## Provenance

- **Under review:** `96c515d` (range mode, strict UTF-8, `intent-trailer-gate` make step, the
  `PROOFBOUND_INTENT_GATE` seam, corpus at 32 rows, DEF-INTENT-CI-REVISION-001), `8cf9bfc`
  (generated pin refresh) and `0dc9e08` (`scripts/mutant-apply.sh`). The reference implementation
  is `kernel/internal/connector/git/gitcmd/gitcmd.go` (`readCommit`, `intentTrailers`,
  `parseScalars`, `files`, `nulUTF8Strings`, `run`) and `kernel/internal/connector/git/connector.go`
  (`resolvedCommit`, `intentIDRE`).
- **HEAD at start:** `0dc9e087454a58748b82b72d531797ff1966a612`. **HEAD at end:**
  `0dc9e087454a58748b82b72d531797ff1966a612` (unchanged; 10 commits ahead of `origin/main`).
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral (one adversarial review).
- **Reviewer identity:** `dispatched-non-author-reviewer-20261116-gate-r3`. This reviewer did
  **not** author the change, so this verdict counts as acceptance evidence under Law 9.
- **Writes:** this verdict is the only repository file written. All probes ran in a session
  scratch directory:
  - a `git clone --no-hardlinks` at the reviewed HEAD, used for mutation runs. Every substitution
    went through `scripts/mutant-apply.sh` and every file was restored; `git diff --quiet` was
    confirmed after each sweep. Two scratch-only probe tests were added to the clone, never to the
    repo.
  - throwaway probe repositories, plus a bare "remote" for the range scenarios. Commits were built
    byte-exactly with `git hash-object`.

  No worktree of the real repository was created. The only command run against the real
  repository was the read-only `scripts/intent-delivery-applicability.sh`. It exits 0 and checks
  all 10 unpushed commits: 3 applicable, each binding its CI.
- **Not run:** a full `make check`. It writes witness spool files into the repository's
  `.proofbound/`, which the review rules forbid.
- No cleanroom redactions were needed.

## Round-2 findings re-verified

- **Round-2 F1 (corpus leaves grammar/value boundaries unpinned): FIXED.**
  - **Calibration.** Shell side: the neutral control (qualifier alternation reordered to
    `(specdir|records)`) survives, and the positive control (value check replaced by `true`) is
    killed. Kernel side: the neutral control (`len(lines) - 1 + 0`) survives, and the positive
    control (`len(lines) < 0`) is killed.
  - **Shell side.** I re-applied all seven round-2 shell survivors through `mutant-apply.sh`. All
    seven are now **killed**: `:.+$`, the dropped `i < 1 ||`, `{5,6}` suffix, `{0,2}` qualifier,
    `{1,61}`, `{3,62}`, and the 6-character `Intent` prefix.
  - **Kernel side.** These mutants are all **killed** by `go test ./internal/connector/git/...`:
    - `intentIDRE` `{1,61}` and `{5,6}`;
    - `i >= 0 && lines[i] != ""`;
    - `strings.HasPrefix(match[1], "Intent")`;
    - `len(parts) < 2`;
    - `utf8.Valid` disabled in `nulUTF8Strings`, and disabled in `parseScalars`.
  - **Residual.** The new range and UTF-8 code has unpinned rules of its own (F2 below).
- **Round-2 F2 (gate examines only HEAD): PARTIALLY FIXED.** With an upstream configured and the
  check run after the last commit, the buried-under-a-pin shape is now refused. I reproduced this
  (scenario R1a below). The harm still lands on every path where the range is empty or absent, so
  it is not closed (F1 below).
- **Round-2 F3 (gate invocation unpinned, self-test silent): FIXED for `delivery-enforce`.** I
  tested two mutants of `scripts/delivery-enforce.sh` on the clone:
  - **Gate line deleted:** `delivery-enforce.test.sh` exits 1 with 325 bytes of output ("delivery
    command count mismatch …" plus the captured outputs).
  - **Gate call made fail-open (`|| true`):** the test exits 1 with "failing intent gate was
    accepted".

  The unmutated test exits 0. However, the **new** `make check` step is not pinned (F2 below).
- **Round-2 F4 (UTF-8 rejection does not match the kernel): PARTIALLY FIXED.**
  - **Now fixed.** The three round-2 probes all fail at the gate, matching the kernel: code point
    above U+10FFFF, Latin-1 author, Latin-1 path. A lenient `utf8` decoder mutant is killed.
  - **Still divergent.** Perl's strict decoder is **not** Go's `utf8.Valid`: it rejects Unicode
    noncharacters that Go accepts (F3 below). That is a regression from round 2, where the U+FFFE
    probe agreed.
- **Round-2 F5 (deferral record stale): PARTIALLY FIXED.** `117d82f` appended an "UPDATE … after
  review rounds 1-2" paragraph to the DEF-INTENT-TRAILER-BLOCK-001 row. `96c515d` did not touch the
  row, so it does not describe the current state (F6 below).

## Findings

### intent-trailer-gate-review-round3-20261116-F1 (MED): the "unpushed range" still lets a governed commit with an unbindable Intent land with every gate green, and the gate's verdict at a fixed HEAD depends on remote-tracking state

**Broken property:**
- **The deferral's harm statement.** "A governed commit whose `Intent:` line sits in a body
  paragraph lands with `make check` green" is again presented as fixed, by `docs/gates.md` row 17
  and CI-intent-trailer-range-gwi9uu.
- **BR-gate-integrity-4b2e08 O-1.** `make check` must be deterministic. The deferral row itself
  cites this obligation to rule out an environment-dependent gate.

**Evidence:** the probe repository has a bare `origin`, a pushed base commit, then the round-2
shape: a governed commit (`scripts/bad.sh`) whose `Intent:` line is separated from
`Co-Authored-By` by a blank line, followed by a docs-only pin commit. The gate runs in default mode,
exactly as `make intent-trailer-gate` runs it.

| Scenario | Gate exit |
|---|---|
| R1a: unpushed, upstream tracked | **1** (correct) |
| R1b: same HEAD after a plain `git push` with no check run | **0** |
| R5: fresh `git clone` of that remote (a CI checkout) | **0** |
| R2: same shape on a new branch with no upstream | **0** |
| R3: detached HEAD on top of the shape | **0** |
| R4: `branch.*.merge` set but the remote-tracking ref missing (pruned or never fetched) | **0** (no diagnostic) |
| `--commit HEAD~1` in any of these | 1 |

On the R5 clone, the real adapter plus `Sync` (scratch probe test) returns `sync-error=<nil>` and
`trailers=[]` for the governed commit. No claim is bound and nothing downstream refuses.

**Why it matters here:**
- **The ordering defeats the gate.** The gate now protects a commit only if `make check` runs
  *after* that commit and *before* it is pushed. `docs/eod-prompt.md` lines 31-34 order the
  closeout the other way: run `make check`, then "commit only coherent completed changes", then
  "push". No git hook enforces the order: `core.hooksPath` is unset, and only `.sample` hooks
  exist.
- **The self-test encodes the laundering.** `intent-delivery-applicability.test.sh` asserts as a
  "neutral control" that once the buried commit is upstream, the default run passes. So a commit
  pushed unchecked is exempt permanently, by design.
- **No range shown.** The no-upstream, detached and missing-ref fallbacks are silent. The make
  step sends stdout to `/dev/null`, and nothing reports which range was checked.
- **Same HEAD, two verdicts.** R1a and R1b show the same HEAD and tree going from red to green
  when only `refs/remotes/origin/main` moves.

### intent-trailer-gate-review-round3-20261116-F2 (LOW): the new range, identity, root-commit and make-step behaviour is unpinned; calibrated mutants survive

**Broken property:**
- **Law 3 and the CI acceptance boundary.** The acceptance boundary requires "calibrated mutation
  evidence on both sides".
- **The `docs/gates.md` row 17 claim.** It says the gate runs "in `make check` as
  `intent-trailer-gate`".

**Evidence:** these are mutants of `scripts/intent-delivery-applicability.sh` applied through
`mutant-apply.sh` (exit 0, so each is a real mutant) against the stock self-test. Calibration as
in the round-2 F1 section above.

| Mutant | Self-test | Witness (unmutated gate → mutated gate) |
|---|---|---|
| `\|\| commits+=("$head")` → `\|\| true` | **survives** | detached HEAD at a body-only-Intent governed commit: exit 1 → **0** (checks nothing) |
| drop `'%cn'` from the identity loop | **survives** | Latin-1 committer name: exit 1 → **0**; kernel `Commits` fails "scalar is not valid UTF-8" |
| drop `'%ae'` / drop `'%ce'` | **survive** | same class |
| root-commit branch `paths_ok=$(…)` → `paths_ok=yes` | **survives** | root commit with a non-UTF-8 path |
| parent `awk '{print $1}'` → `$NF` | **survives** | merge whose first-parent diff differs |

The kill side is sound:
- dropping `'%an'`, emptying the range, `failed=0`, and a lenient `utf8` decoder are all killed;
- excluding the two controls, 11 of the 17 shell mutants I applied are killed (the 7 round-2
  survivors plus these 4), and the 6 in the table survive.

The new `make check` step is also unpinned. Removing ` intent-trailer-gate` from the `check:`
prerequisites (via `mutant-apply.sh`) leaves `scripts/p6-task1-close.sh` exiting 0. Its
`check_prereqs` list (line 68) enumerates every other `check` prerequisite but not this one. No
other script inspects the `check:` line, so the primary invocation can be deleted with every gate
green.

### intent-trailer-gate-review-round3-20261116-F3 (LOW): Perl strict "UTF-8" is not Go `utf8.Valid`; the gate refuses commits the kernel binds

**Broken property:**
- **The claimed match.** "Strict UTF-8 exactly as the kernel judges it (Go utf8.Valid)"
  (`scripts/intent-delivery-applicability.sh` lines 23-26), and the commit message's "matching Go
  utf8.Valid".
- **The test's requirement.** The gate accepts exactly the commits whose claim the kernel binds.

**Evidence:**
- **Byte-sequence comparison.** I ran 29 byte sequences through the gate's exact Perl one-liner
  (Perl 5.34) and through `utf8.Valid` (Go 1.27.4). They agree on surrogates, overlongs, 5/6-byte
  forms, values above U+10FFFF, truncated sequences, lone continuation bytes, and NUL. They
  **disagree** on the four noncharacters tested. Go accepts each of these and Perl rejects it:
  `EF BF BE` (U+FFFE), `EF BF BF` (U+FFFF), `EF B7 90` (U+FDD0), and `F4 8F BF BF` (U+10FFFF).
- **End-to-end probes.** In each probe below the governed commit carries a correct final-block
  `Intent: CI-fixture-item-acde12`. The gate exits 1 ("… is not valid UTF-8, which the kernel
  rejects outright"), while the real adapter plus `Sync` binds the claim with `sync-error=<nil>`:
  - U+FFFE in the body;
  - U+FDD0 in the subject;
  - U+FFFF in the author name;
  - U+10FFFF in a touched path.

**Impact:** this is fail-closed, not a silent false pass. However:
- the diagnostic is false;
- such a commit blocks `make check` until history is rewritten;
- no corpus row covers a noncharacter.

### intent-trailer-gate-review-round3-20261116-F4 (LOW): a governed path that git C-quotes is classified non-applicable, so a governed commit with no Intent at all passes

**Broken property:** the gate must demand a bindable claim from every governed commit. This flaw is
pre-existing, in `scripts/intent-applicability.sh`, which the gate calls. It is outside
CI-intent-trailer-range-gwi9uu's allowed scope, but it sits on the gate's acceptance path.

**Evidence:** `intent-applicability.sh` line 76 runs `diff-tree` with `core.quotePath=false` but
without `-z`. For a path containing `"`, `\`, a tab or a newline, git therefore emits a C-quoted
string, for example `"scripts/\"x/run.sh"`. The leading `"` fails every pattern in `matches()`.

Each probe commit below is message `tweak scripts` with no Intent line. The gate prints
`applicable=false` and exits **0** for:
- `scripts/a"b.sh`;
- `scripts/a\b.sh`;
- `kernel/x<TAB>y.go`;
- `scripts/"x/run.sh` (a whole directory escapes).

The control `scripts/ab.sh` exits 1. The kernel ingests these commits with `trailers=[]` and
`sync-error=<nil>`.

### intent-trailer-gate-review-round3-20261116-F5 (LOW): `mutant-apply.sh` still admits comment-only mutants on code lines and refuses comment-out deletions

**Broken property:** the compiled response to LESSON `fixture-not-what-it-claims` in `0dc9e08`. It
promises that an exit-0 mutant "changed in at least one non-comment line", and that a pattern which
matched a comment cannot be counted.

**Evidence:**
- **Comment-only change counted.** On a two-line Go file `x := 1 // keep` / `y := 2`, the
  substitution `s/keep/kept/` changes only a trailing comment, yet `mutant-apply.sh` exits **0**
  ("real mutant"). The classifier checks only whether the changed line *starts* with `#` or `//`,
  so a comment match on a code line is exactly the lesson's class, still counted.
- **Real mutant refused.** Commenting out a code line (`s/^x := 1/\/\/ x := 1/m`), the classic
  statement-deletion mutant, is refused with exit 4 ("only changed comments").

### intent-trailer-gate-review-round3-20261116-F6 (LOW): DEF-INTENT-TRAILER-BLOCK-001 does not describe the fix under review

**Broken property:** Law 2. The deferral row is the home for this deferral's state.

**Evidence:** `git log -L13,13:docs/deferrals.tsv` shows that the row was last changed in
`117d82f`; `96c515d` only appended DEF-INTENT-CI-REVISION-001. Line 13 is therefore stale:
- it still says "12 cases" and "pinned … by TestIntentTrailers_SharedCorpus" in its FIXED clause,
  and "20 cases" in its update;
- its closing condition still says the range check is a "design choice pending owner", although
  the owner chose the unpushed range;
- its evidence-path list omits `CI-intent-trailer-range-gwi9uu`, `docs/gates.md`, `Makefile` and
  `scripts/delivery-enforce.sh`.

The state of the round-2 fix is recorded only in the journal, `docs/gates.md`, and the new CI.

## What holds

- **Grammar and value parity.** Every round-2 boundary is now pinned on both sides with calibrated
  controls, as listed in the round-2 F1 section above.
- **Range mechanics, when an upstream exists:**
  - `rev-list upstream..HEAD` includes side-branch commits brought in by a merge;
  - a rebased or amended commit gets a new SHA and is rechecked;
  - an upstream ahead of HEAD collapses to HEAD;
  - path checks diff against parent[0], as the kernel's `files()` does.

  Applicability uses `-m` (any parent), which is stricter than the kernel and therefore safe.
- **The `PROOFBOUND_INTENT_GATE` seam is not a net bypass.** Setting it to `true` skips only the
  second invocation, at `delivery-enforce.sh:41`. `make check-witnessed` (line 37) runs `make check`
  through `check-witness.sh:247`, and that runs `intent-trailer-gate`, which calls the script
  directly rather than through the seam.

  The seam also fails closed:
  - a nonexistent path exits 127 under `set -e`;
  - the fail-open (`|| true`) mutant is killed.
- **The real repository is clean against the gate.** The default run checks all 10 unpushed
  commits and passes. `7293be8`, `96c515d` and `0dc9e08` bind their CIs. No kernel production
  code changed, which respects the CI scope.
- **DEF-INTENT-CI-REVISION-001 holds for every shipping provider.** In
  `intent.go:318-343`, `validateReferences` keys on source, then record_kind, and so on, over
  `targets ++ constrained_by ++ supersedes` (line 269). A scratch `ValidateRevision` probe gives:
  - a records-sourced CI superseding itself (with records targets) is rejected, "intent: invalid
    reference";
  - a specdir CI with specdir targets is rejected too.

  **The mechanism is stated more absolutely than the code supports:** a *specdir*-sourced CI whose
  targets are *records*-sourced and which supersedes itself **is accepted** by `ValidateRevision`
  (`err=<nil>`), because `intent.records` sorts before `intent.specdir`. The claim still holds in
  practice: `specdir.parseIntent` always emits `Supersedes: []`, and the records provider can only
  produce the rejected shapes. A future amendment should state the rule per field, not as "sort
  order always descends".
