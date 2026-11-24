---
schema: vera.verdict.v1
verdict_id: agent-track-record-review-20261124
status: NEEDS_WORK
reviewed_commit: bc0ec22dc295467bcb4b532d3761578f4eedaf22
findings:
  - finding_id: agent-track-record-review-20261124-M1
    severity: MED
  - finding_id: agent-track-record-review-20261124-M2
    severity: MED
  - finding_id: agent-track-record-review-20261124-L1
    severity: LOW
  - finding_id: agent-track-record-review-20261124-L2
    severity: LOW
  - finding_id: agent-track-record-review-20261124-L3
    severity: LOW
  - finding_id: agent-track-record-review-20261124-L4
    severity: LOW
  - finding_id: agent-track-record-review-20261124-L5
    severity: LOW
artifact_path: docs/verification/verdicts/agent-track-record-review-20261124.md
artifact_sha: e60de4149833b09d9c01f5e205a2663e35d1fba414d15be58229264405dcf0ca
---
## Provenance
Non-author subagent review (Law 9), the single code round (Law 11), of the uncommitted working tree on top of HEAD: review outcomes by agent (kernel/internal/cli/track.go, track_test.go; the per-commit counts C10, C11 and C12 hand over in area.go, tests.go, provenance.go; audit.go's unexported track field; the Audit room card in ui.go, uikit/audit.html, app.css; `report agents` in cli.go; the sdd package's provenance_e2e_test.go). The spec had one earlier non-author round (NEEDS_WORK: 2 HIGH, 4 MED, folded in at build, docs/plans/agent-track-record-spec.md section 7). No repository file was modified; `make check` was not run; only `go test -run 'TestTrackRecord|TestWriteTrack|TestAuditC1[0-2]' ./internal/cli` ran.

## Status
NEEDS_WORK: two MED, both fixed (M2 with a regression test). Checked and found fine: the transcript and attestations are unchanged (track is unexported; UI fields are json:"-"); every untrusted-record path in C10, C11 and C12 hands over nothing; counts are set once per commit per control; agents dedupe and per-commit sessions are correct; the as-of view uses rep.Head with the shared object database; `report agents` argument parsing refuses a flag as base and extra arguments, and resolves the base before opening the database; html/template escapes names and session ids; the four caveats are honest.

## Findings
- **M1 (MED).** cli.go commandReportAgents did not call attachOutcomes, so every exception counted as open, unlike the Audit room. Fixed: attachOutcomes before computeTrack.
- **M2 (MED).** report agents did not sync, and a commit with no reviewers in the ledger was treated as independent while its OV verdict counted ACCEPTABLE. Fixed: sync first (a failed sync is printed); independence also judged from each verdict's declared_reviewer; test in track_test.go TestTrackRecord.
- **L1 (LOW).** The two survivors called equivalent are test gaps: a verdict re-digested to another commit's path, and a verdict deleted after the as-of commit. Recorded.
- **L2 (LOW).** The Audit room table lacked Commits, Code without a spec and Open exceptions. Fixed.
- **L3 (LOW).** The session table showed a not-independent review as not reviewed. Fixed: its own column.
- **L4 (LOW).** C12's traced-line count discarded a diff error and its session counts are not re-matched to the current diff. The error now fails the audit; the rest is recorded.
- **L5 (LOW).** An area set by an agent counts its hunks against that area. Recorded (the caveat says undecided).
