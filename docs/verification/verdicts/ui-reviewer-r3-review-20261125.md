---
schema: vera.verdict.v1
verdict_id: ui-reviewer-r3-review-20261125
status: NEEDS_WORK
reviewed_commit: fb527cf09efe32513a1a64e319dc01b61e9502e5
findings:
  - finding_id: ui-reviewer-r3-review-20261125-F1
    severity: MED
  - finding_id: ui-reviewer-r3-review-20261125-F2
    severity: MED
  - finding_id: ui-reviewer-r3-review-20261125-F3
    severity: LOW
  - finding_id: ui-reviewer-r3-review-20261125-F4
    severity: LOW
  - finding_id: ui-reviewer-r3-review-20261125-F5
    severity: LOW
  - finding_id: ui-reviewer-r3-review-20261125-F6
    severity: LOW
artifact_path: docs/verification/verdicts/ui-reviewer-r3-review-20261125.md
artifact_sha: 4b3ff2d6a0ff10c95abb73cd934ba90fb3dc578af47f2c069c6af741d7d562aa
---
## Provenance
Non-author review (Law 9), the single code round for R3 (Law 11), by agent-code-reviewer-ui-r3-20261125, which did not write the change. Reviewed: the uncommitted diff on top of HEAD in kernel/internal/cli bench.go, bench_test.go, ui.go, workspaces_http.go, SPEC.md (CLI-INV-76), uikit review.html, static/app.css, static/app.js; docs/releases/r3 RELEASE-NOTES.md and DEMO-AGENT-LIVE.md; against docs/plans/ui-reviewer-spec.md §10. The reviewer ran TestBenchLayout, TestBenchFlags and TestBenchTestsFlow (pass), checked `git show` output for binary, empty-added, deleted, mode-only, spaced and tab-named files in a scratch repository, and ran a mutation sweep on TestBenchLayout with calibration stated first. Positive controls, each died: jumpLineOf returns the line after the first; the area hint removed; data-line rendered wrong; the tests box links line 1; "Tested:" shown for every hunk. Neutral control, survived as it must: the highlight colour changed in app.css. Also survived (gaps, F4): the whole-change sentence inverted; the JS matching on path only; sticky removed. Every file was restored byte for byte (cmp). Endpoints, draft, seal, records and `check` are unchanged; the three workspaces_http.go strings are asserted by no test, CLI output or doc.

## Status
NEEDS_WORK: F1 and F2 (MED) must be fixed; F3 to F6 (LOW) are recorded.

## Findings
- **F1 (MED).** A whole-file part can never be shown in the code, and the page then says something false. bench.go gives line 0 only to a `+++ b/<path>` header, but `git show` prints no such header for any whole-file part: a binary file ("Binary files /dev/null and b/x differ"), an empty file added, a mode-only change, or a deleted file (`+++ /dev/null`). So every link for "binary file", "file added", "mode change" or "file deleted" misses, and app.js:208 replaces it with "Not in the view above (past its end)", although the view was not cut. The same happens for a path git quotes (a tab or a quote in the name: `+++ "b/ta\tb.txt"`). TestBenchLayout only checks that a normal file's header has `data-line="0"`; its fixture has no whole-file part, so it cannot catch this. Fix: anchor line 0 on the `diff --git` line (or the first header line of each file), and make the fallback say "not shown in this view" unless the 1,500-line cut is the reason; add a binary or mode-change part to the test.
- **F2 (MED).** The tests summary (review.html:96) states things that did not happen. Baseline can be pass, fail, timeout or unreliable; Whole can be caught, not caught, not run or undo failed (tests.go). When the baseline fails or times out, Whole is "not run" and the page says "With all the new code removed, they do not fail (not run), so they do not test this change": nothing was run. With Whole "undo failed" it says they "do not fail" when the undo never applied. With baseline "unreliable" it says "The tests do not pass on this change (unreliable)" although they passed first, and it can go on to "they fail, so they do test it", with "Tested:" rows (review.html:98) for results that do not count. Fix: one sentence for each value, e.g. not run means "not tried because the tests do not pass", undo failed means "Proofbound could not remove it all at once", unreliable means "they passed, then failed when run again unchanged, so no result counts".
- **F3 (LOW).** A "removed lines o-p" tests part (code deleted and no test noticed) links to the file header, which can be far from the deletion, and the deleted lines have no flag. The behaviour is honest but not where the spec says ("first added line").
- **F4 (LOW).** TestBenchLayout does not check the whole-change sentence (an inverted condition survived), and it cannot test app.js or app.css (the path-only match and the sticky removal survived). There is no JS harness on this machine, so the owner checks the link and the sticky column once by hand.
- **F5 (LOW).** The spec still uses the old words: ui-reviewer-spec.md:74 (*tests did not notice*) and §10.2 at line 282 ("*Tests did not notice* box"), while the page now says "No test checks this". The BR O-8 statement still reads correctly.
- **F6 (LOW).** A part that was not run (past 20 hunks, or baseline failed) is titled "No test checks this", which is not known. The "(not run)" note and "could not check this part" line soften it. The queue says "every part tested" when the baseline is unreliable.

Checked and fine: new line numbers across several hunks, and through files with spaces (the trailing tab is trimmed); the 1,500-line cut (the fallback text is right there); html/template escapes data-path and data-jump-path the same way, neither name is classed as a URL attribute, and getAttribute compares the decoded strings, so there is no mismatch and no injection; the sticky column has no overflow:hidden ancestor, its grid uses align-items:start, the columns have min-width:0 and the diff scrolls itself, so the page never scrolls sideways; on narrow screens the evidence comes after the decisions.
