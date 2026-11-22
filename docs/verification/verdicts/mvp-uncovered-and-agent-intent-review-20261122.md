---
schema: vera.verdict.v1
verdict_id: mvp-uncovered-and-agent-intent-review-20261122
status: NEEDS_WORK
reviewed_commit: 01eb46badff7de7617d018778d96314e141dc8d9
findings:
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F1
    severity: HIGH
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F2
    severity: MED
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F3
    severity: MED
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F4
    severity: MED
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F5
    severity: LOW
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F6
    severity: LOW
  - finding_id: mvp-uncovered-and-agent-intent-review-20261122-F7
    severity: LOW
artifact_path: docs/verification/verdicts/mvp-uncovered-and-agent-intent-review-20261122.md
artifact_sha: fc35632e03d57dd8bbfdbb397d84cf616408c16ee804478fc2bfc0277adeef58
---
## Provenance
A separate Sonnet reviewer agent (not the author), one code round under Law 11, reviewing the
uncommitted changes on top of reviewed_commit under `CI-mvp-uncovered-code-581081` (Change A) and
`CI-mvp-agent-intent-3k7q2w` (Change B). Ran: `go vet` on internal/cli and the sdd package (clean),
`go test ./internal/cli/` (ok), a scratch repo driving the real built binary and the installed hook
scripts through real `git commit` (plain `-m`, `-a`, `<path>`, editor, `--amend`), and `git diff-tree
--cc` experiments for merges and renames. Calibration (stub `uncoveredCommits` to return nil, expect
`TestPRCheckEndToEnd` to fail) was run on a scratch copy of kernel/, not the worktree, and was
INCONCLUSIVE: an orphaned embedded postgres from an interrupted test run held port 55432, so the test
failed on a store migration error, not on the assertion. By inspection the mixed-PR assertions
(NOT WARRANTED headline, UNCOVERED row) cannot pass with the stub. No tracked file was modified.

## Status
NEEDS_WORK: one HIGH (hooks are bypassed by `git commit -a` and `git commit <path>`) and two MED hook
failure modes plus one MED over-broad exemption.

## Findings
- **F1 (HIGH).** kernel/internal/cli/task.go `stagedCode` (~line 128) runs `git diff --cached` under
  `repositoryGitEnv()`, which strips every `GIT_*` variable including `GIT_INDEX_FILE`. For
  `git commit -a` and `git commit <path>` git runs hooks against a temporary index, so the hook reads
  the real (still empty) index, sees no staged code, does not stamp and commit-msg passes. Reproduced:
  with an active task, `git commit -qam viaA` and `git commit -qm viaPath a.go` produced unclaimed code
  commits with both hooks installed; `git add` then `git commit -m` stamped correctly. This is the most
  common agent commit form, so the feature's central promise (every code commit stamped) fails; the PR
  check would still flag UNCOVERED, so it is not a trust break, but the agent flow is broken. The unit
  tests only use `git add`. Fix: in stagedCode keep `GIT_INDEX_FILE` (and GIT_DIR/GIT_WORK_TREE) from the
  environment when running as a hook, and add a test that runs a real `git commit -a`.
- **F2 (MED).** task.go `withTrailer` (~line 215): with an empty message (editor flow; only `#`
  comments) it emits `["", "Intent: sdd:x"]`, so after the user types a subject on line 1 the message is
  `subject\nIntent: sdd:x` with no blank line (reproduced). gitcmd.intentTrailers needs at least 3 lines
  and a blank line before the block, so the kernel does not bind it (UNBOUND at check), while commit-msg
  accepts it because `claimLines` is position-lenient. Fix: for an empty body emit two blank lines, and
  have `commit-msg` reject a claim the kernel grammar would not parse (reuse intentTrailers' rules).
- **F3 (MED).** starterkit/commit-msg: only checks that `proofbound` is on PATH. An older binary without
  the `hook` subcommand is on this machine's PATH (/home/thamm/.local/bin/proofbound); every commit,
  housekeeping included, was then refused with a usage dump and exit 2 (reproduced), and
  prepare-commit-msg silently does nothing. Fix: have the script detect that failure (exit 2 / usage)
  and print "upgrade proofbound", or call `proofbound version` and require a minimum.
- **F4 (MED).** check.go `housekeepingPath` (~line 222): `.proofbound/` is exempt wholesale, which
  includes `.proofbound/sdd-profile.json` (defines what is governed) and `.proofbound/keys/` (the
  declared warranty keys). An unclaimed PR commit that edits either reads clean. Fix: exempt only what
  init/review writes (`.proofbound/drafts/`, `.proofbound/keys/*.pub` if intended) and make the profile
  count as code, or record the decision explicitly in CLI-INV-33.
- **F5 (LOW).** `housekeepingPath` also exempts every file under `docs/`, `specs/`, `openspec/` and any
  `*.md`, regardless of type: `docs/gen.py`, `specs/x/run.sh` pass, and so do behaviour-bearing
  Markdown files such as AGENTS.md, `.claude/agents/*.md` and this repo's own `go:embed`ded
  `starterkit/*.md`. Fix: restrict by extension inside those dirs, or document the residual risk.
- **F6 (LOW).** Tests that cannot fail on behaviour: `TestInitHookScriptsDelegate` only greps the
  script text; nothing runs the scripts under real git (amend, `-a`, editor, old binary). CLI-INV-33's
  merge, rename and root-commit claims have no test (I confirmed with `git diff-tree --cc` that a clean
  merge lists nothing and a rename lists both paths, so the code is right, but only
  `TestHousekeepingPath` is cited). Fix: add one e2e test with a merge and a rename.
- **F7 (LOW).** `writeNew` never overwrites, so a repo initialised before this change keeps the old
  commit-msg hook (with the retired `PROOFBOUND_NO_INTENT` opt-out) and gets no upgrade path; and
  `git commit --amend` with nothing newly staged is neither stamped nor refused (the hook sees only the
  delta), so an unclaimed code commit can be re-worded past commit-msg. Both are advisory, the PR check
  still catches them. Fix: print a note from `init` when an existing hook differs, document the amend gap.
