---
schema: vera.verdict.v1
verdict_id: intent-trailer-gate-review-round1-20261116
status: NEEDS_WORK
reviewed_commit: 1de1b21ac2b04a04474fcb33d006e95356754f7f
findings:
  - finding_id: intent-trailer-gate-review-round1-20261116-F1
    severity: MED
  - finding_id: intent-trailer-gate-review-round1-20261116-F2
    severity: MED
  - finding_id: intent-trailer-gate-review-round1-20261116-F3
    severity: LOW
  - finding_id: intent-trailer-gate-review-round1-20261116-F4
    severity: LOW
artifact_path: docs/verification/verdicts/intent-trailer-gate-review-round1-20261116.md
artifact_sha: 0638175ded30d1ac74414a2d2a12e6b94e17ea3aa31de3310c8ba52549e2e065
---

# Intent-trailer gate fix (CI-intent-trailer-block-xp2pke / DEF-INTENT-TRAILER-BLOCK-001): review round 1

## Provenance

- **Under review:** commit `07c384d` (the gate port, the shared corpus, both corpus tests, and the
  deferral row update) and pin commit `fb5f2dc`, as present at HEAD. Files:
  - `scripts/intent-delivery-applicability.sh`
  - `scripts/tests/intent-delivery-applicability.test.sh`
  - `scripts/tests/fixtures/intent-trailers/` (`cases.tsv` + 12 `.msg`)
  - `kernel/internal/connector/git/gitcmd/gitcmd_test.go::TestIntentTrailers_SharedCorpus`
  - `docs/intent/records/change-intents/CI-intent-trailer-block-xp2pke/0001.md`
  - the `DEF-INTENT-TRAILER-BLOCK-001` row of `docs/deferrals.tsv`
- **Reference implementation:** `kernel/internal/connector/git/gitcmd/gitcmd.go` (`intentTrailers`,
  `trailerLineRE`, `readCommit`, `run`) and `kernel/internal/connector/git/connector.go`
  (`resolvedCommit`, `intentIDRE`).
- **HEAD at dispatch:** `1de1b21ac2b04a04474fcb33d006e95356754f7f`. **HEAD at completion:**
  `1de1b21ac2b04a04474fcb33d006e95356754f7f` (unchanged).
- **Driving mechanism:** `/proofbound-review` procedure, executed by a dispatched Agent.
- **Tier:** connector/peripheral (one adversarial review).
- **Reviewer identity:** `dispatched-non-author-reviewer-20261116-r2m0`, a **non-author** of this
  change. This verdict counts as acceptance evidence under Law 9.
- **Writes:** no repository files other than this verdict and its sibling
  `r2-m0-review-round1-20261116.md` were written.
- **How the probes ran:**
  - All probes ran in a scratch directory: a throwaway git repo, a `git clone --no-hardlinks` of
    this repo at the reviewed HEAD for mutation runs, and a verbatim copy of `intentTrailers` plus the
    `resolvedCommit` value rule built as a scratch Go program.
  - The real gate script and the real Go tests were run unmodified: `go test ./internal/connector/git/...`
    passes, and `intent-delivery-applicability.test.sh` passes.
  - On real history the gate refuses `a028984` and accepts `a5a085b`, `07c384d`, and `23a6fcc`.
- **Grammar parity probe.** 21 hand-built messages:
  - CRLF; `Intent:` followed only by spaces; `Intent-X:`; lower-case key; tab after the colon;
    U+00E9, fullwidth, and superscript characters in the id; a BOM.
  - A message made only of trailers; a 2-line message; a whitespace-only separator line; a
    trailer-like subject line; trailing blank lines and a trailing CR line.
  - Trailing space in the value; a double qualifier; an upper-case id; one valid and one invalid
    Intent; a Latin-1 byte in a different trailer.
  - The gate and the kernel copy agreed on 20 of 21. The exception is F4.
- No cleanroom redactions were needed.

## Findings

### intent-trailer-gate-review-round1-20261116-F1 (MED): the shared corpus does not pin the ported rules, so the parsers can drift with every test green

