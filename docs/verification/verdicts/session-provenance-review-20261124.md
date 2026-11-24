---
schema: vera.verdict.v1
verdict_id: session-provenance-review-20261124
status: NEEDS_WORK
reviewed_commit: 4f40fada36634aeb06c905c86de662433198ac56
findings:
  - finding_id: session-provenance-review-20261124-F1
    severity: HIGH
  - finding_id: session-provenance-review-20261124-F2
    severity: MED
  - finding_id: session-provenance-review-20261124-F3
    severity: MED
  - finding_id: session-provenance-review-20261124-F4
    severity: MED
  - finding_id: session-provenance-review-20261124-F5
    severity: MED
  - finding_id: session-provenance-review-20261124-F6
    severity: MED
  - finding_id: session-provenance-review-20261124-F7
    severity: MED
  - finding_id: session-provenance-review-20261124-F8
    severity: LOW
  - finding_id: session-provenance-review-20261124-F9
    severity: LOW
  - finding_id: session-provenance-review-20261124-F10
    severity: LOW
  - finding_id: session-provenance-review-20261124-F11
    severity: LOW
  - finding_id: session-provenance-review-20261124-F12
    severity: LOW
artifact_path: docs/verification/verdicts/session-provenance-review-20261124.md
artifact_sha: d151b7309c5bc53e7d273bc3db77c6d4408a5f49e1f22ab1470e6024cb9b0802
---
## Provenance
Non-author subagent review (Law 9), the single code round (Law 11), of the uncommitted working tree on top of HEAD (session-to-line provenance: kernel/internal/cli/provenance.go and its wiring, provenance_test.go, the sdd package's provenance_e2e_test.go, the agent demo). No repository file was modified; `make check` was not run; only targeted `go test ./internal/cli -run ...` ran; one git probe in the session scratchpad confirmed git's diff header formats for paths with a space and non-ASCII characters. The spec (docs/plans/session-provenance-spec.md draft 2) had one earlier non-author round (NEEDS_WORK, folded into draft 2). Calibration of the author's mutation sweep: positive control (unrecorded lines never listed) killed, neutral (a no-op statement) survived; all 24 behavioural mutants and the 9 review-fix mutants killed except one equivalent (core.quotePath on: the C-unquoting path then decodes the same path).

## Status
NEEDS_WORK: one HIGH, six MED, all fixed with regression tests (provenance_test.go TestProvenanceReviewFindings, TestAuditC12JudgesItsCommit); F8, F9, F11, F12 LOW fixed (one line each); F10 LOW recorded. Checked and found fine: seal never reads the draft's provenance; re-seal was already refused by the gates/ready existence check; the Review bench goes through reviewSeal; exception ID and UI label include provenance; frameworks_test ties the control count to controlQuestions; attestations are not recomputed; binary files, deletions and no-newline markers handled; session ids cannot escape the log folder; whole-file aggregation correct; records naming another session skipped.

## Findings
- **F1 (HIGH).** starter.go seal with provenance.go readProvRecord: a reviewer with no declared key sealed an unsigned PROV record, which never verifies, so C12 turned every keyless seal into an exception. Fixed: PROV is written only when the reviewer signs.
- **F2 (MED).** provenance.go commitAddedLines: a path with a space keeps git's trailing tab, and a non-ASCII path is C-quoted, so lines were signed as no recorded agent edit or the file was dropped. Fixed: core.quotePath off, trailing tab removed, unquote before b/.
- **F3 (MED).** provenance.go commitAddedLines: an added line starting with "++ " was read as a +++ header and the rest of the file lost. Fixed: hunks are read to the length their @@ header gives; headers only outside a hunk.
- **F4 (MED).** provenance.go provControl, audit.go: C12 read HELD with one traced commit and many untraced; DEMO-AGENT.md said "never an exception". Fixed: C12 is LISTED, the question is reworded, the page explains which records are exceptions.
- **F5 (MED).** provenance.go readSessionLogs and the diff scanner: an unreadable log or a sessions file aborted draft and seal; a 64 MB line failed the scanner. Fixed: warnings, not errors; no scanner limit.
- **F6 (MED).** provenance_test.go: no test for a record judging another commit, or for a path too long for one record. Fixed: both tested.
- **F7 (MED).** DEMO-MANUAL.md and DEMO.md listed seal output without the PROV line. Fixed.
- **F8 (LOW).** The settings warning describes the reviewer's checkout, not the commit. Fixed: worded "in this checkout".
- **F9 (LOW).** The hook followed a symlink planted at the log path. Fixed: Lstat refuses a non-regular file.
- **F10 (LOW).** Input over 8 MB and a path too long for one record are skipped silently. Recorded.
- **F11 (LOW).** Merge commits are diffed against their first parent, unstated. Fixed: stated in C12's Basis and CLI-INV-71.
- **F12 (LOW).** Session ids "." and ".." passed the regex (harmless). Fixed: no leading dot.
