---
schema: vera.verdict.v1
verdict_id: r2-demo-polish-review-20261119
status: ACCEPTABLE
reviewed_commit: c9818c80b1175430a88979850c8bb116896e83fc
findings:
  - finding_id: r2-demo-polish-review-20261119-F1
    severity: LOW
  - finding_id: r2-demo-polish-review-20261119-F2
    severity: LOW
  - finding_id: r2-demo-polish-review-20261119-F3
    severity: LOW
artifact_path: docs/verification/verdicts/r2-demo-polish-review-20261119.md
artifact_sha: dfaf24cc6ac0496f428479867b6c09817cfa7c65a1a4a5b0dab2387cb8170c65
---

## Provenance
Reviewer: independent non-author (Law 9), connector-tier review under Law 11. Read `git show c9818c8` in full for store.go, gates.go, cli.go, the three SPEC rows, and the new tests. Also read the embedded-postgres v1.31.0 logger path, `store/lock.go`, and `latestIntentProof`. Ran in a detached worktree at <scratch>/review-polish, since removed: store `TestEmbeddedServerOutputGoesToLogFile`, gates reason test, and cli `TestReportIntentResolvesAnSDDChangeName`. All passed. I did not run the sdd M3 e2e test (fixed port 55432) and did not run the demo script. Main checkout unchanged: git status shows only the two pre-existing untracked paths.

## Findings
No HIGH or MED defects.

- F1 (LOW) Log writer failure mode. The library copies the temp log into `Logger` in `flush()` and returns the Write error, so an unwritable `<DataDir>.log` now makes `Start`/`Stop` fail where stdout never could. It does not hang. In practice `acquireLock` already requires the lock directory to be writable, and the lock sits beside the data dir. So the new failure is reachable only when the data dir's parent is writable but the log path is not, such as a pre-existing read-only `.log` file. A start failure then reports the log path in the message. Acceptable. The log also grows without bound (append-only, a few lines per start), which the code comment acknowledges.
- F2 (LOW) The `commit not in the ledger; run sync git` reason fires whenever there is no `commit.recorded` event for that SHA. That is the same condition `latestIntentProof` uses, so it is correct for the readiness rule. The text assumes the cause is an unsynced commit, and it is silent on a SHA that was never a commit. Wording only.
- F3 (LOW) CLI-INV-17's test calls `reportIntentID` directly, so it would not fail if the `runCommand` wiring were reverted to `args[2]`. That wiring is one line. The demo exercises it end to end, per the commit message. A `runCommand`-level case would close it.

Checks on the specific questions:
- Reason on every readiness path:
  - Projection failure with a proof: `projection failed: <err>`, printed through `strconv.Quote`, so newlines are escaped.
  - Projection failure with no proof: returns an error, no Result, so no Reason is expected.
  - No proof: "commit not in the ledger".
  - Zero targets: "commit binds no change intent", BLOCKED.
  - PASS and BLOCKED: counts from the same query that decides the state.
  - `intent-reference-integrity` and `intent-verdict-integrity` return early with an empty Reason, as the SPEC says.
- Output parsers: nothing in the repo's scripts, Go code or Makefile parses `would_block`. The only hits are docs and a demo echo. The `reason="..."` field is appended after `would_block`, and it is omitted when empty, so existing lines are unchanged. Nothing breaks.
- CLI-INV-17:
  - Only an `sdd:` prefix triggers resolution, and anything else passes through untouched.
  - A resolve failure returns an error before `ReportIntent` runs, so nothing is reported.
  - HEAD is resolved with `rev-parse --verify HEAD^{commit}` and a repo-scoped git environment.
  - The SPEC row matches the code, test, and error behaviour.
- SPEC rows: INV-52, GATE-INV-16 and CLI-INV-17 each match their code and a named test.

## Verdict
ACCEPTABLE. No HIGH or MED findings, and the three LOWs are optional.