**Broken property:** the CI's acceptance claim that "parser parity is enforced … by the shared
corpus, which both parsers must satisfy", and the test comment that "a drift in either parser fails
a test".

**Evidence:**
- **The two tests check different things.** The kernel test consumes only column 2 (extraction) and
  reads the raw `.msg` bytes (`gitcmd_test.go:98-107`). The gate test consumes only column 3
  (pass/fail) (`intent-delivery-applicability.test.sh:39-50`). So:
  - The gate's *extraction* is never compared to column 2.
  - The kernel's *value rule* (`connector.go:189-203`, `intentIDRE`) is never driven by the corpus.
    The rows `malformed-id` and `unknown-provider` exercise only the gate's copy.
  - Half of the port, the value regex, therefore has a parity test on one side only.
- **Mutation runs.** These ran on a scratch clone at the reviewed HEAD, with the stock gate test
  (`bash scripts/tests/intent-delivery-applicability.test.sh`).
  - Positive control: replacing the blank-line check with an unconditional `exit 0` is **killed**.
  - Neutral control: a whitespace-only edit **survives**.
  - Then each of the following **survived**:
    - M1: delete `if (n < 3) exit 0` (the "≥ 3 lines" rule).
    - M2: `(records|specdir)` → `(specdir)`. The gate would now refuse a legitimate
      `Intent: records:CI-…` that the kernel binds; the corpus has no `records:` row.
    - M3: slug `{1,62}` → `+`. The gate would pass ids the kernel rejects as malformed.
    - M4: suffix `{6}` → `+`.
    - M5: id character classes made case-insensitive.
- **Kernel side.** On the same clone, replacing `if len(lines) < 3` with `if false && len(lines) < 3`
  in `gitcmd.go` leaves `TestIntentTrailers_SharedCorpus` green.
- **Why the ≥ 3-lines rule is unpinned.** The `too-short` case (`subject\nIntent: …`) is rejected by
  the blank-line rule whether or not the length rule exists. The only shape that discriminates the
  length rule, `\nIntent: CI-…` (an empty subject), is absent.
- The ≥ 12-case calibration counts rows, not discriminated rules, so it does not detect any of these
  gaps.

### intent-trailer-gate-review-round1-20261116-F2 (MED): the reason for replacing condition (3) is wrong about where the gate runs, and `make check` still never examines HEAD's trailer

**Broken property:** the condition-(3) substitution recorded in `DEF-INTENT-TRAILER-BLOCK-001` and
in the CI's "Explicit exclusions".

