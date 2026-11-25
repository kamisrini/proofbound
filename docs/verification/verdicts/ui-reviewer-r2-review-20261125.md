---
schema: vera.verdict.v1
verdict_id: ui-reviewer-r2-review-20261125
status: ACCEPTABLE
reviewed_commit: 93e337256c3da120ca7cb681ae17fbe5f95d790d
findings:
  - finding_id: ui-reviewer-r2-review-20261125-F1
    severity: HIGH
  - finding_id: ui-reviewer-r2-review-20261125-F2
    severity: MED
  - finding_id: ui-reviewer-r2-review-20261125-F3
    severity: MED
  - finding_id: ui-reviewer-r2-review-20261125-F4
    severity: MED
  - finding_id: ui-reviewer-r2-review-20261125-F5
    severity: LOW
  - finding_id: ui-reviewer-r2-review-20261125-F6
    severity: LOW
  - finding_id: ui-reviewer-r2-review-20261125-F7
    severity: LOW
artifact_path: docs/verification/verdicts/ui-reviewer-r2-review-20261125.md
artifact_sha: 095e2dea32392fedc74d10f98c91d9e3a65700bdf3bd79c3930de9858152c018
---
## Provenance
Non-author review (Law 9), the single code round for R2 (Law 11), of the uncommitted diff on top of HEAD: kernel/internal/cli change.go (Changes page, Issue), bundle.go (warrantyBundle split into bundleCompute and the write; CLI behaviour unchanged, checked by the reviewer), ui.go, uikit change.html, audit.html, layout.html, app.js; kernel/cmd/proofbound-verify (the warranty check compiled to WebAssembly, the page and its assembler); kernel/internal/store lock_js.go; scripts/release-build.sh; against docs/plans/ui-reviewer-spec.md draft 2 §2. The reviewer ran the R2 tests, `GOOS=js GOARCH=wasm go vet` and the release-build test. Original verdict NEEDS_WORK (one HIGH, three MED); this record is ACCEPTABLE after the HIGH and every MED were fixed, with mutation controls stated first (positive, each died: Issue writes a partial release, Issue ignores uncovered code, Issue offered to a reviewer, Review this only for the last commit, a subject alone counts as a release, the verify page takes the key from the file; neutral: a comment edit survived), and an Issue on a copy of the live demo verified VALID by `warranty verify --repo` from a fresh clone.

## Status
ACCEPTABLE after fixes: F1 HIGH and F2, F3 MED fixed; F4 MED partly fixed and the rest recorded with reason; F5, F7 LOW fixed; F6 LOW recorded (spec §9).

## Findings
- **F1 (HIGH), fixed.** Issue checked only the bundle's claiming entries, so a period with uncovered code (NOT WARRANTED by `check`) could be issued. Issue now requires every C1 line PASS and no uncovered-code exception (TestIssueWarranty).
- **F2 (MED), fixed.** The Changes page offered Issue from one change's last commit; it now follows BR-audit-room-ui O-7: Review this for any waiting commit, Issue only when the viewer may issue the period's warranty, the reason otherwise (TestChangePage).
- **F3 (MED), fixed.** Issued status came from worktree files and a commit subject; now from HEAD's tree, and a release commit must add exactly its folder; the dirty check runs first.
- **F4 (MED), partly fixed.** Added the author-viewer case and the release-build assertion that the page is built and listed. Recorded: Issue's bundle equals `warranty bundle` only through the shared bundleCompute (a fixture needs the record store); the WebAssembly glue and page script are not run by a test on this machine (no node or browser), so the owner opens the page once by hand.
- **F5 (LOW), fixed.** A HEAD that moves during Issue is refused.
- **F6 (LOW), recorded.** The WebAssembly links the database driver through the connectors (9 MB, about 3 MB compressed).
- **F7 (LOW), fixed.** Page wording (the review account is the file's own), the issue message names the warranted commit, an unreadable file shows a line, the loader check ignores case; commitPaths names what it failed to commit.
