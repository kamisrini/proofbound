# P6 Task 8 — `internal/projections` mutation evidence

This is calibrated author-side mutation evidence for the frozen implementation commit
`be58cd7` (`test: enforce projection report interval boundaries`). It is mechanical evidence only;
it is not the required non-author semantic verdict.

The exact package tree object is
`af840395f0211de0fe375f3d2ef1599e97e4a638` (`git rev-parse be58cd7:kernel/internal/projections`).
The complete tagged package suite passed against the same checkout:

```text
DATABASE_URL=postgres://proofbound@127.0.0.1:55433/postgres?sslmode=disable go test -tags=integration ./internal/projections -count=1
ok   github.com/kamisrini/proofbound/kernel/internal/projections  9.158s
```

The mutation harness used four calibrated runs over the complete candidate universe in the
package. The first three runs were source-scoped proving partitions; every candidate in the named
source file was exercised, and the full tagged package suite above was run separately afterward.
Each run reported `calibration neutral=survived invalid=invalid lethal=killed`.

| Source partition | Candidate range | Candidates | Killed | Invalid | Survived | Test filter |
|---|---:|---:|---:|---:|---:|---|
| `intent_report.go` | 1–58 | 58 | 58 | 0 | 0 | intent, obligation, requirement, deployment, derived-state tests |
| `projections.go` | 59–345 | 287 | 287 | 0 | 0 | complete tagged package suite |
| `report.go` | 346–404 | 59 | 59 | 0 | 0 | report, red-verdict, workflow, and render tests |

Aggregate: **404 candidates, 404 killed, 0 invalid, 0 survived**.

The optimized integration fixture gives each external-`DATABASE_URL` test a unique PostgreSQL
schema and runs independent tests in parallel. The embedded-PostgreSQL fallback remains serial
and unchanged. No candidate result from an interrupted run is included here.
