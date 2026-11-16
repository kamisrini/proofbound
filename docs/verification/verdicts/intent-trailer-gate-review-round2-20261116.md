---
schema: vera.verdict.v1
verdict_id: intent-trailer-gate-review-round2-20261116
status: NEEDS_WORK
reviewed_commit: 04dd68e8ae130a970ff89a5d7f691442cc15b297
findings:
  - finding_id: intent-trailer-gate-review-round2-20261116-F1
    severity: MED
  - finding_id: intent-trailer-gate-review-round2-20261116-F2
    severity: MED
  - finding_id: intent-trailer-gate-review-round2-20261116-F3
    severity: LOW
  - finding_id: intent-trailer-gate-review-round2-20261116-F4
    severity: LOW
  - finding_id: intent-trailer-gate-review-round2-20261116-F5
    severity: LOW
artifact_path: docs/verification/verdicts/intent-trailer-gate-review-round2-20261116.md
artifact_sha: 4e22a5beb18c99bb04f25ad9147b8d40d433bebbc9a0a52dee5e6750d44bcd76
---

# Intent-trailer gate fix (CI-intent-trailer-block-xp2pke / DEF-INTENT-TRAILER-BLOCK-001): review round 2

## Provenance

- **Under review:** `7293be8` (the round-1 fixes: `TestIntentTrailers_SharedCorpusKernelBinds`,
  8 new corpus rows, the gate's `GIT_*` scrub and `GIT_NO_REPLACE_OBJECTS=1`, the UTF-8 check, the
  CRLF diagnostic, `LC_ALL=C`), `04dd68e` (generated pin refresh), and, for context, the original
  fix `07c384d`. Reference implementation: `kernel/internal/connector/git/gitcmd/gitcmd.go`
  (`intentTrailers`, `parseScalars`, `run`, `gitEnvironment`) and
  `kernel/internal/connector/git/connector.go` (`resolvedCommit`, `intentIDRE`).
- **HEAD at start:** `04dd68e8ae130a970ff89a5d7f691442cc15b297`. **HEAD at end:**
  `04dd68e8ae130a970ff89a5d7f691442cc15b297` (unchanged).
- **Driving mechanism:** `/proofbound-review` procedure, executed by a dispatched Agent.
- **Tier:** connector/peripheral (one adversarial review).
- **Reviewer identity:** `dispatched-non-author-reviewer-20261116-gate-r2`, a **non-author** of
  this change; this verdict is acceptance evidence under Law 9.
- **Writes:** no repository file other than this verdict was written. All probes ran in a scratch
  directory: a detached throwaway worktree (removed with `git worktree remove --force`), a
  `git clone --no-hardlinks` at the reviewed HEAD for mutation runs (every mutated file restored,
  `git diff --quiet` confirmed), and 36 throwaway probe repositories whose commits were built
  byte-exactly with `git hash-object --literally`. The kernel side of each probe was evaluated by a
  scratch-only test in the clone that runs the real `gitcmd.New` → `Commits` → `connectorgit.Sync`
  path with the same always-resolving resolver as the author's test.
- No cleanroom redactions were needed.

## Round-1 findings re-verified

- **Round-1 F1 (corpus does not pin the ported rules): PARTIALLY FIXED.** Every named round-1
  survivor now dies, which I reproduced with my own calibration (neutral comment edit survives;
  positive control `i < 1 || line[i] != ""` → `i < 1` is killed): M1 (≥ 3 lines) by
  `two-line-blank-first`, M2 (`records`) by `records-qualified`, M3/M4 (`+` bounds) by
  `slug-too-long`/`suffix-too-long`, M5 (case) by `uppercase-id`. Kernel side: `len(lines) < 3`
  disabled, the provider check widened, `utf8.Valid` disabled, `--no-replace-objects` dropped and the
  env scrub dropped are all killed. The end-to-end test is real (no build tag, runs under plain
  `go test`, drives the actual adapter and `Sync`). But the property the finding named — the corpus
  pins the rules so the parsers cannot drift — is still not established; see new F1.