**Evidence:**
- **Where the gate runs.** The only caller of `scripts/intent-delivery-applicability.sh` is
  `scripts/delivery-enforce.sh:39`. It runs immediately after
  `(cd kernel && go run ./cmd/proofbound sync all)` on line 38, which already requires the embedded
  Postgres and a synced ledger.
  - So the stated reason ("a live `proofbound intent check` inside make check needs embedded
    Postgres + a synced ledger, making the gate environment-dependent") does not apply where the gate
    actually runs: the environment is already present there.
  - At that call site, `sync all` does fail loudly on an unresolvable trailer (`connector.go:127-131`
    aborts on a `resolvedCommit` error). Resolvability is therefore covered there by `sync`, not by
    the substitute.
- **Where it does not run.** `make check` (Makefile line 3 target list) runs only the gate's
  *self-test* via `hooks-test`. No `make check` target, git hook (`core.hooksPath` unset, no
  non-sample `.git/hooks`), or Claude hook runs the gate against HEAD.
  - The deferral's original harm statement ("make check is green either way") is still true after the
    fix. A governed commit whose `Intent:` line sits in a body paragraph still lands with `make check`
    green.
  - Only an explicit `make delivery-enforce` run, or a manual run, would refuse it.
- **Effect.** Original condition (3), "make check gains a kernel-side bind check", existed to close
  exactly that routine-gate gap. The substitute (corpus parity) addresses a different property
  (grammar agreement), and per F1 it does not fully establish even that. The deferral row records the
  condition as met "by corpus parity" while the routine gate's blind spot remains.

### intent-trailer-gate-review-round1-20261116-F3 (LOW): the gate reads a different commit object than the kernel when replace refs exist (silent false-pass reproduced)

**Broken property:** "the gate must accept exactly the commits whose Intent claim the kernel binds"
(`intent-delivery-applicability.test.sh:35-37`).

**Evidence:**
- The kernel runs git with `--no-replace-objects` and a scrubbed `GIT_*` environment
  (`gitcmd.go:404-406`, `gitEnvironment`). The gate runs plain `git show -s --format=%B` with
  neither.
- **Repro** (scratch repo):
  1. Commit `body-paragraph.msg` on an applicable path. The gate refuses it ("OUTSIDE the final
     trailer block").
  2. Run `git replace <sha> <sha'>`, where `sha'` is the same tree with the message
     `subject\n\nIntent: CI-fixture-item-acde12\n`.
  3. The gate now prints `applicable=true … intents=CI-fixture-item-acde12`. Meanwhile
     `git --no-replace-objects show -s --format=%B <sha>` fed to the kernel parser yields no trailer.
- This is the original harm class: the gate passes and the kernel binds nothing, with no error. It
  requires a local replace ref, which is not pushed by default, so the severity is LOW. An inherited
  `GIT_DIR` or `GIT_*` override is a similar unpinned divergence.

### intent-trailer-gate-review-round1-20261116-F4 (LOW): residual parity and diagnostic gaps outside the corpus

**Broken property:** grammar parity, and the accuracy of the gate's diagnostic.

**Evidence:**
- **Non-UTF-8 message.** The message `subject\n\nSigned-off-by: J\xf6rg\nIntent: CI-fixture-item-acde12\n`
  (Latin-1 byte in a different trailer) passes the gate (`applicable=true`). The kernel rejects the
  whole commit at `parseScalars` ("scalar is not valid UTF-8", `gitcmd.go:224-227`), so no claim
  binds.
  - Within `delivery-enforce` this surfaces loudly at `sync all` first. The gate itself, however,
    reports a pass for a commit the kernel cannot read.
- **CRLF, tab-separated, and trailing-CR messages.** The kernel and the gate agree on refusing them.
  The gate's diagnostic, though, says the Intent line is "OUTSIDE the final trailer block" because its
  fallback `grep -q '^Intent:'` matches. For the CRLF case the line *is* in the final block; the real
  cause is the `\r`. The diagnostic misattributes the cause, which steers the author to the wrong
  remedy.

## What holds

- **The awk grammar port matches `intentTrailers` on every UTF-8 input probed:**
  - It drops trailing empty lines, which is equivalent to `TrimRight("\n")`.
  - It walks back while lines match `^[A-Za-z0-9-]+: .+$`.
  - It requires a preceding empty line, rejects `i < 0` (a message made only of trailers), and
    matches the key `Intent` exactly and case-sensitively.
  - `Intent-X:`, `intent:`, a tab after the colon, a whitespace-only separator, and a value of only
    spaces all behave identically. A space-only value is extracted by both and then rejected by both
    value rules.
- **The value regex matches `resolvedCommit` + `intentIDRE` on every probed value:**
  - Accept and reject agree for bare, `records:`, `specdir:`, double-qualified, unknown-provider,
    upper-case, trailing-space, and non-ASCII (é, fullwidth, superscript) values.
  - In this box's `C.UTF-8` locale, bash `[a-z]` does not match `é` or `B`.
  - The `en_US` locale is not installed, so locale-collation divergence was not testable.
- **The gate's real-history acceptance boundary holds:**
  - `a028984` is refused with the correct diagnostic, and `a5a085b`, `07c384d`, and `23a6fcc` pass.
  - `b0f5e5d` is `applicable=false`, because kernel `SPEC.md` paths are excluded by
    `intent-applicability.sh:44`, which is pre-existing policy.
  - I found no regression for a commit that legitimately bound before: every correctly-placed trailer
    form in the corpus still passes.
- **The positive controls are real.** Corrupting the blank-line rule kills the gate test, and both
  corpus tests run ≥ 12 rows. The deferral row is correctly left `OPEN` pending this verdict.
