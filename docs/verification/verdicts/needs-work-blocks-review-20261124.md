---
schema: vera.verdict.v1
verdict_id: needs-work-blocks-review-20261124
status: ACCEPTABLE
reviewed_commit: bdf83e1b222e29c7d23dec793e707f2e304471e9
findings:
  - finding_id: needs-work-blocks-review-20261124-M1
    severity: MED
  - finding_id: needs-work-blocks-review-20261124-M2
    severity: MED
  - finding_id: needs-work-blocks-review-20261124-L1
    severity: LOW
  - finding_id: needs-work-blocks-review-20261124-L2
    severity: LOW
  - finding_id: needs-work-blocks-review-20261124-L3
    severity: LOW
  - finding_id: needs-work-blocks-review-20261124-L4
    severity: LOW
  - finding_id: needs-work-blocks-review-20261124-L5
    severity: LOW
  - finding_id: needs-work-blocks-review-20261124-L6
    severity: LOW
artifact_path: docs/verification/verdicts/needs-work-blocks-review-20261124.md
artifact_sha: b70a031445c317ca7eb332c48c3c979454c53ac82b0932b460f4c13f414d7d8f
---
## Provenance
Non-author review (Law 9), the single code round (Law 11), read-only and without running tests, of the staged diff on top of HEAD: docs/plans/needs-work-blocks-spec.md draft 2 (the contract), kernel/internal/projections (verdict_status column, in-place backfill in Ensure, upsert, Snapshot, intent report), kernel/internal/gates (readinessDecision and its SQL), kernel/internal/warranty (CheckRule), kernel/internal/cli/drift.go, the SPEC rows GATE-INV-17, P-INV-35 and W-INV-3, and the sdd end-to-end test. Every consumer of obligation_verdicts_view and of the readiness gate was traced. The author's mutation check, calibration stated first: positive control (the NEEDS_WORK case dropped from readinessDecision) died; neutral control (a comment edit) survived; eight further mutants (unknown status, warranty status check, on-conflict status update, the backfill, the gate SQL filter, the drift filter, the intent report mapping, and the agent answer's BLOCKED case) all died, the on-conflict one only end to end.

## Status
ACCEPTABLE: no HIGH. M1 and M2 fixed before commit; L1-L6 recorded (Law 11).

## Findings
- **M1 (MED), fixed.** cli/workspaces.go agentNext: a commit BLOCKED by a NEEDS_WORK verdict with every promise SATISFIED still answered `done`. A BLOCKED commit now adds a fix_code step carrying the gate's reason (CLI-INV-48 amended; workspaces_test.go TestAgentNext).
- **M2 (MED), fixed.** The change intent's allowed scope omitted kernel/internal/cli (drift) and named a rebuild instead of the in-place migration.
- **L1 (LOW).** drift groups by verdict id: one NEEDS_WORK and one ACCEPTABLE verdict under different ids still list as warranted; bool_and skips NULL.
- **L2 (LOW).** "Latest revision per verdict id" holds per obligation row; a re-seal that drops an obligation row keeps that row's old status.
- **L3 (LOW).** The backfill UPDATE runs on every Ensure.
- **L4 (LOW).** No fresh-ledger-agreement test (spec section 4) and no pre-fix warranty through warranty verify end to end.
- **L5 (LOW).** The intent report maps NULL status to NEEDS_WORK (fails closed).
- **L6 (LOW).** agent-demo.sh must be re-run before release.