- **Round-1 F2 (`make check` never runs the gate against HEAD): its central claim was WRONG.**
  Reproduced the author's rebuttal independently in a detached worktree at `04dd68e`:
  `bash scripts/tests/delivery-enforce.test.sh` exits 0 on the clean HEAD and exits 1 after a
  governed commit (touching `scripts/`) whose `Intent:` line is a body paragraph. The self-test runs
  the real `scripts/delivery-enforce.sh` with only `make`/`go` shimmed, so line 39 executes the real
  gate against the real HEAD (twice per run). `make hooks-test` runs that self-test inside
  `make check`. Round 1 was wrong to say `make check` never examines HEAD's trailer. Its **residual
  point** (a live kernel bind check after `sync all` in real `delivery-enforce`) is, in my judgement,
  acceptable to leave out of this intent: on 36 out-of-corpus probes the port and the kernel agree on
  every input the kernel can read (see "What holds"), so a live check would add little. The real gap
  is a different one, which the reproduction exposes: the coverage is HEAD-only and incidental (new
  F2, new F3).
- **Round-1 F3 (replace refs / inherited `GIT_*`): FIXED.** The scrub and `GIT_NO_REPLACE_OBJECTS=1`
  (gate lines 8-9) run before the first git call and are inherited by `intent-applicability.sh`.
  Removing either is killed by the new self-test cases (replace-ref case; `GIT_DIR=/nonexistent`
  case). The scrub does not break normal use: the gate ran correctly from the repo, from a
  detached worktree, and through `delivery-enforce.test.sh`.
- **Round-1 F4 (non-UTF-8 / CRLF diagnostic): PARTIALLY FIXED.** The Latin-1-in-message case and
  the CRLF diagnostic are fixed and pinned (removing the iconv check is killed by
  `latin1-other-trailer`; disabling the CRLF branch is killed by the diagnostic assertion). The
  "rejected outright" claim holds only for the message, and only for iconv's definition of UTF-8;
  see new F4.

## Findings

### intent-trailer-gate-review-round2-20261116-F1 (MED): the corpus still leaves non-equivalent grammar and value-rule boundaries unpinned on both sides, including two that reopen the silent false-pass class

**Broken property:** "both parsers are pinned to one shared corpus … so they cannot drift apart"
(gate header comment, `cases.tsv` header, CI-intent-trailer-block-xp2pke "Explicit exclusions").

**Evidence:** These are mutation runs on a scratch clone at `04dd68e`, with the stock
`intent-delivery-applicability.test.sh` and the stock `go test ./internal/connector/git/...`. The
calibration is: the neutral control survives and the positive control is killed on each side.

Shell side (`scripts/intent-delivery-applicability.sh`): the seven mutants below **survive**. For
each, I built a witness probe on which the mutated gate disagrees with the real kernel, while the
unmutated gate agrees with it:

| Mutant | Witness message (governed commit) | Mutated gate | Kernel |
|---|---|---|---|
| line 56: `: .+$` → `:.+$` | `subject\n\nIntent: CI-fixture-item-acde12\nFoo:bar\n` | pass | binds nothing (no error) |
| line 60: drop `i < 1 \|\|` | `Intent: CI-fixture-item-acde12\nFoo: a\nBar: b\n` | pass | binds nothing (no error) |
| line 75: `]{6}$` → `]{5,6}$` | `Intent: CI-fixture-item-acde1` | pass | sync error |
| line 75: `?` → `{0,2}` on the qualifier | `Intent: records:specdir:CI-fixture-item-acde12` | pass | sync error |
| line 75: `{1,62}` → `{1,61}` | 63-character slug | fail | binds |
| line 75: `{1,62}` → `{3,62}` | `Intent: CI-ab-acde12` | fail | binds |
| line 57: key prefix `Intent` (6 chars) | `Intent-Note: hi\nIntent: CI-fixture-item-acde12` | fail | binds |

The first two rows are the exact harm DEF-INTENT-TRAILER-BLOCK-001 exists for: the gate passes and
the kernel silently records no claim.

Kernel side: the mirror mutants all **survive** the whole connector/git test suite. They are:
- `if i < 0 || lines[i] != ""` → `if i >= 0 && lines[i] != ""` (`gitcmd.go:199`);
- `intentIDRE` `{1,62}` → `{1,61}`, `{1,62}` → `{3,62}`, and `{6}` → `{5,6}`;
- `len(parts) != 2` → `len(parts) < 2` (`connector.go:193`);
- `match[1] == "Intent"` → `strings.HasPrefix(match[1], "Intent")`.

