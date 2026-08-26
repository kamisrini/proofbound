---
schema: vera.verdict.v1
verdict_id: task8-current-round1
status: ACCEPTABLE
reviewed_commit: 327219c000000000000000000000000000000000
findings: []
artifact_path: docs/verification/verdicts/task8-current-round1.md
artifact_sha: f89014e56c0fc7b3b1978e16060c3fcfcb12323418b8d49ccf84dec7fbb7b92f
---

ACCEPTABLE

- MED finding resolved: `LEFT JOIN` now retains rows with absent proofs and returns explicit errors for commits, checks, and sessions. Integration regression coverage passed.
- Boundary semantics are correct: `[now-7d, now)` using UTC event time. LOW: no dedicated exact-boundary regression test.
- Supersession logic is sound: `rev-list --all` plus detached `HEAD`; branch deletion and rewrites become unreachable and are retained as `[superseded]`. LOW: direct end-to-end tests for those transitions are absent.
- Mutation calibration remains a LOW verification limitation, accurately disclosed; no mutation result was improperly claimed.

Targeted unit and CLI tests passed, and the PostgreSQL `TestReportWeek` integration tests passed.
