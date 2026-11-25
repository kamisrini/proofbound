---
schema: vera.verdict.v1
verdict_id: ui-reviewer-r1-review-20261125
status: ACCEPTABLE
reviewed_commit: 93e337256c3da120ca7cb681ae17fbe5f95d790d
findings:
  - finding_id: ui-reviewer-r1-review-20261125-F1
    severity: MED
  - finding_id: ui-reviewer-r1-review-20261125-F2
    severity: MED
  - finding_id: ui-reviewer-r1-review-20261125-F3
    severity: MED
  - finding_id: ui-reviewer-r1-review-20261125-F4
    severity: MED
  - finding_id: ui-reviewer-r1-review-20261125-F5
    severity: MED
  - finding_id: ui-reviewer-r1-review-20261125-F6
    severity: LOW
  - finding_id: ui-reviewer-r1-review-20261125-F7
    severity: LOW
  - finding_id: ui-reviewer-r1-review-20261125-F8
    severity: LOW
  - finding_id: ui-reviewer-r1-review-20261125-F9
    severity: LOW
  - finding_id: ui-reviewer-r1-review-20261125-F10
    severity: LOW
  - finding_id: ui-reviewer-r1-review-20261125-F11
    severity: LOW
artifact_path: docs/verification/verdicts/ui-reviewer-r1-review-20261125.md
artifact_sha: 7acfdf190e5a8f86b86e9e84cce25319604b963d99c189ab5db5f91a493af916
---
## Provenance
Non-author review (Law 9), the single code round for R1 (Law 11), of the uncommitted diff on top of HEAD: kernel/internal/cli bench.go (new), workspaces_http.go, ui.go, starter.go (reviewSeal split into readSealDraft and reviewSealWith; mergeDecisions carries tests), cli.go (the CLI seal runs the tests before opening the store), tests.go (progress callback; the tests' environment drops PROOFBOUND_* and DATABASE_URL), uikit review.html, app.js, app.css and the tests, against docs/plans/ui-reviewer-spec.md draft 2 §1. The reviewer ran the bench tests with -race. Original verdict NEEDS_WORK (no HIGH, five MED); this record is ACCEPTABLE after every MED was fixed and verified by the author's mutation sweep, calibration stated first (positive controls each died: the seal skips the second test run, an undecided part not caught is sealed, Save drops the tests, the pricing line loses its area flag, a reviewed commit can run its tests, the tests see PROOFBOUND_KEY; neutral: a comment edit survived), and by a run on a copy of the live demo through the bench's HTTP calls (BLOCKED and PASS as `check` reads them).

## Status
ACCEPTABLE after fixes: F1-F5 MED fixed with regression tests; F7, F8, F9 LOW fixed; F6, F10, F11 LOW recorded (spec §8).

## Findings
- **F1 (MED), fixed.** Run was offered for an invalid tests file: the page shows the reason and no Run; the endpoint refuses (TestBenchInvalidTestsFile).
- **F2 (MED), fixed.** The bench actions did not require the commit to wait for review: every action now requires NO REVIEW (TestBenchTestsFlow, reviewed commit).
- **F3 (MED), fixed.** The flags depended on the reviewer's git settings: `git show` runs with fixed options; a flagged path not shown is listed (TestBenchFlagsPastTheCut).
- **F4 (MED), fixed.** Spec §1.4 said the commit, the code read the parent: §1.4 amended (the parent is what the area check enforces); a commit changing its own spec is marked.
- **F5 (MED), fixed.** Missing tests added (reload, page loads during a job, reviewed commit, invalid tests file, the real seal refusing a part only the second run finds, flags past the cut); a test that could not fail now isolates its reason.
- **F6 (LOW), recorded.** The CLI runs the tests before checking declared_reviewer.
- **F7 (LOW), fixed.** The tests job's draft merge takes the lock; draft.json is written atomically.
- **F8 (LOW), fixed.** A started seal finishes even if the server stops.
- **F9 (LOW), fixed.** Buttons the server refused stay disabled after a job.
- **F10 (LOW), recorded.** The job status is readable with the agent's read token (summary lines only).
- **F11 (LOW), recorded.** The queue computes the area per commit on each load; the needs-work list shows the gate line, not the reviewer's name.