So the kernel's own grammar can move at any of these points with the gate still green.

**Why this happens:** the corpus has no row for an all-trailer message of ≥ 3 lines, a no-space
line inside the final block, a double qualifier, a slug at the lower bound or at the 63-character
upper bound, a 5-character suffix, or an `Intent-*` key. The author's journal reports "slug/suffix
bounds … killed". That is true only for the `+` (unbounded) mutants, not for the boundary values
themselves. The fix killed the named survivors one row at a time instead of covering each rule's
boundary.

### intent-trailer-gate-review-round2-20261116-F2 (MED): the gate examines only HEAD, so in this repository's own commit-then-pin workflow a governed commit is never examined by `make check` or `delivery-enforce`

**Broken property:** the deferral's harm statement, "a governed commit whose `Intent:` line sits in
a body paragraph lands with `make check` green", is recorded as FIXED.

**Evidence:**
- **The only call site checks HEAD.** The gate's sole call site is `scripts/delivery-enforce.sh:39`
  (no arguments, so HEAD only). Both real `delivery-enforce` and the `make check` path through
  `delivery-enforce.test.sh` therefore look at one commit.
- **Reproduction.** In the detached worktree I made the bad governed commit described above. The
  self-test then exits 1. I then added one docs-only commit on top (`README.md`). The self-test
  exits **0**, and the gate prints `applicable=false` for HEAD. Run with `--commit HEAD~1`, the same
  gate still refuses the buried commit ("OUTSIDE the final trailer block").
- **The kernel will not catch it either.** The kernel has no notion of applicability: no
  `applicab*` in non-test kernel code, and an un-trailered commit syncs without error. Nothing
  downstream refuses the buried commit.
- **Real history shows the workflow guarantees this.** The documented commit-then-write-the-pin
  workflow (DEF-ROUTE-MATRIX-PRECOMMIT) puts a non-applicable pin commit directly after each
  governed kernel commit:

  | Governed commit | Pin commit on top | Time the governed commit was HEAD |
  |---|---|---|
  | `7293be8` (`applicable=true`) | `04dd68e` (`applicable=false`), 15:12:19 | 1 second (committed 15:12:18) |
  | `07c384d` (`applicable=true`) | `fb5f2dc` (`applicable=false`) | 37 seconds |

  `7293be8` was HEAD for one second, far shorter than a `make check` run, so the governed commits
  this very intent produced were never the HEAD any gate ran against. A body-only `Intent:` in
  `7293be8` would have landed with every gate green. That is the original harm, unchanged for this
  workflow.

The HEAD-only scope predates this intent. It is in scope here because the deferral's closure claims
the harm is fixed.

### intent-trailer-gate-review-round2-20261116-F3 (LOW): the `make check` coverage that refutes round-1 F2 is incidental, unpinned, and silent when it fires

**Broken property:** Law 3 ("enforcement proves it fires"), applied to the `make check` path the
F2 rebuttal relies on.

**Evidence:**
- **Nothing pins the gate's invocation.** `delivery-enforce.test.sh` asserts the order of the
  shimmed `make`/`go` calls (lines 21-31), but the gate call is not in that list. On the scratch
  clone, deleting line 39 (`scripts/intent-delivery-applicability.sh`) from `delivery-enforce.sh`
  leaves both `delivery-enforce.test.sh` and `intent-delivery-applicability.test.sh` exiting 0. No
  other test references the gate. The only `make check` check of HEAD can therefore be removed
  with every test green.
- **When it fires, it says nothing.** The run that fires writes its output to `$tmp/first.out`
  (line 10), and the `EXIT` trap deletes that file. In my reproduction the failing self-test
  produced **0 bytes** of output. `make check` goes red with no diagnostic: the gate's "OUTSIDE the
  final trailer block" message is lost, and an operator sees an unexplained `hooks-test` failure in
  a lock/ordering test.

### intent-trailer-gate-review-round2-20261116-F4 (LOW): the UTF-8 rejection does not match what the kernel rejects

**Broken property:** "the gate must accept exactly the commits whose Intent claim the kernel binds"
(`intent-delivery-applicability.test.sh`), and the F4 fix claim that invalid UTF-8 is "rejected
outright".

