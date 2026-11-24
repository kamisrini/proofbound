---
schema: vera.verdict.v1
verdict_id: tests-can-fail-review-20261123
status: NEEDS_WORK
reviewed_commit: af89d9958295bdf0e767936107db0bea4b7320d6
findings:
  - finding_id: tests-can-fail-review-20261123-F1
    severity: HIGH
  - finding_id: tests-can-fail-review-20261123-F2
    severity: MED
  - finding_id: tests-can-fail-review-20261123-F3
    severity: MED
  - finding_id: tests-can-fail-review-20261123-F4
    severity: MED
  - finding_id: tests-can-fail-review-20261123-F5
    severity: MED
  - finding_id: tests-can-fail-review-20261123-F6
    severity: LOW
  - finding_id: tests-can-fail-review-20261123-F7
    severity: LOW
  - finding_id: tests-can-fail-review-20261123-F8
    severity: LOW
  - finding_id: tests-can-fail-review-20261123-F9
    severity: LOW
artifact_path: docs/verification/verdicts/tests-can-fail-review-20261123.md
artifact_sha: 71bf62bcf8d59ad61da87e8d5c18ff1e822a932a2b74bc192db9bd365271cf37
---
## Provenance
Non-author review (Law 9), the single code round (Law 11), of the staged diff on top of af89d99: the spec including its §7 amendments, CLI-INV-66..68, tests.go, tests_test.go, and the starter.go, audit.go, acts.go, ui.go and frameworks.go changes; also the bench seal path (workspaces_http.go, ui.go), the deferral row and the BR/CI records. Probes and mutations ran in a scratchpad copy of the kernel (review-tcf/kernel); no repository file was changed, `make check` was not run, port 55432 was not used. Baseline `go test ./internal/cli -run 'TestCommitTests|TestSealAccountsForTests|TestAuditC11'` passes. Calibration stated before the count: positive control (undo returns true without applying) died; neutral control (an error message reworded) survived. 17 mutants against the three new tests: 8 died, 9 survived (F5); survivors were not run against the whole package, so M22 and M24 may be caught by older tests.

## Status
NEEDS_WORK: one HIGH and four MED.

## Findings
- **F1 (HIGH).** tests.go:228-276 (commitTests). The judged commit can make every hunk and the whole change "caught" with a test that never exercises the code, and C11 then holds on a signed record. Shown by two probes: tests/t.sh = `test -z "$(git status --porcelain)"` (the worktree is dirty whenever anything is undone) and a test that fails once a marker such as /tmp/pbmark-$(git rev-parse HEAD) exists (passes on its first run, the baseline); every result was caught. Root cause: no neutral control (nothing re-runs with nothing undone after an undo run and requires a pass), the tests run in a git worktree and can ask git what was undone, and `git clean -fd` (no -x) keeps ignored files between runs. Flaky or state-leaking tests produce false "caught" by accident. Fix: run from git archive/checkout-index into a directory without .git; re-run the baseline after the undo runs and record "unreliable" (refuse Held) if it fails; use a fresh export per run; say in C11's Basis that a test pinning file contents still "catches" every hunk.
- **F2 (MED).** tests.go:148-152, 257-265, 489-491, decideTests 354-370. A "not run" hunk (past 20; binary, an added empty file or a mode change, shown with an empty src/empty.txt; or a silent `git apply -R` failure) is recorded with no decision and no NEEDS_WORK, then flags C11 for ever with no way to clear it; shown with 22 hunks (20 caught, 2 not run, whole caught: needsWork=false, no error). The spec does not say this blocks C11. Fix: require a decision for not-run hunks too and let C11 accept them (or refuse/NEEDS_WORK and say so); record an apply failure as its own result; undo binary and mode-change files with the --binary patch already built.
- **F3 (MED).** Bench seal (workspaces_http.go:366-412; spec §7 A5). The bench overwrites draft.json with a ledger draft whose Tests is nil, erasing the reviewer's decisions, so every bench seal with a not-caught hunk is refused and A5's remedy cannot work. It also runs the commit's code inside the HTTP request while holding the ledger gate, for up to 22 × timeout_seconds, with no warning, and the preview omits the TESTS- file. Fix: carry Tests from the local draft into `edited`, or refuse a bench seal when tests are declared and say so; list the TESTS- record in the preview.
- **F4 (MED).** tests.go:176-192 (testsRunner). On timeout only sh is killed: with run `sh -c "sleep 3; touch $dir/late"; true` and a 300 ms timeout, the runner returned timeout after 2.3 s and $dir/late appeared afterwards. The grandchild keeps running in the worktree through later resets and runs and outlives the worktree's removal; the existing timeout test passes only because dash execs a lone sleep. Fix: Setpgid and a cmd.Cancel that kills the process group; test with a compound command.
- **F5 (MED).** tests_test.go strength: nine mutants survived the new tests: M2 seal NEEDS_WORK on whole not caught; M3 C11 independence; M6 GIT_* removal; M14 review draft printing; M16 the 20-hunk limit; M21 record commit must equal the judged commit; M22 C6 listing tests.json; M24 draft.json carrying tests; M25 C11 accepting accept with an empty reason. Add one case each.
- **F6 (LOW).** tests.go:196-204. Only GIT_* is removed: the judged code inherits PROOFBOUND_KEY (the probe printed it), and `git rev-parse --git-common-dir` gives it the reviewer's real .git. Inherent to running the commit's code and warned about, but drop PROOFBOUND_* and consider an export without .git.
- **F7 (LOW).** starter.go:268-280. review draft runs the whole sweep before refusing because draft.json exists (up to 22 wasted runs); check first.
- **F8 (LOW).** §3 says the TESTS record holds the "base revision" (it has no such field) and C11 does not check paths or timeout_seconds; a cancelled caller context returns "fail", counted as caught (only the inner DeadlineExceeded is a timeout); the label "binary or mode change" differs from C10's "binary file"/"mode change".
- **F9 (LOW).** frameworks.go C11 EU AI Act Art. 17(1)(d): "tests are shown to depend on the change" reads closer to an outcome given F1; prefer "a signed record of whether the tests failed with each hunk undone". The rest of the mapping, including the PW.8.2 caveat, is honest. BR-tests-can-fail-k2h06j is still proposed (no requirement review); the CI's acceptance boundary names a demo act not in this diff.