**Evidence:** Each probe below is a governed commit with a correct final-block
`Intent: CI-fixture-item-acde12`. The **gate passes** (`applicable=true … intents=…`), while the
kernel's `Commits` fails, so nothing binds and `Sync` cannot proceed:

| Probe | Kernel error |
|---|---|
| Message containing `\xf4\x90\x80\x80` (above U+10FFFF) | "scalar is not valid UTF-8" |
| Author name `J\xf6rg`, no `encoding` header | "scalar is not valid UTF-8" |
| Touched path `scripts/f\xe9.sh` | "path is not valid UTF-8" |

Why they diverge:
- glibc `iconv -f UTF-8 -t UTF-8` accepts code points above U+10FFFF (`printf '\xf4\x90\x80\x80' | iconv …` exits 0); Go's `utf8.Valid` rejects them.
- The gate checks only `%B` (line 43). The kernel validates all nine scalars and every path.

In real `delivery-enforce`, `sync all` aborts loudly first. In the `make check` path, `go` is
shimmed, so the gate is the only check and it passes.

Separately, the gate fails closed without `iconv`, which is correct, but with a misleading
diagnostic. With `iconv` removed from `PATH`, a valid ASCII message is refused as "not valid UTF-8,
which the kernel rejects outright".

### intent-trailer-gate-review-round2-20261116-F5 (LOW): the deferral record describes the fix that round 1 rejected, not the current one

**Broken property:** Law 2 (one home per datum): the deferral row is the home for this deferral's
state.

**Evidence:** `docs/deferrals.tsv` line 13 (`DEF-INTENT-TRAILER-BLOCK-001`) was not touched by
`7293be8`, so it still reflects the round-1 version of the fix:
- It still says "12 cases" (the corpus has 20).
- It still says the parsers are pinned "by TestIntentTrailers_SharedCorpus" (extraction only), not
  by the end-to-end test.
- It still justifies the condition-(3) substitution with the Postgres/ledger reasoning that round 1
  disputed.
- Its harm statement does not reflect the HEAD-only reality documented in F2.

The corrected F2 understanding, that `make check` does run the gate on HEAD, lives only in the
journal, which is a session-history file rather than the deferral's home.

## What holds

- **Grammar and value parity on readable inputs.** On 36 probes outside the corpus the unmutated
  gate and the real kernel path agreed on every commit the kernel can read. The probes covered:
  - NUL in the value, subject, or body: git truncates `%B` at the NUL for both, so they agree;
  - a message without a final newline, leading blank lines, an empty message;
  - a tab after the value, a tab-separated key, `Intent : x`, a key with digits and hyphens, a
    vertical tab in another trailer;
  - trailing spaces in the value, two `Intent` lines with one bad value;
  - a merge commit carrying the trailer;
  - an `encoding ISO-8859-1` header with Latin-1 bytes in the message and trailer (git re-encodes,
    both bind), a bogus `encoding` header (both refuse), a surrogate, an overlong encoding, U+FFFE,
    a BOM;
  - a CR-only separator line;
  - slug lengths 1, 2, and 63, a 5-character suffix, a double qualifier, a ≥ 3-line all-trailer
    message, a no-space line inside the block, and an `Intent-Note:` key.

  The only disagreements are the unreadable-commit cases in F4.
- **Round-1 survivors.** Every round-1 survivor is now killed on its own side, with calibrated
  controls. The kernel end-to-end test runs untagged under `make check` and fails when the
  kernel's framing, value rule, replace-object handling, or environment handling regresses.
- **Round-1 F3 is fully closed.** The environment scrub does not break normal use.
- **Self-test floors.** The calibration floors in both tests (≥ 19 cases, ≥ 4 pass, ≥ 10 fail) are
  satisfied by 20 rows (6 pass, 14 fail), and a constant answer is rejected.
- **The rebuttal is correct.** The author's F2 rebuttal is factually correct and was reproduced as
  described.
- **Law compliance.**
  - No production kernel code changed (CI scope respected).
  - `iconv` is already a dependency of `kernel/scripts/check-witness.sh`, so no new dependency was
    added (Law 8).
  - The pin refresh `04dd68e` is a generated-file commit (Law 1).
  - The deferral is correctly left `OPEN` pending a non-author verdict (Law 9).
